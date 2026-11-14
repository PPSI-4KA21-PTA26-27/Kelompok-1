DROP TABLE IF EXISTS post_products;
DROP TABLE IF EXISTS posts;

ALTER TABLE site_settings
    DROP CONSTRAINT IF EXISTS ck_site_settings_payment_window,
    DROP COLUMN IF EXISTS payment_window_hours,
    DROP COLUMN IF EXISTS qris_nmid,
    DROP COLUMN IF EXISTS qris_merchant_name,
    DROP COLUMN IF EXISTS qris_image_url;

DROP INDEX IF EXISTS idx_banners_placement;
ALTER TABLE banners
    DROP CONSTRAINT IF EXISTS ck_banners_schedule,
    DROP CONSTRAINT IF EXISTS ck_banners_placement,
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS ends_at,
    DROP COLUMN IF EXISTS starts_at,
    DROP COLUMN IF EXISTS mobile_image_url,
    DROP COLUMN IF EXISTS cta_label,
    DROP COLUMN IF EXISTS subtitle,
    DROP COLUMN IF EXISTS title,
    DROP COLUMN IF EXISTS placement;

ALTER TABLE shipments DROP COLUMN IF EXISTS updated_at;

ALTER TABLE payments DROP COLUMN IF EXISTS created_at;

DROP INDEX IF EXISTS idx_orders_created;
DROP INDEX IF EXISTS idx_orders_status_due;
ALTER TABLE orders
    DROP COLUMN IF EXISTS cancel_reason,
    DROP COLUMN IF EXISTS payment_due_at,
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS created_at;

DROP INDEX IF EXISTS idx_products_status_created;
ALTER TABLE products DROP COLUMN IF EXISTS created_at;

DROP TABLE IF EXISTS email_verifications;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS ck_users_status,
    DROP CONSTRAINT IF EXISTS ck_users_role,
    DROP COLUMN IF EXISTS created_at;
