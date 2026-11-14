-- Role admin diganti owner: owner adalah pemilik toko dengan hak penuh
-- (kelola pengguna, pengaturan situs, analitik), staff tetap operasional.
ALTER TABLE users DROP CONSTRAINT IF EXISTS ck_users_role;
UPDATE users SET role = 'owner' WHERE role = 'admin';
ALTER TABLE users
    ADD CONSTRAINT ck_users_role CHECK (role IN ('customer', 'owner', 'staff'));
