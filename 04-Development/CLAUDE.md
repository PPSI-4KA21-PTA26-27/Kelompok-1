# capstone_hi — panduan kerja Claude Code

Monorepo satu git: `backend/` (Go, API) dan `frontend/` (Next.js 16, pnpm). Kontrak API: `backend/api/openapi.yaml`.

## Satu agent: `dev`
Default sesi (`.claude/settings.json` -> `"agent": "dev"`). Semua aturan kerja ada di `.claude/agents/dev.md`: ponytail (lazy senior engineer, kode singkat) di atas context-mode, selalu aktif tanpa dipanggil.

Aturan inti:
- Semua kerja file/perintah lewat MCP context-mode; Bash, Grep, Glob, Edit, Write dilarang, Read hanya gambar/PDF.
- Output maksimal 2500 karakter per panggilan; penjelasan maksimal 8 baris.
- Model sonnet, effort medium. Sub-agent ringan (cari/baca/rangkum) pakai `haiku`, tugas berat (arsitektur, debug lintas modul, review keamanan) boleh `opus`. Maksimal 3 paralel; tugas kecil kerjakan sendiri.
- Satu build/test/lint berat sekali jalan.
- Jangan commit/push kecuali diminta.
- Perintah verifikasi tersaring ada di `dev.md`.
