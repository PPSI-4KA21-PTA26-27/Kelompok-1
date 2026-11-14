-- Menghapus penyimpanan bukti pembayaran di basis data (SDD D-07: bukti di
-- Cloudinary, payments hanya menyimpan proof_url). Ditolak bila masih ada
-- bukti lama yang belum dipindahkan, supaya tidak ada berkas yang hilang.
DO $$
DECLARE
    pending INT;
BEGIN
    SELECT count(*) INTO pending FROM payments
     WHERE proof_data IS NOT NULL AND proof_url IS NULL;
    IF pending > 0 THEN
        RAISE EXCEPTION '% bukti pembayaran belum dipindahkan ke Cloudinary. Jalankan: go run ./cmd/proofbackfill', pending;
    END IF;
END $$;

ALTER TABLE payments
    DROP COLUMN proof_data,
    DROP COLUMN proof_type;