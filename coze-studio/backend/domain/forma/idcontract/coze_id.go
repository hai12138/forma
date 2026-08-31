/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 *
 * External/public Coze resource IDs are strings at API boundaries.
 * Domain/repository layers may keep int64.
 */

package idcontract

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

// FormatCozeID serializes a domain int64 Coze resource ID for public JSON.
func FormatCozeID(id int64) string {
	if id <= 0 {
		return ""
	}
	return strconv.FormatInt(id, 10)
}

// ParseCozeID parses a public string Coze resource ID into domain int64.
// Rejects empty, non-digit, floats, scientific notation, zero/negative, and overflow.
func ParseCozeID(raw string) (int64, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, fmt.Errorf("coze id is required")
	}
	if strings.ContainsAny(s, ".eE+") {
		return 0, fmt.Errorf("coze id must be a decimal integer string")
	}
	if s[0] == '-' {
		return 0, fmt.Errorf("coze id must be positive")
	}
	if s[0] == '+' {
		return 0, fmt.Errorf("coze id must be a decimal integer string")
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return 0, fmt.Errorf("coze id must be numeric digits only")
		}
	}
	if len(s) > 1 && s[0] == '0' {
		return 0, fmt.Errorf("coze id must not have leading zeros")
	}
	// Overflow-safe parse via big-endian digit accumulation.
	var n uint64
	for _, r := range s {
		d := uint64(r - '0')
		if n > (math.MaxUint64-d)/10 {
			return 0, fmt.Errorf("coze id overflows uint64")
		}
		n = n*10 + d
	}
	if n == 0 {
		return 0, fmt.Errorf("coze id must be > 0")
	}
	if n > math.MaxInt64 {
		return 0, fmt.Errorf("coze id overflows int64")
	}
	return int64(n), nil
}

// ParseOptionalCozeID treats empty as absent (0, nil).
func ParseOptionalCozeID(raw string) (int64, error) {
	if strings.TrimSpace(raw) == "" {
		return 0, nil
	}
	return ParseCozeID(raw)
}
