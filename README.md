# 📈 Crypto Tracker & Alert Bot in Go 🤖🔔

Benvenuto nel Crypto Tracker & Alert Bot! Questo progetto è un'applicazione Go completa che ti permette di monitorare i prezzi delle criptovalute, impostare alert personalizzati e ricevere notifiche direttamente sul tuo bot Telegram.

Puoi interagire con il bot direttamente su Telegram: [t.me/go_crypto_prices_bot](t.me/go_crypto_prices_bot)

Questo README ti guiderà attraverso l'architettura, le funzionalità e i passaggi necessari per configurare ed eseguire il progetto. È perfetto se vuoi capire come costruire un'applicazione backend in Go con integrazione API, database e un bot interattivo! 🚀

## ✨ Funzionalità Principali

*   **📊 Monitoraggio Prezzi**: Ottiene i prezzi aggiornati delle criptovalute dall'API di CoinGecko.
*   **🔔 Sistema di Alert**:
    *   Crea alert basati su soglie di prezzo (es. "avvisami quando Bitcoin supera i $70,000").
    *   API REST per la gestione programmatica degli alert.
    *   Comandi Telegram per creare, visualizzare, aggiornare ed eliminare alert.
*   **🤖 Bot Telegram Interattivo**:
    *   Ricevi notifiche istantanee quando un alert viene triggerato.
    *   Interroga il bot per il prezzo attuale di qualsiasi criptovaluta.
    *   Gestisci i tuoi alert direttamente dalla chat di Telegram.
*   **⚙️ Servizio Background**: Un monitoraggio continuo verifica gli alert attivi in background.
*   **💾 Database Persistente**: Utilizza GORM e PostgreSQL (configurato per NeonDB) per salvare gli alert degli utenti.
*   **🌐 API RESTful**: Espone endpoint per interagire con il sistema (protetti da CORS).
*   **🐳 Docker Ready**: Include un `Dockerfile` per containerizzare facilmente l'applicazione.

## 🛠️ Tecnologie Utilizzate

