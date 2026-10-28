# handlers

Urusan HTTP saja: baca request, validasi input, panggil service, kirim respons. Tidak ada aturan bisnis atau SQL di sini.

- [x] Health check
- [x] auth: daftar, masuk, lupa kata sandi, reset
- [x] akun: profil pelanggan, ganti kata sandi. Alamat tidak ada di server: desain menyimpannya di browser
- [ ] katalog: beranda, list produk, detail produk, kategori, banner
- [ ] keranjang
- [ ] checkout & order: buat order, daftar pesanan saya, detail pesanan
- [ ] pembayaran: data QRIS + total + kode unik, upload bukti bayar
- [ ] maintenance: status toko, daftar notifikasi
- [ ] admin: ringkasan, analitik
- [ ] admin: pesanan (list, detail, ubah status), pembayaran (verifikasi)
- [ ] admin: produk, stok + log mutasi, kategori, banner
- [ ] admin: ongkir (tarif per kurir + provinsi), subsidi QRIS, toggle maintenance
- [ ] waitlist / VIP drop (kalau masuk scope)
