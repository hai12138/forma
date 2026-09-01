/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"
	"time"

	tenancyentity "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
)

type tenancyAuditHook struct {
	svc TenancyAuditRecorder
}

type TenancyAuditRecorder interface {
	RecordAudit(ctx context.Context, event *tenancyentity.AuditEvent) error
}

func NewTenancyAuditHook(rec TenancyAuditRecorder) *tenancyAuditHook {
	return &tenancyAuditHook{svc: rec}
}

func (h *tenancyAuditHook) RecordAnalystAudit(ctx context.Context, tenantID, actorID, action, resourceID, requestID string) error {
	if h == nil || h.svc == nil {
		return nil
	}
	return h.svc.RecordAudit(ctx, &tenancyentity.AuditEvent{
		TenantID:    tenantID,
		PrincipalID: actorID,
		Action:      action,
		Resource:    resourceID,
		RequestID:   requestID,
		CreatedAt:   time.Now().UTC(),
	})
}
