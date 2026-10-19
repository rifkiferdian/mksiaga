DROP TABLE IF EXISTS daily_cctv_report_events;

UPDATE daily_cctv_reports SET status='submitted' WHERE status='approved';
UPDATE daily_cctv_reports SET status='draft' WHERE status NOT IN ('draft','submitted');

ALTER TABLE daily_cctv_reports
    DROP FOREIGN KEY daily_report_reviewer_fk,
    DROP COLUMN review_notes,
    DROP COLUMN reviewed_at,
    DROP COLUMN reviewed_by_user_id,
    DROP COLUMN accepted_at,
    MODIFY COLUMN status ENUM('draft','submitted') NOT NULL DEFAULT 'draft';

ALTER TABLE user_stores
    DROP FOREIGN KEY user_stores_default_shift_fk,
    DROP INDEX user_stores_default_shift_index,
    DROP COLUMN default_shift_id;
