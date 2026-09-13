/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	datasvc "github.com/coze-dev/coze-studio/backend/domain/forma/data/service"
)

type fieldSide int

const (
	fieldSideInput fieldSide = iota
	fieldSideOutput
)

type capFieldInfo struct {
	LogicalType string
	Required    bool
	Nullable    bool
	Side        fieldSide
}

// validateCompatibility collects stable issue codes for BM pin + contract logical compatibility.
// Infra/port failures must NOT be turned into FAIL codes — callers handle those as clean errors.
func validateCompatibility(
	rev *entity.BusinessCapabilityRevision,
	bm *BusinessModelRevisionEvidence,
	descriptors map[string]*ContractLogicalDescriptor,
) []string {
	var issues []string
	if rev == nil {
		return []string{entity.IssueStructuralInvalid}
	}
	if rev.BusinessModelRevision <= 0 {
		issues = append(issues, entity.IssueBMRevisionMissing)
	} else if bm == nil {
		issues = append(issues, entity.IssueBMNotFound)
	} else {
		if bm.TenantID != "" && bm.TenantID != rev.TenantID {
			issues = append(issues, entity.IssueBMTenantMismatch)
		}
		if bm.BusinessID != "" && bm.BusinessID != rev.BusinessID {
			issues = append(issues, entity.IssueBMBusinessMismatch)
		}
		if bm.Revision != rev.BusinessModelRevision {
			issues = append(issues, entity.IssueBMNotFound)
		}
	}

	bindings := rev.DataContractBindings
	seenPins := map[string]struct{}{}
	for _, b := range bindings {
		pinKey := fmt.Sprintf("%s\x00%s\x00%d", b.DataContractID, b.DataContractRevisionID, b.DataContractVersion)
		if _, dup := seenPins[pinKey]; dup {
			issues = append(issues, entity.IssueContractDuplicatePin)
		}
		seenPins[pinKey] = struct{}{}
	}

	switch rev.CapabilityKind {
	case entity.KindQuery:
		if len(bindings) == 0 {
			issues = append(issues, entity.IssueContractRequired)
		}
		issues = append(issues, validateQueryRules(rev, descriptors)...)
	}

	capFields, typeConflict := collectCapabilityFields(rev)
	if typeConflict {
		issues = append(issues, entity.IssueMappingTypeConflict)
	}
	for _, b := range bindings {
		desc := descriptors[b.DataContractID]
		issues = append(issues, validateBindingAgainstDescriptor(rev, b, desc, capFields)...)
	}
	return uniqueSorted(issues)
}

func validateBindingAgainstDescriptor(
	rev *entity.BusinessCapabilityRevision,
	b entity.DataContractBinding,
	desc *ContractLogicalDescriptor,
	capFields map[string]capFieldInfo,
) []string {
	var issues []string
	if desc == nil {
		return []string{entity.IssueContractNotFound}
	}
	if !strings.EqualFold(desc.Status, "ACTIVE") {
		issues = append(issues, entity.IssueContractNotActive)
	}
	if desc.TenantID != "" && desc.TenantID != rev.TenantID {
		issues = append(issues, entity.IssueContractTenantMismatch)
	}
	if desc.BusinessID != "" && desc.BusinessID != rev.BusinessID {
		issues = append(issues, entity.IssueContractBusinessMismatch)
	}
	if desc.ContractID != b.DataContractID ||
		desc.RevisionID != b.DataContractRevisionID ||
		desc.Version != b.DataContractVersion {
		issues = append(issues, entity.IssueContractPinMismatch)
	}

	contractFields := map[string]ContractLogicalField{}
	for _, f := range desc.LogicalSchema {
		contractFields[f.LogicalKey] = f
	}

	mappedCap := map[string]string{}
	for _, m := range b.LogicalFieldMappings {
		if _, ok := capFields[m.CapabilityLogicalKey]; !ok {
			issues = append(issues, entity.IssueMappingUnknownCapKey)
			continue
		}
		if _, ok := mappedCap[m.CapabilityLogicalKey]; ok {
			issues = append(issues, entity.IssueMappingDuplicateCapKey)
		}
		mappedCap[m.CapabilityLogicalKey] = m.ContractLogicalKey
		cf, ok := contractFields[m.ContractLogicalKey]
		if !ok {
			issues = append(issues, entity.IssueMappingUnknownContractKey)
			continue
		}
		cfld := capFields[m.CapabilityLogicalKey]
		if !datasvc.IsAllowedLogicalType(cfld.LogicalType) || !datasvc.IsAllowedLogicalType(cf.LogicalType) {
			issues = append(issues, entity.IssueTypeMismatch)
		} else if cfld.LogicalType != cf.LogicalType {
			issues = append(issues, entity.IssueTypeMismatch)
		}
		if cfld.Side == fieldSideOutput && cfld.Required && cf.Nullable {
			issues = append(issues, entity.IssueNullability)
		}
	}
	for key := range capFields {
		if _, ok := mappedCap[key]; !ok {
			issues = append(issues, entity.IssueMappingMissing)
		}
	}
	return issues
}

