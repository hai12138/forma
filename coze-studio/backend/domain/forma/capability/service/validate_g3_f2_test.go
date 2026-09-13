/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/fixture"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/repository"
)

func TestG3F2BMCurrentRevisionAdvanceValidateFAIL(t *testing.T) {
	svc, repo, _, bm, contracts := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev)
	bm.Put(&BusinessModelRevisionEvidence{
		TenantID: "t1", BusinessID: "biz-lab", Revision: 2, ContentDigest: "bm-digest-v2",
	})
	out, result, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrValidationFailed)
	require.Equal(t, entity.RevisionDraft, out.Status)
	require.Contains(t, result.IssueCodes, entity.IssueBMNotFound)
	got, err := repo.GetRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Equal(t, entity.RevisionDraft, got.Status)
}

func TestG3F2BMAdvanceAfterValidateActivateMissingEvidence(t *testing.T) {
	svc, repo, _, bm, contracts := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev)
	_, _, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.NoError(t, err)

	bm.Put(&BusinessModelRevisionEvidence{
		TenantID: "t1", BusinessID: "biz-lab", Revision: 2, ContentDigest: "bm-digest-v2",
	})
	_, err = svc.Activate(context.Background(), "t1", rev.RevisionID, testActor, "bm-advanced")
	require.ErrorIs(t, err, entity.ErrMissingValidationEvidence)
	got, err := repo.GetRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Equal(t, entity.RevisionValidated, got.Status)
}

func TestG3F2AIEmptyCapabilityIDFAIL(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, repo, _, bm, contracts := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz-lab", BusinessModelRevision: 1, ClientRequestID: "g3f2-empty-cap", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	rev, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: testActor, CapabilityID: "cap-f2-empty-cap",
	})
	require.NoError(t, err)
	prop, err := repo.GetProposal(context.Background(), "t1", rev.ProposalID)
	require.NoError(t, err)
	prop.CapabilityID = ""
	require.NoError(t, repository.ReplaceProposalForTest(repo, prop))
	seedPortsForRevision(bm, contracts, rev)
	_, _, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrMissingProvenance)
}

func TestG3F2AIEmptyMaterializedRevisionIDFAIL(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, repo, _, bm, contracts := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz-lab", BusinessModelRevision: 1, ClientRequestID: "g3f2-empty-mat", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	rev, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: testActor, CapabilityID: "cap-f2-empty-mat",
	})
	require.NoError(t, err)
	prop, err := repo.GetProposal(context.Background(), "t1", rev.ProposalID)
	require.NoError(t, err)
	prop.MaterializedRevisionID = ""
	require.NoError(t, repository.ReplaceProposalForTest(repo, prop))
	seedPortsForRevision(bm, contracts, rev)
	_, _, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrMissingProvenance)
}

func TestG3F2AIWrongProposalBusinessIDFAIL(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, repo, _, bm, contracts := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz-lab", BusinessModelRevision: 1, ClientRequestID: "g3f2-wrong-biz", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	rev, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: testActor, CapabilityID: "cap-f2-wrong-biz",
	})
	require.NoError(t, err)
	prop, err := repo.GetProposal(context.Background(), "t1", rev.ProposalID)
	require.NoError(t, err)
	prop.BusinessID = "biz-other"
	require.NoError(t, repository.ReplaceProposalForTest(repo, prop))
	seedPortsForRevision(bm, contracts, rev)
	_, _, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrMissingProvenance)
}

func TestG3F2AIDecisionEmptyCapabilityIDFAIL(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, repo, _, bm, contracts := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz-lab", BusinessModelRevision: 1, ClientRequestID: "g3f2-dec-cap", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	rev, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: testActor, CapabilityID: "cap-f2-dec-cap",
	})
	require.NoError(t, err)
	decs, err := repo.ListDecisionsByCapability(context.Background(), "t1", rev.CapabilityID)
	require.NoError(t, err)
	var confirm *entity.CapabilityDecision
	for _, d := range decs {
		if d.ProposalID == rev.ProposalID && (d.Action == entity.DecisionConfirm || d.Action == entity.DecisionEditConfirm) {
			cp := *d
			confirm = &cp
			break
		}
	}
	require.NotNil(t, confirm)
	confirm.CapabilityID = ""
	require.NoError(t, repository.ReplaceDecisionForTest(repo, confirm))
	seedPortsForRevision(bm, contracts, rev)
	_, _, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrMissingProvenance)
}

