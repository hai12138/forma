/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	assetentity "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
)

func baseRev(id string, version int32, status entity.RevisionStatus, name string) *entity.BusinessCapabilityRevision {
	return &entity.BusinessCapabilityRevision{
		RevisionID: id, CapabilityID: "cap1", TenantID: "t1", BusinessID: "b1",
		Version: version, Status: status, Name: name, Description: "d",
		BusinessModelRevision: 1, CapabilityKind: entity.KindQuery,
		QueryOperation: entity.QueryOpList, OutputCardinality: entity.CardinalityMany,
		Source: entity.SourceManualCreated,
	}
}

func TestProjectCapabilityAssetRefPriorities(t *testing.T) {
	cap := &entity.BusinessCapability{CapabilityID: "cap1", TenantID: "t1"}

	t.Run("STALE+DEPRECATED projects max STALE IN_REVIEW", func(t *testing.T) {
		revs := []*entity.BusinessCapabilityRevision{
			baseRev("r1", 1, entity.RevisionDeprecated, "old"),
			baseRev("r2", 3, entity.RevisionStale, "stale-high"),
			baseRev("r3", 2, entity.RevisionStale, "stale-low"),
		}
		proj, err := ProjectCapabilityAssetRef(cap, revs)
		require.NoError(t, err)
		require.Equal(t, "stale-high", proj.Name)
		require.Equal(t, "0.3.0", proj.SemanticVersion)
		require.Equal(t, assetentity.AssetStatusInReview, proj.Status)
	})

	t.Run("ACTIVE RELEASED", func(t *testing.T) {
		cap2 := &entity.BusinessCapability{CapabilityID: "cap1", TenantID: "t1", ActiveRevisionID: "ra"}
		revs := []*entity.BusinessCapabilityRevision{
			baseRev("ra", 2, entity.RevisionActive, "active"),
			baseRev("rd", 5, entity.RevisionDraft, "draft"),
		}
		proj, err := ProjectCapabilityAssetRef(cap2, revs)
		require.NoError(t, err)
		require.Equal(t, "active", proj.Name)
		require.Equal(t, "0.2.0", proj.SemanticVersion)
		require.Equal(t, assetentity.AssetStatusReleased, proj.Status)
	})

	t.Run("empty revisions consistency", func(t *testing.T) {
		_, err := ProjectCapabilityAssetRef(cap, nil)
		require.ErrorIs(t, err, entity.ErrConsistency)
	})

	t.Run("duplicate version consistency", func(t *testing.T) {
		revs := []*entity.BusinessCapabilityRevision{
			baseRev("r1", 1, entity.RevisionDraft, "a"),
			baseRev("r2", 1, entity.RevisionDraft, "b"),
		}
		_, err := ProjectCapabilityAssetRef(cap, revs)
		require.ErrorIs(t, err, entity.ErrConsistency)
	})

	t.Run("orphan ACTIVE with null pointer", func(t *testing.T) {
		revs := []*entity.BusinessCapabilityRevision{
			baseRev("r1", 1, entity.RevisionActive, "a"),
		}
		_, err := ProjectCapabilityAssetRef(cap, revs)
		require.ErrorIs(t, err, entity.ErrConsistency)
	})

	t.Run("pointer + second ACTIVE", func(t *testing.T) {
		cap2 := &entity.BusinessCapability{CapabilityID: "cap1", TenantID: "t1", ActiveRevisionID: "ra"}
		revs := []*entity.BusinessCapabilityRevision{
			baseRev("ra", 1, entity.RevisionActive, "a"),
			baseRev("rb", 2, entity.RevisionActive, "b"),
		}
		_, err := ProjectCapabilityAssetRef(cap2, revs)
		require.ErrorIs(t, err, entity.ErrConsistency)
	})

	t.Run("pointer A while ActiveSet is B", func(t *testing.T) {
		cap2 := &entity.BusinessCapability{CapabilityID: "cap1", TenantID: "t1", ActiveRevisionID: "ra"}
		revs := []*entity.BusinessCapabilityRevision{
			baseRev("rb", 1, entity.RevisionActive, "b"),
			baseRev("ra", 2, entity.RevisionDraft, "a"),
		}
		_, err := ProjectCapabilityAssetRef(cap2, revs)
		require.ErrorIs(t, err, entity.ErrConsistency)
	})

	t.Run("dangling pointer", func(t *testing.T) {
		cap2 := &entity.BusinessCapability{CapabilityID: "cap1", TenantID: "t1", ActiveRevisionID: "missing"}
		revs := []*entity.BusinessCapabilityRevision{
			baseRev("r1", 1, entity.RevisionDraft, "a"),
		}
		_, err := ProjectCapabilityAssetRef(cap2, revs)
		require.ErrorIs(t, err, entity.ErrConsistency)
	})
}

