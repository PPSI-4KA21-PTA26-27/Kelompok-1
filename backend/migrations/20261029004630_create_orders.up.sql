CREATE TABLE orders (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number       TEXT NOT NULL UNIQUE,
    user_id            UUID NOT NULL REFERENCES users(id),
    recipient_name     TEXT NOT NULL,
    recipient_phone    TEXT NOT NULL,
    shipping_address   TEXT NOT NULL,
    shipping_kelurahan TEXT NOT NULL,
    shipping_city      TEXT,
    subtotal           NUMERIC NOT NULL,
    shipping_cost      NUMERIC NOT NULL,
    total_amount       NUMERIC NOT NULL,
    order_status       TEXT NOT NULL
);
