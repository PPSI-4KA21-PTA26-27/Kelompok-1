-- Index untuk query yang paling sering dipanggil (hasil EXPLAIN ANALYZE saat
-- benchmark). Index lama yang sudah tercakup index baru dihapus supaya
-- INSERT/UPDATE tidak memelihara index ganda.

-- Daftar produk publik: WHERE status = 'active' ORDER BY name.
CREATE INDEX idx_products_status_name ON products (status, name, id);

-- Daftar produk admin (tanpa filter status) yang diurutkan nama: tanpa index
-- ini PostgreSQL harus mengurutkan seluruh tabel di setiap halaman.
CREATE INDEX idx_products_name ON products (name, id);

-- Pencarian nama produk (ILIKE '%kata%') memakai trigram. pg_trgm termasuk
-- extension "trusted", jadi pemilik database boleh membuatnya.
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_products_name_trgm ON products USING gin (name gin_trgm_ops);

-- Gambar utama per produk: WHERE product_id = ? ORDER BY is_primary DESC, sort_order.
CREATE INDEX idx_product_images_primary ON product_images (product_id, is_primary DESC, sort_order, id);
DROP INDEX IF EXISTS idx_product_images_product_id;

-- Daftar order milik user, terbaru dulu.
CREATE INDEX idx_orders_user_number ON orders (user_id, order_number DESC);
DROP INDEX IF EXISTS idx_orders_user_id;

-- Antrean verifikasi pembayaran admin (filter status).
CREATE INDEX idx_payments_status ON payments (status);
