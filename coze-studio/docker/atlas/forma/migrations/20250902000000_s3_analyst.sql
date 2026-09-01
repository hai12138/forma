-- Forma S3: AI Business Analyst — sessions, evidence, assertions, proposals.
-- Independent of Coze core tables. Tenant-scoped; no FOREIGN KEY to Coze.

CREATE TABLE IF NOT EXISTS `forma_analyst_session` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `session_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `status` VARCHAR(32) NOT NULL,
  `title` VARCHAR(256) NOT NULL DEFAULT '',
  `runtime_conversation_ref` VARCHAR(128) NOT NULL DEFAULT '',
  `confirmation_policy` VARCHAR(32) NOT NULL DEFAULT 'DEVELOPMENT',
  `created_by` VARCHAR(64) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  `closed_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_analyst_session` (`session_id`),
  KEY `idx_forma_analyst_session_tenant_biz` (`tenant_id`, `business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_analyst_turn` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `turn_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `session_id` VARCHAR(64) NOT NULL,
  `sequence` INT NOT NULL,
  `speaker` VARCHAR(16) NOT NULL,
  `content` TEXT NOT NULL,
  `content_type` VARCHAR(32) NOT NULL DEFAULT 'TEXT',
  `client_request_id` VARCHAR(64) NOT NULL DEFAULT '',
  `model_request_id` VARCHAR(64) NOT NULL DEFAULT '',
  `analysis_status` VARCHAR(32) NOT NULL DEFAULT 'NONE',
  `created_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_analyst_turn` (`turn_id`),
  UNIQUE KEY `uk_forma_analyst_turn_idempotency` (`session_id`, `client_request_id`),
  KEY `idx_forma_analyst_turn_session` (`tenant_id`, `session_id`, `sequence`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_business_evidence` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `evidence_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `session_id` VARCHAR(64) NOT NULL,
  `turn_id` VARCHAR(64) NOT NULL DEFAULT '',
  `source_type` VARCHAR(32) NOT NULL,
  `source_ref` VARCHAR(128) NOT NULL DEFAULT '',
  `quote` TEXT NOT NULL,
  `content_digest` VARCHAR(64) NOT NULL,
  `created_by` VARCHAR(64) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_business_evidence` (`evidence_id`),
  KEY `idx_forma_evidence_business` (`tenant_id`, `business_id`),
  KEY `idx_forma_evidence_session` (`tenant_id`, `session_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_business_assertion` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `assertion_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `session_id` VARCHAR(64) NOT NULL,
  `assertion_type` VARCHAR(64) NOT NULL,
  `subject_ref` VARCHAR(256) NOT NULL DEFAULT '',
  `predicate` VARCHAR(256) NOT NULL DEFAULT '',
  `object_value` VARCHAR(512) NOT NULL DEFAULT '',
  `structured_value_json` TEXT NOT NULL,
  `confidence` DOUBLE NOT NULL DEFAULT 0,
  `status` VARCHAR(32) NOT NULL,
  `source_marker` VARCHAR(32) NOT NULL,
  `derived_from_assertion_id` VARCHAR(64) NOT NULL DEFAULT '',
  `created_by` VARCHAR(64) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_business_assertion` (`assertion_id`),
  KEY `idx_forma_assertion_business` (`tenant_id`, `business_id`),
  KEY `idx_forma_assertion_session` (`tenant_id`, `session_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_assertion_evidence_ref` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` VARCHAR(64) NOT NULL,
  `assertion_id` VARCHAR(64) NOT NULL,
  `evidence_id` VARCHAR(64) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_assertion_evidence` (`tenant_id`, `assertion_id`, `evidence_id`),
  KEY `idx_forma_assertion_evidence_assertion` (`tenant_id`, `assertion_id`),
  KEY `idx_forma_assertion_evidence_evidence` (`tenant_id`, `evidence_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_business_confirmation` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `confirmation_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `assertion_id` VARCHAR(64) NOT NULL,
  `decision` VARCHAR(16) NOT NULL,
  `comment` TEXT NOT NULL,
  `decided_by` VARCHAR(64) NOT NULL,
  `decided_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_business_confirmation` (`confirmation_id`),
  KEY `idx_forma_confirmation_assertion` (`tenant_id`, `assertion_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_assertion_conflict` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `conflict_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `session_id` VARCHAR(64) NOT NULL,
  `assertion_id_a` VARCHAR(64) NOT NULL,
  `assertion_id_b` VARCHAR(64) NOT NULL,
  `subject_ref` VARCHAR(256) NOT NULL DEFAULT '',
  `predicate` VARCHAR(256) NOT NULL DEFAULT '',
  `status` VARCHAR(32) NOT NULL DEFAULT 'OPEN',
  `created_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_assertion_conflict` (`conflict_id`),
  KEY `idx_forma_conflict_business` (`tenant_id`, `business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_analyst_gap` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `gap_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `session_id` VARCHAR(64) NOT NULL,
  `gap_type` VARCHAR(64) NOT NULL DEFAULT 'INFORMATION',
  `question` TEXT NOT NULL,
  `related_assertion_ids_json` TEXT NOT NULL,
  `status` VARCHAR(32) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_analyst_gap` (`gap_id`),
  KEY `idx_forma_gap_business` (`tenant_id`, `business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_business_model_proposal` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `proposal_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `session_id` VARCHAR(64) NOT NULL,
  `base_revision` INT NOT NULL,
  `assertion_ids_json` TEXT NOT NULL,
  `patch_json` TEXT NOT NULL,
  `status` VARCHAR(32) NOT NULL,
  `content_digest` VARCHAR(64) NOT NULL,
  `created_by` VARCHAR(64) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_business_model_proposal` (`proposal_id`),
  KEY `idx_forma_proposal_business` (`tenant_id`, `business_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_revision_provenance` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `revision_no` INT NOT NULL,
  `proposal_id` VARCHAR(64) NOT NULL,
  `assertion_ids_json` TEXT NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_revision_provenance` (`tenant_id`, `business_id`, `revision_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `forma_analyst_model_call` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `request_id` VARCHAR(64) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL,
  `business_id` VARCHAR(64) NOT NULL,
  `session_id` VARCHAR(64) NOT NULL,
  `operation` VARCHAR(64) NOT NULL,
  `model_ref` VARCHAR(128) NOT NULL DEFAULT '',
  `latency_ms` INT NOT NULL DEFAULT 0,
  `success` TINYINT(1) NOT NULL DEFAULT 0,
  `input_tokens` INT NOT NULL DEFAULT 0,
  `output_tokens` INT NOT NULL DEFAULT 0,
  `error_message` VARCHAR(512) NOT NULL DEFAULT '',
  `created_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_forma_analyst_model_call` (`request_id`),
  KEY `idx_forma_model_call_session` (`tenant_id`, `session_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
