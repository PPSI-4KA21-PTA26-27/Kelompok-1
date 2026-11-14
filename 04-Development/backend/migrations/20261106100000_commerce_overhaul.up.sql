-- Perombakan alur bisnis toko:
-- 1. Jejak waktu (created_at/updated_at) untuk user, produk, order, dan banner
--    supaya admin dan pelanggan bisa melihat kapan data dibuat.
-- 2. Verifikasi email wajib sebelum pelanggan boleh memesan.
-- 3. Pembayaran hanya QRIS statis dengan batas waktu bayar.
-- 4. Banner punya penempatan (hero/main/catalog) dan jadwal tayang.
-- 5. Jurnal/blog untuk copywriter, bisa ditautkan ke produk.

ALTER TABLE users
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- Akun yang dibuat sebelum verifikasi email diwajibkan dianggap sudah
-- terverifikasi supaya admin dan pelanggan lama tidak terkunci.
UPDATE users SET email_verified_at = now() WHERE email_verified_at IS NULL;

ALTER TABLE users
    ADD CONSTRAINT ck_users_role CHECK (role IN ('customer', 'admin', 'copywriter')),
    ADD CONSTRAINT ck_users_status CHECK (status IN ('active', 'inactive'));

CREATE TABLE email_verifications (
    token_hash TEXT        NOT NULL,
    user_id    UUID        NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT pk_email_verifications PRIMARY KEY (token_hash),
    CONSTRAINT fk_email_verifications_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE
);
CREATE INDEX idx_email_verifications_user ON email_verifications (user_id);

ALTER TABLE products
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now();
CREATE INDEX idx_products_status_created ON products (status, created_at DESC, id);

ALTER TABLE orders
    ADD COLUMN created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN payment_due_at TIMESTAMPTZ,
    ADD COLUMN cancel_reason  TEXT;
UPDATE orders SET payment_due_at = created_at + INTERVAL '24 hours'
 WHERE order_status = 'pending_payment';
CREATE INDEX idx_orders_status_due ON orders (order_status, payment_due_at);
CREATE INDEX idx_orders_created ON orders (created_at DESC);

ALTER TABLE payments
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE shipments
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE banners
    ADD COLUMN placement        TEXT NOT NULL DEFAULT 'hero',
    ADD COLUMN title            TEXT,
    ADD COLUMN subtitle         TEXT,
    ADD COLUMN cta_label        TEXT,
    ADD COLUMN mobile_image_url TEXT,
    ADD COLUMN starts_at        TIMESTAMPTZ,
    ADD COLUMN ends_at          TIMESTAMPTZ,
    ADD COLUMN created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD CONSTRAINT ck_banners_placement CHECK (placement IN ('hero', 'main', 'catalog')),
    ADD CONSTRAINT ck_banners_schedule CHECK (starts_at IS NULL OR ends_at IS NULL OR ends_at > starts_at);
CREATE INDEX idx_banners_placement ON banners (placement, is_active, sort_order);

ALTER TABLE site_settings
    ADD COLUMN qris_image_url     TEXT,
    ADD COLUMN qris_merchant_name TEXT,
    ADD COLUMN qris_nmid          TEXT,
    ADD COLUMN payment_window_hours INT NOT NULL DEFAULT 24,
    ADD CONSTRAINT ck_site_settings_payment_window CHECK (payment_window_hours BETWEEN 1 AND 168);

CREATE TABLE posts (
    id              UUID        NOT NULL DEFAULT gen_random_uuid(),
    slug            TEXT        NOT NULL,
    title           TEXT        NOT NULL,
    excerpt         TEXT,
    content         TEXT        NOT NULL DEFAULT '',
    cover_image_url TEXT,
    tag             TEXT,
    status          TEXT        NOT NULL DEFAULT 'draft',
    author_id       UUID,
    published_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT pk_posts PRIMARY KEY (id),
    CONSTRAINT uq_posts_slug UNIQUE (slug),
    CONSTRAINT fk_posts_author FOREIGN KEY (author_id)
        REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT ck_posts_status CHECK (status IN ('draft', 'published', 'archived'))
);
CREATE INDEX idx_posts_status_published ON posts (status, published_at DESC);

CREATE TABLE post_products (
    post_id    UUID NOT NULL,
    product_id UUID NOT NULL,
    sort_order INT  NOT NULL DEFAULT 0,

    CONSTRAINT pk_post_products PRIMARY KEY (post_id, product_id),
    CONSTRAINT fk_post_products_post FOREIGN KEY (post_id)
        REFERENCES posts (id) ON DELETE CASCADE,
    CONSTRAINT fk_post_products_product FOREIGN KEY (product_id)
        REFERENCES products (id) ON DELETE CASCADE
);
CREATE INDEX idx_post_products_product ON post_products (product_id);
