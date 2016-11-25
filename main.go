package main

import (
	"net/http"
	"os"
	"time"

	"github.com/ONSdigital/dp-content-resolver/babbage"
	"github.com/ONSdigital/dp-content-resolver/content"
	"github.com/ONSdigital/dp-content-resolver/content/homePage"
	"github.com/ONSdigital/dp-content-resolver/handlers"
	"github.com/ONSdigital/dp-content-resolver/zebedee"
	"github.com/ONSdigital/go-ns/handlers/healthcheck"
	"github.com/ONSdigital/go-ns/handlers/requestID"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/pat"
	"github.com/justinas/alice"
)

func main() {
	bindAddr := os.Getenv("BIND_ADDR")
	if len(bindAddr) == 0 {
		bindAddr = ":20020"
	}

	zebedeeURL := os.Getenv("ZEBEDEE_URL")
	if len(zebedeeURL) == 0 {
		zebedeeURL = "http://localhost:8082"
	}

	babbageURL := os.Getenv("BABBAGE_URL")
	if len(babbageURL) == 0 {
		babbageURL = "http://localhost:8080"
	}

	content.ZebedeeService = zebedee.CreateClient(time.Second*2, zebedeeURL)
	content.BabbageService = babbage.CreateClient(time.Second*2, babbageURL)
	homePage.ZebedeeService = content.ZebedeeService

	log.Namespace = "dp-content-resolver"

	router := pat.New()
	alice := alice.New(log.Handler, requestID.Handler(16)).Then(router)

	router.Get("/healthcheck", healthcheck.Handler)
	router.Get("/{uri:.*}", handlers.Handle)

	log.Debug("Starting server", log.Data{
		"bind_addr":   bindAddr,
		"zebedee_url": zebedeeURL,
	})

	if err := http.ListenAndServe(bindAddr, alice); err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}
}
