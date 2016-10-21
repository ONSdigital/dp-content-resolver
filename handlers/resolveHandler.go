package handlers

import (
	"net/http"

	"github.com/ONSdigital/dp-content-resolver/resolvers"
	"github.com/ONSdigital/go-ns/log"
)

// Handler interface
type Handler interface {
	Handle(w http.ResponseWriter, req *http.Request)
}

// ResolveHandler deals with the http request and forwards the extracted url onto the resolver.
type ResolveHandler struct {
	Resolver resolvers.Resolver
}

// Handle will resolve the page defined by the path.
func (handler *ResolveHandler) Handle(w http.ResponseWriter, req *http.Request) {

	log.DebugR(req, "Resolver handler", nil)

	data, err := handler.Resolver.Resolve(req.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}
