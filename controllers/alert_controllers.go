package controllers

import (
	"crypto-tracker/models"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateAlert(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Struttura per il binding dell'input
		var input struct {
			CryptoID       string  `json:"crypto_id" binding:"required"`
			ThresholdPrice float64 `json:"threshold_price" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verifica che l'ID della criptovaluta esista ottenendo il prezzo attuale
		price, err := GetPriceUSD(input.CryptoID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Impossibile ottenere il prezzo per la criptovaluta fornita. Verifica che l'ID sia corretto."})
			return
		}

		// Crea l'alert
		alert := models.Alert{
			CryptoID:       input.CryptoID,
			ThresholdPrice: input.ThresholdPrice,
			CurrentPrice:   price,
			Triggered:      false,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if err := db.Create(&alert).Error; err != nil {
			log.Printf("Errore nella creazione dell'alert: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Errore nella creazione dell'alert: %v", err)})
			return
		}

		c.JSON(http.StatusCreated, alert)
	}
}

func GetAlerts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var alerts []models.Alert
		if err := db.Find(&alerts).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Errore nel recupero degli alert"})
			return
		}

		if len(alerts) == 0 {
			c.JSON(http.StatusOK, gin.H{"message": "Nessun alert trovato"})
			return
		}

		c.JSON(http.StatusOK, alerts)
	}
}

func GetAlert(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var alert models.Alert

		if err := db.First(&alert, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Nessun alert trovato"})
			return
		}

		c.JSON(http.StatusOK, alert)
	}
}

func UpdateAlert(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var alert models.Alert

		if err := db.First(&alert, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Alert non trovato"})
			return
		}

		var input struct {
			ThresholdPrice float64 `json:"threshold_price"`
			Triggered      bool    `json:"triggered"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Aggiorna i campi
		alert.ThresholdPrice = input.ThresholdPrice
		alert.Triggered = input.Triggered
		alert.UpdatedAt = time.Now()

		// Se l'alert viene reimpostato, aggiorna anche il prezzo corrente
		if !alert.Triggered {
			price, err := GetPriceUSD(alert.CryptoID)
			if err == nil {
				alert.CurrentPrice = price
			}
		}

		db.Save(&alert)
		c.JSON(http.StatusOK, alert)
	}
}

func DeleteAlert(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))

		if err := db.First(&models.Alert{}, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Alert non trovato"})
			return
		}

		if err := db.Delete(&models.Alert{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Errore nella cancellazione"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Alert eliminato"})
	}
}

// GetActiveAlerts restituisce tutti gli alert non ancora triggerati
func GetActiveAlerts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var alerts []models.Alert

		// Ottiene solo gli alert con Triggered = false
		if err := db.Where("triggered = ?", false).Find(&alerts).Error; err != nil {
			log.Printf("Errore nel recupero degli alert attivi: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Errore nel recupero degli alert attivi"})
			return
		}

		if len(alerts) == 0 {
			c.JSON(http.StatusOK, gin.H{"message": "Nessun alert attivo trovato"})
			return
		}

		c.JSON(http.StatusOK, alerts)
	}
}
