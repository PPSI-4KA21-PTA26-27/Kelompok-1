-- Menghapus kolom yang tidak ada di kamus data SDD revisi 13-11-2026:
-- orders.shipping_province (alamat tujuan di SDD: alamat, kelurahan,
-- kecamatan, kota, kode pos) dan shipments.courier_name (kurir pesanan
-- diambil dari orders.courier_code dan courier_service).
ALTER TABLE orders DROP COLUMN shipping_province;
ALTER TABLE shipments DROP COLUMN courier_name;