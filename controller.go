package main

import (
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

	respondWithBytes(resp, req, partsIndex)
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

	respondWithBytes(resp, req, partData)

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

func logInfo(req *http.Request, response string, responseCode int) {
	sout.Printf("%s - request by %s for '%s', response=%s",
		brush.DarkYellow(strconv.Itoa(responseCode)),
		brush.Cyan(req.RemoteAddr),
		brush.Blue(req.RequestURI),
		brush.DarkGreen(response),
	)
}

func logErr(req *http.Request, err error, response string, responseCode int) {
	serr.Printf("%s - request by %s for '%s', err = %s, response=%s",
		brush.Red(strconv.Itoa(responseCode)),
		brush.Cyan(req.RemoteAddr),
		brush.Blue(req.RequestURI),
		brush.Yellow(err.Error()),
		brush.DarkGreen(response),
	)
}
