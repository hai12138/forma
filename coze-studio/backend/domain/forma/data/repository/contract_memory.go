/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package repository

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
)

type memoryContractRepo struct {
	mu         sync.Mutex
	contracts  map[string]*entity.DataContract
	revisions  map[string]*entity.DataContractRevision
	validations map[string]*entity.DataValidationResult
	events     map[string]*entity.DataContractLifecycleEvent
	drifts     map[string]*entity.DataDriftResult
	gaps       map[string]*entity.DataContractGapResult
}

type memoryContractTx struct{ repo *memoryContractRepo }

func NewMemoryContractRepository() ContractRepository {
	return &memoryContractRepo{
		contracts:   map[string]*entity.DataContract{},
		revisions:   map[string]*entity.DataContractRevision{},
		validations: map[string]*entity.DataValidationResult{},
		events:      map[string]*entity.DataContractLifecycleEvent{},
		drifts:      map[string]*entity.DataDriftResult{},
		gaps:        map[string]*entity.DataContractGapResult{},
	}
}

func contractKey(t, id string) string { return t + "\x00" + id }

func cloneContract(v *entity.DataContract) *entity.DataContract {
	if v == nil {
		return nil
	}
	x := *v
	return &x
}

func cloneRevision(v *entity.DataContractRevision) *entity.DataContractRevision {
	if v == nil {
		return nil
	}
	x := *v
	x.RequirementIDs = append([]string(nil), v.RequirementIDs...)
	x.LogicalSchema.Fields = append([]entity.LogicalField(nil), v.LogicalSchema.Fields...)
	x.QueryCapabilities = append([]entity.QueryCapability(nil), v.QueryCapabilities...)
	x.FilterSchema.Fields = make([]entity.FilterFieldSpec, len(v.FilterSchema.Fields))
	for i, f := range v.FilterSchema.Fields {
		x.FilterSchema.Fields[i] = entity.FilterFieldSpec{
			LogicalKey: f.LogicalKey,
			Operators:  append([]entity.FilterOperator(nil), f.Operators...),
		}
	}
	x.SortSchema.Fields = make([]entity.SortFieldSpec, len(v.SortSchema.Fields))
	for i, f := range v.SortSchema.Fields {
		x.SortSchema.Fields[i] = entity.SortFieldSpec{
			LogicalKey: f.LogicalKey,
			Directions:  append([]entity.SortDirection(nil), f.Directions...),
		}
	}
	if v.ClassificationPolicy != nil {
		x.ClassificationPolicy = make(map[string]entity.DataClassification, len(v.ClassificationPolicy))
		for k, val := range v.ClassificationPolicy {
			x.ClassificationPolicy[k] = val
		}
	} else {
		x.ClassificationPolicy = map[string]entity.DataClassification{}
	}
	x.BindingRefs = append([]entity.ContractBinding(nil), v.BindingRefs...)
	return &x
}

func cloneValidation(v *entity.DataValidationResult) *entity.DataValidationResult {
	if v == nil {
		return nil
	}
	x := *v
	x.Errors = append([]entity.ValidationIssue(nil), v.Errors...)
	x.Warnings = append([]entity.ValidationIssue(nil), v.Warnings...)
	if v.SnapshotFingerprints != nil {
		x.SnapshotFingerprints = make(map[string]string, len(v.SnapshotFingerprints))
		for k, val := range v.SnapshotFingerprints {
			x.SnapshotFingerprints[k] = val
		}
	} else {
		x.SnapshotFingerprints = map[string]string{}
	}
	return &x
}

func cloneLifecycle(v *entity.DataContractLifecycleEvent) *entity.DataContractLifecycleEvent {
	if v == nil {
		return nil
	}
	x := *v
	return &x
}

