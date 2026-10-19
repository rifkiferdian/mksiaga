ALTER TABLE user_stores
    ADD COLUMN default_shift_id BIGINT UNSIGNED NULL AFTER store_id,
    ADD KEY user_stores_default_shift_index (default_shift_id),
    ADD CONSTRAINT user_stores_default_shift_fk FOREIGN KEY (default_shift_id) REFERENCES master_shift(id);

ALTER TABLE daily_cctv_reports
    MODIFY COLUMN status ENUM('draft','submitted','waiting_handover','waiting_review','revision','approved') NOT NULL DEFAULT 'draft',
    ADD COLUMN accepted_at DATETIME(6) NULL AFTER submitted_at,
    ADD COLUMN reviewed_by_user_id BIGINT UNSIGNED NULL AFTER accepted_at,
    ADD COLUMN reviewed_at DATETIME(6) NULL AFTER reviewed_by_user_id,
    ADD COLUMN review_notes TEXT NULL AFTER reviewed_at,
    ADD CONSTRAINT daily_report_reviewer_fk FOREIGN KEY (reviewed_by_user_id) REFERENCES users(id);

UPDATE daily_cctv_reports
SET status='approved', reviewed_at=COALESCE(submitted_at,updated_at)
WHERE status='submitted';

ALTER TABLE daily_cctv_reports
    MODIFY COLUMN status ENUM('draft','waiting_handover','waiting_review','revision','approved') NOT NULL DEFAULT 'draft';

CREATE TABLE daily_cctv_report_events (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    report_id BIGINT UNSIGNED NOT NULL,
    actor_user_id BIGINT UNSIGNED NOT NULL,
    from_status VARCHAR(30) NOT NULL,
    to_status VARCHAR(30) NOT NULL,
    notes TEXT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY daily_report_event_report_index (report_id,created_at),
    CONSTRAINT daily_report_event_report_fk FOREIGN KEY (report_id) REFERENCES daily_cctv_reports(id) ON DELETE CASCADE,
    CONSTRAINT daily_report_event_actor_fk FOREIGN KEY (actor_user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
