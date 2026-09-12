/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"errors"
	"strings"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"gorm.io/gorm"
)

// MapRepoError wraps raw DB/driver errors to stable domain errors.
// Known sentinels are returned bare (never wrapped with driver text).
func MapRepoError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, entity.ErrConsistency):
		return entity.ErrConsistency
	case errors.Is(err, entity.ErrConflict):
		return entity.ErrConflict
	case errors.Is(err, entity.ErrIdempotencyConflict):
		return entity.ErrIdempotencyConflict
	case errors.Is(err, entity.ErrNotFound):
		return entity.ErrNotFound
	case errors.Is(err, entity.ErrRevisionNotFound):
		return entity.ErrRevisionNotFound
	case errors.Is(err, entity.ErrProposalNotFound):
		return entity.ErrProposalNotFound
	case errors.Is(err, entity.ErrAnalysisNotFound):
		return entity.ErrAnalysisNotFound
	case errors.Is(err, entity.ErrDecisionNotFound):
		return entity.ErrDecisionNotFound
	case errors.Is(err, entity.ErrIllegalTransition):
		return entity.ErrIllegalTransition
	case errors.Is(err, entity.ErrCrossTenant):
		return entity.ErrCrossTenant
	case errors.Is(err, entity.ErrStaleGeneration):
		return entity.ErrStaleGeneration
	case errors.Is(err, entity.ErrAnalysisNotFailed):
		return entity.ErrAnalysisNotFailed
	case errors.Is(err, entity.ErrMissingProvenance):
		return entity.ErrMissingProvenance
	case errors.Is(err, entity.ErrForbidden):
		return entity.ErrForbidden
	case errors.Is(err, entity.ErrInvalidPayload):
		return entity.ErrInvalidPayload
	case errors.Is(err, entity.ErrNotConfigured):
		return entity.ErrNotConfigured
	case errors.Is(err, entity.ErrActiveConflict):
		return entity.ErrActiveConflict
	case errors.Is(err, entity.ErrRevisionImmutable):
		return entity.ErrRevisionImmutable
	case errors.Is(err, entity.ErrConfirmRequired):
		return entity.ErrConfirmRequired
	case errors.Is(err, entity.ErrInvalidState):
		return entity.ErrInvalidState
	case errors.Is(err, entity.ErrUoWNotConfigured):
		return entity.ErrUoWNotConfigured
	case errors.Is(err, entity.ErrUoWCommitFailed):
		return entity.ErrUoWCommitFailed
	case errors.Is(err, entity.ErrMissingValidationEvidence):
		return entity.ErrMissingValidationEvidence
	case errors.Is(err, entity.ErrMissingImpactEvidence):
		return entity.ErrMissingImpactEvidence
	case errors.Is(err, entity.ErrAnalysisFailed):
		return entity.ErrAnalysisFailed
	case errors.Is(err, gorm.ErrRecordNotFound):
		return entity.ErrNotFound
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "unique_violation") || strings.Contains(msg, "1062") {
		return entity.ErrConflict
	}
	return entity.ErrConsistency
}
