package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shale0315/chirpy/internal/auth"
	"github.com/shale0315/chirpy/internal/database"
)

type ReturnChirp struct {
	ChirpId   uuid.UUID `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) ChirpHandler(w http.ResponseWriter, r *http.Request) {
	token, token_err := auth.GetBearerToken(r.Header)
	if token_err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or malformed auth token", token_err)
		return
	}
	userUUID, validation_error := auth.ValidateJWT(token, cfg.secret)
	if validation_error != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", validation_error)
		return
	}

	type Incoming struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := Incoming{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	cleanChirp := stringCleaner(params.Body)
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanChirp,
		UserID: userUUID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}
	respondWithJson(w, 201, ReturnChirp{
		Body:      chirp.Body,
		UserID:    userUUID,
		ChirpId:   chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
	})
}

func (cfg *apiConfig) SortChirpHandler(w http.ResponseWriter, r *http.Request) {
	var finalChirpSlice []ReturnChirp

	author_id := r.URL.Query().Get("author_id")
	if author_id != "" {
		authorUUID, err := uuid.Parse(author_id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to parse ID", err)
			return
		}
		chirpsByID, err := cfg.dbQueries.GetChirpsByAuthor(r.Context(), authorUUID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Could not retrieve chirps", err)
		}
		for _, chirp := range chirpsByID {
			transformedChirp := ReturnChirp{
				Body:      chirp.Body,
				UserID:    chirp.UserID,
				ChirpId:   chirp.ID,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
			}
			finalChirpSlice = append(finalChirpSlice, transformedChirp)
		}
		respondWithJson(w, http.StatusOK, finalChirpSlice)
		return
	}

	sortedChirps, err := cfg.dbQueries.SortChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve chirps", err)
		return
	}

	for _, chirp := range sortedChirps {
		transformedChirp := ReturnChirp{
			Body:      chirp.Body,
			UserID:    chirp.UserID,
			ChirpId:   chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
		}
		finalChirpSlice = append(finalChirpSlice, transformedChirp)
	}

	respondWithJson(w, http.StatusOK, finalChirpSlice)
}

func (cfg *apiConfig) GetChirp(w http.ResponseWriter, r *http.Request) {
	chirp_id := r.PathValue("chirp_id")
	chirp_id_uuid, err := uuid.Parse(chirp_id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error parsing id", err)
		return
	}
	chirp, err := cfg.dbQueries.GetChirp(r.Context(), (chirp_id_uuid))
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, 404, "Error finding chirp", err)
			return
		}
		respondWithError(w, 400, "Other error", err)
		return
	}

	respondWithJson(w, http.StatusOK, ReturnChirp{
		Body:      chirp.Body,
		UserID:    chirp.UserID,
		ChirpId:   chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
	})
}

func (cfg *apiConfig) DeleteChirpRequest(w http.ResponseWriter, r *http.Request) {
	// Get auth token
	token, token_err := auth.GetBearerToken(r.Header)
	if token_err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", token_err)
		return
	}

	// Validate auth token and get UUID of of user
	userUUID, valid_err := auth.ValidateJWT(token, cfg.secret)
	if valid_err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", valid_err)
		return
	}

	// Parse ID of chirp from path
	chirp_id := r.PathValue("chirp_id")
	chirp_uuid, err := uuid.Parse(chirp_id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error parsing id", err)
		return
	}
	// Check if chirp exists
	chirpJSON, err := cfg.dbQueries.GetChirp(r.Context(), chirp_uuid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Could not find chirp", err)
		return
	}
	if chirpJSON.UserID != userUUID {
		respondWithError(w, http.StatusForbidden, "Forbidden", nil)
		return
	}
	result, err := cfg.dbQueries.DeleteChirp(r.Context(), chirp_uuid)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete chirp", err)
		return
	}
	rows_affected, err := result.RowsAffected()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get rows affected", err)
	}
	if rows_affected == 0 {
		respondWithError(w, http.StatusNotFound, "Could not find chirp", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
