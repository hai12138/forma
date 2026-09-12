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
// Never returns driver strings to callers from the service layer.
func MapRepoError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, entity.ErrConsistency),
		errors.Is(err, entity.ErrConflict),
		errors.Is(err, entity.ErrIdempotencyConflict),
		errors.Is(err, entity.ErrNotFound),
		errors.Is(err, entity.ErrRevisionNotFound),
		errors.Is(err, entity.ErrProposalNotFound),
		errors.Is(err, entity.ErrAnalysisNotFound),
		errors.Is(err, entity.ErrDecisionNotFound),
		errors.Is(err, entity.ErrIllegalTransition),
		errors.Is(err, entity.ErrCrossTenant),
		errors.Is(err, entity.ErrStaleGeneration),
		errors.Is(err, entity.ErrAnalysisNotFailed),
		errors.Is(err, entity.ErrMissingProvenance),
		errors.Is(err, entity.ErrForbidden),
		errors.Is(err, entity.ErrInvalidPayload),
		errors.Is(err, entity.ErrNotConfigured),
		errors.Is(err, entity.ErrActiveConflict),
		errors.Is(err, entity.ErrRevisionImmutable),
		errors.Is(err, entity.ErrConfirmRequired),
		errors.Is(err, entity.ErrInvalidState),
		errors.Is(err, entity.ErrUoWNotConfigured),
		errors.Is(err, entity.ErrMissingValidationEvidence),
		errors.Is(err, entity.ErrMissingImpactEvidence),
		errors.Is(err, entity.ErrAnalysisFailed):
		return err
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
