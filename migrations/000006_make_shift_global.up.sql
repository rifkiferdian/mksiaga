-- Shift is shared by every store. Run once after 000005.
-- Abort before deleting copies if the same code differs between stores.
CREATE TEMPORARY TABLE shift_consistency_guard (guard_value TINYINT UNSIGNED NOT NULL PRIMARY KEY);
INSERT INTO shift_consistency_guard (guard_value) VALUES (1);
INSERT INTO shift_consistency_guard (guard_value)
SELECT 1 FROM master_shift GROUP BY code
HAVING COUNT(DISTINCT CONCAT_WS('|', name, COALESCE(start_time, ''), COALESCE(end_time, ''), sort_order, status)) > 1;
DROP TEMPORARY TABLE shift_consistency_guard;

DELETE duplicate FROM master_shift AS duplicate
JOIN master_shift AS keeper ON keeper.code=duplicate.code AND keeper.id<duplicate.id;

ALTER TABLE master_shift
    DROP FOREIGN KEY master_shift_store_fk,
    DROP INDEX master_shift_store_code_unique,
    DROP COLUMN store_id,
    ADD UNIQUE KEY master_shift_code_unique (code);
