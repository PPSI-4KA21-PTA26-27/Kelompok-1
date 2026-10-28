# repository

Akses database saja: query SQL (parameterized) lewat `db.DB`. Tidak ada aturan bisnis di sini.

- [x] UserRepository (create, cari by email/id, update profil, update password)
- [x] ResetTokenRepository
- [ ] CategoryRepository, ProductRepository (list + filter + pagination, detail + varian), BannerRepository
- [ ] StockRepository (kurangi/tambah stok + catat mutasi dalam satu transaksi)
- [ ] CartRepository
- [ ] ShippingRateRepository
- [ ] OrderRepository (buat order + item dalam satu transaksi, list per user, list admin + filter status, detail)
- [ ] PaymentRepository
- [ ] SettingRepository
- [ ] Query analitik/ringkasan admin (omzet, jumlah order, produk terlaris)
- [ ] WaitlistRepository (kalau Drop/VIP masuk scope)
