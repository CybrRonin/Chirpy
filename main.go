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
	platform       string
	jwtSecret      string
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
	const filepathUsers = "/users"
	const filepathChirps = "/chirps"
	const filepathLogin = "/login"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("error opening database: %s", err)
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             database.New(dbConn),
		platform:       platform,
		jwtSecret:      secret,
	}

	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix(filepathApp, http.FileServer(http.Dir(filepathRoot))))
	mux.Handle(filepathApp+"/", fsHandler)

	mux.HandleFunc("GET "+filepathApi+filepathReadiness, handlerReadiness)

	mux.HandleFunc("POST "+filepathApi+filepathUsers, apiCfg.handlerUsersCreate)
	mux.HandleFunc("POST "+filepathApi+filepathLogin, apiCfg.handlerLogin)

	mux.HandleFunc("POST "+filepathApi+filepathChirps, apiCfg.handlerChirpsCreate)
	mux.HandleFunc("GET "+filepathApi+filepathChirps, apiCfg.handlerChirpsGetAll)
	mux.HandleFunc("GET "+filepathApi+filepathChirps+"/{chirpID}", apiCfg.handlerChirpsGet)

	mux.HandleFunc("GET "+filepathAdmin+filepathMetrics, apiCfg.handlerMetrics)
	mux.HandleFunc("POST "+filepathAdmin+filepathReset, apiCfg.handlerReset)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
