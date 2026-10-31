CREATE TABLE banners (
    id         UUID NOT NULL DEFAULT gen_random_uuid(),
    image_url  TEXT NOT NULL,
    link_url   TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT pk_banners PRIMARY KEY (id)
);
