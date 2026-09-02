/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	dataentity "github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/data/repository"
)

func setExpiredLease(t *testing.T, repo repository.DataRepository, tenantID, runID string) {
	t.Helper()
	if err := repository.SetLeaseExpiresAtForTest(repo, context.Background(), tenantID, runID, time.Now().UTC().Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
}

func abandonPendingRun(t *testing.T, requestID string) (DataService, repository.DataRepository, *stubBusiness, *dataentity.DataRequirementAnalysisRun) {
	t.Helper()
	entered := make(chan struct{})
	model := &FakeFormaDataModel{AnalyzeFn: func(context.Context, *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
		close(entered)
		select {} // simulate crash before completion
	}}
	svc, repo, business := newService(model)
	input := analyzeInput(labBiz, requestID)
	go func() { _, _ = svc.AnalyzeDataRequirements(context.Background(), input) }()
	<-entered
	time.Sleep(20 * time.Millisecond)
	run, err := repository.LookupAnalysisRunByClientRequest(context.Background(), repo, testTenant, labBiz, 1, requestID)
	if err != nil {
		t.Fatal(err)
	}
	setExpiredLease(t, repo, testTenant, run.AnalysisRunID)
	return svc, repo, business, run
}

func TestFreshPendingCannotBeTakenOver(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	model := &FakeFormaDataModel{AnalyzeFn: func(context.Context, *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
		close(entered)
		<-release
		return &AnalyzeDataRequirementsResponse{ModelRef: "model", Proposals: []dataentity.DataRequirementProposal{labProposal()}}, nil
	}}
	svc, _, _ := newService(model)
	input := analyzeInput(labBiz, "fresh-pending")
	go func() { _, _ = svc.AnalyzeDataRequirements(context.Background(), input) }()
	<-entered
	second, err := svc.AnalyzeDataRequirements(context.Background(), input)
	close(release)
	if err != nil {
		t.Fatal(err)
	}
	if second.OwnedExecute || model.Calls != 1 {
		t.Fatalf("second=%+v calls=%d", second, model.Calls)
	}
}

func TestExpiredPendingCanBeTakenOver(t *testing.T) {
	model := successfulModel(labProposal())
	svc, _, _, run := abandonPendingRun(t, "expired-takeover")
	svcImpl := svc.(*dataServiceImpl)
	svcImpl.model = model
	result, err := svc.AnalyzeDataRequirements(context.Background(), analyzeInput(labBiz, "expired-takeover"))
	if err != nil {
		t.Fatal(err)
	}
	if !result.OwnedExecute || result.Run.Status != dataentity.AnalysisSucceeded || result.Run.AnalysisRunID != run.AnalysisRunID {
		t.Fatalf("result=%+v", result)
	}
	if model.Calls != 1 {
		t.Fatalf("model calls=%d want 1 takeover call", model.Calls)
	}
}

func TestConcurrentExpiredTakeoverOnlyOneOwner(t *testing.T) {
	attempt := 0
	var mu sync.Mutex
	model := &FakeFormaDataModel{AnalyzeFn: func(context.Context, *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
		mu.Lock()
		attempt++
		mu.Unlock()
		time.Sleep(50 * time.Millisecond)
		return &AnalyzeDataRequirementsResponse{ModelRef: "model", Proposals: []dataentity.DataRequirementProposal{labProposal()}}, nil
	}}
	svc, _, _, _ := abandonPendingRun(t, "concurrent-expired")
	svcImpl := svc.(*dataServiceImpl)
	svcImpl.model = model
	input := analyzeInput(labBiz, "concurrent-expired")
	var wg sync.WaitGroup
	results := make([]*AnalyzeResult, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			res, err := svc.AnalyzeDataRequirements(context.Background(), input)
			if err != nil {
				t.Error(err)
				return
			}
			results[idx] = res
		}(i)
	}
	wg.Wait()
	owners := 0
	for _, res := range results {
		if res != nil && res.OwnedExecute {
			owners++
		}
	}
	if owners != 1 || attempt != 1 {
		t.Fatalf("owners=%d attempt=%d results=%+v", owners, attempt, results)
	}
}

func TestOldOwnerCannotMarkSucceededAfterTakeover(t *testing.T) {
	repo := repository.NewMemoryDataRepository()
	oldGen := int32(1)
	now := time.Now().UTC()
	exp := now.Add(-time.Minute)
	run := &dataentity.DataRequirementAnalysisRun{
		AnalysisRunID: "run-fence", TenantID: testTenant, BusinessID: labBiz,
		BusinessModelRevision: 1, ClientRequestID: "fence", RequestDigest: "d",
		Status: dataentity.AnalysisPending, ExecutionGeneration: 2,
		ExecutionClaimedAt: &now, LeaseExpiresAt: &exp, CreatedBy: "analyst",
		CreatedAt: now, UpdatedAt: now,
	}
	seedExecutionLease(run, now)
	if _, _, err := repo.CreateOrClaimAnalysisRun(context.Background(), run); err != nil {
		t.Fatal(err)
	}
	if err := repo.MarkAnalysisSucceeded(context.Background(), testTenant, run.AnalysisRunID, "old-model", oldGen); !errors.Is(err, dataentity.ErrRequirementInvalidState) {
		t.Fatalf("old owner succeeded: %v", err)
	}
}

