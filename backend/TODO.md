# TODO Backend — peta dari desain (`../ui-hii-stranger`)

Alur: handlers -> services -> repository -> db. Detail per folder ada di `TODO.md` masing-masing folder.

## Urutan kerja

1. [ ] Auth & akun — daftar_akun, masuk_akun, lupa_kata_sandi, akun_pelanggan
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
