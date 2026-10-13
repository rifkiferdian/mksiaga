CREATE TABLE daily_cctv_reports (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    store_id BIGINT UNSIGNED NOT NULL,
    officer_id BIGINT UNSIGNED NOT NULL,
    supervisor_id BIGINT UNSIGNED NULL,
    shift_id BIGINT UNSIGNED NOT NULL,
    report_date DATE NOT NULL,
    work_start TIME NOT NULL,
    work_end TIME NOT NULL,
    monitoring_location VARCHAR(255) NOT NULL,
    start_notes TEXT NULL,
    end_notes TEXT NULL,
    handover_to_user_id BIGINT UNSIGNED NULL,
    handover_at TIME NULL,
    handover_notes TEXT NULL,
    status ENUM('draft','submitted') NOT NULL DEFAULT 'draft',
    submitted_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY daily_report_identity_unique (store_id,officer_id,report_date,shift_id),
    KEY daily_report_store_date_index (store_id,report_date),
    CONSTRAINT daily_report_store_fk FOREIGN KEY (store_id) REFERENCES stores(id),
    CONSTRAINT daily_report_officer_fk FOREIGN KEY (officer_id) REFERENCES users(id),
    CONSTRAINT daily_report_supervisor_fk FOREIGN KEY (supervisor_id) REFERENCES users(id),
    CONSTRAINT daily_report_shift_fk FOREIGN KEY (shift_id) REFERENCES master_shift(id),
    CONSTRAINT daily_report_handover_user_fk FOREIGN KEY (handover_to_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE daily_cctv_report_checklists (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    report_id BIGINT UNSIGNED NOT NULL,
    checklist_id BIGINT UNSIGNED NOT NULL,
    answer ENUM('yes','no') NOT NULL,
    camera_active SMALLINT UNSIGNED NULL,
    camera_total SMALLINT UNSIGNED NULL,
    notes VARCHAR(500) NULL,
    PRIMARY KEY (id),
    UNIQUE KEY daily_report_checklist_unique (report_id,checklist_id),
    CONSTRAINT daily_report_checklist_report_fk FOREIGN KEY (report_id) REFERENCES daily_cctv_reports(id) ON DELETE CASCADE,
    CONSTRAINT daily_report_checklist_master_fk FOREIGN KEY (checklist_id) REFERENCES master_checklist(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE daily_cctv_monitoring_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    report_id BIGINT UNSIGNED NOT NULL,
    log_time TIME NOT NULL,
    area_id BIGINT UNSIGNED NULL,
    condition_status ENUM('normal','finding') NOT NULL,
    finding TEXT NULL,
    action_taken TEXT NULL,
    PRIMARY KEY (id),
    KEY daily_monitoring_report_time_index (report_id,log_time),
    CONSTRAINT daily_monitoring_report_fk FOREIGN KEY (report_id) REFERENCES daily_cctv_reports(id) ON DELETE CASCADE,
    CONSTRAINT daily_monitoring_area_fk FOREIGN KEY (area_id) REFERENCES master_area(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE daily_cctv_incidents (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    report_id BIGINT UNSIGNED NOT NULL,
    incident_time TIME NULL,
    area_id BIGINT UNSIGNED NULL,
    camera_id BIGINT UNSIGNED NULL,
    incident_type_id BIGINT UNSIGNED NULL,
    other_type VARCHAR(150) NULL,
    description TEXT NULL,
    perpetrator_gender VARCHAR(30) NULL,
    perpetrator_age VARCHAR(30) NULL,
    perpetrator_clothing VARCHAR(255) NULL,
    perpetrator_features VARCHAR(255) NULL,
    response_minutes SMALLINT UNSIGNED NULL,
    incident_status ENUM('completed','in_progress','forwarded') NULL,
    PRIMARY KEY (id),
    KEY daily_incident_report_index (report_id),
    CONSTRAINT daily_incident_report_fk FOREIGN KEY (report_id) REFERENCES daily_cctv_reports(id) ON DELETE CASCADE,
    CONSTRAINT daily_incident_area_fk FOREIGN KEY (area_id) REFERENCES master_area(id),
    CONSTRAINT daily_incident_camera_fk FOREIGN KEY (camera_id) REFERENCES master_kamera(id),
    CONSTRAINT daily_incident_type_fk FOREIGN KEY (incident_type_id) REFERENCES master_jenis_insiden(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE daily_cctv_incident_actions (
    incident_id BIGINT UNSIGNED NOT NULL,
    action_id BIGINT UNSIGNED NOT NULL,
    notes VARCHAR(255) NULL,
    PRIMARY KEY (incident_id,action_id),
    CONSTRAINT daily_incident_action_incident_fk FOREIGN KEY (incident_id) REFERENCES daily_cctv_incidents(id) ON DELETE CASCADE,
    CONSTRAINT daily_incident_action_master_fk FOREIGN KEY (action_id) REFERENCES master_tindakan(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
