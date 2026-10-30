CREATE TABLE product_variants (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku        TEXT NOT NULL UNIQUE,
    size       TEXT NOT NULL,
    color      TEXT,
    stock      INT NOT NULL DEFAULT 0
);
