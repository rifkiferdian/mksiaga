-- Run ONCE after 000003. Back up first; MySQL DDL auto-commits.
-- Existing rows must have no transaction references (verify before migration).
-- Original IDs are retained on the lowest existing store ID. Each other
-- non-deleted store receives an independent copy, including existing edits.
SET @master_store_id = (SELECT MIN(id) FROM stores WHERE deleted_at IS NULL);

ALTER TABLE master_shift
 ADD COLUMN store_id BIGINT UNSIGNED NULL AFTER id,
 DROP INDEX master_shift_code_unique,
 ADD UNIQUE KEY master_shift_store_code_unique (store_id,code),
 ADD CONSTRAINT master_shift_store_fk FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE RESTRICT ON UPDATE RESTRICT;
ALTER TABLE master_checklist
 ADD COLUMN store_id BIGINT UNSIGNED NULL AFTER id,
 DROP INDEX master_checklist_code_unique,
 ADD UNIQUE KEY master_checklist_store_code_unique (store_id,code),
 ADD CONSTRAINT master_checklist_store_fk FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE RESTRICT ON UPDATE RESTRICT;

START TRANSACTION;
UPDATE master_shift SET store_id=@master_store_id;
UPDATE master_checklist SET store_id=@master_store_id;
INSERT INTO master_shift (store_id,code,name,start_time,end_time,sort_order,status,created_at,updated_at)
 SELECT s.id,m.code,m.name,m.start_time,m.end_time,m.sort_order,m.status,m.created_at,m.updated_at
 FROM master_shift m CROSS JOIN stores s
 WHERE m.store_id=@master_store_id AND s.deleted_at IS NULL AND s.id<>@master_store_id;
INSERT INTO master_checklist (store_id,code,phase,name,requires_camera_count,sort_order,status,created_at,updated_at)
 SELECT s.id,m.code,m.phase,m.name,m.requires_camera_count,m.sort_order,m.status,m.created_at,m.updated_at
 FROM master_checklist m CROSS JOIN stores s
 WHERE m.store_id=@master_store_id AND s.deleted_at IS NULL AND s.id<>@master_store_id;
COMMIT;

ALTER TABLE master_shift MODIFY store_id BIGINT UNSIGNED NOT NULL;
ALTER TABLE master_checklist MODIFY store_id BIGINT UNSIGNED NOT NULL;
