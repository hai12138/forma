-- Forma S4-G2-F1: canonical source types and persisted connection health.

UPDATE `forma_data_source`
SET `source_type` = 'RELATIONAL_DATABASE'
WHERE `source_type` = 'DATABASE';

UPDATE `forma_data_source`
SET `source_type` = 'HTTP_API'
WHERE `source_type` = 'HTTP';

ALTER TABLE `forma_data_connection`
  ADD COLUMN `status` VARCHAR(16) NOT NULL DEFAULT 'INACTIVE' AFTER `credential_ref_id`,
  ADD COLUMN `last_test_status` VARCHAR(16) NOT NULL DEFAULT '' AFTER `status`,
  ADD COLUMN `last_test_at` DATETIME(3) NULL AFTER `last_test_status`,
  ADD COLUMN `last_test_error_key` VARCHAR(128) NOT NULL DEFAULT '' AFTER `last_test_at`;
