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
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

type TokenResponse struct {
	Token string `json:"token"`
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
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}
	respondWithJson(w, 201, jsonUser)
}

func (cfg *apiConfig) PostRefresh(w http.ResponseWriter, r *http.Request) {

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

func (cfg *apiConfig) UpdateCredsRequest(w http.ResponseWriter, r *http.Request) {
	token, jwt_err := auth.GetBearerToken(r.Header)
	if jwt_err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", jwt_err)
		return
	}
	userUUID, uuid_err := auth.ValidateJWT(token, cfg.secret)
	if uuid_err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", uuid_err)
		return
	}
	type NewCredsRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	newCreds := NewCredsRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&newCreds)
	if err != nil {
		respondWithError(w, http.StatusExpectationFailed, "Error decoding", err)
		return
	}
	hashedPW, err := auth.HashPassword(newCreds.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password", err)
		return
	}
	updatedUser, err := cfg.dbQueries.UpdateCreds(r.Context(), database.UpdateCredsParams{
		Email:          newCreds.Email,
		HashedPassword: hashedPW,
		ID:             userUUID,
	})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to update", err)
		return
	}
	jsonUser := User{
		ID:        updatedUser.ID,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
		Email:     updatedUser.Email,
	}
	respondWithJson(w, http.StatusOK, jsonUser)
}

func (cfg *apiConfig) UpgradeChirpyRed(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	type Data struct {
		User_id uuid.UUID `json:"user_id"`
	}
	type Webhook struct {
		Event string `json:"event"`
		Data  Data   `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	webhookRequest := Webhook{}
	err = decoder.Decode(&webhookRequest)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error processing request", err)
		return
	}
	if webhookRequest.Event != "user.upgraded" {
		respondWithJson(w, http.StatusNoContent, nil)
		return
	}
	result, err := cfg.dbQueries.UpgradeToChirpyRed(r.Context(), webhookRequest.Data.User_id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to upgrade user", err)
		return
	}
	rows_affected, err := result.RowsAffected()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get rows affected", err)
		return
	}
	if rows_affected == 0 {
		respondWithError(w, http.StatusNotFound, "User not found", nil)
		return
	}
	respondWithJson(w, http.StatusNoContent, nil)
}
