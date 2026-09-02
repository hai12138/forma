/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"

	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
)

// FormaDataModel is the ACL boundary for S4 LLM operations.
// Domain must not call provider SDKs directly.
type FormaDataModel interface {
	AnalyzeDataRequirements(ctx context.Context, req *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error)
	SuggestSemanticMappings(ctx context.Context, req *SuggestSemanticMappingsRequest) (*SuggestSemanticMappingsResponse, error)
}

type AnalyzeDataRequirementsRequest struct {
	RequestID             string
	TenantID              string
	BusinessID            string
	BusinessModelRevision int32
	SemanticModel         *businessentity.SemanticModel
}

type AnalyzeDataRequirementsResponse struct {
	ModelRef   string
	Proposals  []entity.DataRequirementProposal
	InputTokens  int32
	OutputTokens int32
}

type SuggestSemanticMappingsRequest struct {
	RequestID  string
	TenantID   string
	BusinessID string
}

type SuggestSemanticMappingsResponse struct {
	ModelRef string
	// G1: interface reserved; domain does not persist mappings.
}

// FakeFormaDataModel is a deterministic test double (zero provider calls).
type FakeFormaDataModel struct {
	AnalyzeFn func(ctx context.Context, req *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error)
	Calls     int
}

func (f *FakeFormaDataModel) AnalyzeDataRequirements(ctx context.Context, req *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
	f.Calls++
	if f.AnalyzeFn != nil {
		return f.AnalyzeFn(ctx, req)
	}
	return &AnalyzeDataRequirementsResponse{ModelRef: "fake-data-model", Proposals: nil}, nil
}

func (f *FakeFormaDataModel) SuggestSemanticMappings(_ context.Context, _ *SuggestSemanticMappingsRequest) (*SuggestSemanticMappingsResponse, error) {
	return &SuggestSemanticMappingsResponse{ModelRef: "fake-data-model"}, nil
}
