-- Forma S0-B initial schema (independent from Coze core tables)
-- Table forma_coze_resource_ref implements the coze_resource_ref mapping contract.

CREATE TABLE IF NOT EXISTS `forma_asset_ref` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` VARCHAR(64) NOT NULL,
  `asset_id` VARCHAR(64) NOT NULL,
  `kind` VARCHAR(32) NOT NULL,
  `name` VARCHAR(255) NOT NULL,
  `semantic_version` VARCHAR(64) NOT NULL DEFAULT '0.0.0',
  `revision` INT NOT NULL DEFAULT 1,
  `schema_version` VARCHAR(32) NOT NULL DEFAULT '1.0',
  `status` VARCHAR(32) NOT NULL DEFAULT 'DRAFT',
  `owner_id` BIGINT NOT NULL DEFAULT 0,
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `content_digest` VARCHAR(128) DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  `deleted_at` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_asset_tenant_asset_revision` (`tenant_id`, `asset_id`, `revision`),
  KEY `idx_forma_asset_tenant_kind` (`tenant_id`, `kind`),
  KEY `idx_forma_asset_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `forma_coze_resource_ref` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` VARCHAR(64) NOT NULL,
  `asset_id` VARCHAR(64) NOT NULL,
  `asset_revision` INT NOT NULL,
  `coze_resource_type` VARCHAR(32) NOT NULL,
  `coze_resource_id` BIGINT NOT NULL,
  `coze_space_id` BIGINT DEFAULT NULL,
  `coze_version` VARCHAR(64) DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_coze_resource_binding` (
    `tenant_id`, `asset_id`, `asset_revision`, `coze_resource_type`, `coze_resource_id`
  ),
  KEY `idx_forma_coze_resource_lookup` (`coze_resource_type`, `coze_resource_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
