-- Forma S1-G1: one row per coze_space_id for tenant space mapping.
-- Collapse duplicate coze_space_id rows keeping the highest id,
-- then replace (coze_space_id, status) uniqueness with UNIQUE(coze_space_id).

DELETE t1 FROM `forma_tenant_space_ref` t1
INNER JOIN `forma_tenant_space_ref` t2
  ON t1.`coze_space_id` = t2.`coze_space_id`
 AND t1.`id` < t2.`id`;

ALTER TABLE `forma_tenant_space_ref`
  DROP INDEX `uk_forma_space_active`,
  ADD UNIQUE KEY `uk_forma_space_id` (`coze_space_id`);
