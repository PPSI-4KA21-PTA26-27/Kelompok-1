# Hii Stranger — Backend (Go)

API toko streetwear Hii Stranger. Prefix semua endpoint: `/api/v1`.

Alur kode: `routes` -> `middleware` -> `handlers` -> `services` -> `repository` -> `db` (PostgreSQL).

| Folder | Isi | Aturan |
| --- | --- | --- |
| `handlers/` | HTTP saja: baca request, validasi, panggil service, kirim respons | Tidak ada aturan bisnis / SQL |
| `services/` | Aturan bisnis | Tidak ada SQL |
| `repository/` | Query SQL (parameterized) lewat `db.DB` | Tidak ada aturan bisnis |
| `models/` | Struct tabel + bentuk request/response | `password_hash` jangan keluar ke API |
| `middleware/` | Logger, Recover, nanti Auth/CORS/rate limit | Lintas-request |
| `migrations/` | SQL skema, pasangan `.up.sql` + `.down.sql` | Jangan edit migrasi yang sudah jalan, tambah file baru |
| `response/` | Format JSON seragam `{"data": ...}` / `{"error": "..."}` | |
| `config/`, `db/` | Konfigurasi env, koneksi pgx | |
| `uploads/` | File unggahan (bukti bayar, gambar) | Disajikan lewat handler |

Detail tugas per folder ada di `TODO.md` masing-masing folder. Peta besar ada di `TODO.md` root backend.

## Aturan kerja (dari `AGENTS.md`)

- Baca `AGENTS.md` (root + `backend/`) dan cek struktur/`go.mod` sebelum mengubah apa pun.
- Solusi sekecil mungkin tapi lengkap. Tidak ada file, layer, atau dependency yang tidak perlu.
- Tidak ada data palsu, mock response, atau error yang ditelan di jalur produksi.
- Error selalu dikembalikan dengan konteks; jangan kirim detail SQL/stack trace ke klien.
- Query pakai parameter, input divalidasi di server, otorisasi dicek di server (bukan di UI).
- Jangan bikin crypto sendiri: password pakai bcrypt, token pakai library JWT.
- Dependency baru dipasang hanya saat tahapnya dikerjakan, dan dikonfirmasi dulu.
- Setiap selesai: `gofmt`, `go vet ./...`, `go test ./...`, lalu review `git diff`.

## Install & setup awal (sekali di depan)

Urutan:

1. **Go**, versi sesuai `go.mod`.
2. **PostgreSQL**, lalu buat database kosong untuk proyek ini.
3. **File `.env`** di folder `backend/` (jangan di-commit, pastikan masuk `.gitignore`):
   ```
   DB_URL=postgres://USER:PASSWORD@localhost:5432/NAMA_DB?sslmode=disable
   PORT=8080
   ```
   Variabel lain (`JWT_SECRET`, dll.) ditambah saat tahapnya dikerjakan.
4. **Dependency yang sudah ada** (`pgx`, `golang-migrate`, `godotenv`):
   ```powershell
   cd backend
   go mod download
   ```
5. **Jalankan migrasi** (harus dari folder `backend`, karena membaca `.env` dan `migrations/` dari folder aktif):
   ```powershell
   go run ./cmd/migrate
   ```
6. **Jalankan server** lalu cek `GET http://localhost:8080/api/v1/health`:
   ```powershell
   go run .
   ```

Cek kualitas (jalankan dari `backend/`):

```powershell
gofmt -l .
go vet ./...
go test ./...
```

## Urutan pengerjaan

Di tiap tahap, kerjakan folder **dari bawah ke atas** supaya yang di atas selalu punya pondasi:

`migrations` -> `models` -> `repository` -> `services` -> `middleware` (kalau perlu) -> `handlers` -> `routes` -> test

Setelah satu tahap backend selesai dan endpoint-nya jalan, baru buat halaman frontend yang sesuai (lihat bagian Frontend). Jangan kerjakan tahap berikutnya sebelum yang sekarang lolos cek kualitas.

### Tahap 0 — Fondasi

Folder: `config` -> `response` -> `middleware` -> `db`

- `config`: baca `JWT_SECRET` + masa berlaku token (wajib ada saat startup, jangan di-hardcode), origin CORS.
- `response`: format error validasi per field.
- `middleware`: CORS (origin Next.js), batas ukuran request body.
- `db`: batas pool koneksi.

Install: tidak ada.

### Tahap 1 — Auth & akun

Desain: `daftar_akun`, `masuk_akun`, `lupa_kata_sandi`, `akun_pelanggan`

