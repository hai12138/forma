/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	businesssvc "github.com/coze-dev/coze-studio/backend/domain/forma/business/service"
	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/data/repository"
)

type ContractService interface {
	CreateContract(context.Context, *CreateContractInput) (*entity.DataContract, *entity.DataContractRevision, error)
	GetContract(context.Context, string, string) (*entity.DataContract, error)
	ListContracts(context.Context, string, string) ([]*entity.DataContract, error)
	GetRevision(context.Context, string, string) (*entity.DataContractRevision, error)
	ListRevisions(context.Context, string, string) ([]*entity.DataContractRevision, error)
	CreateRevision(context.Context, *CreateRevisionInput) (*entity.DataContractRevision, error)
	ValidateRevision(context.Context, string, string, string) (*entity.DataContractRevision, *entity.DataValidationResult, error)
	ActivateRevision(context.Context, string, string, string, string) (*entity.DataContractRevision, error)
	DeprecateRevision(context.Context, string, string, string, string) (*entity.DataContractRevision, error)
	ListValidationResults(context.Context, string, string) ([]*entity.DataValidationResult, error)
	EvaluateDrift(context.Context, *EvaluateDriftInput) (*entity.DataDriftResult, *entity.DataContractRevision, error)
	ListDriftResults(context.Context, string, string) ([]*entity.DataDriftResult, error)
	EvaluateBusinessGap(context.Context, *EvaluateGapInput) (*entity.DataContractGapResult, error)
	ListGapResults(context.Context, string, string) ([]*entity.DataContractGapResult, error)
	ListLifecycleEvents(context.Context, string, string) ([]*entity.DataContractLifecycleEvent, error)
	BuildContractDescriptor(*entity.DataContractRevision) *entity.DataContractDescriptor
	GetActiveContractDescriptor(context.Context, string, string) (*entity.DataContractDescriptor, error)
}

type ContractComponents struct {
	Contracts  repository.ContractRepository
	Data       repository.DataRepository
	Mappings   repository.MappingRepository
	Sources    repository.DataSourceRepository
	Business   businesssvc.BusinessService
}

type contractService struct {
	contracts repository.ContractRepository
	data      repository.DataRepository
	mappings  repository.MappingRepository
	sources   repository.DataSourceRepository
	business  businesssvc.BusinessService
}

func NewContractService(c *ContractComponents) ContractService {
	if c == nil {
		return &contractService{}
	}
	return &contractService{
		contracts: c.Contracts,
		data:      c.Data,
		mappings:  c.Mappings,
		sources:   c.Sources,
		business:  c.Business,
	}
}

type CreateContractInput struct {
	TenantID              string
	BusinessID            string
	BusinessModelRevision int32
	Name                  string
	Description           string
	RequirementIDs        []string
	LogicalSchema         entity.ContractLogicalSchema
	QueryCapabilities     []entity.QueryCapability
	FilterSchema          entity.FilterSchema
	SortSchema            entity.SortSchema
	PaginationPolicy      entity.PaginationPolicy
	FreshnessPolicy       entity.FreshnessPolicy
	ClassificationPolicy  map[string]entity.DataClassification
	MappingIDs            []string
	AccessPolicyRef       string
	ActorID               string
}

type CreateRevisionInput struct {
	TenantID              string
	ContractID            string
	BaseRevisionID        string
	BusinessModelRevision int32
	Name                  string
	Description           string
	RequirementIDs        []string
	LogicalSchema         entity.ContractLogicalSchema
	QueryCapabilities     []entity.QueryCapability
	FilterSchema          entity.FilterSchema
	SortSchema            entity.SortSchema
	PaginationPolicy      entity.PaginationPolicy
	FreshnessPolicy       entity.FreshnessPolicy
	ClassificationPolicy  map[string]entity.DataClassification
	MappingIDs            []string
	AccessPolicyRef       string
	ActorID               string
}

type EvaluateDriftInput struct {
	TenantID       string
	RevisionID     string
	NewSnapshotIDs map[string]string // keyed by pinned schema_snapshot_id
	ActorID        string
}

type EvaluateGapInput struct {
	TenantID   string
	RevisionID string
	ActorID    string
}

