.PHONY: run build tidy test test-race test-integration test-db-up test-db-down lint vuln security check db-up db-down migrate-up migrate-down migrate-version

TEST_DATABASE_URL ?= postgres://hii_test:hii_test_pw@127.0.0.1:55432/hii_test?sslmode=disable

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api
	go build -o bin/migrate ./cmd/migrate

tidy:
	go mod tidy

# Unit test (integration test dilewati bila TEST_DATABASE_URL kosong)
test:
	go test ./... -count=1 -cover

test-race:
	go test ./... -count=1 -race

# PostgreSQL sementara khusus test (tanpa volume, port 55432)
test-db-up:
	docker run -d --name hii_stranger_test_db -e POSTGRES_USER=hii_test -e POSTGRES_PASSWORD=hii_test_pw -e POSTGRES_DB=hii_test -p 127.0.0.1:55432:5432 --tmpfs /var/lib/postgresql/data postgres:16-alpine

test-db-down:
	docker rm -f hii_stranger_test_db

test-integration:
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test ./... -count=1 -cover

# Migrasi memakai DATABASE_URL dari environment / .env
migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down $(N)

migrate-version:
	go run ./cmd/migrate version

# Install tool scanning (sekali saja)
tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

lint:
	golangci-lint run ./...

vuln:
	govulncheck ./...

security: lint vuln

# Jalankan semua pengecekan sebelum commit
check: build
	go vet ./...
	go test ./... -count=1 -race
	$(MAKE) lint
	$(MAKE) vuln

db-up:
	docker compose up -d db

db-down:
	docker compose down
