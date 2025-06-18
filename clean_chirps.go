package main

import (
	"strings"
)

func stringCleaner(s string) string {
	wordList := strings.Split(s, " ")
	var newWordList []string
	for _, word := range wordList {
		loweredWord := strings.ToLower(word)
		if (loweredWord == "kerfuffle") || (loweredWord == "sharbert") || (loweredWord == "fornax") {
			newWordList = append(newWordList, "****")
		} else {
			newWordList = append(newWordList, word)
		}
	}
	newString := strings.Join(newWordList, " ")
	return newString
}
