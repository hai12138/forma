/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package repository

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
)

type memoryMappingRepo struct {
	mu           sync.Mutex
	mappings     map[string]*entity.SemanticMapping
	runs         map[string]*entity.SemanticMappingAnalysisRun
	runKeys      map[string]string
	decisions    map[string]*entity.SemanticMappingDecision
	decisionKeys map[string]string
}
type memoryMappingTx struct{ repo *memoryMappingRepo }

func NewMemoryMappingRepository() MappingRepository {
	return &memoryMappingRepo{mappings: map[string]*entity.SemanticMapping{}, runs: map[string]*entity.SemanticMappingAnalysisRun{}, runKeys: map[string]string{}, decisions: map[string]*entity.SemanticMappingDecision{}, decisionKeys: map[string]string{}}
}

func mappingKey(t, id string) string    { return t + "\x00" + id }
func mappingRunKey(t, id string) string { return t + "\x00" + id }
func mappingIdemKey(v *entity.SemanticMappingAnalysisRun) string {
	return fmt.Sprintf("%s\x00%s\x00%d\x00%s", v.TenantID, v.BusinessID, v.BusinessModelRevision, v.ClientRequestID)
}
func mappingDecisionKey(t, id string) string       { return t + "\x00" + id }
func mappingDecisionSourceKey(t, id string) string { return t + "\x00" + id }
func cloneMapping(v *entity.SemanticMapping) *entity.SemanticMapping {
	if v == nil {
		return nil
	}
	x := *v
	x.TargetFieldPaths = append([]string(nil), v.TargetFieldPaths...)
	x.TransformSpec = append([]byte(nil), v.TransformSpec...)
	return &x
}
func cloneMappingRun(v *entity.SemanticMappingAnalysisRun) *entity.SemanticMappingAnalysisRun {
	if v == nil {
		return nil
	}
	x := *v
	cloneTime := func(p *time.Time) *time.Time {
		if p == nil {
			return nil
		}
		y := *p
		return &y
	}
	x.StartedAt = cloneTime(v.StartedAt)
	x.CompletedAt = cloneTime(v.CompletedAt)
	x.LastRetryAt = cloneTime(v.LastRetryAt)
	x.ExecutionClaimedAt = cloneTime(v.ExecutionClaimedAt)
	x.LeaseExpiresAt = cloneTime(v.LeaseExpiresAt)
	return &x
}
func cloneMappingDecision(v *entity.SemanticMappingDecision) *entity.SemanticMappingDecision {
	if v == nil {
		return nil
	}
	x := *v
	return &x
}

