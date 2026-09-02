/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package entity

import "errors"

var (
	ErrRequirementNotFound       = errors.New("data requirement not found")
	ErrRequirementAlreadyDecided = errors.New("data requirement already decided")
	ErrRequirementInvalidState   = errors.New("data requirement invalid state")
	ErrAnalysisNotFound          = errors.New("data analysis run not found")
	ErrAnalysisIdempotencyConflict = errors.New("data analysis idempotency conflict")
	ErrAnalysisNotFailed         = errors.New("data analysis run is not failed")
	ErrBusinessRevisionNotFound  = errors.New("data business revision not found")
	ErrBusinessElementRefInvalid = errors.New("data business element ref invalid")
	ErrInvalidProposal           = errors.New("data requirement proposal invalid")
	ErrForbidden                 = errors.New("data action forbidden")
	ErrNotConfigured             = errors.New("data service not configured")
	ErrModelFailed               = errors.New("data model failed")
)
