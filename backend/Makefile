# Load environment variables
-include .env
export

.PHONY: help run build test swagger fmt lint migrate-up migrate-down db-reset

help: ## Show all commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## Run Go server
	go run cmd/server/main.go

build: ## Build binary
	CGO_ENABLED=0 go build -o bin/server cmd/server/main.go

test: ## Run tests
	go test ./... -v -cover

swagger: ## Regenerate Swagger docs
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

fmt: ## Format code
	gofmt -s -w .
	goimports -w .

lint: ## Run golangci-lint
	golangci-lint run ./...

migrate-up: ## Run pending migrations
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)" up

migrate-down: ## Rollback last migration
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)" down 1

db-reset: ## Full database reset (WARNING: deletes all data)
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)" drop -f
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)" up
