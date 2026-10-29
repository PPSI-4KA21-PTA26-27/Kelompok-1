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

## Tahap 2 — endpoint katalog

Semua publik (tanpa login) dan hanya membaca. Daftar query, aturan respons, dan kolom tabel yang dibutuhkan ada di `TODO.md` bagian "Riwayat perubahan tahap 2".

| Method + path | Fungsi |
| --- | --- |
| `GET /api/v1/categories` | Kategori aktif |
| `GET /api/v1/products` | Daftar produk (filter kategori, pencarian, ketersediaan, urutan, paginasi) |
| `GET /api/v1/products/{slug}` | Detail produk + varian |
| `GET /api/v1/banners` | Banner yang sedang tayang (filter `position`) |
| `GET /api/v1/uploads/{name}` | Gambar unggahan |

Endpoint data membutuhkan tabel katalog dari migrasi; sebelum itu dijalankan, responsnya 500.

## Daftar File yang Dibuat pada Setiap Tahap

Bagian ini merinci file yang perlu dibuat **ketika mulai mengerjakan tahap tersebut**. Daftar mengikuti cakupan dan urutan pada dokumen ini. Nama file adalah usulan penamaan yang jelas, bukan klaim bahwa semua file tersebut sudah ada di repository. Sebelum membuat file, cek struktur aktual dan gunakan file yang sudah tersedia. Untuk migrasi, selalu lanjutkan dari nomor terakhir yang benar-benar ada.

### Tahap 0 — Fondasi

Buat atau lengkapi file berikut sesuai kondisi proyek:

```text
backend/
├── config/
│   └── config.go
├── response/
│   └── response.go
├── middleware/
│   ├── cors.go
│   └── request_size.go
├── db/
│   └── postgres.go
├── routes/
│   └── routes.go
└── main.go
```

- `config/config.go` — membaca dan memvalidasi konfigurasi environment, termasuk `JWT_SECRET`, masa berlaku token, dan origin CORS sesuai kebutuhan tahap.
- `response/response.go` — format JSON sukses/error dan error validasi per field.
- `middleware/cors.go` — konfigurasi CORS untuk origin frontend.
- `middleware/request_size.go` — membatasi ukuran request body.
- `db/postgres.go` — koneksi dan pengaturan pool PostgreSQL.
- `routes/routes.go` — registrasi route dasar dan health check jika belum tersedia.
- `main.go` — merangkai konfigurasi, koneksi database, middleware, route, dan server jika entry point belum tersedia.

File yang sudah ada seperti logger, recover, koneksi database, route health check, atau entry point tidak perlu dibuat ulang.

### Tahap 1 — Auth & akun

Buat atau lengkapi file berikut:

```text
backend/
├── migrations/
│   ├── 000002_create_users.up.sql
│   ├── 000002_create_users.down.sql
│   ├── 000003_create_password_reset_tokens.up.sql
│   └── 000003_create_password_reset_tokens.down.sql
├── models/
│   ├── user.go
│   └── password_reset_token.go
├── repository/
│   ├── user_repository.go
│   └── password_reset_token_repository.go
├── services/
│   ├── auth_service.go
│   └── account_service.go
├── middleware/
│   ├── auth.go
│   ├── require_admin.go
│   └── rate_limit.go
├── handlers/
│   ├── auth_handler.go
│   └── account_handler.go
└── routes/
    ├── auth_routes.go
    └── account_routes.go
```

- Migrasi `users` — menyimpan data akun, nama lengkap, nomor HP/WA, role, dan password hash sesuai desain.
- Migrasi `password_reset_tokens` — menyimpan token reset kata sandi.
- `models/user.go` — struct `User` dan role yang diperlukan; jangan pernah mengirim `password_hash` ke API.
- `models/password_reset_token.go` — struct token reset.
- `repository/user_repository.go` — query database untuk pengguna.
- `repository/password_reset_token_repository.go` — query database untuk token reset.
- `services/auth_service.go` — registrasi, login, dan alur lupa/reset kata sandi.
- `services/account_service.go` — membaca dan memperbarui data akun serta mengganti kata sandi.
- `middleware/auth.go` — validasi token JWT.
- `middleware/require_admin.go` — membatasi akses khusus admin.
- `middleware/rate_limit.go` — membatasi percobaan login, registrasi, dan lupa kata sandi.
- `handlers/auth_handler.go` — handler register, login, forgot password, dan reset password.
- `handlers/account_handler.go` — handler `/me`, pembaruan profil, dan penggantian kata sandi.
- `routes/auth_routes.go` — route `/auth/*`.
- `routes/account_routes.go` — route `/me`.

