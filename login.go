package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shale0315/chirpy/internal/auth"
	"github.com/shale0315/chirpy/internal/database"
)

type UserLogin struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) LoginHandler(w http.ResponseWriter, r *http.Request) {
	type Credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		// ExpiresInSeconds *int   `json:"expires_in_seconds,omitempty"`
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

	jwt_token, token_error := auth.MakeJWT(user.ID, cfg.secret)
	if token_error != nil {
		respondWithError(w, http.StatusBadRequest, "Error creating JWT token", token_error)
		return
	}

	refresh_token, refresh_token_err := auth.MakeRefreshToken()
	if refresh_token_err != nil {
		respondWithError(w, http.StatusBadRequest, "Error creating refresh token", refresh_token_err)
		return
	}

	_, store_rt_err := cfg.dbQueries.StoreRefreshToken(r.Context(), database.StoreRefreshTokenParams{
		Token:     refresh_token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	})
	if store_rt_err != nil {
		respondWithError(w, http.StatusExpectationFailed, "Failed to store in database", store_rt_err)
	}

	respondWithJson(w, http.StatusOK, UserLogin{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        jwt_token,
		RefreshToken: refresh_token,
		IsChirpyRed:  user.IsChirpyRed,
	})
}
