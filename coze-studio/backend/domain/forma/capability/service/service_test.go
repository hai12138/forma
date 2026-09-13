/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func newTestService(gen ProposalGenerator) (CapabilityService, repository.CapabilityRepository, *MemoryAssetProjection, *FakeBusinessModelPort, *FakeContractPort) {
	uow := NewMemoryUnitOfWork()
	bm := NewFakeBusinessModelPort()
	contracts := NewFakeContractPort()
	svc := NewCapabilityService(&Components{UoW: uow, Generator: gen, Business: bm, Contract: contracts})
	return svc, uow.Root(), uow.AssetsView(), bm, contracts
}

// rootOverrideUoW overrides Root() for peek/list test seams while keeping owned UoW transactions.
type rootOverrideUoW struct {
	inner CapabilityUnitOfWork
	root  repository.CapabilityRepository
}

func (u *rootOverrideUoW) Root() repository.CapabilityRepository { return u.root }
func (u *rootOverrideUoW) WithinTransaction(ctx context.Context, fn func(tx CapabilityTx) error) error {
	return u.inner.WithinTransaction(ctx, fn)
}

// seedPortsForRevision registers BM + Active contract descriptors matching the revision payload.
func seedPortsForRevision(bm *FakeBusinessModelPort, contracts *FakeContractPort, rev *entity.BusinessCapabilityRevision) {
	bm.Put(&BusinessModelRevisionEvidence{
		TenantID: rev.TenantID, BusinessID: rev.BusinessID,
		Revision: rev.BusinessModelRevision, ContentDigest: "bm-digest-" + rev.BusinessID,
	})
	for _, b := range rev.DataContractBindings {
		fields := make([]ContractLogicalField, 0)
		filters := make([]ContractFilterFieldSpec, 0)
		seen := map[string]struct{}{}
		for _, m := range b.LogicalFieldMappings {
			if _, ok := seen[m.ContractLogicalKey]; ok {
				continue
			}
			seen[m.ContractLogicalKey] = struct{}{}
			lt := "STRING"
			for _, f := range append(append([]entity.LogicalField{}, rev.InputSchema.Fields...), rev.OutputSchema.Fields...) {
				if f.LogicalKey == m.CapabilityLogicalKey {
					lt = f.LogicalType
					break
				}
			}
			fields = append(fields, ContractLogicalField{
				LogicalKey: m.ContractLogicalKey, LogicalType: lt, Nullable: false,
			})
			filters = append(filters, ContractFilterFieldSpec{
				LogicalKey: m.ContractLogicalKey, Operators: []string{"EQ", "NE", "GT", "GTE", "LT", "LTE", "IN"},
			})
		}
		contracts.Put(&ContractLogicalDescriptor{
			TenantID: rev.TenantID, BusinessID: rev.BusinessID,
			ContractID: b.DataContractID, RevisionID: b.DataContractRevisionID, Version: b.DataContractVersion,
			BusinessModelRevision: rev.BusinessModelRevision, Status: "ACTIVE",
			LogicalSchema: fields, QueryCapabilities: []string{"READ", "LIST", "FILTER"}, FilterSchema: filters,
			PaginationPolicy: ContractPaginationPolicy{DefaultLimit: 20, MaxLimit: 100},
		})
	}
}

// forceStatusForTest seeds a revision status via repository (prefer Validate in G3 tests).
func forceStatusForTest(t *testing.T, repo repository.CapabilityRepository, tenantID, revisionID string, from, to entity.RevisionStatus) {
	t.Helper()
	ok, err := repo.UpdateRevisionStatus(context.Background(), tenantID, revisionID, from, to)
	require.NoError(t, err)
	require.True(t, ok)
}

// seedPASSValidationForTest writes PASS evidence matching current ports for Activate gate tests.
func seedPASSValidationForTest(t *testing.T, repo repository.CapabilityRepository, bm *FakeBusinessModelPort, contracts *FakeContractPort, rev *entity.BusinessCapabilityRevision) {
	t.Helper()
	seedPortsForRevision(bm, contracts, rev)
	revDigest, err := CapabilityContentDigest(rev.ToSemanticPayload())
	require.NoError(t, err)
	descs := []*ContractLogicalDescriptor{}
	for _, b := range rev.DataContractBindings {
		d, err := contracts.GetActiveContractLogicalDescriptor(context.Background(), rev.TenantID, rev.BusinessID, b.DataContractID)
		require.NoError(t, err)
		descs = append(descs, d)
	}
	contractDigest, err := ContractEvidenceDigest(descs)
	require.NoError(t, err)
	bmEv, err := bm.GetBusinessModelRevision(context.Background(), rev.TenantID, rev.BusinessID, rev.BusinessModelRevision)
	require.NoError(t, err)
	evidenceDigest, err := ValidationEvidenceDigest(rev.BusinessModelRevision, bmEv.ContentDigest, contractDigest)
	require.NoError(t, err)
	now := time.Now().UTC()
	require.NoError(t, repo.CreateValidationResult(context.Background(), &entity.CapabilityValidationResult{
		ValidationID: "cval_seed_" + rev.RevisionID, TenantID: rev.TenantID, BusinessID: rev.BusinessID,
		CapabilityID: rev.CapabilityID, RevisionID: rev.RevisionID,
		RevisionContentDigest: revDigest, BusinessModelRevision: rev.BusinessModelRevision,
		BusinessModelContentDigest: bmEv.ContentDigest, ContractEvidenceDigest: contractDigest,
		EvidenceDigest: evidenceDigest, Status: entity.ValidationPass, IssueCodes: nil,
		ValidatedBy: testActor, ValidatedAt: now, CreatedAt: now,
	}))
}

