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

// NewMemoryDataRepository returns a concurrency-safe repository intended for
// domain tests and local, non-persistent use.
func NewMemoryDataRepository() DataRepository {
	return &memRepo{
		requirements: make(map[string]*entity.DataRequirement),
		runs:         make(map[string]*entity.DataRequirementAnalysisRun),
		runKeys:      make(map[string]string),
		decisions:    make(map[string]*entity.DataRequirementDecision),
		decisionKeys: make(map[string]string),
	}
}

type memRepo struct {
	mu           sync.Mutex
	requirements map[string]*entity.DataRequirement
	runs         map[string]*entity.DataRequirementAnalysisRun
	runKeys      map[string]string
	decisions    map[string]*entity.DataRequirementDecision
	decisionKeys map[string]string
}

type memRepoTx struct{ repo *memRepo }

func joined(parts ...any) string { return fmt.Sprint(parts...) }

func requirementKey(tenantID, requirementID string) string {
	return joined(tenantID, "\x00", requirementID)
}

func runKey(tenantID, analysisRunID string) string {
	return joined(tenantID, "\x00", analysisRunID)
}

func idempotencyKey(tenantID, businessID string, revision int32, clientRequestID string) string {
	return joined(tenantID, "\x00", businessID, "\x00", revision, "\x00", clientRequestID)
}

func decisionKey(tenantID, decisionID string) string {
	return joined(tenantID, "\x00", decisionID)
}

func decisionSourceKey(tenantID, sourceRequirementID string) string {
	return joined(tenantID, "\x00", sourceRequirementID)
}

func cloneRequirement(in *entity.DataRequirement) *entity.DataRequirement {
	if in == nil {
		return nil
	}
	out := *in
	out.BusinessElementRefs = append([]string(nil), in.BusinessElementRefs...)
	return &out
}

func cloneRun(in *entity.DataRequirementAnalysisRun) *entity.DataRequirementAnalysisRun {
	if in == nil {
		return nil
	}
	out := *in
	if in.StartedAt != nil {
		v := *in.StartedAt
		out.StartedAt = &v
	}
	if in.CompletedAt != nil {
		v := *in.CompletedAt
		out.CompletedAt = &v
	}
	return &out
}

func cloneDecision(in *entity.DataRequirementDecision) *entity.DataRequirementDecision {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func (r *memRepo) CreateRequirement(ctx context.Context, req *entity.DataRequirement) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.createRequirement(ctx, req)
}

func (r *memRepo) createRequirement(_ context.Context, req *entity.DataRequirement) error {
	if req == nil {
		return entity.ErrInvalidProposal
	}
	key := requirementKey(req.TenantID, req.RequirementID)
	if _, exists := r.requirements[key]; exists {
		return fmt.Errorf("duplicate data requirement %q", req.RequirementID)
	}
	r.requirements[key] = cloneRequirement(req)
	return nil
}

func (r *memRepo) GetRequirement(ctx context.Context, tenantID, requirementID string) (*entity.DataRequirement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getRequirement(ctx, tenantID, requirementID)
}

func (r *memRepo) getRequirement(_ context.Context, tenantID, requirementID string) (*entity.DataRequirement, error) {
	req, ok := r.requirements[requirementKey(tenantID, requirementID)]
	if !ok {
		return nil, entity.ErrRequirementNotFound
	}
	return cloneRequirement(req), nil
}

func (r *memRepo) ListRequirementsByRevision(ctx context.Context, tenantID, businessID string, revision int32) ([]*entity.DataRequirement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listRequirementsByRevision(ctx, tenantID, businessID, revision)
}

