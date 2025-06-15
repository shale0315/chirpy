package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	type Email struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	emailAddr := Email{}
	err := decoder.Decode(&emailAddr)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	user, err := cfg.dbQueries.CreateUser(r.Context(), emailAddr.Email)
	if err != nil {
		respondWithError(w, 500, "Error creating user", err)
		return
	}
	jsonUser := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	respondWithJson(w, 201, jsonUser)
}
