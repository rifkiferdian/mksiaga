-- Restores one shift copy for every active store. Historical store-specific
-- differences cannot be reconstructed after the global migration.
ALTER TABLE master_shift DROP INDEX master_shift_code_unique,
    ADD COLUMN store_id BIGINT UNSIGNED NULL AFTER id;

CREATE TEMPORARY TABLE global_shift_copy AS
SELECT code,name,start_time,end_time,sort_order,status FROM master_shift;
DELETE FROM master_shift;
INSERT INTO master_shift (store_id,code,name,start_time,end_time,sort_order,status)
SELECT s.id,g.code,g.name,g.start_time,g.end_time,g.sort_order,g.status
FROM stores s CROSS JOIN global_shift_copy g WHERE s.deleted_at IS NULL;
DROP TEMPORARY TABLE global_shift_copy;

ALTER TABLE master_shift MODIFY store_id BIGINT UNSIGNED NOT NULL,
    ADD UNIQUE KEY master_shift_store_code_unique (store_id,code),
    ADD CONSTRAINT master_shift_store_fk FOREIGN KEY (store_id) REFERENCES stores(id);
