CREATE TABLE shipping_rates (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    origin_region        TEXT NOT NULL,
    destination_province TEXT NOT NULL,
    destination_city     TEXT,
    cost                 NUMERIC NOT NULL
);
