/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package entity

import "errors"

var (
	ErrInvalidModel       = errors.New("invalid business model")
	ErrNotFound           = errors.New("business not found")
	ErrRevisionNotFound   = errors.New("business revision not found")
	ErrRevisionConflict   = errors.New("business model revision conflict")
	ErrLayoutConflict     = errors.New("business layout revision conflict")
	ErrNoChange           = errors.New("business model no change")
	ErrInvalidRelation    = errors.New("invalid business relation")
)