func TestManualCreateLifecycle(t *testing.T) {
	svc, repo, assets, bmPort, contractPort := newTestService(nil)
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

	seedPASSValidationForTest(t, repo, bmPort, contractPort, rev)
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
	svc, _, _, _, _ := newTestService(nil)
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
	svc, repo, assets, _, _ := newTestService(gen)
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
	svc, _, assets, _, _ := newTestService(gen)
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
	svc, repo, assets, bmPort, contractPort := newTestService(&DeterministicFakeGenerator{
		Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()},
	})
	cap, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, CapabilityID: "cap-x",
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	seedPASSValidationForTest(t, repo, bmPort, contractPort, rev)
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
	svc, _, _, _, _ := newTestService(gen)
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
	svc2, _, _, _, _ := newTestService(failGen)
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
	svc, repo, assets, bmPort, contractPort := newTestService(gen)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, CapabilityID: "cap-c",
		Payload: payload,
	})
	require.NoError(t, err)
	seedPASSValidationForTest(t, repo, bmPort, contractPort, rev)
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
	svc, repo, assets, _, _ := newTestService(nil)
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
	svc, _, _, _, _ := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "tenant-a", BusinessID: "biz", ActorID: testActor,
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	_, err = svc.GetRevision(context.Background(), "tenant-b", rev.RevisionID)
	require.ErrorIs(t, err, entity.ErrRevisionNotFound)
}

func TestDualIndustryFixturesAgnostic(t *testing.T) {
	svc, _, _, _, _ := newTestService(nil)
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
	uow := NewMemoryUnitOfWork()
	svc := NewCapabilityService(&Components{UoW: uow}) // no validation ports
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	_, _, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrPortsNotConfigured)
	got, err := uow.Root().GetRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Equal(t, entity.RevisionDraft, got.Status)
}

func TestMarkStaleFailClosedWithoutEvidence(t *testing.T) {
	svc, repo, _, bmPort, contractPort := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	seedPASSValidationForTest(t, repo, bmPort, contractPort, rev)
	forceStatusForTest(t, repo, "t1", rev.RevisionID, entity.RevisionDraft, entity.RevisionValidated)
	_, err = svc.Activate(context.Background(), "t1", rev.RevisionID, testActor, "")
	require.NoError(t, err)
	_, err = svc.MarkStale(context.Background(), "t1", rev.RevisionID, testActor, "impact")
	require.ErrorIs(t, err, entity.ErrMissingImpactEvidence)
}

