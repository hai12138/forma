/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	assetentity "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/fixture"
)

func TestG3ValidateManualCreatedPASS(t *testing.T) {
	svc, repo, assets, bm, contracts := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor,
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev)
	out, result, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.NoError(t, err)
	require.Equal(t, entity.RevisionValidated, out.Status)
	require.Equal(t, entity.ValidationPass, result.Status)
	require.Empty(t, result.IssueCodes)
	asset, err := assets.GetCapabilityAsset(context.Background(), "t1", rev.CapabilityID)
	require.NoError(t, err)
	require.Equal(t, assetentity.AssetStatusVerified, asset.Status)
	list, err := repo.ListValidationsByRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Len(t, list, 1)
}

func TestG3ValidateAIProposalConfirmPASS(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, _, _, bm, contracts := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz-lab", BusinessModelRevision: 1, ClientRequestID: "g3-ai", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	rev, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: testActor, CapabilityID: "cap-ai-ok",
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev)
	out, result, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.NoError(t, err)
	require.Equal(t, entity.ValidationPass, result.Status)
	require.Equal(t, entity.RevisionValidated, out.Status)
}

func TestG3ValidateIdempotentReplay(t *testing.T) {
	svc, repo, _, bm, contracts := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev)
	_, r1, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.NoError(t, err)
	_, r2, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.NoError(t, err)
	require.Equal(t, r1.ValidationID, r2.ValidationID)
	list, err := repo.ListValidationsByRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Len(t, list, 1)
}

func TestG3ValidateFAILLeavesDraft(t *testing.T) {
	svc, repo, assets, bm, contracts := newTestService(nil)
	payload := fixture.LaboratoryFlowCapability()
	payload.DataContractBindings[0].DataContractRevisionID = "dcr_wrong"
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: payload,
	})
	require.NoError(t, err)
	// Seed ACTIVE descriptor with correct revision — pin mismatch → FAIL.
	seedPortsForRevision(bm, contracts, rev)
	contracts.Put(&ContractLogicalDescriptor{
		TenantID: "t1", BusinessID: "biz-lab", ContractID: "dc_lab",
		RevisionID: "dcr_lab_1", Version: 1, Status: "ACTIVE",
		LogicalSchema: []ContractLogicalField{
			{LogicalKey: "work_cell_id", LogicalType: "STRING"},
			{LogicalKey: "device_id", LogicalType: "STRING"},
			{LogicalKey: "device_status", LogicalType: "STRING"},
		},
		QueryCapabilities: []string{"LIST"},
	})
	out, result, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrValidationFailed)
	require.NotNil(t, result)
	require.Equal(t, entity.ValidationFail, result.Status)
	require.Contains(t, result.IssueCodes, entity.IssueContractPinMismatch)
	require.Equal(t, entity.RevisionDraft, out.Status)
	got, err := repo.GetRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Equal(t, entity.RevisionDraft, got.Status)
	list, err := repo.ListValidationsByRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, entity.ValidationFail, list[0].Status)
	gotByEv, err := repo.GetValidationByEvidence(context.Background(), "t1", rev.RevisionID, result.EvidenceDigest)
	require.NoError(t, err)
	require.Equal(t, result.ValidationID, gotByEv.ValidationID)
	asset, err := assets.GetCapabilityAsset(context.Background(), "t1", rev.CapabilityID)
	require.NoError(t, err)
	require.Equal(t, assetentity.AssetStatusDraft, asset.Status)
}

