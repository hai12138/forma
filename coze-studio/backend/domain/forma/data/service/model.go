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
	ModelRef     string
	Proposals    []entity.DataRequirementProposal
	InputTokens  int32
	OutputTokens int32
}

// SuggestSemanticMappingsRequest must never contain Credential, Secret, or PublicConfig data.
type SuggestSemanticMappingsRequest struct {
	RequestID             string
	TenantID              string
	BusinessID            string
	BusinessModelRevision int32
	SemanticModel         *businessentity.SemanticModel
	RequirementIDs        []string
	Requirements          []MappingRequirementMetadata
	SchemaSnapshots       []NormalizedSchemaSnapshot
}

type MappingRequirementMetadata struct {
	RequirementID   string                 `json:"requirement_id"`
	RequirementKind entity.RequirementKind `json:"requirement_kind"`
	SemanticName    string                 `json:"semantic_name"`
	Description     string                 `json:"description"`
	Requiredness    string                 `json:"requiredness"`
	AccessNeed      string                 `json:"access_need"`
}

type NormalizedSchemaSnapshot struct {
	SchemaSnapshotID string                `json:"schema_snapshot_id"`
	SourceID         string                `json:"source_id"`
	ConnectionID     string                `json:"connection_id"`
	AssetID          string                `json:"asset_id"`
	Schema           entity.PhysicalSchema `json:"schema"`
}

type SuggestSemanticMappingsResponse struct {
	ModelRef  string
	Proposals []entity.SemanticMappingProposal
}

// FakeFormaDataModel is a deterministic test double (zero provider calls).
type FakeFormaDataModel struct {
	AnalyzeFn      func(ctx context.Context, req *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error)
	Calls          int
	SuggestFn      func(ctx context.Context, req *SuggestSemanticMappingsRequest) (*SuggestSemanticMappingsResponse, error)
	SuggestCalls   int
	LastSuggestReq *SuggestSemanticMappingsRequest
}

func (f *FakeFormaDataModel) AnalyzeDataRequirements(ctx context.Context, req *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
	f.Calls++
	if f.AnalyzeFn != nil {
		return f.AnalyzeFn(ctx, req)
	}
	return &AnalyzeDataRequirementsResponse{ModelRef: "fake-data-model", Proposals: nil}, nil
}

func (f *FakeFormaDataModel) SuggestSemanticMappings(ctx context.Context, req *SuggestSemanticMappingsRequest) (*SuggestSemanticMappingsResponse, error) {
	f.SuggestCalls++
	f.LastSuggestReq = req
	if f.SuggestFn != nil {
		return f.SuggestFn(ctx, req)
	}
	return &SuggestSemanticMappingsResponse{ModelRef: "fake-data-model"}, nil
}
