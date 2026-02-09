# Makefile for id-100 development and Docker operations
# Usage: make <target>

.PHONY: help run build build-all build-frontend test fmt vet clean
.PHONY: docker-build docker-up docker-down docker-restart docker-logs docker-clean

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
	@echo "  Docker Compose:"
	@echo "    docker-build     - Build Docker images"
	@echo "    docker-up        - Start all services"
	@echo "    docker-down      - Stop all services"
	@echo "    docker-restart   - Restart all services"
	@echo "    docker-logs      - View logs from all services"
	@echo "    docker-clean     - Stop services and remove volumes"
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

# Docker Compose targets
docker-build:
	docker-compose build --no-cache

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-restart:
	docker-compose restart

docker-logs:
	docker-compose logs -f

docker-clean:
	docker-compose down -v
	docker system prune -f

# Combined targets
rebuild: docker-down docker-build docker-up

# Run full stack with logs
dev: docker-up docker-logs