func cloneDrift(v *entity.DataDriftResult) *entity.DataDriftResult {
	if v == nil {
		return nil
	}
	x := *v
	x.Findings = append([]entity.DriftFinding(nil), v.Findings...)
	if v.ComparedSnapshotIDs != nil {
		x.ComparedSnapshotIDs = make(map[string]string, len(v.ComparedSnapshotIDs))
		for k, val := range v.ComparedSnapshotIDs {
			x.ComparedSnapshotIDs[k] = val
		}
	} else {
		x.ComparedSnapshotIDs = map[string]string{}
	}
	return &x
}

func cloneGap(v *entity.DataContractGapResult) *entity.DataContractGapResult {
	if v == nil {
		return nil
	}
	x := *v
	x.NewConfirmedRequirementIDs = append([]string(nil), v.NewConfirmedRequirementIDs...)
	x.UnmappedRequirementIDs = append([]string(nil), v.UnmappedRequirementIDs...)
	return &x
}

type contractSnapshot struct {
	contracts   map[string]*entity.DataContract
	revisions   map[string]*entity.DataContractRevision
	validations map[string]*entity.DataValidationResult
	events      map[string]*entity.DataContractLifecycleEvent
	drifts      map[string]*entity.DataDriftResult
	gaps        map[string]*entity.DataContractGapResult
}

func (r *memoryContractRepo) snapshot() contractSnapshot {
	s := contractSnapshot{
		contracts: map[string]*entity.DataContract{}, revisions: map[string]*entity.DataContractRevision{},
		validations: map[string]*entity.DataValidationResult{}, events: map[string]*entity.DataContractLifecycleEvent{},
		drifts: map[string]*entity.DataDriftResult{}, gaps: map[string]*entity.DataContractGapResult{},
	}
	for k, v := range r.contracts {
		s.contracts[k] = cloneContract(v)
	}
	for k, v := range r.revisions {
		s.revisions[k] = cloneRevision(v)
	}
	for k, v := range r.validations {
		s.validations[k] = cloneValidation(v)
	}
	for k, v := range r.events {
		s.events[k] = cloneLifecycle(v)
	}
	for k, v := range r.drifts {
		s.drifts[k] = cloneDrift(v)
	}
	for k, v := range r.gaps {
		s.gaps[k] = cloneGap(v)
	}
	return s
}

func (r *memoryContractRepo) Transaction(c context.Context, fn func(ContractRepository) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.snapshot()
	if err := fn(&memoryContractTx{r}); err != nil {
		r.contracts = s.contracts
		r.revisions = s.revisions
		r.validations = s.validations
		r.events = s.events
		r.drifts = s.drifts
		r.gaps = s.gaps
		return err
	}
	return nil
}

