package main

import (
	"crypto/md5"
	"fmt"
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

	partData, ok, err := db.Get(fmt.Sprintf("parts/%s", name))

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
