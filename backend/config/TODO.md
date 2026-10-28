# config

Baca konfigurasi dari environment (PORT, dll). Koneksi database ditangani paket `db`.

- [x] PORT (harus angka 1-65535, contoh 8080; salah isi seperti http://localhost:8080 membuat server berhenti dengan pesan jelas)
- [x] JWT_SECRET + masa berlaku token (JWT_TTL_MINUTES, default 15). Ditambah FRONTEND_URL dan SMTP_HOST/PORT/USER/PASS/FROM untuk email reset sandi
- [ ] Origin CORS yang diizinkan
- [ ] Folder + batas ukuran upload
- [ ] Data QRIS toko (gambar/string QRIS) dan nama merchant
- [ ] Kontak WhatsApp toko (untuk tombol hubungi/VIP)
