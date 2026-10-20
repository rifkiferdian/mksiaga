UPDATE blmkc_access_logs SET action='view_case' WHERE action='view_evidence';

ALTER TABLE blmkc_access_logs
    MODIFY COLUMN action ENUM('view_subject','view_case','download_evidence','print_case','export_data') NOT NULL;