func TestEditConfirmRequiresFullPayloadAndReplay(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, _, _, _, _ := newTestService(gen)
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
	svc, _, _, _, _ := newTestService(gen)
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
	svc, repo, assets, bmPort, contractPort := newTestService(nil)
	cap, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, CapabilityID: "cap-cons",
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	seedPASSValidationForTest(t, repo, bmPort, contractPort, rev)
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

func TestUoWFailClosedWithoutUoW(t *testing.T) {
	svc := NewCapabilityService(&Components{})
	_, _, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor,
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.ErrorIs(t, err, entity.ErrNotConfigured)

	svc2 := NewCapabilityService(nil)
	_, _, err = svc2.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor,
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.ErrorIs(t, err, entity.ErrNotConfigured)
}

func TestAssetProjectionFailureFullRollback(t *testing.T) {
	var assetFails atomic.Int32
	uow := NewMemoryUnitOfWorkWithOptions(MemoryUoWOptions{
		FailAssetUpdate: true,
		OnAssetFail:     func() { assetFails.Add(1) },
	})
	svc := NewCapabilityService(&Components{UoW: uow})

	_, _, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, CapabilityID: "cap-rb",
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.Error(t, err)
	require.GreaterOrEqual(t, assetFails.Load(), int32(1))
	_, err = uow.Root().GetCapability(context.Background(), "t1", "cap-rb")
	require.ErrorIs(t, err, entity.ErrNotFound)
	_, err = uow.AssetsView().GetCapabilityAsset(context.Background(), "t1", "cap-rb")
	require.ErrorIs(t, err, entity.ErrNotFound)
}

func TestAssetCreateFailureFullRollback(t *testing.T) {
	var assetFails atomic.Int32
	uow := NewMemoryUnitOfWorkWithOptions(MemoryUoWOptions{
		FailAssetCreate: true,
		OnAssetFail:     func() { assetFails.Add(1) },
	})
	svc := NewCapabilityService(&Components{UoW: uow})

	_, _, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, CapabilityID: "cap-cr",
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.Error(t, err)
	require.Equal(t, int32(1), assetFails.Load())
	_, err = uow.Root().GetCapability(context.Background(), "t1", "cap-cr")
	require.ErrorIs(t, err, entity.ErrNotFound)
	_, err = uow.AssetsView().GetCapabilityAsset(context.Background(), "t1", "cap-cr")
	require.ErrorIs(t, err, entity.ErrNotFound)
}

func TestMemoryUoWCallbackFailureRollsBackCapAndAsset(t *testing.T) {
	uow := NewMemoryUnitOfWork()
	repo := uow.Root()
	assets := uow.AssetsView()
	err := uow.WithinTransaction(context.Background(), func(tx CapabilityTx) error {
		cap := &entity.BusinessCapability{
			CapabilityID: "cap-uow", TenantID: "t1", BusinessID: "biz",
			CreatedBy: testActor, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		}
		require.NoError(t, tx.Repo().CreateCapability(context.Background(), cap))
		require.NoError(t, tx.Assets().CreateCapabilityAsset(context.Background(), &assetentity.AssetRef{
			TenantID: "t1", AssetID: "cap-uow", Kind: assetentity.AssetKindCapability, Name: "x",
			SemanticVersion: "0.0.0", Revision: 1, SchemaVersion: "1.0", Status: assetentity.AssetStatusDraft,
		}))
		return errors.New("force rollback")
	})
	require.Error(t, err)
	_, err = repo.GetCapability(context.Background(), "t1", "cap-uow")
	require.ErrorIs(t, err, entity.ErrNotFound)
	_, err = assets.GetCapabilityAsset(context.Background(), "t1", "cap-uow")
	require.ErrorIs(t, err, entity.ErrNotFound)
}

func TestMemoryUoWCommitFailureRollsBackCapAndAsset(t *testing.T) {
	var commitFails atomic.Int32
	uow := NewMemoryUnitOfWorkWithOptions(MemoryUoWOptions{
		FailCommit:   true,
		OnCommitFail: func() { commitFails.Add(1) },
	})
	err := uow.WithinTransaction(context.Background(), func(tx CapabilityTx) error {
		cap := &entity.BusinessCapability{
			CapabilityID: "cap-commit", TenantID: "t1", BusinessID: "biz",
			CreatedBy: testActor, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		}
		require.NoError(t, tx.Repo().CreateCapability(context.Background(), cap))
		require.NoError(t, tx.Assets().CreateCapabilityAsset(context.Background(), &assetentity.AssetRef{
			TenantID: "t1", AssetID: "cap-commit", Kind: assetentity.AssetKindCapability, Name: "x",
			SemanticVersion: "0.0.0", Revision: 1, SchemaVersion: "1.0", Status: assetentity.AssetStatusDraft,
		}))
		return nil
	})
	require.ErrorIs(t, err, entity.ErrUoWCommitFailed)
	require.Equal(t, int32(1), commitFails.Load())
	_, err = uow.Root().GetCapability(context.Background(), "t1", "cap-commit")
	require.ErrorIs(t, err, entity.ErrNotFound)
	_, err = uow.AssetsView().GetCapabilityAsset(context.Background(), "t1", "cap-commit")
	require.ErrorIs(t, err, entity.ErrNotFound)
}

func TestConcurrentActivateExactlyOneSuccess(t *testing.T) {
	inner := NewMemoryUnitOfWork()
	gate := make(chan struct{})
	var peeks atomic.Int32
	repo := &activatePeekBarrierRepo{CapabilityRepository: inner.Root(), peeks: &peeks, gate: gate}
	uow := &rootOverrideUoW{inner: inner, root: repo}
	bm := NewFakeBusinessModelPort()
	contracts := NewFakeContractPort()
	svc := NewCapabilityService(&Components{UoW: uow, Business: bm, Contract: contracts})

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
	seedPASSValidationForTest(t, inner.Root(), bm, contracts, rev1)
	seedPASSValidationForTest(t, inner.Root(), bm, contracts, rev2)
	forceStatusForTest(t, inner.Root(), "t1", rev1.RevisionID, entity.RevisionDraft, entity.RevisionValidated)
	forceStatusForTest(t, inner.Root(), "t1", rev2.RevisionID, entity.RevisionDraft, entity.RevisionValidated)

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
	got, err := inner.Root().GetCapability(context.Background(), "t1", cap.CapabilityID)
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
	svc, repo, assets, _, _ := newTestService(nil)
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

func TestConfirmedReplayCapabilityIDMismatch(t *testing.T) {
	svc, repo, assets, _, _ := newTestService(&DeterministicFakeGenerator{
		Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()},
	})
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "term-c", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	propID := res.Proposals[0].ProposalID
	rev, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: propID, ActorID: testActor, CapabilityID: "cap-term-a",
	})
	require.NoError(t, err)
	require.NotEmpty(t, rev.RevisionID)

	beforeDecs, err := repo.ListDecisionsByCapability(context.Background(), "t1", "cap-term-a")
	require.NoError(t, err)
	_, err = svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: propID, ActorID: testActor, CapabilityID: "cap-term-b",
	})
	require.ErrorIs(t, err, entity.ErrConflict)
	afterDecs, err := repo.ListDecisionsByCapability(context.Background(), "t1", "cap-term-a")
	require.NoError(t, err)
	require.Equal(t, len(beforeDecs), len(afterDecs))
	_, err = assets.GetCapabilityAsset(context.Background(), "t1", "cap-term-b")
	require.ErrorIs(t, err, entity.ErrNotFound)
	prop, err := repo.GetProposal(context.Background(), "t1", propID)
	require.NoError(t, err)
	require.Equal(t, entity.ProposalConfirmed, prop.Status)
	require.Equal(t, "cap-term-a", prop.CapabilityID)
}

func TestEditConfirmedReplayCapabilityIDMismatch(t *testing.T) {
	svc, repo, assets, _, _ := newTestService(&DeterministicFakeGenerator{
		Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()},
	})
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "term-e", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	propID := res.Proposals[0].ProposalID
	payload := fixture.LaboratoryFlowCapability()
	payload.Name = "EditedTermFlow"
	rev, err := svc.EditConfirmProposal(context.Background(), &EditConfirmInput{
		TenantID: "t1", ProposalID: propID, ActorID: testActor, CapabilityID: "cap-edit-a",
		EffectivePayload: payload,
	})
	require.NoError(t, err)
	require.Equal(t, "EditedTermFlow", rev.Name)

	beforeDecs, err := repo.ListDecisionsByCapability(context.Background(), "t1", "cap-edit-a")
	require.NoError(t, err)
	_, err = svc.EditConfirmProposal(context.Background(), &EditConfirmInput{
		TenantID: "t1", ProposalID: propID, ActorID: testActor, CapabilityID: "cap-edit-b",
		EffectivePayload: payload,
	})
	require.ErrorIs(t, err, entity.ErrConflict)
	afterDecs, err := repo.ListDecisionsByCapability(context.Background(), "t1", "cap-edit-a")
	require.NoError(t, err)
	require.Equal(t, len(beforeDecs), len(afterDecs))
	_, err = assets.GetCapabilityAsset(context.Background(), "t1", "cap-edit-b")
	require.ErrorIs(t, err, entity.ErrNotFound)
	prop, err := repo.GetProposal(context.Background(), "t1", propID)
	require.NoError(t, err)
	require.Equal(t, entity.ProposalEditConfirmed, prop.Status)
	require.Equal(t, "cap-edit-a", prop.CapabilityID)
}