Folder: `migrations` (users + nama lengkap, no. HP/WA, role; `password_reset_tokens`) -> `models` (User, Role) -> `repository` (User, ResetToken) -> `services` (AuthService) -> `middleware` (Auth JWT, RequireAdmin, rate limit login/register/lupa sandi) -> `handlers` (auth, akun) -> `routes` (`/auth/*`, `/me`)

Install saat mulai tahap ini:

```powershell
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto
```

Migrasi baru ditambah sebagai file `000002_...`, bukan mengedit `000001`.

### Tahap 2 — Katalog

Desain: `beranda`, `katalog_produk`, `detail_produk`, `editorial_streetwear`, kategori, banner

Folder: `migrations` (categories, products, product_variants, banners) -> `models` -> `repository` (list + filter + pagination, detail + varian) -> `services` (CatalogService, label pre-order) -> `uploads` (gambar produk/banner, validasi tipe + ukuran, nama acak, `.gitignore`) -> `handlers` -> `routes` (publik: `/products`, `/categories`, `/banners`)

Install: tidak ada (upload pakai standard library).

### Tahap 3 — Keranjang

Desain: `keranjang_belanja`

Folder: `migrations` (carts, cart_items) -> `models` -> `repository` (Cart) -> `services` (CartService, cek stok) -> `handlers` -> `routes` (`/cart`, butuh login)

Install: tidak ada.

### Tahap 4 — Checkout & order

Desain: `checkout_pesanan`, `daftar_pesanan`, `detail_pesanan`

Folder: `migrations` (shipping_rates, stock_movements, orders, order_items, order_status_logs) -> `models` -> `repository` (ShippingRate, Stock, Order; order + item dibuat dalam satu transaksi) -> `services` (ShippingService, StockService, OrderService: subtotal + ongkir, kode unik +Rp1 s/d +Rp999 dibuat di server, pre-order dikirim satu paket) -> `handlers` -> `routes` (`/orders`)

Install: tidak ada.

### Tahap 5 — Pembayaran QRIS

Desain: `detail_pesanan_pembayaran_qris`

Folder: `config` (data QRIS + nama merchant) -> `migrations` (payments, settings untuk subsidi QRIS Rp10.000) -> `models` -> `repository` (Payment, Setting) -> `services` (PaymentService, subsidi QRIS) -> `uploads` (bukti bayar) -> `handlers` -> `routes` (`/orders/{id}/payment`)

Install: tidak ada.

### Tahap 6 — Admin

Desain: `admin_pesanan`, `admin_pembayaran`, `admin_produk`, `admin_kategori`, `admin_banner`, `admin_ongkir`, `admin_ringkasan`, `admin_analitik`

Semua route di `/admin/*`, dibungkus `RequireAdmin`. Kerjakan berurutan (ringkasan/analitik terakhir karena butuh data dari yang lain):

1. Pesanan + verifikasi pembayaran (ubah status, catat `order_status_logs`)
2. Produk + stok + log mutasi (catatan wajib, stok tidak boleh minus)
3. Kategori, banner
4. Ongkir (tarif per kurir + provinsi)
5. Subsidi QRIS (toggle)
6. Ringkasan + analitik (omzet, jumlah order, produk terlaris)

Tiap butir: `repository` -> `services` -> `handlers` -> `routes`. Migrasi tidak perlu baru kecuali ada kolom yang kurang.

Install: tidak ada.

### Tahap 7 — Maintenance storefront

Desain: `storefront_maintenance`, `storefront_maintenance_notifikasi_terdaftar`, `admin_maintenance`

Folder: `migrations` (settings mode maintenance, maintenance_subscribers) -> `models` -> `repository` -> `services` (mode maintenance + notifikasi) -> `middleware` (blokir endpoint toko untuk pelanggan, admin tetap boleh) -> `handlers` -> `routes` (`/store/status` publik, toggle di `/admin/*`)

Install: tidak ada.

### Tahap 8 — Drop 05 / VIP (kalau masuk scope)

Desain: `drop_05_sold_out_vault_waitlist`, `private_catalog_vip_drop_05`, `vip_express_checkout_drop_05`, `vip_instant_qris_slip_drop_05`, `pesan_whatsapp_vip_drop_05`

Folder: `config` (kontak WhatsApp toko) -> `migrations` (waitlist) -> `models` -> `repository` -> `services` (WaitlistService, akses VIP) -> `handlers` -> `routes` (`/waitlist`)

Install: tidak ada.

## Keputusan yang harus diambil sebelum tahapnya

