package telegram

import (
	"crypto-tracker/controllers"
	"crypto-tracker/models"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

// Timezone italiano (UTC+2)
var italianTimezone, _ = time.LoadLocation("Europe/Rome")

// TelegramBot gestisce l'interazione con il bot Telegram
type TelegramBot struct {
	bot      *tgbotapi.BotAPI
	db       *gorm.DB
	chatIDs  map[int64]bool // Mappa delle chat IDs attive
	chatLock sync.RWMutex   // Per accesso thread-safe alla mappa
}

// NewTelegramBot crea una nuova istanza del bot Telegram
func NewTelegramBot(token string, db *gorm.DB) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("errore nell'inizializzazione del bot: %w", err)
	}

	log.Printf("Bot autorizzato con account %s", bot.Self.UserName)

	return &TelegramBot{
		bot:     bot,
		db:      db,
		chatIDs: make(map[int64]bool),
	}, nil
}

// Start avvia il bot e inizia ad ascoltare i messaggi e le notifiche
func (t *TelegramBot) Start() {
	log.Println("[Telegram] Avvio del bot in corso...")
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	log.Println("[Telegram] Configurazione del canale updates completata, richiesta updates a Telegram...")

	updates := t.bot.GetUpdatesChan(u)
	log.Printf("[Telegram] Canale updates ottenuto con successo: %v", updates != nil)

	// Gestisce i messaggi in arrivo
	log.Println("[Telegram] Avvio goroutine di gestione messaggi...")
	go func() {
		log.Println("[Telegram] Goroutine di ascolto messaggi avviata")
		messageCount := 0

		for update := range updates {
			messageCount++
			log.Printf("[Telegram] Ricevuto update #%d da Telegram", messageCount)

			if update.Message == nil {
				log.Println("[Telegram] Update senza messaggio, ignoro")
				continue
			}

			// Memorizza l'ID della chat per le notifiche future
			t.registerChatID(update.Message.Chat.ID)

			chatID := update.Message.Chat.ID
			log.Printf("[Telegram] Processando messaggio da chat ID %d: %s", chatID, update.Message.Text)

			go t.handleMessage(update.Message)
		}

		log.Println("[Telegram] Loop di aggiornamenti interrotto! Il bot non riceverà più messaggi!")
	}()

	log.Println("[Telegram] Bot avviato correttamente e in ascolto di messaggi")
}

// registerChatID registra un ID chat per future notifiche
func (t *TelegramBot) registerChatID(chatID int64) {
	t.chatLock.Lock()
	defer t.chatLock.Unlock()

	if _, exists := t.chatIDs[chatID]; !exists {
		t.chatIDs[chatID] = true
		log.Printf("[Telegram] Nuovo utente registrato con chat ID: %d", chatID)
	}
}

// GetAllChatIDs restituisce tutti gli ID chat registrati
func (t *TelegramBot) GetAllChatIDs() []int64 {
	t.chatLock.RLock()
	defer t.chatLock.RUnlock()

	chatIDs := make([]int64, 0, len(t.chatIDs))
	for id := range t.chatIDs {
		chatIDs = append(chatIDs, id)
	}

	return chatIDs
}

// handleMessage gestisce i messaggi in arrivo
func (t *TelegramBot) handleMessage(message *tgbotapi.Message) {
	log.Printf("[Telegram] Messaggio da %s: %s", message.From.UserName, message.Text)

	if !message.IsCommand() {
		t.sendMessage(message.Chat.ID, "Invia un comando, ad esempio /help o /price bitcoin")
		return
	}

	switch message.Command() {
	case "start", "help":
		t.handleHelp(message)
	case "price":
		t.handlePrice(message)
	case "create_alert":
		t.handleCreateAlert(message)
	case "update_alert":
		t.handleUpdateAlert(message)
	case "alerts":
		t.handleGetAlerts(message)
	case "active_alerts":
		t.handleGetActiveAlerts(message)
	case "alert":
		t.handleGetAlert(message)
	case "delete_alert":
		t.handleDeleteAlert(message)
	default:
		t.sendMessage(message.Chat.ID, "Comando non riconosciuto. Usa /help per vedere i comandi disponibili.")
	}
}

