# TODO Backend — peta dari desain (`../ui-hii-stranger`)

Alur: handlers -> services -> repository -> db. Detail per folder ada di `TODO.md` masing-masing folder.

## Urutan kerja

1. [x] Auth & akun — daftar_akun, masuk_akun, lupa_kata_sandi, akun_pelanggan (selesai, lihat "Riwayat perubahan tahap 1" di bawah)
2. [ ] Katalog — beranda, katalog_produk, detail_produk, editorial, kategori, banner (BE publik sudah dikerjakan, tinggal menunggu migrasi tabelnya; lihat "Riwayat perubahan tahap 2")
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

## Riwayat perubahan tahap 2 (katalog, BE sampai banner)

Urutan pengerjaan: models -> repository -> services -> uploads -> handlers -> routes. Migrasi TIDAK dikerjakan di sini (tugas tim; file migrasi katalog saat ini masih kosong).

Endpoint publik baru (prefix `/api/v1`, tanpa login, hanya membaca):

| Method + path | Fungsi |
| --- | --- |
| `GET /categories` | Kategori aktif, urut `sort_order` |
| `GET /products` | Daftar produk aktif. Query opsional: `category` (slug), `q`, `availability` (all / in_stock / preorder), `sort` (terbaru / harga_terendah / harga_tertinggi / nama_az), `page`, `limit` (1-50, bawaan 12). Respons: `items`, `page`, `limit`, `total` |
| `GET /products/{slug}` | Detail produk + varian |
| `GET /banners` | Banner yang sedang tayang (aktif dan dalam jadwal). Query opsional: `position` (hero / announcement / mid_feed / catalog) |
| `GET /uploads/{name}` | Gambar unggahan |

Query tidak valid -> 422 `{"error", "fields"}`. Produk tidak ada atau nonaktif -> 404. Produk hanya tampil kalau produk dan kategorinya sama-sama aktif.

File BARU:
- `models/category.go`, `models/product.go`, `models/banner.go`
- `repository/category.go`, `repository/product.go`, `repository/banner.go`
- `services/upload.go`: ImageStore (simpan gambar: cek tipe dari isi file, maks 2 MB, nama acak; sajikan dengan nama yang sudah dicek). Belum dipakai endpoint unggah; pemanggilnya baru ada di tahap admin.
- `handlers/catalog.go`, `handlers/upload.go`, `routes/catalog.go`
- Test: `services/catalog_test.go`, `services/upload_test.go`, `repository/product_test.go`, `handlers/catalog_test.go`

File DIUBAH:
- `services/catalog.go`: isinya tadinya hanya `TODO` dan `func main()` placeholder; sekarang CatalogService (validasi filter, diskon, label pre-order, status stok).
- `routes/routes.go`: memanggil `registerCatalogRoutes`.
- `.gitignore`: `uploads/*` tidak di-commit (kecuali `TODO.md`).
- `TODO.md` tiap folder dan `README.md`: centang dan status.

Kontrak skema untuk migrasi (kolom yang dipakai query; nama dan tipe harus cocok, atau kabari supaya query disesuaikan):

```
categories:       id, name, slug (unik), description, sort_order, is_active
products:         id, category_id -> categories.id, name, slug (unik), description,
                  price (BIGINT rupiah), original_price (BIGINT, boleh NULL = tidak diskon),
                  is_preorder, preorder_days (boleh NULL), image_url, is_active, created_at
product_variants: id, product_id -> products.id, color, size, sku, stock (INT, CHECK stock >= 0)
banners:          id, title, position (hero | announcement | mid_feed | catalog), image_url, link_url,
                  cta_label, sort_order, starts_at (boleh NULL), ends_at (boleh NULL), is_active
```

Hasil pengecekan:
- `go build`, `go vet`, `go test -race ./...`: lolos.
- Query dicoba ke PostgreSQL lokal memakai tabel TEMP sesuai kontrak di atas (bukan tabel dari migrasi tim): filter, pencarian (karakter % dan _ dibaca literal), urutan, paginasi, detail + varian, jadwal banner. File pengecekannya sudah dihapus.
- Server dijalankan sungguhan: route baru terdaftar tanpa bentrok, validasi 422, path traversal 404. Endpoint data membalas 500 generik sampai tabel katalog ada (detail SQL hanya di log server).

BELUM ada / diketahui:
- Tabel katalog dari migrasi tim. Sebelum itu jalan, endpoint data tidak bisa mengembalikan data. Catatan: `migrations/20261029004406_create_categories.up.sql` berisi teks nyasar (`Hanif pukamak` dan huruf `s`) yang akan membuat migrasi itu gagal.
- Desain detail produk menampilkan beberapa foto, tapi kontrak di atas hanya satu `image_url` per produk. Kalau mau banyak foto, perlu tabel/kolom baru (keputusan tim migrasi).
- Spesifikasi material, panduan ukuran, dan batas "maks. 3 pcs / transaksi" belum ada di skema; deskripsi teks bebas dipakai dulu.
- Endpoint unggah gambar (admin) dan rate limit untuk endpoint publik katalog belum ada.
- `models/catalog.go` dan `repository/catalog.go` (tipe `Katalog`, tabel `katalog`) tidak dipakai kode baru; tidak ada migrasi untuk tabel itu. Belum dihapus, menunggu keputusan.
