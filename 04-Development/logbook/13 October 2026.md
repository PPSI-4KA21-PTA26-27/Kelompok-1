# Hii Stranger — Backend API

**Hii Stranger** adalah aplikasi e-commerce pakaian yang dikembangkan untuk proyek Pengelolaan Proyek Sistem Informasi (PPSI) kelas 4KA21. Backend ini menyediakan REST API untuk autentikasi pengguna, katalog produk, manajemen stok, pemesanan, dan pembayaran.

> **Project Status:** In Development  
> **Last Updated:** 13 October 2026

## Tech Stack

| Teknologi | Kegunaan |
|---|---|
| Go | Bahasa pemrograman backend |
| Gin | HTTP web framework |
| PostgreSQL 16 | Relational database |
| pgx v5 / pgxpool | Database driver dan connection pooling |
| JWT | Autentikasi berbasis token |
| bcrypt | Password hashing |
| golang-migrate | Database migration |
| Docker | Containerization |
| Railway | Backend deployment |
| Midtrans Sandbox | Pengujian integrasi pembayaran QRIS |

## Features

### Authentication & Authorization
- Customer registration and login.
- Admin login.
- Access token and refresh token.
- Logout and session revocation.
- Password hashing menggunakan bcrypt.
- Password change and password reset.
- Role-based access control.
- Resource ownership validation.

### Product Catalog
- Product listing and product details.
- Category management.
- Product variants, such as size and color.
- Product search and category filtering.
- Pagination and sorting.
- Product availability and stock validation.

### Shopping & Orders
- Shopping cart.
- Checkout and order creation.
- Server-side price calculation.
- Order history for authenticated customers.
- Order detail and status tracking.
- Order management for admins.
- Transaction-safe inventory management.

### Payment
- Payment provider integration.
- QRIS payment through Midtrans Sandbox.
- Payment status verification.
- Webhook signature validation.
- Idempotent payment processing.
- Protection against duplicate payment processing.

### Security
- Input validation.
- Parameterized SQL queries.
- Authentication and authorization middleware.
- Rate limiting for sensitive endpoints.
- CORS allowlist.
- Secure environment variable management.
- Request size limits and HTTP timeouts.
- Structured logging without sensitive credentials.
- Protection against unauthorized order access.

## Architecture

The backend uses a **modular monolith architecture**. Each module has a clear responsibility, while the application remains deployable as one service.

```text
Frontend (React + TypeScript)
            |
            v
      REST API (Gin)
            |
            v
   Authentication & Middleware
            |
            v
       Business Logic
            |
            v
    PostgreSQL (pgxpool)

External Services:
- Midtrans Payment Gateway
- Email Provider (if configured)
```

## Project Structure

```text
backend/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   ├── database/
│   ├── httpapi/
│   │   ├── middleware/
│   │   ├── routes.go
│   │   └── response.go
│   ├── auth/
│   ├── users/
│   ├── admins/
│   ├── categories/
│   ├── products/
│   ├── customers/
│   ├── cart/
│   ├── orders/
│   ├── payments/
│   ├── inventory/
│   ├── email/
│   └── platform/
├── migrations/
├── scripts/
├── tests/
├── .env.example
├── .gitignore
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

The actual structure may evolve during development.

## Requirements

Pastikan perangkat sudah memiliki:

- Go versi yang kompatibel dengan `go.mod`.
- PostgreSQL 16 atau database PostgreSQL yang kompatibel.
- Git.
- Docker (opsional untuk pengembangan lokal, tergantung konfigurasi).
- Railway account untuk deployment.
- Midtrans Sandbox credentials jika pengujian pembayaran dilakukan.

## Environment Variables

Buat file `.env` berdasarkan `.env.example`.

```env
APP_ENV=development
PORT=8080

DATABASE_URL=

JWT_ACCESS_SECRET=
JWT_ISSUER=hii-stranger-api
ACCESS_TOKEN_TTL=
REFRESH_TOKEN_TTL=

FRONTEND_URL=
CORS_ALLOWED_ORIGINS=

MIDTRANS_SERVER_KEY=
MIDTRANS_CLIENT_KEY=
MIDTRANS_IS_PRODUCTION=false

