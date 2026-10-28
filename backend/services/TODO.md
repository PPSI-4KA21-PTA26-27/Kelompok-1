# services

Aturan bisnis aplikasi. Memanggil repository, dipanggil handlers.

- [ ] AuthService: register, login, JWT, lupa/reset kata sandi, hash password
- [ ] CatalogService: produk aktif, varian, label pre-order
- [ ] CartService: tambah/ubah/hapus item, cek stok
- [ ] OrderService: hitung subtotal + ongkir, generate kode unik (+Rp1 s/d +Rp999) di server, buat order, ubah status + catat log
- [ ] Pre-order: kalau ada item pre-order, seluruh item dikirim dalam satu paket (sesuai desain checkout)
- [ ] StockService: mutasi stok (restock/koreksi/cacat), catatan wajib, stok tidak boleh minus
- [ ] ShippingService: tarif per kurir + provinsi
- [ ] PaymentService: terima bukti bayar, verifikasi, ubah order Awaiting Payment -> Processing
- [ ] Subsidi QRIS Rp10.000 (bisa diaktif/nonaktifkan admin)
- [ ] Mode maintenance + notifikasi pelanggan terdaftar
- [ ] Ringkasan & analitik admin
- [ ] WaitlistService / akses VIP (kalau Drop/VIP masuk scope)
