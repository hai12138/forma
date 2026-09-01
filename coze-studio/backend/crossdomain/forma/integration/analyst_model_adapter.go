/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package integration

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"

	"github.com/coze-dev/coze-studio/backend/bizpkg/llm/modelbuilder"
	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
	analystsvc "github.com/coze-dev/coze-studio/backend/domain/forma/analyst/service"
)

// CozeEinoAnalystModel implements FormaAnalystModel via Coze Model Manager / Eino.
type CozeEinoAnalystModel struct {
	EnvPrefix string
}

func NewCozeEinoAnalystModel(envPrefix string) analystsvc.FormaAnalystModel {
	return &CozeEinoAnalystModel{EnvPrefix: envPrefix}
}

func (m *CozeEinoAnalystModel) GenerateInterviewTurn(ctx context.Context, req *analystsvc.InterviewTurnRequest) (*analystsvc.InterviewTurnResponse, error) {
	prompt := integration.AnalystSystemPolicy()
	if req != nil {
		if req.NextQuestion != nil && req.NextQuestion.Question != "" {
			return &analystsvc.InterviewTurnResponse{
				ModelRef: "coze-eino-builtin",
				Content:  req.NextQuestion.Question,
			}, nil
		}
		prompt = req.SystemPolicy
	}
	content, modelRef, err := m.invoke(ctx, prompt, req.UserMessage)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", entity.ErrModelFailed, err)
	}
	return &analystsvc.InterviewTurnResponse{
		ModelRef: modelRef,
		Content:  content,
	}, nil
}

func (m *CozeEinoAnalystModel) ExtractAssertions(ctx context.Context, req *analystsvc.ExtractionRequest) (*entity.ExtractionResult, error) {
	system := extractionSystemPrompt()
	user := fmt.Sprintf("User turn (%s): %s\n\nReturn JSON only.", req.UserTurnID, req.UserTurnContent)
	raw, modelRef, err := m.invoke(ctx, system, user)
	if err != nil {
		// Fallback to deterministic heuristic via fake model path
		fake := analystsvc.NewDeterministicFakeModel()
		return fake.ExtractAssertions(ctx, req)
	}
	res, parseErr := analystsvc.ParseStructuredExtraction(raw, req.UserTurnID)
	if parseErr != nil {
		fake := analystsvc.NewDeterministicFakeModel()
		return fake.ExtractAssertions(ctx, req)
	}
	_ = modelRef
	return res, nil
}

func (m *CozeEinoAnalystModel) ProposeModelPatch(ctx context.Context, req *analystsvc.ProposalRequest) (*entity.SemanticModelPatch, error) {
	if req == nil || len(req.Assertions) == 0 {
		return &entity.SemanticModelPatch{}, nil
	}
	return analystsvc.BuildProposalPatch(req.Assertions), nil
}

func (m *CozeEinoAnalystModel) invoke(ctx context.Context, system, user string) (string, string, error) {
	model, ok, err := modelbuilder.GetBuiltinChatModel(ctx, m.EnvPrefix)
	if err != nil || !ok || model == nil {
		return "", "", fmt.Errorf("builtin chat model not configured")
	}
	msgs := []*schema.Message{
		schema.SystemMessage(system),
		schema.UserMessage(user),
	}
	out, err := model.Generate(ctx, msgs)
	if err != nil {
		return "", "", err
	}
	content := strings.TrimSpace(out.Content)
	return content, "coze-eino-builtin", nil
}

func extractionSystemPrompt() string {
	return `Extract business facts as JSON:
{
  "assertions": [{"assertion_type":"ACTOR_EXISTS|BUSINESS_OBJECT_EXISTS|PROCESS_EXISTS|EVENT_EXISTS|SYSTEM_EXISTS|POLICY_EXISTS|RELATION_EXISTS|STATE_EXISTS|STATE_TRANSITION|BUSINESS_RULE|PROPERTY","subject_ref":"","predicate":"","object_value":"","confidence":0.0,"evidence_turn_ids":[]}],
  "evidence_links": [{"assertion_index":0,"evidence_turn_id":""}],
  "gaps": [{"gap_type":"INFORMATION","question":""}],
  "conflicts": []
}
Never auto-confirm. JSON only.`
}

// AnalystSystemPolicy re-export for integration package.
func AnalystSystemPolicy() string {
	return analystsvc.AnalystSystemPolicyExport()
}
