# Makefile for id-100 development and Docker operations
# Usage: make <target>

.PHONY: help run build build-all build-frontend test fmt vet clean
.PHONY: docker-build docker-up docker-down docker-restart docker-logs docker-clean
.PHONY: docker-dev-up docker-dev-down docker-dev-restart docker-dev-logs docker-dev-clean docker-dev-rebuild

# Default target - show help
help:
	@echo "Available targets:"
	@echo "  Development:"
	@echo "    run              - Run the app locally (requires local Postgres)"
	@echo "    build            - Build the Go binary"
	@echo "    build-frontend   - Build TypeScript frontend"
	@echo "    build-all        - Build frontend + backend"
	@echo "    test             - Run Go tests"
	@echo "    fmt              - Format Go code"
	@echo "    vet              - Run Go vet"
	@echo "    clean            - Remove build artifacts"
	@echo ""
	@echo "  Docker Compose (Production):"
	@echo "    docker-build     - Build Docker images"
	@echo "    docker-up        - Start all services"
	@echo "    docker-down      - Stop all services"
	@echo "    docker-restart   - Restart all services"
	@echo "    docker-logs      - View logs from all services"
	@echo "    docker-clean     - Stop services and remove volumes"
	@echo ""
	@echo "  Docker Compose (Development):"
	@echo "    docker-dev-up    - Start all services for local development"
	@echo "    docker-dev-down  - Stop development services"
	@echo "    docker-dev-restart - Restart development services"
	@echo "    docker-dev-logs  - View logs from development services"
	@echo "    docker-dev-clean - Stop development services and remove volumes"
	@echo "    docker-dev-rebuild - Clean rebuild of development services"
	@echo ""

# Development targets
run:
	go run ./cmd/id-100

build:
	go build -o bin/id-100 ./cmd/id-100

build-frontend:
	npm run build

build-all: build-frontend build

test:
	go test ./...
	npm test

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -f bin/id-100
	rm -f web/static/main.js web/static/main.js.map

# Docker Compose targets (Production)
docker-build:
	docker compose build --no-cache

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-restart:
	docker compose restart

docker-logs:
	docker compose logs -f

docker-clean:
	docker compose down -v
	docker system prune -f

# Docker Compose targets (Development)
docker-dev-up:
	docker compose -f docker-compose.dev.yml --env-file .env.dev up -d

docker-dev-down:
	docker compose -f docker-compose.dev.yml down

docker-dev-restart:
	docker compose -f docker-compose.dev.yml restart

docker-dev-logs:
	docker compose -f docker-compose.dev.yml logs -f

docker-dev-clean:
	docker compose -f docker-compose.dev.yml down -v
	docker system prune -f

docker-dev-rebuild:
	docker compose -f docker-compose.dev.yml --env-file .env.dev down
	docker compose -f docker-compose.dev.yml --env-file .env.dev build --no-cache
	docker compose -f docker-compose.dev.yml --env-file .env.dev up -d

# Combined targets
rebuild: docker-down docker-build docker-up

# Run full stack with logs
dev: docker-up docker-logs
