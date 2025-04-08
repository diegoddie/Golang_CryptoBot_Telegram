package services

import (
	"crypto-tracker/controllers"
	"crypto-tracker/models"
	"log"
	"time"

	"gorm.io/gorm"
)

type AlertMonitor struct {
	db               *gorm.DB
	interval         time.Duration
	stopChan         chan struct{}
	telegramNotifyCh chan *models.Alert // Canale per notifiche Telegram
}

// NewAlertMonitor crea una nuova istanza del monitor degli alert
func NewAlertMonitor(db *gorm.DB, interval time.Duration) *AlertMonitor {
	if interval < time.Second {
		interval = time.Minute // Valore di default
	}

	return &AlertMonitor{
		db:               db,
		interval:         interval,
		stopChan:         make(chan struct{}),
		telegramNotifyCh: make(chan *models.Alert, 100), // Buffer per le notifiche
	}
}

// SetTelegramNotificationChannel imposta il canale per inviare notifiche Telegram
func (am *AlertMonitor) SetTelegramNotificationChannel(ch chan *models.Alert) {
	am.telegramNotifyCh = ch
}

// Start avvia il monitoraggio in background
func (am *AlertMonitor) Start() {
	log.Printf("[AlertMonitor] Avvio del monitoraggio (intervallo: %v)", am.interval)

	go func() {
		// Esegui subito il primo controllo
		am.checkAlerts()

		ticker := time.NewTicker(am.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				am.checkAlerts()
			case <-am.stopChan:
				log.Println("[AlertMonitor] Monitoraggio terminato")
				return
			}
		}
	}()
}

// Stop interrompe il monitoraggio in background
func (am *AlertMonitor) Stop() {
	log.Println("[AlertMonitor] Arresto del monitoraggio...")
	close(am.stopChan)
}

// checkAlerts verifica tutti gli alert attivi
func (am *AlertMonitor) checkAlerts() {
	log.Println("[AlertMonitor] Controllo degli alert attivi...")

	var activeAlerts []models.Alert
	if err := am.db.Where("triggered = ?", false).Find(&activeAlerts).Error; err != nil {
		log.Printf("[AlertMonitor] Errore nel recupero degli alert: %v", err)
		return
	}

	log.Printf("[AlertMonitor] Trovati %d alert attivi", len(activeAlerts))
	if len(activeAlerts) == 0 {
		return
	}

	// Controlla ogni alert in sequenza
	// Non c'è vero bisogno di parallelizzare questa operazione
	// a meno che non si abbiano migliaia di alert da controllare
	for i := range activeAlerts {
		if err := am.processSingleAlert(&activeAlerts[i]); err != nil {
			log.Printf("[AlertMonitor] Errore per alert ID %d: %v", activeAlerts[i].ID, err)
		}
	}
}

// processSingleAlert verifica e aggiorna un singolo alert
func (am *AlertMonitor) processSingleAlert(alert *models.Alert) error {
	// Ottieni il prezzo corrente
	price, err := controllers.GetPriceUSD(alert.CryptoID)
	if err != nil {
		return err
	}

	// Aggiorna il prezzo corrente
	alert.CurrentPrice = price
	now := time.Now().UTC() // Usa UTC per i timestamp nel database
	alert.UpdatedAt = now

	// Salva le vecchie informazioni per decidere se inviare notifiche
	wasTriggeredBefore := alert.Triggered

	// Verifica la condizione di trigger
	if price >= alert.ThresholdPrice {
		log.Printf("[AlertMonitor] ALERT TRIGGERATO! ID: %d, Crypto: %s, Soglia: %.2f, Prezzo: %.2f",
			alert.ID, alert.CryptoID, alert.ThresholdPrice, price)

		alert.Triggered = true
		alert.NotifiedAt = &now
	}

	// Invia notifica Telegram se l'alert è stato appena triggerato
	if alert.Triggered && !wasTriggeredBefore {
		am.sendTelegramNotification(alert)
	}

	// Salva nel database
	return am.db.Save(alert).Error
}

// sendTelegramNotification invia una notifica Telegram
func (am *AlertMonitor) sendTelegramNotification(alert *models.Alert) {
	log.Printf("[AlertMonitor] Tentativo di invio notifica Telegram per alert ID %d...", alert.ID)

	// Controllo se il canale è stato configurato
	if am.telegramNotifyCh == nil {
		log.Printf("[AlertMonitor] ⚠️ ERRORE: canale notifiche Telegram non configurato!")
		return
	}

	log.Printf("[AlertMonitor] Canale notifiche Telegram presente, invio notifica...")

	select {
	case am.telegramNotifyCh <- alert:
		log.Printf("[AlertMonitor] ✅ Notifica Telegram inviata con successo per alert ID %d", alert.ID)
	default:
		log.Printf("[AlertMonitor] ⚠️ Buffer di notifiche Telegram pieno, impossibile inviare notifica per alert ID %d", alert.ID)
	}
}
