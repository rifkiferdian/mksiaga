UPDATE blmkc_evidence_files SET file_category='cctv_capture' WHERE file_category='cctv_video';

ALTER TABLE blmkc_evidence_files
    DROP FOREIGN KEY blmkc_evidence_verifier_fk,
    DROP KEY blmkc_evidence_verification_index,
    DROP COLUMN verification_notes,
    DROP COLUMN verified_at,
    DROP COLUMN verified_by_user_id,
    DROP COLUMN verification_status,
    MODIFY COLUMN file_category ENUM(
        'subject_photo','identity_card','incident_photo','cctv_capture','vehicle_photo',
        'stolen_item_photo','statement_letter','settlement_letter','police_report',
        'case_resolution','other'
    ) NOT NULL;
