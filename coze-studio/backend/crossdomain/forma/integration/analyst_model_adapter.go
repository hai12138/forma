/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package integration

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"

	"github.com/coze-dev/coze-studio/backend/bizpkg/llm/modelbuilder"
	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
	analystsvc "github.com/coze-dev/coze-studio/backend/domain/forma/analyst/service"
)

// CozeEinoAnalystModel implements FormaAnalystModel via Coze Model Manager / Eino.
// Production: no heuristic fallback on failure.
type CozeEinoAnalystModel struct {
	EnvPrefix string
}

func NewCozeEinoAnalystModel(envPrefix string) analystsvc.FormaAnalystModel {
	return &CozeEinoAnalystModel{EnvPrefix: envPrefix}
}

func (m *CozeEinoAnalystModel) GenerateInterviewTurn(ctx context.Context, req *analystsvc.InterviewTurnRequest) (*analystsvc.InterviewTurnResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: request required", entity.ErrModelFailed)
	}
	system := req.SystemPolicy
	if system == "" {
		system = AnalystSystemPolicy()
	}
	userMsg := buildAnalystUserMessage(req)
	start := time.Now()
	content, modelRef, inTok, outTok, err := m.invoke(ctx, system, userMsg)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", entity.ErrModelFailed, err)
	}
	_ = start
	return &analystsvc.InterviewTurnResponse{
		ModelRef:     modelRef,
		Content:      content,
		InputTokens:  inTok,
		OutputTokens: outTok,
	}, nil
}

func buildAnalystUserMessage(req *analystsvc.InterviewTurnRequest) string {
	parts := []string{fmt.Sprintf("User message: %s", req.UserMessage)}
	if req.NextQuestion != nil {
		if req.NextQuestion.Goal != "" {
			parts = append(parts, fmt.Sprintf("Interview goal: %s", req.NextQuestion.Goal))
		}
		if req.NextQuestion.Question != "" {
			parts = append(parts, fmt.Sprintf("Suggested follow-up focus: %s", req.NextQuestion.Question))
		}
	}
	parts = append(parts, "Respond as the Forma AI Business Analyst. Ask clarifying business questions. Do not auto-confirm facts.")
	return strings.Join(parts, "\n")
}

func (m *CozeEinoAnalystModel) ExtractAssertions(ctx context.Context, req *analystsvc.ExtractionRequest) (*entity.ExtractionResult, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: request required", entity.ErrInvalidExtraction)
	}
	system := extractionSystemPrompt()
	user := fmt.Sprintf("User turn (%s): %s\n\nReturn JSON only.", req.UserTurnID, req.UserTurnContent)
	raw, _, _, _, err := m.invoke(ctx, system, user)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", entity.ErrModelFailed, err)
	}
	res, parseErr := analystsvc.ParseStructuredExtraction(raw, req.UserTurnID)
	if parseErr != nil {
		return nil, fmt.Errorf("%w: %v", entity.ErrInvalidExtraction, parseErr)
	}
	return res, nil
}

func (m *CozeEinoAnalystModel) ProposeModelPatch(ctx context.Context, req *analystsvc.ProposalRequest) (*entity.SemanticModelPatch, error) {
	if req == nil || len(req.Assertions) == 0 {
		return &entity.SemanticModelPatch{}, nil
	}
	return analystsvc.BuildProposalPatch(req.Assertions), nil
}

func (m *CozeEinoAnalystModel) invoke(ctx context.Context, system, user string) (content, modelRef string, inputTokens, outputTokens int32, err error) {
	model, ok, err := modelbuilder.GetBuiltinChatModel(ctx, m.EnvPrefix)
	if err != nil || !ok || model == nil {
		return "", "", 0, 0, fmt.Errorf("builtin chat model not configured")
	}
	msgs := []*schema.Message{
		schema.SystemMessage(system),
		schema.UserMessage(user),
	}
	out, err := model.Generate(ctx, msgs)
	if err != nil {
		return "", "", 0, 0, err
	}
	return strings.TrimSpace(out.Content), "coze-eino-builtin", 0, 0, nil
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

// AnalystSystemPolicy exposes system policy for ACL layer.
func AnalystSystemPolicy() string {
	return analystsvc.AnalystSystemPolicyExport()
}