func TestOldOwnerCannotMarkFailedAfterTakeover(t *testing.T) {
	repo := repository.NewMemoryDataRepository()
	oldGen := int32(1)
	now := time.Now().UTC()
	exp := now.Add(-time.Minute)
	run := &dataentity.DataRequirementAnalysisRun{
		AnalysisRunID: "run-fence-fail", TenantID: testTenant, BusinessID: labBiz,
		BusinessModelRevision: 1, ClientRequestID: "fence-fail", RequestDigest: "d",
		Status: dataentity.AnalysisPending, ExecutionGeneration: 2,
		ExecutionClaimedAt: &now, LeaseExpiresAt: &exp, CreatedBy: "analyst",
		CreatedAt: now, UpdatedAt: now,
	}
	if _, _, err := repo.CreateOrClaimAnalysisRun(context.Background(), run); err != nil {
		t.Fatal(err)
	}
	if err := repo.MarkAnalysisFailed(context.Background(), testTenant, run.AnalysisRunID, "OLD", "old", oldGen); !errors.Is(err, dataentity.ErrRequirementInvalidState) {
		t.Fatalf("old owner failed: %v", err)
	}
}

func TestCurrentOwnerCanComplete(t *testing.T) {
	repo := repository.NewMemoryDataRepository()
	now := time.Now().UTC()
	run := &dataentity.DataRequirementAnalysisRun{
		AnalysisRunID: "run-current", TenantID: testTenant, BusinessID: labBiz,
		BusinessModelRevision: 1, ClientRequestID: "current", RequestDigest: "d",
		Status: dataentity.AnalysisPending, ExecutionGeneration: 1, CreatedBy: "analyst",
		CreatedAt: now, UpdatedAt: now, StartedAt: &now,
	}
	seedExecutionLease(run, now)
	if _, _, err := repo.CreateOrClaimAnalysisRun(context.Background(), run); err != nil {
		t.Fatal(err)
	}
	if err := repo.MarkAnalysisSucceeded(context.Background(), testTenant, run.AnalysisRunID, "model", 1); err != nil {
		t.Fatal(err)
	}
	final, err := repo.GetAnalysisRun(context.Background(), testTenant, run.AnalysisRunID)
	if err != nil || final.Status != dataentity.AnalysisSucceeded {
		t.Fatalf("final=%+v err=%v", final, err)
	}
}

