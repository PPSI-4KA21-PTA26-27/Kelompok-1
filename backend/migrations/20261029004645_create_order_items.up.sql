CREATE TABLE order_items (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id     UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id   UUID NOT NULL REFERENCES products(id),
    variant_id   UUID NOT NULL REFERENCES product_variants(id),
    product_name TEXT NOT NULL,
    size         TEXT NOT NULL,
    unit_price   NUMERIC NOT NULL,
    quantity     INT NOT NULL,
    subtotal     NUMERIC NOT NULL
);
