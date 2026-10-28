# migrations

File SQL skema database, berpasangan `.up.sql` dan `.down.sql`. Jangan edit migrasi yang sudah dijalankan; tambah file baru.

- [x] 000001 users
- [ ] users: tambah nama lengkap, no. HP/WhatsApp, role (customer/admin)
- [ ] password_reset_tokens (lupa kata sandi)
- [ ] categories
- [ ] products (nama, deskripsi, harga, pre-order + estimasi hari, status aktif)
- [ ] product_variants (warna, ukuran, SKU, stok)
- [ ] banners
- [ ] stock_movements (restock / koreksi / cacat, jumlah, catatan wajib, oleh siapa)
- [ ] carts + cart_items
- [ ] shipping_rates (kurir, provinsi, tarif dasar, estimasi hari)
- [ ] orders (status, alamat kirim, subtotal, ongkir, kode unik 1-999, total)
- [ ] order_items (snapshot nama/varian/harga saat order)
- [ ] order_status_logs (audit trail: siapa, kapan, dari status ke status)
- [ ] payments (bukti bayar, status verifikasi, diverifikasi oleh/kapan)
- [ ] settings (mode maintenance, subsidi QRIS Rp10.000 aktif/tidak)
- [ ] maintenance_subscribers (notifikasi saat toko buka lagi)
- [ ] waitlist (kontak WA/email, ukuran, drop) — kalau Drop/VIP masuk scope
