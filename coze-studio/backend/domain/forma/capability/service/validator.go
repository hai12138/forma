/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
)

var (
	allowedPredicates = map[string]struct{}{
		"EXISTS": {}, "EQ": {}, "NEQ": {}, "IN": {}, "NOT_IN": {},
		"GT": {}, "GTE": {}, "LT": {}, "LTE": {}, "EMPTY": {}, "NOT_EMPTY": {},
	}
	fieldBasedPredicates = map[string]struct{}{
		"EQ": {}, "NEQ": {}, "IN": {}, "NOT_IN": {},
		"GT": {}, "GTE": {}, "LT": {}, "LTE": {}, "EXISTS": {}, "EMPTY": {}, "NOT_EMPTY": {},
	}
	allowedEffectKinds = map[string]struct{}{
		"READ_ONLY": {}, "INTENT": {}, "STATE_CHANGE": {}, "NOTIFY": {},
	}
	executablePattern = regexp.MustCompile(`(?i)(SELECT\s|;|\$\(|eval\(|os\.system|import\s|require\(|Function\(|` + "`[^`]*\\$[^`]*`)")
	secretPattern     = regexp.MustCompile(`(?i)(password|token|authorization|api[_-]?key|cookie|bearer)`)
)

// ValidateMaterializationPayload rejects illegal or secret-bearing semantic payloads.
func ValidateMaterializationPayload(p entity.SemanticPayload) error {
	if strings.TrimSpace(p.Name) == "" {
		return entity.ErrInvalidPayload
	}
	if p.BusinessModelRevision <= 0 {
		return entity.ErrInvalidPayload
	}
	switch p.CapabilityKind {
	case entity.KindQuery:
		if p.QueryOperation != entity.QueryOpRead && p.QueryOperation != entity.QueryOpList && p.QueryOperation != entity.QueryOpFilter {
			return entity.ErrInvalidPayload
		}
		if p.OutputCardinality != entity.CardinalityOne && p.OutputCardinality != entity.CardinalityMany {
			return entity.ErrInvalidPayload
		}
	case entity.KindCommand:
		// COMMAND must not require query ops; empty is OK.
	default:
		return entity.ErrInvalidPayload
	}
	if containsSecret(p.Name) || containsSecret(p.Description) || containsExecutable(p.Name) || containsExecutable(p.Description) {
		return entity.ErrInvalidPayload
	}
	for _, pc := range p.Preconditions {
		if err := validatePrecondition(pc); err != nil {
			return err
		}
	}
	for _, ef := range p.Effects {
		if err := validateEffect(ef); err != nil {
			return err
		}
	}
	return nil
}

func validatePrecondition(pc entity.Precondition) error {
	pred := strings.TrimSpace(pc.Predicate)
	if _, ok := allowedPredicates[pred]; !ok {
		return entity.ErrInvalidPayload
	}
	if _, ok := fieldBasedPredicates[pred]; ok {
		if strings.TrimSpace(pc.LogicalKey) == "" {
			return entity.ErrInvalidPayload
		}
	}
	if containsSecret(pc.Description) || containsSecret(pc.LogicalKey) || containsExecutable(pc.Description) {
		return entity.ErrInvalidPayload
	}
	if pc.Comparand != nil {
		if !isLiteralComparand(pc.Comparand) {
			return entity.ErrInvalidPayload
		}
		if s, ok := pc.Comparand.(string); ok {
			if containsExecutable(s) || containsSecret(s) {
				return entity.ErrInvalidPayload
			}
		}
	}
	return nil
}

func validateEffect(ef entity.Effect) error {
	kind := strings.TrimSpace(ef.Kind)
	if _, ok := allowedEffectKinds[kind]; !ok {
		return entity.ErrInvalidPayload
	}
	if containsSecret(ef.Description) || containsExecutable(ef.Description) {
		return entity.ErrInvalidPayload
	}
	return nil
}

func isLiteralComparand(v any) bool {
	switch v.(type) {
	case nil, string, bool, float64, float32, int, int32, int64, uint, uint32, uint64:
		return true
	case []string:
		return true
	case []any:
		items := v.([]any)
		for _, item := range items {
			switch item.(type) {
			case nil, string, bool, float64, float32, int, int32, int64:
				continue
			default:
				return false
			}
		}
		return true
	default:
		return false
	}
}

func containsExecutable(s string) bool {
	return executablePattern.MatchString(s)
}

func containsSecret(s string) bool {
	return secretPattern.MatchString(s)
}

// SanitizeAnalysisError maps generator/persistence errors to a stable domain error.
// Never returns raw upstream generator text to callers.
func SanitizeAnalysisError(err error) error {
	if err == nil {
		return nil
	}
	if err == entity.ErrStaleGeneration || err == entity.ErrInvalidPayload || err == entity.ErrConsistency ||
		err == entity.ErrIdempotencyConflict || err == entity.ErrNotConfigured || err == entity.ErrUoWNotConfigured {
		return err
	}
	return entity.ErrAnalysisFailed
}

// SanitizedAnalysisErrorCode returns a stable error_code stored on failed analysis runs.
func SanitizedAnalysisErrorCode(err error) string {
	if err == nil {
		return ""
	}
	switch {
	case err == entity.ErrStaleGeneration:
		return "FORMA_CAPABILITY_STALE_GENERATION"
	case err == entity.ErrInvalidPayload:
		return "FORMA_CAPABILITY_INVALID_REQUEST"
	case err == entity.ErrConsistency:
		return "FORMA_CAPABILITY_CONSISTENCY"
	default:
		return "FORMA_CAPABILITY_ANALYSIS_FAILED"
	}
}

func parseOwnerID(actor string) (int64, error) {
	s := strings.TrimSpace(actor)
	if s == "" {
		return 0, nil
	}
	var n int64
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return 0, entity.ErrInvalidPayload
	}
	// Reject non-numeric remainder (e.g. "12abc").
	if fmt.Sprintf("%d", n) != s {
		return 0, entity.ErrInvalidPayload
	}
	return n, nil
}