func TestIllegalMaterializationPayload(t *testing.T) {
	svc, _, _, _, _ := newTestService(nil)
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
	bad.Preconditions = []entity.Precondition{{ID: "x", Predicate: entity.PredicateKind("DROP_TABLE"), LogicalKey: "a"}}
	_, _, err = svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, Payload: bad,
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	bad = fixture.LaboratoryFlowCapability()
	bad.Effects = []entity.Effect{{ID: "e", Kind: entity.EffectKind("EXEC"), Description: "x"}}
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
	svc, _, _, _, _ := newTestService(nil)
	ok := fixture.LaboratoryFlowCapability()
	ok.Name = "GetPasswordReset"
	ok.Description = "password reset flow"
	_, _, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, Payload: ok,
	})
	require.NoError(t, err)

	ok = fixture.LaboratoryFlowCapability()
	ok.Name = "ReviewTradeSecretPolicy"
	ok.Description = "Review trade secret policy"
	_, _, err = svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz2", ActorID: testActor, Payload: ok,
	})
	require.NoError(t, err)

	bad := fixture.LaboratoryFlowCapability()
	bad.Description = "password=hunter2"
	_, _, err = svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, Payload: bad,
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	bad = fixture.LaboratoryFlowCapability()
	bad.InputSchema.Fields = []entity.LogicalField{{LogicalKey: "api_key", LogicalType: "STRING"}}
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
	svc, _, _, _, _ := newTestService(gen)
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
	inner := NewMemoryUnitOfWork()
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	wrapped := &listFailRepo{CapabilityRepository: inner.Root()}
	uow := &rootOverrideUoW{inner: inner, root: wrapped}
	svc := NewCapabilityService(&Components{UoW: uow, Generator: gen})

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
		Preconditions: []entity.Precondition{{ID: "1", Predicate: entity.PredicateEQ, LogicalKey: "k", Comparand: map[string]any{"sql": "SELECT 1"}}},
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
}

func TestDeprecateCreatesDecision(t *testing.T) {
	svc, repo, _, bmPort, contractPort := newTestService(nil)
	cap, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, CapabilityID: "cap-dep",
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	seedPASSValidationForTest(t, repo, bmPort, contractPort, rev)
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

func TestQueryPairingTable(t *testing.T) {
	cases := []struct {
		name string
		p    entity.SemanticPayload
		ok   bool
	}{
		{"READ_ONE", entity.SemanticPayload{Name: "a", CapabilityKind: entity.KindQuery, BusinessModelRevision: 1, QueryOperation: entity.QueryOpRead, OutputCardinality: entity.CardinalityOne}, true},
		{"READ_MANY", entity.SemanticPayload{Name: "a", CapabilityKind: entity.KindQuery, BusinessModelRevision: 1, QueryOperation: entity.QueryOpRead, OutputCardinality: entity.CardinalityMany}, false},
		{"LIST_MANY", entity.SemanticPayload{Name: "a", CapabilityKind: entity.KindQuery, BusinessModelRevision: 1, QueryOperation: entity.QueryOpList, OutputCardinality: entity.CardinalityMany}, true},
		{"LIST_ONE", entity.SemanticPayload{Name: "a", CapabilityKind: entity.KindQuery, BusinessModelRevision: 1, QueryOperation: entity.QueryOpList, OutputCardinality: entity.CardinalityOne}, false},
		{"FILTER_MANY", entity.SemanticPayload{Name: "a", CapabilityKind: entity.KindQuery, BusinessModelRevision: 1, QueryOperation: entity.QueryOpFilter, OutputCardinality: entity.CardinalityMany}, true},
		{"FILTER_ONE", entity.SemanticPayload{Name: "a", CapabilityKind: entity.KindQuery, BusinessModelRevision: 1, QueryOperation: entity.QueryOpFilter, OutputCardinality: entity.CardinalityOne}, false},
		{"COMMAND_empty", entity.SemanticPayload{Name: "a", CapabilityKind: entity.KindCommand, BusinessModelRevision: 1}, true},
		{"COMMAND_with_op", entity.SemanticPayload{Name: "a", CapabilityKind: entity.KindCommand, BusinessModelRevision: 1, QueryOperation: entity.QueryOpRead}, false},
		{"COMMAND_with_card", entity.SemanticPayload{Name: "a", CapabilityKind: entity.KindCommand, BusinessModelRevision: 1, OutputCardinality: entity.CardinalityOne}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMaterializationPayload(tc.p)
			if tc.ok {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, entity.ErrInvalidPayload)
			}
		})
	}
}

