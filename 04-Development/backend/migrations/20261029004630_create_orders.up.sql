CREATE TABLE orders (
    id                 UUID NOT NULL DEFAULT gen_random_uuid(),
    order_number       TEXT NOT NULL,
    user_id            UUID NOT NULL,
    recipient_name     TEXT NOT NULL,
    recipient_phone    TEXT NOT NULL,
    shipping_address   TEXT NOT NULL,
    shipping_city      TEXT,
    shipping_kelurahan TEXT NOT NULL,
    subtotal           NUMERIC NOT NULL,
    shipping_cost      NUMERIC NOT NULL,
    total_amount       NUMERIC NOT NULL,
    order_status       TEXT NOT NULL,

    CONSTRAINT pk_orders PRIMARY KEY (id),
    CONSTRAINT uq_orders_order_number UNIQUE (order_number),
    CONSTRAINT fk_orders_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT ck_orders_subtotal CHECK (subtotal >= 0),
    CONSTRAINT ck_orders_shipping_cost CHECK (shipping_cost >= 0),
    CONSTRAINT ck_orders_total_amount CHECK (total_amount >= 0)
);

CREATE INDEX idx_orders_user_id ON orders (user_id);
