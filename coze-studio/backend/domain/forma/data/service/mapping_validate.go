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

func buildFieldPathIndex(schema *entity.PhysicalSchema) map[string]entity.PhysicalField {
	out := map[string]entity.PhysicalField{}
	if schema == nil {
		return out
	}
	for _, field := range schema.Fields {
		path := strings.TrimSpace(field.Path)
		if path == "" {
			path = strings.TrimSpace(field.Name)
		}
		if path != "" {
			out[path] = field
		}
	}
	return out
}

func ValidateMappingTarget(paths []string, schema *entity.PhysicalSchema) error {
	if len(paths) == 0 || schema == nil {
		return entity.ErrMappingTargetInvalid
	}
	index := buildFieldPathIndex(schema)
	seen := map[string]struct{}{}
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if _, ok := index[path]; !ok {
			return fmt.Errorf("%w: field path %q", entity.ErrMappingTargetInvalid, path)
		}
		if _, ok := seen[path]; ok {
			return fmt.Errorf("%w: duplicate field path %q", entity.ErrMappingTargetInvalid, path)
		}
		seen[path] = struct{}{}
	}
	return nil
}

func ValidateTransformSpec(mappingType entity.MappingType, raw json.RawMessage) error {
	if !json.Valid(raw) {
		return entity.ErrMappingTransformInvalid
	}
	var header struct {
		Type entity.MappingType `json:"type"`
	}
	if json.Unmarshal(raw, &header) != nil || header.Type != mappingType {
		return entity.ErrMappingTransformInvalid
	}
	valid := func(err error, required ...string) error {
		if err != nil {
			return entity.ErrMappingTransformInvalid
		}
		for _, value := range required {
			if strings.TrimSpace(value) == "" {
				return entity.ErrMappingTransformInvalid
			}
		}
		return nil
	}
	switch mappingType {
	case entity.MappingTypeDirect:
		var v entity.DirectTransformSpec
		return valid(json.Unmarshal(raw, &v))
	case entity.MappingTypeCast:
		var v entity.CastTransformSpec
		return valid(json.Unmarshal(raw, &v), v.To)
	case entity.MappingTypeEnumMap:
		var v entity.EnumMapTransformSpec
		if err := json.Unmarshal(raw, &v); err != nil || len(v.Values) == 0 {
			return entity.ErrMappingTransformInvalid
		}
	case entity.MappingTypeUnitConvert:
		var v entity.UnitConvertTransformSpec
		return valid(json.Unmarshal(raw, &v), v.From, v.To)
	case entity.MappingTypeTimeNormalize:
		var v entity.TimeNormalizeTransformSpec
		return valid(json.Unmarshal(raw, &v), v.InputFormat, v.OutputFormat)
	case entity.MappingTypeFieldPath:
		var v entity.FieldPathTransformSpec
		return valid(json.Unmarshal(raw, &v), v.Path)
	case entity.MappingTypeJoinRef:
		var v entity.JoinRefTransformSpec
		if err := json.Unmarshal(raw, &v); err != nil || v.Relationship == "" || v.ToSchema == "" || len(v.FromFields) == 0 || len(v.FromFields) != len(v.ToFields) {
			return entity.ErrMappingTransformInvalid
		}
	default:
		return entity.ErrMappingTransformInvalid
	}
	return nil
}

func ValidateJoinRef(raw json.RawMessage, schema *entity.PhysicalSchema) error {
	var spec entity.JoinRefTransformSpec
	if err := json.Unmarshal(raw, &spec); err != nil || schema == nil {
		return entity.ErrMappingTransformInvalid
	}
	for _, rel := range schema.Relationships {
		if rel.Name == spec.Relationship && rel.ToSchema == spec.ToSchema && stringSlicesEqual(rel.FromFields, spec.FromFields) && stringSlicesEqual(rel.ToFields, spec.ToFields) {
			return nil
		}
	}
	return entity.ErrMappingLineageInvalid
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