Dokumen sumber menyebut migrasi tahap 1 bernomor `000002` dan `000003`. Gunakan nomor tersebut hanya jika memang belum terpakai; jangan menimpa migrasi yang sudah ada.

Dependency yang dipasang saat mulai tahap 1:

```powershell
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto
```

Konfigurasi `.env` tahap 1 mengikuti bagian **Tahap 1 — cara menjalankan** pada dokumen ini.

### Tahap 2 — Katalog

Buat atau lengkapi file berikut:

```text
backend/
├── migrations/
│   ├── [nomor berikutnya]_create_categories.up.sql
│   ├── [nomor berikutnya]_create_categories.down.sql
│   ├── [nomor berikutnya]_create_products.up.sql
│   ├── [nomor berikutnya]_create_products.down.sql
│   ├── [nomor berikutnya]_create_product_variants.up.sql
│   ├── [nomor berikutnya]_create_product_variants.down.sql
│   ├── [nomor berikutnya]_create_banners.up.sql
│   └── [nomor berikutnya]_create_banners.down.sql
├── models/
│   ├── category.go
│   ├── product.go
│   ├── product_variant.go
│   └── banner.go
├── repository/
│   ├── category_repository.go
│   ├── product_repository.go
│   └── banner_repository.go
├── services/
│   └── catalog_service.go
├── handlers/
│   ├── category_handler.go
│   ├── product_handler.go
│   └── banner_handler.go
├── routes/
│   └── catalog_routes.go
└── uploads/
    └── .gitkeep
```

- Migrasi membuat tabel kategori, produk, varian produk, dan banner.
- `models/` berisi struct untuk kategori, produk, varian, dan banner.
- `repository/` menangani query daftar produk dengan filter/pagination, detail produk beserta varian, kategori, dan banner.
- `services/catalog_service.go` menangani aturan katalog, termasuk label pre-order.
- `handlers/` menangani endpoint katalog publik.
- `routes/catalog_routes.go` mendaftarkan `/products`, `/categories`, dan `/banners`.
- `uploads/.gitkeep` hanya diperlukan jika folder upload kosong dan perlu dipertahankan di Git. File gambar yang diunggah tidak di-commit.

Nama file migrasi di atas merupakan nama dasar. Ganti `[nomor berikutnya]` dengan nomor unik berurutan sesuai format migrasi proyek. Jangan menyalin teks placeholder itu secara literal.

### Tahap 3 — Keranjang

Buat atau lengkapi file berikut:

```text
backend/
├── migrations/
│   ├── [nomor berikutnya]_create_carts.up.sql
│   ├── [nomor berikutnya]_create_carts.down.sql
│   ├── [nomor berikutnya]_create_cart_items.up.sql
│   └── [nomor berikutnya]_create_cart_items.down.sql
├── models/
│   ├── cart.go
│   └── cart_item.go
├── repository/
│   └── cart_repository.go
├── services/
│   └── cart_service.go
├── handlers/
│   └── cart_handler.go
└── routes/
    └── cart_routes.go
```

- Migrasi membuat tabel `carts` dan `cart_items`.
- `models/cart.go` dan `models/cart_item.go` menyimpan bentuk data keranjang.
- `repository/cart_repository.go` menangani query keranjang.
- `services/cart_service.go` menerapkan aturan keranjang dan pengecekan stok.
- `handlers/cart_handler.go` menangani request keranjang.
- `routes/cart_routes.go` mendaftarkan `/cart` dengan akses login.

Gunakan nomor migrasi unik berikutnya; jangan gunakan placeholder sebagai nomor migrasi sungguhan.

### Tahap 4 — Checkout & order

Buat atau lengkapi file berikut:

```text
backend/
├── migrations/
│   ├── [nomor berikutnya]_create_shipping_rates.up.sql
│   ├── [nomor berikutnya]_create_shipping_rates.down.sql
│   ├── [nomor berikutnya]_create_stock_movements.up.sql
│   ├── [nomor berikutnya]_create_stock_movements.down.sql
│   ├── [nomor berikutnya]_create_orders.up.sql
│   ├── [nomor berikutnya]_create_orders.down.sql
│   ├── [nomor berikutnya]_create_order_items.up.sql
│   ├── [nomor berikutnya]_create_order_items.down.sql
│   ├── [nomor berikutnya]_create_order_status_logs.up.sql
│   └── [nomor berikutnya]_create_order_status_logs.down.sql
├── models/
│   ├── shipping_rate.go
│   ├── stock_movement.go
│   ├── order.go
│   ├── order_item.go
│   └── order_status_log.go
├── repository/
│   ├── shipping_rate_repository.go
│   ├── stock_repository.go
│   └── order_repository.go
├── services/
│   ├── shipping_service.go
│   ├── stock_service.go
│   └── order_service.go
├── handlers/
│   └── order_handler.go
└── routes/
    └── order_routes.go
```

