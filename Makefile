# Load environment variables
-include .env
export

.PHONY: start stop \
        dev dev-log dev-infra dev-backend dev-frontend dev-stop dev-wait \
        run run-stop ngrok-start ngrok-update-env \
        up down build logs \
        migrate-up migrate-down db-shell db-reset \
        seed test lint swagger clean help

# ──────────────────────────────────────────────
# Quick Start (just run this)
# ──────────────────────────────────────────────

## One command to rule them all: checks deps, starts infra, backend, frontend
start: env-file start-check-deps dev-infra dev-wait dev-backend dev-frontend
	@echo ""
	@echo "========================================"
	@echo "  phone-call-receptionist is running!"
	@echo ""
	@echo "  Frontend : http://localhost:3000"
	@echo "  Backend  : http://localhost:$${PORT:-8082}"
	@echo "  Health   : http://localhost:$${PORT:-8082}/api/health"
	@echo "  Swagger  : http://localhost:$${PORT:-8082}/swagger/index.html"
	@echo "========================================"
	@echo ""
	@echo "To stop:  make stop"
	@echo "Logs:     make dev-log"
	@echo ""

## Stop everything
stop: dev-stop

## Check that Docker, Go, and Node are available
start-check-deps:
	@echo ""
	@echo "Checking dependencies..."
	@command -v docker > /dev/null 2>&1 || (echo "ERROR: Docker is not installed. Install it from https://docker.com" && exit 1)
	@docker info > /dev/null 2>&1 || (echo "ERROR: Docker is not running. Start Docker Desktop first." && exit 1)
	@command -v go > /dev/null 2>&1 || (echo "ERROR: Go is not installed. Install it from https://go.dev" && exit 1)
	@command -v node > /dev/null 2>&1 || (echo "ERROR: Node.js is not installed. Install it from https://nodejs.org" && exit 1)
	@echo "  Docker  OK"
	@echo "  Go      OK ($$(go version | cut -d' ' -f3))"
	@echo "  Node    OK ($$(node -v))"
	@test -d frontend/node_modules || (echo "Installing frontend dependencies..." && cd frontend && npm install)
	@echo ""

# ──────────────────────────────────────────────
# Full Stack with ngrok (for real Twilio webhooks)
# ──────────────────────────────────────────────

NGROK_PORT ?= 8082

## Start everything including ngrok (single command for real Twilio calls)
run: env-file dev-infra dev-wait ngrok-start ngrok-update-env dev-backend dev-frontend
	@echo ""
	@echo "========================================"
	@echo "  phone-call-receptionist is running with ngrok!"
	@echo "  Frontend : http://localhost:3000"
	@echo "  Backend  : http://localhost:$${PORT:-8082}"
	@echo "  ngrok    : $$(cat /tmp/pcr-ngrok-url.txt 2>/dev/null)"
	@echo "  ngrok UI : http://127.0.0.1:4040"
	@echo "========================================"
	@echo ""
	@echo "Twilio webhooks are routed through ngrok."
	@echo "Press Ctrl+C or run 'make run-stop' to stop."

## Stop everything (ngrok + backend + frontend + infra)
run-stop: dev-stop
	@-kill $$(cat /tmp/pcr-ngrok.pid 2>/dev/null) 2>/dev/null; rm -f /tmp/pcr-ngrok.pid /tmp/pcr-ngrok-url.txt
	@echo "ngrok stopped."