func TestG3F2AnalysisRunBusinessIDMismatchFAIL(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, repo, _, bm, contracts := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz-lab", BusinessModelRevision: 1, ClientRequestID: "g3f2-run-biz", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	rev, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: testActor, CapabilityID: "cap-f2-run-biz",
	})
	require.NoError(t, err)
	run, err := repo.GetAnalysisRun(context.Background(), "t1", rev.AnalysisRunID)
	require.NoError(t, err)
	run.BusinessID = "biz-other"
	require.NoError(t, repository.ReplaceAnalysisRunForTest(repo, run))
	seedPortsForRevision(bm, contracts, rev)
	_, _, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrMissingProvenance)
}

func TestG3F2AnalysisRunBMRevisionMismatchFAIL(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, repo, _, bm, contracts := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz-lab", BusinessModelRevision: 1, ClientRequestID: "g3f2-run-bm", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	rev, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: testActor, CapabilityID: "cap-f2-run-bm",
	})
	require.NoError(t, err)
	run, err := repo.GetAnalysisRun(context.Background(), "t1", rev.AnalysisRunID)
	require.NoError(t, err)
	run.BusinessModelRevision = 99
	require.NoError(t, repository.ReplaceAnalysisRunForTest(repo, run))
	seedPortsForRevision(bm, contracts, rev)
	_, _, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrMissingProvenance)
}

func TestG3F2DerivedSourceMissingFAIL(t *testing.T) {
	svc, repo, _, bm, contracts := newTestService(nil)
	_, rev1, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	payload2 := fixture.LaboratoryFlowCapability()
	payload2.Name = "ListLaboratoryDevices-v2"
	rev2, _, err := svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: rev1.CapabilityID, SourceRevisionID: rev1.RevisionID,
		ClientRequestID: "derive-f2-miss", ActorID: testActor, Action: entity.DecisionDerive, Payload: payload2,
	})
	require.NoError(t, err)
	decs, err := repo.ListDecisionsByCapability(context.Background(), "t1", rev2.CapabilityID)
	require.NoError(t, err)
	var derive *entity.CapabilityDecision
	for _, d := range decs {
		if d.TargetRevisionID == rev2.RevisionID && (d.Action == entity.DecisionDerive || d.Action == entity.DecisionEdit) {
			cp := *d
			derive = &cp
			break
		}
	}
	require.NotNil(t, derive)
	derive.SourceRevisionID = "crev_missing_src"
	require.NoError(t, repository.ReplaceDecisionForTest(repo, derive))
	rev2.DerivedFromRevisionID = "crev_missing_src"
	require.NoError(t, repository.ReplaceRevisionForTest(repo, rev2))
	seedPortsForRevision(bm, contracts, rev2)
	_, _, err = svc.Validate(context.Background(), "t1", rev2.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrMissingProvenance)
}

func TestG3F2DerivedSourceOtherBusinessFAIL(t *testing.T) {
	svc, repo, _, bm, contracts := newTestService(nil)
	_, rev1, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	payload2 := fixture.LaboratoryFlowCapability()
	payload2.Name = "ListLaboratoryDevices-v2"
	rev2, _, err := svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: rev1.CapabilityID, SourceRevisionID: rev1.RevisionID,
		ClientRequestID: "derive-f2-biz", ActorID: testActor, Action: entity.DecisionDerive, Payload: payload2,
	})
	require.NoError(t, err)
	now := time.Now().UTC()
	foreign := &entity.BusinessCapabilityRevision{
		RevisionID: "crev_foreign_biz", CapabilityID: "cap_foreign_biz", TenantID: "t1", BusinessID: "biz-other",
		Version: 1, Status: entity.RevisionDraft, Name: "foreign", BusinessModelRevision: 1,
		CapabilityKind: entity.KindQuery, Source: entity.SourceManualCreated,
		CreatedBy: testActor, CreatedAt: now,
	}
	require.NoError(t, repo.CreateCapability(context.Background(), &entity.BusinessCapability{
		CapabilityID: "cap_foreign_biz", TenantID: "t1", BusinessID: "biz-other", CreatedBy: testActor, CreatedAt: now, UpdatedAt: now,
	}))
	require.NoError(t, repo.CreateRevision(context.Background(), foreign))

	decs, err := repo.ListDecisionsByCapability(context.Background(), "t1", rev2.CapabilityID)
	require.NoError(t, err)
	var derive *entity.CapabilityDecision
	for _, d := range decs {
		if d.TargetRevisionID == rev2.RevisionID && (d.Action == entity.DecisionDerive || d.Action == entity.DecisionEdit) {
			cp := *d
			derive = &cp
			break
		}
	}
	require.NotNil(t, derive)
	derive.SourceRevisionID = foreign.RevisionID
	require.NoError(t, repository.ReplaceDecisionForTest(repo, derive))
	rev2.DerivedFromRevisionID = foreign.RevisionID
	require.NoError(t, repository.ReplaceRevisionForTest(repo, rev2))
	seedPortsForRevision(bm, contracts, rev2)
	_, _, err = svc.Validate(context.Background(), "t1", rev2.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrMissingProvenance)
}

