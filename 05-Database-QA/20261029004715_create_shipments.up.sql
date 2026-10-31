CREATE TABLE shipments (
    id              UUID NOT NULL DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL,
    courier_name    TEXT NOT NULL,
    tracking_number TEXT,
    coverage_status TEXT NOT NULL,
    shipping_status TEXT NOT NULL,

    CONSTRAINT pk_shipments PRIMARY KEY (id),
    CONSTRAINT uq_shipments_order UNIQUE (order_id),
    CONSTRAINT fk_shipments_order FOREIGN KEY (order_id)
        REFERENCES orders (id) ON DELETE CASCADE
);
