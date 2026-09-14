/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/fixture"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/repository"
)

func assertPersistedFailedDTO(t *testing.T, res *AnalysisResult, wantRunID string, wantAttemptMin int32) {
	t.Helper()
	require.NotNil(t, res, "persisted FAILED must return AnalysisResult DTO")
	require.NotNil(t, res.Run)
	require.Equal(t, entity.AnalysisFailed, res.Run.Status)
	require.True(t, res.OwnedExecute)
	require.NotEmpty(t, res.Run.AnalysisRunID)
	if wantRunID != "" {
		require.Equal(t, wantRunID, res.Run.AnalysisRunID)
	}
	require.GreaterOrEqual(t, res.Run.Attempt, wantAttemptMin)
	require.NotEmpty(t, res.Run.ErrorCode)
}

func TestF3RetryInvalidProposalReturnsFailedDTO(t *testing.T) {
	gen := &DeterministicFakeGenerator{Err: errors.New("boom")}
	svc, repo, _, _, _ := newTestService(gen)
	first, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "f3-retry-invprop", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.ErrorIs(t, err, entity.ErrAnalysisFailed)
	assertPersistedFailedDTO(t, first, "", 1)
	runID := first.Run.AnalysisRunID

	bad := fixture.LaboratoryFlowCapability()
	bad.Name = ""
	gen.Err = nil
	gen.Proposals = []entity.SemanticPayload{bad}

	retried, err := svc.RetryFailedAnalysis(context.Background(), "t1", runID, testActor)
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
	assertPersistedFailedDTO(t, retried, runID, 2)
	require.Equal(t, "FORMA_CAPABILITY_INVALID_REQUEST", retried.Run.ErrorCode)

	props, listErr := repo.ListProposalsByAnalysisRun(context.Background(), "t1", runID)
	require.NoError(t, listErr)
	require.Empty(t, props, "invalid proposal must not persist proposals")

	// Same client_request_id must not create a second AnalysisRun.
	again, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "f3-retry-invprop", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	require.Equal(t, runID, again.Run.AnalysisRunID)
	require.False(t, again.OwnedExecute)
}

func TestF3RetryCorruptRequestJSONReturnsFailedDTO(t *testing.T) {
	gen := &DeterministicFakeGenerator{Err: errors.New("boom")}
	svc, repo, _, _, _ := newTestService(gen)
	first, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "f3-retry-corrupt", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.ErrorIs(t, err, entity.ErrAnalysisFailed)
	assertPersistedFailedDTO(t, first, "", 1)
	runID := first.Run.AnalysisRunID

	run, err := repo.GetAnalysisRun(context.Background(), "t1", runID)
	require.NoError(t, err)
	run.RequestJSON = "{not-json"
	require.NoError(t, repository.ReplaceAnalysisRunForTest(repo, run))

	gen.Err = nil
	gen.Proposals = []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}
	callsBefore := gen.CallCount()

	retried, err := svc.RetryFailedAnalysis(context.Background(), "t1", runID, testActor)
	require.ErrorIs(t, err, entity.ErrConsistency)
	assertPersistedFailedDTO(t, retried, runID, 2)
	require.Equal(t, "FORMA_CAPABILITY_INVALID_REQUEST", retried.Run.ErrorCode)
	require.Equal(t, callsBefore, gen.CallCount(), "corrupt RequestJSON must never reach generator")

	props, listErr := repo.ListProposalsByAnalysisRun(context.Background(), "t1", runID)
	require.NoError(t, listErr)
	require.Empty(t, props)
}

func TestF3RetryDigestMismatchReturnsFailedDTO(t *testing.T) {
	gen := &DeterministicFakeGenerator{Err: errors.New("boom")}
	svc, repo, _, _, _ := newTestService(gen)
	first, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "f3-retry-digest", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.ErrorIs(t, err, entity.ErrAnalysisFailed)
	assertPersistedFailedDTO(t, first, "", 1)
	runID := first.Run.AnalysisRunID

	run, err := repo.GetAnalysisRun(context.Background(), "t1", runID)
	require.NoError(t, err)
	// Body digest no longer matches stored RequestDigest.
	run.RequestJSON = `{"business_model_revision":1,"data_contract_pins":[{"data_contract_id":"dc_other","data_contract_version":1}],"requirement_refs":null}`
	require.NoError(t, repository.ReplaceAnalysisRunForTest(repo, run))

	gen.Err = nil
	gen.Proposals = []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}
	callsBefore := gen.CallCount()

	retried, err := svc.RetryFailedAnalysis(context.Background(), "t1", runID, testActor)
	require.ErrorIs(t, err, entity.ErrConsistency)
	assertPersistedFailedDTO(t, retried, runID, 2)
	require.Equal(t, callsBefore, gen.CallCount())
}

