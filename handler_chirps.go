package main

import (
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/CybrRonin/Chirpy/internal/auth"
	"github.com/CybrRonin/Chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
		//UserID uuid.UUID `json:"user_id"`
	}

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	uID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	reqParams := parameters{}
	err = decodeJSON(req.Body, &reqParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to decode chirp parameters", err)
		return
	}

	cleaned, err := validateChirp(reqParams.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp: ", err)
		return
	}

	params := database.CreateChirpParams{
		Body:   cleaned,
		UserID: uID,
	}
	ch, err := cfg.db.CreateChirp(req.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create chirp", err)
		return
	}

	chirp := mapChirp(ch)
	respondWithJSON(w, http.StatusCreated, chirp)
}

func (cfg *apiConfig) handlerChirpsGetAll(w http.ResponseWriter, req *http.Request) {
	resp, err := cfg.db.GetAllChirps(req.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to retrieve chirps", err)
		return
	}

	chirps := []Chirp{}
	for _, entry := range resp {
		chirps = append(chirps, mapChirp(entry))
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, req *http.Request) {
	chirpID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid chirp ID", err)
		return
	}

	dbChirp, err := cfg.db.GetChirp(req.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "couldn't retrieve chirp", err)
		return
	}

	respondWithJSON(w, http.StatusOK, mapChirp(dbChirp))
}

func mapChirp(ch database.Chirp) Chirp {
	return Chirp{
		ID:        ch.ID,
		CreatedAt: ch.CreatedAt,
		UpdatedAt: ch.UpdatedAt,
		Body:      ch.Body,
		UserID:    ch.UserID,
	}
}

func validateChirp(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	cleaned := cleanPost(body)
	return cleaned, nil
}

func cleanPost(msg string) string {
	var censoredWords = []string{"kerfuffle", "sharbert", "fornax"}
	const censored = "****"
	msgWords := strings.Split(msg, " ")

	for i, word := range msgWords {
		if slices.Contains(censoredWords, strings.ToLower(word)) {
			msgWords[i] = censored
		}
	}

	return strings.Join(msgWords, " ")
}
