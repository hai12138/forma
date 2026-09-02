/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"fmt"
	"strings"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
)

// NormalizePhysicalDataType maps a native/physical type string to a Forma logical type.
// Returns "" when the input cannot be classified.
func NormalizePhysicalDataType(native string) string {
	s := strings.ToLower(strings.TrimSpace(native))
	if s == "" {
		return ""
	}
	// Check datetime/timestamp before date so "datetime" is not classified as DATE.
	if strings.Contains(s, "datetime") || strings.Contains(s, "timestamp") {
		return "DATETIME"
	}
	if s == "date" || strings.HasPrefix(s, "date") {
		return "DATE"
	}
	if s == "time" || (strings.Contains(s, "time") && !strings.Contains(s, "datetime") && !strings.Contains(s, "timestamp")) {
		return "TIME"
	}
	switch {
	case strings.Contains(s, "json") || strings.Contains(s, "object") || strings.Contains(s, "array"):
		return "JSON"
	case strings.Contains(s, "bool"):
		return "BOOLEAN"
	case strings.Contains(s, "bigint") || strings.Contains(s, "smallint") || strings.Contains(s, "tinyint") ||
		s == "int" || strings.Contains(s, "integer") || strings.HasPrefix(s, "int"):
		return "INTEGER"
	case strings.Contains(s, "decimal") || strings.Contains(s, "numeric") || strings.Contains(s, "float") ||
		strings.Contains(s, "double") || strings.Contains(s, "real") || s == "number" || strings.Contains(s, "number"):
		return "DECIMAL"
	case strings.Contains(s, "varchar") || strings.Contains(s, "char") || strings.Contains(s, "text") ||
		s == "string" || strings.Contains(s, "string"):
		return "STRING"
	}
	return ""
}

// ResolveMappingOutputContractType derives the logical output type guaranteed by a mapping
// against a physical schema.
func ResolveMappingOutputContractType(mapping *entity.SemanticMapping, schema *entity.PhysicalSchema) (string, error) {
	if mapping == nil || schema == nil {
		return "", entity.ErrContractInvalidPayload
	}
	switch mapping.MappingType {
	case entity.MappingTypeDirect, entity.MappingTypeFieldPath:
		field, err := requireSingleTarget(mapping.TargetFieldPaths, schema)
		if err != nil {
			return "", err
		}
		return NormalizePhysicalDataType(field.DataType), nil
	case entity.MappingTypeCast:
		var spec entity.CastTransformSpec
		if err := decodeStrictJSON(mapping.TransformSpec, &spec); err != nil {
			return "", entity.ErrContractInvalidPayload
		}
		to := normalizedType(spec.ToType)
		if !isAllowedMappingType(to) {
			return "", entity.ErrContractInvalidPayload
		}
		return to, nil
	case entity.MappingTypeEnumMap:
		return "STRING", nil
	case entity.MappingTypeUnitConvert:
		return "DECIMAL", nil
	case entity.MappingTypeTimeNormalize:
		return "DATETIME", nil
	case entity.MappingTypeJoinRef:
		if len(mapping.TargetFieldPaths) == 0 {
			return "", entity.ErrContractInvalidPayload
		}
		if len(mapping.TargetFieldPaths) != 1 {
			return "", fmt.Errorf("%w: JOIN_REF requires exactly one target path", entity.ErrContractInvalidPayload)
		}
		field, err := requireSingleTarget(mapping.TargetFieldPaths, schema)
		if err != nil {
			return "", err
		}
		return NormalizePhysicalDataType(field.DataType), nil
	default:
		return "", entity.ErrContractInvalidPayload
	}
}
