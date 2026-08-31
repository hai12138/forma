/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/service"
)

func TestSpaceBindUnbindRebindCycle(t *testing.T) {
	svc, audit := newTestServiceWithAudit()
	ctx := context.Background()

	owner, err := svc.ResolveOrCreatePrincipal(ctx, 31, "OwnerA")
	require.NoError(t, err)
	tenant, err := svc.CreateTenant(ctx, &service.CreateTenantRequest{
		Name:             "TenantA",
		OwnerPrincipalID: owner.PrincipalID,
	})
	require.NoError(t, err)

	const spaceID int64 = 9001
	var lastID int64
	for i := 0; i < 5; i++ {
		ref, bindErr := svc.BindSpace(ctx, &service.BindSpaceRequest{
			TenantID:         tenant.TenantID,
			CozeSpaceID:      spaceID,
			Purpose:          entity.SpacePurposeDefault,
			ActorPrincipalID: owner.PrincipalID,
			RequestID:        "bind-cycle",
		})
		require.NoError(t, bindErr)
		require.Equal(t, entity.SpaceRefActive, ref.Status)
		if i == 0 {
			lastID = ref.ID
		} else {
			assert.Equal(t, lastID, ref.ID, "rebind must reuse the same row")
		}

		err = svc.UnbindSpace(ctx, tenant.TenantID, spaceID, owner.PrincipalID, "unbind-cycle")
		require.NoError(t, err)
	}

	bound := audit.byAction(entity.AuditSpaceBound)
	unbound := audit.byAction(entity.AuditSpaceUnbound)
	assert.Len(t, bound, 5)
	assert.Len(t, unbound, 5)
	for _, e := range bound {
		assert.Equal(t, owner.PrincipalID, e.PrincipalID)
	}
}

func TestSpaceCannotStealActiveBinding_ThenRebindAfterUnbind(t *testing.T) {
	svc, _ := newTestServiceWithAudit()
	ctx := context.Background()

	ownerA, err := svc.ResolveOrCreatePrincipal(ctx, 41, "OwnerA")
	require.NoError(t, err)
	ownerB, err := svc.ResolveOrCreatePrincipal(ctx, 42, "OwnerB")
	require.NoError(t, err)

	tenantA, err := svc.CreateTenant(ctx, &service.CreateTenantRequest{
		Name: "A", OwnerPrincipalID: ownerA.PrincipalID,
	})
	require.NoError(t, err)
	tenantB, err := svc.CreateTenant(ctx, &service.CreateTenantRequest{
		Name: "B", OwnerPrincipalID: ownerB.PrincipalID,
	})
	require.NoError(t, err)

	const spaceID int64 = 9100
	refA, err := svc.BindSpace(ctx, &service.BindSpaceRequest{
		TenantID: tenantA.TenantID, CozeSpaceID: spaceID,
		ActorPrincipalID: ownerA.PrincipalID,
	})
	require.NoError(t, err)
	require.Equal(t, entity.SpaceRefActive, refA.Status)

	_, err = svc.BindSpace(ctx, &service.BindSpaceRequest{
		TenantID: tenantB.TenantID, CozeSpaceID: spaceID,
		ActorPrincipalID: ownerB.PrincipalID,
	})
	require.ErrorIs(t, err, entity.ErrSpaceOwned)

	err = svc.UnbindSpace(ctx, tenantA.TenantID, spaceID, ownerA.PrincipalID, "req")
	require.NoError(t, err)

	refB, err := svc.BindSpace(ctx, &service.BindSpaceRequest{
		TenantID: tenantB.TenantID, CozeSpaceID: spaceID,
		Purpose: entity.SpacePurposeDelivery, ActorPrincipalID: ownerB.PrincipalID,
	})
	require.NoError(t, err)
	assert.Equal(t, tenantB.TenantID, refB.TenantID)
	assert.Equal(t, entity.SpaceRefActive, refB.Status)
	assert.Equal(t, refA.ID, refB.ID, "rebind after unbind must update same row")
	assert.Equal(t, entity.SpacePurposeDelivery, refB.Purpose)
}