func (s *contractService) configured() bool {
	return s.contracts != nil && s.data != nil && s.mappings != nil && s.sources != nil && s.business != nil
}

func (s *contractService) CreateContract(ctx context.Context, in *CreateContractInput) (*entity.DataContract, *entity.DataContractRevision, error) {
	if !s.configured() {
		return nil, nil, entity.ErrNotConfigured
	}
	if in == nil || strings.TrimSpace(in.TenantID) == "" || strings.TrimSpace(in.BusinessID) == "" || in.BusinessModelRevision <= 0 || strings.TrimSpace(in.Name) == "" {
		return nil, nil, entity.ErrContractInvalidPayload
	}
	if err := ValidateRequirementIDsUnique(in.RequirementIDs); err != nil {
		return nil, nil, err
	}
	if err := ValidateQueryCapabilities(in.QueryCapabilities); err != nil {
		return nil, nil, err
	}
	bindings, err := s.materializeBindings(ctx, in.TenantID, in.BusinessID, in.BusinessModelRevision, in.MappingIDs)
	if err != nil {
		return nil, nil, err
	}
	now := time.Now().UTC()
	contractID := newID("ctr")
	revisionID := newID("crev")
	contract := &entity.DataContract{
		ContractID: contractID, TenantID: in.TenantID, BusinessID: in.BusinessID,
		CreatedBy: in.ActorID, CreatedAt: now, UpdatedAt: now,
	}
	rev := &entity.DataContractRevision{
		RevisionID: revisionID, TenantID: in.TenantID, BusinessID: in.BusinessID, ContractID: contractID,
		Version: 1, Status: entity.ContractStatusDraft, BusinessModelRevision: in.BusinessModelRevision,
		Name: in.Name, Description: in.Description,
		RequirementIDs: append([]string(nil), in.RequirementIDs...),
		LogicalSchema: entity.ContractLogicalSchema{Fields: append([]entity.LogicalField(nil), in.LogicalSchema.Fields...)},
		QueryCapabilities: append([]entity.QueryCapability(nil), in.QueryCapabilities...),
		FilterSchema: in.FilterSchema, SortSchema: in.SortSchema, PaginationPolicy: in.PaginationPolicy,
		FreshnessPolicy: in.FreshnessPolicy, ClassificationPolicy: cloneClassPolicy(in.ClassificationPolicy),
		BindingRefs: bindings, AccessPolicyRef: in.AccessPolicyRef,
		CreatedBy: in.ActorID, CreatedAt: now, UpdatedAt: now,
	}
	err = s.contracts.Transaction(ctx, func(tx repository.ContractRepository) error {
		if err := tx.CreateContract(ctx, contract); err != nil {
			return err
		}
		return tx.CreateRevision(ctx, rev)
	})
	if err != nil {
		return nil, nil, err
	}
	return contract, rev, nil
}

func (s *contractService) GetContract(ctx context.Context, tenantID, contractID string) (*entity.DataContract, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	return s.contracts.GetContract(ctx, tenantID, contractID)
}

func (s *contractService) ListContracts(ctx context.Context, tenantID, businessID string) ([]*entity.DataContract, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	return s.contracts.ListContracts(ctx, tenantID, businessID)
}

func (s *contractService) GetRevision(ctx context.Context, tenantID, revisionID string) (*entity.DataContractRevision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	return s.contracts.GetRevision(ctx, tenantID, revisionID)
}

func (s *contractService) ListRevisions(ctx context.Context, tenantID, contractID string) ([]*entity.DataContractRevision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	return s.contracts.ListRevisions(ctx, tenantID, contractID)
}

