# Architecture

## Alur Request

```
Client -> Router -> Middleware -> Handler -> Service -> Repository -> PostgreSQL
```

| Layer | Folder | Tugas |
|---|---|---|
| Router | `internal/router` | Daftar route dan wiring dependency |
| Middleware | `internal/middleware` | Validasi JWT, cek role, CORS |
| Handler | `internal/handler` | Baca request, validasi input, kirim response |
| Service | `internal/service` | Business logic (hash password, bikin token, aturan bisnis) |
| Repository | `internal/repository` | Query database lewat GORM |
| Model | `internal/models` | Struct tabel database |
| DTO | `internal/dto` | Bentuk request dan aturan validasi |
| Utils | `internal/utils` | Helper response, JWT, password |

## Aturan Antar Layer

- Handler hanya boleh panggil service, tidak boleh langsung ke repository atau DB
- Service hanya boleh panggil repository
- Repository dan service dibuat sebagai interface, jadi gampang di-mock waktu bikin unit test
- Error bisnis (misal `ErrEmailTaken`) didefinisikan di service, lalu handler yang menerjemahkannya ke status HTTP

## Dependency Wiring

Semua objek dirangkai manual di `internal/router/router.go`:

```
repository -> service -> handler
```

Tidak pakai framework DI supaya alurnya jelas dan gampang dibaca.

## Konfigurasi

Dibaca dari environment variable (`.env` kalau ada) lewat `internal/config`. Daftar variabelnya ada di `.env.example`.

## Database

- PostgreSQL, diakses lewat GORM
- Tabel dibuat otomatis lewat `AutoMigrate` di `cmd/api/main.go`
- Model baru wajib didaftarkan di `models.All()` supaya ikut ter-migrate
- Hapus data pakai soft delete (`deleted_at`)

## Menambah Fitur Baru

1. Buat model di `internal/models`, daftarkan di `models.All()`
2. Buat DTO di `internal/dto`
3. Buat repository di `internal/repository`
4. Buat service di `internal/service`
5. Buat handler di `internal/handler`
6. Rangkai dan daftarkan route di `internal/router/router.go`
7. Catat endpoint barunya di `docs/api.md`