func (r *memoryContractRepo) CreateContract(c context.Context, v *entity.DataContract) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.createContract(c, v)
}
func (r *memoryContractRepo) createContract(_ context.Context, v *entity.DataContract) error {
	if v == nil {
		return entity.ErrContractInvalidPayload
	}
	k := contractKey(v.TenantID, v.ContractID)
	if _, ok := r.contracts[k]; ok {
		return entity.ErrContractInvalidState
	}
	r.contracts[k] = cloneContract(v)
	return nil
}
func (r *memoryContractRepo) GetContract(c context.Context, t, id string) (*entity.DataContract, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getContract(c, t, id)
}
func (r *memoryContractRepo) getContract(_ context.Context, t, id string) (*entity.DataContract, error) {
	v := r.contracts[contractKey(t, id)]
	if v == nil {
		return nil, entity.ErrContractNotFound
	}
	return cloneContract(v), nil
}
func (r *memoryContractRepo) ListContracts(c context.Context, t, b string) ([]*entity.DataContract, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listContracts(c, t, b)
}
func (r *memoryContractRepo) listContracts(_ context.Context, t, b string) ([]*entity.DataContract, error) {
	out := []*entity.DataContract{}
	for _, v := range r.contracts {
		if v.TenantID == t && v.BusinessID == b {
			out = append(out, cloneContract(v))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}
func (r *memoryContractRepo) UpdateContractActiveRevision(c context.Context, t, contractID, revisionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.updateActiveRevision(c, t, contractID, revisionID)
}
func (r *memoryContractRepo) updateActiveRevision(_ context.Context, t, contractID, revisionID string) error {
	v := r.contracts[contractKey(t, contractID)]
	if v == nil {
		return entity.ErrContractNotFound
	}
	v.ActiveRevisionID = revisionID
	v.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *memoryContractRepo) CreateRevision(c context.Context, v *entity.DataContractRevision) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.createRevision(c, v)
}
func (r *memoryContractRepo) createRevision(_ context.Context, v *entity.DataContractRevision) error {
	if v == nil {
		return entity.ErrContractInvalidPayload
	}
	k := contractKey(v.TenantID, v.RevisionID)
	if _, ok := r.revisions[k]; ok {
		return entity.ErrContractInvalidState
	}
	for _, existing := range r.revisions {
		if existing.TenantID == v.TenantID && existing.ContractID == v.ContractID && existing.Version == v.Version {
			return entity.ErrContractVersionConflict
		}
	}
	r.revisions[k] = cloneRevision(v)
	return nil
}
func (r *memoryContractRepo) GetRevision(c context.Context, t, id string) (*entity.DataContractRevision, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getRevision(c, t, id)
}
func (r *memoryContractRepo) getRevision(_ context.Context, t, id string) (*entity.DataContractRevision, error) {
	v := r.revisions[contractKey(t, id)]
	if v == nil {
		return nil, entity.ErrContractRevisionNotFound
	}
	return cloneRevision(v), nil
}
func (r *memoryContractRepo) ListRevisions(c context.Context, t, contractID string) ([]*entity.DataContractRevision, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listRevisions(c, t, contractID)
}
func (r *memoryContractRepo) listRevisions(_ context.Context, t, contractID string) ([]*entity.DataContractRevision, error) {
	out := []*entity.DataContractRevision{}
	for _, v := range r.revisions {
		if v.TenantID == t && v.ContractID == contractID {
			out = append(out, cloneRevision(v))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
}
func (r *memoryContractRepo) GetRevisionByVersion(c context.Context, t, contractID string, version int32) (*entity.DataContractRevision, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getRevisionByVersion(c, t, contractID, version)
}
func (r *memoryContractRepo) getRevisionByVersion(_ context.Context, t, contractID string, version int32) (*entity.DataContractRevision, error) {
	for _, v := range r.revisions {
		if v.TenantID == t && v.ContractID == contractID && v.Version == version {
			return cloneRevision(v), nil
		}
	}
	return nil, entity.ErrContractRevisionNotFound
}
func (r *memoryContractRepo) AllocateNextVersion(c context.Context, t, contractID string) (int32, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.allocateNextVersion(c, t, contractID)
}
func (r *memoryContractRepo) allocateNextVersion(_ context.Context, t, contractID string) (int32, error) {
	if r.contracts[contractKey(t, contractID)] == nil {
		return 0, entity.ErrContractNotFound
	}
	var max int32
	for _, v := range r.revisions {
		if v.TenantID == t && v.ContractID == contractID && v.Version > max {
			max = v.Version
		}
	}
	return max + 1, nil
}
func (r *memoryContractRepo) UpdateRevisionStatus(c context.Context, t, revisionID string, from, to entity.ContractStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.updateRevisionStatus(c, t, revisionID, from, to)
}
func (r *memoryContractRepo) updateRevisionStatus(_ context.Context, t, revisionID string, from, to entity.ContractStatus) error {
	v := r.revisions[contractKey(t, revisionID)]
	if v == nil || v.Status != from {
		return entity.ErrContractInvalidState
	}
	v.Status = to
	v.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *memoryContractRepo) CreateValidationResult(c context.Context, v *entity.DataValidationResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.createValidation(c, v)
}
func (r *memoryContractRepo) createValidation(_ context.Context, v *entity.DataValidationResult) error {
	if v == nil {
		return entity.ErrContractInvalidPayload
	}
	k := contractKey(v.TenantID, v.ValidationID)
	if _, ok := r.validations[k]; ok {
		return entity.ErrContractInvalidState
	}
	r.validations[k] = cloneValidation(v)
	return nil
}
func (r *memoryContractRepo) ListValidationResults(c context.Context, t, revisionID string) ([]*entity.DataValidationResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listValidations(c, t, revisionID)
}
func (r *memoryContractRepo) listValidations(_ context.Context, t, revisionID string) ([]*entity.DataValidationResult, error) {
	out := []*entity.DataValidationResult{}
	for _, v := range r.validations {
		if v.TenantID == t && v.RevisionID == revisionID {
			out = append(out, cloneValidation(v))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}
func (r *memoryContractRepo) GetValidationResult(c context.Context, t, id string) (*entity.DataValidationResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getValidation(c, t, id)
}
func (r *memoryContractRepo) getValidation(_ context.Context, t, id string) (*entity.DataValidationResult, error) {
	v := r.validations[contractKey(t, id)]
	if v == nil {
		return nil, entity.ErrContractValidationFailed
	}
	return cloneValidation(v), nil
}

func (r *memoryContractRepo) CreateLifecycleEvent(c context.Context, v *entity.DataContractLifecycleEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.createLifecycle(c, v)
}
func (r *memoryContractRepo) createLifecycle(_ context.Context, v *entity.DataContractLifecycleEvent) error {
	if v == nil {
		return entity.ErrContractInvalidPayload
	}
	k := contractKey(v.TenantID, v.EventID)
	if _, ok := r.events[k]; ok {
		return entity.ErrContractInvalidState
	}
	r.events[k] = cloneLifecycle(v)
	return nil
}
func (r *memoryContractRepo) ListLifecycleEvents(c context.Context, t, contractID string) ([]*entity.DataContractLifecycleEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listLifecycle(c, t, contractID)
}
func (r *memoryContractRepo) listLifecycle(_ context.Context, t, contractID string) ([]*entity.DataContractLifecycleEvent, error) {
	out := []*entity.DataContractLifecycleEvent{}
	for _, v := range r.events {
		if v.TenantID == t && v.ContractID == contractID {
			out = append(out, cloneLifecycle(v))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (r *memoryContractRepo) CreateDriftResult(c context.Context, v *entity.DataDriftResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.createDrift(c, v)
}
func (r *memoryContractRepo) createDrift(_ context.Context, v *entity.DataDriftResult) error {
	if v == nil {
		return entity.ErrContractInvalidPayload
	}
	k := contractKey(v.TenantID, v.DriftResultID)
	if _, ok := r.drifts[k]; ok {
		return entity.ErrContractInvalidState
	}
	r.drifts[k] = cloneDrift(v)
	return nil
}
func (r *memoryContractRepo) ListDriftResults(c context.Context, t, revisionID string) ([]*entity.DataDriftResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listDrifts(c, t, revisionID)
}
func (r *memoryContractRepo) listDrifts(_ context.Context, t, revisionID string) ([]*entity.DataDriftResult, error) {
	out := []*entity.DataDriftResult{}
	for _, v := range r.drifts {
		if v.TenantID == t && v.RevisionID == revisionID {
			out = append(out, cloneDrift(v))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (r *memoryContractRepo) CreateGapResult(c context.Context, v *entity.DataContractGapResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.createGap(c, v)
}
func (r *memoryContractRepo) createGap(_ context.Context, v *entity.DataContractGapResult) error {
	if v == nil {
		return entity.ErrContractInvalidPayload
	}
	k := contractKey(v.TenantID, v.GapResultID)
	if _, ok := r.gaps[k]; ok {
		return entity.ErrContractInvalidState
	}
	r.gaps[k] = cloneGap(v)
	return nil
}
func (r *memoryContractRepo) ListGapResults(c context.Context, t, revisionID string) ([]*entity.DataContractGapResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listGaps(c, t, revisionID)
}
func (r *memoryContractRepo) listGaps(_ context.Context, t, revisionID string) ([]*entity.DataContractGapResult, error) {
	out := []*entity.DataContractGapResult{}
	for _, v := range r.gaps {
		if v.TenantID == t && v.RevisionID == revisionID {
			out = append(out, cloneGap(v))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (tx *memoryContractTx) Transaction(_ context.Context, fn func(ContractRepository) error) error {
	return fn(tx)
}
func (tx *memoryContractTx) CreateContract(c context.Context, v *entity.DataContract) error {
	return tx.repo.createContract(c, v)
}
func (tx *memoryContractTx) GetContract(c context.Context, t, id string) (*entity.DataContract, error) {
	return tx.repo.getContract(c, t, id)
}
func (tx *memoryContractTx) ListContracts(c context.Context, t, b string) ([]*entity.DataContract, error) {
	return tx.repo.listContracts(c, t, b)
}
func (tx *memoryContractTx) UpdateContractActiveRevision(c context.Context, t, contractID, revisionID string) error {
	return tx.repo.updateActiveRevision(c, t, contractID, revisionID)
}
func (tx *memoryContractTx) CreateRevision(c context.Context, v *entity.DataContractRevision) error {
	return tx.repo.createRevision(c, v)
}
func (tx *memoryContractTx) GetRevision(c context.Context, t, id string) (*entity.DataContractRevision, error) {
	return tx.repo.getRevision(c, t, id)
}
func (tx *memoryContractTx) ListRevisions(c context.Context, t, contractID string) ([]*entity.DataContractRevision, error) {
	return tx.repo.listRevisions(c, t, contractID)
}
func (tx *memoryContractTx) GetRevisionByVersion(c context.Context, t, contractID string, version int32) (*entity.DataContractRevision, error) {
	return tx.repo.getRevisionByVersion(c, t, contractID, version)
}
func (tx *memoryContractTx) AllocateNextVersion(c context.Context, t, contractID string) (int32, error) {
	return tx.repo.allocateNextVersion(c, t, contractID)
}
func (tx *memoryContractTx) UpdateRevisionStatus(c context.Context, t, revisionID string, from, to entity.ContractStatus) error {
	return tx.repo.updateRevisionStatus(c, t, revisionID, from, to)
}
func (tx *memoryContractTx) CreateValidationResult(c context.Context, v *entity.DataValidationResult) error {
	return tx.repo.createValidation(c, v)
}
func (tx *memoryContractTx) ListValidationResults(c context.Context, t, revisionID string) ([]*entity.DataValidationResult, error) {
	return tx.repo.listValidations(c, t, revisionID)
}
func (tx *memoryContractTx) GetValidationResult(c context.Context, t, id string) (*entity.DataValidationResult, error) {
	return tx.repo.getValidation(c, t, id)
}
func (tx *memoryContractTx) CreateLifecycleEvent(c context.Context, v *entity.DataContractLifecycleEvent) error {
	return tx.repo.createLifecycle(c, v)
}
func (tx *memoryContractTx) ListLifecycleEvents(c context.Context, t, contractID string) ([]*entity.DataContractLifecycleEvent, error) {
	return tx.repo.listLifecycle(c, t, contractID)
}
func (tx *memoryContractTx) CreateDriftResult(c context.Context, v *entity.DataDriftResult) error {
	return tx.repo.createDrift(c, v)
}
func (tx *memoryContractTx) ListDriftResults(c context.Context, t, revisionID string) ([]*entity.DataDriftResult, error) {
	return tx.repo.listDrifts(c, t, revisionID)
}
func (tx *memoryContractTx) CreateGapResult(c context.Context, v *entity.DataContractGapResult) error {
	return tx.repo.createGap(c, v)
}
func (tx *memoryContractTx) ListGapResults(c context.Context, t, revisionID string) ([]*entity.DataContractGapResult, error) {
	return tx.repo.listGaps(c, t, revisionID)
}
