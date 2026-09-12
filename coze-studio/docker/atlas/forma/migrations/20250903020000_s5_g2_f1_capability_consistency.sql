-- Forma S5-G2-F1: Capability aggregate generation for Activate CAS.
ALTER TABLE forma_business_capability
  ADD COLUMN aggregate_generation BIGINT NOT NULL DEFAULT 0 AFTER active_revision_id;