func TestG3BMIsolation(t *testing.T) {
	svc, _, _, bm, contracts := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-a", ActorID: testActor, Payload: fixture.LaboratoryCommandCapability(),
	})
	require.NoError(t, err)
	bm.Put(&BusinessModelRevisionEvidence{TenantID: "t2", BusinessID: "biz-a", Revision: 1, ContentDigest: "x"})
	_, result, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrValidationFailed)
	require.Contains(t, result.IssueCodes, entity.IssueBMNotFound)

	bm.Put(&BusinessModelRevisionEvidence{TenantID: "t1", BusinessID: "biz-b", Revision: 1, ContentDigest: "x"})
	_, result, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrValidationFailed)
	require.Contains(t, result.IssueCodes, entity.IssueBMNotFound)

	_ = contracts
	seedPortsForRevision(bm, contracts, rev)
	out, result, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.NoError(t, err)
	require.Equal(t, entity.ValidationPass, result.Status)
	require.Equal(t, entity.RevisionValidated, out.Status)
}

func TestG3ContractCompatibilityMatrix(t *testing.T) {
	svc, _, _, bm, contracts := newTestService(nil)
	payload := fixture.LaboratoryFlowCapability()
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: payload,
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev)

	// unknown contract key
	bad := *contracts
	_ = bad
	contracts.Put(&ContractLogicalDescriptor{
		TenantID: "t1", BusinessID: "biz-lab", ContractID: "dc_lab",
		RevisionID: "dcr_lab_1", Version: 1, Status: "ACTIVE",
		LogicalSchema: []ContractLogicalField{
			{LogicalKey: "work_cell_id", LogicalType: "STRING"},
			// missing device fields
		},
		QueryCapabilities: []string{"LIST"},
	})
	_, result, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrValidationFailed)
	require.Contains(t, result.IssueCodes, entity.IssueMappingUnknownContractKey)

	// type mismatch
	seedPortsForRevision(bm, contracts, rev)
	desc, _ := contracts.GetActiveContractLogicalDescriptor(context.Background(), "t1", "biz-lab", "dc_lab")
	for i := range desc.LogicalSchema {
		if desc.LogicalSchema[i].LogicalKey == "device_id" {
			desc.LogicalSchema[i].LogicalType = "INT64"
		}
	}
	contracts.Put(desc)
	_, result, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrValidationFailed)
	require.Contains(t, result.IssueCodes, entity.IssueTypeMismatch)

	// required output → nullable contract
	seedPortsForRevision(bm, contracts, rev)
	payload2 := fixture.LaboratoryFlowCapability()
	payload2.Name = "LabRead"
	payload2.QueryOperation = entity.QueryOpRead
	payload2.OutputCardinality = entity.CardinalityOne
	payload2.OutputSchema.Fields[0].Required = true
	_, rev2, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, CapabilityID: "cap-null", Payload: payload2,
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev2)
	desc2, _ := contracts.GetActiveContractLogicalDescriptor(context.Background(), "t1", "biz-lab", "dc_lab")
	for i := range desc2.LogicalSchema {
		if desc2.LogicalSchema[i].LogicalKey == "device_id" {
			desc2.LogicalSchema[i].Nullable = true
		}
	}
	contracts.Put(desc2)
	_, result, err = svc.Validate(context.Background(), "t1", rev2.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrValidationFailed)
	require.Contains(t, result.IssueCodes, entity.IssueNullability)
}

func TestG3QueryRules(t *testing.T) {
	svc, _, _, bm, contracts := newTestService(nil)

	// READ/MANY invalid — structural rejects at ManualCreate
	bad := fixture.LaboratoryFlowCapability()
	bad.QueryOperation = entity.QueryOpRead
	bad.OutputCardinality = entity.CardinalityMany
	_, _, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz", ActorID: testActor, Payload: bad,
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	// FILTER without comparison predicate
	filterPayload := fixture.ProcurementQueryCapability()
	filterPayload.Preconditions = nil
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-proc", ActorID: testActor, Payload: filterPayload,
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev)
	_, result, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrValidationFailed)
	require.Contains(t, result.IssueCodes, entity.IssueQueryFilterPredicate)

	// FILTER PASS
	_, revOK, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-proc", ActorID: testActor, CapabilityID: "cap-filter-ok",
		Payload: fixture.ProcurementQueryCapability(),
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, revOK)
	_, result, err = svc.Validate(context.Background(), "t1", revOK.RevisionID, testActor)
	require.NoError(t, err)
	require.Equal(t, entity.ValidationPass, result.Status)

	// COMMAND without binding PASS
	_, revCmd, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-proc", ActorID: testActor, CapabilityID: "cap-cmd",
		Payload: fixture.ProcurementApprovalCapability(),
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, revCmd)
	_, result, err = svc.Validate(context.Background(), "t1", revCmd.RevisionID, testActor)
	require.NoError(t, err)
	require.Equal(t, entity.ValidationPass, result.Status)
}

