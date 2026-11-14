-- Ongkir kini dihitung RajaOngkir sampai tingkat kelurahan, jadi order
-- menyimpan wilayah lengkap dan kurir yang dipilih. Kolom nullable supaya
-- order lama tetap valid.
ALTER TABLE orders
    ADD COLUMN shipping_province       TEXT,
    ADD COLUMN shipping_district       TEXT,
    ADD COLUMN shipping_subdistrict_id INT,
    ADD COLUMN shipping_zip_code       TEXT,
    ADD COLUMN courier_code            TEXT,
    ADD COLUMN courier_service         TEXT,
    ADD COLUMN shipping_etd            TEXT,
    ADD COLUMN shipping_weight_gram    INT,
    ADD CONSTRAINT ck_orders_shipping_weight CHECK (shipping_weight_gram IS NULL OR shipping_weight_gram >= 0);

-- Tarif ongkir manual tidak dipakai lagi (diganti RajaOngkir).
DROP TABLE IF EXISTS shipping_rates;
