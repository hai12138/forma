/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
)

// unavailableAnalystModel fails closed — production must inject a real model or this.
type unavailableAnalystModel struct{}

func NewUnavailableAnalystModel() FormaAnalystModel {
	return unavailableAnalystModel{}
}

func (unavailableAnalystModel) GenerateInterviewTurn(_ context.Context, _ *InterviewTurnRequest) (*InterviewTurnResponse, error) {
	return nil, entity.ErrModelFailed
}

func (unavailableAnalystModel) ExtractAssertions(_ context.Context, _ *ExtractionRequest) (*entity.ExtractionResult, error) {
	return nil, entity.ErrModelFailed
}

func (unavailableAnalystModel) ProposeModelPatch(_ context.Context, req *ProposalRequest) (*entity.SemanticModelPatch, error) {
	if req == nil || len(req.Assertions) == 0 {
		return &entity.SemanticModelPatch{}, nil
	}
	return BuildProposalPatch(req.Assertions), nil
}
