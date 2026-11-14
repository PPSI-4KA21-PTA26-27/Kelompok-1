-- Menyelaraskan skema dengan SDD revisi 13-11-2026 (Bagian 5, D-07, D-13):
-- 1. Role pindah ke tabel roles; users.role (teks) diganti users.role_id (FK).
-- 2. Penempatan banner pindah ke tabel banner_placements; banners.placement
--    diganti banners.placement_id (FK), dan banners.image_url menjadi
--    desktop_image_url.
-- 3. Bukti pembayaran disimpan di Cloudinary (payments.proof_url). Kolom lama
--    proof_data/proof_type dihapus di migrasi 20261113100100 setelah berkas
--    lama dipindahkan dengan `go run ./cmd/proofbackfill`.

-- 1. roles ---------------------------------------------------------------
CREATE TABLE roles (
    id   UUID NOT NULL DEFAULT gen_random_uuid(),
    code TEXT NOT NULL,
    name TEXT NOT NULL,

    CONSTRAINT pk_roles PRIMARY KEY (id),
    CONSTRAINT uq_roles_code UNIQUE (code),
    CONSTRAINT ck_roles_code CHECK (code IN ('customer', 'staff', 'owner'))
);

INSERT INTO roles (code, name) VALUES
    ('customer', 'Pelanggan'),
    ('staff', 'Staff'),
    ('owner', 'Owner');

ALTER TABLE users ADD COLUMN role_id UUID;
UPDATE users u SET role_id = r.id FROM roles r WHERE r.code = u.role;
ALTER TABLE users
    ALTER COLUMN role_id SET NOT NULL,
    ADD CONSTRAINT fk_users_role FOREIGN KEY (role_id)
        REFERENCES roles (id) ON DELETE RESTRICT,
    DROP CONSTRAINT IF EXISTS ck_users_role,
    DROP COLUMN role;
CREATE INDEX idx_users_role_id ON users (role_id);

-- 2. banner_placements ---------------------------------------------------
CREATE TABLE banner_placements (
    id   UUID NOT NULL DEFAULT gen_random_uuid(),
    code TEXT NOT NULL,
    name TEXT NOT NULL,

    CONSTRAINT pk_banner_placements PRIMARY KEY (id),
    CONSTRAINT uq_banner_placements_code UNIQUE (code)
);

INSERT INTO banner_placements (code, name) VALUES
    ('hero', 'Hero beranda'),
    ('main', 'Banner utama beranda'),
    ('catalog', 'Banner katalog');

ALTER TABLE banners ADD COLUMN placement_id UUID;
UPDATE banners b SET placement_id = bp.id FROM banner_placements bp WHERE bp.code = b.placement;
DROP INDEX IF EXISTS idx_banners_placement;
ALTER TABLE banners
    ALTER COLUMN placement_id SET NOT NULL,
    ADD CONSTRAINT fk_banners_placement FOREIGN KEY (placement_id)
        REFERENCES banner_placements (id) ON DELETE RESTRICT,
    DROP CONSTRAINT IF EXISTS ck_banners_placement,
    DROP COLUMN placement;
ALTER TABLE banners RENAME COLUMN image_url TO desktop_image_url;
CREATE INDEX idx_banners_placement ON banners (placement_id, is_active, sort_order);

-- 3. payments.proof_url ----------------------------------------------------
ALTER TABLE payments ADD COLUMN proof_url TEXT;
