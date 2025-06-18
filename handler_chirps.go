package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
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
		UserID: params.UserID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}
	respondWithJson(w, 201, ReturnChirp{
		Body:      chirp.Body,
		UserID:    chirp.UserID,
		ChirpId:   chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
	})
}

func (cfg *apiConfig) SortChirpHandler(w http.ResponseWriter, r *http.Request) {
	var finalChirpSlice []ReturnChirp
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
