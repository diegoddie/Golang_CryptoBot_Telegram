package controllers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"gorm.io/gorm"
)

// GetPriceUSD prende l'ID della crypto (es. "bitcoin") e restituisce il prezzo in USD
func GetPriceUSD(coinID string) (float64, error) {
	client := resty.New()

	url := "https://api.coingecko.com/api/v3/simple/price"
	log.Printf("[CoinGecko] Inizio richiesta prezzo per: %s", coinID)
	startTime := time.Now()

	// Ottieni la chiave API dalla variabile d'ambiente o usa un valore di default
	apiKey := os.Getenv("COINGECKO_API_KEY")
	if apiKey == "" {
		log.Println("[CoinGecko] AVVISO: COINGECKO_API_KEY non impostata. L'API potrebbe avere limitazioni.")
	} else {
		log.Printf("[CoinGecko] Chiave API trovata (lunghezza: %d caratteri)", len(apiKey))
	}

	// Prepara la richiesta
	request := client.R().
		SetQueryParams(map[string]string{
			"ids":           coinID,
			"vs_currencies": "usd",
		}).
		SetResult(map[string]map[string]float64{})

	// Aggiungi l'header API key se disponibile
	if apiKey != "" {
		request.SetHeader("x-cg-demo-api-key", apiKey)
		log.Println("[CoinGecko] Header API key aggiunto alla richiesta")
	}

	// Esegui la richiesta GET
	log.Println("[CoinGecko] Invio richiesta a:", url)
	resp, err := request.Get(url)

	elapsedTime := time.Since(startTime)
	log.Printf("[CoinGecko] Tempo di risposta: %v", elapsedTime)

	if err != nil {
		log.Printf("[CoinGecko] ERRORE nella richiesta: %v", err)
		return 0, fmt.Errorf("errore nella richiesta a CoinGecko: %w", err)
	}

	log.Printf("[CoinGecko] Stato risposta: %s", resp.Status())

	if resp.IsError() {
		log.Printf("[CoinGecko] ERRORE risposta non valida: %s, Body: %s", resp.Status(), resp.String())
		return 0, fmt.Errorf("risposta non valida da CoinGecko: %s", resp.Status())
	}

	// Parsing del risultato
	result := resp.Result().(*map[string]map[string]float64)
	log.Printf("[CoinGecko] Risposta JSON ricevuta: %v", *result)

	price, ok := (*result)[coinID]["usd"]
	if !ok {
		log.Printf("[CoinGecko] ERRORE prezzo non trovato per %s", coinID)
		return 0, fmt.Errorf("prezzo non trovato per %s", coinID)
	}

	log.Printf("[CoinGecko] Prezzo ottenuto con successo per %s: $%.2f USD", coinID, price)
	return price, nil
}

// GetCryptoPriceHandler restituisce un handler Gin per ottenere il prezzo di una criptovaluta
func GetCryptoPriceHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		coinID := c.Param("id")

		if coinID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID criptovaluta non fornito"})
			return
		}

		// Ottieni il prezzo usando la funzione GetPriceUSD
		price, err := GetPriceUSD(coinID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Impossibile ottenere il prezzo. Verifica che l'ID della criptovaluta sia corretto."})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":        coinID,
			"price":     price,
			"timestamp": time.Now(),
		})
	}
}
