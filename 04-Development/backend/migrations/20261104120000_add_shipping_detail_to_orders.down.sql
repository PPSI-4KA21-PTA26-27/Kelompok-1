CREATE TABLE shipping_rates (
    id                   UUID NOT NULL DEFAULT gen_random_uuid(),
    origin_region        TEXT NOT NULL,
    destination_province TEXT NOT NULL,
    destination_city     TEXT,
    cost                 NUMERIC NOT NULL,

    CONSTRAINT pk_shipping_rates PRIMARY KEY (id),
    CONSTRAINT ck_shipping_rates_cost CHECK (cost >= 0)
);

ALTER TABLE orders
    DROP CONSTRAINT IF EXISTS ck_orders_shipping_weight,
    DROP COLUMN IF EXISTS shipping_province,
    DROP COLUMN IF EXISTS shipping_district,
    DROP COLUMN IF EXISTS shipping_subdistrict_id,
    DROP COLUMN IF EXISTS shipping_zip_code,
    DROP COLUMN IF EXISTS courier_code,
    DROP COLUMN IF EXISTS courier_service,
    DROP COLUMN IF EXISTS shipping_etd,
    DROP COLUMN IF EXISTS shipping_weight_gram;
