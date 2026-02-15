package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	const port = "8080"
	const filepathRoot = "."
	const filepathReadiness = "/healthz"
	const filepathApp = "/app"
	const filepathMetrics = "/metrics"
	const filepathReset = "/reset"
	const filepathApi = "/api"
	const filepathAdmin = "/admin"
	const filepathValidateChirp = "/validate_chirp"

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}
	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix(filepathApp, http.FileServer(http.Dir(filepathRoot))))
	mux.Handle(filepathApp+"/", fsHandler)

	mux.HandleFunc("GET "+filepathApi+filepathReadiness, handlerReadiness)
	mux.HandleFunc("POST "+filepathApi+filepathValidateChirp, handlerValidateChirp)

	mux.HandleFunc("GET "+filepathAdmin+filepathMetrics, apiCfg.handlerMetrics)
	mux.HandleFunc("POST "+filepathAdmin+filepathReset, apiCfg.handlerReset)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
