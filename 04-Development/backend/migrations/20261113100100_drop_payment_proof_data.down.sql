-- Kolom dikembalikan kosong; berkas tetap berada di Cloudinary (proof_url).
ALTER TABLE payments
    ADD COLUMN proof_data BYTEA,
    ADD COLUMN proof_type TEXT;