- Migrasi menambahkan skema ongkir, mutasi stok, pesanan, item pesanan, dan log status pesanan.
- `repository/order_repository.go` harus membuat order dan item order dalam satu transaksi.
- `services/shipping_service.go` menangani ongkir.
- `services/stock_service.go` menangani pengecekan dan perubahan stok.
- `services/order_service.go` menghitung subtotal dan ongkir, membuat kode unik tambahan Rp1–Rp999 di server, dan mengikuti aturan pengiriman pre-order satu paket.
- `handlers/order_handler.go` menangani checkout dan endpoint pesanan.
- `routes/order_routes.go` mendaftarkan endpoint `/orders`.

Buat migrasi memakai nomor berikutnya yang belum terpakai.

### Tahap 5 — Pembayaran QRIS

Buat atau lengkapi file berikut:

```text
backend/
├── config/
│   └── qris.go
├── migrations/
│   ├── [nomor berikutnya]_create_payments.up.sql
│   ├── [nomor berikutnya]_create_payments.down.sql
│   ├── [nomor berikutnya]_create_settings.up.sql
│   └── [nomor berikutnya]_create_settings.down.sql
├── models/
│   ├── payment.go
│   └── setting.go
├── repository/
│   ├── payment_repository.go
│   └── setting_repository.go
├── services/
│   └── payment_service.go
├── handlers/
│   └── payment_handler.go
└── routes/
    └── payment_routes.go
```

- `config/qris.go` — konfigurasi data QRIS dan nama merchant.
- Migrasi — tabel pembayaran dan pengaturan, termasuk pengaturan subsidi QRIS Rp10.000.
- `models/payment.go` dan `models/setting.go` — bentuk data pembayaran dan pengaturan.
- Repository — query pembayaran dan pengaturan.
- `services/payment_service.go` — aturan pembayaran dan subsidi QRIS.
- `handlers/payment_handler.go` — request pembayaran dan unggah bukti bayar.
- `routes/payment_routes.go` — route `/orders/{id}/payment`.
- Bukti bayar disimpan di folder `uploads/` saat aplikasi berjalan; validasi tipe dan ukuran file.

### Tahap 6 — Admin

Buat atau lengkapi file berikut sesuai urutan pengerjaan admin:

```text
backend/
├── handlers/
│   ├── admin_order_handler.go
│   ├── admin_payment_handler.go
│   ├── admin_product_handler.go
│   ├── admin_category_handler.go
│   ├── admin_banner_handler.go
│   ├── admin_shipping_handler.go
│   ├── admin_setting_handler.go
│   └── admin_analytics_handler.go
├── services/
│   ├── admin_order_service.go
│   ├── admin_payment_service.go
│   ├── admin_product_service.go
│   ├── admin_category_service.go
│   ├── admin_banner_service.go
│   ├── admin_shipping_service.go
│   ├── admin_setting_service.go
│   └── admin_analytics_service.go
└── routes/
    └── admin_routes.go
```

Repository dan model baru **tidak otomatis wajib dibuat**. Periksa repository/model dari tahap sebelumnya dan tambahkan hanya jika operasi admin membutuhkan query atau bentuk data baru.

Urutan implementasi:
1. Pesanan dan verifikasi pembayaran; catat perubahan di `order_status_logs`.
2. Produk dan stok; wajib ada catatan mutasi dan stok tidak boleh minus.
3. Kategori dan banner.
4. Ongkir per kurir dan provinsi.
5. Toggle subsidi QRIS.
6. Ringkasan dan analitik, setelah data fitur sebelumnya tersedia.

Semua endpoint admin berada di `/admin/*` dan harus dilindungi `RequireAdmin`. Migrasi baru hanya diperlukan jika skema database memang belum memiliki kolom/tabel yang dibutuhkan.

### Tahap 7 — Maintenance storefront

Buat atau lengkapi file berikut:

