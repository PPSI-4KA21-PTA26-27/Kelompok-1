.PHONY: run build tidy test db-up db-down

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

tidy:
	go mod tidy

test:
	go test ./... -cover

db-up:
	docker compose up -d db

db-down:
	docker compose down
