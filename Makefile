.PHONY: help dev infra-up infra-down stop health db-shell db-seed logs

help: ## Show all commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Start everything with docker-compose
	docker-compose up --build -d

infra-up: ## Start postgres + redis + weaviate
	docker-compose up -d postgres redis weaviate

infra-down: ## Stop infrastructure services
	docker-compose stop postgres redis weaviate

stop: ## Stop all services
	docker-compose down

health: ## Check all services health
	@echo "PostgreSQL:" && docker-compose exec postgres pg_isready -U postgres 2>/dev/null || echo "  DOWN"
	@echo "Redis:" && docker-compose exec redis redis-cli ping 2>/dev/null || echo "  DOWN"
	@echo "Weaviate:" && curl -sf http://localhost:8081/v1/.well-known/ready > /dev/null && echo "  READY" || echo "  DOWN"
	@echo "Backend:" && curl -sf http://localhost:8080/api/health > /dev/null && echo "  HEALTHY" || echo "  DOWN"
	@echo "Frontend:" && curl -sf http://localhost:3000 > /dev/null && echo "  UP" || echo "  DOWN"

db-shell: ## Open PostgreSQL shell
	docker-compose exec postgres psql -U postgres -d voice_ai

db-seed: ## Seed test data
	cd backend && go run cmd/seed/main.go

logs: ## Tail all logs
	docker-compose logs -f