## Start ngrok in background and save the public URL
ngrok-start:
	@echo "Starting ngrok on port $(NGROK_PORT)..."
	@ngrok http $(NGROK_PORT) --log=stdout --log-format=json > /tmp/pcr-ngrok.log 2>&1 & echo $$! > /tmp/pcr-ngrok.pid
	@sleep 3
	@NGROK_URL=$$(curl -s http://127.0.0.1:4040/api/tunnels | grep -o '"public_url":"https://[^"]*"' | head -1 | cut -d'"' -f4) && \
		if [ -z "$$NGROK_URL" ]; then echo "ERROR: ngrok failed to start. Is it installed? Run: brew install ngrok"; exit 1; fi && \
		echo "$$NGROK_URL" > /tmp/pcr-ngrok-url.txt && \
		echo "ngrok tunnel: $$NGROK_URL"

## Update .env and Twilio webhooks with the current ngrok URL
ngrok-update-env:
	@NGROK_URL=$$(cat /tmp/pcr-ngrok-url.txt 2>/dev/null) && \
		if [ -z "$$NGROK_URL" ]; then echo "ERROR: ngrok URL not found"; exit 1; fi && \
		sed -i '' "s|^TWILIO_WEBHOOK_BASE_URL=.*|TWILIO_WEBHOOK_BASE_URL=$$NGROK_URL|" .env && \
		cp .env backend/.env 2>/dev/null || true && \
		echo ".env updated: TWILIO_WEBHOOK_BASE_URL=$$NGROK_URL"
	@NGROK_URL=$$(cat /tmp/pcr-ngrok-url.txt 2>/dev/null) && \
		TWILIO_SID=$$(grep '^TWILIO_ACCOUNT_SID=' backend/.env | cut -d= -f2) && \
		TWILIO_TOKEN=$$(grep '^TWILIO_AUTH_TOKEN=' backend/.env | cut -d= -f2) && \
		TWILIO_PHONE=$$(grep '^TWILIO_PHONE_NUMBER=' backend/.env | cut -d= -f2) && \
		if [ -z "$$TWILIO_SID" ] || [ -z "$$TWILIO_TOKEN" ] || [ -z "$$TWILIO_PHONE" ]; then \
			echo "Twilio credentials not found in .env, skipping webhook update"; \
		else \
			PHONE_SID=$$(curl -s -u "$$TWILIO_SID:$$TWILIO_TOKEN" \
				"https://api.twilio.com/2010-04-01/Accounts/$$TWILIO_SID/IncomingPhoneNumbers.json?PhoneNumber=$$(echo $$TWILIO_PHONE | sed 's/+/%2B/')" \
				| python3 -c "import sys,json; print(json.load(sys.stdin)['incoming_phone_numbers'][0]['sid'])" 2>/dev/null) && \
			if [ -z "$$PHONE_SID" ]; then echo "WARNING: Could not find Twilio phone number SID"; else \
				curl -s -X POST \
					"https://api.twilio.com/2010-04-01/Accounts/$$TWILIO_SID/IncomingPhoneNumbers/$$PHONE_SID.json" \
					-u "$$TWILIO_SID:$$TWILIO_TOKEN" \
					-d "VoiceUrl=$$NGROK_URL/api/webhooks/twilio/voice" \
					-d "VoiceMethod=POST" \
					-d "StatusCallback=$$NGROK_URL/api/webhooks/twilio/status" \
					-d "StatusCallbackMethod=POST" > /dev/null && \
				echo "Twilio webhooks updated: $$NGROK_URL/api/webhooks/twilio/voice"; \
			fi; \
		fi

# ──────────────────────────────────────────────
# Local Development (without ngrok)
# ──────────────────────────────────────────────

## Start everything locally: infra (Docker) + backend + frontend
dev: env-file dev-infra dev-wait dev-backend dev-frontend
	@echo ""
	@echo "========================================"
	@echo "  phone-call-receptionist is running!"
	@echo "  Frontend : http://localhost:3000"
	@echo "  Backend  : http://localhost:$${PORT:-8082}"
	@echo "  Health   : http://localhost:$${PORT:-8082}/api/health"
	@echo "  Swagger  : http://localhost:$${PORT:-8082}/swagger/index.html"
	@echo "========================================"
	@echo ""
	@echo "Press Ctrl+C or run 'make dev-stop' to stop."

## Start everything with all logs in foreground (Ctrl+C to stop)
dev-log: env-file dev-infra dev-wait
	@echo ""
	@echo "========================================"
	@echo "  phone-call-receptionist starting with logs..."
	@echo "  Frontend : http://localhost:3000"
	@echo "  Backend  : http://localhost:$${PORT:-8082}"
	@echo "  Swagger  : http://localhost:$${PORT:-8082}/swagger/index.html"
	@echo "========================================"
	@echo ""
	@trap 'kill 0; make dev-stop' EXIT; \
		(cd backend && go run ./cmd/server 2>&1 | sed 's/^/[backend]  /') & \
		(cd frontend && PORT=3000 npm run dev 2>&1 | sed 's/^/[frontend] /') & \
		wait

## Create .env from .env.example if it doesn't exist
env-file:
	@test -f .env || (cp .env.example .env && echo "Created .env from .env.example")

## Start Postgres + Redis + Weaviate in Docker, wait healthy, run migrations
dev-infra:
	@echo "Starting Postgres, Redis, and Weaviate..."
	@docker compose up -d postgres redis weaviate
	@echo "Waiting for Postgres to be healthy..."
	@until docker compose exec -T postgres pg_isready -U $${DB_USER:-postgres} > /dev/null 2>&1; do sleep 1; done
	@echo "Postgres is ready."
	@echo "Running migrations..."
	@cd backend && migrate -path migrations -database "postgres://$${DB_USER:-postgres}:$${DB_PASSWORD:-postgres}@localhost:$${DB_PORT:-5433}/$${DB_NAME:-voice_ai}?sslmode=$${DB_SSL_MODE:-disable}" up 2>&1 || true
	@echo "Migrations done."

## Wait for infra to be ready (used internally)
dev-wait:
	@until docker compose exec -T postgres pg_isready -U $${DB_USER:-postgres} > /dev/null 2>&1; do sleep 1; done

## Start backend in background
dev-backend:
	@echo "Starting backend..."
	@-lsof -i :$${PORT:-8082} -sTCP:LISTEN -t 2>/dev/null | xargs kill 2>/dev/null || true
	@cd backend && go run ./cmd/server & echo $$! > /tmp/pcr-backend.pid
	@sleep 2
	@echo "Backend started (PID $$(cat /tmp/pcr-backend.pid))."

## Start frontend in background
dev-frontend:
	@echo "Starting frontend..."
	@cd frontend && PORT=3000 npm run dev & echo $$! > /tmp/pcr-frontend.pid
	@sleep 2
	@echo "Frontend started (PID $$(cat /tmp/pcr-frontend.pid))."

## Stop local dev processes and infra
dev-stop:
	@echo "Stopping dev processes..."
	@-kill $$(cat /tmp/pcr-ngrok.pid 2>/dev/null) 2>/dev/null; rm -f /tmp/pcr-ngrok.pid /tmp/pcr-ngrok-url.txt
	@-kill $$(cat /tmp/pcr-backend.pid 2>/dev/null) 2>/dev/null; rm -f /tmp/pcr-backend.pid
	@-kill $$(cat /tmp/pcr-frontend.pid 2>/dev/null) 2>/dev/null; rm -f /tmp/pcr-frontend.pid
	@-pkill -f "go run ./cmd/server" 2>/dev/null || true
	@-pkill -f "nuxt" 2>/dev/null || true
	@-pkill -f "ngrok" 2>/dev/null || true
	@-lsof -i :$${PORT:-8082} -sTCP:LISTEN -t 2>/dev/null | xargs kill 2>/dev/null || true
	@-lsof -i :3000 -sTCP:LISTEN -t 2>/dev/null | xargs kill 2>/dev/null || true
	@docker compose down
	@echo "All stopped."

# ──────────────────────────────────────────────
# Docker Compose (full stack)
# ──────────────────────────────────────────────

## Start all services in Docker
up: env-file
	docker compose up -d --build

## Stop all services
down:
	docker compose down

## Rebuild all Docker images
build:
	docker compose build --no-cache

## Tail logs for all services
logs:
	docker compose logs -f

# ──────────────────────────────────────────────
# Database
# ──────────────────────────────────────────────

## Run pending migrations
migrate-up:
	cd backend && migrate -path migrations -database "postgres://$${DB_USER:-postgres}:$${DB_PASSWORD:-postgres}@localhost:$${DB_PORT:-5433}/$${DB_NAME:-voice_ai}?sslmode=$${DB_SSL_MODE:-disable}" up

## Rollback last migration
migrate-down:
	cd backend && migrate -path migrations -database "postgres://$${DB_USER:-postgres}:$${DB_PASSWORD:-postgres}@localhost:$${DB_PORT:-5433}/$${DB_NAME:-voice_ai}?sslmode=$${DB_SSL_MODE:-disable}" down 1

## Open psql shell
db-shell:
	docker compose exec postgres psql -U $${DB_USER:-postgres} -d $${DB_NAME:-voice_ai}

## Reset database (drop all, re-migrate)
db-reset:
	cd backend && migrate -path migrations -database "postgres://$${DB_USER:-postgres}:$${DB_PASSWORD:-postgres}@localhost:$${DB_PORT:-5433}/$${DB_NAME:-voice_ai}?sslmode=$${DB_SSL_MODE:-disable}" drop -f
	cd backend && migrate -path migrations -database "postgres://$${DB_USER:-postgres}:$${DB_PASSWORD:-postgres}@localhost:$${DB_PORT:-5433}/$${DB_NAME:-voice_ai}?sslmode=$${DB_SSL_MODE:-disable}" up

## Seed sample data (admin user + IT FAQ documents)
seed:
	cd backend && go run cmd/seed/main.go

# ──────────────────────────────────────────────
# Quality
# ──────────────────────────────────────────────

## Run Go tests
test:
	cd backend && go test ./...

## Run go vet
lint:
	cd backend && go vet ./...

## Generate Swagger docs
swagger:
	cd backend && swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

# ──────────────────────────────────────────────
# Cleanup
# ──────────────────────────────────────────────

## Remove all containers, volumes, and temp files
clean:
	docker compose down -v
	rm -f /tmp/pcr-backend.pid /tmp/pcr-frontend.pid /tmp/pcr-ngrok.pid /tmp/pcr-ngrok-url.txt

# ──────────────────────────────────────────────
# Help
# ──────────────────────────────────────────────

## Show available commands
help:
	@echo ""
	@echo "phone-call-receptionist - Voice AI Receptionist"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Quick Start:"
	@echo "  start          Run everything (checks deps, starts DB, backend, frontend)"
	@echo "  stop           Stop everything"
	@echo ""
	@echo "Development:"
	@echo "  dev            Start locally (infra + backend + frontend)"
	@echo "  dev-log        Start with logs in foreground (Ctrl+C to stop)"
	@echo "  run            Start everything + ngrok (for real Twilio calls)"
	@echo "  run-stop       Stop everything (ngrok + backend + frontend + infra)"
	@echo "  seed           Create sample data (admin user + knowledge docs)"
	@echo ""
	@echo "Docker:"
	@echo "  up             Start all services in Docker"
	@echo "  down           Stop all services"
	@echo "  build          Rebuild Docker images"
	@echo "  logs           Tail service logs"
	@echo ""
	@echo "Database:"
	@echo "  migrate-up     Run pending migrations"
	@echo "  migrate-down   Rollback last migration"
	@echo "  db-shell       Open psql shell"
	@echo "  db-reset       Drop and re-create database"
	@echo ""
	@echo "Quality:"
	@echo "  test           Run Go tests"
	@echo "  lint           Run go vet"
	@echo "  swagger        Generate Swagger docs"
	@echo ""
	@echo "Cleanup:"
	@echo "  clean          Remove containers, volumes, temp files"
	@echo ""
