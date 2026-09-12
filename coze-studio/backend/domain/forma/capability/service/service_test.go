/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	assetentity "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/fixture"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/repository"
)

func newTestService(gen ProposalGenerator) (CapabilityService, repository.CapabilityRepository, *MemoryAssetProjection) {
	repo := repository.NewMemoryCapabilityRepository()
	assets := NewMemoryAssetProjection()
	svc := NewCapabilityService(&Components{Repo: repo, Assets: assets, Generator: gen})
	return svc, repo, assets
}

func TestManualCreateLifecycle(t *testing.T) {
	svc, _, assets := newTestService(nil)
	cap, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: "actor",
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	require.Equal(t, entity.RevisionDraft, rev.Status)
	require.Equal(t, entity.SourceManualCreated, rev.Source)
	require.Equal(t, int32(1), rev.Version)

	asset, err := assets.GetCapabilityAsset(context.Background(), "t1", cap.CapabilityID)
	require.NoError(t, err)
	require.Equal(t, assetentity.AssetKindCapability, asset.Kind)
	require.Equal(t, int32(1), asset.Revision)
	require.Equal(t, "1.0", asset.SchemaVersion)
	require.Equal(t, assetentity.AssetStatusDraft, asset.Status)

	validated, err := svc.Validate(context.Background(), "t1", rev.RevisionID, "actor")
	require.NoError(t, err)
	require.Equal(t, entity.RevisionValidated, validated.Status)

	active, err := svc.Activate(context.Background(), "t1", rev.RevisionID, "owner", "")
	require.NoError(t, err)
	require.Equal(t, entity.RevisionActive, active.Status)
	asset, _ = assets.GetCapabilityAsset(context.Background(), "t1", cap.CapabilityID)
	require.Equal(t, assetentity.AssetStatusReleased, asset.Status)
	require.Equal(t, "0.1.0", asset.SemanticVersion)
}

func TestDeriveIdempotency(t *testing.T) {
	svc, _, _ := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: "a", CapabilityID: "cap-d",
		Payload: fixture.ProcurementApprovalCapability(),
	})
	require.NoError(t, err)
	payload := fixture.ProcurementApprovalCapability()
	payload.Name = "SubmitProcurementApproval-v2"
	r1, d1, err := svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: "cap-d", SourceRevisionID: rev.RevisionID,
		ClientRequestID: "derive-1", ActorID: "a", Payload: payload,
	})
	require.NoError(t, err)
	require.Equal(t, int32(2), r1.Version)
	require.Equal(t, entity.SourceDerivedEdit, r1.Source)

	r2, d2, err := svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: "cap-d", SourceRevisionID: rev.RevisionID,
		ClientRequestID: "derive-1", ActorID: "a", Payload: payload,
	})
	require.NoError(t, err)
	require.Equal(t, r1.RevisionID, r2.RevisionID)
	require.Equal(t, d1.DecisionID, d2.DecisionID)

	payload.Description = "changed"
	_, _, err = svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: "cap-d", SourceRevisionID: rev.RevisionID,
		ClientRequestID: "derive-1", ActorID: "a", Payload: payload,
	})
	require.ErrorIs(t, err, entity.ErrIdempotencyConflict)
}

func TestRejectNoShell(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, repo, assets := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "an1", ActorID: "a",
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	require.Len(t, res.Proposals, 1)
	dec, err := svc.RejectProposal(context.Background(), &RejectInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: "a",
	})
	require.NoError(t, err)
	require.Equal(t, entity.DecisionReject, dec.Action)
	require.Empty(t, dec.CapabilityID)
	_, err = repo.GetCapability(context.Background(), "t1", "anything")
	require.ErrorIs(t, err, entity.ErrNotFound)
	_, err = assets.GetCapabilityAsset(context.Background(), "t1", "nope")
	require.ErrorIs(t, err, entity.ErrNotFound)
}

