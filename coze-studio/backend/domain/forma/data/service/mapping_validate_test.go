/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
)

func validationSchema() *entity.PhysicalSchema {
	return &entity.PhysicalSchema{
		Name: "readings",
		Fields: []entity.PhysicalField{
			{Name: "temperature", Path: "sensor.temperature", DataType: "DECIMAL"},
			{Name: "recorded_at", Path: "sensor.recorded_at", DataType: "DATETIME"},
			{Name: "sensor_id", Path: "sensor.id", DataType: "STRING"},
		},
		Relationships: []entity.PhysicalRelationship{{
			Name: "sensor", FromFields: []string{"sensor_id"}, ToSchema: "sensors", ToFields: []string{"id"},
		}},
	}
}

func TestValidateMappingTargetRequiresCanonicalUniquePaths(t *testing.T) {
	tests := []struct {
		name   string
		paths  []string
		schema *entity.PhysicalSchema
	}{
		{name: "empty request", paths: nil, schema: validationSchema()},
		{name: "name is not path", paths: []string{"temperature"}, schema: validationSchema()},
		{name: "duplicate request path", paths: []string{"sensor.temperature", "sensor.temperature"}, schema: validationSchema()},
		{name: "empty schema path", paths: []string{"sensor.temperature"}, schema: &entity.PhysicalSchema{Fields: []entity.PhysicalField{{Name: "bad", Path: " "}, {Path: "sensor.temperature"}}}},
		{name: "duplicate schema path", paths: []string{"sensor.temperature"}, schema: &entity.PhysicalSchema{Fields: []entity.PhysicalField{{Path: "sensor.temperature"}, {Path: " sensor.temperature "}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateMappingTarget(tt.paths, tt.schema); !errors.Is(err, entity.ErrMappingTargetInvalid) {
				t.Fatalf("expected target error, got %v", err)
			}
		})
	}
	if err := ValidateMappingTarget([]string{"sensor.temperature"}, validationSchema()); err != nil {
		t.Fatalf("canonical path rejected: %v", err)
	}
}

func TestValidateTransformSpecContract(t *testing.T) {
	schema := validationSchema()
	tests := []struct {
		name        string
		mappingType entity.MappingType
		raw         string
		paths       []string
		wantErr     error
	}{
		{"direct", entity.MappingTypeDirect, `{"type":"DIRECT"}`, []string{"sensor.temperature"}, nil},
		{"direct zero target", entity.MappingTypeDirect, `{"type":"DIRECT"}`, []string{}, entity.ErrMappingTransformInvalid},
		{"direct multi target", entity.MappingTypeDirect, `{"type":"DIRECT"}`, []string{"sensor.temperature", "sensor.recorded_at"}, entity.ErrMappingTransformInvalid},
		{"direct extra field", entity.MappingTypeDirect, `{"type":"DIRECT","expr":"x"}`, []string{"sensor.temperature"}, entity.ErrMappingTransformInvalid},
		{"field path", entity.MappingTypeFieldPath, `{"type":"FIELD_PATH","path":"sensor.temperature"}`, []string{"sensor.temperature"}, nil},
		{"field path mismatch", entity.MappingTypeFieldPath, `{"type":"FIELD_PATH","path":"sensor.id"}`, []string{"sensor.temperature"}, entity.ErrMappingTransformInvalid},
		{"cast", entity.MappingTypeCast, `{"type":"CAST","from_type":"decimal","to_type":"STRING"}`, []string{"sensor.temperature"}, nil},
		{"cast source mismatch", entity.MappingTypeCast, `{"type":"CAST","from_type":"INTEGER","to_type":"STRING"}`, []string{"sensor.temperature"}, entity.ErrMappingTransformInvalid},
		{"cast unsupported target", entity.MappingTypeCast, `{"type":"CAST","from_type":"DECIMAL","to_type":"BINARY"}`, []string{"sensor.temperature"}, entity.ErrMappingTransformInvalid},
		{"enum map", entity.MappingTypeEnumMap, `{"type":"ENUM_MAP","pairs":{"yes":"true"}}`, []string{"sensor.temperature"}, nil},
		{"enum map empty", entity.MappingTypeEnumMap, `{"type":"ENUM_MAP","pairs":{}}`, []string{"sensor.temperature"}, entity.ErrMappingTransformInvalid},
		{"enum map blank value", entity.MappingTypeEnumMap, `{"type":"ENUM_MAP","pairs":{"yes":" "}}`, []string{"sensor.temperature"}, entity.ErrMappingTransformInvalid},
		{"unit convert", entity.MappingTypeUnitConvert, `{"type":"UNIT_CONVERT","from_unit":"C","to_unit":"F","factor":1.8,"offset":32}`, []string{"sensor.temperature"}, nil},
		{"unit zero factor", entity.MappingTypeUnitConvert, `{"type":"UNIT_CONVERT","from_unit":"C","to_unit":"F","factor":0,"offset":32}`, []string{"sensor.temperature"}, entity.ErrMappingTransformInvalid},
		{"unit inf factor", entity.MappingTypeUnitConvert, `{"type":"UNIT_CONVERT","from_unit":"C","to_unit":"F","factor":"Inf","offset":32}`, []string{"sensor.temperature"}, entity.ErrMappingTransformInvalid},
		{"time normalize", entity.MappingTypeTimeNormalize, `{"type":"TIME_NORMALIZE","source_timezone":"Asia/Shanghai","target_timezone":"UTC","format":"2006-01-02T15:04:05Z07:00"}`, []string{"sensor.recorded_at"}, nil},
		{"time bad timezone", entity.MappingTypeTimeNormalize, `{"type":"TIME_NORMALIZE","source_timezone":"Mars/Base","target_timezone":"UTC","format":"RFC3339"}`, []string{"sensor.recorded_at"}, entity.ErrMappingTransformInvalid},
		{"join invented", entity.MappingTypeJoinRef, `{"type":"JOIN_REF","relationship":"ghost","from_fields":["sensor_id"],"to_schema":"sensors","to_fields":["id"]}`, []string{"sensor.id"}, entity.ErrMappingLineageInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTransformSpec(tt.mappingType, json.RawMessage(tt.raw), tt.paths, schema)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateJoinRefRequiresCanonicalPhysicalFromFields(t *testing.T) {
	schema := validationSchema()
	valid := json.RawMessage(`{"type":"JOIN_REF","relationship":"sensor","from_fields":["sensor_id"],"to_schema":"sensors","to_fields":["id"]}`)
	if err := ValidateTransformSpec(entity.MappingTypeJoinRef, valid, []string{"sensor.id"}, schema); err != nil {
		t.Fatalf("valid join rejected: %v", err)
	}
	schema.Fields[2].Path = ""
	if err := ValidateTransformSpec(entity.MappingTypeJoinRef, valid, []string{"sensor.temperature"}, schema); !errors.Is(err, entity.ErrMappingTargetInvalid) && !errors.Is(err, entity.ErrMappingLineageInvalid) {
		t.Fatalf("expected invalid join field, got %v", err)
	}
}

func TestValidateConfidence(t *testing.T) {
	for _, c := range []float64{0, .5, 1} {
		if err := ValidateConfidence(c, true); err != nil {
			t.Fatalf("valid required confidence %v rejected: %v", c, err)
		}
	}
	if err := ValidateConfidence(0, false); err != nil {
		t.Fatalf("omitted manual confidence rejected: %v", err)
	}
	for _, c := range []float64{-0.1, 1.1} {
		if err := ValidateConfidence(c, true); !errors.Is(err, entity.ErrMappingTransformInvalid) {
			t.Fatalf("expected confidence error for %v, got %v", c, err)
		}
	}
}
