package main

import (
	"net/http"
	"slices"
	"strings"
	"time"

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
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	reqParams := parameters{}
	err := decodeJSON(req.Body, &reqParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to decode chirp parameters", err)
		return
	}

	if !validateChirp(w, &reqParams.Body) {
		return
	}

	params := database.CreateChirpParams{
		Body:   reqParams.Body,
		UserID: reqParams.UserID,
	}
	ch, err := cfg.db.CreateChirp(req.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create chirp", err)
		return
	}

	chirp := mapChirp(ch)
	respondWithJSON(w, http.StatusCreated, chirp)
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

func validateChirp(w http.ResponseWriter, body *string) bool {
	const maxChirpLength = 140
	if len(*body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return false
	}

	*body = cleanPost(*body)
	return true
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
