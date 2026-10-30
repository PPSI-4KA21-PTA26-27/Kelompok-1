CREATE TABLE shipments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
    courier_name    TEXT NOT NULL,
    tracking_number TEXT,
    coverage_status TEXT NOT NULL,
    shipping_status TEXT NOT NULL
);
