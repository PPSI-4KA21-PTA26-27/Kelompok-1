ALTER TABLE users DROP CONSTRAINT IF EXISTS ck_users_role;
UPDATE users SET role = 'copywriter' WHERE role = 'staff';
ALTER TABLE users
    ADD CONSTRAINT ck_users_role CHECK (role IN ('customer', 'admin', 'copywriter'));