func collectCapabilityFields(rev *entity.BusinessCapabilityRevision) (map[string]capFieldInfo, bool) {
	out := map[string]capFieldInfo{}
	typeConflict := false
	for _, f := range rev.InputSchema.Fields {
		out[f.LogicalKey] = capFieldInfo{
			LogicalType: f.LogicalType, Required: f.Required, Nullable: f.Nullable, Side: fieldSideInput,
		}
	}
	for _, f := range rev.OutputSchema.Fields {
		if existing, ok := out[f.LogicalKey]; ok {
			if existing.LogicalType != f.LogicalType {
				typeConflict = true
			}
			// Prefer output side for nullability when key appears on both.
			existing.Side = fieldSideOutput
			if f.Required {
				existing.Required = true
			}
			out[f.LogicalKey] = existing
		} else {
			out[f.LogicalKey] = capFieldInfo{
				LogicalType: f.LogicalType, Required: f.Required, Nullable: f.Nullable, Side: fieldSideOutput,
			}
		}
	}
	return out, typeConflict
}

func validateQueryRules(rev *entity.BusinessCapabilityRevision, descriptors map[string]*ContractLogicalDescriptor) []string {
	var issues []string
	switch rev.QueryOperation {
	case entity.QueryOpRead:
		if rev.OutputCardinality != entity.CardinalityOne {
			issues = append(issues, entity.IssueQueryPairingInvalid)
		}
	case entity.QueryOpList, entity.QueryOpFilter:
		if rev.OutputCardinality != entity.CardinalityMany {
			issues = append(issues, entity.IssueQueryPairingInvalid)
		}
	default:
		issues = append(issues, entity.IssueQueryOpUnsupported)
		return issues
	}

	op := string(rev.QueryOperation)
	for _, b := range rev.DataContractBindings {
		desc := descriptors[b.DataContractID]
		if desc == nil {
			continue
		}
		if !queryCapAllows(desc.QueryCapabilities, op) {
			issues = append(issues, entity.IssueQueryOpNotAllowed)
		}
	}

	inputRequired := map[string]bool{}
	inputKeys := map[string]struct{}{}
	for _, f := range rev.InputSchema.Fields {
		inputKeys[f.LogicalKey] = struct{}{}
		if f.Required {
			inputRequired[f.LogicalKey] = true
		}
	}
	for _, pc := range rev.Preconditions {
		if pc.LogicalKey != "" {
			if _, ok := inputKeys[pc.LogicalKey]; !ok {
				issues = append(issues, entity.IssueQueryPreconditionKey)
			}
		}
	}

	switch rev.QueryOperation {
	case entity.QueryOpFilter:
		if len(inputRequired) == 0 {
			issues = append(issues, entity.IssueQueryFilterRequiredInput)
		}
		hasCmp := false
		for _, pc := range rev.Preconditions {
			if !isComparisonPredicate(pc.Predicate) {
				continue
			}
			if !inputRequired[pc.LogicalKey] {
				continue
			}
			hasCmp = true
			issues = append(issues, validateFilterPredicate(pc, rev.DataContractBindings, descriptors)...)
		}
		if !hasCmp {
			issues = append(issues, entity.IssueQueryFilterPredicate)
		}
	case entity.QueryOpRead, entity.QueryOpList:
		for _, pc := range rev.Preconditions {
			if isComparisonPredicate(pc.Predicate) {
				issues = append(issues, validateFilterPredicate(pc, rev.DataContractBindings, descriptors)...)
			}
		}
	}
	return issues
}

func validateFilterPredicate(pc entity.Precondition, bindings []entity.DataContractBinding, descriptors map[string]*ContractLogicalDescriptor) []string {
	op, ok := mapPredicateToFilterOp(pc.Predicate)
	if !ok {
		return []string{entity.IssueQueryFilterOperator}
	}
	contractKey := ""
	var desc *ContractLogicalDescriptor
	for _, b := range bindings {
		for _, m := range b.LogicalFieldMappings {
			if m.CapabilityLogicalKey == pc.LogicalKey {
				contractKey = m.ContractLogicalKey
				desc = descriptors[b.DataContractID]
				break
			}
		}
		if contractKey != "" {
			break
		}
	}
	if desc == nil || contractKey == "" {
		return []string{entity.IssueQueryFilterOperator}
	}
	for _, fs := range desc.FilterSchema {
		if fs.LogicalKey != contractKey {
			continue
		}
		for _, allowed := range fs.Operators {
			if allowed == op {
				return nil
			}
		}
		return []string{entity.IssueQueryFilterOperator}
	}
	return []string{entity.IssueQueryFilterOperator}
}

func isComparisonPredicate(p entity.PredicateKind) bool {
	switch p {
	case entity.PredicateEQ, entity.PredicateNEQ, entity.PredicateGT, entity.PredicateGTE,
		entity.PredicateLT, entity.PredicateLTE, entity.PredicateIN, entity.PredicateNotIn:
		return true
	default:
		return false
	}
}

func mapPredicateToFilterOp(p entity.PredicateKind) (string, bool) {
	switch p {
	case entity.PredicateEQ:
		return "EQ", true
	case entity.PredicateNEQ:
		return "NE", true
	case entity.PredicateGT:
		return "GT", true
	case entity.PredicateGTE:
		return "GTE", true
	case entity.PredicateLT:
		return "LT", true
	case entity.PredicateLTE:
		return "LTE", true
	case entity.PredicateIN:
		return "IN", true
	default:
		return "", false
	}
}

func queryCapAllows(caps []string, op string) bool {
	for _, c := range caps {
		if c == op {
			return true
		}
	}
	return false
}

func uniqueSorted(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, c := range in {
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}
