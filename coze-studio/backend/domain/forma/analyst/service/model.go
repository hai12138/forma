/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
)

type ExtractionOutcome struct {
	Result       *entity.ExtractionResult
	ModelRef     string
	InputTokens  int32
	OutputTokens int32
}

// FormaAnalystModel is the ACL boundary for LLM operations.
// Domain must not call provider SDKs directly.
type FormaAnalystModel interface {
	GenerateInterviewTurn(ctx context.Context, req *InterviewTurnRequest) (*InterviewTurnResponse, error)
	ExtractAssertions(ctx context.Context, req *ExtractionRequest) (*ExtractionOutcome, error)
	ProposeModelPatch(ctx context.Context, req *ProposalRequest) (*entity.SemanticModelPatch, error)
}

type InterviewTurnRequest struct {
	RequestID       string
	TenantID        string
	BusinessID      string
	SessionID       string
	ContextManifest *entity.ContextManifest
	ContextText     string
	SystemPolicy    string
	UserMessage     string
	NextQuestion    *entity.NextQuestionPlan
}

type InterviewTurnResponse struct {
	ModelRef     string
	Content      string
	InputTokens  int32
	OutputTokens int32
}

type ExtractionRequest struct {
	RequestID       string
	TenantID        string
	BusinessID      string
	SessionID       string
	ContextManifest *entity.ContextManifest
	ContextText     string
	SystemPolicy    string
	UserTurnContent string
	UserTurnID      string
}

type ProposalRequest struct {
	RequestID    string
	TenantID     string
	BusinessID   string
	SessionID    string
	Assertions   []*entity.BusinessAssertion
	BaseRevision int32
}
