/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	assetentity "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/fixture"
)

// TestG3F3ActivateListRevisionsRaceCaughtByFence proves Contract/BM drift injected during
// ListRevisions (after PASS check, before first status write) is caught by the final fence.
func TestG3F3ActivateListRevisionsRaceCaughtByFence(t *testing.T) {
	uow := NewMemoryUnitOfWork()
	bm := NewFakeBusinessModelPort()
	contracts := NewFakeContractPort()
	svc := NewCapabilityService(&Components{UoW: uow, Business: bm, Contract: contracts})

	_, rev1, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, OwnerID: testOwnerID, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev1)
	_, _, err = svc.Validate(context.Background(), "t1", rev1.RevisionID, testActor)
	require.NoError(t, err)
	active, err := svc.Activate(context.Background(), "t1", rev1.RevisionID, testActor, "first-active")
	require.NoError(t, err)
	require.Equal(t, entity.RevisionActive, active.Status)

	payload2 := fixture.LaboratoryFlowCapability()
	payload2.Name = "ListLaboratoryDevices-v2"
	rev2, _, err := svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: rev1.CapabilityID, SourceRevisionID: rev1.RevisionID,
		ClientRequestID: "derive-f3-race", ActorID: testActor, Action: entity.DecisionDerive, Payload: payload2,
	})
	require.NoError(t, err)
	seedPortsForRevision(bm, contracts, rev2)
	_, _, err = svc.Validate(context.Background(), "t1", rev2.RevisionID, testActor)
	require.NoError(t, err)

	capBefore, err := uow.Root().GetCapability(context.Background(), "t1", rev1.CapabilityID)
	require.NoError(t, err)
	assetBefore, err := uow.AssetsView().GetCapabilityAsset(context.Background(), "t1", rev1.CapabilityID)
	require.NoError(t, err)
	decsBefore, err := uow.Root().ListDecisionsByCapability(context.Background(), "t1", rev1.CapabilityID)
	require.NoError(t, err)
	activateCountBefore := 0
	for _, d := range decsBefore {
		if d.Action == entity.DecisionActivate {
			activateCountBefore++
		}
	}

	desc, err := contracts.GetActiveContractLogicalDescriptor(context.Background(), "t1", "biz-lab", "dc_lab")
	require.NoError(t, err)
	drifted := cloneContractDescriptor(desc)
	drifted.Classification = map[string]string{"device_id": "INTERNAL"}

	var flipped atomic.Bool
	uow.SetOnBeforeListRevisions(func() {
		if flipped.CompareAndSwap(false, true) {
			contracts.Put(drifted)
		}
	})

	_, err = svc.Activate(context.Background(), "t1", rev2.RevisionID, testActor, "race-activate")
	require.ErrorIs(t, err, entity.ErrConsistency)
	require.True(t, flipped.Load(), "ListRevisions hook must have fired")

	got1, err := uow.Root().GetRevision(context.Background(), "t1", rev1.RevisionID)
	require.NoError(t, err)
	require.Equal(t, entity.RevisionActive, got1.Status)
	got2, err := uow.Root().GetRevision(context.Background(), "t1", rev2.RevisionID)
	require.NoError(t, err)
	require.Equal(t, entity.RevisionValidated, got2.Status)

	capAfter, err := uow.Root().GetCapability(context.Background(), "t1", rev1.CapabilityID)
	require.NoError(t, err)
	require.Equal(t, capBefore.ActiveRevisionID, capAfter.ActiveRevisionID)
	require.Equal(t, rev1.RevisionID, capAfter.ActiveRevisionID)
	require.Equal(t, capBefore.AggregateGeneration, capAfter.AggregateGeneration)

	decsAfter, err := uow.Root().ListDecisionsByCapability(context.Background(), "t1", rev1.CapabilityID)
	require.NoError(t, err)
	require.Len(t, decsAfter, len(decsBefore))
	activateCountAfter := 0
	for _, d := range decsAfter {
		if d.Action == entity.DecisionActivate {
			activateCountAfter++
		}
	}
	require.Equal(t, activateCountBefore, activateCountAfter)

	assetAfter, err := uow.AssetsView().GetCapabilityAsset(context.Background(), "t1", rev1.CapabilityID)
	require.NoError(t, err)
	require.Equal(t, assetBefore.Status, assetAfter.Status)
	require.Equal(t, assetentity.AssetStatusReleased, assetAfter.Status)
	require.Equal(t, assetBefore.SemanticVersion, assetAfter.SemanticVersion)
	require.Equal(t, assetBefore.ContentDigest, assetAfter.ContentDigest)
}
