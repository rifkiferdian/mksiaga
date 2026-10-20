ALTER TABLE blmkc_access_logs
    MODIFY COLUMN action ENUM('view_subject','view_case','view_evidence','download_evidence','print_case','export_data') NOT NULL;