func TestConfirmFirstCreateAndReplay(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, _, assets := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "an2", ActorID: "a",
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	propID := res.Proposals[0].ProposalID
	rev, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: propID, ActorID: "a", CapabilityID: "cap-ai",
	})
	require.NoError(t, err)
	require.Equal(t, entity.SourceAIProposal, rev.Source)
	asset, err := assets.GetCapabilityAsset(context.Background(), "t1", "cap-ai")
	require.NoError(t, err)
	require.Equal(t, int32(1), asset.Revision)

	rev2, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: propID, ActorID: "a", CapabilityID: "cap-ai",
	})
	require.NoError(t, err)
	require.Equal(t, rev.RevisionID, rev2.RevisionID)
}

func TestConfirmOntoExistingKeepsACTIVEProjection(t *testing.T) {
	svc, _, assets := newTestService(&DeterministicFakeGenerator{
		Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()},
	})
	cap, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: "a", CapabilityID: "cap-x",
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	_, err = svc.Validate(context.Background(), "t1", rev.RevisionID, "a")
	require.NoError(t, err)
	_, err = svc.Activate(context.Background(), "t1", rev.RevisionID, "a", "")
	require.NoError(t, err)
	before, _ := assets.GetCapabilityAsset(context.Background(), "t1", cap.CapabilityID)

	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "an3", ActorID: "a",
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	draft, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: "a", CapabilityID: cap.CapabilityID,
	})
	require.NoError(t, err)
	require.Equal(t, entity.RevisionDraft, draft.Status)
	require.Equal(t, int32(2), draft.Version)

	after, _ := assets.GetCapabilityAsset(context.Background(), "t1", cap.CapabilityID)
	require.Equal(t, before.Name, after.Name)
	require.Equal(t, before.SemanticVersion, after.SemanticVersion)
	require.Equal(t, before.ContentDigest, after.ContentDigest)
	require.Equal(t, assetentity.AssetStatusReleased, after.Status)
	require.Equal(t, int32(1), after.Revision)
}

func TestAnalysisIdempotencyAndRetry(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.ProcurementQueryCapability()}}
	svc, _, _ := newTestService(gen)
	in := &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 2, ClientRequestID: "same", ActorID: "a",
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 2, RequirementRefs: []string{"r1"}},
	}
	first, err := svc.StartAnalysis(context.Background(), in)
	require.NoError(t, err)
	require.True(t, first.OwnedExecute)
	require.Equal(t, 1, gen.CallCount())

	second, err := svc.StartAnalysis(context.Background(), in)
	require.NoError(t, err)
	require.False(t, second.OwnedExecute)
	require.Equal(t, first.Run.AnalysisRunID, second.Run.AnalysisRunID)
	require.Equal(t, 1, gen.CallCount())

	in.Analysis.RequirementRefs = []string{"r2"}
	_, err = svc.StartAnalysis(context.Background(), in)
	require.ErrorIs(t, err, entity.ErrIdempotencyConflict)

	failGen := &DeterministicFakeGenerator{Err: errors.New("boom")}
	svc2, _, _ := newTestService(failGen)
	failed, err := svc2.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 2, ClientRequestID: "fail", ActorID: "a",
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 2},
	})
	require.Error(t, err)
	require.Equal(t, entity.AnalysisFailed, failed.Run.Status)

	failGen.Err = nil
	failGen.Proposals = []entity.SemanticPayload{fixture.ProcurementQueryCapability()}
	retried, err := svc2.RetryFailedAnalysis(context.Background(), "t1", failed.Run.AnalysisRunID, "a")
	require.NoError(t, err)
	require.Equal(t, entity.AnalysisSucceeded, retried.Run.Status)
	require.Equal(t, failed.Run.AnalysisRunID, retried.Run.AnalysisRunID)
	require.GreaterOrEqual(t, retried.Run.Attempt, int32(2))
}

func TestConcurrentConfirmDistinctDrafts(t *testing.T) {
	payload := fixture.LaboratoryFlowCapability()
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{payload, payload}}
	svc, _, assets := newTestService(gen)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: "a", CapabilityID: "cap-c",
		Payload: payload,
	})
	require.NoError(t, err)
	_, _ = svc.Validate(context.Background(), "t1", rev.RevisionID, "a")
	_, _ = svc.Activate(context.Background(), "t1", rev.RevisionID, "a", "")

	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "multi", ActorID: "a",
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	require.Len(t, res.Proposals, 2)

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	revs := make([]*entity.BusinessCapabilityRevision, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			r, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
				TenantID: "t1", ProposalID: res.Proposals[idx].ProposalID, ActorID: "a", CapabilityID: "cap-c",
			})
			revs[idx] = r
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	require.NotEqual(t, revs[0].RevisionID, revs[1].RevisionID)
	require.NotEqual(t, revs[0].Version, revs[1].Version)
	asset, _ := assets.GetCapabilityAsset(context.Background(), "t1", "cap-c")
	require.Equal(t, assetentity.AssetStatusReleased, asset.Status)
	require.Equal(t, "0.1.0", asset.SemanticVersion)
}

