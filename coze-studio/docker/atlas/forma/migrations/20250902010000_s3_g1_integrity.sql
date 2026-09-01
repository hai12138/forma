-- Forma S3-G1: integrity constraints, session turn sequence allocator, turn lineage.
-- Does not modify prior frozen migrations.

ALTER TABLE `forma_analyst_session`
  ADD COLUMN `next_turn_sequence` INT NOT NULL DEFAULT 1;

ALTER TABLE `forma_analyst_turn`
  ADD COLUMN `reply_to_turn_id` VARCHAR(64) NOT NULL DEFAULT '',
  ADD COLUMN `reserved_reply_sequence` INT NOT NULL DEFAULT 0;

ALTER TABLE `forma_analyst_turn`
  ADD UNIQUE KEY `uk_forma_analyst_turn_sequence` (`tenant_id`, `session_id`, `sequence`);

ALTER TABLE `forma_assertion_conflict`
  ADD UNIQUE KEY `uk_forma_assertion_conflict_pair` (`tenant_id`, `business_id`, `assertion_id_a`, `assertion_id_b`);
