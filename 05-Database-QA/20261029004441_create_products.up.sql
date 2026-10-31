CREATE TABLE products (
    id                UUID NOT NULL DEFAULT gen_random_uuid(),
    category_id       UUID,
    name              TEXT NOT NULL,
    description       TEXT,
    price             NUMERIC NOT NULL,
    discount_percent  NUMERIC NOT NULL DEFAULT 0,
    weight_gram       INT NOT NULL,
    availability_type TEXT NOT NULL,
    po_estimate_days  INT,
    status            TEXT NOT NULL,

    CONSTRAINT pk_products PRIMARY KEY (id),
    CONSTRAINT fk_products_category FOREIGN KEY (category_id)
        REFERENCES categories (id) ON DELETE SET NULL,
    CONSTRAINT ck_products_price CHECK (price >= 0),
    CONSTRAINT ck_products_discount CHECK (discount_percent BETWEEN 0 AND 100),
    CONSTRAINT ck_products_weight CHECK (weight_gram > 0),
    CONSTRAINT ck_products_po_days CHECK (po_estimate_days IS NULL OR po_estimate_days >= 0)
);

CREATE INDEX idx_products_category_id ON products (category_id);
