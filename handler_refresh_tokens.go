package main

import (
	"net/http"
	"time"

	"github.com/CybrRonin/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefreshTokensRefresh(w http.ResponseWriter, req *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid header", err)
		return
	}

	/*
		refToken, err := cfg.db.GetRefreshToken(req.Context(), token)
		if err != nil || refToken.RevokedAt.Valid || time.Now().UTC().After(refToken.ExpiresAt) {
			respondWithError(w, http.StatusUnauthorized, "unauthorized access", err)
			return
		}
	*/

	user, err := cfg.db.GetUserFromRefreshToken(req.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get user for refresh token", err)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't validate token", err)
		return
	}

	resp := response{
		Token: accessToken,
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (cfg *apiConfig) handlerRefreshTokensRevoke(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "invalid header", err)
		return
	}

	_, err = cfg.db.RevokeRefreshToken(req.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to revoke token", err)
		return
	}

	//respondWithJSON(w, http.StatusNoContent, nil)
	w.WriteHeader(http.StatusNoContent)
}
