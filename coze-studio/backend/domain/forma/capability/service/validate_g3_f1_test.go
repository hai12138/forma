/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/fixture"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/repository"
)

func TestG3F1FAILPersistedReadableAndDraft(t *testing.T) {
	svc, repo, _, bm, contracts := newTestService(nil)
	payload := fixture.LaboratoryFlowCapability()
	payload.DataContractBindings[0].DataContractRevisionID = "dcr_wrong"
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: payload,
	})
	require.NoError(t, err)
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
	require.Equal(t, entity.RevisionDraft, out.Status)
	require.Equal(t, entity.ValidationFail, result.Status)

	list, err := repo.ListValidationsByRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, result.ValidationID, list[0].ValidationID)
	got, err := repo.GetValidationByEvidence(context.Background(), "t1", rev.RevisionID, result.EvidenceDigest)
	require.NoError(t, err)
	require.Equal(t, entity.ValidationFail, got.Status)
	gotRev, err := repo.GetRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Equal(t, entity.RevisionDraft, gotRev.Status)
}

func TestG3F1FAILReplaySameValidationID(t *testing.T) {
	svc, repo, _, bm, contracts := newTestService(nil)
	payload := fixture.LaboratoryFlowCapability()
	payload.DataContractBindings[0].DataContractRevisionID = "dcr_wrong"
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: payload,
	})
	require.NoError(t, err)
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
	_, r1, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrValidationFailed)
	_, r2, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrValidationFailed)
	require.Equal(t, r1.ValidationID, r2.ValidationID)
	list, err := repo.ListValidationsByRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Len(t, list, 1)
}

func TestG3F1ConcurrentFAILReplaySameValidationID(t *testing.T) {
	svc, repo, _, bm, contracts := newTestService(nil)
	payload := fixture.LaboratoryFlowCapability()
	payload.DataContractBindings[0].DataContractRevisionID = "dcr_wrong"
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: payload,
	})
	require.NoError(t, err)
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
	var wg sync.WaitGroup
	ids := make(chan string, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, result, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
			require.ErrorIs(t, err, entity.ErrValidationFailed)
			require.NotNil(t, result)
			ids <- result.ValidationID
		}()
	}
	wg.Wait()
	close(ids)
	seen := map[string]struct{}{}
	for id := range ids {
		seen[id] = struct{}{}
	}
	require.Len(t, seen, 1)
	list, err := repo.ListValidationsByRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Len(t, list, 1)
}

func TestG3F1MultiContractFieldSplitPASS(t *testing.T) {
	svc, _, _, bm, contracts := newTestService(nil)
	payload := fixture.LaboratoryFlowCapability()
	payload.DataContractBindings = []entity.DataContractBinding{
		{
			DataContractID: "dc_lab_in", DataContractRevisionID: "dcr_in_1", DataContractVersion: 1,
			LogicalFieldMappings: []entity.LogicalFieldMapping{
				{CapabilityLogicalKey: "work_cell_id", ContractLogicalKey: "work_cell_id"},
			},
		},
		{
			DataContractID: "dc_lab_out", DataContractRevisionID: "dcr_out_1", DataContractVersion: 1,
			LogicalFieldMappings: []entity.LogicalFieldMapping{
				{CapabilityLogicalKey: "device_id", ContractLogicalKey: "device_id"},
				{CapabilityLogicalKey: "device_status", ContractLogicalKey: "device_status"},
			},
		},
	}
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, CapabilityID: "cap-split", Payload: payload,
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev)
	out, result, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.NoError(t, err)
	require.Equal(t, entity.ValidationPass, result.Status)
	require.Equal(t, entity.RevisionValidated, out.Status)
}

