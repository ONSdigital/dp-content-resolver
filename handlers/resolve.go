package handlers

import (
	"net/http"
	"github.com/ONSdigital/dp-content-resolver/content"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/dp-content-resolver/model"
	"encoding/json"
)

// Handler interface
type Handler interface {
	Handle(w http.ResponseWriter, req *http.Request)
}

// ResolveHandler deals with the http request and forwards the extracted url onto the resolver.
type ResolveHandler struct {
	Resolver content.Resolver
}

// Handle will resolve the page defined by the path.
func (handler *ResolveHandler) Handle(w http.ResponseWriter, req *http.Request) {

	log.DebugR(req, "Resolver handler", nil)

	w.Header().Set("Content-Type", "application/json")

	data, err := handler.Resolver.Resolve(req.URL.Path)
	if err != nil {
		writeErrorResponse(err, w)
		log.ErrorR(req, err, nil)
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
