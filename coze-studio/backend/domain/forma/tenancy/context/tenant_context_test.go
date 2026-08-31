/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package tenantctx_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tenantctx "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/context"
	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
)

func TestTenantContext_WithAndFrom(t *testing.T) {
	tc := &tenantctx.TenantContext{
		TenantID:        "ten_1",
		PrincipalID:     "prin_1",
		MembershipRole:  entity.RoleOwner,
		CozeUserID:      99,
		AllowedSpaceIDs: []int64{1, 2},
		RequestID:       "req-1",
		TenantStatus:    entity.TenantStatusActive,
	}
	ctx := tenantctx.WithTenantContext(context.Background(), tc)
	got, ok := tenantctx.FromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, "ten_1", got.TenantID)
	assert.Equal(t, entity.RoleOwner, got.MembershipRole)

	_, ok = tenantctx.FromContext(context.Background())
	assert.False(t, ok)
}
