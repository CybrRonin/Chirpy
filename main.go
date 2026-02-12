package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"
	const filepathRoot = "."
	const filepathReadiness = "/healthz"
	const filepathApp = "/app"

	serveMux := http.NewServeMux()
	serveMux.Handle(filepathApp+"/", http.StripPrefix(filepathApp, http.FileServer(http.Dir(filepathRoot))))
	serveMux.HandleFunc(filepathReadiness, handlerReadiness)
	srv := http.Server{
		Addr:    ":" + port,
		Handler: serveMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
