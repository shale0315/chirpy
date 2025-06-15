package main

import "net/http"

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, 403, "Forbidden", nil)
	}
	cfg.fileserverHits.Store(0)
	_, err := cfg.dbQueries.ResetUsers(r.Context())
	if err != nil {
		respondWithError(w, 500, "Could not reset database", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
