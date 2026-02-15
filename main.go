package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/CybrRonin/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
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

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("error opening database: %s", err)
	}

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             database.New(dbConn),
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