// handleHelp gestisce il comando /help
func (t *TelegramBot) handleHelp(message *tgbotapi.Message) {
	helpText := `
Comandi disponibili:
/price <crypto_id> - Ottiene il prezzo attuale (es: /price bitcoin)
/create_alert <crypto_id> <threshold_price> - Crea un nuovo alert (es: /create_alert bitcoin 30000)
/update_alert <id> <threshold_price> - Aggiorna un alert esistente (es: /update_alert 1 32000)
/alerts - Mostra tutti gli alert
/active_alerts - Mostra solo gli alert attivi (non triggerati)
/alert <id> - Mostra i dettagli di un alert specifico
/delete_alert <id> - Elimina un alert specifico
/help - Mostra questo messaggio
`
	t.sendMessage(message.Chat.ID, helpText)
}

// handlePrice gestisce il comando /price
func (t *TelegramBot) handlePrice(message *tgbotapi.Message) {
	args := strings.Fields(message.CommandArguments())
	if len(args) < 1 {
		t.sendMessage(message.Chat.ID, "Specifica l'ID della criptovaluta. Esempio: /price bitcoin")
		return
	}

	// Converti l'ID in lowercase per evitare problemi con maiuscole/minuscole
	coinID := strings.ToLower(args[0])
	price, err := controllers.GetPriceUSD(coinID)
	if err != nil {
		t.sendMessage(message.Chat.ID, fmt.Sprintf("Errore: %v", err))
		return
	}

	t.sendMessage(message.Chat.ID, fmt.Sprintf("💰 %s: $%.2f USD", coinID, price))
}

// handleCreateAlert gestisce il comando /create_alert
func (t *TelegramBot) handleCreateAlert(message *tgbotapi.Message) {
	args := strings.Fields(message.CommandArguments())
	if len(args) < 2 {
		t.sendMessage(message.Chat.ID, "Formato: /create_alert <crypto_id> <threshold_price>")
		return
	}

	// Converti l'ID in lowercase
	coinID := strings.ToLower(args[0])
	thresholdPrice, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		t.sendMessage(message.Chat.ID, "Prezzo non valido. Usa un numero decimale.")
		return
	}

	// Usa la stessa logica di validazione presente in controllers.CreateAlert
	price, err := controllers.GetPriceUSD(coinID)
	if err != nil {
		t.sendMessage(message.Chat.ID, fmt.Sprintf("Errore: Impossibile ottenere il prezzo per '%s'. Verifica che l'ID sia corretto.", coinID))
		return
	}

	// Crea l'alert utilizzando le stesse logiche dei controller
	// Usa UTC per i timestamp nel database
	now := time.Now().UTC()
	alert := models.Alert{
		CryptoID:       coinID,
		ThresholdPrice: thresholdPrice,
		CurrentPrice:   price,
		Triggered:      false,
		CreatedAt:      now,
		UpdatedAt:      now,
		UserChatID:     message.Chat.ID, // Salva l'ID della chat dell'utente
	}

	if err := t.db.Create(&alert).Error; err != nil {
		t.sendMessage(message.Chat.ID, fmt.Sprintf("Errore nella creazione dell'alert: %v", err))
		return
	}

	// Converti la data in fuso orario italiano solo per la visualizzazione
	createdAtLocal := alert.CreatedAt.In(italianTimezone)
	t.sendMessage(message.Chat.ID, fmt.Sprintf("✅ Alert creato! ID: %d\nCrypto: %s\nSoglia: $%.2f\nPrezzo attuale: $%.2f\nCreato il: %s (CET)",
		alert.ID, alert.CryptoID, alert.ThresholdPrice, alert.CurrentPrice, createdAtLocal.Format("02/01/2006 15:04")))
}