EMAIL_PROVIDER=
EMAIL_API_KEY=
```

Isi nilai sesuai konfigurasi lokal atau deployment. Nama variabel dan format durasi harus mengikuti implementasi aktual di source code.

**Security Notes:**
- Jangan commit file `.env` ke repository.
- Jangan menaruh credential production di source code.
- Jangan menggunakan secret JWT default.
- Jangan membagikan server key Midtrans.
- Gunakan secret manager atau environment variables pada platform deployment.

## Installation

Clone repository dan masuk ke direktori backend:

```bash
git clone <REPOSITORY_URL>
cd <REPOSITORY_FOLDER>/backend
```

Unduh dependency Go:

```bash
go mod download
```

Konfigurasikan PostgreSQL dan environment variables sebelum menjalankan aplikasi.

## Database Migration

Gunakan migration yang tersedia di direktori `migrations/`.

Pastikan database sudah dibuat dan `DATABASE_URL` sudah benar sebelum menjalankan migration.

Perintah migration harus mengikuti konfigurasi dan script yang disediakan oleh project. Jangan menjalankan migration production secara manual tanpa memeriksa perubahan schema terlebih dahulu.

## Running Locally

Setelah database dan konfigurasi siap:

```bash
go run ./cmd/api
```

Jika entry point atau struktur repository berbeda, sesuaikan perintah dengan implementasi aktual.

Backend akan berjalan pada port yang ditentukan melalui environment variable `PORT`, jika konfigurasi tersebut telah diterapkan.

## API Endpoints

Semua endpoint API menggunakan prefix `/api/v1`.

### Authentication

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/auth/register` | Registrasi customer |
| POST | `/api/v1/auth/login` | Login customer |
| POST | `/api/v1/auth/logout` | Logout customer |
| POST | `/api/v1/auth/refresh` | Refresh token |
| POST | `/api/v1/auth/forgot-password` | Memulai reset password |
| POST | `/api/v1/auth/reset-password` | Reset password |
| GET | `/api/v1/me` | Melihat profil sendiri |
| PATCH | `/api/v1/me` | Memperbarui profil |
| POST | `/api/v1/me/change-password` | Mengganti password |

### Admin Authentication

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/admin/auth/login` | Login admin |
| POST | `/api/v1/admin/auth/logout` | Logout admin |
| POST | `/api/v1/admin/auth/refresh` | Refresh token admin |

### Products & Categories

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/categories` | Daftar kategori |
| GET | `/api/v1/products` | Daftar produk |
| GET | `/api/v1/products/:id` | Detail produk |
| GET | `/api/v1/products/slug/:slug` | Detail produk berdasarkan slug |
| POST | `/api/v1/admin/categories` | Membuat kategori |
| PATCH | `/api/v1/admin/categories/:id` | Memperbarui kategori |
| POST | `/api/v1/admin/products` | Membuat produk |
| PATCH | `/api/v1/admin/products/:id` | Memperbarui produk |
| DELETE | `/api/v1/admin/products/:id` | Menghapus atau menonaktifkan produk |

### Orders & Payments

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/orders` | Membuat pesanan |
| GET | `/api/v1/orders` | Riwayat pesanan customer |
| GET | `/api/v1/orders/:id` | Detail pesanan |
| GET | `/api/v1/orders/number/:orderNumber` | Pencarian pesanan berdasarkan nomor |
| GET | `/api/v1/admin/orders` | Daftar pesanan untuk admin |
| PATCH | `/api/v1/admin/orders/:id/status` | Memperbarui status pesanan |
| POST | `/api/v1/payments/create` | Membuat transaksi pembayaran |
| GET | `/api/v1/payments/:id` | Memeriksa status pembayaran |
| POST | `/api/v1/webhooks/midtrans` | Menerima notifikasi pembayaran |

Endpoint di atas merupakan rancangan awal. Ketersediaan endpoint dan metode HTTP harus mengikuti implementasi aktual.

## API Response Format

Contoh respons berhasil:

```json
{
  "success": true,
  "message": "Products retrieved successfully",
  "data": [],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 0
  }
}
```

Contoh respons error:

```json
{
  "success": false,
  "message": "Invalid request",
  "error": {
    "code": "VALIDATION_ERROR",
    "details": []
  }
}
```

## Testing

Jalankan unit test:

```bash
go test ./...
```

Jalankan pemeriksaan kode:

```bash
go vet ./...
```

Integration test membutuhkan PostgreSQL testing environment yang sesuai.

Pengujian wajib mencakup autentikasi, authorization, validasi input, checkout, konkurensi stok, dan verifikasi webhook pembayaran.

## Deployment

Backend direncanakan untuk di-deploy ke **Railway**.

Garis besar deployment:

1. Siapkan database PostgreSQL.
2. Atur environment variables melalui Railway.
3. Konfigurasikan `DATABASE_URL` dan secret yang diperlukan.
4. Jalankan database migration menggunakan prosedur deployment yang aman.
5. Deploy backend.
6. Pastikan health check berhasil.
7. Konfigurasikan CORS untuk domain frontend.
8. Uji autentikasi, katalog, checkout, dan integrasi pembayaran Sandbox.

Jangan gunakan credential Sandbox untuk transaksi production. Aktifkan pembayaran production hanya setelah konfigurasi dan pengujian resmi selesai.

## Development Guidelines

- Ikuti prinsip clean code dan separation of concerns.
- Gunakan parameterized SQL queries.
- Validasi input pada backend.
- Terapkan least privilege.
- Jangan mempercayai harga atau status pembayaran dari frontend.
- Gunakan transaksi database untuk operasi yang membutuhkan konsistensi.
- Tulis test untuk business logic penting.
- Dokumentasikan perubahan endpoint dan database.
- Jalankan test sebelum melakukan pull request atau merge.
- Jangan menambahkan kompleksitas infrastruktur tanpa kebutuhan yang jelas.

## Project Status

Proyek masih dalam tahap pengembangan. Fitur yang tercantum dalam README merupakan target atau rancangan sampai implementasinya benar-benar tersedia dan diuji.

## Team

**Project:** Hii Stranger  
**Course:** Pengelolaan Proyek Sistem Informasi (PPSI)  
**Class:** 4KA21  
**Academic Year:** 2026/2027