| Sebelum tahap | Pertanyaan |
| --- | --- |
| 1 | Token reset kata sandi dikirim lewat apa (email atau WhatsApp)? Desain belum menentukan. |
| 1 | Akun admin pertama dibuat dengan cara apa (SQL manual atau perintah khusus)? |
| 4 | Ongkir: tarif manual per kurir + provinsi, atau sinkron API kurir (JNE/SiCepat/J&T)? Checkout menulis "JNE Reguler", admin pesanan menulis "Flat Rate". |
| 4 | Order yang tidak dibayar kedaluwarsa otomatis atau tidak? |
| 5 | Verifikasi pembayaran otomatis atau manual oleh admin? Checkout menulis "verifikasi otomatis instan", admin pesanan menulis "verified by HQ Controller". |
| 8 | Fitur Drop/VIP masuk scope sekarang atau nanti? |

Status order dari desain: Awaiting Payment -> Processing. Nama status final ditentukan saat mulai tahap 4.

## Frontend (`../frontend`)

Folder `frontend/` saat ini baru berisi `AGENTS.md`, belum ada project Next.js. Desain acuan ada di `../ui-hii-stranger` (HTML + PNG, jangan diubah atau dipindah).

Urutan:

1. Setelah tahap 0 backend selesai, buat project Next.js (App Router + TypeScript) di `frontend/`. Desain memakai Tailwind, font Bricolage Grotesque + Geist, dan ikon Material Symbols; token warna/tipografi ada di `../ui-hii-stranger/editorial_streetwear/DESIGN.md`.
2. Buat halaman per tahap, **setelah** endpoint backend tahap itu jalan. Frontend tidak boleh memakai endpoint karangan atau data palsu.
3. Urutan halaman mengikuti tahap backend: auth -> beranda/katalog/detail -> keranjang -> checkout/pesanan -> pembayaran QRIS -> admin -> maintenance -> drop/VIP.

## Tahap 1 — cara menjalankan

Tambahkan ke `backend/.env` (server tidak mau jalan kalau `JWT_SECRET` kosong atau kurang dari 32 karakter):

```
JWT_SECRET=<isi acak, minimal 32 karakter>
JWT_TTL_MINUTES=15
FRONTEND_URL=http://localhost:3000
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASS=
SMTP_FROM=Hii Stranger <noreply@domain-kamu.com>
```

Buat `JWT_SECRET` acak (PowerShell):

```powershell
$b = New-Object byte[] 48; [System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($b); [Convert]::ToBase64String($b)
```

Lalu dari folder `backend`: `go run ./cmd/migrate` (menjalankan migrasi 000002 dan 000003), kemudian `go run .`.

| Method + path | Akses | Fungsi |
| --- | --- | --- |
| `POST /api/v1/auth/register` | publik | Daftar, langsung dapat token |
| `POST /api/v1/auth/login` | publik | Masuk, dapat token |
| `POST /api/v1/auth/forgot-password` | publik | Kirim tautan reset ke email (selalu 200) |
| `POST /api/v1/auth/reset-password` | publik | Sandi baru dengan token dari email |
| `GET /api/v1/me` | login | Data akun sendiri |
| `PUT /api/v1/me` | login | Ubah nama dan nomor HP |
| `PUT /api/v1/me/password` | login | Ganti kata sandi |

Token dikirim di header `Authorization: Bearer <token>` dan berlaku 15 menit. Keluar akun cukup dengan membuang token di klien (token disimpan di memori, sesuai desain).

Akun admin pertama: daftar seperti biasa, lalu ubah rolenya lewat SQL (pendaftaran publik tidak bisa membuat admin):

```sql
UPDATE users SET role = 'admin' WHERE email = 'email-kamu@contoh.com';
```

Email reset kata sandi memakai SMTP bawaan Go dan butuh port STARTTLS (587). Kalau `SMTP_*` kosong, server tetap jalan tapi email reset tidak terkirim (ada peringatan di log).

## Status

Sudah ada: health check, Logger, Recover, koneksi PostgreSQL, runner migrasi, tahap 1 (auth & akun).
Belum ada: tahap 2–8, CORS, dan pengiriman email SMTP yang sudah dites sungguhan.
Tahap 1 sudah dites langsung ke server lokal + PostgreSQL (21 skenario lolos) selain unit test, build, dan vet. Rincian perubahan kode ada di `TODO.md` bagian "Riwayat perubahan tahap 1".

Perhatian: `PORT` di `.env` harus angka saja (`PORT=8080`), bukan alamat lengkap.
