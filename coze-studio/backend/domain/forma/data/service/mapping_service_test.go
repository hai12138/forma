/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/data/repository"
)

func mappingFixture(t *testing.T) (MappingService, repository.MappingRepository, *FakeFormaDataModel, string) {
	t.Helper()
	ctx := context.Background()
	data := repository.NewMemoryDataRepository()
	sources := repository.NewMemoryDataSourceRepository()
	mappings := repository.NewMemoryMappingRepository()
	now := time.Now().UTC()
	req := &entity.DataRequirement{RequirementID: "req_temperature", TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7, RequirementKind: entity.KindAttribute, SemanticName: "temperature", Status: entity.StatusConfirmed, CreatedAt: now, UpdatedAt: now}
	if err := data.CreateRequirement(ctx, req); err != nil {
		t.Fatal(err)
	}
	if err := sources.CreateSource(ctx, &entity.DataSource{SourceID: "source", TenantID: "tenant"}); err != nil {
		t.Fatal(err)
	}
	if err := sources.CreateConnection(ctx, &entity.DataConnection{ConnectionID: "connection", SourceID: "source", TenantID: "tenant"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := sources.UpsertAsset(ctx, &entity.DataAsset{AssetID: "asset", TenantID: "tenant", SourceID: "source", ConnectionID: "connection", LocatorDigest: "asset"}); err != nil {
		t.Fatal(err)
	}
	schema, _ := json.Marshal(entity.PhysicalSchema{Name: "readings", Fields: []entity.PhysicalField{{Name: "temp_c", Path: "temp_c", DataType: "number"}}, Relationships: []entity.PhysicalRelationship{{Name: "sensor", FromFields: []string{"sensor_id"}, ToSchema: "sensors", ToFields: []string{"id"}}}})
	if err := sources.CreateSnapshot(ctx, &entity.SchemaSnapshot{SnapshotID: "snapshot", TenantID: "tenant", SourceID: "source", ConnectionID: "connection", AssetID: "asset", SchemaJSON: string(schema)}); err != nil {
		t.Fatal(err)
	}
	model := &FakeFormaDataModel{SuggestFn: func(context.Context, *SuggestSemanticMappingsRequest) (*SuggestSemanticMappingsResponse, error) {
		return &SuggestSemanticMappingsResponse{ModelRef: "fake", Proposals: []entity.SemanticMappingProposal{{
			RequirementID: "req_temperature", SourceID: "source", ConnectionID: "connection", AssetID: "asset", SchemaSnapshotID: "snapshot",
			TargetFieldPaths: []string{"temp_c"}, MappingType: entity.MappingTypeDirect, TransformSpec: json.RawMessage(`{"type":"DIRECT"}`), Confidence: .93,
		}}}, nil
	}}
	business := newStubBusiness()
	business.currentRevision = 7
	return NewMappingService(&MappingComponents{MappingRepo: mappings, DataRepo: data, DataSourceRepo: sources, BusinessSVC: business, Model: model}), mappings, model, req.RequirementID
}

func TestAnalyzeSemanticMappingsPersistsValidBatchIdempotently(t *testing.T) {
	svc, _, model, reqID := mappingFixture(t)
	in := &AnalyzeSemanticMappingsInput{TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7, RequirementIDs: []string{reqID}, SchemaSnapshotIDs: []string{"snapshot"}, ClientRequestID: "request-1", ActorID: "actor"}
	first, err := svc.AnalyzeSemanticMappings(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.AnalyzeSemanticMappings(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if first.Run.AnalysisRunID != second.Run.AnalysisRunID || model.SuggestCalls != 1 || len(first.Mappings) != 1 {
		t.Fatalf("idempotency failed: first=%+v second=%+v calls=%d", first, second, model.SuggestCalls)
	}
}

func TestAnalyzeSemanticMappingsLoadsPinnedSemanticModel(t *testing.T) {
	svc, _, model, reqID := mappingFixture(t)
	_, err := svc.AnalyzeSemanticMappings(context.Background(), &AnalyzeSemanticMappingsInput{
		TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7,
		RequirementIDs: []string{reqID}, SchemaSnapshotIDs: []string{"snapshot"},
		ClientRequestID: "semantic-context", ActorID: "actor",
	})
	if err != nil {
		t.Fatal(err)
	}
	if model.LastSuggestReq == nil || model.LastSuggestReq.SemanticModel == nil || len(model.LastSuggestReq.SemanticModel.Nodes) == 0 {
		t.Fatalf("semantic model was not supplied: %+v", model.LastSuggestReq)
	}
	if model.LastSuggestReq.BusinessModelRevision != 7 {
		t.Fatalf("business revision changed: %d", model.LastSuggestReq.BusinessModelRevision)
	}
}

func TestAnalyzeSemanticMappingsRejectsConfidenceOutOfRange(t *testing.T) {
	svc, repo, model, reqID := mappingFixture(t)
	model.SuggestFn = func(context.Context, *SuggestSemanticMappingsRequest) (*SuggestSemanticMappingsResponse, error) {
		return &SuggestSemanticMappingsResponse{Proposals: []entity.SemanticMappingProposal{{
			RequirementID: reqID, SourceID: "source", ConnectionID: "connection", AssetID: "asset", SchemaSnapshotID: "snapshot",
			TargetFieldPaths: []string{"temp_c"}, MappingType: entity.MappingTypeDirect,
			TransformSpec: json.RawMessage(`{"type":"DIRECT"}`), Confidence: 1.01,
		}}}, nil
	}
	result, err := svc.AnalyzeSemanticMappings(context.Background(), &AnalyzeSemanticMappingsInput{
		TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7,
		RequirementIDs: []string{reqID}, SchemaSnapshotIDs: []string{"snapshot"},
		ClientRequestID: "bad-confidence", ActorID: "actor",
	})
	if !errors.Is(err, entity.ErrMappingTransformInvalid) || result.Run.Status != entity.AnalysisFailed {
		t.Fatalf("expected failed confidence validation, got %+v %v", result, err)
	}
	got, _ := repo.ListMappings(context.Background(), "tenant", "lab", 7, "")
	if len(got) != 0 {
		t.Fatalf("invalid proposal persisted: %d", len(got))
	}
}

func TestAnalyzeSemanticMappingsIdempotencyConflict(t *testing.T) {
	svc, _, _, reqID := mappingFixture(t)
	base := &AnalyzeSemanticMappingsInput{TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7, RequirementIDs: []string{reqID}, SchemaSnapshotIDs: []string{"snapshot"}, ClientRequestID: "same", ActorID: "actor"}
	if _, err := svc.AnalyzeSemanticMappings(context.Background(), base); err != nil {
		t.Fatal(err)
	}
	changed := *base
	changed.SchemaSnapshotIDs = []string{"different"}
	if _, err := svc.AnalyzeSemanticMappings(context.Background(), &changed); !errors.Is(err, entity.ErrMappingAnalysisIdempotencyConflict) {
		t.Fatalf("expected idempotency conflict, got %v", err)
	}
}

func TestAnalyzeSemanticMappingsConcurrentClaimHasSingleOwner(t *testing.T) {
	svc, _, model, reqID := mappingFixture(t)
	in := &AnalyzeSemanticMappingsInput{TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7, RequirementIDs: []string{reqID}, SchemaSnapshotIDs: []string{"snapshot"}, ClientRequestID: "concurrent", ActorID: "actor"}
	start := make(chan struct{})
	done := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() { <-start; _, err := svc.AnalyzeSemanticMappings(context.Background(), in); done <- err }()
	}
	close(start)
	for i := 0; i < 2; i++ {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
	if model.SuggestCalls != 1 {
		t.Fatalf("expected one model execution, got %d", model.SuggestCalls)
	}
}

func TestAnalyzeSemanticMappingsRejectsInvalidBatchAtomically(t *testing.T) {
	svc, repo, model, reqID := mappingFixture(t)
	model.SuggestFn = func(context.Context, *SuggestSemanticMappingsRequest) (*SuggestSemanticMappingsResponse, error) {
		return &SuggestSemanticMappingsResponse{Proposals: []entity.SemanticMappingProposal{
			{RequirementID: reqID, SourceID: "source", ConnectionID: "connection", AssetID: "asset", SchemaSnapshotID: "snapshot", TargetFieldPaths: []string{"temp_c"}, MappingType: entity.MappingTypeDirect, TransformSpec: json.RawMessage(`{"type":"DIRECT"}`)},
			{RequirementID: reqID, SourceID: "source", ConnectionID: "connection", AssetID: "asset", SchemaSnapshotID: "snapshot", TargetFieldPaths: []string{"missing"}, MappingType: entity.MappingTypeDirect, TransformSpec: json.RawMessage(`{"type":"DIRECT"}`)},
		}}, nil
	}
	result, err := svc.AnalyzeSemanticMappings(context.Background(), &AnalyzeSemanticMappingsInput{TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7, RequirementIDs: []string{reqID}, SchemaSnapshotIDs: []string{"snapshot"}, ClientRequestID: "bad", ActorID: "actor"})
	if !errors.Is(err, entity.ErrMappingTargetInvalid) || result.Run.Status != entity.AnalysisFailed {
		t.Fatalf("expected failed target validation, got %+v %v", result, err)
	}
	got, _ := repo.ListMappings(context.Background(), "tenant", "lab", 7, "")
	if len(got) != 0 {
		t.Fatalf("partial batch persisted: %d", len(got))
	}
}

func TestMappingDecisionLifecycleAndCoverage(t *testing.T) {
	svc, _, _, reqID := mappingFixture(t)
	manual, err := svc.CreateManualMapping(context.Background(), &ManualMappingInput{TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7, RequirementID: reqID, SourceID: "source", ConnectionID: "connection", AssetID: "asset", SchemaSnapshotID: "snapshot", TargetFieldPaths: []string{"temp_c"}, MappingType: entity.MappingTypeDirect, TransformSpec: json.RawMessage(`{"type":"DIRECT"}`), ActorID: "actor"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = svc.ConfirmMapping(context.Background(), "tenant", manual.MappingID, "actor", "verified"); err != nil {
		t.Fatal(err)
	}
	if _, _, err = svc.ConfirmMapping(context.Background(), "tenant", manual.MappingID, "actor", "again"); !errors.Is(err, entity.ErrMappingInvalidState) {
		t.Fatalf("expected invalid state, got %v", err)
	}
	coverage, err := svc.GetMappingCoverage(context.Background(), "tenant", "lab", 7)
	if err != nil || coverage.TotalConfirmedRequirements != 1 || coverage.ConfirmedMappings != 1 || coverage.Coverage != 1 {
		t.Fatalf("bad coverage: %+v %v", coverage, err)
	}
}

func TestDuplicateConfirmedMappingAndEditLineage(t *testing.T) {
	svc, _, _, reqID := mappingFixture(t)
	create := func() *entity.SemanticMapping {
		m, err := svc.CreateManualMapping(context.Background(), &ManualMappingInput{TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7, RequirementID: reqID, SourceID: "source", ConnectionID: "connection", AssetID: "asset", SchemaSnapshotID: "snapshot", TargetFieldPaths: []string{"temp_c"}, MappingType: entity.MappingTypeDirect, TransformSpec: json.RawMessage(`{"type":"DIRECT"}`), ActorID: "actor"})
		if err != nil {
			t.Fatal(err)
		}
		return m
	}
	first := create()
	if _, _, err := svc.ConfirmMapping(context.Background(), "tenant", first.MappingID, "actor", "first"); err != nil {
		t.Fatal(err)
	}
	second := create()
	if _, _, err := svc.ConfirmMapping(context.Background(), "tenant", second.MappingID, "actor", "second"); !errors.Is(err, entity.ErrMappingAlreadyConfirmed) {
		t.Fatalf("expected duplicate confirm rejection, got %v", err)
	}

	svc2, _, _, reqID2 := mappingFixture(t)
	editable, err := svc2.CreateManualMapping(context.Background(), &ManualMappingInput{TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7, RequirementID: reqID2, SourceID: "source", ConnectionID: "connection", AssetID: "asset", SchemaSnapshotID: "snapshot", TargetFieldPaths: []string{"temp_c"}, MappingType: entity.MappingTypeDirect, TransformSpec: json.RawMessage(`{"type":"DIRECT"}`), ActorID: "actor"})
	if err != nil {
		t.Fatal(err)
	}
	old, replacement, decision, err := svc2.EditConfirmMapping(context.Background(), &EditConfirmMappingInput{TenantID: "tenant", SourceMappingID: editable.MappingID, Reason: "curated", ActorID: "reviewer"})
	if err != nil {
		t.Fatal(err)
	}
	if old.Status != entity.MappingStatusSuperseded || replacement.Source != entity.MappingSourceManualModified || replacement.DerivedFromMappingID != old.MappingID || decision.TargetMappingID != replacement.MappingID {
		t.Fatalf("invalid edit lineage: old=%+v replacement=%+v decision=%+v", old, replacement, decision)
	}
}

func TestFailedAnalysisCanRetryWithoutRealModel(t *testing.T) {
	svc, _, model, reqID := mappingFixture(t)
	model.SuggestFn = func(context.Context, *SuggestSemanticMappingsRequest) (*SuggestSemanticMappingsResponse, error) {
		return nil, errors.New("provider secret must be sanitized")
	}
	result, err := svc.AnalyzeSemanticMappings(context.Background(), &AnalyzeSemanticMappingsInput{TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7, RequirementIDs: []string{reqID}, SchemaSnapshotIDs: []string{"snapshot"}, ClientRequestID: "retry", ActorID: "actor"})
	if !errors.Is(err, entity.ErrModelFailed) || result.Run.Status != entity.AnalysisFailed || result.Run.ErrorMessageSanitized != "sanitized error" {
		t.Fatalf("unexpected failure: %+v %v", result, err)
	}
	model.SuggestFn = func(context.Context, *SuggestSemanticMappingsRequest) (*SuggestSemanticMappingsResponse, error) {
		return &SuggestSemanticMappingsResponse{ModelRef: "fake", Proposals: nil}, nil
	}
	retried, err := svc.RetryFailedMappingAnalysis(context.Background(), "tenant", result.Run.AnalysisRunID, "reviewer")
	if err != nil || retried.Run.Status != entity.AnalysisSucceeded || model.SuggestCalls != 2 {
		t.Fatalf("retry failed: %+v %v calls=%d", retried, err, model.SuggestCalls)
	}
}

func TestOnlyConfirmedRequirementsCanBeMapped(t *testing.T) {
	svc, _, _, reqID := mappingFixture(t)
	data := repository.NewMemoryDataRepository()
	_ = reqID
	_ = data
	_, err := svc.CreateManualMapping(context.Background(), &ManualMappingInput{TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7, RequirementID: "missing"})
	if !errors.Is(err, entity.ErrMappingRequirementNotConfirmed) {
		t.Fatalf("expected requirement gate, got %v", err)
	}
}

func TestValidateTransformAndJoinRef(t *testing.T) {
	schema := &entity.PhysicalSchema{Name: "orders", Fields: []entity.PhysicalField{{Name: "customer_id", Path: "customer_id"}}, Relationships: []entity.PhysicalRelationship{{Name: "customer", FromFields: []string{"customer_id"}, ToSchema: "customers", ToFields: []string{"id"}}}}
	if err := ValidateMappingTarget([]string{"customer_id"}, schema); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMappingTarget([]string{"unknown"}, schema); !errors.Is(err, entity.ErrMappingTargetInvalid) {
		t.Fatalf("expected target error, got %v", err)
	}
	valid := json.RawMessage(`{"type":"JOIN_REF","relationship":"customer","from_fields":["customer_id"],"to_schema":"customers","to_fields":["id"]}`)
	if err := ValidateTransformSpec(entity.MappingTypeJoinRef, valid, []string{"customer_id"}, schema); err != nil {
		t.Fatal(err)
	}
	if err := ValidateJoinRef(valid, schema); err != nil {
		t.Fatal(err)
	}
	if err := ValidateTransformSpec(entity.MappingTypeCast, json.RawMessage(`{"type":"CAST"}`), []string{"customer_id"}, schema); !errors.Is(err, entity.ErrMappingTransformInvalid) {
		t.Fatalf("expected transform error, got %v", err)
	}
}