func TestMaterializationBypassAttempts(t *testing.T) {
	// Operator SQL via Predicate
	err := ValidateMaterializationPayload(entity.SemanticPayload{
		Name: "x", CapabilityKind: entity.KindQuery, BusinessModelRevision: 1,
		QueryOperation: entity.QueryOpRead, OutputCardinality: entity.CardinalityOne,
		Preconditions: []entity.Precondition{{ID: "1", Predicate: entity.PredicateKind("SELECT *"), LogicalKey: "k"}},
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	// list comparand shell
	err = ValidateMaterializationPayload(entity.SemanticPayload{
		Name: "x", CapabilityKind: entity.KindQuery, BusinessModelRevision: 1,
		QueryOperation: entity.QueryOpRead, OutputCardinality: entity.CardinalityOne,
		Preconditions: []entity.Precondition{{ID: "1", Predicate: entity.PredicateIN, LogicalKey: "k", Comparand: []string{"ok", "rm -rf /"}}},
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	// binding credential key name "token"
	err = ValidateMaterializationPayload(entity.SemanticPayload{
		Name: "x", CapabilityKind: entity.KindQuery, BusinessModelRevision: 1,
		QueryOperation: entity.QueryOpRead, OutputCardinality: entity.CardinalityOne,
		DataContractBindings: []entity.DataContractBinding{{
			DataContractID: "token", DataContractRevisionID: "r1", DataContractVersion: 1,
		}},
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	// effect logical key bypass
	err = ValidateMaterializationPayload(entity.SemanticPayload{
		Name: "x", CapabilityKind: entity.KindCommand, BusinessModelRevision: 1,
		Effects: []entity.Effect{{ID: "e1", Kind: entity.EffectIntent, LogicalKey: "password"}},
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
}

func TestAnalysisOpaqueIDRejection(t *testing.T) {
	svc, _, _, _, _ := newTestService(&DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}})
	_, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "oid", ActorID: testActor,
		Analysis: entity.AnalysisRequest{
			BusinessModelRevision: 1,
			DataContractPins:      []entity.DataContractPin{{DataContractID: "dc drop table", DataContractVersion: 1}},
		},
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	_, err = svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "oid2", ActorID: testActor,
		Analysis: entity.AnalysisRequest{
			BusinessModelRevision: 1,
			RequirementRefs:       []string{"req; SELECT 1"},
		},
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
}

func TestLeaseTakeoverUsesPersistedRequestJSON(t *testing.T) {
	fixed := &fixedClock{t: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	uow := NewMemoryUnitOfWork()
	gen := &recordingGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc := NewCapabilityService(&Components{UoW: uow, Generator: gen, Clock: fixed})

	persisted := entity.AnalysisRequest{
		BusinessModelRevision: 1,
		DataContractPins:      []entity.DataContractPin{{DataContractID: "dc_persisted", DataContractVersion: 1}},
	}
	digest, err := AnalysisRequestDigest(persisted)
	require.NoError(t, err)
	persistedJSON, err := json.Marshal(persisted)
	require.NoError(t, err)

	exp := fixed.Now().Add(-time.Minute)
	claimedAt := fixed.Now().Add(-10 * time.Minute)
	seed := &entity.CapabilityAnalysisRun{
		AnalysisRunID: "run-expired", TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1,
		ClientRequestID: "lease-exp", RequestDigest: digest, Status: entity.AnalysisPending, Attempt: 1,
		RequestJSON: string(persistedJSON), ExecutionClaimedAt: &claimedAt, LeaseExpiresAt: &exp,
		CreatedBy: testActor, CreatedAt: fixed.Now(), UpdatedAt: fixed.Now(),
	}
	_, created, err := uow.Root().CreateOrClaimAnalysisRun(context.Background(), seed)
	require.NoError(t, err)
	require.True(t, created)
	require.NoError(t, uow.Root().CreateAnalysisAttempt(context.Background(), &entity.CapabilityAnalysisAttempt{
		AttemptID: "att-lease", AnalysisRunID: "run-expired", TenantID: "t1", Attempt: 1,
		ActorPrincipalID: testActor, TriggerKind: entity.AttemptTriggerFirst,
		ResultStatus: entity.AttemptResultPending, CreatedAt: fixed.Now(),
	}))

	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "lease-exp", ActorID: testActor,
		Analysis: persisted,
	})
	require.NoError(t, err)
	require.True(t, res.OwnedExecute)
	require.Equal(t, "dc_persisted", gen.lastAnalysis.DataContractPins[0].DataContractID)

	atts, err := uow.Root().ListAnalysisAttempts(context.Background(), "t1", "run-expired")
	require.NoError(t, err)
	require.Len(t, atts, 2)
	require.Equal(t, entity.AttemptResultSuperseded, atts[0].ResultStatus)
	require.NotNil(t, atts[0].CompletedAt)
	require.Equal(t, entity.AttemptTriggerLeaseTakeover, atts[1].TriggerKind)
	require.Equal(t, entity.AttemptResultSucceeded, atts[1].ResultStatus)
	require.NotNil(t, atts[1].CompletedAt)
}

func TestLeaseTakeoverDigestMismatchFailsClosed(t *testing.T) {
	fixed := &fixedClock{t: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	uow := NewMemoryUnitOfWork()
	gen := &recordingGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc := NewCapabilityService(&Components{UoW: uow, Generator: gen, Clock: fixed})

	caller := entity.AnalysisRequest{
		BusinessModelRevision: 1,
		DataContractPins:      []entity.DataContractPin{{DataContractID: "dc_caller", DataContractVersion: 1}},
	}
	digest, err := AnalysisRequestDigest(caller)
	require.NoError(t, err)
	// Forged: RequestDigest matches caller, but JSON body differs.
	persistedJSON := `{"business_model_revision":1,"data_contract_pins":[{"data_contract_id":"dc_persisted","data_contract_version":1}],"requirement_refs":null}`
	exp := fixed.Now().Add(-time.Minute)
	claimedAt := fixed.Now().Add(-10 * time.Minute)
	seed := &entity.CapabilityAnalysisRun{
		AnalysisRunID: "run-forged", TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1,
		ClientRequestID: "lease-forge", RequestDigest: digest, Status: entity.AnalysisPending, Attempt: 1,
		RequestJSON: persistedJSON, ExecutionClaimedAt: &claimedAt, LeaseExpiresAt: &exp,
		CreatedBy: testActor, CreatedAt: fixed.Now(), UpdatedAt: fixed.Now(),
	}
	_, created, err := uow.Root().CreateOrClaimAnalysisRun(context.Background(), seed)
	require.NoError(t, err)
	require.True(t, created)
	require.NoError(t, uow.Root().CreateAnalysisAttempt(context.Background(), &entity.CapabilityAnalysisAttempt{
		AttemptID: "att-forge", AnalysisRunID: "run-forged", TenantID: "t1", Attempt: 1,
		ActorPrincipalID: testActor, TriggerKind: entity.AttemptTriggerFirst,
		ResultStatus: entity.AttemptResultPending, CreatedAt: fixed.Now(),
	}))

	_, err = svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "lease-forge", ActorID: testActor,
		Analysis: caller,
	})
	require.ErrorIs(t, err, entity.ErrConsistency)
	require.Empty(t, gen.lastAnalysis.DataContractPins) // generator never called
	failed, err := uow.Root().GetAnalysisRun(context.Background(), "t1", "run-forged")
	require.NoError(t, err)
	require.Equal(t, entity.AnalysisFailed, failed.Status)
}

type fixedClock struct{ t time.Time }

func (c *fixedClock) Now() time.Time { return c.t }

type recordingGenerator struct {
	Proposals    []entity.SemanticPayload
	lastAnalysis entity.AnalysisRequest
	Err          error
}

func (g *recordingGenerator) Generate(_ context.Context, req GenerateRequest) (*GenerateResult, error) {
	g.lastAnalysis = req.Analysis
	if g.Err != nil {
		return nil, g.Err
	}
	return &GenerateResult{ModelRef: "fake", Proposals: g.Proposals}, nil
}

func TestCorruptRequestJSONMarkFailedPropagates(t *testing.T) {
	uow := NewMemoryUnitOfWork()
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	now := time.Now().UTC()
	exp := now.Add(-time.Minute)
	caller := entity.AnalysisRequest{BusinessModelRevision: 1}
	digest, err := AnalysisRequestDigest(caller)
	require.NoError(t, err)
	seed := &entity.CapabilityAnalysisRun{
		AnalysisRunID: "run-corrupt", TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1,
		ClientRequestID: "corrupt", RequestDigest: digest, Status: entity.AnalysisPending, Attempt: 1,
		RequestJSON: "{not-json", ExecutionClaimedAt: &now, LeaseExpiresAt: &exp,
		CreatedBy: testActor, CreatedAt: now, UpdatedAt: now,
	}
	_, created, err := uow.Root().CreateOrClaimAnalysisRun(context.Background(), seed)
	require.NoError(t, err)
	require.True(t, created)
	require.NoError(t, uow.Root().CreateAnalysisAttempt(context.Background(), &entity.CapabilityAnalysisAttempt{
		AttemptID: "att1", AnalysisRunID: "run-corrupt", TenantID: "t1", Attempt: 1,
		ActorPrincipalID: testActor, TriggerKind: entity.AttemptTriggerFirst,
		ResultStatus: entity.AttemptResultPending, CreatedAt: now,
	}))

	svc := NewCapabilityService(&Components{UoW: uow, Generator: gen})
	_, err = svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "corrupt", ActorID: testActor,
		Analysis: caller,
	})
	require.ErrorIs(t, err, entity.ErrConsistency)
	failed, err := uow.Root().GetAnalysisRun(context.Background(), "t1", "run-corrupt")
	require.NoError(t, err)
	require.Equal(t, entity.AnalysisFailed, failed.Status)
}

func TestAnalysisAttemptAuditOnRetry(t *testing.T) {
	gen := &DeterministicFakeGenerator{Err: errors.New("boom")}
	svc, repo, _, _, _ := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "retry-aud", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.ErrorIs(t, err, entity.ErrAnalysisFailed)
	require.Equal(t, entity.AnalysisFailed, res.Run.Status)

	atts, err := repo.ListAnalysisAttempts(context.Background(), "t1", res.Run.AnalysisRunID)
	require.NoError(t, err)
	require.Len(t, atts, 1)
	require.Equal(t, entity.AttemptTriggerFirst, atts[0].TriggerKind)
	require.Equal(t, entity.AttemptResultFailed, atts[0].ResultStatus)

	gen.Err = nil
	gen.Proposals = []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}
	res2, err := svc.RetryFailedAnalysis(context.Background(), "t1", res.Run.AnalysisRunID, testActor)
	require.NoError(t, err)
	require.Equal(t, entity.AnalysisSucceeded, res2.Run.Status)

	atts, err = repo.ListAnalysisAttempts(context.Background(), "t1", res.Run.AnalysisRunID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(atts), 2)
	var foundRetry bool
	for _, a := range atts {
		if a.TriggerKind == entity.AttemptTriggerRetry {
			foundRetry = true
			require.Equal(t, entity.AttemptResultSucceeded, a.ResultStatus)
		}
	}
	require.True(t, foundRetry)
}

func TestConfirmedReplayCapabilityMismatch(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, repo, _, _, _ := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "bind-term", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	rev, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: testActor, CapabilityID: "cap-term",
	})
	require.NoError(t, err)
	require.Equal(t, "cap-term", rev.CapabilityID)

	_, err = svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: testActor, CapabilityID: "cap-other",
	})
	require.ErrorIs(t, err, entity.ErrConflict)
	got, err := repo.GetProposal(context.Background(), "t1", res.Proposals[0].ProposalID)
	require.NoError(t, err)
	require.Equal(t, entity.ProposalConfirmed, got.Status)
}