func (r *memRepo) listRequirementsByRevision(_ context.Context, tenantID, businessID string, revision int32) ([]*entity.DataRequirement, error) {
	out := make([]*entity.DataRequirement, 0)
	for _, req := range r.requirements {
		if req.TenantID == tenantID && req.BusinessID == businessID && req.BusinessModelRevision == revision {
			out = append(out, cloneRequirement(req))
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (r *memRepo) UpdateRequirementStatusCAS(ctx context.Context, tenantID, requirementID string, from, to entity.RequirementStatus) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.updateRequirementStatusCAS(ctx, tenantID, requirementID, from, to)
}

func (r *memRepo) updateRequirementStatusCAS(_ context.Context, tenantID, requirementID string, from, to entity.RequirementStatus) (bool, error) {
	req, ok := r.requirements[requirementKey(tenantID, requirementID)]
	if !ok || req.Status != from {
		return false, nil
	}
	req.Status = to
	req.UpdatedAt = time.Now().UTC()
	return true, nil
}

func (r *memRepo) CreateOrClaimAnalysisRun(ctx context.Context, run *entity.DataRequirementAnalysisRun) (*entity.DataRequirementAnalysisRun, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.createOrClaimAnalysisRun(ctx, run)
}

func (r *memRepo) createOrClaimAnalysisRun(_ context.Context, run *entity.DataRequirementAnalysisRun) (*entity.DataRequirementAnalysisRun, bool, error) {
	if run == nil {
		return nil, false, entity.ErrInvalidProposal
	}
	key := idempotencyKey(run.TenantID, run.BusinessID, run.BusinessModelRevision, run.ClientRequestID)
	if id, ok := r.runKeys[key]; ok {
		existing := r.runs[runKey(run.TenantID, id)]
		if existing.RequestDigest != run.RequestDigest {
			return nil, false, entity.ErrAnalysisIdempotencyConflict
		}
		return cloneRun(existing), false, nil
	}
	idKey := runKey(run.TenantID, run.AnalysisRunID)
	if _, exists := r.runs[idKey]; exists {
		return nil, false, fmt.Errorf("duplicate data analysis run %q", run.AnalysisRunID)
	}
	r.runs[idKey] = cloneRun(run)
	r.runKeys[key] = run.AnalysisRunID
	return cloneRun(run), true, nil
}

func (r *memRepo) GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*entity.DataRequirementAnalysisRun, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getAnalysisRun(ctx, tenantID, analysisRunID)
}

func (r *memRepo) getAnalysisRun(_ context.Context, tenantID, analysisRunID string) (*entity.DataRequirementAnalysisRun, error) {
	run, ok := r.runs[runKey(tenantID, analysisRunID)]
	if !ok {
		return nil, entity.ErrAnalysisNotFound
	}
	return cloneRun(run), nil
}

func (r *memRepo) GetAnalysisRunByIdempotencyKey(ctx context.Context, tenantID, businessID string, revision int32, clientRequestID string) (*entity.DataRequirementAnalysisRun, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getAnalysisRunByIdempotencyKey(ctx, tenantID, businessID, revision, clientRequestID)
}

func (r *memRepo) getAnalysisRunByIdempotencyKey(ctx context.Context, tenantID, businessID string, revision int32, clientRequestID string) (*entity.DataRequirementAnalysisRun, error) {
	id, ok := r.runKeys[idempotencyKey(tenantID, businessID, revision, clientRequestID)]
	if !ok {
		return nil, entity.ErrAnalysisNotFound
	}
	return r.getAnalysisRun(ctx, tenantID, id)
}

func (r *memRepo) MarkAnalysisSucceeded(ctx context.Context, tenantID, analysisRunID, modelRef string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.markAnalysisSucceeded(ctx, tenantID, analysisRunID, modelRef)
}

func (r *memRepo) markAnalysisSucceeded(_ context.Context, tenantID, analysisRunID, modelRef string) error {
	run, ok := r.runs[runKey(tenantID, analysisRunID)]
	if !ok {
		return entity.ErrAnalysisNotFound
	}
	if run.Status != entity.AnalysisPending {
		return entity.ErrRequirementInvalidState
	}
	now := time.Now().UTC()
	run.Status, run.ModelRef, run.CompletedAt, run.UpdatedAt = entity.AnalysisSucceeded, modelRef, &now, now
	run.ErrorKey, run.ErrorMessageSanitized = "", ""
	return nil
}

func (r *memRepo) MarkAnalysisFailed(ctx context.Context, tenantID, analysisRunID, errorKey, sanitizedMsg string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.markAnalysisFailed(ctx, tenantID, analysisRunID, errorKey, sanitizedMsg)
}

func (r *memRepo) markAnalysisFailed(_ context.Context, tenantID, analysisRunID, errorKey, sanitizedMsg string) error {
	run, ok := r.runs[runKey(tenantID, analysisRunID)]
	if !ok {
		return entity.ErrAnalysisNotFound
	}
	if run.Status != entity.AnalysisPending {
		return entity.ErrRequirementInvalidState
	}
	if len(sanitizedMsg) > 1024 {
		sanitizedMsg = sanitizedMsg[:1024]
	}
	now := time.Now().UTC()
	run.Status, run.ErrorKey, run.ErrorMessageSanitized = entity.AnalysisFailed, errorKey, sanitizedMsg
	run.CompletedAt, run.UpdatedAt = &now, now
	return nil
}

func (r *memRepo) ClaimAnalysisRetry(ctx context.Context, tenantID, analysisRunID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.claimAnalysisRetry(ctx, tenantID, analysisRunID)
}

func (r *memRepo) claimAnalysisRetry(_ context.Context, tenantID, analysisRunID string) (bool, error) {
	run, ok := r.runs[runKey(tenantID, analysisRunID)]
	if !ok || run.Status != entity.AnalysisFailed {
		return false, nil
	}
	now := time.Now().UTC()
	run.Status, run.RetryCount, run.StartedAt, run.CompletedAt = entity.AnalysisPending, run.RetryCount+1, &now, nil
	run.ErrorKey, run.ErrorMessageSanitized, run.UpdatedAt = "", "", now
	return true, nil
}

func (r *memRepo) CreateDecision(ctx context.Context, d *entity.DataRequirementDecision) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.createDecision(ctx, d)
}

