package resolver

import (
	"github.com/ONSdigital/dp-content-resolver/resolver"
	"github.com/ONSdigital/go-ns/log"
	"net/http"
)

// Handler will resolve the page defined by the path.
func Handler(w http.ResponseWriter, req *http.Request) {

	log.DebugR(req, "Resolver handler", nil)

	data, err := resolver.Resolve(req.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write(data)
}