```text
backend/
├── migrations/
│   ├── [nomor berikutnya]_create_maintenance_settings.up.sql
│   ├── [nomor berikutnya]_create_maintenance_settings.down.sql
│   ├── [nomor berikutnya]_create_maintenance_subscribers.up.sql
│   └── [nomor berikutnya]_create_maintenance_subscribers.down.sql
├── models/
│   └── maintenance_subscriber.go
├── repository/
│   ├── maintenance_repository.go
│   └── maintenance_subscriber_repository.go
├── services/
│   └── maintenance_service.go
├── middleware/
│   └── maintenance.go
├── handlers/
│   ├── store_status_handler.go
│   └── admin_maintenance_handler.go
└── routes/
    └── maintenance_routes.go
```

- Migrasi — pengaturan mode maintenance dan tabel pelanggan notifikasi.
- `models/maintenance_subscriber.go` — bentuk data pelanggan notifikasi.
- Repository — query pengaturan maintenance dan pelanggan notifikasi.
- `services/maintenance_service.go` — aturan mode maintenance dan notifikasi.
- `middleware/maintenance.go` — memblokir endpoint toko untuk pelanggan, sementara admin tetap dapat mengaksesnya.
- Handler dan route — status publik `/store/status` serta pengaturan melalui endpoint admin.

### Tahap 8 — Drop 05 / VIP (jika masuk scope)

Buat atau lengkapi file berikut:

```text
backend/
├── config/
│   └── whatsapp.go
├── migrations/
│   ├── [nomor berikutnya]_create_waitlist.up.sql
│   └── [nomor berikutnya]_create_waitlist.down.sql
├── models/
│   └── waitlist.go
├── repository/
│   └── waitlist_repository.go
├── services/
│   ├── waitlist_service.go
│   └── vip_service.go
├── handlers/
│   ├── waitlist_handler.go
│   └── vip_handler.go
└── routes/
    ├── waitlist_routes.go
    └── vip_routes.go
```

- `config/whatsapp.go` — kontak WhatsApp toko.
- Migrasi — tabel waitlist.
- `models/waitlist.go` — bentuk data waitlist.
- `repository/waitlist_repository.go` — query waitlist.
- `services/waitlist_service.go` — aturan pendaftaran waitlist.
- `services/vip_service.go` — aturan akses VIP.
- Handler dan route — waitlist serta endpoint VIP yang disepakati.

Tahap ini dikerjakan hanya jika Drop 05/VIP masuk scope proyek.

### File test yang dibuat seiring pengerjaan

Buat test pada tahap terkait, bukan semuanya di awal:

```text
config/config_test.go
response/response_test.go
middleware/auth_test.go
services/auth_service_test.go
services/catalog_service_test.go
services/cart_service_test.go
services/order_service_test.go
services/payment_service_test.go
```

Daftar tersebut adalah titik awal yang disarankan. Tambahkan test untuk repository, handler, dan fitur lain sesuai kebutuhan serta pola test proyek yang sudah ada.

### Catatan penting sebelum membuat file

- Daftar ini merinci file per tahap, tetapi tidak berarti seluruh file harus dibuat sekaligus.
- Nama file adalah konvensi yang disarankan; ikuti pola yang sudah dipakai proyek bila berbeda.
- Jangan membuat ulang `main.go`, `config.go`, koneksi DB, middleware, model, repository, atau handler yang sudah ada.
- Jangan menyalin `[nomor berikutnya]` secara literal ke nama migrasi. Tentukan nomor berdasarkan file migrasi aktual.
- Jangan mengedit migrasi yang sudah dijalankan; buat migrasi baru jika skema perlu berubah.
- Tetap ikuti urutan pengerjaan dan pemeriksaan kualitas di bagian **Urutan pengerjaan** dan **Aturan kerja**.

---

## Status

Sudah ada: health check, Logger, Recover, koneksi PostgreSQL, runner migrasi, tahap 1 (auth & akun), tahap 2 sisi backend publik (kategori, produk, banner, gambar; tinggal menunggu migrasi tabelnya dari tim).
Belum ada: tahap 3–8, CORS, dan pengiriman email SMTP yang sudah dites sungguhan.
Tahap 1 sudah dites langsung ke server lokal + PostgreSQL (21 skenario lolos) selain unit test, build, dan vet. Rincian perubahan kode ada di `TODO.md` bagian "Riwayat perubahan tahap 1".

Perhatian: `PORT` di `.env` harus angka saja (`PORT=8080`), bukan alamat lengkap.
