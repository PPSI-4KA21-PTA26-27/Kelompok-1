# Backend Hii Stranger

REST API toko online (Go + PostgreSQL). Kontrak lengkap endpoint ada di
[`api/openapi.yaml`](api/openapi.yaml); semua route berawalan `/api/v1`. Versi lengkap (`APP_VERSION`) juga dilayani di `/api/v<versi>`, dan versi lama yang masih didukung didaftarkan lewat `API_LEGACY_VERSIONS` (dipisah koma). Header `X-API-Version` menunjukkan versi aktif.

## Struktur folder

```text
backend/
├── api/openapi.yaml          Dokumentasi/kontrak API
├── cmd/
│   ├── api/                  Entry point server HTTP
│   └── migration/            Tool migrasi database (up, down, status, create)
├── docs/                     Catatan penulisan
├── internal/
│   ├── app/api/              Fitur API, satu package per fitur:
│   │   ├── auth, user        Register, login, profil, kelola user (admin)
│   │   ├── category, product Kategori, produk, varian, gambar produk
│   │   ├── banner            Banner beranda
│   │   ├── sitesetting       Mode maintenance
│   │   ├── ongkir            Wilayah (prov/kota/kec/kel) & cek ongkir RajaOngkir
│   │   ├── cart              Keranjang
│   │   ├── order             Checkout & status order
│   │   ├── payment           Tagihan, bukti transfer, verifikasi admin
│   │   ├── shipment          Kurir & nomor resi
│   │   ├── upload            Unggah gambar ke Cloudinary
│   │   ├── middleware        Auth, CORS, rate limit, logger, dll.
│   │   ├── routes.go         Daftar semua route
│   │   └── server.go         Merangkai repository -> service -> handler
│   └── pkg/                  Helper bersama: config, database, response,
│                             validator, token, password, mailer (SMTP),
│                             cloudinary, rajaongkir, fileupload
├── migrations/               File SQL migrasi
├── .env.example              Contoh konfigurasi
└── go.mod
```

Setiap package fitur memakai pola yang sama:

- `model.go` – struct data, konstanta status, dan error domain
- `repository.go` – query database
- `service.go` – aturan bisnis
- `handler.go` – baca request, validasi, tulis response

## Menjalankan

```bash
cp .env.example .env          # lalu isi nilainya
go run ./cmd/migration up     # buat tabel
go run ./cmd/api              # jalankan server di :8080
```

**Di belakang proxy (Railway/Cloudflare)** – set `TRUST_PROXY_HEADERS=true`
supaya rate limit memakai IP asli klien (`CF-Connecting-IP`, `X-Real-IP`,
lalu elemen pertama `X-Forwarded-For`). Default `false`: hanya RemoteAddr;
jangan aktifkan tanpa proxy karena header bisa dipalsukan klien. Flag ini juga
membuat `X-Forwarded-Proto: https` dipercaya untuk header HSTS.

**Batas beban** – rute terautentikasi dibatasi per user (cadangan per IP):
mutasi pelanggan 60/menit, GET 120/menit, upload 20/menit, admin 120/menit
(429 + `Retry-After`). `MAX_CONCURRENT_REQUESTS` (default 200, 0 = nonaktif)
membatasi request bersamaan; bila penuh dijawab 503 `server_busy`. Health
check dan `/metrics` tidak ikut dihitung.

## Layanan luar

**Email (Gmail SMTP)** – isi `SMTP_HOST=smtp.gmail.com`, `SMTP_PORT=587`,
`SMTP_USER`, dan `SMTP_PASS` dengan *App Password* Google (bukan password
akun). Email dikirim di background:

- ke `OWNER_EMAIL`: order baru dan bukti pembayaran baru
- ke pelanggan: hasil verifikasi pembayaran, perubahan status order, info resi

Bila `SMTP_HOST` kosong, email dilewati tanpa error.

