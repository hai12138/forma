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
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	assetentity "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/fixture"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/repository"
)

const testActor = "1001"

func newTestService(gen ProposalGenerator) (CapabilityService, repository.CapabilityRepository, *MemoryAssetProjection) {
	repo := repository.NewMemoryCapabilityRepository()
	assets := NewMemoryAssetProjection()
	svc := NewCapabilityService(&Components{Repo: repo, Assets: assets, Generator: gen})
	return svc, repo, assets
}

// forceStatusForTest seeds a revision status via repository (Validate is fail-closed until G3).
func forceStatusForTest(t *testing.T, repo repository.CapabilityRepository, tenantID, revisionID string, from, to entity.RevisionStatus) {
	t.Helper()
	ok, err := repo.UpdateRevisionStatus(context.Background(), tenantID, revisionID, from, to)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestManualCreateLifecycle(t *testing.T) {
	svc, repo, assets := newTestService(nil)
	cap, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor,
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

	forceStatusForTest(t, repo, "t1", rev.RevisionID, entity.RevisionDraft, entity.RevisionValidated)

	active, err := svc.Activate(context.Background(), "t1", rev.RevisionID, testActor, "go-live")
	require.NoError(t, err)
	require.Equal(t, entity.RevisionActive, active.Status)
	asset, _ = assets.GetCapabilityAsset(context.Background(), "t1", cap.CapabilityID)
	require.Equal(t, assetentity.AssetStatusReleased, asset.Status)
	require.Equal(t, "0.1.0", asset.SemanticVersion)

	decs, err := repo.ListDecisionsByCapability(context.Background(), "t1", cap.CapabilityID)
	require.NoError(t, err)
	var foundActivate bool
	for _, d := range decs {
		if d.Action == entity.DecisionActivate {
			foundActivate = true
			require.Equal(t, testActor, d.ActorPrincipalID)
			require.Equal(t, "go-live", d.Reason)
			require.Equal(t, rev.RevisionID, d.TargetRevisionID)
		}
	}
	require.True(t, foundActivate)
}

func TestDeriveIdempotency(t *testing.T) {
	svc, _, _ := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, CapabilityID: "cap-d",
		Payload: fixture.ProcurementApprovalCapability(),
	})
	require.NoError(t, err)
	payload := fixture.ProcurementApprovalCapability()
	payload.Name = "SubmitProcurementApproval-v2"
	r1, d1, err := svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: "cap-d", SourceRevisionID: rev.RevisionID,
		ClientRequestID: "derive-1", ActorID: testActor, Payload: payload,
	})
	require.NoError(t, err)
	require.Equal(t, int32(2), r1.Version)
	require.Equal(t, entity.SourceDerivedEdit, r1.Source)

	r2, d2, err := svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: "cap-d", SourceRevisionID: rev.RevisionID,
		ClientRequestID: "derive-1", ActorID: testActor, Payload: payload,
	})
	require.NoError(t, err)
	require.Equal(t, r1.RevisionID, r2.RevisionID)
	require.Equal(t, d1.DecisionID, d2.DecisionID)

	payload.Description = "changed"
	_, _, err = svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: "cap-d", SourceRevisionID: rev.RevisionID,
		ClientRequestID: "derive-1", ActorID: testActor, Payload: payload,
	})
	require.ErrorIs(t, err, entity.ErrIdempotencyConflict)
}

func TestRejectNoShell(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, repo, assets := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "an1", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	require.Len(t, res.Proposals, 1)
	dec, err := svc.RejectProposal(context.Background(), &RejectInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: testActor,
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
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "an2", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	propID := res.Proposals[0].ProposalID
	rev, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: propID, ActorID: testActor, CapabilityID: "cap-ai",
	})
	require.NoError(t, err)
	require.Equal(t, entity.SourceAIProposal, rev.Source)
	asset, err := assets.GetCapabilityAsset(context.Background(), "t1", "cap-ai")
	require.NoError(t, err)
	require.Equal(t, int32(1), asset.Revision)

	rev2, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: propID, ActorID: testActor, CapabilityID: "cap-ai",
	})
	require.NoError(t, err)
	require.Equal(t, rev.RevisionID, rev2.RevisionID)
}

