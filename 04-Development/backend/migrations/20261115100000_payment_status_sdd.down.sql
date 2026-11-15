ALTER TABLE payments DROP CONSTRAINT ck_payments_status;
UPDATE payments SET status = CASE status
    WHEN 'menunggu' THEN 'pending' WHEN 'valid' THEN 'approved' WHEN 'ditolak' THEN 'rejected' ELSE status END;
