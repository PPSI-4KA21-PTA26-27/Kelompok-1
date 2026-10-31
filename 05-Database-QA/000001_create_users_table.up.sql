CREATE TABLE users (
    id                UUID NOT NULL DEFAULT gen_random_uuid(),
    name              TEXT NOT NULL,
    email             TEXT NOT NULL,
    password_hash     TEXT NOT NULL,
    phone             TEXT,
    role              TEXT NOT NULL DEFAULT 'customer',
    status            TEXT NOT NULL DEFAULT 'active',
    email_verified_at TIMESTAMPTZ,

    CONSTRAINT pk_users PRIMARY KEY (id),
    CONSTRAINT uq_users_email UNIQUE (email)
);
