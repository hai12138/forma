/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package repository

import (
	"context"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
)

// SetLeaseExpiresAtForTest adjusts lease expiry in memory repositories (tests only).
func SetLeaseExpiresAtForTest(repo DataRepository, ctx context.Context, tenantID, analysisRunID string, expiresAt time.Time) error {
	type leaseSetter interface {
		SetLeaseExpiresAtForTest(context.Context, string, string, time.Time) error
	}
	s, ok := repo.(leaseSetter)
	if !ok {
		return entity.ErrNotConfigured
	}
	return s.SetLeaseExpiresAtForTest(ctx, tenantID, analysisRunID, expiresAt)
}

// LookupAnalysisRunByClientRequest returns the run for an idempotency key (tests only).
func LookupAnalysisRunByClientRequest(ctx context.Context, repo DataRepository, tenantID, businessID string, revision int32, clientRequestID string) (*entity.DataRequirementAnalysisRun, error) {
	type idempotencyLookup interface {
		GetAnalysisRunByIdempotencyKey(context.Context, string, string, int32, string) (*entity.DataRequirementAnalysisRun, error)
	}
	l, ok := repo.(idempotencyLookup)
	if !ok {
		return nil, entity.ErrNotConfigured
	}
	return l.GetAnalysisRunByIdempotencyKey(ctx, tenantID, businessID, revision, clientRequestID)
}
