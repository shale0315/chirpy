package main

import (
	"encoding/json"
	"net/http"
)

func handlerChripsValidate(w http.ResponseWriter, r *http.Request) {
	type incoming struct {
		Body string `json:"body"`
	}

	type returnValid struct {
		Body string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := incoming{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	chirp := stringCleaner(params.Body)
	respondWithJson(w, http.StatusOK, returnValid{
		Body: chirp,
	})

}
