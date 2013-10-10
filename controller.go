package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/aybabtme/bricklayer/bricks"
	"github.com/aybabtme/color/brush"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

func RequestTimeLog(start time.Time) {
	sout.Printf("%s %s", brush.DarkGray("done in"), brush.DarkGreen(time.Since(start).String()))
}

func AllPartsHandler(resp http.ResponseWriter, req *http.Request) {
	defer RequestTimeLog(time.Now())

	partsIndex, ok, err := db.Get(fmt.Sprintf("%s/%s", indexPath, partsIndex))
	if err != nil {
		response := "cannot fulfill request for all parts"
		responseCode := http.StatusServiceUnavailable
		logErr(req, err, response, responseCode)
		http.Error(resp, response, responseCode)
		return
	}
	if !ok {
		response := "index not built"
		responseCode := http.StatusServiceUnavailable
		logErr(req, err, response, responseCode)
		http.Error(resp, response, responseCode)
		return
	}

	respondIfChanged(resp, req, partsIndex)
}

func PartsHandler(resp http.ResponseWriter, req *http.Request) {
	defer RequestTimeLog(time.Now())

	vars := mux.Vars(req)
	name := vars["name"]

	if name == "" {
		response := "missing biobrick name"
		responseCode := http.StatusBadRequest
		logInfo(req, response, responseCode)
		http.Error(resp, response, responseCode)
		return
	}

	partData, ok, err := db.Get(fmt.Sprintf("%s/%s", partsPath, name))

	if err != nil {
		response := "cannot fulfill request for part named " + name
		responseCode := http.StatusServiceUnavailable
		logErr(req, err, response, responseCode)
		http.Error(resp, response, responseCode)
		return
	}

	if !ok {
		logInfo(req, "not found", http.StatusNotFound)
		http.NotFound(resp, req)
		return
	}

	respondIfChanged(resp, req, partData)

}

func ExtendedPartsHandler(resp http.ResponseWriter, req *http.Request) {
	defer RequestTimeLog(time.Now())
	vars := mux.Vars(req)
	name := vars["name"]

	if name == "" {
		response := "missing biobrick name"
		responseCode := http.StatusBadRequest
		logInfo(req, response, responseCode)
		http.Error(resp, response, responseCode)
		return
	}

	partData, ok, err := db.Get(fmt.Sprintf("%s/%s", extendedPartsPath, name))

	if err != nil {
		response := "cannot fulfill request for part named " + name
		responseCode := http.StatusServiceUnavailable
		logErr(req, err, response, responseCode)
		http.Error(resp, response, responseCode)
		return
	}

	if ok {
		respondIfChanged(resp, req, partData)
		return
	}

	// If the part is not already in the DB, verify that it's actually a real part
	_, partExists, err := db.Get(fmt.Sprintf("%s/%s", partsPath, name))
	if err != nil {
		response := "cannot fulfill request for reverse lookup on part named " + name
		responseCode := http.StatusServiceUnavailable
		logErr(req, err, response, responseCode)
		http.Error(resp, response, responseCode)
		return
	}
	// If it's not, give up
	if !partExists {
		logInfo(req, "not found", http.StatusNotFound)
		http.NotFound(resp, req)
		return
	}

	// If it's a real part, go query it from the iGem API
	sout.Printf("extended part '%s' not found locally, querying iGem API", name)
	s := time.Now()
	parts, err := bricks.QueryExtendedBiobricks(name)
	sout.Printf("done in %s", time.Since(s))

	if err != nil {
		response := "cannot fulfill request, iGem API lookup failed for part " + name
		responseCode := http.StatusServiceUnavailable
		logErr(req, err, response, responseCode)
		http.Error(resp, response, responseCode)
		return
	}

	if len(parts) < 1 {
		logInfo(req, "part not found on iGem API, but should be there, part="+name, http.StatusNotFound)
		http.NotFound(resp, req)
		return
	}

	if len(parts) > 1 {
		sout.Printf("found %d parts instead of 1 for name=%s", len(parts), name)
	}

	partRawJSON := make([][]byte, len(parts))
	for i, part := range parts {
		partRawJSON[i], err = json.Marshal(part)
		if err != nil {
			response := "cannot fulfill request, iGem API lookup failed for part " + name
			responseCode := http.StatusServiceUnavailable
			logErr(req, err, response, responseCode)
			http.Error(resp, response, responseCode)
			return
		}
		respondIfChanged(resp, req, partRawJSON[i])

		// Persist it asynchronously
		go func(name string, data []byte) {
			db.Put(fmt.Sprintf("%s/%s", extendedPartsPath, name), data)
		}(part.Name, partRawJSON[i])
	}
}

func computeHash(data []byte) string {
	h := md5.New()
	_, err := h.Write(data)
	if err != nil {
		panic(err)
	}
	return string(h.Sum(nil))
}

func respondIfChanged(resp http.ResponseWriter, req *http.Request, data []byte) {
	currentHash := computeHash(data)
	ifNotMatchHeader := "If-None-Match"

	eTagHash := req.Header.Get(ifNotMatchHeader)
	if len(eTagHash) == 0 || eTagHash != currentHash {
		resp.Header().Set("ETag", currentHash)

		response := "ok"
		responseCode := http.StatusOK
		logInfo(req, response, responseCode)

		respondWithBytes(resp, req, data)
		return
	}

	response := "not modified"
	responseCode := http.StatusNotModified
	logInfo(req, response, responseCode)
	resp.WriteHeader(responseCode)
}

func respondWithBytes(resp http.ResponseWriter, req *http.Request, data []byte) {
	n, err := resp.Write(data)
	if err != nil {
		logErr(req, err, "error writing to client", 0)
	}
	if n != len(data) {
		logErr(req, err, fmt.Sprintf("error writing to client, %d/%d bytes written", n, len(data)), 0)
	}
}

func colorizeHttpCode(httpCode int) string {
	codeStr := strconv.Itoa(httpCode)
	if httpCode >= 500 {
		return brush.Red(codeStr).String()
	} else if httpCode >= 400 {
		return brush.DarkRed(codeStr).String()
	} else if httpCode >= 300 {
		return brush.DarkYellow(codeStr).String()
	} else if httpCode >= 200 {
		return brush.DarkGreen(codeStr).String()
	}
	return brush.Red(codeStr).String()
}

func logInfo(req *http.Request, response string, responseCode int) {
	sout.Printf("%s - request by %s for '%s', response=%s",
		colorizeHttpCode(responseCode),
		brush.Cyan(req.RemoteAddr),
		brush.Blue(req.RequestURI),
		brush.DarkGreen(response),
	)
}

func logErr(req *http.Request, err error, response string, responseCode int) {

	serr.Printf("%s - request by %s for '%s', err = %s, response=%s",
		colorizeHttpCode(responseCode),
		brush.Cyan(req.RemoteAddr),
		brush.Blue(req.RequestURI),
		brush.Yellow(err.Error()),
		brush.DarkGreen(response),
	)
}