func TestEditConfirmedReplayCapabilityMismatch(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, repo, _, _, _ := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "bind-edit", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	eff := fixture.LaboratoryFlowCapability()
	eff.Name = "ListLaboratoryDevices-edited"
	rev, err := svc.EditConfirmProposal(context.Background(), &EditConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: testActor, CapabilityID: "cap-edit2",
		EffectivePayload: eff,
	})
	require.NoError(t, err)
	require.Equal(t, "cap-edit2", rev.CapabilityID)

	_, err = svc.EditConfirmProposal(context.Background(), &EditConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: testActor, CapabilityID: "cap-wrong",
		EffectivePayload: eff,
	})
	require.ErrorIs(t, err, entity.ErrConflict)
	got, err := repo.GetProposal(context.Background(), "t1", res.Proposals[0].ProposalID)
	require.NoError(t, err)
	require.Equal(t, entity.ProposalEditConfirmed, got.Status)
}

func TestMapRepoErrorNeverLeaksDriver(t *testing.T) {
	err := MapRepoError(errors.New("Error 1062: Duplicate entry 'x' for key 'PRIMARY'"))
	require.ErrorIs(t, err, entity.ErrConflict)
	require.Equal(t, entity.ErrConflict.Error(), err.Error())
	require.NotContains(t, strings.ToLower(err.Error()), "sql")
	require.NotContains(t, strings.ToLower(err.Error()), "mysql")
	require.NotContains(t, strings.ToLower(err.Error()), "duplicate")
	require.NotContains(t, err.Error(), "1062")

	wrapped := fmt.Errorf("driver: %w", entity.ErrConflict)
	err = MapRepoError(wrapped)
	require.Equal(t, entity.ErrConflict, err)
	require.Equal(t, entity.ErrConflict.Error(), err.Error())

	err = MapRepoError(errors.New("pq: connection refused"))
	require.ErrorIs(t, err, entity.ErrConsistency)
	require.Equal(t, entity.ErrConsistency.Error(), err.Error())
	require.NotContains(t, err.Error(), "pq:")
}

