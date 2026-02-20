package main

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/CybrRonin/Chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUpgradeUser(w http.ResponseWriter, req *http.Request) {
	type requestParams struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unable to find API key", err)
		return
	}
	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "API key is invalid", err)
		return
	}

	reqParams := requestParams{}
	err = decodeJSON(req.Body, &reqParams)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't decode request", err)
		return
	}

	if reqParams.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = cfg.db.UpgradeUser(req.Context(), reqParams.Data.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "user not found", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to update user", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
