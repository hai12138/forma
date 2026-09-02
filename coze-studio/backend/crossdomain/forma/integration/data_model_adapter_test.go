/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package integration

import (
	"context"
	"strings"
	"testing"
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
	if err == nil || !strings.Contains(err.Error(), "requirements and schemas required") {
		t.Fatalf("expected request validation, got %v", err)
	}
}
