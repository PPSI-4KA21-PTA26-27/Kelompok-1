# Security

Semua pengecekan keamanan backend ada di dalam folder `backend` ini dan dijalankan lokal.

## Scanning

| Tool | Fungsi | Perintah |
|---|---|---|
| golangci-lint (termasuk gosec) | Static analysis dan scan celah keamanan di kode: hardcoded secret, SQL injection, crypto lemah, error yang tidak dicek | `make lint` |
| govulncheck | Cek dependency yang punya celah keamanan yang sudah diketahui, hanya yang benar-benar dipakai kodenya | `make vuln` |
| go vet, build, test | Cek dasar kode | `make check` |

Konfigurasi linter ada di `.golangci.yml`.

### Cara jalanin

Windows tanpa `make`, cukup satu perintah (tool di-install otomatis kalau belum ada):

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\check.ps1
```

Kalau sudah punya `make`:

```bash
make tools    # install tool, sekali saja
make check    # build + test + lint + vuln
```

Jalankan sebelum commit atau push. Kalau ada temuan, perbaiki dulu atau catat alasannya kalau memang dibiarkan (false positive).

## Proteksi yang sudah ada di kode

- Password di-hash dengan bcrypt, tidak disimpan polos dan tidak pernah dikirim di response (`json:"-"`)
- JWT ditandatangani HS256 dan metode signing diverifikasi saat parsing
- Pesan error login sama untuk email salah dan password salah, supaya email tidak bisa ditebak
- Role-based access (`RequireRole`) untuk endpoint admin
- Register selalu membuat role `user`, tidak bisa memilih role sendiri
- Validasi input lewat tag `binding` di DTO, dan batas `limit` pagination maksimal 100
- Query lewat GORM dengan parameter binding, bukan string concatenation
- Header pengaman di semua response (`internal/middleware/security.go`): `X-Content-Type-Options`, `X-Frame-Options`, `Content-Security-Policy`, `Referrer-Policy`, `Cache-Control: no-store`
- Aplikasi menolak start di production kalau `JWT_SECRET` default atau kurang dari 32 karakter (`internal/config/config.go`)
- Timeout server (`ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`) untuk mengurangi risiko slowloris
- `.env` masuk `.gitignore`, hanya `.env.example` yang boleh di-commit

## Belum ada (bisa ditambah kalau dibutuhkan)

- Rate limiting dan pembatasan percobaan login (brute force)
- Refresh token dan logout/revoke token
- Batas ukuran body request
- Audit log
- Unit test untuk logic auth

## Checklist sebelum deploy

- [ ] `APP_ENV=production`
- [ ] `JWT_SECRET` acak minimal 32 karakter
- [ ] Password database kuat dan bukan default
- [ ] `DB_SSLMODE` bukan `disable` kalau database di luar mesin yang sama
- [ ] `CORS_ORIGINS` hanya berisi domain frontend yang sebenarnya
- [ ] `make check` lolos
- [ ] Tidak ada file `.env` yang ikut ter-commit

## Catatan tentang GitHub

Workflow GitHub Actions hanya bisa jalan kalau filenya ada di `.github/workflows/` di root repo, tidak bisa dari dalam subfolder. Karena itu scanning di sini dibuat lokal. Fitur bawaan GitHub seperti Dependabot alerts dan secret scanning diatur lewat Settings repo, bukan lewat folder ini.
