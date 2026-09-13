-- Forma S5-G3: Capability validation evidence (PASS/FAIL results).
-- Tenant-scoped; no FOREIGN KEY.

CREATE TABLE IF NOT EXISTS `forma_capability_validation_result` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `validation_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `capability_id` VARCHAR(64) NOT NULL,
  `revision_id` VARCHAR(64) NOT NULL,
  `revision_content_digest` VARCHAR(128) NOT NULL,
  `business_model_revision` INT NOT NULL,
  `business_model_content_digest` VARCHAR(128) NOT NULL,
  `contract_evidence_digest` VARCHAR(128) NOT NULL,
  `evidence_digest` VARCHAR(128) NOT NULL,
  `status` VARCHAR(16) NOT NULL,
  `issue_codes_json` LONGTEXT NOT NULL,
  `validated_by` VARCHAR(64) NOT NULL,
  `validated_at` DATETIME(3) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_capability_validation` (`validation_id`),
  UNIQUE KEY `uk_forma_capability_validation_evidence` (`tenant_id`, `revision_id`, `evidence_digest`),
  KEY `idx_forma_capability_validation_revision` (`tenant_id`, `revision_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
