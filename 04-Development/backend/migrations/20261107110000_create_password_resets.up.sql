-- Token reset password (lupa password). Yang disimpan hanya hash SHA-256
-- tokennya; tautan mentah hanya ada di email pengguna. Sekali pakai.
CREATE TABLE password_resets (
    token_hash TEXT        NOT NULL,
    user_id    UUID        NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT pk_password_resets PRIMARY KEY (token_hash),
    CONSTRAINT fk_password_resets_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE
);
CREATE INDEX idx_password_resets_user ON password_resets (user_id);