func TestF3LeaseTakeoverIllegalRequestReturnsFailedDTO(t *testing.T) {
	fixed := &fixedClock{t: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)}
	uow := NewMemoryUnitOfWork()
	gen := &recordingGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc := NewCapabilityService(&Components{UoW: uow, Generator: gen, Clock: fixed})

	caller := entity.AnalysisRequest{BusinessModelRevision: 1}
	digest, err := AnalysisRequestDigest(caller)
	require.NoError(t, err)
	exp := fixed.Now().Add(-time.Minute)
	claimedAt := fixed.Now().Add(-10 * time.Minute)
	seed := &entity.CapabilityAnalysisRun{
		AnalysisRunID: "run-f3-lease", TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1,
		ClientRequestID: "f3-lease-bad", RequestDigest: digest, Status: entity.AnalysisPending, Attempt: 1,
		RequestJSON: "{not-json", ExecutionClaimedAt: &claimedAt, LeaseExpiresAt: &exp,
		CreatedBy: testActor, CreatedAt: fixed.Now(), UpdatedAt: fixed.Now(),
	}
	_, created, err := uow.Root().CreateOrClaimAnalysisRun(context.Background(), seed)
	require.NoError(t, err)
	require.True(t, created)
	require.NoError(t, uow.Root().CreateAnalysisAttempt(context.Background(), &entity.CapabilityAnalysisAttempt{
		AttemptID: "att-f3-lease", AnalysisRunID: "run-f3-lease", TenantID: "t1", Attempt: 1,
		ActorPrincipalID: testActor, TriggerKind: entity.AttemptTriggerFirst,
		ResultStatus: entity.AttemptResultPending, CreatedAt: fixed.Now(),
	}))

	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "f3-lease-bad", ActorID: testActor,
		Analysis: caller,
	})
	require.ErrorIs(t, err, entity.ErrConsistency)
	assertPersistedFailedDTO(t, res, "run-f3-lease", 2)
	require.Equal(t, "FORMA_CAPABILITY_INVALID_REQUEST", res.Run.ErrorCode)
	require.Empty(t, gen.lastAnalysis.DataContractPins)

	props, listErr := uow.Root().ListProposalsByAnalysisRun(context.Background(), "t1", "run-f3-lease")
	require.NoError(t, listErr)
	require.Empty(t, props)
}

func TestF3PersistedFailedPathsShareSameRunDTO(t *testing.T) {
	// Generator fail + invalid proposal + corrupt retry — all keep the same analysis_run_id.
	gen := &DeterministicFakeGenerator{Err: errors.New("boom")}
	svc, repo, _, _, _ := newTestService(gen)
	first, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "f3-same-run", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.ErrorIs(t, err, entity.ErrAnalysisFailed)
	assertPersistedFailedDTO(t, first, "", 1)
	runID := first.Run.AnalysisRunID

	bad := fixture.LaboratoryFlowCapability()
	bad.Name = ""
	gen.Err = nil
	gen.Proposals = []entity.SemanticPayload{bad}
	second, err := svc.RetryFailedAnalysis(context.Background(), "t1", runID, testActor)
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
	assertPersistedFailedDTO(t, second, runID, 2)

	run, err := repo.GetAnalysisRun(context.Background(), "t1", runID)
	require.NoError(t, err)
	run.RequestJSON = "{broken"
	require.NoError(t, repository.ReplaceAnalysisRunForTest(repo, run))

	third, err := svc.RetryFailedAnalysis(context.Background(), "t1", runID, testActor)
	require.ErrorIs(t, err, entity.ErrConsistency)
	assertPersistedFailedDTO(t, third, runID, 3)

	props, listErr := repo.ListProposalsByAnalysisRun(context.Background(), "t1", runID)
	require.NoError(t, listErr)
	require.Empty(t, props)
}

