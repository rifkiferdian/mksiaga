ALTER TABLE blmkc_evidence_files
    ADD COLUMN verification_status ENUM('pending','verified','rejected') NOT NULL DEFAULT 'pending' AFTER description,
    ADD COLUMN verified_by_user_id BIGINT UNSIGNED NULL AFTER verification_status,
    ADD COLUMN verified_at DATETIME(6) NULL AFTER verified_by_user_id,
    ADD COLUMN verification_notes VARCHAR(500) NULL AFTER verified_at,
    ADD KEY blmkc_evidence_verification_index (verification_status,created_at),
    ADD CONSTRAINT blmkc_evidence_verifier_fk FOREIGN KEY (verified_by_user_id) REFERENCES users(id) ON DELETE SET NULL;
