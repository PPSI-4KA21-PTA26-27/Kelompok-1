-- Data uji lokal (bukan migrasi). Jalankan sekali: psql "$DB_URL" -f scripts/seed.sql
-- Semua akun berkata sandi: Password123
BEGIN;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO users (name, email, password_hash, phone, role_id)
SELECT v.name, v.email, crypt('Password123', gen_salt('bf', 10)), v.phone, r.id
FROM (VALUES
  ('Owner Hii Stranger', 'owner@hiistranger.test',  '081200000001', 'owner'),
  ('Staff Hii Stranger', 'staff@hiistranger.test',  '081200000002', 'staff'),
  ('Budi Pelanggan',     'budi@hiistranger.test',   '081200000003', 'customer'),
  ('Sari Pelanggan',     'sari@hiistranger.test',   '081200000004', 'customer')
) v(name, email, phone, code) JOIN roles r ON r.code = v.code;

INSERT INTO categories (name) VALUES ('Kaos'), ('Hoodie'), ('Celana'), ('Aksesoris');

INSERT INTO products (category_id, name, description, price, discount_percent, weight_gram, availability_type, po_estimate_days, status)
SELECT c.id, p.name, p.descr, p.price, p.disc, p.weight, p.avail, p.po, 'active'
FROM (VALUES
  ('Kaos',      'Kaos Oversize Stranger Hitam', 'Kaos cotton combed 24s, potongan oversize.',  129000, 0,  250, 'ready',    NULL),
  ('Hoodie',    'Hoodie Zipper Stranger Abu',   'Hoodie fleece tebal dengan zipper YKK.',       289000, 10, 700, 'ready',    NULL),
  ('Celana',    'Cargo Pants Stranger Olive',   'Celana cargo 6 saku, bahan ripstop.',          249000, 0,  550, 'ready',    NULL),
  ('Aksesoris', 'Topi Baseball Stranger',       'Topi baseball bordir logo, strap adjustable.',  89000, 0,  120, 'ready',    NULL),
  ('Kaos',      'Kaos Limited Edition Rilisan Baru', 'Rilisan terbatas, produksi setelah pesanan.', 159000, 0, 250, 'preorder', 14)
) p(cat, name, descr, price, disc, weight, avail, po) JOIN categories c ON c.name = p.cat;

INSERT INTO product_variants (product_id, sku, size, color, stock)
SELECT p.id, v.sku, v.size, v.color, v.stock
FROM (VALUES
  ('Kaos Oversize Stranger Hitam', 'KOS-HTM-M',  'M',  'Hitam', 20),
  ('Kaos Oversize Stranger Hitam', 'KOS-HTM-L',  'L',  'Hitam', 15),
  ('Kaos Oversize Stranger Hitam', 'KOS-HTM-XL', 'XL', 'Hitam', 10),
  ('Hoodie Zipper Stranger Abu',   'HZS-ABU-M',  'M',  'Abu',   8),
  ('Hoodie Zipper Stranger Abu',   'HZS-ABU-L',  'L',  'Abu',   12),
  ('Cargo Pants Stranger Olive',   'CPS-OLV-30', '30', 'Olive', 10),
  ('Cargo Pants Stranger Olive',   'CPS-OLV-32', '32', 'Olive', 10),
  ('Topi Baseball Stranger',       'TBS-HTM-ALL','All Size', 'Hitam', 30),
  ('Kaos Limited Edition Rilisan Baru', 'KLE-PTH-L', 'L', 'Putih', 0)
) v(pname, sku, size, color, stock) JOIN products p ON p.name = v.pname;

-- Foto lokal di frontend/public/products. Cargo & Topi belum ada fotonya -> placeholder picsum.
INSERT INTO product_images (product_id, image_url, sort_order, is_primary)
SELECT p.id, i.url, i.ord, i.ord = 0
FROM (VALUES
  ('Kaos Oversize Stranger Hitam', '/products/img_6151.jpg', 0),
  ('Kaos Oversize Stranger Hitam', '/products/img_6157.jpg', 1),
  ('Kaos Oversize Stranger Hitam', '/products/img_6142.jpg', 2),
  ('Kaos Oversize Stranger Hitam', '/products/size-chart-1.jpg', 3),
  ('Hoodie Zipper Stranger Abu',   '/products/dsc00522.jpg', 0),
  ('Hoodie Zipper Stranger Abu',   '/products/dsc00525.jpg', 1),
  ('Hoodie Zipper Stranger Abu',   '/products/dsc00527.jpg', 2),
  ('Hoodie Zipper Stranger Abu',   '/products/dsc00542.jpg', 3),
  ('Kaos Limited Edition Rilisan Baru', '/products/img_4378.jpg', 0),
  ('Kaos Limited Edition Rilisan Baru', '/products/img_4380.jpg', 1)
) i(pname, url, ord) JOIN products p ON p.name = i.pname
UNION ALL
SELECT id, 'https://picsum.photos/seed/' || replace(lower(name), ' ', '-') || '/600/600', 0, true
FROM products WHERE name IN ('Cargo Pants Stranger Olive', 'Topi Baseball Stranger');

