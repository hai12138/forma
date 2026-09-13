/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"strings"
	"unicode/utf8"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
)

const (
	maxAuditReasonLen          = 1024
	maxAuditClientRequestIDLen = 128
)

var auditCredentialSubstrings = []string{
	"password",
	"token",
	"cookie",
	"authorization",
	"bearer",
	"jwt",
	"api_key",
	"api-key",
	"private_key",
	"private-key",
	"secret",
	"-----begin",
}

// ValidateAuditMetadata rejects unsafe reason / client_request_id values.
// Empty strings are allowed. Failures return entity.ErrInvalidPayload (reject, never redact).
func ValidateAuditMetadata(reason, clientRequestID string) error {
	if err := validateAuditField(reason, maxAuditReasonLen); err != nil {
		return err
	}
	return validateAuditField(clientRequestID, maxAuditClientRequestIDLen)
}

func validateAuditField(value string, maxLen int) error {
	if value == "" {
		return nil
	}
	if !utf8.ValidString(value) {
		return entity.ErrInvalidPayload
	}
	if utf8.RuneCountInString(value) > maxLen {
		return entity.ErrInvalidPayload
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return entity.ErrInvalidPayload
		}
	}
	lower := strings.ToLower(value)
	for _, sub := range auditCredentialSubstrings {
		if strings.Contains(lower, sub) {
			return entity.ErrInvalidPayload
		}
	}
	return nil
}
