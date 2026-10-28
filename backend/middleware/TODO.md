# middleware

Urusan lintas-request: logging, recover dari panic, dan nanti auth/CORS/rate limit.

- [x] Logger, Recover
- [x] Auth JWT (verifikasi token, simpan user id + role di context)
- [x] RequireAdmin (cek role di server, bukan di frontend)
- [ ] CORS (origin frontend Next.js)
- [x] Rate limit untuk login, register, lupa kata sandi (per IP, disimpan di memori)
- [ ] Batas ukuran request body (terutama upload bukti bayar)
- [ ] Mode maintenance: blokir endpoint toko untuk pelanggan saat aktif, admin tetap boleh
