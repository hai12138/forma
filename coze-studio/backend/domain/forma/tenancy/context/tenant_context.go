/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package tenantctx

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
)

type TenantContext struct {
	TenantID        string
	PrincipalID     string
	MembershipRole  entity.MembershipRole
	CozeUserID      int64
	AllowedSpaceIDs []int64
	RequestID       string
	TenantStatus    entity.TenantStatus
}

type ctxKey struct{}

func WithTenantContext(ctx context.Context, tc *TenantContext) context.Context {
	return context.WithValue(ctx, ctxKey{}, tc)
}

func FromContext(ctx context.Context) (*TenantContext, bool) {
	tc, ok := ctx.Value(ctxKey{}).(*TenantContext)
	return tc, ok && tc != nil
}
