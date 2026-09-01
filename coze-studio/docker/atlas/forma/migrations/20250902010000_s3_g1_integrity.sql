-- Forma S3-G1: integrity constraints for analyst idempotency and conflict dedup.
-- Does not modify prior frozen migrations.

CREATE UNIQUE INDEX IF NOT EXISTS `uk_forma_analyst_turn_sequence`
  ON `forma_analyst_turn` (`tenant_id`, `session_id`, `sequence`);

CREATE UNIQUE INDEX IF NOT EXISTS `uk_forma_assertion_conflict_pair`
  ON `forma_assertion_conflict` (`tenant_id`, `business_id`, `assertion_id_a`, `assertion_id_b`);
