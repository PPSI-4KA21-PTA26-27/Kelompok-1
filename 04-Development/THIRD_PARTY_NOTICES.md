# Third-Party Notices — Hii Stranger

Proyek ini **proprietary** (lihat `LICENSE`). Daftar di bawah adalah komponen pihak
ketiga yang dipakai beserta lisensinya. Semuanya lisensi *permissive* yang boleh
dipakai di produk komersial tertutup, dengan syarat utama: **simpan teks lisensi &
copyright aslinya** (sudah ada di `node_modules/*/LICENSE` dan module cache Go).
Hasil audit 6 Nov 2026 (`go list -m all`, `pnpm licenses list`).

## Backend (Go) — ikut ter-compile ke binary
| Paket | Lisensi |
|---|---|
| github.com/jackc/pgx/v5 (+ pgpassfile, pgservicefile, puddle) | MIT |
| github.com/golang-jwt/jwt/v5 | MIT |
| github.com/joho/godotenv | MIT |
| github.com/redis/go-redis/v9 | BSD-2-Clause |
| github.com/prometheus/client_golang, client_model, common, procfs | Apache-2.0 |
| github.com/klauspost/compress | Apache-2.0 / BSD |
| github.com/beorn7/perks, cespare/xxhash, go.uber.org/atomic | MIT |
| github.com/munnerz/goautoneg, google.golang.org/protobuf | BSD-3-Clause |
| golang.org/x/crypto, x/sys, x/text, x/sync | BSD-3-Clause |
| Go standard library | BSD-3-Clause |

Apache-2.0: bila binary didistribusikan ke pihak lain, sertakan file NOTICE paket
tersebut. Untuk pemakaian server internal tidak ada kewajiban tambahan.

## Frontend (runtime)
| Paket | Lisensi | Catatan |
|---|---|---|
| next, react, react-dom | MIT | |
| @reduxjs/toolkit, react-redux | MIT | |
| react-hot-toast | MIT | |
| lucide-react | ISC | ikon |
| sharp | Apache-2.0 | dipasang otomatis oleh Next.js |
| @img/sharp-* (libvips) | **LGPL-3.0** + Apache-2.0 | boleh dipakai di software tertutup selama libvips dipakai sebagai library terpisah tanpa dimodifikasi (kondisi bawaan npm). Jangan menyalin/memodifikasi kodenya ke dalam proyek. |
| caniuse-lite | CC-BY-4.0 | data browser saat build; cukup atribusi (file ini). |

## Frontend (hanya saat development/build, tidak ikut dikirim ke pengguna)
MIT/ISC/BSD/Apache mayoritas. Pengecualian yang tetap aman:
`lightningcss`, `axe-core` (MPL-2.0 — copyleft per file, hanya berlaku bila file
mereka dimodifikasi), `argparse` (Python-2.0), `language-subtag-registry` (CC0).

## Font
| Font | Lisensi | Catatan |
|---|---|---|
| Geist (Vercel) | SIL Open Font License 1.1 | boleh komersial; font tidak boleh dijual terpisah |
| Bricolage Grotesque | SIL Open Font License 1.1 | sama |

Font **Euclid Circular A** dari template lama sudah dihapus — itu font komersial
(Swiss Typefaces) yang tidak punya lisensi di proyek ini.

## Yang TIDAK boleh ada di repositori
- `frontend/rules-direction/` — salinan source Next.js (MIT, milik Vercel) dan
  "agent-skills" tanpa file lisensi. Hanya bahan referensi lokal; diabaikan git.
- Aset template lama (NextMerce, logo Next.js/Vercel) — sudah dihapus.
- File `.env` berisi kredensial — diabaikan git.

## Layanan pihak ketiga (tunduk pada Terms of Service, bukan lisensi kode)
RajaOngkir/Komerce (ongkir & wilayah), Cloudinary (gambar), Gmail SMTP (email),
Google Fonts (diunduh saat build), cekresi.com & situs kurir (hanya ditautkan).
Untuk volume email produksi disarankan Google Workspace atau layanan email
transaksional (Gmail pribadi dibatasi ±500 email/hari).