*   **Go**: Linguaggio di programmazione principale. [https://go.dev/](https://go.dev/)
*   **Gin**: Framework web veloce per la creazione dell'API REST. [https://gin-gonic.com/](https://gin-gonic.com/)
*   **GORM**: Fantastico ORM (Object-Relational Mapper) per interagire con il database. [https://gorm.io/](https://gorm.io/)
*   **PostgreSQL (NeonDB)**: Database relazionale serverless per memorizzare gli alert. [https://neon.tech/](https://neon.tech/)
*   **Go Telegram Bot API**: Libreria per interagire con l'API di Telegram. [https://github.com/go-telegram-bot-api/telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
*   **CoinGecko API**: API per ottenere i dati sui prezzi delle criptovalute. [https://www.coingecko.com/en/api](https://www.coingecko.com/en/api)
*   **Resty**: Client HTTP per Go, usato per le chiamate all'API CoinGecko. [https://github.com/go-resty/resty](https://github.com/go-resty/resty)
*   **Godotenv**: Libreria per caricare variabili d'ambiente da file `.env`. [https://github.com/joho/godotenv](https://github.com/joho/godotenv)
*   **Docker**: Piattaforma per containerizzare l'applicazione. [https://www.docker.com/](https://www.docker.com/)

## 🏗️ Architettura Generale

Il sistema è composto da diversi componenti che lavorano insieme:

1.  **API Server (Gin)**: Gestisce le richieste HTTP per creare/leggere/aggiornare/eliminare alert e ottenere prezzi. Interagisce direttamente con il database.
2.  **Database (PostgreSQL/NeonDB)**: Memorizza le informazioni sugli alert creati dagli utenti.
3.  **Telegram Bot (go-telegram-bot-api)**: Fornisce un'interfaccia utente tramite chat. Riceve comandi, interagisce con l'API server (logicamente, chiamando le stesse funzioni dei controller o funzioni dedicate) e invia notifiche.
4.  **Alert Monitor (Servizio Background)**: Un goroutine che periodicamente:
    *   Recupera tutti gli alert attivi dal database.
    *   Ottiene i prezzi correnti da CoinGecko.
    *   Verifica se qualche alert è stato triggerato.
    *   Aggiorna lo stato dell'alert nel database.
    *   Invia una notifica tramite il canale del Bot Telegram.
6.  **CoinGecko API**: Fonte esterna per i dati sui prezzi.

## 📋 Prerequisiti

Prima di iniziare, assicurati di avere installato:

*   **Go**: Versione 1.18 o successiva.
*   **Git**: Per clonare il repository.
*   **Docker** (Opzionale): Se vuoi eseguire l'applicazione in un container.
*   **Un account Telegram**: Per creare il tuo bot.
*   **Un account CoinGecko**: Per ottenere una chiave API (consigliato per evitare limiti di rate).
*   **Un database PostgreSQL**: Puoi usare un'istanza locale o un servizio cloud come [NeonDB](https://neon.tech/) (consigliato e gratuito).

## 🏁 Getting Started

Segui questi passaggi per configurare ed eseguire il progetto localmente.

### 1. Clona il Repository

```bash
git clone https://github.com/diegoddie/Golang_CryptoBot_Telegram
cd crypto-tracker-golang
```

### 2. Configura le Variabili d'Ambiente 🔑

Crea un file `.env` nella root del progetto. Inserisci i seguenti valori:

*   **`DATABASE_URL`**: La stringa di connessione al tuo database PostgreSQL. Se usi NeonDB, trovi la stringa nella tua dashboard.
*   **`COINGECKO_API_KEY`**: La tua chiave API di CoinGecko. Anche senza chiave funziona, ma potresti incorrere in limiti di utilizzo più restrittivi.
*   **`TELEGRAM_BOT_TOKEN`**: Il token univoco del tuo bot Telegram. Creane uno parlando con `@BotFather` su Telegram e seguendo le istruzioni.

### 3. Configura il Database 💾

L'applicazione usa GORM per gestire le migrazioni del database. Quando avvii l'applicazione per la prima volta, GORM creerà automaticamente la tabella `alerts` se non esiste, basandosi sulla struttura definita in `models/alert.go`. Assicurati che la `DATABASE_URL` nel tuo file `.env` sia corretta e che il database sia accessibile.

### 4. Installa le Dipendenze

Apri il terminale nella directory del progetto ed esegui:

```bash
go mod tidy
```

Questo comando scaricherà tutte le librerie necessarie definite nel file `go.mod`.

### 5. Esegui l'Applicazione 🚀

Ora sei pronto per avviare il server!

```bash
go run main.go
```

## 🌐 API Endpoints

L'applicazione espone i seguenti endpoint API (base path: `http://localhost:8080`):

### Alert API (`/alerts`)

*   `POST /alerts`
    *   Crea un nuovo alert.
    *   **Body (JSON):** `{ "crypto_id": "bitcoin", "threshold_price": 65000, "user_chat_id": 12345678 }` (user\_chat\_id è opzionale per test API)
    *   **Risposta:** Dettagli dell'alert creato.
*   `GET /alerts`
    *   Ottiene tutti gli alert.
    *   **Query Params (Opzionale):** `user_chat_id=12345678` per filtrare per utente Telegram.
    *   **Risposta:** Lista di alert.
*   `GET /alerts/active`
    *   Ottiene solo gli alert attivi (non ancora triggerati).
    *   **Query Params (Opzionale):** `user_chat_id=12345678` per filtrare per utente Telegram.
    *   **Risposta:** Lista di alert attivi.
*   `GET /alerts/:id`
    *   Ottiene i dettagli di un alert specifico per ID.
    *   **Risposta:** Dettagli dell'alert.
*   `PUT /alerts/:id`
    *   Aggiorna un alert esistente (es. soglia, stato triggered).
    *   **Body (JSON):** `{ "threshold_price": 70000, "triggered": false }`
    *   **Risposta:** Dettagli dell'alert aggiornato.
*   `DELETE /alerts/:id`
    *   Elimina un alert specifico per ID.
    *   **Risposta:** Messaggio di conferma.

### Crypto Price API (`/crypto`)

*   `GET /crypto/price/:id`
    *   Ottiene il prezzo attuale per una criptovaluta specifica (usa ID CoinGecko, es. `bitcoin`).
    *   **Risposta:** `{ "id": "bitcoin", "price": 68123.45, "timestamp": "..." }`

*(Nota: Gli endpoint sono protetti da CORS, configurato in `main.go` per permettere richieste da specifici domini/localhost)*

## 🤖 Comandi del Bot Telegram

Puoi interagire con il bot direttamente su Telegram: [t.me/go_crypto_prices_bot](t.me/go_crypto_prices_bot)

Usa i seguenti comandi:

*   `/start` o `/help`: Mostra il messaggio di aiuto con la lista dei comandi.
*   `/price <crypto_id_o_simbolo>`: Mostra il prezzo attuale della criptovaluta specificata (es. `/price btc` o `/price bitcoin`).
*   `/create_alert <crypto_id_o_simbolo> <prezzo_soglia>`: Crea un nuovo alert per te (es. `/create_alert solana 150`).
*   `/alerts`: Mostra tutti gli alert che hai creato.
*   `/active_alerts`: Mostra solo i tuoi alert che non sono ancora stati triggerati.
*   `/alert <id>`: Mostra i dettagli di un tuo alert specifico usando il suo ID numerico (es. `/alert 5`).
*   `/update_alert <id> <nuovo_prezzo_soglia>`: Aggiorna la soglia di un tuo alert esistente (es. `/update_alert 5 160`).
*   `/update_alert <id> <nuovo_prezzo_soglia> reset`: Aggiorna la soglia e reimposta lo stato `triggered` a `false` (utile se vuoi riattivare un alert già scattato).
*   `/delete_alert <id>`: Elimina un tuo alert specifico (es. `/delete_alert 5`).

---

