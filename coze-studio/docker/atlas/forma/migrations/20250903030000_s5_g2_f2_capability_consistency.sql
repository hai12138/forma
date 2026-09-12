-- Forma S5-G2-F2: Capability analysis attempt audit for claim/retry/lease takeover.
CREATE TABLE IF NOT EXISTS `forma_capability_analysis_attempt` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `attempt_id` VARCHAR(64) NOT NULL,
  `analysis_run_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `attempt` INT NOT NULL,
  `actor_principal_id` VARCHAR(64) NOT NULL,
  `trigger_kind` VARCHAR(32) NOT NULL,
  `result_status` VARCHAR(32) NOT NULL,
  `error_code` VARCHAR(128) NOT NULL DEFAULT '',
  `created_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_capability_analysis_attempt` (`attempt_id`),
  KEY `idx_forma_capability_analysis_attempt_run` (`tenant_id`, `analysis_run_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