// handleUpdateAlert gestisce il comando /update_alert
func (t *TelegramBot) handleUpdateAlert(message *tgbotapi.Message) {
	args := strings.Fields(message.CommandArguments())
	if len(args) < 2 {
		t.sendMessage(message.Chat.ID, "Formato: /update_alert <id> <threshold_price>")
		return
	}

	// Estrai ID e nuovo prezzo
	id, err := strconv.Atoi(args[0])
	if err != nil || id <= 0 {
		t.sendMessage(message.Chat.ID, "ID non valido. Usa un numero intero positivo.")
		return
	}

	thresholdPrice, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		t.sendMessage(message.Chat.ID, "Prezzo non valido. Usa un numero decimale.")
		return
	}

	// Verifica che l'alert esista e appartenga all'utente corrente
	var alert models.Alert
	if err := t.db.Where("id = ? AND user_chat_id = ?", id, message.Chat.ID).First(&alert).Error; err != nil {
		t.sendMessage(message.Chat.ID, "Alert non trovato o non hai i permessi per modificarlo.")
		return
	}

	// Aggiorna i campi dell'alert
	alert.ThresholdPrice = thresholdPrice

	// Aggiorna anche il prezzo corrente e reimposta lo stato triggered se necessario
	wasTriggered := alert.Triggered
	resetTrigger := false

	// Se viene specificato un terzo argomento "reset", reimposta lo stato triggered
	if len(args) > 2 && args[2] == "reset" {
		resetTrigger = true
	}

	// Aggiorna il prezzo corrente
	price, err := controllers.GetPriceUSD(alert.CryptoID)
	if err != nil {
		t.sendMessage(message.Chat.ID, fmt.Sprintf("Avviso: Impossibile ottenere il prezzo aggiornato: %v", err))
	} else {
		alert.CurrentPrice = price
	}

	// Se il prezzo attuale è sotto la soglia o è stato richiesto un reset, reimposta lo stato
	if resetTrigger || alert.CurrentPrice < alert.ThresholdPrice {
		alert.Triggered = false
	}

	alert.UpdatedAt = time.Now().UTC()

	// Salva le modifiche
	if err := t.db.Save(&alert).Error; err != nil {
		t.sendMessage(message.Chat.ID, fmt.Sprintf("Errore nell'aggiornamento dell'alert: %v", err))
		return
	}

	// Prepara il messaggio di risposta
	statusChange := ""
	if wasTriggered && !alert.Triggered {
		statusChange = "\n⚠️ Lo stato è stato reimpostato da triggerato a attivo!"
	}

	status := "⏳ In attesa"
	if alert.Triggered {
		status = "✅ Triggerato"
	}

	t.sendMessage(message.Chat.ID, fmt.Sprintf("✅ Alert aggiornato! ID: %d\nCrypto: %s\nNuova soglia: $%.2f\nPrezzo attuale: $%.2f\nStato: %s%s",
		alert.ID, alert.CryptoID, alert.ThresholdPrice, alert.CurrentPrice, status, statusChange))
}

// handleGetAlerts gestisce il comando /alerts
func (t *TelegramBot) handleGetAlerts(message *tgbotapi.Message) {
	var alerts []models.Alert

	// Filtra gli alert per l'ID della chat dell'utente corrente
	if err := t.db.Where("user_chat_id = ?", message.Chat.ID).Find(&alerts).Error; err != nil {
		t.sendMessage(message.Chat.ID, fmt.Sprintf("Errore nel recupero degli alert: %v", err))
		return
	}

	if len(alerts) == 0 {
		t.sendMessage(message.Chat.ID, "Non hai alert salvati.")
		return
	}

	var response strings.Builder
	response.WriteString("📊 I tuoi alert:\n\n")

	for _, alert := range alerts {
		status := "⏳ In attesa"
		triggerInfo := ""

		if alert.Triggered {
			status = "✅ Triggerato"
			if alert.NotifiedAt != nil {
				triggerInfo = fmt.Sprintf("Triggerato il: %s (CET)\n", alert.NotifiedAt.In(italianTimezone).Format("02/01/2006 15:04"))
			}
		}

		createdAt := fmt.Sprintf("Creato il: %s (CET)\n", alert.CreatedAt.In(italianTimezone).Format("02/01/2006 15:04"))

		response.WriteString(fmt.Sprintf("ID: %d | %s\nCrypto: %s\nSoglia: $%.2f\nPrezzo attuale: $%.2f\n%s%s\n",
			alert.ID, status, alert.CryptoID, alert.ThresholdPrice, alert.CurrentPrice, createdAt, triggerInfo))
	}

	t.sendMessage(message.Chat.ID, response.String())
}

// handleGetActiveAlerts gestisce il comando /active_alerts
func (t *TelegramBot) handleGetActiveAlerts(message *tgbotapi.Message) {
	var alerts []models.Alert

	// Filtra gli alert per l'ID della chat dell'utente corrente e non triggerati
	if err := t.db.Where("user_chat_id = ? AND triggered = ?", message.Chat.ID, false).Find(&alerts).Error; err != nil {
		t.sendMessage(message.Chat.ID, fmt.Sprintf("Errore nel recupero degli alert attivi: %v", err))
		return
	}

	if len(alerts) == 0 {
		t.sendMessage(message.Chat.ID, "Non hai alert attivi.")
		return
	}

	var response strings.Builder
	response.WriteString("⚡ I tuoi alert attivi:\n\n")

	for _, alert := range alerts {
		createdAt := fmt.Sprintf("Creato il: %s (CET)", alert.CreatedAt.In(italianTimezone).Format("02/01/2006 15:04"))

		response.WriteString(fmt.Sprintf("ID: %d\nCrypto: %s\nSoglia: $%.2f\nPrezzo attuale: $%.2f\n%s\n\n",
			alert.ID, alert.CryptoID, alert.ThresholdPrice, alert.CurrentPrice, createdAt))
	}

	t.sendMessage(message.Chat.ID, response.String())
}

