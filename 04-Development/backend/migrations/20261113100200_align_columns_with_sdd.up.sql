-- Menyelaraskan kolom dan kunci dengan kamus data SDD revisi 13-11-2026
-- (Bagian 5.3). Kolom banners.title (boleh kosong), orders.shipping_province,
-- dan shipments.courier_name sengaja dipertahankan karena dipakai aplikasi.

-- created_at pada tabel yang belum memilikinya.
ALTER TABLE categories       ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE product_variants ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE product_images   ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE cart_items       ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE order_items      ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE shipments        ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- cart_items: satu baris per pasangan pengguna + varian (D-16).
ALTER TABLE cart_items ADD CONSTRAINT uq_cart_items_user_variant UNIQUE (user_id, variant_id);

-- order_items: salinan warna varian saat pesanan dibuat (D-14).
ALTER TABLE order_items ADD COLUMN color TEXT;
UPDATE order_items oi SET color = v.color FROM product_variants v WHERE v.id = oi.variant_id;

-- Token verifikasi email dan reset password: id sebagai PK, hash token unik.
ALTER TABLE email_verifications
    DROP CONSTRAINT pk_email_verifications,
    ADD COLUMN id UUID NOT NULL DEFAULT gen_random_uuid(),
    ADD CONSTRAINT pk_email_verifications PRIMARY KEY (id),
    ADD CONSTRAINT uq_email_verifications_token_hash UNIQUE (token_hash);
ALTER TABLE password_resets
    DROP CONSTRAINT pk_password_resets,
    ADD COLUMN id UUID NOT NULL DEFAULT gen_random_uuid(),
    ADD CONSTRAINT pk_password_resets PRIMARY KEY (id),
    ADD CONSTRAINT uq_password_resets_token_hash UNIQUE (token_hash);

-- orders: data pengiriman wajib diisi.
ALTER TABLE orders
    ALTER COLUMN shipping_city SET NOT NULL,
    ALTER COLUMN shipping_district SET NOT NULL,
    ALTER COLUMN shipping_subdistrict_id SET NOT NULL,
    ALTER COLUMN courier_code SET NOT NULL,
    ALTER COLUMN courier_service SET NOT NULL,
    ALTER COLUMN shipping_weight_gram SET NOT NULL;

-- posts: setiap artikel wajib punya penulis.
ALTER TABLE posts
    DROP CONSTRAINT fk_posts_author,
    ALTER COLUMN author_id SET NOT NULL,
    ADD CONSTRAINT fk_posts_author FOREIGN KEY (author_id)
        REFERENCES users (id) ON DELETE RESTRICT;

-- site_settings: id UUID; tabel tetap satu baris lewat indeks unik konstan.
ALTER TABLE site_settings
    DROP CONSTRAINT pk_site_settings,
    DROP COLUMN id,
    ADD COLUMN id UUID NOT NULL DEFAULT gen_random_uuid(),
    ADD CONSTRAINT pk_site_settings PRIMARY KEY (id),
    ALTER COLUMN maintenance_message DROP NOT NULL;
CREATE UNIQUE INDEX uq_site_settings_singleton ON site_settings ((true));