func TestG3ActivateEvidenceGate(t *testing.T) {
	svc, repo, _, bm, contracts := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	forceStatusForTest(t, repo, "t1", rev.RevisionID, entity.RevisionDraft, entity.RevisionValidated)
	seedPortsForRevision(bm, contracts, rev)
	_, err = svc.Activate(context.Background(), "t1", rev.RevisionID, testActor, "x")
	require.ErrorIs(t, err, entity.ErrMissingValidationEvidence)

	seedPASSValidationForTest(t, repo, bm, contracts, rev)
	active, err := svc.Activate(context.Background(), "t1", rev.RevisionID, testActor, "x")
	require.NoError(t, err)
	require.Equal(t, entity.RevisionActive, active.Status)

	// Drift Active descriptor after Validate → Activate fail-closed on second revision
	_, rev2, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, CapabilityID: "cap-drift",
		Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev2)
	_, _, err = svc.Validate(context.Background(), "t1", rev2.RevisionID, testActor)
	require.NoError(t, err)
	contracts.Put(&ContractLogicalDescriptor{
		TenantID: "t1", BusinessID: "biz-lab", ContractID: "dc_lab",
		RevisionID: "dcr_lab_2", Version: 2, Status: "ACTIVE",
		LogicalSchema: []ContractLogicalField{
			{LogicalKey: "work_cell_id", LogicalType: "STRING"},
			{LogicalKey: "device_id", LogicalType: "STRING"},
			{LogicalKey: "device_status", LogicalType: "STRING"},
		},
		QueryCapabilities: []string{"LIST"},
	})
	_, err = svc.Activate(context.Background(), "t1", rev2.RevisionID, testActor, "drift")
	require.ErrorIs(t, err, entity.ErrMissingValidationEvidence)
}

func TestG3ValidateAssetFailureRollsBack(t *testing.T) {
	uow := NewMemoryUnitOfWork()
	bm := NewFakeBusinessModelPort()
	contracts := NewFakeContractPort()
	svc := NewCapabilityService(&Components{UoW: uow, Business: bm, Contract: contracts})
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev)
	uow.SetFailAssetUpdate(true)
	_, _, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.Error(t, err)
	got, err := uow.Root().GetRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Equal(t, entity.RevisionDraft, got.Status)
	list, err := uow.Root().ListValidationsByRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestG3ConcurrentValidateIdempotent(t *testing.T) {
	svc, repo, _, bm, contracts := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev)
	var wg sync.WaitGroup
	var passCount atomic.Int32
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, result, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
			if err == nil && result != nil && result.Status == entity.ValidationPass {
				passCount.Add(1)
			}
		}()
	}
	wg.Wait()
	require.GreaterOrEqual(t, passCount.Load(), int32(1))
	list, err := repo.ListValidationsByRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	got, err := repo.GetRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Equal(t, entity.RevisionValidated, got.Status)
}