func TestAllowTransition(t *testing.T) {
	require.True(t, AllowTransition(entity.RevisionDraft, entity.RevisionValidated))
	require.True(t, AllowTransition(entity.RevisionValidated, entity.RevisionActive))
	require.True(t, AllowTransition(entity.RevisionActive, entity.RevisionDeprecated))
	require.True(t, AllowTransition(entity.RevisionActive, entity.RevisionStale))
	require.True(t, AllowTransition(entity.RevisionStale, entity.RevisionDeprecated))
	require.False(t, AllowTransition(entity.RevisionDraft, entity.RevisionActive))
	require.False(t, AllowTransition(entity.RevisionValidated, entity.RevisionDeprecated))
	require.False(t, AllowTransition(entity.RevisionStale, entity.RevisionActive))
	require.False(t, AllowTransition(entity.RevisionDeprecated, entity.RevisionActive))
}

func TestContentDigestProvenanceNeutral(t *testing.T) {
	payload := entity.SemanticPayload{
		Name: "X", Description: "Y", CapabilityKind: entity.KindQuery, BusinessModelRevision: 1,
		InputSchema: entity.LogicalSchema{Fields: []entity.LogicalField{{LogicalKey: "a", LogicalType: "STRING"}}},
		OutputSchema: entity.LogicalSchema{Fields: []entity.LogicalField{{LogicalKey: "b", LogicalType: "STRING"}}},
		Preconditions: []entity.Precondition{{ID: "p2", Predicate: "EQ"}, {ID: "p1", Predicate: "EXISTS"}},
		Effects: []entity.Effect{{ID: "e1", Kind: "READ"}},
		DataContractBindings: []entity.DataContractBinding{
			{DataContractID: "dc2", DataContractRevisionID: "r2"},
			{DataContractID: "dc1", DataContractRevisionID: "r1"},
		},
		QueryOperation: entity.QueryOpRead, OutputCardinality: entity.CardinalityOne,
	}
	d1, err := CapabilityContentDigest(payload)
	require.NoError(t, err)

	revManual := &entity.BusinessCapabilityRevision{
		RevisionID: "r1", Version: 1, Status: entity.RevisionDraft, Source: entity.SourceManualCreated,
		Name: payload.Name, Description: payload.Description, CapabilityKind: payload.CapabilityKind,
		BusinessModelRevision: payload.BusinessModelRevision, InputSchema: payload.InputSchema,
		OutputSchema: payload.OutputSchema, Preconditions: payload.Preconditions, Effects: payload.Effects,
		DataContractBindings: payload.DataContractBindings, QueryOperation: payload.QueryOperation,
		OutputCardinality: payload.OutputCardinality, CreatedBy: "a",
	}
	revAI := *revManual
	revAI.Source = entity.SourceAIProposal
	revAI.ProposalID = "prop"
	revAI.AnalysisRunID = "run"
	revAI.RevisionID = "r2"
	revAI.Version = 9

	d2, err := CapabilityContentDigest(revManual.ToSemanticPayload())
	require.NoError(t, err)
	d3, err := CapabilityContentDigest(revAI.ToSemanticPayload())
	require.NoError(t, err)
	require.Equal(t, d1, d2)
	require.Equal(t, d1, d3)

	// set-like ordering of preconditions must not affect digest
	payload2 := payload
	payload2.Preconditions = []entity.Precondition{{ID: "p1", Predicate: "EXISTS"}, {ID: "p2", Predicate: "EQ"}}
	d4, err := CapabilityContentDigest(payload2)
	require.NoError(t, err)
	require.Equal(t, d1, d4)
}

func TestAnalysisRequestDigestStable(t *testing.T) {
	req := entity.AnalysisRequest{
		BusinessModelRevision: 1,
		DataContractPins: []entity.DataContractPin{
			{DataContractID: "b", DataContractVersion: 2},
			{DataContractID: "a", DataContractVersion: 1},
		},
		RequirementRefs: []string{"r2", "r1"},
		Options:         map[string]any{"z": 1, "a": "x"},
	}
	d1, err := AnalysisRequestDigest(req)
	require.NoError(t, err)
	req2 := entity.AnalysisRequest{
		BusinessModelRevision: 1,
		DataContractPins: []entity.DataContractPin{
			{DataContractID: "a", DataContractVersion: 1},
			{DataContractID: "b", DataContractVersion: 2},
		},
		RequirementRefs: []string{"r1", "r2"},
		Options:         map[string]any{"a": "x", "z": 1},
	}
	d2, err := AnalysisRequestDigest(req2)
	require.NoError(t, err)
	require.Equal(t, d1, d2)
}
