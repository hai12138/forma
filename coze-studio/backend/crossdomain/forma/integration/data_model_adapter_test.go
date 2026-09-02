/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package integration

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	datasvc "github.com/coze-dev/coze-studio/backend/domain/forma/data/service"
)

func TestParseDataRequirementProposals(t *testing.T) {
	proposals, err := parseDataRequirementProposals(`{
		"requirements": [{
			"requirement_kind": "ENTITY",
			"semantic_name": "customer_profile",
			"description": "Customer profile data",
			"business_element_refs": ["customer"],
			"requiredness": "REQUIRED",
			"freshness_requirement": "DAILY",
			"access_need": "READ"
		}]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(proposals) != 1 || proposals[0].SemanticName != "customer_profile" {
		t.Fatalf("unexpected proposals: %+v", proposals)
	}
	if len(proposals[0].BusinessElementRefs) != 1 || proposals[0].BusinessElementRefs[0] != "customer" {
		t.Fatalf("unexpected references: %+v", proposals[0].BusinessElementRefs)
	}
}

func TestSuggestSemanticMappingsRejectsEmptyInputWithoutModelCall(t *testing.T) {
	model := &CozeEinoDataModel{EnvPrefix: "FORMA_DATA"}
	_, err := model.SuggestSemanticMappings(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "semantic model, requirements, and schemas required") {
		t.Fatalf("expected request validation, got %v", err)
	}
}

func TestSuggestSemanticMappingsRequiresSemanticModel(t *testing.T) {
	model := &CozeEinoDataModel{EnvPrefix: "FORMA_DATA"}
	_, err := model.SuggestSemanticMappings(context.Background(), &datasvc.SuggestSemanticMappingsRequest{
		Requirements:    []datasvc.MappingRequirementMetadata{{RequirementID: "req"}},
		SchemaSnapshots: []datasvc.NormalizedSchemaSnapshot{{SchemaSnapshotID: "snap"}},
	})
	if err == nil || !strings.Contains(err.Error(), "semantic model") {
		t.Fatalf("expected semantic model validation, got %v", err)
	}
}

func TestSemanticMappingPayloadContainsCanonicalContext(t *testing.T) {
	req := &datasvc.SuggestSemanticMappingsRequest{
		BusinessID:            "lab",
		BusinessModelRevision: 7,
		SemanticModel: &businessentity.SemanticModel{
			SchemaVersion: businessentity.SemanticSchemaVersion,
			Nodes:         []businessentity.SemanticNode{{ID: "sample", Name: "Sample"}},
		},
		Requirements:    []datasvc.MappingRequirementMetadata{{RequirementID: "req"}},
		SchemaSnapshots: []datasvc.NormalizedSchemaSnapshot{{SchemaSnapshotID: "snap"}},
	}
	raw, err := semanticMappingPayload(req)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"semantic_model", "requirements", "schema_snapshots", "business_id", "business_model_revision"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("payload missing %q: %s", key, raw)
		}
	}
}