func TestG3DerivedEditProvenance(t *testing.T) {
	svc, _, _, bm, contracts := newTestService(nil)
	_, rev1, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	payload2 := fixture.LaboratoryFlowCapability()
	payload2.Name = "ListLaboratoryDevices-v2"
	rev2, _, err := svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: rev1.CapabilityID, SourceRevisionID: rev1.RevisionID,
		ClientRequestID: "derive-g3", ActorID: testActor, Action: entity.DecisionDerive, Payload: payload2,
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev2)
	out, result, err := svc.Validate(context.Background(), "t1", rev2.RevisionID, testActor)
	require.NoError(t, err)
	require.Equal(t, entity.ValidationPass, result.Status)
	require.Equal(t, entity.RevisionValidated, out.Status)
}

func TestG3GeneratorCallsRemainZero(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, _, _, bm, contracts := newTestService(gen)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev)
	_, _, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.NoError(t, err)
	require.Equal(t, 0, gen.Calls)
}

func TestG3AIProposalMissingConfirmProvenance(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, repo, _, bm, contracts := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz-lab", BusinessModelRevision: 1, ClientRequestID: "g3-noconfirm", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	// Materialize a DRAFT with AI_PROPOSAL source without Confirm Decision (direct CreateRevision seam).
	payload := fixture.LaboratoryFlowCapability()
	now := time.Now().UTC()
	rev := &entity.BusinessCapabilityRevision{
		RevisionID: "crev_ai_orphan", CapabilityID: "cap_ai_orphan", TenantID: "t1", BusinessID: "biz-lab",
		Version: 1, Status: entity.RevisionDraft, Name: payload.Name, Description: payload.Description,
		BusinessModelRevision: 1, CapabilityKind: payload.CapabilityKind,
		InputSchema: payload.InputSchema, OutputSchema: payload.OutputSchema,
		Preconditions: payload.Preconditions, Effects: payload.Effects,
		DataContractBindings: payload.DataContractBindings,
		QueryOperation: payload.QueryOperation, OutputCardinality: payload.OutputCardinality,
		AnalysisRunID: res.Run.AnalysisRunID, ProposalID: res.Proposals[0].ProposalID,
		Source: entity.SourceAIProposal, CreatedBy: testActor, CreatedAt: now,
	}
	require.NoError(t, repo.CreateCapability(context.Background(), &entity.BusinessCapability{
		CapabilityID: "cap_ai_orphan", TenantID: "t1", BusinessID: "biz-lab", CreatedBy: testActor, CreatedAt: now, UpdatedAt: now,
	}))
	require.NoError(t, repo.CreateRevision(context.Background(), rev))
	seedPortsForRevision(bm, contracts, rev)
	_, _, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrMissingProvenance)
}

func TestG3UnmappedCapabilityKeyFAIL(t *testing.T) {
	svc, _, _, bm, contracts := newTestService(nil)
	payload := fixture.LaboratoryFlowCapability()
	payload.DataContractBindings[0].LogicalFieldMappings = payload.DataContractBindings[0].LogicalFieldMappings[:1] // drop outputs
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: payload,
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev)
	_, result, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrValidationFailed)
	require.Contains(t, result.IssueCodes, entity.IssueMappingMissing)
	for _, code := range result.IssueCodes {
		require.NotContains(t, strings.ToLower(code), "token")
		require.NotContains(t, strings.ToLower(code), "password")
		require.NotContains(t, strings.ToLower(code), "mysql")
	}
}

func TestG3HistoricalBusinessModelRevisionRejectedWhenNotCurrent(t *testing.T) {
	svc, _, _, bm, contracts := newTestService(nil)
	payload := fixture.LaboratoryCommandCapability()
	payload.BusinessModelRevision = 3
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: payload,
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev)
	// Master CurrentRevision advances past the Capability pin — fail closed (S5-G3-F2).
	bm.Put(&BusinessModelRevisionEvidence{TenantID: "t1", BusinessID: "biz-lab", Revision: 9, ContentDigest: "latest-bm-9"})
	out, result, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrValidationFailed)
	require.Equal(t, entity.RevisionDraft, out.Status)
	require.Contains(t, result.IssueCodes, entity.IssueBMNotFound)
}