func TestF3MarkFailedFailureNoForgedDTO(t *testing.T) {
	inner := NewMemoryUnitOfWork()
	uow := &markFailInjectUoW{inner: inner, failMark: true}
	gen := &DeterministicFakeGenerator{Err: errors.New("boom")}
	svc := NewCapabilityService(&Components{UoW: uow, Generator: gen})

	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "f3-mark-fail", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.Error(t, err)
	require.Nil(t, res, "mark-failed failure must not forge a FAILED DTO")

	run, getErr := inner.Root().GetAnalysisRun(context.Background(), "t1", findAnalysisRunID(t, inner.Root(), "t1", "biz", 1, "f3-mark-fail"))
	require.NoError(t, getErr)
	require.Equal(t, entity.AnalysisPending, run.Status, "run must remain PENDING when mark fails")
}

func TestF3RefetchFailureNoForgedDTO(t *testing.T) {
	inner := NewMemoryUnitOfWork()
	wrapped := &getFailAfterMarkRepo{CapabilityRepository: inner.Root()}
	uow := &rootOverrideUoW{inner: inner, root: wrapped}
	gen := &DeterministicFakeGenerator{Err: errors.New("boom")}
	svc := NewCapabilityService(&Components{UoW: uow, Generator: gen})

	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "f3-refetch-fail", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.Error(t, err)
	require.Nil(t, res, "refetch failure after mark must not forge a FAILED DTO")

	// Mark via txn still succeeded; status is FAILED in store, but caller got nil result.
	run, getErr := inner.Root().GetAnalysisRun(context.Background(), "t1", findAnalysisRunID(t, inner.Root(), "t1", "biz", 1, "f3-refetch-fail"))
	require.NoError(t, getErr)
	require.Equal(t, entity.AnalysisFailed, run.Status)
}

// markFailInjectUoW forces MarkAnalysisFailed to fail inside owned transactions.
type markFailInjectUoW struct {
	inner    *MemoryUnitOfWork
	failMark bool
}

func (u *markFailInjectUoW) Root() repository.CapabilityRepository { return u.inner.Root() }
func (u *markFailInjectUoW) Assets() AssetProjection               { return u.inner.Assets() }

func (u *markFailInjectUoW) WithinTransaction(ctx context.Context, fn func(tx CapabilityTx) error) error {
	return u.inner.WithinTransaction(ctx, func(tx CapabilityTx) error {
		repo := tx.Repo()
		if u.failMark {
			repo = &markFailRepo{CapabilityRepository: repo}
		}
		return fn(&capabilityTx{repo: repo, assets: tx.Assets()})
	})
}

type markFailRepo struct {
	repository.CapabilityRepository
}

func (r *markFailRepo) MarkAnalysisFailed(context.Context, string, string, string, int32) error {
	return errors.New("injected mark-failed failure")
}

// getFailAfterMarkRepo lets CreateOrClaim / early gets succeed, then fails the post-mark refetch.
type getFailAfterMarkRepo struct {
	repository.CapabilityRepository
	gets atomic.Int32
}

func (r *getFailAfterMarkRepo) GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*entity.CapabilityAnalysisRun, error) {
	r.gets.Add(1)
	run, err := r.CapabilityRepository.GetAnalysisRun(ctx, tenantID, analysisRunID)
	if err != nil {
		return nil, err
	}
	// Fail the post-mark refetch used by returnPersistedFailedAnalysis (status already FAILED).
	if run != nil && run.Status == entity.AnalysisFailed {
		return nil, errors.New("injected refetch failure")
	}
	return run, nil
}

func findAnalysisRunID(t *testing.T, repo repository.CapabilityRepository, tenantID, businessID string, bmRev int32, clientRequestID string) string {
	t.Helper()
	digest, err := AnalysisRequestDigest(entity.AnalysisRequest{BusinessModelRevision: bmRev})
	require.NoError(t, err)
	raw, err := json.Marshal(entity.AnalysisRequest{BusinessModelRevision: bmRev})
	require.NoError(t, err)
	probe := &entity.CapabilityAnalysisRun{
		AnalysisRunID: "probe", TenantID: tenantID, BusinessID: businessID, BusinessModelRevision: bmRev,
		ClientRequestID: clientRequestID, RequestDigest: digest, RequestJSON: string(raw),
		Status: entity.AnalysisPending, Attempt: 1, CreatedBy: testActor,
	}
	existing, created, err := repo.CreateOrClaimAnalysisRun(context.Background(), probe)
	require.NoError(t, err)
	require.False(t, created, "expected existing run for client_request_id")
	require.NotNil(t, existing)
	return existing.AnalysisRunID
}
