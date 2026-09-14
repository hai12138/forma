/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/fixture"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/repository"
)

func assertPersistedFailedDTOExact(t *testing.T, res *AnalysisResult, wantRunID string, wantAttempt int32) {
	t.Helper()
	require.NotNil(t, res, "persisted FAILED must return AnalysisResult DTO")
	require.NotNil(t, res.Run)
	require.Equal(t, entity.AnalysisFailed, res.Run.Status)
	require.True(t, res.OwnedExecute)
	require.NotEmpty(t, res.Run.AnalysisRunID)
	if wantRunID != "" {
		require.Equal(t, wantRunID, res.Run.AnalysisRunID)
	}
	require.Equal(t, wantAttempt, res.Run.Attempt, "FAILED DTO Attempt must equal owned attempt exactly")
	require.NotEmpty(t, res.Run.ErrorCode)
}

func TestF4GeneratorFailReturnsExactAttempt(t *testing.T) {
	gen := &DeterministicFakeGenerator{Err: errors.New("boom")}
	svc, repo, _, _, _ := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "f4-gen-fail", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.ErrorIs(t, err, entity.ErrAnalysisFailed)
	assertPersistedFailedDTOExact(t, res, "", 1)

	props, listErr := repo.ListProposalsByAnalysisRun(context.Background(), "t1", res.Run.AnalysisRunID)
	require.NoError(t, listErr)
	require.Empty(t, props)
}

func TestF4InvalidProposalReturnsExactAttempt(t *testing.T) {
	gen := &DeterministicFakeGenerator{Err: errors.New("boom")}
	svc, repo, _, _, _ := newTestService(gen)
	first, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "f4-invprop", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.ErrorIs(t, err, entity.ErrAnalysisFailed)
	assertPersistedFailedDTOExact(t, first, "", 1)
	runID := first.Run.AnalysisRunID

	bad := fixture.LaboratoryFlowCapability()
	bad.Name = ""
	gen.Err = nil
	gen.Proposals = []entity.SemanticPayload{bad}

	retried, err := svc.RetryFailedAnalysis(context.Background(), "t1", runID, testActor)
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
	assertPersistedFailedDTOExact(t, retried, runID, 2)
	require.Equal(t, "FORMA_CAPABILITY_INVALID_REQUEST", retried.Run.ErrorCode)

	props, listErr := repo.ListProposalsByAnalysisRun(context.Background(), "t1", runID)
	require.NoError(t, listErr)
	require.Empty(t, props)
}

func TestF4RetryCorruptRequestJSONExactAttempt(t *testing.T) {
	gen := &DeterministicFakeGenerator{Err: errors.New("boom")}
	svc, repo, _, _, _ := newTestService(gen)
	first, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "f4-retry-corrupt", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.ErrorIs(t, err, entity.ErrAnalysisFailed)
	assertPersistedFailedDTOExact(t, first, "", 1)
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
	assertPersistedFailedDTOExact(t, retried, runID, 2)
	require.Equal(t, "FORMA_CAPABILITY_INVALID_REQUEST", retried.Run.ErrorCode)
	require.Equal(t, callsBefore, gen.CallCount())

	props, listErr := repo.ListProposalsByAnalysisRun(context.Background(), "t1", runID)
	require.NoError(t, listErr)
	require.Empty(t, props)
}

func TestF4RetryDigestMismatchExactAttempt(t *testing.T) {
	gen := &DeterministicFakeGenerator{Err: errors.New("boom")}
	svc, repo, _, _, _ := newTestService(gen)
	first, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "f4-retry-digest", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.ErrorIs(t, err, entity.ErrAnalysisFailed)
	assertPersistedFailedDTOExact(t, first, "", 1)
	runID := first.Run.AnalysisRunID

	run, err := repo.GetAnalysisRun(context.Background(), "t1", runID)
	require.NoError(t, err)
	run.RequestJSON = `{"business_model_revision":1,"data_contract_pins":[{"data_contract_id":"dc_other","data_contract_version":1}],"requirement_refs":null}`
	require.NoError(t, repository.ReplaceAnalysisRunForTest(repo, run))

	gen.Err = nil
	gen.Proposals = []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}
	callsBefore := gen.CallCount()

	retried, err := svc.RetryFailedAnalysis(context.Background(), "t1", runID, testActor)
	require.ErrorIs(t, err, entity.ErrConsistency)
	assertPersistedFailedDTOExact(t, retried, runID, 2)
	require.Equal(t, callsBefore, gen.CallCount())
}

func TestF4LeaseTakeoverIllegalRequestExactAttempt(t *testing.T) {
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
		AnalysisRunID: "run-f4-lease", TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1,
		ClientRequestID: "f4-lease-bad", RequestDigest: digest, Status: entity.AnalysisPending, Attempt: 1,
		RequestJSON: "{not-json", ExecutionClaimedAt: &claimedAt, LeaseExpiresAt: &exp,
		CreatedBy: testActor, CreatedAt: fixed.Now(), UpdatedAt: fixed.Now(),
	}
	_, created, err := uow.Root().CreateOrClaimAnalysisRun(context.Background(), seed)
	require.NoError(t, err)
	require.True(t, created)
	require.NoError(t, uow.Root().CreateAnalysisAttempt(context.Background(), &entity.CapabilityAnalysisAttempt{
		AttemptID: "att-f4-lease", AnalysisRunID: "run-f4-lease", TenantID: "t1", Attempt: 1,
		ActorPrincipalID: testActor, TriggerKind: entity.AttemptTriggerFirst,
		ResultStatus: entity.AttemptResultPending, CreatedAt: fixed.Now(),
	}))

	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "f4-lease-bad", ActorID: testActor,
		Analysis: caller,
	})
	require.ErrorIs(t, err, entity.ErrConsistency)
	assertPersistedFailedDTOExact(t, res, "run-f4-lease", 2)
	require.Equal(t, "FORMA_CAPABILITY_INVALID_REQUEST", res.Run.ErrorCode)

	props, listErr := uow.Root().ListProposalsByAnalysisRun(context.Background(), "t1", "run-f4-lease")
	require.NoError(t, listErr)
	require.Empty(t, props)
}

