# Kel-1 Backend (Go)

Template backend generik: Gin + GORM + PostgreSQL + JWT. Fitur domain ditambah belakangan setelah topik project fix.

## Stack
- Go, Gin (HTTP), GORM (ORM), PostgreSQL
- JWT (golang-jwt) + bcrypt
- godotenv untuk konfigurasi, Docker + docker-compose

## Struktur
```
cmd/api/main.go          entry point + graceful shutdown
internal/config          load env
internal/database        koneksi DB
internal/models          struct tabel (daftarkan di models.All())
internal/dto             request/response payload + validasi
internal/repository      akses data (interface + implementasi GORM)
internal/service         business logic
internal/handler         HTTP handler
internal/middleware      auth JWT, role, CORS
internal/router          wiring dependency + definisi route
internal/utils           response helper, jwt, password
```
Alur: `handler -> service -> repository -> DB`.

## Mulai
```bash
cp .env.example .env        # sesuaikan isinya
make db-up                  # jalankan PostgreSQL via Docker
go mod tidy                 # download dependency
make run
```

## Endpoint
| Method | Path | Akses |
|---|---|---|
| GET | /health | public |
| POST | /api/v1/auth/register | public |
| POST | /api/v1/auth/login | public |
| GET | /api/v1/me | login |
| GET | /api/v1/users?page=1&limit=10 | admin |
| GET/PUT/DELETE | /api/v1/users/:id | admin |

Format response:
```json
{ "success": true, "message": "...", "data": {}, "meta": {}, "errors": null }
```

## Tambah fitur baru (contoh: Product)
1. Buat `internal/models/product.go`, daftarkan di `models.All()`
2. Tambah DTO di `internal/dto`
3. Buat `repository/product_repository.go` dan `service/product_service.go`
4. Buat `handler/product_handler.go`
5. Wiring + route di `internal/router/router.go`

## Catatan
- Register selalu membuat role `user`. Untuk bikin admin pertama, ubah kolom `role` di DB jadi `admin`.
- Ganti `JWT_SECRET` sebelum deploy.
- AutoMigrate cukup untuk tahap awal; kalau schema sudah rumit, pindah ke golang-migrate.
