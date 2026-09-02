-- Forma S4-G4: data contract, immutable revisions, validation/lifecycle/drift/gap.
-- Intentionally no FOREIGN KEY constraints.

CREATE TABLE IF NOT EXISTS `forma_data_contract` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `contract_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `active_revision_id` VARCHAR(64) NOT NULL DEFAULT '',
  `created_by` VARCHAR(64) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_data_contract` (`tenant_id`, `contract_id`),
  KEY `idx_forma_data_contract_scope` (`tenant_id`, `business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_data_contract_revision` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `revision_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `contract_id` VARCHAR(64) NOT NULL,
  `version` INT NOT NULL,
  `status` VARCHAR(16) NOT NULL DEFAULT 'DRAFT',
  `business_model_revision` INT NOT NULL,
  `name` VARCHAR(256) NOT NULL,
  `description` TEXT NOT NULL,
  `requirement_ids_json` LONGTEXT NOT NULL,
  `logical_schema_json` LONGTEXT NOT NULL,
  `query_capabilities_json` LONGTEXT NOT NULL,
  `filter_schema_json` LONGTEXT NOT NULL,
  `sort_schema_json` LONGTEXT NOT NULL,
  `pagination_policy_json` LONGTEXT NOT NULL,
  `freshness_policy` VARCHAR(32) NOT NULL,
  `classification_policy_json` LONGTEXT NOT NULL,
  `binding_refs_json` LONGTEXT NOT NULL,
  `access_policy_ref` VARCHAR(64) NOT NULL DEFAULT '',
  `derived_from_revision_id` VARCHAR(64) NOT NULL DEFAULT '',
  `created_by` VARCHAR(64) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_data_contract_revision` (`tenant_id`, `revision_id`),
  UNIQUE KEY `uk_forma_data_contract_revision_version` (`tenant_id`, `contract_id`, `version`),
  KEY `idx_forma_data_contract_revision_scope` (`tenant_id`, `contract_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_data_contract_validation_result` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `validation_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `contract_id` VARCHAR(64) NOT NULL,
  `revision_id` VARCHAR(64) NOT NULL,
  `version` INT NOT NULL,
  `status` VARCHAR(16) NOT NULL,
  `errors_json` LONGTEXT NOT NULL,
  `warnings_json` LONGTEXT NOT NULL,
  `snapshot_fingerprints_json` LONGTEXT NOT NULL,
  `validated_by` VARCHAR(64) NOT NULL,
  `validated_at` DATETIME(3) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_data_contract_validation` (`tenant_id`, `validation_id`),
  KEY `idx_forma_data_contract_validation_rev` (`tenant_id`, `revision_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_data_contract_lifecycle_event` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `event_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `contract_id` VARCHAR(64) NOT NULL,
  `revision_id` VARCHAR(64) NOT NULL,
  `version` INT NOT NULL,
  `action` VARCHAR(32) NOT NULL,
  `actor_principal_id` VARCHAR(64) NOT NULL,
  `reason` TEXT NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_data_contract_lifecycle_event` (`tenant_id`, `event_id`),
  KEY `idx_forma_data_contract_lifecycle_scope` (`tenant_id`, `contract_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_data_contract_drift_result` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `drift_result_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `contract_id` VARCHAR(64) NOT NULL,
  `revision_id` VARCHAR(64) NOT NULL,
  `version` INT NOT NULL,
  `severity` VARCHAR(16) NOT NULL,
  `findings_json` LONGTEXT NOT NULL,
  `compared_snapshot_ids_json` LONGTEXT NOT NULL,
  `evaluated_by` VARCHAR(64) NOT NULL,
  `evaluated_at` DATETIME(3) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_data_contract_drift` (`tenant_id`, `drift_result_id`),
  KEY `idx_forma_data_contract_drift_rev` (`tenant_id`, `revision_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_data_contract_gap_result` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `gap_result_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `contract_id` VARCHAR(64) NOT NULL,
  `revision_id` VARCHAR(64) NOT NULL,
  `version` INT NOT NULL,
  `from_business_revision` INT NOT NULL,
  `current_business_revision` INT NOT NULL,
  `new_confirmed_requirement_ids_json` LONGTEXT NOT NULL,
  `unmapped_requirement_ids_json` LONGTEXT NOT NULL,
  `gap_status` VARCHAR(16) NOT NULL,
  `evaluated_by` VARCHAR(64) NOT NULL,
  `evaluated_at` DATETIME(3) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_data_contract_gap` (`tenant_id`, `gap_result_id`),
  KEY `idx_forma_data_contract_gap_rev` (`tenant_id`, `revision_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