func TestConfirmOntoExistingKeepsACTIVEProjection(t *testing.T) {
	svc, repo, assets := newTestService(&DeterministicFakeGenerator{
		Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()},
	})
	cap, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, CapabilityID: "cap-x",
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	forceStatusForTest(t, repo, "t1", rev.RevisionID, entity.RevisionDraft, entity.RevisionValidated)
	_, err = svc.Activate(context.Background(), "t1", rev.RevisionID, testActor, "")
	require.NoError(t, err)
	before, _ := assets.GetCapabilityAsset(context.Background(), "t1", cap.CapabilityID)

	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "an3", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	draft, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: testActor, CapabilityID: cap.CapabilityID,
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
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 2, ClientRequestID: "same", ActorID: testActor,
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

	failGen := &DeterministicFakeGenerator{Err: errors.New("boom secret password=xyz")}
	svc2, _, _ := newTestService(failGen)
	failed, err := svc2.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 2, ClientRequestID: "fail", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 2},
	})
	require.ErrorIs(t, err, entity.ErrAnalysisFailed)
	require.Equal(t, entity.AnalysisFailed, failed.Run.Status)
	require.NotContains(t, err.Error(), "password")
	require.NotContains(t, failed.Run.ErrorCode, "password")

	failGen.Err = nil
	failGen.Proposals = []entity.SemanticPayload{fixture.ProcurementQueryCapability()}
	retried, err := svc2.RetryFailedAnalysis(context.Background(), "t1", failed.Run.AnalysisRunID, testActor)
	require.NoError(t, err)
	require.Equal(t, entity.AnalysisSucceeded, retried.Run.Status)
	require.Equal(t, failed.Run.AnalysisRunID, retried.Run.AnalysisRunID)
	require.GreaterOrEqual(t, retried.Run.Attempt, int32(2))
}

func TestConcurrentConfirmDistinctDrafts(t *testing.T) {
	payload := fixture.LaboratoryFlowCapability()
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{payload, payload}}
	svc, repo, assets := newTestService(gen)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, CapabilityID: "cap-c",
		Payload: payload,
	})
	require.NoError(t, err)
	forceStatusForTest(t, repo, "t1", rev.RevisionID, entity.RevisionDraft, entity.RevisionValidated)
	_, err = svc.Activate(context.Background(), "t1", rev.RevisionID, testActor, "")
	require.NoError(t, err)

	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "multi", ActorID: testActor,
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
				TenantID: "t1", ProposalID: res.Proposals[idx].ProposalID, ActorID: testActor, CapabilityID: "cap-c",
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
				TenantID: "t1", BusinessID: "biz", ActorID: testActor, CapabilityID: "same-id",
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
		TenantID: "tenant-a", BusinessID: "biz", ActorID: testActor,
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	_, err = svc.GetRevision(context.Background(), "tenant-b", rev.RevisionID)
	require.ErrorIs(t, err, entity.ErrRevisionNotFound)
}

func TestDualIndustryFixturesAgnostic(t *testing.T) {
	svc, _, _ := newTestService(nil)
	_, r1, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "lab", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	_, r2, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "proc", ActorID: testActor, Payload: fixture.ProcurementApprovalCapability(),
	})
	require.NoError(t, err)
	require.Equal(t, entity.KindQuery, r1.CapabilityKind)
	require.Equal(t, entity.KindCommand, r2.CapabilityKind)
}

func TestValidateFailClosedWithoutEvidence(t *testing.T) {
	svc, repo, _ := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	_, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrMissingValidationEvidence)
	got, err := repo.GetRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Equal(t, entity.RevisionDraft, got.Status)
}

func TestMarkStaleFailClosedWithoutEvidence(t *testing.T) {
	svc, repo, _ := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	forceStatusForTest(t, repo, "t1", rev.RevisionID, entity.RevisionDraft, entity.RevisionValidated)
	_, err = svc.Activate(context.Background(), "t1", rev.RevisionID, testActor, "")
	require.NoError(t, err)
	_, err = svc.MarkStale(context.Background(), "t1", rev.RevisionID, testActor, "impact")
	require.ErrorIs(t, err, entity.ErrMissingImpactEvidence)
}

