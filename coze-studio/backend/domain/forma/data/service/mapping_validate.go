/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
)

func buildFieldPathIndex(schema *entity.PhysicalSchema) map[string]entity.PhysicalField {
	out := map[string]entity.PhysicalField{}
	if schema == nil {
		return out
	}
	for _, field := range schema.Fields {
		path := strings.TrimSpace(field.Path)
		if path != "" {
			out[path] = field
		}
	}
	return out
}

func ValidateSchemaPaths(schema *entity.PhysicalSchema) error {
	if schema == nil {
		return entity.ErrMappingTargetInvalid
	}
	seen := make(map[string]struct{}, len(schema.Fields))
	for _, field := range schema.Fields {
		path := strings.TrimSpace(field.Path)
		if path == "" {
			return fmt.Errorf("%w: physical field %q has no canonical path", entity.ErrMappingTargetInvalid, field.Name)
		}
		if _, ok := seen[path]; ok {
			return fmt.Errorf("%w: duplicate schema field path %q", entity.ErrMappingTargetInvalid, path)
		}
		seen[path] = struct{}{}
	}
	return nil
}

func ValidateMappingTarget(paths []string, schema *entity.PhysicalSchema) error {
	if err := ValidateSchemaPaths(schema); err != nil {
		return err
	}
	if len(paths) == 0 {
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

var normalizedMappingTypes = map[string]struct{}{
	"STRING": {}, "INTEGER": {}, "DECIMAL": {}, "BOOLEAN": {},
	"DATE": {}, "DATETIME": {}, "TIME": {}, "JSON": {},
}

func normalizedType(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func isAllowedMappingType(value string) bool {
	_, ok := normalizedMappingTypes[normalizedType(value)]
	return ok
}

func requireSingleTarget(paths []string, schema *entity.PhysicalSchema) (entity.PhysicalField, error) {
	if len(paths) != 1 {
		return entity.PhysicalField{}, entity.ErrMappingTransformInvalid
	}
	if err := ValidateMappingTarget(paths, schema); err != nil {
		return entity.PhysicalField{}, err
	}
	return buildFieldPathIndex(schema)[strings.TrimSpace(paths[0])], nil
}

func ValidateTransformSpec(mappingType entity.MappingType, raw json.RawMessage, paths []string, schema *entity.PhysicalSchema) error {
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
		if _, err := requireSingleTarget(paths, schema); err != nil {
			return err
		}
		var v entity.DirectTransformSpec
		dec := json.NewDecoder(strings.NewReader(string(raw)))
		dec.DisallowUnknownFields()
		return valid(dec.Decode(&v))
	case entity.MappingTypeCast:
		field, err := requireSingleTarget(paths, schema)
		if err != nil {
			return err
		}
		var v entity.CastTransformSpec
		if err := valid(json.Unmarshal(raw, &v), v.FromType, v.ToType); err != nil || !isAllowedMappingType(v.ToType) {
			return entity.ErrMappingTransformInvalid
		}
		if strings.TrimSpace(field.DataType) == "" {
			if !isAllowedMappingType(v.FromType) {
				return entity.ErrMappingTransformInvalid
			}
		} else if !strings.EqualFold(strings.TrimSpace(field.DataType), strings.TrimSpace(v.FromType)) {
			return entity.ErrMappingTransformInvalid
		}
		return nil
	case entity.MappingTypeEnumMap:
		if _, err := requireSingleTarget(paths, schema); err != nil {
			return err
		}
		var v entity.EnumMapTransformSpec
		if err := json.Unmarshal(raw, &v); err != nil || len(v.Pairs) == 0 {
			return entity.ErrMappingTransformInvalid
		}
		seenKeys := make(map[string]struct{}, len(v.Pairs))
		for key, value := range v.Pairs {
			if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
				return entity.ErrMappingTransformInvalid
			}
			if _, ok := seenKeys[key]; ok {
				return entity.ErrMappingTransformInvalid
			}
			seenKeys[key] = struct{}{}
		}
		return nil
	case entity.MappingTypeUnitConvert:
		if _, err := requireSingleTarget(paths, schema); err != nil {
			return err
		}
		var v entity.UnitConvertTransformSpec
		if err := valid(json.Unmarshal(raw, &v), v.FromUnit, v.ToUnit); err != nil || math.IsNaN(v.Factor) || math.IsInf(v.Factor, 0) || v.Factor == 0 {
			return entity.ErrMappingTransformInvalid
		}
		return nil
	case entity.MappingTypeTimeNormalize:
		if _, err := requireSingleTarget(paths, schema); err != nil {
			return err
		}
		var v entity.TimeNormalizeTransformSpec
		if err := valid(json.Unmarshal(raw, &v), v.SourceTimezone, v.TargetTimezone, v.Format); err != nil {
			return err
		}
		if _, err := time.LoadLocation(v.SourceTimezone); err != nil {
			return entity.ErrMappingTransformInvalid
		}
		if _, err := time.LoadLocation(v.TargetTimezone); err != nil {
			return entity.ErrMappingTransformInvalid
		}
		return nil
	case entity.MappingTypeFieldPath:
		if _, err := requireSingleTarget(paths, schema); err != nil {
			return err
		}
		var v entity.FieldPathTransformSpec
		if err := valid(json.Unmarshal(raw, &v), v.Path); err != nil || strings.TrimSpace(v.Path) != strings.TrimSpace(paths[0]) {
			return entity.ErrMappingTransformInvalid
		}
		return nil
	case entity.MappingTypeJoinRef:
		var v entity.JoinRefTransformSpec
		if err := json.Unmarshal(raw, &v); err != nil || v.Relationship == "" || v.ToSchema == "" || len(v.FromFields) == 0 || len(v.FromFields) != len(v.ToFields) {
			return entity.ErrMappingTransformInvalid
		}
		return ValidateJoinRef(raw, schema)
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
			for _, fromField := range rel.FromFields {
				if !resolvesRelationshipField(fromField, schema.Fields) {
					return entity.ErrMappingLineageInvalid
				}
			}
			return nil
		}
	}
	return entity.ErrMappingLineageInvalid
}

func resolvesRelationshipField(fromField string, fields []entity.PhysicalField) bool {
	needle := strings.TrimSpace(fromField)
	for _, field := range fields {
		path := strings.TrimSpace(field.Path)
		if path == "" {
			continue
		}
		if needle == path || needle == strings.TrimSpace(field.Name) {
			return true
		}
	}
	return false
}

func ValidateConfidence(c float64, required bool) error {
	if math.IsNaN(c) || math.IsInf(c, 0) || c < 0 || c > 1 {
		return entity.ErrMappingTransformInvalid
	}
	if !required && c == 0 {
		return nil
	}
	return nil
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
