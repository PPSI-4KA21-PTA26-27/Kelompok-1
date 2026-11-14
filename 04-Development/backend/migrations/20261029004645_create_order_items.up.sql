CREATE TABLE order_items (
    id           UUID NOT NULL DEFAULT gen_random_uuid(),
    order_id     UUID NOT NULL,
    product_id   UUID NOT NULL,
    variant_id   UUID NOT NULL,
    product_name TEXT NOT NULL,
    size         TEXT NOT NULL,
    unit_price   NUMERIC NOT NULL,
    quantity     INT NOT NULL,
    subtotal     NUMERIC NOT NULL,

    CONSTRAINT pk_order_items PRIMARY KEY (id),
    CONSTRAINT fk_order_items_order FOREIGN KEY (order_id)
        REFERENCES orders (id) ON DELETE CASCADE,
    CONSTRAINT fk_order_items_product FOREIGN KEY (product_id)
        REFERENCES products (id) ON DELETE RESTRICT,
    CONSTRAINT fk_order_items_variant FOREIGN KEY (variant_id)
        REFERENCES product_variants (id) ON DELETE RESTRICT,
    CONSTRAINT ck_order_items_unit_price CHECK (unit_price >= 0),
    CONSTRAINT ck_order_items_quantity CHECK (quantity > 0),
    CONSTRAINT ck_order_items_subtotal CHECK (subtotal >= 0)
);

CREATE INDEX idx_order_items_order_id ON order_items (order_id);
CREATE INDEX idx_order_items_product_id ON order_items (product_id);
CREATE INDEX idx_order_items_variant_id ON order_items (variant_id);
