/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package entity

import "errors"

var (
	ErrConsistency           = errors.New("forma capability consistency error")
	ErrConflict              = errors.New("forma capability conflict")
	ErrIdempotencyConflict   = errors.New("forma capability idempotency conflict")
	ErrNotFound              = errors.New("forma capability not found")
	ErrRevisionNotFound      = errors.New("forma capability revision not found")
	ErrProposalNotFound      = errors.New("forma capability proposal not found")
	ErrAnalysisNotFound      = errors.New("forma capability analysis run not found")
	ErrDecisionNotFound      = errors.New("forma capability decision not found")
	ErrIllegalTransition     = errors.New("forma capability illegal transition")
	ErrCrossTenant           = errors.New("forma capability cross-tenant reference denied")
	ErrStaleGeneration       = errors.New("forma capability stale analysis generation")
	ErrAnalysisNotFailed     = errors.New("forma capability analysis is not failed")
	ErrMissingProvenance     = errors.New("forma capability missing decision provenance")
	ErrForbidden             = errors.New("forma capability forbidden")
	ErrInvalidPayload        = errors.New("forma capability invalid payload")
	ErrNotConfigured         = errors.New("forma capability service not configured")
	ErrActiveConflict        = errors.New("forma capability active revision conflict")
	ErrRevisionImmutable     = errors.New("forma capability revision immutable")
	ErrConfirmRequired            = errors.New("forma capability confirm required")
	ErrInvalidState               = errors.New("forma capability invalid state")
	ErrUoWNotConfigured           = errors.New("forma capability unit of work not configured")
	ErrMissingValidationEvidence  = errors.New("forma capability missing validation evidence")
	ErrMissingImpactEvidence      = errors.New("forma capability missing impact evidence")
	ErrAnalysisFailed             = errors.New("forma capability analysis failed")
)
