/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"unicode/utf8"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
)

const (
	maxAuditReasonLen          = 1024
	maxAuditClientRequestIDLen = 128
)

// ValidateAuditMetadata rejects unsafe reason / client_request_id values.
// Empty strings are allowed. Failures return entity.ErrInvalidPayload (reject, never redact).
//
// reason uses containsSecret (shared credential shapes + freeTextAssignmentPatterns), not bare keyword substrings.
// client_request_id (when non-empty) requires ValidateOpaqueID + containsCredentialShape + length ≤128
// (assignment forms cannot pass ValidateOpaqueID; keep this split — do not run freeTextAssignmentPatterns here).
func ValidateAuditMetadata(reason, clientRequestID string) error {
	if err := validateAuditReason(reason); err != nil {
		return err
	}
	return validateAuditClientRequestID(clientRequestID)
}

func validateAuditReason(reason string) error {
	if reason == "" {
		return nil
	}
	if !utf8.ValidString(reason) {
		return entity.ErrInvalidPayload
	}
	if utf8.RuneCountInString(reason) > maxAuditReasonLen {
		return entity.ErrInvalidPayload
	}
	for _, r := range reason {
		if r < 0x20 || r == 0x7f {
			return entity.ErrInvalidPayload
		}
	}
	if containsSecret(reason) {
		return entity.ErrInvalidPayload
	}
	return nil
}

func validateAuditClientRequestID(clientRequestID string) error {
	if clientRequestID == "" {
		return nil
	}
	if utf8.RuneCountInString(clientRequestID) > maxAuditClientRequestIDLen {
		return entity.ErrInvalidPayload
	}
	if err := ValidateOpaqueID(clientRequestID); err != nil {
		return entity.ErrInvalidPayload
	}
	if containsCredentialShape(clientRequestID) {
		return entity.ErrInvalidPayload
	}
	return nil
}
