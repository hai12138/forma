/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	datasvc "github.com/coze-dev/coze-studio/backend/domain/forma/data/service"
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
	// Shared credential-shape patterns for free-text AND opaque IDs that already passed ValidateOpaqueID.
	// Intentionally omits bare \bsecret\b so "trade secret" / ReviewTradeSecretPolicy remain allowed.
	// Do NOT ban "password" as a substring of business names (GetPasswordReset).
	sharedCredentialShapePatterns = []*regexp.Regexp{
		regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]+\.`),           // JWT
		regexp.MustCompile(`(?i)\bghp_[A-Za-z0-9]{20,}\b`),                        // ghp_ prefix token
		regexp.MustCompile(`(?i)\bgithub_pat_[A-Za-z0-9_]{20,}\b`),                // github_pat_ prefix token
		regexp.MustCompile(`(?i)\bsk-proj-[A-Za-z0-9_-]{16,}\b`),                  // sk-proj- prefix token
		regexp.MustCompile(`(?i)\bsk-[A-Za-z0-9]{20,}\b`),                         // sk- prefix token
		regexp.MustCompile(`(?i)\bxox[baprs]-[A-Za-z0-9-]{10,}\b`),                // xox* prefix token
		regexp.MustCompile(`(?i)\bxapp-[A-Za-z0-9-]{10,}\b`),                      // xapp- prefix token
		regexp.MustCompile(`(?i)-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----`), // PEM private key
		regexp.MustCompile(`^[A-Za-z0-9]{48,}$`),                                  // long separator-free random token
		regexp.MustCompile(`^[A-Za-z0-9+/]{64,}={0,2}$`),                          // long base64 (free-text)
	}
	// Assignment / delimiter forms — contain '=' or spaces, cannot pass ValidateOpaqueID.
	// Require non-empty RHS. Do NOT ban bare keywords (trade secret, cookie policy, tokenization).
	freeTextAssignmentPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(api[_-]?key|authorization)\s*[:=]\s*\S+`),
		regexp.MustCompile(`(?i)password\s*[:=]\s*\S+`),
		regexp.MustCompile(`(?i)\btoken\s*=\s*\S+`),
		regexp.MustCompile(`(?i)\bcookie\s*=\s*\S+`),
		regexp.MustCompile(`(?i)\bsecret\s*=\s*\S+`),
		regexp.MustCompile(`(?i)(session|sid|jsessionid)\s*=\s*\S+`),
		regexp.MustCompile(`(?i)client_secret\s*=\s*\S+`),
		regexp.MustCompile(`(?i)access_token\s*=\s*\S+`),
		regexp.MustCompile(`(?i)refresh_token\s*=\s*\S+`),
		regexp.MustCompile(`(?i)api_key\s*=\s*\S+`),
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
		if pin.DataContractVersion <= 0 {
			return entity.ErrInvalidPayload
		}
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
		if err := ValidateOpaqueID(f.LogicalKey); err != nil {
			return entity.ErrInvalidPayload
		}
		if isCredentialKeyName(f.LogicalKey) {
			return entity.ErrInvalidPayload
		}
		lt := strings.TrimSpace(f.LogicalType)
		if lt == "" || lt != f.LogicalType {
			return entity.ErrInvalidPayload
		}
		if !datasvc.IsAllowedLogicalType(f.LogicalType) {
			return entity.ErrInvalidPayload
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
	if containsCredentialShape(b.DataContractID) || containsCredentialShape(b.DataContractRevisionID) {
		return entity.ErrInvalidPayload
	}
	if b.DataContractVersion <= 0 {
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

// containsBearerCredential rejects Authorization-style Bearer tokens (any non-whitespace token).
// Allows ONLY when the entire trimmed input equals "bearer of responsibility" (case-insensitive).
// Mid-string occurrences of that phrase are NOT exempted. Whitespace after Bearer is detected via
// unicode.IsSpace (not regexp \s), so NBSP and other Unicode spaces count.
func containsBearerCredential(s string) bool {
	if strings.EqualFold(strings.TrimSpace(s), "bearer of responsibility") {
		return false
	}
	const keyword = "bearer"
	runes := []rune(s)
	n := len(runes)
	kwLen := len(keyword)
	for i := 0; i <= n-kwLen; i++ {
		if i > 0 && isASCIIWordChar(runes[i-1]) {
			continue
		}
		matched := true
		for j := 0; j < kwLen; j++ {
			if unicode.ToLower(runes[i+j]) != rune(keyword[j]) {
				matched = false
				break
			}
		}
		if !matched {
			continue
		}
		end := i + kwLen
		if end < n && isASCIIWordChar(runes[end]) {
			continue
		}
		j := end
		spaces := 0
		for j < n && unicode.IsSpace(runes[j]) {
			spaces++
			j++
		}
		if spaces == 0 {
			continue
		}
		if j < n {
			return true
		}
	}
	return false
}

func isASCIIWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

func containsSecret(s string) bool {
	if strings.TrimSpace(s) == "" {
		return false
	}
	if containsBearerCredential(s) {
		return true
	}
	for _, re := range sharedCredentialShapePatterns {
		if re.MatchString(s) {
			return true
		}
	}
	for _, re := range freeTextAssignmentPatterns {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}

// containsCredentialShape blocks credential-like tokens in opaque analysis IDs.
// Uses the shared pattern set + Bearer (no assignment forms that cannot pass ValidateOpaqueID).
// Does NOT reject the bare word "secret" alone.
func containsCredentialShape(s string) bool {
	if strings.TrimSpace(s) == "" {
		return false
	}
	if containsBearerCredential(s) {
		return true
	}
	for _, re := range sharedCredentialShapePatterns {
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

