ALTER TABLE daily_cctv_reports
    ADD COLUMN handover_from_signature MEDIUMTEXT NULL AFTER handover_notes,
    ADD COLUMN handover_to_signature MEDIUMTEXT NULL AFTER handover_from_signature;
