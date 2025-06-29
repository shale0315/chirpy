package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shale0315/chirpy/internal/auth"
	"github.com/shale0315/chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	type Body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	userInfo := Body{}
	err := decoder.Decode(&userInfo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	hashed_pw, hashing_err := auth.HashPassword(userInfo.Password)
	if hashing_err != nil {
		respondWithError(w, http.StatusBadRequest, "Error hashing password", hashing_err)
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          userInfo.Email,
		HashedPassword: hashed_pw,
	})
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

func (cfg *apiConfig) PostRefresh(w http.ResponseWriter, r *http.Request) {
	type TokenResponse struct {
		Token string `json:"token"`
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}
	refresh_token, get_user_err := cfg.dbQueries.GetUserFromRefresh(r.Context(), token)
	if get_user_err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to retrieve user information", get_user_err)
		return
	}
	if (time.Now().After(refresh_token.ExpiresAt)) || (refresh_token.RevokedAt.Valid) {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", nil)
		return
	}
	jwt_token, jwt_err := auth.MakeJWT(refresh_token.UserID, cfg.secret)
	if jwt_err != nil {
		respondWithError(w, http.StatusBadRequest, "Error forming token", jwt_err)
		return
	}
	respondWithJson(w, http.StatusOK, TokenResponse{jwt_token})
}

func (cfg *apiConfig) PostRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}
	_, revoke_err := cfg.dbQueries.RevokeToken(r.Context(), token)
	if revoke_err != nil {
		respondWithError(w, http.StatusBadRequest, "Error updating database", revoke_err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
