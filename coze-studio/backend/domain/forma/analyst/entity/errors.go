/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package entity

import "errors"

var (
	ErrSessionNotFound           = errors.New("analyst session not found")
	ErrSessionClosed             = errors.New("analyst session closed")
	ErrInvalidTurn               = errors.New("invalid analyst turn")
	ErrModelFailed               = errors.New("analyst model failed")
	ErrInvalidExtraction         = errors.New("analyst invalid extraction")
	ErrAssertionNotFound         = errors.New("assertion not found")
	ErrAssertionAlreadyDecided   = errors.New("assertion already decided")
	ErrAssertionEvidenceRequired = errors.New("assertion evidence required")
	ErrProposalNotFound          = errors.New("proposal not found")
	ErrProposalStale             = errors.New("proposal stale")
	ErrProposalInvalid           = errors.New("proposal invalid")
	ErrProposalAlreadyApplied    = errors.New("proposal already applied")
	ErrForbidden                 = errors.New("analyst forbidden")
	ErrNotFound                  = errors.New("analyst not found")
	ErrNotConfigured             = errors.New("analyst service not configured")
	ErrGapNotFound               = errors.New("analyst gap not found")
)