func TestConcurrentFirstCreateRace(t *testing.T) {
	svc, repo, assets := newTestService(nil)
	payload := fixture.ProcurementApprovalCapability()

	var wg sync.WaitGroup
	results := make([]error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, _, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
				TenantID: "t1", BusinessID: "biz", ActorID: "a", CapabilityID: "same-id",
				Payload: payload,
			})
			results[idx] = err
		}(i)
	}
	wg.Wait()
	wins, loses := 0, 0
	for _, err := range results {
		if err == nil {
			wins++
		} else if errors.Is(err, entity.ErrConflict) {
			loses++
		} else {
			t.Fatalf("unexpected err: %v", err)
		}
	}
	require.Equal(t, 1, wins)
	require.Equal(t, 1, loses)
	_, err := repo.GetCapability(context.Background(), "t1", "same-id")
	require.NoError(t, err)
	_, err = assets.GetCapabilityAsset(context.Background(), "t1", "same-id")
	require.NoError(t, err)
}

func TestTenantIsolation(t *testing.T) {
	svc, _, _ := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "tenant-a", BusinessID: "biz", ActorID: "a",
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	_, err = svc.GetRevision(context.Background(), "tenant-b", rev.RevisionID)
	require.ErrorIs(t, err, entity.ErrRevisionNotFound)
}

func TestDualIndustryFixturesAgnostic(t *testing.T) {
	svc, _, _ := newTestService(nil)
	_, r1, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "lab", ActorID: "a", Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	_, r2, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "proc", ActorID: "a", Payload: fixture.ProcurementApprovalCapability(),
	})
	require.NoError(t, err)
	require.Equal(t, entity.KindQuery, r1.CapabilityKind)
	require.Equal(t, entity.KindCommand, r2.CapabilityKind)
}

func TestIllegalValidateWithoutProvenanceBlocked(t *testing.T) {
	// Direct DRAFT without CREATE decision cannot happen via ManualCreate; simulate via repo.
	repo := repository.NewMemoryCapabilityRepository()
	assets := NewMemoryAssetProjection()
	svc := NewCapabilityService(&Components{Repo: repo, Assets: assets})
	ctx := context.Background()
	_ = repo.CreateCapability(ctx, &entity.BusinessCapability{CapabilityID: "c", TenantID: "t", BusinessID: "b", CreatedBy: "a"})
	rev := &entity.BusinessCapabilityRevision{
		RevisionID: "r", CapabilityID: "c", TenantID: "t", BusinessID: "b", Version: 1,
		Status: entity.RevisionDraft, Name: "n", CapabilityKind: entity.KindCommand,
		BusinessModelRevision: 1, Source: entity.SourceManualCreated,
	}
	_ = repo.CreateRevision(ctx, rev)
	_, err := svc.Validate(ctx, "t", "r", "a")
	require.ErrorIs(t, err, entity.ErrMissingProvenance)
}

func TestEditConfirmRequiresFullPayloadAndReplay(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, _, _ := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "edit-an", ActorID: "a",
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	payload := fixture.LaboratoryFlowCapability()
	payload.Name = "EditedLabFlow"
	rev, err := svc.EditConfirmProposal(context.Background(), &EditConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: "a", CapabilityID: "cap-edit",
		EffectivePayload: payload,
	})
	require.NoError(t, err)
	require.Equal(t, "EditedLabFlow", rev.Name)
	require.Equal(t, entity.SourceAIProposal, rev.Source)

	rev2, err := svc.EditConfirmProposal(context.Background(), &EditConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: "a", CapabilityID: "cap-edit",
		EffectivePayload: payload,
	})
	require.NoError(t, err)
	require.Equal(t, rev.RevisionID, rev2.RevisionID)

	payload.Description = "different"
	_, err = svc.EditConfirmProposal(context.Background(), &EditConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: "a", CapabilityID: "cap-edit",
		EffectivePayload: payload,
	})
	require.ErrorIs(t, err, entity.ErrIdempotencyConflict)
}

