package main

import (
	"crypto-tracker/database"
	"crypto-tracker/models"
	"crypto-tracker/routes"
	"crypto-tracker/services"
	"crypto-tracker/services/telegram"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Carica le variabili d'ambiente dal file .env
	if err := godotenv.Load(); err != nil {
		log.Println("File .env non trovato, uso variabili d'ambiente del sistema")
	}

	// Inizializza il database
	db := database.InitDB()
	if err := db.Error; err != nil {
		log.Fatalf("Errore nell'inizializzazione del database: %v", err)
	}

	// Migrazione automatica degli schemi
	err := db.AutoMigrate(&models.Alert{})
	if err != nil {
		log.Fatalf("Errore durante la migrazione: %v", err)
	}

	alertMonitor := services.NewAlertMonitor(db, 5*time.Minute)
	alertMonitor.Start()
	defer alertMonitor.Stop()

	// Inizializza il bot Telegram
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if telegramToken == "" {
		log.Println("TELEGRAM_BOT_TOKEN non impostato, il bot Telegram non sarà avviato")
	} else {
		log.Printf("Avvio del bot Telegram con token: %s***", telegramToken[:10])

		bot, err := telegram.NewTelegramBot(telegramToken, db)
		if err != nil {
			log.Printf("⚠️ ERRORE nell'inizializzazione del bot Telegram: %v", err)
		} else {
			log.Println("Bot Telegram inizializzato correttamente, avvio in corso...")
			go func() {
				log.Println("Goroutine bot Telegram avviata")
				bot.Start()
				log.Println("Bot Telegram terminato")
			}()
			log.Println("Bot Telegram avviato con successo")

			// Collega il canale di notifica del bot al monitor degli alert
			alertMonitor.SetTelegramNotificationChannel(bot.GetNotificationChannel())
			log.Println("Notifiche Telegram configurate per gli alert")
		}
	}

	// Inizializza il router Gin
	router := gin.Default()

	// Configurazione CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://golangcryptobottelegram-production.up.railway.app/", "http://localhost:8080", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Imposta le routes
	routes.SetupAlertRoutes(router, db)
	routes.SetupCryptoRoutes(router, db)

	// Avvia il server
	port := ":8080"
	fmt.Printf("Server in ascolto su http://localhost%s\n", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Errore nell'avvio del server: %v", err)
	}
}
