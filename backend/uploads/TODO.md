# uploads

Folder baru: tempat file yang diunggah (disajikan lewat handler, bukan langsung dari disk publik).

- [ ] Bukti pembayaran (wajib untuk order QRIS)
- [ ] Gambar produk dan banner (penyajian sudah ada: GET /api/v1/uploads/{name}; endpoint unggah menyusul di tahap admin)
- [x] Validasi tipe + ukuran file, ganti nama file jadi acak (cegah path traversal)
- [x] Tambahkan `uploads/*` ke `.gitignore` (kecuali TODO.md)
- [ ] Pertimbangkan pindah ke object storage kalau nanti deploy