func (r *memRepo) createDecision(_ context.Context, d *entity.DataRequirementDecision) error {
	if d == nil {
		return entity.ErrInvalidProposal
	}
	sourceKey := decisionSourceKey(d.TenantID, d.SourceRequirementID)
	if _, exists := r.decisionKeys[sourceKey]; exists {
		return entity.ErrRequirementAlreadyDecided
	}
	key := decisionKey(d.TenantID, d.DecisionID)
	if _, exists := r.decisions[key]; exists {
		return entity.ErrRequirementAlreadyDecided
	}
	r.decisions[key] = cloneDecision(d)
	r.decisionKeys[sourceKey] = d.DecisionID
	return nil
}

func (r *memRepo) GetDecisionBySource(ctx context.Context, tenantID, sourceRequirementID string) (*entity.DataRequirementDecision, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getDecisionBySource(ctx, tenantID, sourceRequirementID)
}

func (r *memRepo) getDecisionBySource(_ context.Context, tenantID, sourceRequirementID string) (*entity.DataRequirementDecision, error) {
	id, ok := r.decisionKeys[decisionSourceKey(tenantID, sourceRequirementID)]
	if !ok {
		return nil, entity.ErrRequirementNotFound
	}
	return cloneDecision(r.decisions[decisionKey(tenantID, id)]), nil
}

func (r *memRepo) ListDecisionsByRequirement(ctx context.Context, tenantID, requirementID string) ([]*entity.DataRequirementDecision, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.listDecisionsByRequirement(ctx, tenantID, requirementID)
}

