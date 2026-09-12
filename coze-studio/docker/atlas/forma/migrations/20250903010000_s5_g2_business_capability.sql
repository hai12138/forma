-- Forma S5-G2: Business Capability domain
-- Independent of Coze core tables. Tenant-scoped; no FOREIGN KEY to Coze.

CREATE TABLE IF NOT EXISTS `forma_business_capability` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `capability_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `active_revision_id` VARCHAR(64) NULL,
  `created_by` VARCHAR(64) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_business_capability` (`tenant_id`, `capability_id`),
  KEY `idx_forma_capability_business` (`tenant_id`, `business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_business_capability_revision` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `revision_id` VARCHAR(64) NOT NULL,
  `capability_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `version` INT NOT NULL,
  `status` VARCHAR(16) NOT NULL,
  `name` VARCHAR(256) NOT NULL,
  `description` TEXT NOT NULL,
  `business_model_revision` INT NOT NULL,
  `capability_kind` VARCHAR(16) NOT NULL,
  `input_schema_json` LONGTEXT NOT NULL,
  `output_schema_json` LONGTEXT NOT NULL,
  `preconditions_json` LONGTEXT NOT NULL,
  `effects_json` LONGTEXT NOT NULL,
  `data_contract_bindings_json` LONGTEXT NOT NULL,
  `query_operation` VARCHAR(32) NOT NULL DEFAULT '',
  `output_cardinality` VARCHAR(16) NOT NULL DEFAULT '',
  `derived_from_revision_id` VARCHAR(64) NOT NULL DEFAULT '',
  `analysis_run_id` VARCHAR(64) NOT NULL DEFAULT '',
  `proposal_id` VARCHAR(64) NOT NULL DEFAULT '',
  `source` VARCHAR(32) NOT NULL,
  `created_by` VARCHAR(64) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_capability_revision` (`tenant_id`, `revision_id`),
  UNIQUE KEY `uk_forma_capability_revision_version` (`tenant_id`, `capability_id`, `version`),
  KEY `idx_forma_capability_revision_status` (`tenant_id`, `capability_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_capability_analysis_run` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `analysis_run_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `business_model_revision` INT NOT NULL,
  `client_request_id` VARCHAR(64) NOT NULL,
  `request_digest` VARCHAR(128) NOT NULL,
  `status` VARCHAR(32) NOT NULL,
  `attempt` INT NOT NULL DEFAULT 1,
  `error_code` VARCHAR(128) NOT NULL DEFAULT '',
  `model_ref` VARCHAR(128) NOT NULL DEFAULT '',
  `request_json` LONGTEXT NOT NULL,
  `execution_claimed_at` DATETIME(3) NULL,
  `lease_expires_at` DATETIME(3) NULL,
  `created_by` VARCHAR(64) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_capability_analysis_run` (`analysis_run_id`),
  UNIQUE KEY `uk_forma_capability_analysis_idempotency` (`tenant_id`, `business_id`, `business_model_revision`, `client_request_id`),
  KEY `idx_forma_capability_analysis_status` (`tenant_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_capability_proposal` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `proposal_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `analysis_run_id` VARCHAR(64) NOT NULL,
  `capability_id` VARCHAR(64) NULL,
  `status` VARCHAR(32) NOT NULL,
  `payload_json` LONGTEXT NOT NULL,
  `materialized_revision_id` VARCHAR(64) NULL,
  `created_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_capability_proposal` (`tenant_id`, `proposal_id`),
  KEY `idx_forma_capability_proposal_run` (`tenant_id`, `analysis_run_id`),
  KEY `idx_forma_capability_proposal_status` (`tenant_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_capability_decision` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `decision_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `capability_id` VARCHAR(64) NULL,
  `proposal_id` VARCHAR(64) NULL,
  `source_revision_id` VARCHAR(64) NULL,
  `target_revision_id` VARCHAR(64) NULL,
  `action` VARCHAR(32) NOT NULL,
  `payload_digest` VARCHAR(128) NOT NULL DEFAULT '',
  `client_request_id` VARCHAR(64) NULL,
  `actor_principal_id` VARCHAR(64) NOT NULL,
  `reason` VARCHAR(1024) NOT NULL DEFAULT '',
  `created_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_capability_decision` (`tenant_id`, `decision_id`),
  UNIQUE KEY `uk_forma_capability_decision_proposal` (`tenant_id`, `proposal_id`),
  UNIQUE KEY `uk_forma_capability_decision_derive` (`tenant_id`, `capability_id`, `source_revision_id`, `client_request_id`),
  KEY `idx_forma_capability_decision_cap` (`tenant_id`, `capability_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
