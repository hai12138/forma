/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package entity

import "errors"

var (
	ErrRevisionConflict = errors.New("revision conflict")
	ErrNotFound         = errors.New("not found")
)
