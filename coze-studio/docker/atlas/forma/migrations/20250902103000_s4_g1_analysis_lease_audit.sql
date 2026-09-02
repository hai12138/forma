-- Forma S4-G1-F1: Analysis execution lease recovery + retry actor audit.
-- Incremental ALTER only; does not modify prior S4-G1 migration.

ALTER TABLE `forma_data_requirement_analysis_run`
  ADD COLUMN `last_retry_by` VARCHAR(64) NOT NULL DEFAULT '' AFTER `retry_count`,
  ADD COLUMN `last_retry_at` DATETIME(3) NULL AFTER `last_retry_by`,
  ADD COLUMN `execution_generation` INT NOT NULL DEFAULT 1 AFTER `last_retry_at`,
  ADD COLUMN `execution_claimed_at` DATETIME(3) NULL AFTER `execution_generation`,
  ADD COLUMN `lease_expires_at` DATETIME(3) NULL AFTER `execution_claimed_at`;

UPDATE `forma_data_requirement_analysis_run`
SET
  `execution_claimed_at` = COALESCE(`started_at`, `created_at`),
  `lease_expires_at` = DATE_ADD(COALESCE(`started_at`, `created_at`), INTERVAL 5 MINUTE),
  `execution_generation` = 1
WHERE `execution_claimed_at` IS NULL;
