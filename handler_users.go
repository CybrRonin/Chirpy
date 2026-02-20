package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/CybrRonin/Chirpy/internal/auth"
	"github.com/CybrRonin/Chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID    `json:"id"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	Email        string       `json:"email"`
	Password     string       `json:"-"`
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	RevokedAt    sql.NullTime `json:"revoked_at"`
	IsChirpyRed  bool         `json:"is_chirpy_red"`
}

type userParameters struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, req *http.Request) {
	params := userParameters{}

	err := decodeJSON(req.Body, &params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error decoding user's email or password", err)
		return
	}

	hashedPwd, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error hashing password", err)
		return
	}

	dbParams := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPwd,
	}
	u, err := cfg.db.CreateUser(req.Context(), dbParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create user", err)
		return
	}

	user := mapUser(u)
	respondWithJSON(w, http.StatusCreated, user)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, req *http.Request) {
	const (
		defaultAccessExpiration  = time.Hour
		defaultRefreshExpiration = time.Hour * 1440 // 60 days' worth of hours
	)

	params := userParameters{}

	err := decodeJSON(req.Body, &params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to decode user parameters", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(req.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, defaultAccessExpiration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to generate JWT", err)
		return
	}

	refreshToken := auth.MakeRefreshToken()

	refreshArgs := database.CreateRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().UTC().Add(defaultRefreshExpiration),
	}
	_, err = cfg.db.CreateRefreshToken(req.Context(), refreshArgs)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create refresh token entry", err)
		return
	}

	respondWithJSON(w, http.StatusOK, mapUser(user, accessToken, refreshToken))
}

func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, req *http.Request) {
	accessToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "couldn't find JWT", err)
		return
	}

	accessID, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "couldn't validate JWT", err)
		return
	}

	userParams := userParameters{}
	err = decodeJSON(req.Body, &userParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to decode parameters", err)
		return
	}

	hashedPwd, err := auth.HashPassword(userParams.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to hash password", err)
		return
	}

	pwdArgs := database.UpdateUserParams{
		ID:             accessID,
		Email:          userParams.Email,
		HashedPassword: hashedPwd,
	}

	user, err := cfg.db.UpdateUser(req.Context(), pwdArgs)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error updating password", err)
		return
	}

	respondWithJSON(w, http.StatusOK, mapUser(user))
}

func mapUser(user database.User, options ...string) User {
	newUser := User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}

	if len(options) > 1 {
		newUser.Token = options[0]
		newUser.RefreshToken = options[1]
		//newUser.RevokedAt = sql.NullTime{Valid: false}
	}

	return newUser
}
