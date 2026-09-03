-- S5-G0: Platform Admin / User Management Foundation

CREATE TABLE IF NOT EXISTS `forma_platform_role` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `principal_id` VARCHAR(64) NOT NULL,
  `role` VARCHAR(32) NOT NULL DEFAULT 'USER',
  `password_change_required` TINYINT(1) NOT NULL DEFAULT 0,
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_platform_role_principal` (`principal_id`),
  KEY `idx_forma_platform_role_role` (`role`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
