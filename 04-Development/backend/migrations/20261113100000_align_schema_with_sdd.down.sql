-- Mengembalikan skema sebelum penyelarasan dengan SDD.

ALTER TABLE payments DROP COLUMN proof_url;

DROP INDEX IF EXISTS idx_banners_placement;
ALTER TABLE banners RENAME COLUMN desktop_image_url TO image_url;
ALTER TABLE banners ADD COLUMN placement TEXT NOT NULL DEFAULT 'hero';
UPDATE banners b SET placement = bp.code FROM banner_placements bp WHERE bp.id = b.placement_id;
ALTER TABLE banners
    DROP CONSTRAINT fk_banners_placement,
    DROP COLUMN placement_id,
    ADD CONSTRAINT ck_banners_placement CHECK (placement IN ('hero', 'main', 'catalog'));
CREATE INDEX idx_banners_placement ON banners (placement, is_active, sort_order);
DROP TABLE banner_placements;

DROP INDEX IF EXISTS idx_users_role_id;
ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'customer';
UPDATE users u SET role = r.code FROM roles r WHERE r.id = u.role_id;
ALTER TABLE users
    DROP CONSTRAINT fk_users_role,
    DROP COLUMN role_id,
    ADD CONSTRAINT ck_users_role CHECK (role IN ('customer', 'owner', 'staff'));
DROP TABLE roles;