# routes

Daftar semua endpoint (prefix `/api/v1`) dan middleware global dipasang di sini.

- [x] GET /api/v1/health
- [x] /auth/* (publik)
- [x] /products, /categories, /banners (publik), /uploads/{name} (gambar)
- [ ] /cart, /orders, /orders/{id}/payment (butuh login). /me sudah ada
- [ ] /store/status (publik, untuk halaman maintenance)
- [ ] /admin/* (butuh login + role admin)
- [ ] /waitlist (kalau Drop/VIP masuk scope)