func TestG3F1CrossBindingDuplicateCapKeyFAIL(t *testing.T) {
	svc, _, _, bm, contracts := newTestService(nil)
	payload := fixture.LaboratoryFlowCapability()
	payload.DataContractBindings = []entity.DataContractBinding{
		{
			DataContractID: "dc_a", DataContractRevisionID: "dcr_a_1", DataContractVersion: 1,
			LogicalFieldMappings: []entity.LogicalFieldMapping{
				{CapabilityLogicalKey: "work_cell_id", ContractLogicalKey: "work_cell_id"},
				{CapabilityLogicalKey: "device_id", ContractLogicalKey: "device_id"},
			},
		},
		{
			DataContractID: "dc_b", DataContractRevisionID: "dcr_b_1", DataContractVersion: 1,
			LogicalFieldMappings: []entity.LogicalFieldMapping{
				{CapabilityLogicalKey: "work_cell_id", ContractLogicalKey: "work_cell_id"},
				{CapabilityLogicalKey: "device_status", ContractLogicalKey: "device_status"},
			},
		},
	}
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, CapabilityID: "cap-dup", Payload: payload,
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev)
	_, result, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrValidationFailed)
	require.Contains(t, result.IssueCodes, entity.IssueMappingDuplicateCapKey)
}

func TestG3F1AIProvenanceWrongTargetFAIL(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, repo, _, bm, contracts := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz-lab", BusinessModelRevision: 1, ClientRequestID: "g3f1-ai-tgt", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	rev, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: testActor, CapabilityID: "cap-ai-tgt",
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
	confirm.TargetRevisionID = "crev_other"
	require.NoError(t, repository.ReplaceDecisionForTest(repo, confirm))

	seedPortsForRevision(bm, contracts, rev)
	_, _, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrMissingProvenance)
}

func TestG3F1AIProvenanceWrongProposalCapabilityFAIL(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, repo, _, bm, contracts := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz-lab", BusinessModelRevision: 1, ClientRequestID: "g3f1-ai-cap", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	rev, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: testActor, CapabilityID: "cap-ai-cap",
	})
	require.NoError(t, err)
	prop, err := repo.GetProposal(context.Background(), "t1", rev.ProposalID)
	require.NoError(t, err)
	prop.CapabilityID = "cap_other"
	require.NoError(t, repository.ReplaceProposalForTest(repo, prop))
	seedPortsForRevision(bm, contracts, rev)
	_, _, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrMissingProvenance)
}

func TestG3F1AIProvenanceWrongAnalysisRunFAIL(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, repo, _, bm, contracts := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz-lab", BusinessModelRevision: 1, ClientRequestID: "g3f1-ai-run", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	rev, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: res.Proposals[0].ProposalID, ActorID: testActor, CapabilityID: "cap-ai-run",
	})
	require.NoError(t, err)
	prop, err := repo.GetProposal(context.Background(), "t1", rev.ProposalID)
	require.NoError(t, err)
	prop.AnalysisRunID = "carun_other"
	require.NoError(t, repository.ReplaceProposalForTest(repo, prop))
	seedPortsForRevision(bm, contracts, rev)
	_, _, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrMissingProvenance)
}

func TestG3F1DerivedWrongSourceTargetFAIL(t *testing.T) {
	svc, repo, _, bm, contracts := newTestService(nil)
	_, rev1, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	payload2 := fixture.LaboratoryFlowCapability()
	payload2.Name = "ListLaboratoryDevices-v2"
	rev2, _, err := svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: rev1.CapabilityID, SourceRevisionID: rev1.RevisionID,
		ClientRequestID: "derive-g3f1", ActorID: testActor, Action: entity.DecisionDerive, Payload: payload2,
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
	derive.SourceRevisionID = "crev_wrong_src"
	require.NoError(t, repository.ReplaceDecisionForTest(repo, derive))
	seedPortsForRevision(bm, contracts, rev2)
	_, _, err = svc.Validate(context.Background(), "t1", rev2.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrMissingProvenance)

	derive.SourceRevisionID = rev1.RevisionID
	derive.TargetRevisionID = "crev_wrong_tgt"
	require.NoError(t, repository.ReplaceDecisionForTest(repo, derive))
	_, _, err = svc.Validate(context.Background(), "t1", rev2.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrMissingProvenance)
}

