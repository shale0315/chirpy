package main

import (
	"encoding/json"
	"net/http"

	"github.com/shale0315/chirpy/internal/auth"
)

func (cfg *apiConfig) LoginHandler(w http.ResponseWriter, r *http.Request) {
	type Credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	userCreds := Credentials{}
	err := decoder.Decode(&userCreds)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	user, user_lookup_err := cfg.dbQueries.Login(r.Context(), userCreds.Email)
	if user_lookup_err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", user_lookup_err)
		return
	}
	check_hash_err := auth.CheckPasswordHash(userCreds.Password, user.HashedPassword)
	if check_hash_err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", check_hash_err)
		return
	}
	respondWithJson(w, http.StatusOK, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}
