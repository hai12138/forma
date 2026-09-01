-- Forma S2: Business Model (Source of Truth) — independent of Coze core tables.
-- business_id == asset_id (ADR-013). Semantic revisions are immutable; layout is separate.

CREATE TABLE IF NOT EXISTS `forma_business_model` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `asset_id` VARCHAR(64) NOT NULL,
  `current_revision` INT NOT NULL DEFAULT 1,
  `schema_version` VARCHAR(32) NOT NULL DEFAULT '2.0',
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_business_tenant_business` (`tenant_id`, `business_id`),
  KEY `idx_forma_business_tenant_asset` (`tenant_id`, `asset_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `forma_business_model_revision` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `revision_no` INT NOT NULL,
  `base_revision_no` INT NOT NULL DEFAULT 0,
  `schema_version` VARCHAR(32) NOT NULL DEFAULT '2.0',
  `semantic_model_json` LONGTEXT NOT NULL,
  `content_digest` VARCHAR(128) NOT NULL,
  `change_summary` VARCHAR(512) NOT NULL DEFAULT '',
  `created_by` VARCHAR(64) NOT NULL DEFAULT '',
  `created_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_business_revision` (`tenant_id`, `business_id`, `revision_no`),
  KEY `idx_forma_business_revision_digest` (`tenant_id`, `business_id`, `content_digest`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `forma_business_model_layout` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `layout_revision` INT NOT NULL DEFAULT 1,
  `based_on_model_revision` INT NOT NULL DEFAULT 1,
  `layout_json` LONGTEXT NOT NULL,
  `updated_by` VARCHAR(64) NOT NULL DEFAULT '',
  `updated_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_business_layout` (`tenant_id`, `business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
