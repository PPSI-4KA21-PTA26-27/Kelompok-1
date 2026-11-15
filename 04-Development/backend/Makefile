# Perintah pengembangan backend. Jalankan dari direktori backend.
# Redis & PostgreSQL memakai yang terpasang di laptop (bukan Docker).
#   make test        semua test (test Redis di-skip bila Redis belum jalan)
#   make test-cache  test cache Redis saja, verbose
#   make check       fmt + vet + test

# Berurutan agar tidak berebut CPU: satu paket dan satu tes pada satu waktu.
GOFLAGS_SEQ = -p 1 -parallel 1

REDIS_TEST_ADDR ?= localhost:6379
export REDIS_TEST_ADDR

.PHONY: test test-v test-race test-cache vet fmt check sec backup

test:
	go test $(GOFLAGS_SEQ) ./... -count=1

test-v:
	go test $(GOFLAGS_SEQ) ./... -count=1 -v

test-race:
	go test $(GOFLAGS_SEQ) ./... -count=1 -race

test-cache:
	go test $(GOFLAGS_SEQ) ./internal/app/api/middleware -run ResponseCache -count=1 -v

vet:
	go vet ./...

fmt:
	gofmt -w cmd internal

check: fmt vet test

# Keamanan: govulncheck memblokir (CVE terpanggil); gosec hanya lapor (temuan lama: SHA1 Cloudinary, CLI migrasi, cast config), blokir bila sudah dibersihkan.
sec:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	go run github.com/securego/gosec/v2/cmd/gosec@latest -no-fail -quiet -severity medium -confidence medium ./...

# Cadangan DB: DATABASE_URL=... make backup  (hasil: backup-<tanggal>.dump, uji restore tiap bulan)
backup:
	sh scripts/backup.sh
