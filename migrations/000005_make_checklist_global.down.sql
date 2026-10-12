-- Restores per-store copies from the current global checklist.
-- This cannot restore historical per-store differences removed by the up migration.
SET @checklist_store_id = (SELECT MIN(id) FROM stores WHERE deleted_at IS NULL);

ALTER TABLE master_checklist
    ADD COLUMN store_id BIGINT UNSIGNED NULL AFTER id,
    DROP INDEX master_checklist_code_unique,
    ADD UNIQUE KEY master_checklist_store_code_unique (store_id, code),
    ADD CONSTRAINT master_checklist_store_fk
        FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE RESTRICT ON UPDATE RESTRICT;

START TRANSACTION;
UPDATE master_checklist SET store_id=@checklist_store_id;
INSERT INTO master_checklist
    (store_id, code, phase, name, requires_camera_count, sort_order, status, created_at, updated_at)
SELECT s.id, c.code, c.phase, c.name, c.requires_camera_count,
       c.sort_order, c.status, c.created_at, c.updated_at
FROM master_checklist c
CROSS JOIN stores s
WHERE c.store_id=@checklist_store_id
  AND s.deleted_at IS NULL
  AND s.id<>@checklist_store_id;
COMMIT;

ALTER TABLE master_checklist MODIFY store_id BIGINT UNSIGNED NOT NULL;