func TestConcurrentFailedRetryOnlyOneOwner(t *testing.T) {
	model := &FakeFormaDataModel{AnalyzeFn: func(context.Context, *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
		return nil, errors.New("fail")
	}}
	svc, _, _ := newService(model)
	failed, err := svc.AnalyzeDataRequirements(context.Background(), analyzeInput(labBiz, "retry-concurrent"))
	if !errors.Is(err, dataentity.ErrModelFailed) {
		t.Fatal(err)
	}
	model.AnalyzeFn = func(context.Context, *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
		time.Sleep(50 * time.Millisecond)
		return &AnalyzeDataRequirementsResponse{ModelRef: "model", Proposals: []dataentity.DataRequirementProposal{labProposal()}}, nil
	}
	model.Calls = 0
	var wg sync.WaitGroup
	successes := 0
	var mu sync.Mutex
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := svc.RetryFailedAnalysis(context.Background(), testTenant, failed.Run.AnalysisRunID, "retry-actor")
			if err == nil && res.Run.Status == dataentity.AnalysisSucceeded {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if successes != 1 {
		t.Fatalf("successes=%d", successes)
	}
}

func TestRetryCountIncrements(t *testing.T) {
	model := &FakeFormaDataModel{AnalyzeFn: func(context.Context, *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
		return nil, errors.New("fail")
	}}
	svc, _, _ := newService(model)
	failed, _ := svc.AnalyzeDataRequirements(context.Background(), analyzeInput(labBiz, "retry-count"))
	model.AnalyzeFn = func(context.Context, *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
		return &AnalyzeDataRequirementsResponse{ModelRef: "model", Proposals: []dataentity.DataRequirementProposal{labProposal()}}, nil
	}
	result, err := svc.RetryFailedAnalysis(context.Background(), testTenant, failed.Run.AnalysisRunID, "retry-actor")
	if err != nil || result.Run.RetryCount != 1 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestRetryActorRecorded(t *testing.T) {
	model := &FakeFormaDataModel{AnalyzeFn: func(context.Context, *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
		return nil, errors.New("fail")
	}}
	svc, _, _ := newService(model)
	failed, _ := svc.AnalyzeDataRequirements(context.Background(), analyzeInput(labBiz, "retry-actor"))
	model.AnalyzeFn = func(context.Context, *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
		return &AnalyzeDataRequirementsResponse{ModelRef: "model", Proposals: []dataentity.DataRequirementProposal{labProposal()}}, nil
	}
	result, err := svc.RetryFailedAnalysis(context.Background(), testTenant, failed.Run.AnalysisRunID, "auditor-42")
	if err != nil || result.Run.LastRetryBy != "auditor-42" || result.Run.LastRetryAt == nil {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if result.Run.CreatedBy != failed.Run.CreatedBy {
		t.Fatalf("created_by mutated: got=%q want=%q", result.Run.CreatedBy, failed.Run.CreatedBy)
	}
}

func TestRetryTimeRecorded(t *testing.T) {
	before := time.Now().UTC()
	model := &FakeFormaDataModel{AnalyzeFn: func(context.Context, *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
		return nil, errors.New("fail")
	}}
	svc, _, _ := newService(model)
	failed, _ := svc.AnalyzeDataRequirements(context.Background(), analyzeInput(labBiz, "retry-time"))
	model.AnalyzeFn = func(context.Context, *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
		return &AnalyzeDataRequirementsResponse{ModelRef: "model", Proposals: []dataentity.DataRequirementProposal{labProposal()}}, nil
	}
	result, err := svc.RetryFailedAnalysis(context.Background(), testTenant, failed.Run.AnalysisRunID, "auditor")
	if err != nil || result.Run.LastRetryAt == nil || result.Run.LastRetryAt.Before(before) {
		t.Fatalf("result=%+v err=%v before=%v", result, err, before)
	}
}

func TestNoDuplicateRequirementsAfterTakeover(t *testing.T) {
	model := successfulModel(labProposal())
	svc, repo, _, _ := abandonPendingRun(t, "no-dup-reqs")
	svcImpl := svc.(*dataServiceImpl)
	svcImpl.model = model
	result, err := svc.AnalyzeDataRequirements(context.Background(), analyzeInput(labBiz, "no-dup-reqs"))
	if err != nil {
		t.Fatal(err)
	}
	reqs, _ := repo.ListRequirementsByRevision(context.Background(), testTenant, labBiz, 1)
	count := 0
	for _, r := range reqs {
		if r.AnalysisRunID == result.Run.AnalysisRunID {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("requirements for run=%d", count)
	}
}

func TestNoDuplicateModelCallForFreshPending(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	model := &FakeFormaDataModel{AnalyzeFn: func(context.Context, *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
		close(entered)
		<-release
		return &AnalyzeDataRequirementsResponse{ModelRef: "model", Proposals: []dataentity.DataRequirementProposal{labProposal()}}, nil
	}}
	svc, _, _ := newService(model)
	input := analyzeInput(labBiz, "no-dup-model")
	go func() { _, _ = svc.AnalyzeDataRequirements(context.Background(), input) }()
	<-entered
	for i := 0; i < 3; i++ {
		res, err := svc.AnalyzeDataRequirements(context.Background(), input)
		if err != nil {
			t.Fatal(err)
		}
		if res.OwnedExecute {
			t.Fatal("fresh pending must not re-execute")
		}
	}
	close(release)
	if model.Calls != 1 {
		t.Fatalf("model calls=%d", model.Calls)
	}
}

func TestLeaseRecoveryBusinessModelRevisionUnchanged(t *testing.T) {
	model := successfulModel(labProposal())
	svc, _, business, _ := abandonPendingRun(t, "bm-unchanged")
	svcImpl := svc.(*dataServiceImpl)
	svcImpl.model = model
	before := business.currentRevision
	if _, err := svc.AnalyzeDataRequirements(context.Background(), analyzeInput(labBiz, "bm-unchanged")); err != nil {
		t.Fatal(err)
	}
	if business.currentRevision != before {
		t.Fatalf("revision changed %d -> %d", before, business.currentRevision)
	}
}

func TestLeaseRecoveryTenantIsolation(t *testing.T) {
	_, _, _, run := abandonPendingRun(t, "tenant-iso")
	svc, _, _ := newService(successfulModel(labProposal()))
	_, err := svc.GetAnalysisRun(context.Background(), "tenant-b", run.AnalysisRunID)
	if !errors.Is(err, dataentity.ErrAnalysisNotFound) {
		t.Fatalf("cross-tenant get=%v", err)
	}
}
