# TODO Backend — peta dari desain (`../ui-hii-stranger`)

Alur: handlers -> services -> repository -> db. Detail per folder ada di `TODO.md` masing-masing folder.

## Urutan kerja

1. [x] Auth & akun — daftar_akun, masuk_akun, lupa_kata_sandi, akun_pelanggan (selesai, lihat "Riwayat perubahan tahap 1" di bawah)
2. [ ] Katalog — beranda, katalog_produk, detail_produk, editorial, kategori, banner
3. [ ] Keranjang — keranjang_belanja
4. [ ] Checkout & order — checkout_pesanan, daftar_pesanan, detail_pesanan (alamat, ongkir, pre-order)
5. [ ] Pembayaran QRIS — detail_pesanan_pembayaran_qris (kode unik, bukti bayar, verifikasi)
6. [ ] Admin — ringkasan, analitik, pesanan, produk + stok, kategori, banner, ongkir, pembayaran, maintenance
7. [ ] Maintenance storefront — storefront_maintenance (+ notifikasi terdaftar)
8. [ ] Drop 05 / VIP — waitlist sold-out vault, katalog privat, express checkout, QRIS instan, WhatsApp VIP

## Perlu keputusan dulu (desain belum jelas / saling beda)

- Verifikasi pembayaran: halaman checkout bilang "verifikasi otomatis instan", tapi halaman admin pesanan menulis "Payment proof verified by HQ Controller" (manual). Pilih otomatis atau manual?
- Ongkir: checkout menampilkan "JNE Reguler", admin pesanan menampilkan "Flat Rate", admin ongkir punya tarif per kurir + provinsi + tombol sinkron API JNE/SiCepat/J&T. Pilih tarif manual per kurir+provinsi atau API kurir?
- Fitur Drop/VIP (nomor 8): masuk scope sekarang atau nanti?
- Order yang tidak dibayar: kedaluwarsa otomatis atau tidak?

## Catatan

- Daftar ini disusun dari nama halaman desain; checkout, admin produk, admin ongkir, admin pesanan, dan waitlist drop sudah dibaca detailnya. Halaman lain baru berdasarkan nama folder, jadi cek lagi saat mengerjakannya.
- Status order dari desain: Awaiting Payment -> Processing (nama status final tentukan saat mulai bagian 4).

## Riwayat perubahan tahap 1 (auth & akun)

Urutan pengerjaan: migrations -> models -> repository -> services -> middleware -> handlers -> routes.

File BARU:
- `migrations/000002_extend_users_for_auth` (up/down): tambah `full_name`, `phone`, `role` (customer/admin). `username` dibuat boleh kosong, tidak dihapus, karena form daftar di desain tidak punya username.
- `migrations/000003_create_password_reset_tokens` (up/down): tabel token reset; yang disimpan hash SHA-256, bukan token asli.
- `models/user.go`: User, Role, dan bentuk request/respons auth. `password_hash` tidak ikut ke JSON.
- `repository/user.go`: create, cari by email/id, ubah profil, ubah password. `repository/reset_token.go`: simpan token, dan reset sandi dalam satu transaksi (token sekali pakai).
- `services/auth.go`: daftar, masuk, lupa/reset sandi, profil, ganti sandi (bcrypt). `services/token.go`: buat dan baca JWT HS256. `services/validation.go`: aturan input sesuai desain. `services/mailer.go`: kirim email reset lewat SMTP bawaan Go.
- `middleware/auth.go`: Auth (cek token) dan RequireAdmin (cek role). `middleware/ratelimit.go`: batas percobaan per IP.
- `handlers/account.go`: GET/PUT `/me`, PUT `/me/password`. `handlers/request.go`: baca JSON (batas 16 KB, tolak field asing) dan terjemah error ke status HTTP.
- Test: `services/validation_test.go`, `services/token_test.go`, `middleware/auth_test.go`, `middleware/ratelimit_test.go`.
- `.gitignore` (isi: `.env`).

File DIUBAH:
- `config/config.go`: tambah JWT_SECRET (wajib, min 32 karakter), JWT_TTL_MINUTES (default 15), FRONTEND_URL, SMTP_*; validasi PORT harus angka 1-65535.
- `response/response.go`: tambah `ValidationError` ({"error", "fields"}).
- `handlers/auth.go`: menggantikan draft lama (struct huruf kecil yang belum bisa dipakai) dengan handler daftar/masuk/lupa/reset.
- `routes/auth.go`: menggantikan draft lama (memanggil handler yang tidak ada); sekarang menyambung db -> repository -> services -> handlers dan mendaftarkan 7 route.
- `routes/routes.go`: `New(cfg)` sekarang menerima config dan mengembalikan error; tetap memanggil `registerAuthRoutes`.
- `main.go`: memanggil `routes.New(cfg)` dan berhenti dengan pesan jelas kalau gagal.
- `go.mod` / `go.sum`: tambah `golang-jwt/jwt/v5` dan `golang.org/x/crypto`; `x/sync` dan `x/text` ikut naik versi otomatis.
- `TODO.md` tiap folder dan `README.md`: centang dan cara menjalankan.

Hasil pengecekan:
- `go build`, `go vet`, `go test -race ./...`: lolos.
- Tes langsung ke server lokal + PostgreSQL: 21 dari 21 lolos (daftar, masuk, /me, ganti sandi, lupa sandi, reset sandi sampai berhasil, token kedaluwarsa dan sekali pakai, batas body 413, rate limit 429). Akun tes sudah dihapus.
- Migrasi 000002 dan 000003 sudah ter-apply (versi 3, tidak dirty).

BELUM diuji / belum ada:
- Pengiriman email SMTP sungguhan (SMTP_* belum diisi; saat kosong, server hanya mencatat di log bahwa email tidak terkirim).
- CORS (tahap 0, belum dikerjakan; dibutuhkan saat frontend Next.js memanggil API dari browser).
- Pemeriksaan di `gofmt` untuk `cmd/migrate/main.go` dan `db/connetion.go` (sudah belum rapi sebelum tahap 1, tidak disentuh).

Keputusan sementara (bisa diubah):
- Tautan reset dikirim lewat email (sesuai desain). Kalau mau lewat WhatsApp, ganti implementasi `ResetSender` di `services/mailer.go`.
- Admin pertama dibuat manual lewat SQL (`UPDATE users SET role = 'admin' WHERE email = '...'`).
- Sesi = satu access token 15 menit, tanpa refresh token. Keluar akun = klien membuang token.

Risiko yang diketahui:
- `backend/.env` sudah ter-commit di Git (commit "Update"). `.gitignore` baru belum mencabutnya dari Git; perlu `git rm --cached backend/.env` dan ganti password database kalau isinya rahasia asli.
- Rate limit disimpan di memori satu server (hilang saat restart, tidak dibagi antar-server) dan memakai IP langsung dari koneksi; kalau nanti di belakang reverse proxy, IP asli perlu diambil dari proxy yang dipercaya.
- Role di token berlaku sampai token habis (maks 15 menit), jadi perubahan role baru terasa setelah masuk ulang.