func TestEditConfirmRequiresFullPayloadAndReplay(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, _, _ := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "edit-an", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	payload := fixture.LaboratoryFlowCapability()
	payload.Name = "EditedLabFlow"
	rev, err := svc.EditConfirmProposal(context.Background(), &EditConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: testActor, CapabilityID: "cap-edit",
		EffectivePayload: payload,
	})
	require.NoError(t, err)
	require.Equal(t, "EditedLabFlow", rev.Name)
	require.Equal(t, entity.SourceAIProposal, rev.Source)

	rev2, err := svc.EditConfirmProposal(context.Background(), &EditConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: testActor, CapabilityID: "cap-edit",
		EffectivePayload: payload,
	})
	require.NoError(t, err)
	require.Equal(t, rev.RevisionID, rev2.RevisionID)

	payload.Description = "different"
	_, err = svc.EditConfirmProposal(context.Background(), &EditConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: testActor, CapabilityID: "cap-edit",
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
				TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 3, ClientRequestID: "once", ActorID: testActor,
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
		CreatedBy: testActor,
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
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, CapabilityID: "cap-cons",
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	forceStatusForTest(t, repo, "t1", rev.RevisionID, entity.RevisionDraft, entity.RevisionValidated)
	_, err = svc.Activate(context.Background(), "t1", rev.RevisionID, testActor, "")
	require.NoError(t, err)
	before, err := assets.GetCapabilityAsset(context.Background(), "t1", cap.CapabilityID)
	require.NoError(t, err)

	err = repo.Transaction(context.Background(), func(tx repository.CapabilityRepository) error {
		return tx.UpdateActiveRevisionID(context.Background(), "t1", cap.CapabilityID, "")
	})
	require.NoError(t, err)

	payload := fixture.LaboratoryFlowCapability()
	payload.Name = "DerivedAfterCorrupt"
	_, _, err = svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: cap.CapabilityID, SourceRevisionID: rev.RevisionID,
		ClientRequestID: "corrupt-derive", ActorID: testActor, Payload: payload,
	})
	require.ErrorIs(t, err, entity.ErrConsistency)

	after, err := assets.GetCapabilityAsset(context.Background(), "t1", cap.CapabilityID)
	require.NoError(t, err)
	require.Equal(t, before.Name, after.Name)
	require.Equal(t, before.ContentDigest, after.ContentDigest)
	require.Equal(t, before.Status, after.Status)

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

func TestUoWFailClosedWithoutDB(t *testing.T) {
	repo := repository.NewMemoryCapabilityRepository()
	// GormAssetProjection without DB — configured() must fail closed.
	svc := NewCapabilityService(&Components{
		Repo:   repo,
		Assets: NewGormAssetProjection(nil),
	})
	_, _, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor,
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.ErrorIs(t, err, entity.ErrNotConfigured)
}

func TestAssetProjectionFailureFullRollback(t *testing.T) {
	repo := repository.NewMemoryCapabilityRepository()
	inner := NewMemoryAssetProjection()
	failing := &FailingAssetProjection{Inner: inner, FailUpdate: true}
	svc := NewCapabilityService(&Components{Repo: repo, Assets: failing})

	_, _, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, CapabilityID: "cap-rb",
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.Error(t, err)
	_, err = repo.GetCapability(context.Background(), "t1", "cap-rb")
	require.ErrorIs(t, err, entity.ErrNotFound)
	_, err = inner.GetCapabilityAsset(context.Background(), "t1", "cap-rb")
	require.ErrorIs(t, err, entity.ErrNotFound)
}

func TestConcurrentActivateExactlyOneSuccess(t *testing.T) {
	base := repository.NewMemoryCapabilityRepository()
	assets := NewMemoryAssetProjection()
	gate := make(chan struct{})
	var peeks atomic.Int32
	repo := &activatePeekBarrierRepo{CapabilityRepository: base, peeks: &peeks, gate: gate}
	svc := NewCapabilityService(&Components{Repo: repo, Assets: assets})

	cap, rev1, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, CapabilityID: "cap-act",
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	payload2 := fixture.LaboratoryFlowCapability()
	payload2.Name = "ListLaboratoryDevices-v2"
	rev2, _, err := svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: cap.CapabilityID, SourceRevisionID: rev1.RevisionID,
		ClientRequestID: "d2", ActorID: testActor, Payload: payload2,
	})
	require.NoError(t, err)
	forceStatusForTest(t, base, "t1", rev1.RevisionID, entity.RevisionDraft, entity.RevisionValidated)
	forceStatusForTest(t, base, "t1", rev2.RevisionID, entity.RevisionDraft, entity.RevisionValidated)

	var wg sync.WaitGroup
	errs := make([]error, 2)
	ids := []string{rev1.RevisionID, rev2.RevisionID}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := svc.Activate(context.Background(), "t1", ids[idx], testActor, "race")
			errs[idx] = err
		}(i)
	}
	wg.Wait()
	wins, conflicts := 0, 0
	for _, err := range errs {
		if err == nil {
			wins++
		} else if errors.Is(err, entity.ErrActiveConflict) {
			conflicts++
		} else {
			t.Fatalf("unexpected: %v", err)
		}
	}
	require.Equal(t, 1, wins)
	require.Equal(t, 1, conflicts)
	got, err := base.GetCapability(context.Background(), "t1", cap.CapabilityID)
	require.NoError(t, err)
	require.Equal(t, int64(1), got.AggregateGeneration)
	require.NotEmpty(t, got.ActiveRevisionID)
}

