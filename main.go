package main

import (
	"github.com/ONSdigital/dp-content-resolver/handlers/resolver"
	"github.com/ONSdigital/dp-content-resolver/zebedee"
	"github.com/ONSdigital/go-ns/handlers/healthcheck"
	"github.com/ONSdigital/go-ns/handlers/requestID"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/pat"
	"github.com/justinas/alice"
	"net/http"
	"os"
	"time"
)

func main() {
	bindAddr := os.Getenv("BIND_ADDR")
	if len(bindAddr) == 0 {
		bindAddr = ":8088"
	}

	zebedee.Init(time.Second*2, "http://localhost:8082")

	log.Namespace = "dp-content-resolver"

	router := pat.New()
	alice := alice.New(log.Handler, requestID.Handler(16)).Then(router)

	router.Get("/healthcheck", healthcheck.Handler)

	router.Get("/resolve", resolver.Handler)

	log.Debug("Starting server", log.Data{"bind_addr": bindAddr})

	if err := http.ListenAndServe(bindAddr, alice); err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}
}
