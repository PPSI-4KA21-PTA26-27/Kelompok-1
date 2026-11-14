ALTER TABLE shipments ADD COLUMN courier_name TEXT NOT NULL DEFAULT '';
UPDATE shipments s
   SET courier_name = upper(o.courier_code) || ' ' || o.courier_service
  FROM orders o WHERE o.id = s.order_id;
ALTER TABLE shipments ALTER COLUMN courier_name DROP DEFAULT;
ALTER TABLE orders ADD COLUMN shipping_province TEXT;