-- Rollback ini menghapus seluruh data BLMKC. Gunakan hanya setelah backup.
DROP TABLE IF EXISTS blmkc_access_logs;
DROP TABLE IF EXISTS blmkc_case_events;
DROP TABLE IF EXISTS blmkc_case_resolutions;
DROP TABLE IF EXISTS blmkc_evidence_files;
DROP TABLE IF EXISTS blmkc_vehicles;
DROP TABLE IF EXISTS blmkc_case_items;
DROP TABLE IF EXISTS blmkc_case_subjects;
DROP TABLE IF EXISTS blmkc_cases;
DROP TABLE IF EXISTS blmkc_subject_aliases;
DROP TABLE IF EXISTS blmkc_subjects;
