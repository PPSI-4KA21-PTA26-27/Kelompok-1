CREATE TABLE payments (
    id          UUID NOT NULL DEFAULT gen_random_uuid(),
    order_id    UUID NOT NULL,
    method      TEXT NOT NULL,
    amount      NUMERIC NOT NULL,
    unique_code INT NOT NULL,
    proof_data  BYTEA,
    proof_type  TEXT,
    status      TEXT NOT NULL,
    paid_at     TIMESTAMPTZ,
    reviewed_by UUID,
    reviewed_at TIMESTAMPTZ,

    CONSTRAINT pk_payments PRIMARY KEY (id),
    CONSTRAINT fk_payments_order FOREIGN KEY (order_id)
        REFERENCES orders (id) ON DELETE CASCADE,
    CONSTRAINT fk_payments_reviewed_by FOREIGN KEY (reviewed_by)
        REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT ck_payments_amount CHECK (amount >= 0),
    CONSTRAINT ck_payments_unique_code CHECK (unique_code >= 0)
);

CREATE INDEX idx_payments_order_id ON payments (order_id);
CREATE INDEX idx_payments_reviewed_by ON payments (reviewed_by);
