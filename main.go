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

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}
	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix(filepathApp, http.FileServer(http.Dir(filepathRoot))))
	mux.Handle(filepathApp+"/", fsHandler)
	mux.HandleFunc("GET "+filepathReadiness, handlerReadiness)
	mux.HandleFunc("GET "+filepathMetrics, apiCfg.handlerMetrics)
	mux.HandleFunc("POST "+filepathReset, apiCfg.handlerReset)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
