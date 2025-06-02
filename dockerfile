# ── build stage ────────────────────────────────────────────────────
FROM golang:1.24.3-alpine AS builder
WORKDIR /src

# Descargamos dependencias primero (cache)
COPY go.mod go.sum ./
RUN go mod download

# Copiamos el resto y compilamos binario estático
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags "-s -w" -o /url-shortener ./main.go

# ── runtime stage ──────────────────────────────────────────────────
FROM alpine:3.20
RUN adduser -D -g '' app
USER app

COPY --from=builder /url-shortener /usr/local/bin/url-shortener
EXPOSE 8080
ENTRYPOINT ["url-shortener"]
