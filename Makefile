-include .env
export

BINARY := bin/api

.PHONY: run build test vet fmt check up down migrate-up migrate-down migrate-status

run:
	go run ./cmd/api

build:
	go build -o $(BINARY) ./cmd/api

test:
	go test ./... -race -count=1

vet:
	go vet ./...

fmt:
	gofmt -w .

check: fmt vet test

up:
	docker compose up -d

down:
	docker compose down

MIGRATIONS := migrations

migrate-up:
	goose -dir $(MIGRATIONS) postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir $(MIGRATIONS) postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir $(MIGRATIONS) postgres "$(DATABASE_URL)" status