// handleGetAlert gestisce il comando /alert
func (t *TelegramBot) handleGetAlert(message *tgbotapi.Message) {
	args := strings.Fields(message.CommandArguments())
	if len(args) < 1 {
		t.sendMessage(message.Chat.ID, "Specifica l'ID dell'alert. Esempio: /alert 1")
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil || id <= 0 {
		t.sendMessage(message.Chat.ID, "ID non valido. Usa un numero intero positivo.")
		return
	}

	var alert models.Alert
	// Filtra per id dell'alert e utente corrente
	if err := t.db.Where("id = ? AND user_chat_id = ?", id, message.Chat.ID).First(&alert).Error; err != nil {
		t.sendMessage(message.Chat.ID, "Alert non trovato o non hai i permessi per visualizzarlo.")
		return
	}

	status := "⏳ In attesa"
	if alert.Triggered {
		status = "✅ Triggerato"
	}

	createdAt := fmt.Sprintf("Creato il: %s (CET)", alert.CreatedAt.In(italianTimezone).Format("02/01/2006 15:04"))

	response := fmt.Sprintf("🔔 Alert #%d\n\nCrypto: %s\nSoglia: $%.2f\nPrezzo attuale: $%.2f\nStato: %s\n%s",
		alert.ID, alert.CryptoID, alert.ThresholdPrice, alert.CurrentPrice, status, createdAt)

	if alert.Triggered && alert.NotifiedAt != nil {
		response += fmt.Sprintf("\nTriggerato il: %s (CET)", alert.NotifiedAt.In(italianTimezone).Format("02/01/2006 15:04"))
	}

	t.sendMessage(message.Chat.ID, response)
}

// handleDeleteAlert gestisce il comando /delete_alert
func (t *TelegramBot) handleDeleteAlert(message *tgbotapi.Message) {
	args := strings.Fields(message.CommandArguments())
	if len(args) < 1 {
		t.sendMessage(message.Chat.ID, "Specifica l'ID dell'alert. Esempio: /delete_alert 1")
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil || id <= 0 {
		t.sendMessage(message.Chat.ID, "ID non valido. Usa un numero intero positivo.")
		return
	}

	// Prima verifica che l'alert esista e appartenga all'utente corrente
	var alert models.Alert
	if err := t.db.Where("id = ? AND user_chat_id = ?", id, message.Chat.ID).First(&alert).Error; err != nil {
		t.sendMessage(message.Chat.ID, "Alert non trovato o non hai i permessi per eliminarlo.")
		return
	}

	// Procedi con l'eliminazione
	if err := t.db.Delete(&alert).Error; err != nil {
		t.sendMessage(message.Chat.ID, fmt.Sprintf("Errore nella cancellazione: %v", err))
		return
	}

	t.sendMessage(message.Chat.ID, fmt.Sprintf("🗑️ Alert #%d eliminato con successo.", id))
}

// sendMessage invia un messaggio a una chat
func (t *TelegramBot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := t.bot.Send(msg)
	if err != nil {
		log.Printf("[Telegram] Errore nell'invio del messaggio: %v", err)
	}
}

// GetNotificationChannel restituisce un canale per ricevere notifiche da inviare tramite Telegram
func (t *TelegramBot) GetNotificationChannel() chan *models.Alert {
	notifyCh := make(chan *models.Alert, 100)

	// Goroutine che ascolta le notifiche e le invia
	go func() {
		for alert := range notifyCh {
			t.sendAlertNotification(alert)
		}
	}()

	return notifyCh
}

// sendAlertNotification invia una notifica quando un alert viene triggerato
func (t *TelegramBot) sendAlertNotification(alert *models.Alert) {
	// Invia la notifica solo all'utente che ha creato l'alert
	chatID := alert.UserChatID

	// Formatta l'ora corrente in timezone italiano per visualizzazione
	currentTime := time.Now().In(italianTimezone)

	message := fmt.Sprintf("🚨 ALERT TRIGGERATO! 🚨\n\nID: %d\nCrypto: %s\nSoglia: $%.2f\nPrezzo attuale: $%.2f\nData: %s (CET)",
		alert.ID, alert.CryptoID, alert.ThresholdPrice, alert.CurrentPrice, currentTime.Format("02/01/2006 15:04"))

	log.Printf("[Telegram] Invio notifica di alert triggerato all'utente %d", chatID)
	t.sendMessage(chatID, message)
}
