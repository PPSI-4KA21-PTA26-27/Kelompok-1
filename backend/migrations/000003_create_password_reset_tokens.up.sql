-- Token lupa kata sandi (desain: lupa_kata_sandi, tautan berlaku 15 menit).
-- Yang disimpan adalah hash SHA-256 token, bukan token aslinya, supaya kebocoran database
-- tidak langsung bisa dipakai untuk mereset akun orang.
CREATE TABLE password_reset_tokens (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash CHAR(64) UNIQUE NOT NULL,
  expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
  used_at TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens (user_id);
