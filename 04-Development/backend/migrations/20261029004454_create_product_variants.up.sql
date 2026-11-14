CREATE TABLE product_variants (
    id         UUID NOT NULL DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    sku        TEXT NOT NULL,
    size       TEXT NOT NULL,
    color      TEXT,
    stock      INT NOT NULL DEFAULT 0,

    CONSTRAINT pk_product_variants PRIMARY KEY (id),
    CONSTRAINT uq_product_variants_sku UNIQUE (sku),
    CONSTRAINT fk_product_variants_product FOREIGN KEY (product_id)
        REFERENCES products (id) ON DELETE CASCADE,
    CONSTRAINT ck_product_variants_stock CHECK (stock >= 0)
);

CREATE INDEX idx_product_variants_product_id ON product_variants (product_id);
