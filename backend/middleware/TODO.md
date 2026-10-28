# middleware

Urusan lintas-request: logging, recover dari panic, dan nanti auth/CORS/rate limit.

- [x] Logger, Recover
- [ ] Auth JWT (verifikasi token, simpan user id + role di context)
- [ ] RequireAdmin (cek role di server, bukan di frontend)
- [ ] CORS (origin frontend Next.js)
- [ ] Rate limit untuk login, register, lupa kata sandi
- [ ] Batas ukuran request body (terutama upload bukti bayar)
- [ ] Mode maintenance: blokir endpoint toko untuk pelanggan saat aktif, admin tetap boleh
