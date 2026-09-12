/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import "github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"

// AllowTransition returns true only for legal revision status transitions (§8.1).
func AllowTransition(from, to entity.RevisionStatus) bool {
	switch from {
	case entity.RevisionDraft:
		return to == entity.RevisionValidated
	case entity.RevisionValidated:
		return to == entity.RevisionActive
	case entity.RevisionActive:
		return to == entity.RevisionDeprecated || to == entity.RevisionStale
	case entity.RevisionStale:
		return to == entity.RevisionDeprecated
	default:
		return false
	}
}
