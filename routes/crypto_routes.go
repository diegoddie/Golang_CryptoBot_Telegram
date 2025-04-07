package routes

import (
	"crypto-tracker/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupCryptoRoutes configura le routes per le operazioni relative alle criptovalute
func SetupCryptoRoutes(router *gin.Engine, db *gorm.DB) {
	// Endpoint per ottenere il prezzo di una criptovaluta
	router.GET("/price/:id", controllers.GetCryptoPriceHandler(db))
}
