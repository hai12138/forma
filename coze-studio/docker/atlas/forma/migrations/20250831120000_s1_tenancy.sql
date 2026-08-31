-- Forma S1 tenancy / identity / product foundation
-- Independent from Coze core tables (ID references only, no FK).

CREATE TABLE IF NOT EXISTS `forma_principal` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `principal_id` VARCHAR(64) NOT NULL,
  `principal_type` VARCHAR(32) NOT NULL,
  `provider` VARCHAR(64) NOT NULL DEFAULT 'coze',
  `external_subject` VARCHAR(255) NOT NULL,
  `coze_user_id` BIGINT NOT NULL,
  `display_name` VARCHAR(255) NOT NULL DEFAULT '',
  `status` VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_principal_id` (`principal_id`),
  UNIQUE KEY `uk_forma_principal_provider_subject` (`provider`, `external_subject`),
  UNIQUE KEY `uk_forma_principal_coze_user` (`coze_user_id`),
  KEY `idx_forma_principal_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `forma_tenant` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` VARCHAR(64) NOT NULL,
  `tenant_key` VARCHAR(128) NOT NULL,
  `name` VARCHAR(255) NOT NULL,
  `display_name` VARCHAR(255) NOT NULL,
  `status` VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
  `owner_principal_id` VARCHAR(64) NOT NULL,
  `revision` INT NOT NULL DEFAULT 1,
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  `deleted_at` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_tenant_id` (`tenant_id`),
  UNIQUE KEY `uk_forma_tenant_key` (`tenant_key`),
  KEY `idx_forma_tenant_status` (`status`),
  KEY `idx_forma_tenant_owner` (`owner_principal_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `forma_tenant_membership` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` VARCHAR(64) NOT NULL,
  `principal_id` VARCHAR(64) NOT NULL,
  `role` VARCHAR(32) NOT NULL,
  `status` VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
  `revision` INT NOT NULL DEFAULT 1,
  `joined_at` DATETIME(3) NOT NULL,
  `created_by` VARCHAR(64) NOT NULL DEFAULT '',
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_membership_tenant_principal` (`tenant_id`, `principal_id`),
  KEY `idx_forma_membership_principal` (`principal_id`),
  KEY `idx_forma_membership_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `forma_tenant_space_ref` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` VARCHAR(64) NOT NULL,
  `coze_space_id` BIGINT NOT NULL,
  `purpose` VARCHAR(64) NOT NULL DEFAULT 'DEFAULT',
  `status` VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_space_active` (`coze_space_id`, `status`),
  KEY `idx_forma_tenant_space_tenant` (`tenant_id`),
  KEY `idx_forma_tenant_space_purpose` (`tenant_id`, `purpose`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `forma_audit_event` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` VARCHAR(64) NOT NULL DEFAULT '',
  `principal_id` VARCHAR(64) NOT NULL DEFAULT '',
  `action` VARCHAR(64) NOT NULL,
  `resource` VARCHAR(255) NOT NULL DEFAULT '',
  `request_id` VARCHAR(128) NOT NULL DEFAULT '',
  `created_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_forma_audit_tenant_created` (`tenant_id`, `created_at`),
  KEY `idx_forma_audit_action` (`action`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
