CREATE TABLE categories (
    id        UUID NOT NULL DEFAULT gen_random_uuid(),
    name      TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT pk_categories PRIMARY KEY (id),
    CONSTRAINT uq_categories_name UNIQUE (name)
);
