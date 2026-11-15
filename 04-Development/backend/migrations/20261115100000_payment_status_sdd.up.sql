-- Status verifikasi pembayaran mengikuti kamus data SDD: menunggu, valid, ditolak.
UPDATE payments SET status = CASE status
    WHEN 'pending' THEN 'menunggu' WHEN 'approved' THEN 'valid' WHEN 'rejected' THEN 'ditolak' ELSE status END;
ALTER TABLE payments ADD CONSTRAINT ck_payments_status CHECK (status IN ('menunggu', 'valid', 'ditolak'));
