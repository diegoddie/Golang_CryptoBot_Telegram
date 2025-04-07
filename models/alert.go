package models

import (
	"time"
)

// Alert rappresenta una soglia di prezzo per una criptovaluta
type Alert struct {
	ID             uint       `gorm:"primaryKey"`
	CryptoID       string     `gorm:"column:cryptocurrency_id;type:varchar(50);not null;index"` // ID della criptovaluta per CoinGecko (es. "bitcoin", "ethereum")
	ThresholdPrice float64    `gorm:"type:decimal(20,8);not null"`                              // Prezzo soglia per l'alert
	CurrentPrice   float64    `gorm:"type:decimal(20,8);"`                                      // Prezzo corrente (ultimo noto)
	Triggered      bool       `gorm:"default:false;not null"`                                   // Se l'alert è stato attivato
	NotifiedAt     *time.Time `gorm:"type:timestamp"`                                           // Quando è stata inviata la notifica
	CreatedAt      time.Time  `gorm:"type:timestamp;not null"`
	UpdatedAt      time.Time  `gorm:"type:timestamp;not null"`
}
