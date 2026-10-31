CREATE TABLE shipping_rates (
    id                   UUID NOT NULL DEFAULT gen_random_uuid(),
    origin_region        TEXT NOT NULL,
    destination_province TEXT NOT NULL,
    destination_city     TEXT,
    cost                 NUMERIC NOT NULL,

    CONSTRAINT pk_shipping_rates PRIMARY KEY (id),
    CONSTRAINT ck_shipping_rates_cost CHECK (cost >= 0)
);
