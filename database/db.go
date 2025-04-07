package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Inizializza la connessione al database
func InitDB() *gorm.DB {
	// Carica le variabili d'ambiente
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Errore nel caricare il file .env: %v", err)
	}

	// Ottieni la stringa di connessione dal file .env
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL non è impostata nell'ambiente")
	}

	// Connessione a PostgreSQL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Errore nella connessione al database: %v", err)
	}

	// Verifica la connessione
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Errore nell'ottenere il db: %v", err)
	}
	err = sqlDB.Ping()
	if err != nil {
		log.Fatalf("Errore nel ping del database: %v", err)
	}

	// Log per confermare che la connessione è stata stabilita
	fmt.Println("Connessione al database avvenuta con successo!")

	return db
}
