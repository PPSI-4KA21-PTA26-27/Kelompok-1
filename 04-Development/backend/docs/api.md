# API Documentation

Base URL (lokal): `http://localhost:8080`
Prefix API: `/api/v1`

## Format Response

Semua response pakai bentuk yang sama:

```json
{
  "success": true,
  "message": "keterangan singkat",
  "data": {},
  "meta": { "page": 1, "limit": 10, "total": 25 },
  "errors": null
}
```

- `data` hanya ada kalau ada isinya
- `meta` hanya ada di endpoint yang pakai pagination
- `errors` hanya ada kalau `success: false`

## Autentikasi

Endpoint yang butuh login wajib kirim header:

```
Authorization: Bearer <token>
```

Token didapat dari `POST /api/v1/auth/login`. Endpoint khusus admin butuh token dengan role `admin`.

## Endpoint

### GET /health
Public. Cek server dan koneksi database.

### POST /api/v1/auth/register
Public. Role otomatis `user`.

Request:
```json
{ "name": "Bos", "email": "bos@example.com", "password": "minimal8karakter" }
```
Status: `201` berhasil, `400` validasi gagal, `409` email sudah terdaftar.

### POST /api/v1/auth/login
Public.

Request:
```json
{ "email": "bos@example.com", "password": "minimal8karakter" }
```
Response `data`:
```json
{ "token": "<jwt>", "user": { "id": 1, "name": "Bos", "email": "bos@example.com", "role": "user" } }
```
Status: `200` berhasil, `401` email atau password salah.

### GET /api/v1/me
Login. Profil user yang sedang login.

### GET /api/v1/users?page=1&limit=10
Admin. Daftar user dengan pagination (limit maksimal 100).

### GET /api/v1/users/:id
Admin. Detail satu user. `404` kalau tidak ada.

### PUT /api/v1/users/:id
Admin. Request (semua field opsional):
```json
{ "name": "Nama Baru", "role": "admin" }
```
`role` hanya boleh `user` atau `admin`.

### DELETE /api/v1/users/:id
Admin. Soft delete (data tetap ada di DB dengan `deleted_at` terisi).

## Kode Status Umum

| Kode | Arti |
|---|---|
| 400 | Request tidak valid |
| 401 | Token tidak ada, salah, atau kedaluwarsa |
| 403 | Role tidak punya akses |
| 404 | Data atau route tidak ditemukan |
| 409 | Konflik data (misal email dobel) |
| 500 | Error di server |

## Contoh curl

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"bos@example.com","password":"minimal8karakter"}'

curl http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer <token>"
```