// activatePeekBarrierRepo holds Activate optimistic peeks until both goroutines have read.
type activatePeekBarrierRepo struct {
	repository.CapabilityRepository
	peeks *atomic.Int32
	gate  chan struct{}
}

func (r *activatePeekBarrierRepo) GetCapability(ctx context.Context, tenantID, capabilityID string) (*entity.BusinessCapability, error) {
	cap, err := r.CapabilityRepository.GetCapability(ctx, tenantID, capabilityID)
	if err != nil {
		return nil, err
	}
	if r.peeks.Add(1) == 2 {
		close(r.gate)
	}
	<-r.gate
	return cap, nil
}

func TestProposalTargetMismatchNoWrites(t *testing.T) {
	svc, repo, assets := newTestService(nil)
	_, _, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, CapabilityID: "cap-a",
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	p := &entity.CapabilityProposal{
		ProposalID: "cprop_bound", TenantID: "t1", BusinessID: "biz", AnalysisRunID: "runx",
		CapabilityID: "cap-a", Status: entity.ProposalProposed, Payload: fixture.LaboratoryFlowCapability(),
		CreatedAt: time.Now().UTC(),
	}
	require.NoError(t, repo.CreateProposal(context.Background(), p))
	beforeRevs, err := repo.ListRevisions(context.Background(), "t1", "cap-a")
	require.NoError(t, err)
	_, err = svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: "cprop_bound", ActorID: testActor, CapabilityID: "cap-b",
	})
	require.ErrorIs(t, err, entity.ErrConflict)
	afterRevs, err := repo.ListRevisions(context.Background(), "t1", "cap-a")
	require.NoError(t, err)
	require.Equal(t, len(beforeRevs), len(afterRevs))
	_, err = assets.GetCapabilityAsset(context.Background(), "t1", "cap-b")
	require.ErrorIs(t, err, entity.ErrNotFound)
	got, err := repo.GetProposal(context.Background(), "t1", "cprop_bound")
	require.NoError(t, err)
	require.Equal(t, entity.ProposalProposed, got.Status)
}

func TestIllegalMaterializationPayload(t *testing.T) {
	svc, _, _ := newTestService(nil)
	bad := fixture.LaboratoryFlowCapability()
	bad.Name = ""
	_, _, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, Payload: bad,
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	bad = fixture.LaboratoryFlowCapability()
	bad.CapabilityKind = "UNKNOWN"
	_, _, err = svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, Payload: bad,
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	bad = fixture.LaboratoryFlowCapability()
	bad.QueryOperation = ""
	_, _, err = svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, Payload: bad,
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	bad = fixture.LaboratoryFlowCapability()
	bad.Preconditions = []entity.Precondition{{ID: "x", Predicate: "DROP_TABLE", LogicalKey: "a"}}
	_, _, err = svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, Payload: bad,
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	bad = fixture.LaboratoryFlowCapability()
	bad.Effects = []entity.Effect{{ID: "e", Kind: "EXEC", Description: "x"}}
	_, _, err = svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, Payload: bad,
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	bad = fixture.LaboratoryFlowCapability()
	bad.Description = "run SELECT * FROM users"
	_, _, err = svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, Payload: bad,
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
}