func TestConcurrentAnalysisSingleGeneratorInvocation(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.ProcurementQueryCapability()}}
	svc, _, _ := newTestService(gen)
	var wg sync.WaitGroup
	results := make([]*AnalysisResult, 8)
	errs := make([]error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
				TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 3, ClientRequestID: "once", ActorID: "a",
				Analysis: entity.AnalysisRequest{BusinessModelRevision: 3},
			})
			results[idx] = res
			errs[idx] = err
		}(i)
	}
	wg.Wait()
	for _, err := range errs {
		require.NoError(t, err)
	}
	require.Equal(t, 1, gen.CallCount())
	id := results[0].Run.AnalysisRunID
	for _, r := range results {
		require.Equal(t, id, r.Run.AnalysisRunID)
		require.Equal(t, entity.AnalysisSucceeded, r.Run.Status)
	}
}

func TestStaleGenerationCompletionRejected(t *testing.T) {
	repo := repository.NewMemoryCapabilityRepository()
	run := &entity.CapabilityAnalysisRun{
		AnalysisRunID: "run-stale", TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1,
		ClientRequestID: "stale", RequestDigest: "d", Status: entity.AnalysisPending, Attempt: 2,
		CreatedBy: "a",
	}
	_, created, err := repo.CreateOrClaimAnalysisRun(context.Background(), run)
	require.NoError(t, err)
	require.True(t, created)
	err = repo.MarkAnalysisSucceeded(context.Background(), "t1", "run-stale", "fake", 1)
	require.ErrorIs(t, err, entity.ErrStaleGeneration)
}

func TestProjectionConsistencyRollsBackAsset(t *testing.T) {
	svc, repo, assets := newTestService(nil)
	cap, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: "a", CapabilityID: "cap-cons",
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	_, err = svc.Validate(context.Background(), "t1", rev.RevisionID, "a")
	require.NoError(t, err)
	_, err = svc.Activate(context.Background(), "t1", rev.RevisionID, "a", "")
	require.NoError(t, err)
	before, err := assets.GetCapabilityAsset(context.Background(), "t1", cap.CapabilityID)
	require.NoError(t, err)

	// Inject orphan ACTIVE inconsistency: clear pointer while ACTIVE row remains.
	err = repo.Transaction(context.Background(), func(tx repository.CapabilityRepository) error {
		return tx.UpdateActiveRevisionID(context.Background(), "t1", cap.CapabilityID, "")
	})
	require.NoError(t, err)

	payload := fixture.LaboratoryFlowCapability()
	payload.Name = "DerivedAfterCorrupt"
	_, _, err = svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: cap.CapabilityID, SourceRevisionID: rev.RevisionID,
		ClientRequestID: "corrupt-derive", ActorID: "a", Payload: payload,
	})
	require.ErrorIs(t, err, entity.ErrConsistency)

	after, err := assets.GetCapabilityAsset(context.Background(), "t1", cap.CapabilityID)
	require.NoError(t, err)
	require.Equal(t, before.Name, after.Name)
	require.Equal(t, before.ContentDigest, after.ContentDigest)
	require.Equal(t, before.Status, after.Status)

	// Corrupted aggregate must not leave a derived revision.
	revs, err := repo.ListRevisions(context.Background(), "t1", cap.CapabilityID)
	require.NoError(t, err)
	require.Len(t, revs, 1)
}

func TestNoSecretFieldsInDomainPayload(t *testing.T) {
	payload := fixture.LaboratoryFlowCapability()
	raw, err := json.Marshal(payload)
	require.NoError(t, err)
	lower := strings.ToLower(string(raw))
	for _, bad := range []string{"password", "authorization", "api_key", "apikey", "cookie", "bearer ", "secret"} {
		require.NotContains(t, lower, bad)
	}
}