func TestG3F1DescriptorDriftValidateActivateFailClosed(t *testing.T) {
	svc, _, _, bm, contracts := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev)
	_, _, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.NoError(t, err)

	desc, err := contracts.GetActiveContractLogicalDescriptor(context.Background(), "t1", "biz-lab", "dc_lab")
	require.NoError(t, err)
	desc.Classification = map[string]string{"device_id": "INTERNAL"}
	contracts.Put(desc)

	_, _, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrMissingValidationEvidence)
	_, err = svc.Activate(context.Background(), "t1", rev.RevisionID, testActor, "drift")
	require.ErrorIs(t, err, entity.ErrMissingValidationEvidence)
}

func TestG3F1OptionalInputUnsupportedFILTEROperatorFAIL(t *testing.T) {
	svc, _, _, bm, contracts := newTestService(nil)
	payload := fixture.ProcurementQueryCapability()
	payload.InputSchema.Fields = append(payload.InputSchema.Fields, entity.LogicalField{
		LogicalKey: "optional_tag", LogicalType: "STRING", Required: false,
	})
	payload.Preconditions = append(payload.Preconditions, entity.Precondition{
		ID: "pc_opt", Predicate: entity.PredicateNotIn, LogicalKey: "optional_tag", Comparand: []string{"x"},
	})
	payload.DataContractBindings[0].LogicalFieldMappings = append(
		payload.DataContractBindings[0].LogicalFieldMappings,
		entity.LogicalFieldMapping{CapabilityLogicalKey: "optional_tag", ContractLogicalKey: "optional_tag"},
	)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-proc", ActorID: testActor, CapabilityID: "cap-opt-filter", Payload: payload,
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev)
	_, result, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.ErrorIs(t, err, entity.ErrValidationFailed)
	require.Contains(t, result.IssueCodes, entity.IssueQueryFilterOperator)
}

func TestG3F1SortSchemaClassificationDigestBlocksActivate(t *testing.T) {
	svc, repo, _, bm, contracts := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev)
	_, result, err := svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.NoError(t, err)
	require.Equal(t, entity.ValidationPass, result.Status)

	d1 := result.ContractEvidenceDigest
	desc, err := contracts.GetActiveContractLogicalDescriptor(context.Background(), "t1", "biz-lab", "dc_lab")
	require.NoError(t, err)
	desc.SortSchema = []ContractSortFieldSpec{{LogicalKey: "device_id", Directions: []string{"ASC"}}}
	contracts.Put(desc)
	d2, err := ContractEvidenceDigest([]*ContractLogicalDescriptor{desc})
	require.NoError(t, err)
	require.NotEqual(t, d1, d2)

	_, err = svc.Activate(context.Background(), "t1", rev.RevisionID, testActor, "sort-drift")
	require.ErrorIs(t, err, entity.ErrMissingValidationEvidence)

	forceStatusForTest(t, repo, "t1", rev.RevisionID, entity.RevisionValidated, entity.RevisionDraft)
	seedPortsForRevision(bm, contracts, rev)
	_, result, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.NoError(t, err)
	desc, err = contracts.GetActiveContractLogicalDescriptor(context.Background(), "t1", "biz-lab", "dc_lab")
	require.NoError(t, err)
	desc.Classification = map[string]string{"work_cell_id": "PII"}
	contracts.Put(desc)
	_, err = svc.Activate(context.Background(), "t1", rev.RevisionID, testActor, "class-drift")
	require.ErrorIs(t, err, entity.ErrMissingValidationEvidence)
}

func TestG3F1CreateValidationFailInjectRollsBack(t *testing.T) {
	uow := NewMemoryUnitOfWork()
	bm := NewFakeBusinessModelPort()
	contracts := NewFakeContractPort()
	svc := NewCapabilityService(&Components{UoW: uow, Business: bm, Contract: contracts})
	payload := fixture.LaboratoryFlowCapability()
	payload.DataContractBindings[0].DataContractRevisionID = "dcr_wrong"
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, Payload: payload,
	})
	require.NoError(t, err)
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
	uow.SetFailCreateValidation(true)
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
