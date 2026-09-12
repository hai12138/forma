-- Forma S5-G2-F3: Capability analysis attempt safety (completed_at + unique run/attempt).
ALTER TABLE `forma_capability_analysis_attempt`
  ADD COLUMN `completed_at` DATETIME(3) NULL AFTER `created_at`;

ALTER TABLE `forma_capability_analysis_attempt`
  ADD UNIQUE KEY `uk_forma_capability_analysis_attempt_run_attempt` (`tenant_id`, `analysis_run_id`, `attempt`);