func (r *memRepo) listDecisionsByRequirement(_ context.Context, tenantID, requirementID string) ([]*entity.DataRequirementDecision, error) {
	out := make([]*entity.DataRequirementDecision, 0)
	for _, d := range r.decisions {
		if d.TenantID == tenantID && (d.SourceRequirementID == requirementID || d.TargetRequirementID == requirementID) {
			out = append(out, cloneDecision(d))
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (r *memRepo) CreateRequirementsBatch(ctx context.Context, reqs []*entity.DataRequirement) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.createRequirementsBatch(ctx, reqs)
}

func (r *memRepo) createRequirementsBatch(ctx context.Context, reqs []*entity.DataRequirement) error {
	seen := make(map[string]struct{}, len(reqs))
	for _, req := range reqs {
		if req == nil {
			return entity.ErrInvalidProposal
		}
		key := requirementKey(req.TenantID, req.RequirementID)
		if _, exists := r.requirements[key]; exists {
			return fmt.Errorf("duplicate data requirement %q", req.RequirementID)
		}
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate data requirement %q", req.RequirementID)
		}
		seen[key] = struct{}{}
	}
	for _, req := range reqs {
		if err := r.createRequirement(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

type memSnapshot struct {
	requirements map[string]*entity.DataRequirement
	runs         map[string]*entity.DataRequirementAnalysisRun
	runKeys      map[string]string
	decisions    map[string]*entity.DataRequirementDecision
	decisionKeys map[string]string
}

func (r *memRepo) snapshot() memSnapshot {
	s := memSnapshot{
		requirements: make(map[string]*entity.DataRequirement, len(r.requirements)),
		runs:         make(map[string]*entity.DataRequirementAnalysisRun, len(r.runs)),
		runKeys:      make(map[string]string, len(r.runKeys)),
		decisions:    make(map[string]*entity.DataRequirementDecision, len(r.decisions)),
		decisionKeys: make(map[string]string, len(r.decisionKeys)),
	}
	for k, v := range r.requirements {
		s.requirements[k] = cloneRequirement(v)
	}
	for k, v := range r.runs {
		s.runs[k] = cloneRun(v)
	}
	for k, v := range r.runKeys {
		s.runKeys[k] = v
	}
	for k, v := range r.decisions {
		s.decisions[k] = cloneDecision(v)
	}
	for k, v := range r.decisionKeys {
		s.decisionKeys[k] = v
	}
	return s
}

func (r *memRepo) restore(s memSnapshot) {
	r.requirements, r.runs, r.runKeys = s.requirements, s.runs, s.runKeys
	r.decisions, r.decisionKeys = s.decisions, s.decisionKeys
}

func (r *memRepo) Transaction(ctx context.Context, fn func(txRepo DataRepository) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.transaction(ctx, fn)
}

func (r *memRepo) transaction(_ context.Context, fn func(txRepo DataRepository) error) error {
	snapshot := r.snapshot()
	if err := fn(&memRepoTx{repo: r}); err != nil {
		r.restore(snapshot)
		return err
	}
	return nil
}

func (tx *memRepoTx) CreateRequirement(ctx context.Context, req *entity.DataRequirement) error {
	return tx.repo.createRequirement(ctx, req)
}
func (tx *memRepoTx) GetRequirement(ctx context.Context, tenantID, requirementID string) (*entity.DataRequirement, error) {
	return tx.repo.getRequirement(ctx, tenantID, requirementID)
}
func (tx *memRepoTx) ListRequirementsByRevision(ctx context.Context, tenantID, businessID string, revision int32) ([]*entity.DataRequirement, error) {
	return tx.repo.listRequirementsByRevision(ctx, tenantID, businessID, revision)
}
func (tx *memRepoTx) UpdateRequirementStatusCAS(ctx context.Context, tenantID, requirementID string, from, to entity.RequirementStatus) (bool, error) {
	return tx.repo.updateRequirementStatusCAS(ctx, tenantID, requirementID, from, to)
}
func (tx *memRepoTx) CreateOrClaimAnalysisRun(ctx context.Context, run *entity.DataRequirementAnalysisRun) (*entity.DataRequirementAnalysisRun, bool, error) {
	return tx.repo.createOrClaimAnalysisRun(ctx, run)
}
func (tx *memRepoTx) GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*entity.DataRequirementAnalysisRun, error) {
	return tx.repo.getAnalysisRun(ctx, tenantID, analysisRunID)
}
func (tx *memRepoTx) GetAnalysisRunByIdempotencyKey(ctx context.Context, tenantID, businessID string, revision int32, clientRequestID string) (*entity.DataRequirementAnalysisRun, error) {
	return tx.repo.getAnalysisRunByIdempotencyKey(ctx, tenantID, businessID, revision, clientRequestID)
}
func (tx *memRepoTx) MarkAnalysisSucceeded(ctx context.Context, tenantID, analysisRunID, modelRef string) error {
	return tx.repo.markAnalysisSucceeded(ctx, tenantID, analysisRunID, modelRef)
}
func (tx *memRepoTx) MarkAnalysisFailed(ctx context.Context, tenantID, analysisRunID, errorKey, sanitizedMsg string) error {
	return tx.repo.markAnalysisFailed(ctx, tenantID, analysisRunID, errorKey, sanitizedMsg)
}
func (tx *memRepoTx) ClaimAnalysisRetry(ctx context.Context, tenantID, analysisRunID string) (bool, error) {
	return tx.repo.claimAnalysisRetry(ctx, tenantID, analysisRunID)
}
func (tx *memRepoTx) CreateDecision(ctx context.Context, d *entity.DataRequirementDecision) error {
	return tx.repo.createDecision(ctx, d)
}
func (tx *memRepoTx) GetDecisionBySource(ctx context.Context, tenantID, sourceRequirementID string) (*entity.DataRequirementDecision, error) {
	return tx.repo.getDecisionBySource(ctx, tenantID, sourceRequirementID)
}
func (tx *memRepoTx) ListDecisionsByRequirement(ctx context.Context, tenantID, requirementID string) ([]*entity.DataRequirementDecision, error) {
	return tx.repo.listDecisionsByRequirement(ctx, tenantID, requirementID)
}
func (tx *memRepoTx) CreateRequirementsBatch(ctx context.Context, reqs []*entity.DataRequirement) error {
	return tx.repo.createRequirementsBatch(ctx, reqs)
}
func (tx *memRepoTx) Transaction(ctx context.Context, fn func(txRepo DataRepository) error) error {
	return tx.repo.transaction(ctx, fn)
}