func TestSecretRejectionInPayloadAndAnalysis(t *testing.T) {
	svc, _, _ := newTestService(nil)
	bad := fixture.LaboratoryFlowCapability()
	bad.Name = "GetPasswordReset"
	_, _, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, Payload: bad,
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	bad = fixture.LaboratoryFlowCapability()
	bad.Description = "uses api_key for upstream"
	_, _, err = svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, Payload: bad,
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	_, err = parseOwnerID("not-a-number")
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
	_, _, err = svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: "alice",
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
}

func TestGeneratorErrorSanitized(t *testing.T) {
	gen := &DeterministicFakeGenerator{Err: errors.New("openai api_key leaked in stack")}
	svc, _, _ := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "san", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.ErrorIs(t, err, entity.ErrAnalysisFailed)
	require.NotNil(t, res)
	require.NotNil(t, res.Run)
	require.Equal(t, "FORMA_CAPABILITY_ANALYSIS_FAILED", res.Run.ErrorCode)
	require.NotContains(t, err.Error(), "api_key")
	require.NotContains(t, err.Error(), "openai")
}

type listFailRepo struct {
	repository.CapabilityRepository
	failList bool
}

func (r *listFailRepo) ListProposalsByAnalysisRun(ctx context.Context, tenantID, analysisRunID string) ([]*entity.CapabilityProposal, error) {
	if r.failList {
		return nil, errors.New("list boom")
	}
	return r.CapabilityRepository.ListProposalsByAnalysisRun(ctx, tenantID, analysisRunID)
}

func TestAnalysisRepoErrorPropagation(t *testing.T) {
	base := repository.NewMemoryCapabilityRepository()
	assets := NewMemoryAssetProjection()
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	wrapped := &listFailRepo{CapabilityRepository: base}
	svc := NewCapabilityService(&Components{Repo: wrapped, Assets: assets, Generator: gen})

	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "prop", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	require.NotNil(t, res.Run)

	wrapped.failList = true
	_, err = svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "prop", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.Error(t, err)
}

func TestMemoryLeasePointerIsolation(t *testing.T) {
	repo := repository.NewMemoryCapabilityRepository()
	now := time.Now().UTC()
	exp := now.Add(time.Minute)
	run := &entity.CapabilityAnalysisRun{
		AnalysisRunID: "run-lease", TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1,
		ClientRequestID: "lease", RequestDigest: "d", Status: entity.AnalysisPending, Attempt: 1,
		ExecutionClaimedAt: &now, LeaseExpiresAt: &exp, CreatedBy: testActor,
	}
	got, created, err := repo.CreateOrClaimAnalysisRun(context.Background(), run)
	require.NoError(t, err)
	require.True(t, created)
	require.NotNil(t, got.LeaseExpiresAt)
	*got.LeaseExpiresAt = now.Add(-time.Hour)
	again, err := repo.GetAnalysisRun(context.Background(), "t1", "run-lease")
	require.NoError(t, err)
	require.True(t, again.LeaseExpiresAt.After(now))
}

func TestCorruptJSONFailClosed(t *testing.T) {
	// Memory repo stores via JSON clone; verify ValidateMaterialization rejects bad payloads.
	err := ValidateMaterializationPayload(entity.SemanticPayload{
		Name: "x", CapabilityKind: entity.KindQuery, BusinessModelRevision: 1,
		QueryOperation: entity.QueryOpRead, OutputCardinality: entity.CardinalityOne,
		Preconditions: []entity.Precondition{{ID: "1", Predicate: "EQ", LogicalKey: "k", Comparand: map[string]any{"sql": "SELECT 1"}}},
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
}

func TestDeprecateCreatesDecision(t *testing.T) {
	svc, repo, _ := newTestService(nil)
	cap, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, CapabilityID: "cap-dep",
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	forceStatusForTest(t, repo, "t1", rev.RevisionID, entity.RevisionDraft, entity.RevisionValidated)
	_, err = svc.Activate(context.Background(), "t1", rev.RevisionID, testActor, "on")
	require.NoError(t, err)
	_, err = svc.Deprecate(context.Background(), "t1", rev.RevisionID, testActor, "retire")
	require.NoError(t, err)
	decs, err := repo.ListDecisionsByCapability(context.Background(), "t1", cap.CapabilityID)
	require.NoError(t, err)
	var found bool
	for _, d := range decs {
		if d.Action == entity.DecisionDeprecate {
			found = true
			require.Equal(t, "retire", d.Reason)
			require.Equal(t, rev.RevisionID, d.TargetRevisionID)
		}
	}
	require.True(t, found)
}
