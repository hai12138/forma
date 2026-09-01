-- S3-G2: session focused gap for gap-driven interview follow-up
ALTER TABLE `forma_analyst_session`
  ADD COLUMN `focus_gap_id` VARCHAR(64) NOT NULL DEFAULT '' AFTER `next_turn_sequence`;
