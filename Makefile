## ——— Configuración ————————————————————————————————————————
COMPOSE     = docker compose -f docker-compose.yml

## ——— Build & Stack ————————————————————————————————
.PHONY: build up down restart logs clean

build:          ## Compila imágenes
	$(COMPOSE) build

up:             ## Levanta stack en segundo plano
	$(COMPOSE) up -d

down:           ## Derriba stack
	$(COMPOSE) down

restart: down up

logs:           ## Logs en vivo
	$(COMPOSE) logs -f --tail=80

clean:          ## Derriba y limpia recursos innecesarios
	make down
	docker system prune -f

## ——— Package Lambda ——————————————————————————————
.PHONY: package
package:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/shortener main.go
	zip -j -q bin/shortener.zip bin/shortener
