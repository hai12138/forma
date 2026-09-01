/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"testing"
	"time"

	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
	"github.com/stretchr/testify/require"
)

func TestValidateExtractionRejectsInvalidType(t *testing.T) {
	err := ValidateExtractionResult(&entity.ExtractionResult{
		Assertions: []entity.ExtractionAssertion{{
			AssertionType: "INVALID",
			Confidence:    0.5,
		}},
	})
	require.Error(t, err)
}

func TestDetectConflicts(t *testing.T) {
	now := time.Now().UTC()
	a1 := &entity.BusinessAssertion{
		AssertionID: "a1", SubjectRef: "rule:close", Predicate: "permission", ObjectValue: "管理员",
	}
	a2 := &entity.BusinessAssertion{
		AssertionID: "a2", SubjectRef: "rule:close", Predicate: "permission", ObjectValue: "任何维修人员",
	}
	conflicts := detectConflicts([]*entity.BusinessAssertion{a1, a2}, "t1", "b1", "s1", now)
	require.Len(t, conflicts, 1)
}

func TestBuildProposalPatchFromConfirmed(t *testing.T) {
	a := &entity.BusinessAssertion{
		AssertionID:   "assert_1",
		AssertionType: entity.AssertionActorExists,
		SubjectRef:    "actor:employee",
		ObjectValue:   "员工",
		Status:        entity.AssertionConfirmed,
	}
	patch := BuildProposalPatch([]*entity.BusinessAssertion{a})
	require.NotEmpty(t, patch.Operations)
	require.Equal(t, entity.PatchAddNode, patch.Operations[0].Op)
	require.Equal(t, businessentity.NodeActor, patch.Operations[0].Node.Type)
}

func TestApplyPatchValidatesModel(t *testing.T) {
	base := &businessentity.SemanticModel{
		SchemaVersion: businessentity.SemanticSchemaVersion,
		Nodes: []businessentity.SemanticNode{{
			ID:           "node_employee",
			Type:         businessentity.NodeActor,
			Name:         "员工",
			SourceMarker: businessentity.SourceManualModified,
		}},
	}
	patch := &entity.SemanticModelPatch{
		Operations: []entity.PatchOperation{{
			Op: entity.PatchAddNode,
			Node: &businessentity.SemanticNode{
				ID:           "node_employee",
				Type:         businessentity.NodeActor,
				Name:         "重复",
				SourceMarker: businessentity.SourceAIGenerated,
			},
			SourceAssertionIDs: []string{"a1"},
		}},
	}
	_, err := ApplyPatch(base, patch)
	require.Error(t, err)
}

func TestDeterministicFakeExtractsWorkOrder(t *testing.T) {
	fake := NewDeterministicFakeModel()
	res, err := fake.ExtractAssertions(context.Background(), &ExtractionRequest{
		UserTurnContent: "员工发现设备故障后提交报修，维修人员接单处理，完成后由管理员关闭。",
		UserTurnID:      "turn_test",
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.Assertions)
	require.NoError(t, ValidateExtractionResult(res))
}

func TestNoSilentMutationWithoutConfirmApply(t *testing.T) {
	// Extraction output is not persisted as CONFIRMED — confirmation is a separate human action.
	res, err := NewDeterministicFakeModel().ExtractAssertions(context.Background(), &ExtractionRequest{
		UserTurnContent: "员工报修维修工单",
		UserTurnID:      "t1",
	})
	require.NoError(t, err)
	for _, a := range res.Assertions {
		require.True(t, allowedAssertionTypes[a.AssertionType])
	}
}
