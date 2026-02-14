package main

import (
	"encoding/json"
	"net/http"
)

func handlerValidateChirp(w http.ResponseWriter, req *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}

	type returnVals struct {
		Valid bool `json:"valid"`
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

	resp := returnVals{
		Valid: true,
	}
	respondWithJSON(w, http.StatusOK, resp)
}
