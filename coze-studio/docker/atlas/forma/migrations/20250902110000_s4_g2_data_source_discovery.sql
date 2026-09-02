-- Forma S4-G2: tenant-scoped data sources, encrypted credential references,
-- discovery assets, and immutable physical schema snapshots.
-- Independent of Coze core tables; intentionally no FOREIGN KEY constraints.

CREATE TABLE IF NOT EXISTS `forma_data_source` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `source_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `name` VARCHAR(256) NOT NULL,
  `source_type` VARCHAR(32) NOT NULL,
  `status` VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
  `created_by` VARCHAR(64) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_data_source` (`tenant_id`, `source_id`),
  KEY `idx_forma_data_source_tenant_status` (`tenant_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_data_connection` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `connection_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `source_id` VARCHAR(64) NOT NULL,
  `name` VARCHAR(256) NOT NULL,
  `environment` VARCHAR(16) NOT NULL,
  `adapter_type` VARCHAR(32) NOT NULL,
  `public_config_json` TEXT NOT NULL,
  `credential_ref_id` VARCHAR(64) NOT NULL DEFAULT '',
  `created_by` VARCHAR(64) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_data_connection` (`tenant_id`, `connection_id`),
  KEY `idx_forma_data_connection_source` (`tenant_id`, `source_id`, `environment`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_data_credential_ref` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `credential_ref_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `provider` VARCHAR(32) NOT NULL,
  `secret_type` VARCHAR(64) NOT NULL,
  `key_version` INT NOT NULL DEFAULT 1,
  `status` VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
  `created_by` VARCHAR(64) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  `rotated_at` DATETIME(3) NULL,
  `revoked_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_data_credential_ref` (`tenant_id`, `credential_ref_id`),
  KEY `idx_forma_data_credential_status` (`tenant_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_data_secret_local` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `credential_ref_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `ciphertext` LONGBLOB NOT NULL,
  `key_version` INT NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_data_secret_local` (`tenant_id`, `credential_ref_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_data_asset` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `asset_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `source_id` VARCHAR(64) NOT NULL,
  `connection_id` VARCHAR(64) NOT NULL,
  `asset_type` VARCHAR(32) NOT NULL,
  `name` VARCHAR(256) NOT NULL,
  `physical_locator_json` TEXT NOT NULL,
  `locator_digest` VARCHAR(64) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_data_asset` (`tenant_id`, `asset_id`),
  UNIQUE KEY `uk_forma_data_asset_locator` (`tenant_id`, `source_id`, `connection_id`, `locator_digest`),
  KEY `idx_forma_data_asset_source` (`tenant_id`, `source_id`, `asset_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_data_schema_snapshot` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `snapshot_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `source_id` VARCHAR(64) NOT NULL,
  `connection_id` VARCHAR(64) NOT NULL,
  `asset_id` VARCHAR(64) NOT NULL,
  `schema_json` LONGTEXT NOT NULL,
  `fingerprint` VARCHAR(64) NOT NULL,
  `created_by` VARCHAR(64) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_data_schema_snapshot` (`tenant_id`, `snapshot_id`),
  KEY `idx_forma_data_schema_asset` (`tenant_id`, `asset_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