func (s *contractService) CreateRevision(ctx context.Context, in *CreateRevisionInput) (*entity.DataContractRevision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	if in == nil || strings.TrimSpace(in.TenantID) == "" || strings.TrimSpace(in.ContractID) == "" || strings.TrimSpace(in.BaseRevisionID) == "" || in.BusinessModelRevision <= 0 || strings.TrimSpace(in.Name) == "" {
		return nil, entity.ErrContractInvalidPayload
	}
	if err := ValidateRequirementIDsUnique(in.RequirementIDs); err != nil {
		return nil, err
	}
	if err := ValidateQueryCapabilities(in.QueryCapabilities); err != nil {
		return nil, err
	}
	contract, err := s.contracts.GetContract(ctx, in.TenantID, in.ContractID)
	if err != nil {
		return nil, err
	}
	base, err := s.contracts.GetRevision(ctx, in.TenantID, in.BaseRevisionID)
	if err != nil {
		return nil, err
	}
	if base.ContractID != contract.ContractID || base.BusinessID != contract.BusinessID {
		return nil, entity.ErrContractInvalidPayload
	}
	bindings, err := s.materializeBindings(ctx, in.TenantID, contract.BusinessID, in.BusinessModelRevision, in.MappingIDs)
	if err != nil {
		return nil, err
	}
	var out *entity.DataContractRevision
	err = s.contracts.Transaction(ctx, func(tx repository.ContractRepository) error {
		version, err := tx.AllocateNextVersion(ctx, in.TenantID, in.ContractID)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		rev := &entity.DataContractRevision{
			RevisionID: newID("crev"), TenantID: in.TenantID, BusinessID: contract.BusinessID, ContractID: in.ContractID,
			Version: version, Status: entity.ContractStatusDraft, BusinessModelRevision: in.BusinessModelRevision,
			Name: in.Name, Description: in.Description,
			RequirementIDs: append([]string(nil), in.RequirementIDs...),
			LogicalSchema: entity.ContractLogicalSchema{Fields: append([]entity.LogicalField(nil), in.LogicalSchema.Fields...)},
			QueryCapabilities: append([]entity.QueryCapability(nil), in.QueryCapabilities...),
			FilterSchema: in.FilterSchema, SortSchema: in.SortSchema, PaginationPolicy: in.PaginationPolicy,
			FreshnessPolicy: in.FreshnessPolicy, ClassificationPolicy: cloneClassPolicy(in.ClassificationPolicy),
			BindingRefs: bindings, AccessPolicyRef: in.AccessPolicyRef, DerivedFromRevisionID: in.BaseRevisionID,
			CreatedBy: in.ActorID, CreatedAt: now, UpdatedAt: now,
		}
		if err := tx.CreateRevision(ctx, rev); err != nil {
			return err
		}
		out = rev
		return nil
	})
	return out, err
}

func (s *contractService) materializeBindings(ctx context.Context, tenantID, businessID string, bmRev int32, mappingIDs []string) ([]entity.ContractBinding, error) {
	if len(mappingIDs) == 0 {
		return nil, fmt.Errorf("%w: mapping_ids required", entity.ErrContractBindingInvalid)
	}
	seen := map[string]struct{}{}
	out := make([]entity.ContractBinding, 0, len(mappingIDs))
	for _, id := range mappingIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			return nil, entity.ErrContractBindingInvalid
		}
		if _, ok := seen[id]; ok {
			return nil, fmt.Errorf("%w: duplicate mapping_id %q", entity.ErrContractBindingInvalid, id)
		}
		seen[id] = struct{}{}
		m, err := s.mappings.GetMapping(ctx, tenantID, id)
		if err != nil {
			return nil, err
		}
		if m.Status != entity.MappingStatusConfirmed {
			return nil, fmt.Errorf("%w: mapping %q is not confirmed", entity.ErrContractBindingInvalid, id)
		}
		if m.BusinessID != businessID || m.BusinessModelRevision != bmRev {
			return nil, fmt.Errorf("%w: mapping %q lineage mismatch", entity.ErrContractBindingInvalid, id)
		}
		out = append(out, entity.ContractBinding{
			RequirementID: m.RequirementID, MappingID: m.MappingID,
			SourceID: m.SourceID, ConnectionID: m.ConnectionID, AssetID: m.AssetID, SchemaSnapshotID: m.SchemaSnapshotID,
		})
	}
	return out, nil
}