func TestAnalysisActorIDRequired(t *testing.T) {
	svc, _, _, _, _ := newTestService(&DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}})
	_, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "no-actor",
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
}

func TestAnalysisCredentialShapeRejection(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, _, _, _, _ := newTestService(gen)

	// Samples must pass ValidateOpaqueID first, then fail credential-shape (no '=' / spaces).
	opaqueSecrets := []string{
		"eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.sig",
		"ghp_abcdefghijklmnopqrstuvwxyz12",
		"sk-proj-abcdefghijklmnopQR",
	}
	for i, tok := range opaqueSecrets {
		require.NoError(t, ValidateOpaqueID(tok), "sample %q must pass ValidateOpaqueID", tok)
		require.True(t, containsCredentialShape(tok), "sample %q must hit credential-shape", tok)
		before := gen.CallCount()
		_, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
			TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1,
			ClientRequestID: fmt.Sprintf("cred-%d", i), ActorID: testActor,
			Analysis: entity.AnalysisRequest{
				BusinessModelRevision: 1,
				RequirementRefs:       []string{tok},
			},
		})
		require.ErrorIs(t, err, entity.ErrInvalidPayload)
		require.Equal(t, before, gen.CallCount(), "generator must not run for %q", tok)
		require.NotContains(t, err.Error(), tok)
	}

	// Ordinary opaque ref with substring "secret" remains allowed.
	err := ValidateAnalysisRequest(entity.AnalysisRequest{
		BusinessModelRevision: 1,
		RequirementRefs:       []string{"req_secret_ref"},
	})
	require.NoError(t, err)
}