-- Pesanan 1: menunggu pembayaran (Budi)
-- Pesanan 2: sudah dibayar, bukti valid (Sari)
-- Pesanan 3: sudah dikirim, ada resi (Budi)
INSERT INTO orders (order_number, user_id, recipient_name, recipient_phone, shipping_address, shipping_city, shipping_kelurahan,
                    subtotal, shipping_cost, total_amount, order_status, shipping_district, shipping_subdistrict_id, shipping_zip_code,
                    courier_code, courier_service, shipping_etd, shipping_weight_gram, created_at, payment_due_at)
SELECT o.no, u.id, u.name, u.phone, o.addr, o.city, o.kel, o.sub, o.ship, o.sub + o.ship, o.st, o.kec, o.sid, o.zip,
       'pos', 'Paket Kilat Khusus', '2-3 hari', o.wt, now() - o.age, now() - o.age + interval '1 day'
FROM (VALUES
  ('HS-20261120-0001', 'budi@hiistranger.test', 'Jl. Merdeka No. 10', 'Bandung',  'Braga',      'Sumur Bandung', 1001, '40111', 129000 + 89000,  18000, 'pending_payment', 370,  interval '1 hour'),
  ('HS-20261118-0002', 'sari@hiistranger.test', 'Jl. Melati No. 5',   'Jakarta Selatan', 'Tebet', 'Tebet',       1002, '12810', 260100,          22000, 'paid',            700,  interval '3 days'),
  ('HS-20261112-0003', 'budi@hiistranger.test', 'Jl. Merdeka No. 10', 'Bandung',  'Braga',      'Sumur Bandung', 1001, '40111', 249000,          18000, 'shipped',         550,  interval '10 days')
) o(no, email, addr, city, kel, kec, sid, zip, sub, ship, st, wt, age) JOIN users u ON u.email = o.email
WHERE o.kec IS NOT NULL;

INSERT INTO order_items (order_id, product_id, variant_id, product_name, size, color, unit_price, quantity, subtotal)
SELECT o.id, p.id, v.id, p.name, v.size, v.color, i.price, i.qty, i.price * i.qty
FROM (VALUES
  ('HS-20261120-0001', 'KOS-HTM-L',   129000, 1),
  ('HS-20261120-0001', 'TBS-HTM-ALL',  89000, 1),
  ('HS-20261118-0002', 'HZS-ABU-M',   260100, 1),
  ('HS-20261112-0003', 'CPS-OLV-32',  249000, 1)
) i(no, sku, price, qty)
JOIN orders o ON o.order_number = i.no
JOIN product_variants v ON v.sku = i.sku
JOIN products p ON p.id = v.product_id;

INSERT INTO payments (order_id, method, amount, unique_code, status, paid_at, reviewed_by, reviewed_at, proof_url, created_at)
SELECT o.id, 'qris', o.total_amount + x.code, x.code, x.st,
       CASE WHEN x.st = 'valid' THEN o.created_at + interval '1 hour' END,
       CASE WHEN x.st = 'valid' THEN (SELECT id FROM users WHERE email = 'staff@hiistranger.test') END,
       CASE WHEN x.st = 'valid' THEN o.created_at + interval '2 hours' END,
       CASE WHEN x.st = 'valid' THEN 'https://picsum.photos/seed/bukti-' || o.order_number || '/400/600.png' END,
       o.created_at
FROM (VALUES ('HS-20261120-0001', 123, 'menunggu'), ('HS-20261118-0002', 217, 'valid'), ('HS-20261112-0003', 88, 'valid')) x(no, code, st)
JOIN orders o ON o.order_number = x.no;

INSERT INTO shipments (order_id, tracking_number, coverage_status, shipping_status)
SELECT id, 'P2611120003ID', 'covered', 'shipped' FROM orders WHERE order_number = 'HS-20261112-0003';

-- Keranjang Sari: 1 kaos
INSERT INTO cart_items (user_id, variant_id, quantity)
SELECT u.id, v.id, 2 FROM users u, product_variants v WHERE u.email = 'sari@hiistranger.test' AND v.sku = 'KOS-HTM-M';
COMMIT;