func cloneClassPolicy(in map[string]entity.DataClassification) map[string]entity.DataClassification {
	if in == nil {
		return map[string]entity.DataClassification{}
	}
	out := make(map[string]entity.DataClassification, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func (s *contractService) ValidateRevision(ctx context.Context, tenantID, revisionID, actorID string) (*entity.DataContractRevision, *entity.DataValidationResult, error) {
	if !s.configured() {
		return nil, nil, entity.ErrNotConfigured
	}
	rev, err := s.contracts.GetRevision(ctx, tenantID, revisionID)
	if err != nil {
		return nil, nil, err
	}
	if rev.Status != entity.ContractStatusDraft {
		return nil, nil, entity.ErrContractInvalidState
	}

	issues := []entity.ValidationIssue{}
	fingerprints := map[string]string{}
	if err := ValidateRequirementIDsUnique(rev.RequirementIDs); err != nil {
		issues = append(issues, entity.ValidationIssue{Code: "REQUIREMENT_IDS_DUPLICATE", Message: err.Error()})
	}
	reqSet := map[string]struct{}{}
	for _, id := range rev.RequirementIDs {
		reqSet[strings.TrimSpace(id)] = struct{}{}
	}

	if _, _, err := s.business.GetRevision(ctx, tenantID, rev.BusinessID, rev.BusinessModelRevision); err != nil {
		issues = append(issues, entity.ValidationIssue{Code: "BUSINESS_REVISION", Message: "pinned business model revision not found"})
	}

	for _, reqID := range rev.RequirementIDs {
		r, err := s.data.GetRequirement(ctx, tenantID, reqID)
		if err != nil || r.Status != entity.StatusConfirmed || r.BusinessID != rev.BusinessID || r.BusinessModelRevision != rev.BusinessModelRevision {
			issues = append(issues, entity.ValidationIssue{Code: "REQUIREMENT", Message: fmt.Sprintf("requirement %q must be confirmed at pinned business revision", reqID)})
		}
	}

	bindingByReq := map[string]entity.ContractBinding{}
	bindingCount := map[string]int{}
	for _, b := range rev.BindingRefs {
		bindingCount[b.RequirementID]++
		bindingByReq[b.RequirementID] = b
		m, err := s.mappings.GetMapping(ctx, tenantID, b.MappingID)
		if err != nil || m.Status != entity.MappingStatusConfirmed {
			issues = append(issues, entity.ValidationIssue{Code: "BINDING_MAPPING", Message: fmt.Sprintf("binding mapping %q must be confirmed", b.MappingID)})
			continue
		}
		if m.BusinessID != rev.BusinessID || m.BusinessModelRevision != rev.BusinessModelRevision ||
			m.RequirementID != b.RequirementID || m.SourceID != b.SourceID || m.ConnectionID != b.ConnectionID ||
			m.AssetID != b.AssetID || m.SchemaSnapshotID != b.SchemaSnapshotID {
			issues = append(issues, entity.ValidationIssue{Code: "BINDING_LINEAGE", Message: fmt.Sprintf("binding %q lineage mismatch", b.MappingID)})
		}
		snap, err := s.sources.GetSnapshot(ctx, tenantID, b.SchemaSnapshotID)
		if err != nil {
			issues = append(issues, entity.ValidationIssue{Code: "SNAPSHOT", Message: fmt.Sprintf("schema snapshot %q not found", b.SchemaSnapshotID)})
			continue
		}
		fingerprints[b.SchemaSnapshotID] = snap.Fingerprint
		var schema entity.PhysicalSchema
		if json.Unmarshal([]byte(snap.SchemaJSON), &schema) != nil {
			issues = append(issues, entity.ValidationIssue{Code: "SCHEMA_JSON", Message: fmt.Sprintf("schema snapshot %q invalid JSON", b.SchemaSnapshotID)})
			continue
		}
		if err := ValidateSchemaPaths(&schema); err != nil {
			issues = append(issues, entity.ValidationIssue{Code: "SCHEMA_PATHS", Message: err.Error()})
		}
		if err := ValidateMappingTarget(m.TargetFieldPaths, &schema); err != nil {
			issues = append(issues, entity.ValidationIssue{Code: "MAPPING_TARGET", Message: err.Error()})
		}
		if err := ValidateTransformSpec(m.MappingType, m.TransformSpec, m.TargetFieldPaths, &schema); err != nil {
			issues = append(issues, entity.ValidationIssue{Code: "TRANSFORM", Message: err.Error()})
		}
	}
	for reqID := range reqSet {
		if bindingCount[reqID] != 1 {
			issues = append(issues, entity.ValidationIssue{Code: "BINDING_COUNT", Message: fmt.Sprintf("requirement %q must have exactly one binding", reqID)})
		}
	}
	for reqID := range bindingCount {
		if _, ok := reqSet[reqID]; !ok {
			issues = append(issues, entity.ValidationIssue{Code: "BINDING_ORPHAN", Message: fmt.Sprintf("binding requirement %q not in requirement_ids", reqID)})
		}
	}

	if err := ValidateLogicalSchema(rev.LogicalSchema, reqSet); err != nil {
		issues = append(issues, entity.ValidationIssue{Code: "LOGICAL_SCHEMA", Message: err.Error()})
	} else {
		fieldReqs := map[string]struct{}{}
		for _, f := range rev.LogicalSchema.Fields {
			fieldReqs[strings.TrimSpace(f.RequirementID)] = struct{}{}
		}
		for reqID := range reqSet {
			if _, ok := fieldReqs[reqID]; !ok {
				issues = append(issues, entity.ValidationIssue{Code: "LOGICAL_COVERAGE", Message: fmt.Sprintf("requirement %q missing logical field", reqID)})
			}
		}
	}
	if err := ValidateQueryCapabilities(rev.QueryCapabilities); err != nil {
		issues = append(issues, entity.ValidationIssue{Code: "QUERY_CAPABILITIES", Message: err.Error()})
	}
	if err := ValidateFilterSchema(rev.FilterSchema, rev.LogicalSchema); err != nil {
		issues = append(issues, entity.ValidationIssue{Code: "FILTER_SCHEMA", Message: err.Error()})
	}
	if err := ValidateSortSchema(rev.SortSchema, rev.LogicalSchema); err != nil {
		issues = append(issues, entity.ValidationIssue{Code: "SORT_SCHEMA", Message: err.Error()})
	}
	if err := ValidatePaginationPolicy(rev.PaginationPolicy); err != nil {
		issues = append(issues, entity.ValidationIssue{Code: "PAGINATION", Message: err.Error()})
	}
	if err := ValidateFreshnessPolicy(rev.FreshnessPolicy); err != nil {
		issues = append(issues, entity.ValidationIssue{Code: "FRESHNESS", Message: err.Error()})
	}
	if err := ValidateClassification(rev.ClassificationPolicy, rev.LogicalSchema); err != nil {
		issues = append(issues, entity.ValidationIssue{Code: "CLASSIFICATION", Message: err.Error()})
	}

	// Type + nullability guarantees against resolved mapping output.
	for _, field := range rev.LogicalSchema.Fields {
		binding, ok := bindingByReq[strings.TrimSpace(field.RequirementID)]
		if !ok {
			continue
		}
		m, err := s.mappings.GetMapping(ctx, tenantID, binding.MappingID)
		if err != nil {
			issues = append(issues, entity.ValidationIssue{Code: "BINDING_MAPPING", Message: err.Error()})
			continue
		}
		snap, err := s.sources.GetSnapshot(ctx, tenantID, binding.SchemaSnapshotID)
		if err != nil {
			continue
		}
		var schema entity.PhysicalSchema
		if json.Unmarshal([]byte(snap.SchemaJSON), &schema) != nil {
			continue
		}
		resolved, err := ResolveMappingOutputContractType(m, &schema)
		if err != nil {
			issues = append(issues, entity.ValidationIssue{Code: "LOGICAL_TYPE_MISMATCH", Message: err.Error()})
			continue
		}
		if resolved != normalizedType(field.LogicalType) {
			issues = append(issues, entity.ValidationIssue{
				Code:    "LOGICAL_TYPE_MISMATCH",
				Message: fmt.Sprintf("field %q resolved type %q != logical %q", field.LogicalKey, resolved, field.LogicalType),
			})
		}
		if !field.Nullable {
			index := buildFieldPathIndex(&schema)
			for _, path := range m.TargetFieldPaths {
				pf, ok := index[strings.TrimSpace(path)]
				if !ok {
					continue
				}
				if pf.Nullable {
					issues = append(issues, entity.ValidationIssue{
						Code:    "NULLABILITY_GUARANTEE_LOST",
						Message: fmt.Sprintf("field %q requires non-null but physical path %q is nullable", field.LogicalKey, path),
					})
				}
			}
		}
	}

	now := time.Now().UTC()
	status := entity.ValidationStatusPass
	action := entity.LifecycleActionValidatePass
	pass := len(issues) == 0
	if !pass {
		status = entity.ValidationStatusFail
		action = entity.LifecycleActionValidateFail
	}
	result := &entity.DataValidationResult{
		ValidationID: newID("cval"), TenantID: tenantID, BusinessID: rev.BusinessID,
		ContractID: rev.ContractID, RevisionID: rev.RevisionID, Version: rev.Version,
		Status: status, Errors: issues, Warnings: []entity.ValidationIssue{},
		SnapshotFingerprints: fingerprints, ValidatedBy: actorID, ValidatedAt: now, CreatedAt: now,
	}
	event := &entity.DataContractLifecycleEvent{
		EventID: newID("clife"), TenantID: tenantID, BusinessID: rev.BusinessID,
		ContractID: rev.ContractID, RevisionID: rev.RevisionID, Version: rev.Version,
		Action: action, ActorPrincipalID: actorID, Reason: string(status), CreatedAt: now,
	}

	err = s.contracts.Transaction(ctx, func(tx repository.ContractRepository) error {
		if pass {
			if err := tx.UpdateRevisionStatus(ctx, tenantID, revisionID, entity.ContractStatusDraft, entity.ContractStatusValidated); err != nil {
				return err
			}
			rev.Status = entity.ContractStatusValidated
			rev.UpdatedAt = now
		}
		if err := tx.CreateValidationResult(ctx, result); err != nil {
			return err
		}
		return tx.CreateLifecycleEvent(ctx, event)
	})
	if err != nil {
		return nil, nil, err
	}
	return rev, result, nil
}

func (s *contractService) ActivateRevision(ctx context.Context, tenantID, revisionID, actorID, reason string) (*entity.DataContractRevision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	var out *entity.DataContractRevision
	err := s.contracts.Transaction(ctx, func(tx repository.ContractRepository) error {
		rev, err := tx.GetRevision(ctx, tenantID, revisionID)
		if err != nil {
			return err
		}
		// Serialize activates on the contract row lock first.
		contract, err := tx.GetContractForUpdate(ctx, tenantID, rev.ContractID)
		if err != nil {
			return err
		}
		rev, err = tx.GetRevision(ctx, tenantID, revisionID)
		if err != nil {
			return err
		}
		if rev.Status != entity.ContractStatusValidated {
			return entity.ErrContractInvalidState
		}
		if err := tx.UpdateRevisionStatus(ctx, tenantID, revisionID, entity.ContractStatusValidated, entity.ContractStatusActive); err != nil {
			if errors.Is(err, entity.ErrContractInvalidState) {
				return entity.ErrContractVersionConflict
			}
			return err
		}
		now := time.Now().UTC()
		if contract.ActiveRevisionID != "" && contract.ActiveRevisionID != revisionID {
			prev, err := tx.GetRevision(ctx, tenantID, contract.ActiveRevisionID)
			if err != nil {
				return err
			}
			if prev.Status == entity.ContractStatusActive {
				if err := tx.UpdateRevisionStatus(ctx, tenantID, prev.RevisionID, entity.ContractStatusActive, entity.ContractStatusDeprecated); err != nil {
					return err
				}
				if err := tx.CreateLifecycleEvent(ctx, &entity.DataContractLifecycleEvent{
					EventID: newID("clife"), TenantID: tenantID, BusinessID: rev.BusinessID,
					ContractID: rev.ContractID, RevisionID: prev.RevisionID, Version: prev.Version,
					Action: entity.LifecycleActionDeprecate, ActorPrincipalID: actorID,
					Reason: "superseded by " + revisionID, CreatedAt: now,
				}); err != nil {
					return err
				}
			}
		}
		if err := tx.UpdateContractActiveRevision(ctx, tenantID, rev.ContractID, revisionID); err != nil {
			return err
		}
		if err := tx.CreateLifecycleEvent(ctx, &entity.DataContractLifecycleEvent{
			EventID: newID("clife"), TenantID: tenantID, BusinessID: rev.BusinessID,
			ContractID: rev.ContractID, RevisionID: rev.RevisionID, Version: rev.Version,
			Action: entity.LifecycleActionActivate, ActorPrincipalID: actorID, Reason: reason, CreatedAt: now,
		}); err != nil {
			return err
		}
		rev.Status = entity.ContractStatusActive
		rev.UpdatedAt = now
		out = rev
		return nil
	})
	return out, err
}

func (s *contractService) DeprecateRevision(ctx context.Context, tenantID, revisionID, actorID, reason string) (*entity.DataContractRevision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	var out *entity.DataContractRevision
	err := s.contracts.Transaction(ctx, func(tx repository.ContractRepository) error {
		rev, err := tx.GetRevision(ctx, tenantID, revisionID)
		if err != nil {
			return err
		}
		// Lock contract root first so pointer decisions serialize with Activate/Drift.
		contract, err := tx.GetContractForUpdate(ctx, tenantID, rev.ContractID)
		if err != nil {
			return err
		}
		rev, err = tx.GetRevision(ctx, tenantID, revisionID)
		if err != nil {
			return err
		}
		from := rev.Status
		switch from {
		case entity.ContractStatusActive:
			// ACTIVE owns the pointer; mismatch is an invariant violation.
			if contract.ActiveRevisionID != revisionID {
				return entity.ErrContractInvalidState
			}
			if err := tx.UpdateRevisionStatus(ctx, tenantID, revisionID, from, entity.ContractStatusDeprecated); err != nil {
				return err
			}
			if err := tx.ClearContractActiveRevisionIfMatch(ctx, tenantID, rev.ContractID, revisionID); err != nil {
				return err
			}
		case entity.ContractStatusStale:
			// Historical STALE never owns another revision's ACTIVE pointer.
			if err := tx.UpdateRevisionStatus(ctx, tenantID, revisionID, from, entity.ContractStatusDeprecated); err != nil {
				return err
			}
			// CASE A: pointer empty → no clear needed.
			// CASE B: pointer points at another ACTIVE → leave untouched.
			// CASE C: legacy inconsistent pointer == this STALE revision → clear safely.
			if contract.ActiveRevisionID == revisionID {
				if err := tx.ClearContractActiveRevisionIfMatch(ctx, tenantID, rev.ContractID, revisionID); err != nil {
					return err
				}
			}
		default:
			return entity.ErrContractInvalidState
		}
		now := time.Now().UTC()
		if err := tx.CreateLifecycleEvent(ctx, &entity.DataContractLifecycleEvent{
			EventID: newID("clife"), TenantID: tenantID, BusinessID: rev.BusinessID,
			ContractID: rev.ContractID, RevisionID: rev.RevisionID, Version: rev.Version,
			Action: entity.LifecycleActionDeprecate, ActorPrincipalID: actorID, Reason: reason, CreatedAt: now,
		}); err != nil {
			return err
		}
		rev.Status = entity.ContractStatusDeprecated
		rev.UpdatedAt = now
		out = rev
		return nil
	})
	return out, err
}

func (s *contractService) ListValidationResults(ctx context.Context, tenantID, revisionID string) ([]*entity.DataValidationResult, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	return s.contracts.ListValidationResults(ctx, tenantID, revisionID)
}

func (s *contractService) ListDriftResults(ctx context.Context, tenantID, revisionID string) ([]*entity.DataDriftResult, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	return s.contracts.ListDriftResults(ctx, tenantID, revisionID)
}

func (s *contractService) ListGapResults(ctx context.Context, tenantID, revisionID string) ([]*entity.DataContractGapResult, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	return s.contracts.ListGapResults(ctx, tenantID, revisionID)
}

func (s *contractService) ListLifecycleEvents(ctx context.Context, tenantID, contractID string) ([]*entity.DataContractLifecycleEvent, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	return s.contracts.ListLifecycleEvents(ctx, tenantID, contractID)
}

func (s *contractService) EvaluateBusinessGap(ctx context.Context, in *EvaluateGapInput) (*entity.DataContractGapResult, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	if in == nil || strings.TrimSpace(in.TenantID) == "" || strings.TrimSpace(in.RevisionID) == "" {
		return nil, entity.ErrContractGapInvalid
	}
	rev, err := s.contracts.GetRevision(ctx, in.TenantID, in.RevisionID)
	if err != nil {
		return nil, err
	}
	model, _, _, err := s.business.GetModel(ctx, in.TenantID, rev.BusinessID)
	if err != nil || model == nil {
		return nil, entity.ErrBusinessRevisionNotFound
	}
	currentRev := model.CurrentRevision
	now := time.Now().UTC()
	result := &entity.DataContractGapResult{
		GapResultID: newID("cgap"), TenantID: in.TenantID, BusinessID: rev.BusinessID,
		ContractID: rev.ContractID, RevisionID: rev.RevisionID, Version: rev.Version,
		FromBusinessRevision: rev.BusinessModelRevision, CurrentBusinessRevision: currentRev,
		NewConfirmedRequirementIDs: []string{}, UnmappedRequirementIDs: []string{},
		GapStatus: "NONE", EvaluatedBy: in.ActorID, EvaluatedAt: now, CreatedAt: now,
	}
	if currentRev == rev.BusinessModelRevision {
		if err := s.contracts.CreateGapResult(ctx, result); err != nil {
			return nil, err
		}
		return result, nil
	}
	reqs, err := s.data.ListRequirementsByRevision(ctx, in.TenantID, rev.BusinessID, currentRev)
	if err != nil {
		return nil, err
	}
	pinned := map[string]struct{}{}
	for _, id := range rev.RequirementIDs {
		pinned[id] = struct{}{}
	}
	confirmedMappings, err := s.mappings.ListMappings(ctx, in.TenantID, rev.BusinessID, currentRev, entity.MappingStatusConfirmed)
	if err != nil {
		return nil, err
	}
	mappedReqs := map[string]struct{}{}
	for _, m := range confirmedMappings {
		mappedReqs[m.RequirementID] = struct{}{}
	}
	for _, r := range reqs {
		if r.Status != entity.StatusConfirmed {
			continue
		}
		if _, ok := pinned[r.RequirementID]; !ok {
			result.NewConfirmedRequirementIDs = append(result.NewConfirmedRequirementIDs, r.RequirementID)
		}
		if _, ok := mappedReqs[r.RequirementID]; !ok {
			result.UnmappedRequirementIDs = append(result.UnmappedRequirementIDs, r.RequirementID)
		}
	}
	result.GapStatus = "GAP"
	if err := s.contracts.CreateGapResult(ctx, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *contractService) BuildContractDescriptor(rev *entity.DataContractRevision) *entity.DataContractDescriptor {
	if rev == nil {
		return nil
	}
	classification := map[string]entity.DataClassification{}
	for _, field := range rev.LogicalSchema.Fields {
		key := strings.TrimSpace(field.LogicalKey)
		if key == "" {
			continue
		}
		classification[key] = field.Classification
	}
	for key, class := range rev.ClassificationPolicy {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		classification[key] = class
	}
	return &entity.DataContractDescriptor{
		ContractID: rev.ContractID, RevisionID: rev.RevisionID, Version: rev.Version,
		BusinessModelRevision: rev.BusinessModelRevision,
		LogicalSchema: entity.ContractLogicalSchema{Fields: append([]entity.LogicalField(nil), rev.LogicalSchema.Fields...)},
		QueryCapabilities: append([]entity.QueryCapability(nil), rev.QueryCapabilities...),
		FilterSchema: rev.FilterSchema, SortSchema: rev.SortSchema, PaginationPolicy: rev.PaginationPolicy,
		FreshnessPolicy: rev.FreshnessPolicy, Classification: classification,
		AccessPolicyRef: rev.AccessPolicyRef, Status: rev.Status,
	}
}

func (s *contractService) GetActiveContractDescriptor(ctx context.Context, tenantID, contractID string) (*entity.DataContractDescriptor, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	contract, err := s.contracts.GetContract(ctx, tenantID, contractID)
	if err != nil {
		return nil, err
	}
	if contract.ActiveRevisionID == "" {
		return nil, entity.ErrContractNotActive
	}
	rev, err := s.contracts.GetRevision(ctx, tenantID, contract.ActiveRevisionID)
	if err != nil {
		return nil, err
	}
	if rev.Status != entity.ContractStatusActive {
		return nil, entity.ErrContractNotActive
	}
	return s.BuildContractDescriptor(rev), nil
}
