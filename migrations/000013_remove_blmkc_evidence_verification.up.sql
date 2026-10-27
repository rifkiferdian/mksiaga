-- Verifikasi dilakukan pada tingkat kasus. Bukti menjadi lampiran pemeriksaan
-- dan tidak lagi memiliki keputusan verifikasi tersendiri.
ALTER TABLE blmkc_evidence_files
    DROP FOREIGN KEY blmkc_evidence_verifier_fk,
    DROP KEY blmkc_evidence_verification_index,
    DROP COLUMN verification_notes,
    DROP COLUMN verified_at,
    DROP COLUMN verified_by_user_id,
    DROP COLUMN verification_status;
