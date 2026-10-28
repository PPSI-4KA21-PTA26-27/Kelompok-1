-- Melengkapi tabel users untuk fitur daftar/masuk (desain: daftar_akun, masuk_akun, akun_pelanggan).
-- Kolom username TIDAK dihapus (aturan: jangan drop kolom diam-diam), hanya dibuat boleh kosong,
-- karena form daftar di desain tidak punya field username.
ALTER TABLE users
  ADD COLUMN full_name VARCHAR(100) NOT NULL DEFAULT '',
  ADD COLUMN phone VARCHAR(20),
  ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'customer'
    CONSTRAINT users_role_check CHECK (role IN ('customer', 'admin')),
  ALTER COLUMN username DROP NOT NULL;

-- Default '' hanya untuk baris lama; akun baru wajib mengisi nama lengkap dari aplikasi.
ALTER TABLE users ALTER COLUMN full_name DROP DEFAULT;
