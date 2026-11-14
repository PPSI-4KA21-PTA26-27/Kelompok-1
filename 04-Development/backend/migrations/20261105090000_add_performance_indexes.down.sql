DROP INDEX IF EXISTS idx_payments_status;

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders (user_id);
DROP INDEX IF EXISTS idx_orders_user_number;

CREATE INDEX IF NOT EXISTS idx_product_images_product_id ON product_images (product_id);
DROP INDEX IF EXISTS idx_product_images_primary;

DROP INDEX IF EXISTS idx_products_name_trgm;
DROP INDEX IF EXISTS idx_products_name;
DROP INDEX IF EXISTS idx_products_status_name;
-- Extension pg_trgm sengaja tidak di-drop karena bisa dipakai objek lain.
