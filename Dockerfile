# Usa una base Go minimale
FROM golang:1.24.1-alpine

# Crea directory di lavoro
WORKDIR /app

# Copia i moduli Go e scaricali
COPY go.mod go.sum ./
RUN go mod download

# Copia tutto il progetto
COPY . .

# Compila il progetto
RUN go build -o main .

# Espone la porta (assumendo che l’API ascolti su :8080)
EXPOSE 8080

# Comando di avvio
CMD ["./main"]