func TestF4SameRunIDAcrossRetryExactAttempts(t *testing.T) {
	gen := &DeterministicFakeGenerator{Err: errors.New("boom")}
	svc, repo, _, _, _ := newTestService(gen)
	first, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "f4-same-run", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.ErrorIs(t, err, entity.ErrAnalysisFailed)
	assertPersistedFailedDTOExact(t, first, "", 1)
	runID := first.Run.AnalysisRunID

	bad := fixture.LaboratoryFlowCapability()
	bad.Name = ""
	gen.Err = nil
	gen.Proposals = []entity.SemanticPayload{bad}
	second, err := svc.RetryFailedAnalysis(context.Background(), "t1", runID, testActor)
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
	assertPersistedFailedDTOExact(t, second, runID, 2)

	run, err := repo.GetAnalysisRun(context.Background(), "t1", runID)
	require.NoError(t, err)
	run.RequestJSON = "{broken"
	require.NoError(t, repository.ReplaceAnalysisRunForTest(repo, run))

	third, err := svc.RetryFailedAnalysis(context.Background(), "t1", runID, testActor)
	require.ErrorIs(t, err, entity.ErrConsistency)
	assertPersistedFailedDTOExact(t, third, runID, 3)

	props, listErr := repo.ListProposalsByAnalysisRun(context.Background(), "t1", runID)
	require.NoError(t, listErr)
	require.Empty(t, props)
}

// attemptBumpOnFailedGetRepo simulates a concurrent retry advancing Attempt after mark-failed:
// post-mark refetch sees FAILED with Attempt N+1 while the older caller still owns N.
type attemptBumpOnFailedGetRepo struct {
	repository.CapabilityRepository
}

func (r *attemptBumpOnFailedGetRepo) GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*entity.CapabilityAnalysisRun, error) {
	run, err := r.CapabilityRepository.GetAnalysisRun(ctx, tenantID, analysisRunID)
	if err != nil || run == nil {
		return run, err
	}
	if run.Status == entity.AnalysisFailed {
		bumped := *run
		bumped.Attempt = run.Attempt + 1
		return &bumped, nil
	}
	return run, nil
}

func TestF4InterleavedRaceRefetchNewerAttemptFailsClosed(t *testing.T) {
	inner := NewMemoryUnitOfWork()
	wrapped := &attemptBumpOnFailedGetRepo{CapabilityRepository: inner.Root()}
	uow := &rootOverrideUoW{inner: inner, root: wrapped}
	gen := &DeterministicFakeGenerator{Err: errors.New("boom")}
	svc := NewCapabilityService(&Components{UoW: uow, Generator: gen})

	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "f4-race-attempt", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.ErrorIs(t, err, entity.ErrConsistency)
	require.Nil(t, res, "older caller must not receive newer-attempt FAILED DTO")

	// Store still has the owned attempt's mark (Attempt 1); only the Root refetch was bumped.
	run, getErr := inner.Root().GetAnalysisRun(context.Background(), "t1", findAnalysisRunID(t, inner.Root(), "t1", "biz", 1, "f4-race-attempt"))
	require.NoError(t, getErr)
	require.Equal(t, entity.AnalysisFailed, run.Status)
	require.Equal(t, int32(1), run.Attempt)

	props, listErr := inner.Root().ListProposalsByAnalysisRun(context.Background(), "t1", run.AnalysisRunID)
	require.NoError(t, listErr)
	require.Empty(t, props)
}

func TestF4StatusMismatchAfterMarkFailsClosed(t *testing.T) {
	inner := NewMemoryUnitOfWork()
	wrapped := &statusOverrideOnFailedGetRepo{CapabilityRepository: inner.Root(), override: entity.AnalysisPending}
	uow := &rootOverrideUoW{inner: inner, root: wrapped}
	gen := &DeterministicFakeGenerator{Err: errors.New("boom")}
	svc := NewCapabilityService(&Components{UoW: uow, Generator: gen})

	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "f4-status-mismatch", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.ErrorIs(t, err, entity.ErrConsistency)
	require.Nil(t, res, "status mismatch must fail closed without forging DTO")
}

// statusOverrideOnFailedGetRepo rewrites Status on FAILED refetch to force status mismatch.
type statusOverrideOnFailedGetRepo struct {
	repository.CapabilityRepository
	override entity.AnalysisStatus
}

func (r *statusOverrideOnFailedGetRepo) GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*entity.CapabilityAnalysisRun, error) {
	run, err := r.CapabilityRepository.GetAnalysisRun(ctx, tenantID, analysisRunID)
	if err != nil || run == nil {
		return run, err
	}
	if run.Status == entity.AnalysisFailed {
		mut := *run
		mut.Status = r.override
		return &mut, nil
	}
	return run, nil
}
