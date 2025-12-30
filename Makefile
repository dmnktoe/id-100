# Simple Makefile for id-100 development
# Usage: make run | build | test | docker-db | docker-stop | clean

.PHONY: run build test fmt vet docker-db docker-stop clean

run:
	go run ./cmd/id-100

build:
	go build -o bin/id-100 ./cmd/id-100

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

# Run a local Postgres for development
docker-db:
	docker run --name id100-db -e POSTGRES_PASSWORD=pass -e POSTGRES_USER=dev -e POSTGRES_DB=id100 -p 5432:5432 -d postgres:15

# Stop and remove the dev Postgres
docker-stop:
	docker stop id100-db || true
	docker rm id100-db || true

clean:
	rm -f bin/id-100
