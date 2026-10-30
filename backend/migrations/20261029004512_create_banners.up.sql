CREATE TABLE banners (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    image_url  TEXT NOT NULL,
    link_url   TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE
);
