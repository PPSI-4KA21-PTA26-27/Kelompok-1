CREATE TABLE products (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id       UUID REFERENCES categories(id),
    name              TEXT NOT NULL,
    description       TEXT,
    price             NUMERIC NOT NULL,
    discount_percent  NUMERIC NOT NULL DEFAULT 0,
    weight_gram       INT NOT NULL,
    availability_type TEXT NOT NULL,
    po_estimate_days  INT,
    status            TEXT NOT NULL
);
