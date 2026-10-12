-- Checklist is shared by every store. Run once after 000004.
-- Abort before deleting copies if any store has changed the same code differently.
-- The duplicate-key guard is deliberately compatible with migration runners that
-- send the complete file as one multi-statement query (no DELIMITER handling).
CREATE TEMPORARY TABLE checklist_consistency_guard (
    guard_value TINYINT UNSIGNED NOT NULL PRIMARY KEY
);
INSERT INTO checklist_consistency_guard (guard_value) VALUES (1);
INSERT INTO checklist_consistency_guard (guard_value)
SELECT 1
FROM master_checklist
GROUP BY code
HAVING COUNT(DISTINCT CONCAT_WS('|', phase, name, requires_camera_count, sort_order, status)) > 1;
DROP TEMPORARY TABLE checklist_consistency_guard;

DELETE duplicate
FROM master_checklist AS duplicate
JOIN master_checklist AS keeper
  ON keeper.code = duplicate.code
 AND keeper.id < duplicate.id;

ALTER TABLE master_checklist
    DROP FOREIGN KEY master_checklist_store_fk,
    DROP INDEX master_checklist_store_code_unique,
    DROP COLUMN store_id,
    ADD UNIQUE KEY master_checklist_code_unique (code);
