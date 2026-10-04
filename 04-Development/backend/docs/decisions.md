# Decisions

Catatan keputusan teknis beserta alasannya. Tambah entri baru tiap ada keputusan penting.

## 2026-10-05 - Template backend generik

**Konteks:** topik proyek kelompok belum ditentukan, tapi backend perlu mulai disiapkan.
**Keputusan:** bikin template generik berisi auth, role, dan CRUD user. Fitur domain ditambah setelah topik fix.
**Alasan:** auth dan user hampir pasti dibutuhkan di topik apa pun.

## 2026-10-05 - Bahasa dan framework: Go + Gin

**Alasan:** Gin populer, dokumentasinya banyak, dan routing serta middleware-nya sederhana. Cocok buat tim yang baru di Go.
**Alternatif:** Fiber, Echo, dan net/http standar. Semuanya bisa, tapi Gin paling banyak referensinya.

## 2026-10-05 - Database: PostgreSQL + GORM

**Alasan:** PostgreSQL stabil dan umum dipakai. GORM mempercepat CRUD dan punya AutoMigrate, cocok buat tahap awal.
**Catatan:** kalau skema sudah rumit, pertimbangkan pindah ke migrasi berbasis file (golang-migrate).

## 2026-10-05 - Autentikasi: JWT + bcrypt

**Alasan:** stateless, gampang dipakai dari frontend mana pun. Password di-hash pakai bcrypt, tidak disimpan polos.
**Catatan:** `JWT_SECRET` wajib diganti sebelum deploy.

## 2026-10-05 - Struktur: layered (handler, service, repository)

**Alasan:** tanggung jawab tiap layer jelas, gampang ditest, dan mudah dibagi kerja antar anggota kelompok.

## 2026-10-05 - Lokasi dokumentasi

**Keputusan:** dokumentasi teknis di `backend/docs`, logbook harian di `04-Development/logbook`.
**Alasan:** dokumen teknis dekat dengan kodenya, sedangkan logbook adalah catatan kerja pribadi per orang.
