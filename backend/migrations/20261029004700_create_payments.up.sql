CREATE TABLE payments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    method      TEXT NOT NULL,
    amount      NUMERIC NOT NULL,
    unique_code INT NOT NULL,
    proof_data  BYTEA,
    proof_type  TEXT,
    status      TEXT NOT NULL,
    paid_at     TIMESTAMPTZ,
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMPTZ
);
