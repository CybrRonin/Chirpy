package main

import (
	"net/http"
	"time"

	"github.com/CybrRonin/Chirpy/internal/auth"
	"github.com/CybrRonin/Chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Token     string    `json:"token"`
}

type userParameters struct {
	Email    string `json:"email"`
	Password string `json:"password"`

	ExpiresInSeconds int `json:"expires_in_seconds"`
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
		defaultExpiration = 3600 // number of seconds in an hour
		defaultDuration   = time.Hour
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

	duration := defaultDuration
	if params.ExpiresInSeconds > 0 && params.ExpiresInSeconds < defaultExpiration {
		duration = time.Duration(params.ExpiresInSeconds) * time.Second
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, duration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to generate JWT", err)
		return
	}

	respondWithJSON(w, http.StatusOK, mapUser(user, token))
}

func mapUser(user database.User, options ...string) User {
	newUser := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	if len(options) > 0 {
		newUser.Token = options[0]
	}

	return newUser
}
