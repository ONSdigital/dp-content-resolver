package handlers

import (
    "encoding/json"
    "github.com/ONSdigital/dp-content-resolver/content"
    "github.com/ONSdigital/dp-content-resolver/model"
    "github.com/ONSdigital/go-ns/log"
    "net/http"
)

// Resolve is the function called to resolve page data.
// Its defined as an exported package member allowing
// alternative implementations to be injected
var Resolve = content.Resolve

// Handle will resolve the page defined by the path.
func Handle(w http.ResponseWriter, req *http.Request) {

    log.DebugR(req, "Resolver handler", nil)

    w.Header().Set("Content-Type", "application/json")

    data, err := Resolve(req)
    if err != nil {
        writeErrorResponse(err, w)
        log.ErrorR(req, err, err.GetLogData())
    }

    w.WriteHeader(200)
    w.Write(data)
}

func writeErrorResponse(err error, w http.ResponseWriter) {
    w.WriteHeader(http.StatusBadRequest)
    jsonEncoder := json.NewEncoder(w)
    jsonEncoder.Encode(model.ErrorResponse{
        Error: err.Error(),
    })
}
