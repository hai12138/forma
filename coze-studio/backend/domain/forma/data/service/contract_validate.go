/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
)

// ForbiddenPhysicalLogicalKeys are JSON object keys that must never appear in a logical schema.
var ForbiddenPhysicalLogicalKeys = []string{
	"source_id",
	"connection_id",
	"asset_id",
	"schema_snapshot_id",
	"table",
	"column",
	"endpoint",
	"physical_path",
}

var allowedQueryCapabilities = map[entity.QueryCapability]struct{}{
	entity.QueryCapabilityRead:      {},
	entity.QueryCapabilityLookup:    {},
	entity.QueryCapabilityList:      {},
	entity.QueryCapabilityFilter:    {},
	entity.QueryCapabilityAggregate: {},
}

var allowedClassifications = map[entity.DataClassification]struct{}{
	entity.DataClassificationPublic:       {},
	entity.DataClassificationInternal:     {},
	entity.DataClassificationConfidential: {},
	entity.DataClassificationPII:          {},
	entity.DataClassificationSecret:       {},
}

var allowedFreshness = map[entity.FreshnessPolicy]struct{}{
	entity.FreshnessPolicyRealtime:     {},
	entity.FreshnessPolicyNearRealtime: {},
	entity.FreshnessPolicyHourly:       {},
	entity.FreshnessPolicyDaily:        {},
	entity.FreshnessPolicyOnDemand:     {},
}

var operatorsByLogicalType = map[string]map[entity.FilterOperator]struct{}{
	"STRING": {
		entity.FilterOperatorEQ: {}, entity.FilterOperatorNE: {},
		entity.FilterOperatorIN: {}, entity.FilterOperatorContains: {},
	},
	"INTEGER": {
		entity.FilterOperatorEQ: {}, entity.FilterOperatorNE: {},
		entity.FilterOperatorGT: {}, entity.FilterOperatorGTE: {},
		entity.FilterOperatorLT: {}, entity.FilterOperatorLTE: {},
		entity.FilterOperatorIN: {}, entity.FilterOperatorBetween: {},
	},
	"DECIMAL": {
		entity.FilterOperatorEQ: {}, entity.FilterOperatorNE: {},
		entity.FilterOperatorGT: {}, entity.FilterOperatorGTE: {},
		entity.FilterOperatorLT: {}, entity.FilterOperatorLTE: {},
		entity.FilterOperatorIN: {}, entity.FilterOperatorBetween: {},
	},
	"BOOLEAN": {
		entity.FilterOperatorEQ: {}, entity.FilterOperatorNE: {},
	},
	"DATE": {
		entity.FilterOperatorEQ: {}, entity.FilterOperatorNE: {},
		entity.FilterOperatorGT: {}, entity.FilterOperatorGTE: {},
		entity.FilterOperatorLT: {}, entity.FilterOperatorLTE: {},
		entity.FilterOperatorIN: {}, entity.FilterOperatorBetween: {},
	},
	"DATETIME": {
		entity.FilterOperatorEQ: {}, entity.FilterOperatorNE: {},
		entity.FilterOperatorGT: {}, entity.FilterOperatorGTE: {},
		entity.FilterOperatorLT: {}, entity.FilterOperatorLTE: {},
		entity.FilterOperatorIN: {}, entity.FilterOperatorBetween: {},
	},
	"TIME": {
		entity.FilterOperatorEQ: {}, entity.FilterOperatorNE: {},
		entity.FilterOperatorGT: {}, entity.FilterOperatorGTE: {},
		entity.FilterOperatorLT: {}, entity.FilterOperatorLTE: {},
		entity.FilterOperatorIN: {}, entity.FilterOperatorBetween: {},
	},
	"JSON": {
		entity.FilterOperatorEQ: {}, entity.FilterOperatorNE: {},
		entity.FilterOperatorContains: {},
	},
}

// ValidateLogicalSchemaIndependence rejects physical binding keys leaking into logical schema JSON.
func ValidateLogicalSchemaIndependence(schema entity.ContractLogicalSchema) error {
	raw, err := json.Marshal(schema)
	if err != nil {
		return entity.ErrContractLogicalSchemaInvalid
	}
	lower := strings.ToLower(string(raw))
	for _, key := range ForbiddenPhysicalLogicalKeys {
		needle := `"` + strings.ToLower(key) + `"`
		if strings.Contains(lower, needle) {
			return fmt.Errorf("%w: forbidden physical key %q", entity.ErrContractLogicalSchemaInvalid, key)
		}
	}
	var tree any
	if err := json.Unmarshal(raw, &tree); err != nil {
		return entity.ErrContractLogicalSchemaInvalid
	}
	if err := walkRejectForbiddenKeys(tree); err != nil {
		return err
	}
	return nil
}

