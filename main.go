package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
)

const creditScoreMin = 500
const creditScoreMax = 900

type credit_rating struct {
	CreditRating int `json:"credit_rating"`
}

func getCreditScore(w http.ResponseWriter, r *http.Request) {
	var creditRating = credit_rating{
		CreditRating: (rand.Intn(creditScoreMax-creditScoreMin) + creditScoreMin),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(creditRating)
}

func handleRequests() {
	http.Handle("/creditscore", http.HandlerFunc(getCreditScore))
	log.Fatal(http.ListenAndServe(":8123", nil))
}

func main() {
	handleRequests()
}
