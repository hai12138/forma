/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"testing"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
)

func TestValidateLogicalSchemaIndependenceRejectsPhysicalKeys(t *testing.T) {
	ok := entity.ContractLogicalSchema{Fields: []entity.LogicalField{{
		LogicalKey: "temperature", SemanticName: "temperature", LogicalType: "DECIMAL",
		RequirementID: "req1", Classification: entity.DataClassificationInternal,
	}}}
	if err := ValidateLogicalSchemaIndependence(ok); err != nil {
		t.Fatal(err)
	}
	tree := map[string]any{"fields": []any{map[string]any{"logical_key": "a", "table": "readings"}}}
	if err := walkRejectForbiddenKeys(tree); err == nil {
		t.Fatal("expected forbidden physical key")
	}
}

func TestValidateQueryCapabilitiesDeniesWrite(t *testing.T) {
	err := ValidateQueryCapabilities([]entity.QueryCapability{"WRITE"})
	if err == nil {
		t.Fatal("expected write denial")
	}
	err = ValidateQueryCapabilities([]entity.QueryCapability{"CREATE", entity.QueryCapabilityRead})
	if err == nil {
		t.Fatal("expected create denial")
	}
	if err := ValidateQueryCapabilities([]entity.QueryCapability{entity.QueryCapabilityRead, entity.QueryCapabilityList}); err != nil {
		t.Fatal(err)
	}
}
