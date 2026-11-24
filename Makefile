-include .env
export

BINARY := bin/api

.PHONY: run build test vet fmt check up down

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