func TestG3F2DerivedSourceOtherCapabilityFAIL(t *testing.T) {
	svc, repo, _, bm, contracts := newTestService(nil)
	_, rev1, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	_, other, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, CapabilityID: "cap-other-src",
		Payload: fixture.LaboratoryCommandCapability(),
	})
	require.NoError(t, err)
	payload2 := fixture.LaboratoryFlowCapability()
	payload2.Name = "ListLaboratoryDevices-v2"
	rev2, _, err := svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: rev1.CapabilityID, SourceRevisionID: rev1.RevisionID,
		ClientRequestID: "derive-f2-cap", ActorID: testActor, Action: entity.DecisionDerive, Payload: payload2,
	})
	require.NoError(t, err)
	decs, err := repo.ListDecisionsByCapability(context.Background(), "t1", rev2.CapabilityID)
	require.NoError(t, err)
	var derive *entity.CapabilityDecision
	for _, d := range decs {
		if d.TargetRevisionID == rev2.RevisionID && (d.Action == entity.DecisionDerive || d.Action == entity.DecisionEdit) {
			cp := *d
			derive = &cp
			break
		}
	}
	require.NotNil(t, derive)
	derive.SourceRevisionID = other.RevisionID
	require.NoError(t, repository.ReplaceDecisionForTest(repo, derive))
	rev2.DerivedFromRevisionID = other.RevisionID
	require.NoError(t, repository.ReplaceRevisionForTest(repo, rev2))
	seedPortsForRevision(bm, contracts, rev2)
	_, _, err = svc.Validate(context.Background(), "t1", rev2.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrMissingProvenance)
}

// flipAfterReadsContractPort returns stable descriptors until the Nth Get, then swaps to drifted.
type flipAfterReadsContractPort struct {
	mu       sync.Mutex
	reads    int
	flipAt   int
	stable   *FakeContractPort
	drifted  *ContractLogicalDescriptor
	driftKey string
}

func (p *flipAfterReadsContractPort) GetActiveContractLogicalDescriptor(ctx context.Context, tenantID, businessID, contractID string) (*ContractLogicalDescriptor, error) {
	p.mu.Lock()
	p.reads++
	n := p.reads
	p.mu.Unlock()
	if n == p.flipAt && p.drifted != nil {
		p.stable.Put(p.drifted)
	}
	return p.stable.GetActiveContractLogicalDescriptor(ctx, tenantID, businessID, contractID)
}

func TestG3F2FinalEvidenceFenceValidateDetectsDrift(t *testing.T) {
	uow := NewMemoryUnitOfWork()
	bm := NewFakeBusinessModelPort()
	stable := NewFakeContractPort()
	flip := &flipAfterReadsContractPort{stable: stable, flipAt: 2}
	svc := NewCapabilityService(&Components{UoW: uow, Business: bm, Contract: flip})
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, stable, rev)
	desc, err := stable.GetActiveContractLogicalDescriptor(context.Background(), "t1", "biz-lab", "dc_lab")
	require.NoError(t, err)
	drifted := cloneContractDescriptor(desc)
	drifted.Classification = map[string]string{"device_id": "INTERNAL"}
	flip.drifted = drifted
	// seedPorts + the peek above already read once; reset counter after seeding.
	flip.mu.Lock()
	flip.reads = 0
	flip.mu.Unlock()

	_, result, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrConsistency)
	require.Nil(t, result)
	got, err := uow.Root().GetRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Equal(t, entity.RevisionDraft, got.Status)
	list, err := uow.Root().ListValidationsByRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestG3F2FinalEvidenceFenceActivateDetectsDrift(t *testing.T) {
	uow := NewMemoryUnitOfWork()
	bm := NewFakeBusinessModelPort()
	stable := NewFakeContractPort()
	svcPass := NewCapabilityService(&Components{UoW: uow, Business: bm, Contract: stable})
	_, rev, err := svcPass.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, stable, rev)
	_, _, err = svcPass.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.NoError(t, err)

	desc, err := stable.GetActiveContractLogicalDescriptor(context.Background(), "t1", "biz-lab", "dc_lab")
	require.NoError(t, err)
	drifted := cloneContractDescriptor(desc)
	drifted.Classification = map[string]string{"device_id": "INTERNAL"}
	flip := &flipAfterReadsContractPort{stable: stable, flipAt: 2, drifted: drifted}
	svcAct := NewCapabilityService(&Components{UoW: uow, Business: bm, Contract: flip})

	_, err = svcAct.Activate(context.Background(), "t1", rev.RevisionID, testActor, "fence-drift")
	require.ErrorIs(t, err, entity.ErrConsistency)
	got, err := uow.Root().GetRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Equal(t, entity.RevisionValidated, got.Status)
}