func walkRejectForbiddenKeys(node any) error {
	switch v := node.(type) {
	case map[string]any:
		for key, child := range v {
			for _, forbidden := range ForbiddenPhysicalLogicalKeys {
				if strings.EqualFold(key, forbidden) {
					return fmt.Errorf("%w: forbidden physical key %q", entity.ErrContractLogicalSchemaInvalid, key)
				}
			}
			if err := walkRejectForbiddenKeys(child); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range v {
			if err := walkRejectForbiddenKeys(child); err != nil {
				return err
			}
		}
	}
	return nil
}

// ValidateLogicalSchema checks field identity, types, classification, and requirement pinning.
func ValidateLogicalSchema(schema entity.ContractLogicalSchema, requirementIDs map[string]struct{}) error {
	if err := ValidateLogicalSchemaIndependence(schema); err != nil {
		return err
	}
	if len(schema.Fields) == 0 {
		return fmt.Errorf("%w: logical schema requires at least one field", entity.ErrContractLogicalSchemaInvalid)
	}
	seenKeys := map[string]struct{}{}
	seenReqs := map[string]struct{}{}
	for _, field := range schema.Fields {
		key := strings.TrimSpace(field.LogicalKey)
		if key == "" {
			return fmt.Errorf("%w: empty logical_key", entity.ErrContractLogicalSchemaInvalid)
		}
		if _, ok := seenKeys[key]; ok {
			return fmt.Errorf("%w: duplicate logical_key %q", entity.ErrContractLogicalSchemaInvalid, key)
		}
		seenKeys[key] = struct{}{}
		if strings.TrimSpace(field.SemanticName) == "" {
			return fmt.Errorf("%w: empty semantic_name for %q", entity.ErrContractLogicalSchemaInvalid, key)
		}
		if !IsAllowedLogicalType(field.LogicalType) {
			return fmt.Errorf("%w: invalid logical_type %q", entity.ErrContractLogicalSchemaInvalid, field.LogicalType)
		}
		if _, ok := allowedClassifications[field.Classification]; !ok {
			return fmt.Errorf("%w: invalid classification %q", entity.ErrContractLogicalSchemaInvalid, field.Classification)
		}
		reqID := strings.TrimSpace(field.RequirementID)
		if reqID == "" {
			return fmt.Errorf("%w: field %q missing requirement_id", entity.ErrContractLogicalSchemaInvalid, key)
		}
		if requirementIDs != nil {
			if _, ok := requirementIDs[reqID]; !ok {
				return fmt.Errorf("%w: field %q references unknown requirement %q", entity.ErrContractLogicalSchemaInvalid, key, reqID)
			}
		}
		if _, ok := seenReqs[reqID]; ok {
			return fmt.Errorf("%w: requirement %q mapped to multiple logical fields", entity.ErrContractLogicalSchemaInvalid, reqID)
		}
		seenReqs[reqID] = struct{}{}
	}
	return nil
}

func ValidateQueryCapabilities(caps []entity.QueryCapability) error {
	if len(caps) == 0 {
		return fmt.Errorf("%w: query_capabilities required", entity.ErrContractInvalidPayload)
	}
	seen := map[entity.QueryCapability]struct{}{}
	for _, cap := range caps {
		if _, ok := allowedQueryCapabilities[cap]; !ok {
			return fmt.Errorf("%w: forbidden query capability %q", entity.ErrContractInvalidPayload, cap)
		}
		if _, ok := seen[cap]; ok {
			return fmt.Errorf("%w: duplicate query capability %q", entity.ErrContractInvalidPayload, cap)
		}
		seen[cap] = struct{}{}
	}
	return nil
}

func logicalKeyTypes(schema entity.ContractLogicalSchema) map[string]string {
	out := make(map[string]string, len(schema.Fields))
	for _, field := range schema.Fields {
		out[strings.TrimSpace(field.LogicalKey)] = normalizedType(field.LogicalType)
	}
	return out
}

func ValidateFilterSchema(filter entity.FilterSchema, schema entity.ContractLogicalSchema) error {
	types := logicalKeyTypes(schema)
	seen := map[string]struct{}{}
	for _, field := range filter.Fields {
		key := strings.TrimSpace(field.LogicalKey)
		lt, ok := types[key]
		if !ok {
			return fmt.Errorf("%w: filter references unknown logical_key %q", entity.ErrContractInvalidPayload, key)
		}
		if _, dup := seen[key]; dup {
			return fmt.Errorf("%w: duplicate filter logical_key %q", entity.ErrContractInvalidPayload, key)
		}
		seen[key] = struct{}{}
		if len(field.Operators) == 0 {
			return fmt.Errorf("%w: filter %q requires operators", entity.ErrContractInvalidPayload, key)
		}
		allowed := operatorsByLogicalType[lt]
		opSeen := map[entity.FilterOperator]struct{}{}
		for _, op := range field.Operators {
			if _, ok := allowed[op]; !ok {
				return fmt.Errorf("%w: operator %q not allowed for type %q", entity.ErrContractInvalidPayload, op, lt)
			}
			if _, ok := opSeen[op]; ok {
				return fmt.Errorf("%w: duplicate operator %q on %q", entity.ErrContractInvalidPayload, op, key)
			}
			opSeen[op] = struct{}{}
		}
	}
	return nil
}

func ValidateSortSchema(sortSchema entity.SortSchema, schema entity.ContractLogicalSchema) error {
	types := logicalKeyTypes(schema)
	seen := map[string]struct{}{}
	for _, field := range sortSchema.Fields {
		key := strings.TrimSpace(field.LogicalKey)
		if _, ok := types[key]; !ok {
			return fmt.Errorf("%w: sort references unknown logical_key %q", entity.ErrContractInvalidPayload, key)
		}
		if _, dup := seen[key]; dup {
			return fmt.Errorf("%w: duplicate sort logical_key %q", entity.ErrContractInvalidPayload, key)
		}
		seen[key] = struct{}{}
		if len(field.Directions) == 0 {
			return fmt.Errorf("%w: sort %q requires directions", entity.ErrContractInvalidPayload, key)
		}
		dirSeen := map[entity.SortDirection]struct{}{}
		for _, dir := range field.Directions {
			if dir != entity.SortDirectionASC && dir != entity.SortDirectionDESC {
				return fmt.Errorf("%w: invalid sort direction %q", entity.ErrContractInvalidPayload, dir)
			}
			if _, ok := dirSeen[dir]; ok {
				return fmt.Errorf("%w: duplicate sort direction %q on %q", entity.ErrContractInvalidPayload, dir, key)
			}
			dirSeen[dir] = struct{}{}
		}
	}
	return nil
}

func ValidatePaginationPolicy(policy entity.PaginationPolicy) error {
	if policy.DefaultLimit < 1 {
		return fmt.Errorf("%w: default_limit must be >= 1", entity.ErrContractInvalidPayload)
	}
	if policy.MaxLimit > 1000 {
		return fmt.Errorf("%w: max_limit must be <= 1000", entity.ErrContractInvalidPayload)
	}
	if policy.DefaultLimit > policy.MaxLimit {
		return fmt.Errorf("%w: default_limit must be <= max_limit", entity.ErrContractInvalidPayload)
	}
	if policy.MaxLimit < 1 {
		return fmt.Errorf("%w: max_limit must be >= 1", entity.ErrContractInvalidPayload)
	}
	return nil
}

func ValidateClassification(policy map[string]entity.DataClassification, schema entity.ContractLogicalSchema) error {
	logicalKeys := map[string]entity.LogicalField{}
	for _, field := range schema.Fields {
		logicalKeys[strings.TrimSpace(field.LogicalKey)] = field
	}
	for key, class := range policy {
		key = strings.TrimSpace(key)
		for _, forbidden := range ForbiddenPhysicalLogicalKeys {
			if strings.EqualFold(key, forbidden) {
				return fmt.Errorf("%w: classification policy key %q is physical", entity.ErrContractInvalidPayload, key)
			}
		}
		field, ok := logicalKeys[key]
		if !ok {
			return fmt.Errorf("%w: classification policy key %q is not a logical_key", entity.ErrContractInvalidPayload, key)
		}
		if _, ok := allowedClassifications[class]; !ok {
			return fmt.Errorf("%w: invalid classification %q for %q", entity.ErrContractInvalidPayload, class, key)
		}
		if field.Classification != "" && field.Classification != class {
			return fmt.Errorf("%w: classification conflict on %q", entity.ErrContractInvalidPayload, key)
		}
	}
	for _, field := range schema.Fields {
		if _, ok := allowedClassifications[field.Classification]; !ok {
			return fmt.Errorf("%w: invalid field classification %q", entity.ErrContractInvalidPayload, field.Classification)
		}
	}
	return nil
}

// ValidateRequirementIDsUnique requires trimmed non-empty unique requirement IDs.
func ValidateRequirementIDsUnique(ids []string) error {
	seen := map[string]struct{}{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			return fmt.Errorf("%w: empty requirement_id", entity.ErrContractInvalidPayload)
		}
		if _, ok := seen[id]; ok {
			return fmt.Errorf("%w: REQUIREMENT_IDS_DUPLICATE", entity.ErrContractInvalidPayload)
		}
		seen[id] = struct{}{}
	}
	return nil
}

func ValidateFreshnessPolicy(policy entity.FreshnessPolicy) error {
	if _, ok := allowedFreshness[policy]; !ok {
		return fmt.Errorf("%w: invalid freshness_policy %q", entity.ErrContractInvalidPayload, policy)
	}
	return nil
}