func (r *memoryMappingRepo) CreateMapping(c context.Context, v *entity.SemanticMapping) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.createMapping(c, v)
}
func (r *memoryMappingRepo) createMapping(_ context.Context, v *entity.SemanticMapping) error {
	if v == nil {
		return entity.ErrMappingTransformInvalid
	}
	k := mappingKey(v.TenantID, v.MappingID)
	if _, ok := r.mappings[k]; ok {
		return entity.ErrMappingInvalidState
	}
	r.mappings[k] = cloneMapping(v)
	return nil
}
func (r *memoryMappingRepo) CreateMappingsBatch(c context.Context, vs []*entity.SemanticMapping) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.createMappingsBatch(c, vs)
}
func (r *memoryMappingRepo) createMappingsBatch(c context.Context, vs []*entity.SemanticMapping) error {
	seen := map[string]bool{}
	for _, v := range vs {
		if v == nil || seen[mappingKey(v.TenantID, v.MappingID)] || r.mappings[mappingKey(v.TenantID, v.MappingID)] != nil {
			return entity.ErrMappingInvalidState
		}
		seen[mappingKey(v.TenantID, v.MappingID)] = true
	}
	for _, v := range vs {
		if err := r.createMapping(c, v); err != nil {
			return err
		}
	}
	return nil
}
func (r *memoryMappingRepo) GetMapping(c context.Context, t, id string) (*entity.SemanticMapping, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getMapping(c, t, id)
}
func (r *memoryMappingRepo) getMapping(_ context.Context, t, id string) (*entity.SemanticMapping, error) {
	v := r.mappings[mappingKey(t, id)]
	if v == nil {
		return nil, entity.ErrMappingNotFound
	}
	return cloneMapping(v), nil
}
func (r *memoryMappingRepo) ListMappings(c context.Context, t, b string, rev int32, s entity.MappingStatus) ([]*entity.SemanticMapping, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listMappings(c, t, b, rev, s)
}
func (r *memoryMappingRepo) listMappings(_ context.Context, t, b string, rev int32, s entity.MappingStatus) ([]*entity.SemanticMapping, error) {
	out := []*entity.SemanticMapping{}
	for _, v := range r.mappings {
		if v.TenantID == t && v.BusinessID == b && v.BusinessModelRevision == rev && (s == "" || v.Status == s) {
			out = append(out, cloneMapping(v))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}
func (r *memoryMappingRepo) UpdateMappingStatusCAS(c context.Context, t, id string, f, to entity.MappingStatus) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.updateStatus(c, t, id, f, to)
}
func (r *memoryMappingRepo) updateStatus(_ context.Context, t, id string, f, to entity.MappingStatus) (bool, error) {
	v := r.mappings[mappingKey(t, id)]
	if v == nil || v.Status != f {
		return false, nil
	}
	v.Status = to
	v.UpdatedAt = time.Now().UTC()
	return true, nil
}
func (r *memoryMappingRepo) GetConfirmedMappingByRequirement(c context.Context, t, b string, rev int32, req string) (*entity.SemanticMapping, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getConfirmed(c, t, b, rev, req)
}
func (r *memoryMappingRepo) getConfirmed(_ context.Context, t, b string, rev int32, req string) (*entity.SemanticMapping, error) {
	for _, v := range r.mappings {
		if v.TenantID == t && v.BusinessID == b && v.BusinessModelRevision == rev && v.RequirementID == req && v.Status == entity.MappingStatusConfirmed {
			return cloneMapping(v), nil
		}
	}
	return nil, entity.ErrMappingNotFound
}

func (r *memoryMappingRepo) CreateOrClaimMappingAnalysisRun(c context.Context, v *entity.SemanticMappingAnalysisRun) (*entity.SemanticMappingAnalysisRun, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.createRun(c, v)
}
func (r *memoryMappingRepo) createRun(_ context.Context, v *entity.SemanticMappingAnalysisRun) (*entity.SemanticMappingAnalysisRun, bool, error) {
	k := mappingIdemKey(v)
	if id, ok := r.runKeys[k]; ok {
		old := r.runs[mappingRunKey(v.TenantID, id)]
		if old.RequestDigest != v.RequestDigest {
			return nil, false, entity.ErrMappingAnalysisIdempotencyConflict
		}
		return cloneMappingRun(old), false, nil
	}
	x := cloneMappingRun(v)
	if x.ExecutionGeneration == 0 {
		x.ExecutionGeneration = 1
	}
	r.runs[mappingRunKey(v.TenantID, v.AnalysisRunID)] = x
	r.runKeys[k] = v.AnalysisRunID
	return cloneMappingRun(x), true, nil
}
func (r *memoryMappingRepo) GetMappingAnalysisRun(c context.Context, t, id string) (*entity.SemanticMappingAnalysisRun, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getRun(c, t, id)
}
func (r *memoryMappingRepo) getRun(_ context.Context, t, id string) (*entity.SemanticMappingAnalysisRun, error) {
	v := r.runs[mappingRunKey(t, id)]
	if v == nil {
		return nil, entity.ErrMappingAnalysisNotFound
	}
	return cloneMappingRun(v), nil
}
func (r *memoryMappingRepo) MarkMappingAnalysisSucceeded(c context.Context, t, id, m string, g int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.markRun(c, t, id, entity.AnalysisSucceeded, m, "", "", g)
}
func (r *memoryMappingRepo) MarkMappingAnalysisFailed(c context.Context, t, id, k, msg string, g int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.markRun(c, t, id, entity.AnalysisFailed, "", k, msg, g)
}
func (r *memoryMappingRepo) markRun(_ context.Context, t, id string, status entity.AnalysisRunStatus, model, key, msg string, g int32) error {
	v := r.runs[mappingRunKey(t, id)]
	if v == nil {
		return entity.ErrMappingAnalysisNotFound
	}
	if v.Status != entity.AnalysisPending || v.ExecutionGeneration != g {
		return entity.ErrMappingInvalidState
	}
	now := time.Now().UTC()
	v.Status = status
	v.ModelRef = model
	v.ErrorKey = key
	v.ErrorMessageSanitized = msg
	v.CompletedAt = &now
	v.UpdatedAt = now
	return nil
}
func (r *memoryMappingRepo) ClaimMappingAnalysisRetry(c context.Context, t, id, a string) (bool, int32, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.claimRetry(c, t, id, a)
}
func (r *memoryMappingRepo) claimRetry(_ context.Context, t, id, a string) (bool, int32, error) {
	v := r.runs[mappingRunKey(t, id)]
	if v == nil || v.Status != entity.AnalysisFailed {
		return false, 0, nil
	}
	now := time.Now().UTC()
	exp := now.Add(5 * time.Minute)
	v.Status = entity.AnalysisPending
	v.RetryCount++
	v.LastRetryBy = a
	v.LastRetryAt = &now
	v.ExecutionGeneration++
	v.ExecutionClaimedAt = &now
	v.LeaseExpiresAt = &exp
	v.CompletedAt = nil
	v.UpdatedAt = now
	return true, v.ExecutionGeneration, nil
}
func (r *memoryMappingRepo) ClaimExpiredMappingAnalysis(c context.Context, t, id string, g int32, now time.Time) (*entity.SemanticMappingAnalysisRun, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v := r.runs[mappingRunKey(t, id)]
	if v == nil {
		return nil, false, entity.ErrMappingAnalysisNotFound
	}
	if v.Status != entity.AnalysisPending || v.ExecutionGeneration != g || v.LeaseExpiresAt == nil || now.Before(*v.LeaseExpiresAt) {
		return cloneMappingRun(v), false, nil
	}
	v.ExecutionGeneration++
	exp := now.Add(5 * time.Minute)
	v.ExecutionClaimedAt = &now
	v.LeaseExpiresAt = &exp
	return cloneMappingRun(v), true, nil
}
func (r *memoryMappingRepo) CreateMappingDecision(c context.Context, v *entity.SemanticMappingDecision) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.createDecision(c, v)
}
func (r *memoryMappingRepo) createDecision(_ context.Context, v *entity.SemanticMappingDecision) error {
	k := mappingDecisionSourceKey(v.TenantID, v.SourceMappingID)
	if r.decisionKeys[k] != "" {
		return entity.ErrMappingAlreadyDecided
	}
	r.decisions[mappingDecisionKey(v.TenantID, v.DecisionID)] = cloneMappingDecision(v)
	r.decisionKeys[k] = v.DecisionID
	return nil
}
func (r *memoryMappingRepo) ListMappingDecisions(c context.Context, t, id string) ([]*entity.SemanticMappingDecision, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listDecisions(c, t, id)
}
func (r *memoryMappingRepo) listDecisions(_ context.Context, t, id string) ([]*entity.SemanticMappingDecision, error) {
	out := []*entity.SemanticMappingDecision{}
	for _, v := range r.decisions {
		if v.TenantID == t && (v.SourceMappingID == id || v.TargetMappingID == id) {
			out = append(out, cloneMappingDecision(v))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

type mappingSnapshot struct {
	mappings     map[string]*entity.SemanticMapping
	runs         map[string]*entity.SemanticMappingAnalysisRun
	runKeys      map[string]string
	decisions    map[string]*entity.SemanticMappingDecision
	decisionKeys map[string]string
}

func (r *memoryMappingRepo) snapshot() mappingSnapshot {
	s := mappingSnapshot{map[string]*entity.SemanticMapping{}, map[string]*entity.SemanticMappingAnalysisRun{}, map[string]string{}, map[string]*entity.SemanticMappingDecision{}, map[string]string{}}
	for k, v := range r.mappings {
		s.mappings[k] = cloneMapping(v)
	}
	for k, v := range r.runs {
		s.runs[k] = cloneMappingRun(v)
	}
	for k, v := range r.runKeys {
		s.runKeys[k] = v
	}
	for k, v := range r.decisions {
		s.decisions[k] = cloneMappingDecision(v)
	}
	for k, v := range r.decisionKeys {
		s.decisionKeys[k] = v
	}
	return s
}
func (r *memoryMappingRepo) Transaction(c context.Context, fn func(MappingRepository) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.snapshot()
	if err := fn(&memoryMappingTx{r}); err != nil {
		r.mappings = s.mappings
		r.runs = s.runs
		r.runKeys = s.runKeys
		r.decisions = s.decisions
		r.decisionKeys = s.decisionKeys
		return err
	}
	return nil
}

func (tx *memoryMappingTx) CreateMapping(c context.Context, v *entity.SemanticMapping) error {
	return tx.repo.createMapping(c, v)
}
func (tx *memoryMappingTx) CreateMappingsBatch(c context.Context, v []*entity.SemanticMapping) error {
	return tx.repo.createMappingsBatch(c, v)
}
func (tx *memoryMappingTx) GetMapping(c context.Context, t, id string) (*entity.SemanticMapping, error) {
	return tx.repo.getMapping(c, t, id)
}
func (tx *memoryMappingTx) ListMappings(c context.Context, t, b string, r int32, s entity.MappingStatus) ([]*entity.SemanticMapping, error) {
	return tx.repo.listMappings(c, t, b, r, s)
}
func (tx *memoryMappingTx) UpdateMappingStatusCAS(c context.Context, t, id string, f, to entity.MappingStatus) (bool, error) {
	return tx.repo.updateStatus(c, t, id, f, to)
}
func (tx *memoryMappingTx) GetConfirmedMappingByRequirement(c context.Context, t, b string, r int32, q string) (*entity.SemanticMapping, error) {
	return tx.repo.getConfirmed(c, t, b, r, q)
}
func (tx *memoryMappingTx) CreateOrClaimMappingAnalysisRun(c context.Context, v *entity.SemanticMappingAnalysisRun) (*entity.SemanticMappingAnalysisRun, bool, error) {
	return tx.repo.createRun(c, v)
}
func (tx *memoryMappingTx) GetMappingAnalysisRun(c context.Context, t, id string) (*entity.SemanticMappingAnalysisRun, error) {
	return tx.repo.getRun(c, t, id)
}
func (tx *memoryMappingTx) MarkMappingAnalysisSucceeded(c context.Context, t, id, m string, g int32) error {
	return tx.repo.markRun(c, t, id, entity.AnalysisSucceeded, m, "", "", g)
}
func (tx *memoryMappingTx) MarkMappingAnalysisFailed(c context.Context, t, id, k, msg string, g int32) error {
	return tx.repo.markRun(c, t, id, entity.AnalysisFailed, "", k, msg, g)
}
func (tx *memoryMappingTx) ClaimMappingAnalysisRetry(c context.Context, t, id, a string) (bool, int32, error) {
	return tx.repo.claimRetry(c, t, id, a)
}
func (tx *memoryMappingTx) ClaimExpiredMappingAnalysis(c context.Context, t, id string, g int32, n time.Time) (*entity.SemanticMappingAnalysisRun, bool, error) {
	v := tx.repo.runs[mappingRunKey(t, id)]
	if v == nil {
		return nil, false, entity.ErrMappingAnalysisNotFound
	}
	return cloneMappingRun(v), false, nil
}
func (tx *memoryMappingTx) CreateMappingDecision(c context.Context, v *entity.SemanticMappingDecision) error {
	return tx.repo.createDecision(c, v)
}
func (tx *memoryMappingTx) ListMappingDecisions(c context.Context, t, id string) ([]*entity.SemanticMappingDecision, error) {
	return tx.repo.listDecisions(c, t, id)
}
func (tx *memoryMappingTx) Transaction(c context.Context, fn func(MappingRepository) error) error {
	return fn(tx)
}
