CREATE TABLE cart_items (
    id         UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL,
    variant_id UUID NOT NULL,
    quantity   INT NOT NULL,

    CONSTRAINT pk_cart_items PRIMARY KEY (id),
    CONSTRAINT fk_cart_items_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_cart_items_variant FOREIGN KEY (variant_id)
        REFERENCES product_variants (id) ON DELETE CASCADE,
    CONSTRAINT ck_cart_items_quantity CHECK (quantity > 0)
);

CREATE INDEX idx_cart_items_user_id ON cart_items (user_id);
CREATE INDEX idx_cart_items_variant_id ON cart_items (variant_id);
