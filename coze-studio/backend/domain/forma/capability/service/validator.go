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
	allowedPredicates = map[entity.PredicateKind]struct{}{
		entity.PredicateExists: {}, entity.PredicateEQ: {}, entity.PredicateNEQ: {},
		entity.PredicateIN: {}, entity.PredicateNotIn: {},
		entity.PredicateGT: {}, entity.PredicateGTE: {}, entity.PredicateLT: {}, entity.PredicateLTE: {},
		entity.PredicateEmpty: {}, entity.PredicateNotEmpty: {},
	}
	fieldBasedPredicates = map[entity.PredicateKind]struct{}{
		entity.PredicateEQ: {}, entity.PredicateNEQ: {}, entity.PredicateIN: {}, entity.PredicateNotIn: {},
		entity.PredicateGT: {}, entity.PredicateGTE: {}, entity.PredicateLT: {}, entity.PredicateLTE: {},
		entity.PredicateExists: {}, entity.PredicateEmpty: {}, entity.PredicateNotEmpty: {},
	}
	scalarComparandPredicates = map[entity.PredicateKind]struct{}{
		entity.PredicateEQ: {}, entity.PredicateNEQ: {},
		entity.PredicateGT: {}, entity.PredicateGTE: {}, entity.PredicateLT: {}, entity.PredicateLTE: {},
	}
	listComparandPredicates = map[entity.PredicateKind]struct{}{
		entity.PredicateIN: {}, entity.PredicateNotIn: {},
	}
	nilComparandPredicates = map[entity.PredicateKind]struct{}{
		entity.PredicateExists: {}, entity.PredicateEmpty: {}, entity.PredicateNotEmpty: {},
	}
	allowedEffectKinds = map[entity.EffectKind]struct{}{
		entity.EffectReadOnly: {}, entity.EffectIntent: {}, entity.EffectStateChange: {}, entity.EffectNotify: {},
	}
	executablePattern = regexp.MustCompile(`(?i)(SELECT\s|;|\$\(|eval\(|os\.system|import\s|require\(|Function\(|` + "`[^`]*\\$[^`]*`)")
	shellPattern      = regexp.MustCompile(`(?i)(\brm\s+-rf\b|\bcurl\s+|\bwget\s+|\|.*sh\b|/bin/sh|/bin/bash)`)
	// Free-text secret shapes — includes bare word "secret". Do NOT apply to opaque analysis IDs.
	// Do NOT ban "password" as a substring of business names (GetPasswordReset).
	freeTextSecretPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bsecret\b`),
		regexp.MustCompile(`(?i)(api[_-]?key|authorization)\s*[:=]`),
		regexp.MustCompile(`(?i)bearer\s+[a-z0-9._\-]{8,}`),
		regexp.MustCompile(`(?i)password\s*[:=]\s*\S+`),
		regexp.MustCompile(`(?i)authorization\s*:`),
		regexp.MustCompile(`(?i)(session|sid|jsessionid)\s*=\s*\S+`),
		regexp.MustCompile(`(?i)-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----`),
		regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]+\.`),
		regexp.MustCompile(`^[A-Za-z0-9+/]{64,}={0,2}$`),
		regexp.MustCompile(`(?i)client_secret\s*=`),
		regexp.MustCompile(`(?i)access_token\s*=`),
		regexp.MustCompile(`(?i)refresh_token\s*=`),
		regexp.MustCompile(`(?i)api_key\s*=`),
	}
	// Credential shapes for opaque analysis pin/ref strings — no bare \bsecret\b.
	credentialShapePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(api[_-]?key|authorization)\s*[:=]`),
		regexp.MustCompile(`(?i)bearer\s+[a-z0-9._\-]{8,}`),
		regexp.MustCompile(`(?i)password\s*[:=]\s*\S+`),
		regexp.MustCompile(`(?i)authorization\s*:`),
		regexp.MustCompile(`(?i)(session|sid|jsessionid)\s*=\s*\S+`),
		regexp.MustCompile(`(?i)-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----`),
		regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]+\.`),
		regexp.MustCompile(`^[A-Za-z0-9+/]{64,}={0,2}$`),
		regexp.MustCompile(`(?i)client_secret\s*=`),
		regexp.MustCompile(`(?i)access_token\s*=`),
		regexp.MustCompile(`(?i)refresh_token\s*=`),
		regexp.MustCompile(`(?i)api_key\s*=`),
	}
	credentialKeyName = regexp.MustCompile(`(?i)^(api[_-]?key|token|password|authorization|secret|cookie|bearer|access_token|refresh_token)$`)
	opaqueIDPattern   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:\-]{0,127}$`)
)

// ValidateOpaqueID rejects spaces, SQL, and other unsafe identifier shapes.
func ValidateOpaqueID(id string) error {
	s := strings.TrimSpace(id)
	if s == "" || s != id {
		return entity.ErrInvalidPayload
	}
	if !opaqueIDPattern.MatchString(s) {
		return entity.ErrInvalidPayload
	}
	if containsExecutable(s) || shellPattern.MatchString(s) {
		return entity.ErrInvalidPayload
	}
	return nil
}

// ValidateAnalysisRequestIDs checks DataContract pins and requirement refs before persist/generator.
func ValidateAnalysisRequestIDs(a entity.AnalysisRequest) error {
	for _, pin := range a.DataContractPins {
		if err := ValidateOpaqueID(pin.DataContractID); err != nil {
			return err
		}
	}
	for _, ref := range a.RequirementRefs {
		if err := ValidateOpaqueID(ref); err != nil {
			return err
		}
	}
	return nil
}

// ValidateAnalysisRequest validates opaque IDs plus credential-shape secret checks on each pin/ref.
func ValidateAnalysisRequest(a entity.AnalysisRequest) error {
	if err := ValidateAnalysisRequestIDs(a); err != nil {
		return err
	}
	for _, pin := range a.DataContractPins {
		if containsCredentialShape(pin.DataContractID) {
			return entity.ErrInvalidPayload
		}
	}
	for _, ref := range a.RequirementRefs {
		if containsCredentialShape(ref) {
			return entity.ErrInvalidPayload
		}
	}
	return nil
}

// ValidateMaterializationPayload rejects illegal or secret-bearing semantic payloads.
func ValidateMaterializationPayload(p entity.SemanticPayload) error {
	if strings.TrimSpace(p.Name) == "" {
		return entity.ErrInvalidPayload
	}
	if p.BusinessModelRevision <= 0 {
		return entity.ErrInvalidPayload
	}
	if err := validateQueryPairing(p); err != nil {
		return err
	}
	if containsSecret(p.Name) || containsSecret(p.Description) || containsExecutable(p.Name) || containsExecutable(p.Description) {
		return entity.ErrInvalidPayload
	}
	if err := validateLogicalSchema(p.InputSchema); err != nil {
		return err
	}
	if err := validateLogicalSchema(p.OutputSchema); err != nil {
		return err
	}
	for _, b := range p.DataContractBindings {
		if err := validateBinding(b); err != nil {
			return err
		}
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

// validateQueryPairing enforces strict QUERY/COMMAND cardinality rules.
// READ → ONE; LIST/FILTER → MANY; COMMAND → both empty.
func validateQueryPairing(p entity.SemanticPayload) error {
	switch p.CapabilityKind {
	case entity.KindQuery:
		switch p.QueryOperation {
		case entity.QueryOpRead:
			if p.OutputCardinality != entity.CardinalityOne {
				return entity.ErrInvalidPayload
			}
		case entity.QueryOpList, entity.QueryOpFilter:
			if p.OutputCardinality != entity.CardinalityMany {
				return entity.ErrInvalidPayload
			}
		default:
			return entity.ErrInvalidPayload
		}
	case entity.KindCommand:
		if p.QueryOperation != "" || p.OutputCardinality != "" {
			return entity.ErrInvalidPayload
		}
	default:
		return entity.ErrInvalidPayload
	}
	return nil
}

func validateLogicalSchema(schema entity.LogicalSchema) error {
	for _, f := range schema.Fields {
		if isCredentialKeyName(f.LogicalKey) {
			return entity.ErrInvalidPayload
		}
		if f.LogicalKey != "" {
			if err := ValidateOpaqueID(f.LogicalKey); err != nil {
				return entity.ErrInvalidPayload
			}
		}
		for _, s := range []string{f.LogicalKey, f.LogicalType, f.Description} {
			if containsSecret(s) || containsExecutable(s) || shellPattern.MatchString(s) {
				return entity.ErrInvalidPayload
			}
		}
	}
	return nil
}

func validateBinding(b entity.DataContractBinding) error {
	if err := ValidateOpaqueID(b.DataContractID); err != nil {
		return entity.ErrInvalidPayload
	}
	if err := ValidateOpaqueID(b.DataContractRevisionID); err != nil {
		return entity.ErrInvalidPayload
	}
	if isCredentialKeyName(b.DataContractID) || isCredentialKeyName(b.DataContractRevisionID) {
		return entity.ErrInvalidPayload
	}
	if containsSecret(b.DataContractID) || containsSecret(b.DataContractRevisionID) {
		return entity.ErrInvalidPayload
	}
	// Version > 0 if set (zero means unset).
	if b.DataContractVersion < 0 {
		return entity.ErrInvalidPayload
	}
	for _, m := range b.LogicalFieldMappings {
		if err := ValidateOpaqueID(m.CapabilityLogicalKey); err != nil {
			return entity.ErrInvalidPayload
		}
		if err := ValidateOpaqueID(m.ContractLogicalKey); err != nil {
			return entity.ErrInvalidPayload
		}
		for _, s := range []string{m.CapabilityLogicalKey, m.ContractLogicalKey} {
			if isCredentialKeyName(s) || containsSecret(s) || containsExecutable(s) {
				return entity.ErrInvalidPayload
			}
		}
	}
	return nil
}

func validatePrecondition(pc entity.Precondition) error {
	// Exact match — NO TrimSpace for allowlist (" EQ " must fail).
	if _, ok := allowedPredicates[pc.Predicate]; !ok {
		return entity.ErrInvalidPayload
	}
	if err := ValidateOpaqueID(pc.ID); err != nil {
		return entity.ErrInvalidPayload
	}
	if isCredentialKeyName(pc.ID) {
		return entity.ErrInvalidPayload
	}
	if _, ok := fieldBasedPredicates[pc.Predicate]; ok {
		if strings.TrimSpace(pc.LogicalKey) == "" {
			return entity.ErrInvalidPayload
		}
		if err := ValidateOpaqueID(pc.LogicalKey); err != nil {
			return entity.ErrInvalidPayload
		}
	} else if pc.LogicalKey != "" {
		if err := ValidateOpaqueID(pc.LogicalKey); err != nil {
			return entity.ErrInvalidPayload
		}
	}
	if isCredentialKeyName(pc.LogicalKey) {
		return entity.ErrInvalidPayload
	}
	for _, s := range []string{pc.ID, pc.LogicalKey, pc.Description} {
		if containsSecret(s) || containsExecutable(s) || shellPattern.MatchString(s) {
			return entity.ErrInvalidPayload
		}
	}
	if _, ok := nilComparandPredicates[pc.Predicate]; ok {
		if pc.Comparand != nil {
			return entity.ErrInvalidPayload
		}
	}
	if _, ok := listComparandPredicates[pc.Predicate]; ok {
		if err := validateListComparand(pc.Comparand); err != nil {
			return err
		}
	}
	if _, ok := scalarComparandPredicates[pc.Predicate]; ok {
		if pc.Comparand == nil || !isScalarLiteral(pc.Comparand) {
			return entity.ErrInvalidPayload
		}
		if err := validateComparandSecrets(pc.Comparand); err != nil {
			return err
		}
	}
	if pc.Comparand != nil {
		if !isLiteralComparand(pc.Comparand) {
			return entity.ErrInvalidPayload
		}
		if err := validateComparandSecrets(pc.Comparand); err != nil {
			return err
		}
	}
	return nil
}

func validateListComparand(v any) error {
	switch t := v.(type) {
	case []string:
		if len(t) == 0 {
			return entity.ErrInvalidPayload
		}
		for _, s := range t {
			if err := validateComparandSecrets(s); err != nil {
				return err
			}
		}
		return nil
	case []any:
		if len(t) == 0 {
			return entity.ErrInvalidPayload
		}
		for _, item := range t {
			if !isScalarLiteral(item) {
				return entity.ErrInvalidPayload
			}
			if err := validateComparandSecrets(item); err != nil {
				return err
			}
		}
		return nil
	default:
		return entity.ErrInvalidPayload
	}
}

func validateEffect(ef entity.Effect) error {
	// Exact match — NO TrimSpace for allowlist.
	if _, ok := allowedEffectKinds[ef.Kind]; !ok {
		return entity.ErrInvalidPayload
	}
	if err := ValidateOpaqueID(ef.ID); err != nil {
		return entity.ErrInvalidPayload
	}
	if isCredentialKeyName(ef.ID) {
		return entity.ErrInvalidPayload
	}
	if ef.LogicalKey != "" {
		if err := ValidateOpaqueID(ef.LogicalKey); err != nil {
			return entity.ErrInvalidPayload
		}
	}
	if isCredentialKeyName(ef.LogicalKey) {
		return entity.ErrInvalidPayload
	}
	for _, s := range []string{ef.ID, ef.LogicalKey, ef.Description} {
		if containsSecret(s) || containsExecutable(s) || shellPattern.MatchString(s) {
			return entity.ErrInvalidPayload
		}
	}
	return nil
}

func isCredentialKeyName(s string) bool {
	return credentialKeyName.MatchString(strings.TrimSpace(s))
}

func validateComparandSecrets(v any) error {
	switch t := v.(type) {
	case string:
		if containsExecutable(t) || containsSecret(t) || shellPattern.MatchString(t) {
			return entity.ErrInvalidPayload
		}
	case []string:
		for _, s := range t {
			if containsExecutable(s) || containsSecret(s) || shellPattern.MatchString(s) {
				return entity.ErrInvalidPayload
			}
		}
	case []any:
		for _, item := range t {
			if err := validateComparandSecrets(item); err != nil {
				return err
			}
		}
	}
	return nil
}

func isScalarLiteral(v any) bool {
	switch v.(type) {
	case string, bool, float64, float32, int, int32, int64, uint, uint32, uint64:
		return true
	default:
		return false
	}
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
			if !isScalarLiteral(item) && item != nil {
				return false
			}
			switch item.(type) {
			case nil, string, bool, float64, float32, int, int32, int64, uint, uint32, uint64:
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
	if strings.TrimSpace(s) == "" {
		return false
	}
	for _, re := range freeTextSecretPatterns {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}

// containsCredentialShape blocks credential-like tokens in opaque analysis IDs.
// Does NOT reject the bare word "secret" alone.
func containsCredentialShape(s string) bool {
	if strings.TrimSpace(s) == "" {
		return false
	}
	for _, re := range credentialShapePatterns {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}

// SanitizeAnalysisError maps generator/persistence errors to a stable domain error.
// Never returns raw upstream generator text to callers.
func SanitizeAnalysisError(err error) error {
	if err == nil {
		return nil
	}
	mapped := MapRepoError(err)
	if mapped == entity.ErrConsistency || mapped == entity.ErrConflict || mapped == entity.ErrNotFound {
		return mapped
	}
	if err == entity.ErrStaleGeneration || err == entity.ErrInvalidPayload || err == entity.ErrConsistency ||
		err == entity.ErrIdempotencyConflict || err == entity.ErrNotConfigured || err == entity.ErrUoWNotConfigured ||
		err == entity.ErrUoWCommitFailed {
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
