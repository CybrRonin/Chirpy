package main

import (
	"fmt"
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
	mux.Handle(filepathApp+"/", apiCfg.middlewareMetricsInc(http.StripPrefix(filepathApp, http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc(filepathReadiness, handlerReadiness)
	mux.HandleFunc(filepathMetrics, apiCfg.handlerMetrics)
	mux.HandleFunc(filepathReset, apiCfg.handlerReset)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, req)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	msg := fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())
	w.Write([]byte(msg))
}
