-- Skema Sprint 2 (SDD Bagian 10.2): tabel banner_placements dan banners
-- sesuai kamus data SDD Bagian 5.3 (UC-18, KF-14).

CREATE TABLE banner_placements (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code text NOT NULL,
    name text NOT NULL,
    CONSTRAINT pk_banner_placements PRIMARY KEY (id),
    CONSTRAINT uq_banner_placements_code UNIQUE (code)
);

CREATE TABLE banners (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    placement_id uuid NOT NULL,
    title text,
    subtitle text,
    cta_label text,
    desktop_image_url text NOT NULL,
    mobile_image_url text,
    link_url text,
    sort_order integer DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    starts_at timestamp with time zone,
    ends_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT pk_banners PRIMARY KEY (id),
    CONSTRAINT fk_banners_placement FOREIGN KEY (placement_id)
        REFERENCES banner_placements (id) ON DELETE RESTRICT,
    CONSTRAINT ck_banners_schedule CHECK (starts_at IS NULL OR ends_at IS NULL OR ends_at > starts_at)
);

CREATE INDEX idx_banners_placement ON banners USING btree (placement_id, is_active, sort_order);

INSERT INTO banner_placements (code, name) VALUES
    ('hero', 'Hero beranda'),
    ('main', 'Banner utama beranda'),
    ('catalog', 'Banner katalog');
