package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
)

func handlerValidateChirp(w http.ResponseWriter, req *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}

	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(req.Body)
	chp := chirp{}

	err := decoder.Decode(&chp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode chirp", err)
		return
	}

	const maxChirpLength = 140
	if len(chp.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	cleanedPost := cleanPost(chp.Body)
	resp := returnVals{
		CleanedBody: cleanedPost,
	}
	respondWithJSON(w, http.StatusOK, resp)
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
