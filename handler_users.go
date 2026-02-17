package main

import (
	"net/http"
	"time"

	"github.com/CybrRonin/Chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, req *http.Request) {
	type email struct {
		Email string
	}

	params := email{}

	err := decodeJSON(req.Body, &params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't decode user's email", err)
		return
	}

	u, err := cfg.db.CreateUser(req.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create user", err)
		return
	}

	user := mapUser(u)
	respondWithJSON(w, http.StatusCreated, user)
}

func mapUser(user database.User) User {
	newUser := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	return newUser
}
