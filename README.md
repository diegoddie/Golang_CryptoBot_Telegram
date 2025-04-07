# Crypto Tracker

Un'applicazione per il monitoraggio dei prezzi delle criptovalute e la creazione di alert quando vengono raggiunte determinate soglie di prezzo.

## Funzionalità

- Ottenere prezzi aggiornati delle criptovalute tramite l'API di CoinGecko
- Creare alert per monitorare i prezzi
- API RESTful per gestire gli alert

## Setup

### Requisiti

- Go 1.15+ 
- MySQL o altro database supportato da GORM

### Installazione

1. Clona il repository:
```
git clone <repository-url>
cd crypto-tracker-golang
```

2. Installa le dipendenze:
```
go mod download
```

3. Configura le variabili d'ambiente:
```
# Configura la connessione al database
export DB_USER=root
export DB_PASSWORD=password
export DB_NAME=crypto_tracker
export DB_HOST=localhost
export DB_PORT=3306

# Configura la chiave API di CoinGecko (necessaria per la nuova versione dell'API)
export COINGECKO_API_KEY=your_api_key_here
```

Per ottenere una chiave API di CoinGecko:
1. Registrati su [CoinGecko](https://www.coingecko.com/)
2. Vai al [dashboard per sviluppatori](https://www.coingecko.com/en/api/pricing) per ottenere una chiave API gratuita o a pagamento
3. Imposta la variabile d'ambiente `COINGECKO_API_KEY` con la tua chiave API

4. Avvia l'applicazione:
```
go run main.go
```

L'applicazione sarà disponibile su http://localhost:8080

## API Endpoints

### Prezzi delle criptovalute

- `GET /price/:id` - Ottiene il prezzo attuale di una criptovaluta tramite il suo ID CoinGecko (es. "bitcoin", "ethereum")

### Alert

- `GET /alerts` - Ottiene tutti gli alert
- `GET /alerts/:id` - Ottiene un alert specifico
- `POST /alerts` - Crea un nuovo alert
  ```json
  {
    "crypto_id": "bitcoin",
    "threshold_price": 30000
  }
  ```
- `PUT /alerts/:id` - Aggiorna un alert esistente
  ```json
  {
    "threshold_price": 32000,
    "triggered": false
  }
  ```
- `DELETE /alerts/:id` - Elimina un alert

## Note sulla nuova API di CoinGecko

CoinGecko ha recentemente aggiornato la propria politica e ora richiede una chiave API per tutte le richieste. Se non imposti una chiave API valida, potresti riscontrare limitazioni di rate o risposte di errore.