func TestPredicateKindNoTrimSpace(t *testing.T) {
	err := ValidateMaterializationPayload(entity.SemanticPayload{
		Name: "X", CapabilityKind: entity.KindQuery, BusinessModelRevision: 1,
		QueryOperation: entity.QueryOpRead, OutputCardinality: entity.CardinalityOne,
		Preconditions: []entity.Precondition{{ID: "p1", Predicate: entity.PredicateKind(" EQ "), LogicalKey: "k", Comparand: "v"}},
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
}

func TestComparandRulesByPredicate(t *testing.T) {
	base := entity.SemanticPayload{
		Name: "X", CapabilityKind: entity.KindQuery, BusinessModelRevision: 1,
		QueryOperation: entity.QueryOpRead, OutputCardinality: entity.CardinalityOne,
	}
	p := base
	p.Preconditions = []entity.Precondition{{ID: "p1", Predicate: entity.PredicateIN, LogicalKey: "k", Comparand: []string{}}}
	require.ErrorIs(t, ValidateMaterializationPayload(p), entity.ErrInvalidPayload)

	p = base
	p.Preconditions = []entity.Precondition{{ID: "p1", Predicate: entity.PredicateEQ, LogicalKey: "k"}}
	require.ErrorIs(t, ValidateMaterializationPayload(p), entity.ErrInvalidPayload)

	p = base
	p.Preconditions = []entity.Precondition{{ID: "p1", Predicate: entity.PredicateExists, LogicalKey: "k", Comparand: "x"}}
	require.ErrorIs(t, ValidateMaterializationPayload(p), entity.ErrInvalidPayload)

	p = base
	p.Preconditions = []entity.Precondition{{ID: "p1", Predicate: entity.PredicateEQ, LogicalKey: "k", Comparand: "ok"}}
	require.NoError(t, ValidateMaterializationPayload(p))
}

func TestConcurrentLeaseClaimExactlyOne(t *testing.T) {
	fixed := &fixedClock{t: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)}
	uow := NewMemoryUnitOfWork()
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc := NewCapabilityService(&Components{UoW: uow, Generator: gen, Clock: fixed})

	req := entity.AnalysisRequest{BusinessModelRevision: 1}
	digest, err := AnalysisRequestDigest(req)
	require.NoError(t, err)
	raw, err := json.Marshal(req)
	require.NoError(t, err)
	exp := fixed.Now().Add(-time.Minute)
	claimedAt := fixed.Now().Add(-10 * time.Minute)
	seed := &entity.CapabilityAnalysisRun{
		AnalysisRunID: "run-race", TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1,
		ClientRequestID: "lease-race", RequestDigest: digest, Status: entity.AnalysisPending, Attempt: 1,
		RequestJSON: string(raw), ExecutionClaimedAt: &claimedAt, LeaseExpiresAt: &exp,
		CreatedBy: testActor, CreatedAt: fixed.Now(), UpdatedAt: fixed.Now(),
	}
	_, created, err := uow.Root().CreateOrClaimAnalysisRun(context.Background(), seed)
	require.NoError(t, err)
	require.True(t, created)
	require.NoError(t, uow.Root().CreateAnalysisAttempt(context.Background(), &entity.CapabilityAnalysisAttempt{
		AttemptID: "att-race", AnalysisRunID: "run-race", TenantID: "t1", Attempt: 1,
		ActorPrincipalID: testActor, TriggerKind: entity.AttemptTriggerFirst,
		ResultStatus: entity.AttemptResultPending, CreatedAt: fixed.Now(),
	}))

	var wg sync.WaitGroup
	results := make([]*AnalysisResult, 2)
	errs := make([]error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = svc.StartAnalysis(context.Background(), &StartAnalysisInput{
				TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "lease-race", ActorID: testActor,
				Analysis: req,
			})
		}(i)
	}
	wg.Wait()
	owners := 0
	for i := 0; i < 2; i++ {
		if errs[i] == nil && results[i] != nil && results[i].OwnedExecute {
			owners++
		}
	}
	require.Equal(t, 1, owners)
}

func TestStructuralLogicalFieldValidation(t *testing.T) {
	base := func() entity.SemanticPayload {
		return entity.SemanticPayload{
			Name: "X", CapabilityKind: entity.KindQuery, BusinessModelRevision: 1,
			QueryOperation: entity.QueryOpRead, OutputCardinality: entity.CardinalityOne,
			InputSchema: entity.LogicalSchema{Fields: []entity.LogicalField{
				{LogicalKey: "work_cell_id", LogicalType: "STRING"},
			}},
		}
	}

	emptyKey := base()
	emptyKey.InputSchema.Fields = []entity.LogicalField{{LogicalKey: "", LogicalType: "STRING"}}
	require.ErrorIs(t, ValidateMaterializationPayload(emptyKey), entity.ErrInvalidPayload)

	emptyType := base()
	emptyType.InputSchema.Fields = []entity.LogicalField{{LogicalKey: "work_cell_id", LogicalType: ""}}
	require.ErrorIs(t, ValidateMaterializationPayload(emptyType), entity.ErrInvalidPayload)

	unknownType := base()
	unknownType.InputSchema.Fields = []entity.LogicalField{{LogicalKey: "work_cell_id", LogicalType: "VARCHAR"}}
	require.ErrorIs(t, ValidateMaterializationPayload(unknownType), entity.ErrInvalidPayload)

	paddedType := base()
	paddedType.InputSchema.Fields = []entity.LogicalField{{LogicalKey: "work_cell_id", LogicalType: " STRING"}}
	require.ErrorIs(t, ValidateMaterializationPayload(paddedType), entity.ErrInvalidPayload)
}

func TestStructuralBindingAndPinVersions(t *testing.T) {
	svcPayload := func(version int32) entity.SemanticPayload {
		return entity.SemanticPayload{
			Name: "X", CapabilityKind: entity.KindQuery, BusinessModelRevision: 1,
			QueryOperation: entity.QueryOpRead, OutputCardinality: entity.CardinalityOne,
			DataContractBindings: []entity.DataContractBinding{{
				DataContractID: "dc_lab", DataContractRevisionID: "dcr_lab_1", DataContractVersion: version,
			}},
		}
	}
	require.ErrorIs(t, ValidateMaterializationPayload(svcPayload(0)), entity.ErrInvalidPayload)
	require.ErrorIs(t, ValidateMaterializationPayload(svcPayload(-1)), entity.ErrInvalidPayload)
	require.NoError(t, ValidateMaterializationPayload(svcPayload(1)))

	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, _, _, _, _ := newTestService(gen)
	for i, ver := range []int32{0, -3} {
		before := gen.CallCount()
		_, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
			TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1,
			ClientRequestID: fmt.Sprintf("pin-ver-%d", i), ActorID: testActor,
			Analysis: entity.AnalysisRequest{
				BusinessModelRevision: 1,
				DataContractPins:      []entity.DataContractPin{{DataContractID: "dc_ok", DataContractVersion: ver}},
			},
		})
		require.ErrorIs(t, err, entity.ErrInvalidPayload)
		require.Equal(t, before, gen.CallCount())
	}
}
