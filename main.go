package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type CurrencyResponse struct {
	USD string `json:"dolar"`
	EUR string `json:"euro"`
}

type CurrencyData struct {
	Code string `json:"code"`
	Bid  string `json:"bid"`
}

type CurrencyCache struct {
	sync.Mutex
	Data      map[string]CurrencyData
	ExpiresAt time.Time
}

var cache CurrencyCache

func getCurrencyData() (usdValue, eurValue float64, err error) {
	cache.Lock()
	defer cache.Unlock()

	if time.Now().Before(cache.ExpiresAt) {
		usdValue, err = strconv.ParseFloat(cache.Data["USDBRL"].Bid, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("Erro ao converter valor de USD para float64: %v", err)
		}

		eurValue, err = strconv.ParseFloat(cache.Data["EURBRL"].Bid, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("Erro ao converter valor de EUR para float64: %v", err)
		}

		return usdValue, eurValue, nil
	}

	url := "https://economia.awesomeapi.com.br/last/USD-BRL,EUR-BRL"

	response, err := http.Get(url)
	if err != nil {
		return 0, 0, fmt.Errorf("Erro ao fazer a solicitação HTTP: %v", err)
	}
	defer response.Body.Close()

	var data map[string]CurrencyData
	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		return 0, 0, fmt.Errorf("Erro ao analisar a resposta JSON: %v", err)
	}

	cache.Data = data
	cache.ExpiresAt = time.Now().Add(1 * time.Minute) // Cache por 1 minuto

	usdValue, err = strconv.ParseFloat(data["USDBRL"].Bid, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("Erro ao converter valor de USD para float64: %v", err)
	}

	eurValue, err = strconv.ParseFloat(data["EURBRL"].Bid, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("Erro ao converter valor de EUR para float64: %v", err)
	}

	return usdValue, eurValue, nil
}

func currencyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Erro ao ler o corpo da solicitação", http.StatusBadRequest)
		return
	}

	var requestBody map[string]string
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		http.Error(w, "Erro ao analisar o JSON da solicitação", http.StatusBadRequest)
		return
	}

	usdValue, eurValue, err := getCurrencyData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	brlValue, err := strconv.ParseFloat(requestBody["real"], 64)
	if err != nil {
		http.Error(w, "Valor em BRL inválido", http.StatusBadRequest)
		return
	}

	usdAmount := brlValue * usdValue
	eurAmount := brlValue * eurValue

	currencyResponse := CurrencyResponse{
		USD: fmt.Sprintf("%.2f", usdAmount),
		EUR: fmt.Sprintf("%.2f", eurAmount),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=60") // Cache por 60 segundos

	json.NewEncoder(w).Encode(currencyResponse)
}

func main() {
	cache = CurrencyCache{
		Data:      make(map[string]CurrencyData),
		ExpiresAt: time.Time{},
	}

	http.HandleFunc("/convertamoeda", currencyHandler)
	http.ListenAndServe(":8123", nil)
}
