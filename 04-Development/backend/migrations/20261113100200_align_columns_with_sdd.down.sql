DROP INDEX IF EXISTS uq_site_settings_singleton;
UPDATE site_settings SET maintenance_message = '' WHERE maintenance_message IS NULL;
ALTER TABLE site_settings
    ALTER COLUMN maintenance_message SET NOT NULL,
    DROP CONSTRAINT pk_site_settings,
    DROP COLUMN id,
    ADD COLUMN id INT NOT NULL DEFAULT 1,
    ADD CONSTRAINT pk_site_settings PRIMARY KEY (id);
ALTER TABLE site_settings ALTER COLUMN id DROP DEFAULT;

ALTER TABLE posts
    DROP CONSTRAINT fk_posts_author,
    ALTER COLUMN author_id DROP NOT NULL,
    ADD CONSTRAINT fk_posts_author FOREIGN KEY (author_id)
        REFERENCES users (id) ON DELETE SET NULL;

ALTER TABLE orders
    ALTER COLUMN shipping_city DROP NOT NULL,
    ALTER COLUMN shipping_district DROP NOT NULL,
    ALTER COLUMN shipping_subdistrict_id DROP NOT NULL,
    ALTER COLUMN courier_code DROP NOT NULL,
    ALTER COLUMN courier_service DROP NOT NULL,
    ALTER COLUMN shipping_weight_gram DROP NOT NULL;

ALTER TABLE password_resets
    DROP CONSTRAINT uq_password_resets_token_hash,
    DROP CONSTRAINT pk_password_resets,
    DROP COLUMN id,
    ADD CONSTRAINT pk_password_resets PRIMARY KEY (token_hash);
ALTER TABLE email_verifications
    DROP CONSTRAINT uq_email_verifications_token_hash,
    DROP CONSTRAINT pk_email_verifications,
    DROP COLUMN id,
    ADD CONSTRAINT pk_email_verifications PRIMARY KEY (token_hash);

ALTER TABLE order_items DROP COLUMN color;
ALTER TABLE cart_items DROP CONSTRAINT uq_cart_items_user_variant;

ALTER TABLE shipments        DROP COLUMN created_at;
ALTER TABLE order_items      DROP COLUMN created_at;
ALTER TABLE cart_items       DROP COLUMN created_at;
ALTER TABLE product_images   DROP COLUMN created_at;
ALTER TABLE product_variants DROP COLUMN created_at;
ALTER TABLE categories       DROP COLUMN created_at;