-- Baris yang username-nya kosong diberi nilai sementara agar NOT NULL bisa dipasang lagi.
UPDATE users SET username = 'user' || id WHERE username IS NULL;

ALTER TABLE users
  ALTER COLUMN username SET NOT NULL,
  DROP COLUMN role,
  DROP COLUMN phone,
  DROP COLUMN full_name;
