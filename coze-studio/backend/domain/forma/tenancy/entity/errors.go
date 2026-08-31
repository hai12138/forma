/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package entity

import "errors"

var (
	ErrRevisionConflict      = errors.New("revision conflict")
	ErrNotFound              = errors.New("not found")
	ErrLastOwner             = errors.New("cannot demote or remove the last OWNER")
	ErrInvalidRole           = errors.New("invalid membership role")
	ErrPrimaryOwnerImmutable = errors.New("primary owner role is immutable")
	ErrSpaceOwned            = errors.New("space is already bound to another tenant")
)