**Cloudinary** – isi `CLOUDINARY_CLOUD_NAME`, `CLOUDINARY_API_KEY`,
`CLOUDINARY_API_SECRET`, dan `CLOUDINARY_FOLDER`. Alur unggah gambar:

1. Admin `POST /admin/uploads/images` (multipart, field `file`) → dapat `url`.
2. Kirim `url` itu sebagai `image_url` ke `POST /products/{id}/images` atau `POST /banners`.

Bukti pembayaran tidak ke Cloudinary; disimpan di kolom `payments.proof_data`.

**RajaOngkir (ongkir sampai kelurahan)** – isi `RAJAONGKIR_API_KEY` dan
`RAJAONGKIR_ORIGIN_ID` (ID kelurahan toko). Alur di frontend:

1. Dropdown berjenjang: `GET /regions/provinces` → `/regions/provinces/{id}/cities`
   → `/regions/cities/{id}/districts` (kecamatan) → `/regions/districts/{id}/subdistricts` (kelurahan).
2. `GET /shipping/costs?subdistrict_id=...` → ongkir; berat dihitung dari keranjang.
   Kurir ditentukan toko lewat `RAJAONGKIR_COURIERS` (saat ini `pos`), layanan termurah
   dipakai otomatis. Pelanggan tidak memilih kurir.
3. `POST /orders` dengan `province_id`, `city_id`, `district_id`, `subdistrict_id`.
   Server memvalidasi wilayah dan menghitung ulang ongkir.

Data wilayah di-cache 24 jam di memori supaya kuota API hemat.

## Alur order

```text
GET  /shipping/costs             -> lihat ongkir
POST /orders (checkout)          -> pending_payment
POST /orders/{id}/payment        -> tagihan + kode unik
PUT  /orders/{id}/payment/proof  -> unggah bukti
POST /admin/payments/{id}/review -> approved: order jadi paid
PUT  /admin/orders/{id}/status   -> processing
PUT  /admin/orders/{id}/status   -> shipped + tracking_number (wajib) -> email resi ke pelanggan
PUT  /admin/orders/{id}/status   -> completed
PUT  /orders/{id}/shipment       -> (opsional) koreksi kurir/resi
```

Order bisa dibatalkan (`cancelled`) selama belum `shipped`; stok dikembalikan.

## Performa

Optimasi yang dipakai (hasil benchmark ada di `hasil-testing.html`):

- **Cache publik di memori** (`middleware/cache.go`): GET produk, kategori, banner, dan
  pengaturan situs tanpa token dilayani dari memori (header `X-Cache: HIT`). Cache otomatis
  dikosongkan setiap ada perubahan data lewat API (termasuk checkout yang mengubah stok),
  dengan batas aman 30 detik.
- **Lebih sedikit perjalanan ke database**: daftar & detail produk serta daftar & detail order
  memakai `pgx.Batch` (beberapa query dikirim sekaligus).
- **Query daftar produk**: gambar utama hanya dicari untuk produk di halaman itu, bukan
  semua produk sebelum OFFSET.
- **Index database** (migrasi `20261105090000_add_performance_indexes`), termasuk
  `pg_trgm` untuk pencarian nama produk.
- **Pool koneksi** default 10 (`DB_MAX_CONNS`).
- **PGO** (profile-guided optimization): `cmd/api/default.pgo` otomatis dipakai `go build`.
  Perbarui profilnya bila kode berubah banyak (ambil CPU profile dari server saat dibebani,
  simpan sebagai `cmd/api/default.pgo`).

Password tetap memakai bcrypt cost 12; login sengaja lambat (~250 ms) demi keamanan.

## Perintah pengembangan

```bash
gofmt -l .        # cek format
go vet ./...      # analisis statis
go test ./...     # jalankan test
go build ./...    # pastikan semua package bisa di-build
go build -o NULL nama_file # build per field untuk mengetahui binary
go run nama_nama # menjalankan perfile untuk cek output
```
