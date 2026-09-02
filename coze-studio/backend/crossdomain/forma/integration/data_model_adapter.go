/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"

	"github.com/coze-dev/coze-studio/backend/bizpkg/llm/modelbuilder"
	dataentity "github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	datasvc "github.com/coze-dev/coze-studio/backend/domain/forma/data/service"
)

// CozeEinoDataModel implements FormaDataModel through Coze Model Manager / Eino.
type CozeEinoDataModel struct {
	EnvPrefix string
}

func NewCozeEinoDataModel(envPrefix string) datasvc.FormaDataModel {
	return &CozeEinoDataModel{EnvPrefix: envPrefix}
}

func (m *CozeEinoDataModel) AnalyzeDataRequirements(ctx context.Context, req *datasvc.AnalyzeDataRequirementsRequest) (*datasvc.AnalyzeDataRequirementsResponse, error) {
	if req == nil || req.SemanticModel == nil {
		return nil, fmt.Errorf("analyze data requirements: request and semantic model required")
	}
	modelJSON, err := json.Marshal(req.SemanticModel)
	if err != nil {
		return nil, fmt.Errorf("marshal semantic model: %w", err)
	}
	model, ok, err := modelbuilder.GetBuiltinChatModel(ctx, m.EnvPrefix)
	if err != nil || !ok || model == nil {
		return nil, fmt.Errorf("builtin chat model not configured")
	}
	out, err := model.Generate(ctx, []*schema.Message{
		schema.SystemMessage(dataRequirementSystemPrompt),
		schema.UserMessage(fmt.Sprintf(
			"Business ID: %s\nBusiness model revision: %d\nSemantic model JSON:\n%s",
			req.BusinessID,
			req.BusinessModelRevision,
			modelJSON,
		)),
	})
	if err != nil {
		return nil, err
	}
	proposals, err := parseDataRequirementProposals(out.Content)
	if err != nil {
		return nil, err
	}
	return &datasvc.AnalyzeDataRequirementsResponse{
		ModelRef:  "coze-eino-builtin",
		Proposals: proposals,
	}, nil
}

func (m *CozeEinoDataModel) SuggestSemanticMappings(ctx context.Context, req *datasvc.SuggestSemanticMappingsRequest) (*datasvc.SuggestSemanticMappingsResponse, error) {
	if req == nil || len(req.Requirements) == 0 || len(req.SchemaSnapshots) == 0 {
		return nil, fmt.Errorf("suggest semantic mappings: requirements and schemas required")
	}
	payload, err := json.Marshal(struct {
		Requirements    any `json:"requirements"`
		SchemaSnapshots any `json:"schema_snapshots"`
	}{req.Requirements, req.SchemaSnapshots})
	if err != nil {
		return nil, err
	}
	model, ok, err := modelbuilder.GetBuiltinChatModel(ctx, m.EnvPrefix)
	if err != nil || !ok || model == nil {
		return nil, fmt.Errorf("builtin chat model not configured")
	}
	out, err := model.Generate(ctx, []*schema.Message{
		schema.SystemMessage(semanticMappingSystemPrompt),
		schema.UserMessage(fmt.Sprintf("Business ID: %s\nBusiness model revision: %d\nInput JSON:\n%s", req.BusinessID, req.BusinessModelRevision, payload)),
	})
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(out.Content), "```json"), "```"), "```"))
	var decoded struct {
		Proposals []dataentity.SemanticMappingProposal `json:"proposals"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &decoded); err != nil {
		return nil, fmt.Errorf("parse semantic mapping output: %w", err)
	}
	return &datasvc.SuggestSemanticMappingsResponse{ModelRef: "coze-eino-builtin", Proposals: decoded.Proposals}, nil
}

type dataRequirementModelOutput struct {
	Requirements []struct {
		RequirementKind      string   `json:"requirement_kind"`
		SemanticName         string   `json:"semantic_name"`
		Description          string   `json:"description"`
		BusinessElementRefs  []string `json:"business_element_refs"`
		Requiredness         string   `json:"requiredness"`
		FreshnessRequirement string   `json:"freshness_requirement"`
		AccessNeed           string   `json:"access_need"`
	} `json:"requirements"`
}

func parseDataRequirementProposals(raw string) ([]dataentity.DataRequirementProposal, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	var decoded dataRequirementModelOutput
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &decoded); err != nil {
		return nil, fmt.Errorf("parse data requirement output: %w", err)
	}
	out := make([]dataentity.DataRequirementProposal, 0, len(decoded.Requirements))
	for _, item := range decoded.Requirements {
		out = append(out, dataentity.DataRequirementProposal{
			RequirementKind:      dataentity.RequirementKind(item.RequirementKind),
			SemanticName:         item.SemanticName,
			Description:          item.Description,
			BusinessElementRefs:  item.BusinessElementRefs,
			Requiredness:         item.Requiredness,
			FreshnessRequirement: item.FreshnessRequirement,
			AccessNeed:           item.AccessNeed,
		})
	}
	return out, nil
}

const dataRequirementSystemPrompt = `Analyze the supplied business semantic model and propose domain-agnostic data requirements.
Return JSON only:
{"requirements":[{"requirement_kind":"ENTITY|ATTRIBUTE|RELATION|EVENT|METRIC|STATE|TIME_SERIES|DOCUMENT|LOOKUP|HISTORY","semantic_name":"","description":"","business_element_refs":["existing semantic model element id"],"requiredness":"","freshness_requirement":"","access_need":""}]}
Every requirement must reference at least one existing semantic model node, edge, rule, or state ID.
Do not include data sources, credentials, mappings, contracts, physical schemas, status, or source fields.`

const semanticMappingSystemPrompt = `Propose semantic-to-physical mappings from confirmed requirements to the supplied normalized physical schemas.
Return JSON only: {"proposals":[{"requirement_id":"","source_id":"","connection_id":"","asset_id":"","schema_snapshot_id":"","target_field_paths":[""],"mapping_type":"DIRECT|CAST|ENUM_MAP|UNIT_CONVERT|TIME_NORMALIZE|FIELD_PATH|JOIN_REF","transform_spec":{"type":"DIRECT"},"confidence":0.0,"reason":""}]}.
Use only IDs and field paths present in the input. transform_spec must be declarative JSON and its type must equal mapping_type. For JOIN_REF, reference one supplied physical relationship exactly. Never emit credentials, connection configuration, secrets, status, source, SQL, scripts, or executable code.`
