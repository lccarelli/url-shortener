## ——— Configuración ————————————————————————————————————————
COMPOSE     = docker compose -f docker-compose.yml
K6_IMG      = grafana/k6:0.50.0
SCRIPT_DIR  = $(PWD)/loadtest
RESULTS_DIR = $(PWD)/results

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

## ——— Tests Go ———————————————————————————————————————
.PHONY: test
test:
	go test ./...

## ——— Smoke test ————————————————————————————————
APP_URL      ?= http://localhost
SMOKE_TARGET ?= https://www.mercadolibre.com

.PHONY: smoke
smoke:
	@echo "🏃 Smoke test…"
	@set -e; \
	r=$$(curl -s -w '\n%{http_code}' -X POST $(APP_URL)/shorten \
	       -H 'Content-Type: application/json' \
	       -d '{"url":"$(SMOKE_TARGET)"}'); \
	body=$$(echo "$$r"|head -n1); code=$$(echo "$$r"|tail -n1); \
	[ "$$code" -eq 200 ] || { echo "POST fail $$code"; exit 1; }; \
	key=$$(echo "$$body"|jq -r .short); \
	echo " short = $$key"; \
	code2=$$(curl -s -o /dev/null -w '%{http_code}' $(APP_URL)/$$key); \
	[ "$$code2" -eq 302 ] && echo "✅ ok" || { echo "redirect fail $$code2"; exit 1; }

## ——— Load tests sin balanceador —————————————————————
.PHONY: load-smoke load-10k load-20k load-50k

load-smoke:
	$(COMPOSE) run --rm --no-deps \
		-v $(SCRIPT_DIR):/scripts \
		k6 run /scripts/smoke_load.js

load-10k:
	$(COMPOSE) run --rm --no-deps \
		-e K6_CONNECTION_REUSE=true \
		-e K6_HTTP_KEEPALIVE=true \
		-e K6_CONNS=10000 \
		-v $(SCRIPT_DIR):/scripts \
		-v $(RESULTS_DIR):/results \
		k6 run /scripts/shortener_10k.js

load-20k:
	$(COMPOSE) run --rm --no-deps \
		-e K6_CONNECTION_REUSE=true \
		-e K6_HTTP_KEEPALIVE=true \
		-e K6_CONNS=20000 \
		-v $(SCRIPT_DIR):/scripts \
		-v $(RESULTS_DIR):/results \
		k6 run /scripts/shortener_20k.js

load-50k:
	$(COMPOSE) run --rm --no-deps \
		-e K6_CONNECTION_REUSE=true \
		-e K6_HTTP_KEEPALIVE=true \
		-e K6_CONNS=30000 \
		-v $(SCRIPT_DIR):/scripts \
		-v $(RESULTS_DIR):/results \
		k6 run /scripts/shortener_50k.js

## ——— Package Lambda ——————————————————————————————
.PHONY: package
package:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/shortener main.go
	zip -j -q bin/shortener.zip bin/shortener
