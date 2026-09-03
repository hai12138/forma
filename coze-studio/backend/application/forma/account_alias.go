/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"regexp"
	"strings"
	"unicode"

	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
)

// Forma Local Account Alias contract:
//
//	admin              → admin@forma.local
//	user01             → user01@forma.local
//	email@example.com  → email@example.com
//
// Local aliases and their @forma.local email form are the same identity.
// Creating both "user01" and "user01@forma.local" is rejected as a collision.
const FormaLocalEmailDomain = "forma.local"

var localAccountPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,62}$`)

// NormalizedAccount is the canonical login identity used by Coze email auth.
type NormalizedAccount struct {
	// Account is the product-facing account (local alias or full email).
	Account string
	// Email is the Coze User Domain login identifier.
	Email string
	// IsLocalAlias is true when the account maps through @forma.local.
	IsLocalAlias bool
}

// NormalizeAccount validates and normalizes a Forma login/create account.
func NormalizeAccount(raw string) (*NormalizedAccount, error) {
	account := strings.TrimSpace(raw)
	if account == "" {
		return nil, formaerrors.AdminBadRequest("account is required")
	}
	if strings.ContainsAny(account, " \t\r\n") {
		return nil, formaerrors.AdminBadRequest("account must not contain whitespace")
	}
	for _, r := range account {
		if unicode.IsControl(r) {
			return nil, formaerrors.AdminBadRequest("account must not contain control characters")
		}
	}
	if strings.ContainsAny(account, `/\`) {
		return nil, formaerrors.AdminBadRequest("account must not contain path characters")
	}

	at := strings.IndexByte(account, '@')
	if at < 0 {
		local := strings.ToLower(account)
		if !localAccountPattern.MatchString(local) {
			return nil, formaerrors.AdminBadRequest("invalid local account format")
		}
		return &NormalizedAccount{
			Account:      local,
			Email:        local + "@" + FormaLocalEmailDomain,
			IsLocalAlias: true,
		}, nil
	}

	if at == 0 || at == len(account)-1 {
		return nil, formaerrors.AdminBadRequest("invalid email account")
	}
	if strings.Count(account, "@") != 1 {
		return nil, formaerrors.AdminBadRequest("invalid email account")
	}

	email := strings.ToLower(account)
	local, domain, ok := strings.Cut(email, "@")
	if !ok || local == "" || domain == "" {
		return nil, formaerrors.AdminBadRequest("invalid email account")
	}
	if domain == FormaLocalEmailDomain {
		if !localAccountPattern.MatchString(local) {
			return nil, formaerrors.AdminBadRequest("invalid local account format")
		}
		return &NormalizedAccount{
			Account:      local,
			Email:        local + "@" + FormaLocalEmailDomain,
			IsLocalAlias: true,
		}, nil
	}

	return &NormalizedAccount{
		Account:      email,
		Email:        email,
		IsLocalAlias: false,
	}, nil
}

// AccountFromEmail projects a Coze email back to the Forma product account.
func AccountFromEmail(email string) string {
	email = strings.TrimSpace(strings.ToLower(email))
	if strings.HasSuffix(email, "@"+FormaLocalEmailDomain) {
		return strings.TrimSuffix(email, "@"+FormaLocalEmailDomain)
	}
	return email
}
