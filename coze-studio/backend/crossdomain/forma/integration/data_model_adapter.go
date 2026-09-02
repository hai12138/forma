/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package integration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"

	"github.com/coze-dev/coze-studio/backend/bizpkg/llm/modelbuilder"
	dataentity "github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	datasvc "github.com/coze-dev/coze-studio/backend/domain/forma/data/service"
)

var ErrNotImplemented = errors.New("not implemented")

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

func (m *CozeEinoDataModel) SuggestSemanticMappings(context.Context, *datasvc.SuggestSemanticMappingsRequest) (*datasvc.SuggestSemanticMappingsResponse, error) {
	return nil, fmt.Errorf("suggest semantic mappings: %w", ErrNotImplemented)
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
