/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"encoding/json"
	"strings"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
)

var forbiddenPublicConfigKeys = map[string]struct{}{
	"password": {}, "passwd": {}, "api_key": {}, "apikey": {}, "token": {},
	"access_token": {}, "refresh_token": {}, "authorization": {}, "cookie": {},
	"secret": {}, "client_secret": {}, "private_key": {}, "credential": {},
}

func ValidatePublicConfig(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return entity.ErrPublicConfigInvalid
	}
	if containsSecretKey(value) {
		return entity.ErrPublicConfigInvalid
	}
	return nil
}

func containsSecretKey(value any) bool {
	switch v := value.(type) {
	case map[string]any:
		for key, child := range v {
			normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "-", "_"), " ", "_"))
			compact := strings.ReplaceAll(normalized, "_", "")
			if _, forbidden := forbiddenPublicConfigKeys[normalized]; forbidden {
				return true
			}
			if compact == "clientsecret" || compact == "apikey" || compact == "accesstoken" ||
				compact == "refreshtoken" || compact == "privatekey" {
				return true
			}
			if strings.Contains(normalized, "password") || strings.HasSuffix(normalized, "_secret") ||
				strings.HasSuffix(normalized, "_token") || strings.HasSuffix(normalized, "_api_key") {
				return true
			}
			if containsSecretKey(child) {
				return true
			}
		}
	case []any:
		for _, child := range v {
			if containsSecretKey(child) {
				return true
			}
		}
	}
	return false
}
