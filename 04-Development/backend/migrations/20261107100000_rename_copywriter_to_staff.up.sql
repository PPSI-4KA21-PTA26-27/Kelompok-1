-- Role copywriter diganti staff: staff mengelola operasional toko (katalog,
-- pesanan, pembayaran, banner, jurnal). Pengaturan toko & pengguna tetap admin.
ALTER TABLE users DROP CONSTRAINT IF EXISTS ck_users_role;
UPDATE users SET role = 'staff' WHERE role = 'copywriter';
ALTER TABLE users
    ADD CONSTRAINT ck_users_role CHECK (role IN ('customer', 'admin', 'staff'));
