/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package repository

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
)

func TestMemoryRepositoryAnalysisIdempotency(t *testing.T) {
	repo := NewMemoryDataRepository()
	ctx := context.Background()
	base := entity.DataRequirementAnalysisRun{
		AnalysisRunID: "run-1", TenantID: "tenant", BusinessID: "business",
		BusinessModelRevision: 1, ClientRequestID: "request", RequestDigest: "digest-a",
		Status: entity.AnalysisPending,
	}

	got, created, err := repo.CreateOrClaimAnalysisRun(ctx, &base)
	if err != nil || !created || got.AnalysisRunID != base.AnalysisRunID {
		t.Fatalf("first claim = (%+v, %v, %v)", got, created, err)
	}
	duplicate := base
	duplicate.AnalysisRunID = "run-2"
	got, created, err = repo.CreateOrClaimAnalysisRun(ctx, &duplicate)
	if err != nil || created || got.AnalysisRunID != base.AnalysisRunID {
		t.Fatalf("duplicate claim = (%+v, %v, %v)", got, created, err)
	}
	conflict := duplicate
	conflict.RequestDigest = "digest-b"
	if _, _, err = repo.CreateOrClaimAnalysisRun(ctx, &conflict); !errors.Is(err, entity.ErrAnalysisIdempotencyConflict) {
		t.Fatalf("conflict error = %v", err)
	}
}

func TestMemoryRepositoryConcurrentClaimHasOneOwner(t *testing.T) {
	repo := NewMemoryDataRepository()
	ctx := context.Background()
	const workers = 32
	var wg sync.WaitGroup
	var mu sync.Mutex
	owners := 0
	runIDs := map[string]struct{}{}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			run := &entity.DataRequirementAnalysisRun{
				AnalysisRunID: "candidate", TenantID: "tenant", BusinessID: "business",
				BusinessModelRevision: 1, ClientRequestID: "request", RequestDigest: "digest",
				Status: entity.AnalysisPending,
			}
			got, created, err := repo.CreateOrClaimAnalysisRun(ctx, run)
			if err != nil {
				t.Errorf("claim: %v", err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			if created {
				owners++
			}
			runIDs[got.AnalysisRunID] = struct{}{}
		}()
	}
	wg.Wait()
	if owners != 1 || len(runIDs) != 1 {
		t.Fatalf("owners=%d run IDs=%v", owners, runIDs)
	}
}

func TestMemoryRepositoryCASRetryDecisionAndTenantIsolation(t *testing.T) {
	repo := NewMemoryDataRepository()
	ctx := context.Background()
	req := &entity.DataRequirement{
		RequirementID: "req", TenantID: "tenant-a", BusinessID: "business",
		BusinessModelRevision: 1, Status: entity.StatusProposed,
	}
	if err := repo.CreateRequirement(ctx, req); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetRequirement(ctx, "tenant-b", req.RequirementID); !errors.Is(err, entity.ErrRequirementNotFound) {
		t.Fatalf("cross-tenant get error = %v", err)
	}
	ok, err := repo.UpdateRequirementStatusCAS(ctx, "tenant-a", "req", entity.StatusProposed, entity.StatusConfirmed)
	if err != nil || !ok {
		t.Fatalf("CAS = (%v, %v)", ok, err)
	}
	ok, err = repo.UpdateRequirementStatusCAS(ctx, "tenant-a", "req", entity.StatusProposed, entity.StatusRejected)
	if err != nil || ok {
		t.Fatalf("second CAS = (%v, %v)", ok, err)
	}

	run := &entity.DataRequirementAnalysisRun{
		AnalysisRunID: "run", TenantID: "tenant-a", BusinessID: "business",
		BusinessModelRevision: 1, ClientRequestID: "retry", RequestDigest: "digest",
		Status: entity.AnalysisFailed,
	}
	if _, _, err := repo.CreateOrClaimAnalysisRun(ctx, run); err != nil {
		t.Fatal(err)
	}
	claimed, gen, err := repo.ClaimAnalysisRetry(ctx, "tenant-a", "run", "retry-actor")
	if err != nil || !claimed || gen == 0 {
		t.Fatalf("retry = (%v, %v, %v)", claimed, gen, err)
	}
	retried, err := repo.GetAnalysisRun(ctx, "tenant-a", "run")
	if err != nil || retried.Status != entity.AnalysisPending || retried.RetryCount != 1 ||
		retried.LastRetryBy != "retry-actor" || retried.LastRetryAt == nil {
		t.Fatalf("retried run = (%+v, %v)", retried, err)
	}

	decision := &entity.DataRequirementDecision{
		DecisionID: "decision-1", TenantID: "tenant-a", BusinessID: "business",
		SourceRequirementID: "req", Action: entity.DecisionConfirm,
	}
	if err := repo.CreateDecision(ctx, decision); err != nil {
		t.Fatal(err)
	}
	duplicate := *decision
	duplicate.DecisionID = "decision-2"
	if err := repo.CreateDecision(ctx, &duplicate); !errors.Is(err, entity.ErrRequirementAlreadyDecided) {
		t.Fatalf("duplicate decision error = %v", err)
	}
}

func TestMemoryRepositoryTransactionDoesNotDeadlockAndRollsBack(t *testing.T) {
	repo := NewMemoryDataRepository()
	ctx := context.Background()
	errSentinel := errors.New("rollback")
	err := repo.Transaction(ctx, func(tx DataRepository) error {
		if err := tx.CreateRequirement(ctx, &entity.DataRequirement{
			RequirementID: "req", TenantID: "tenant", BusinessID: "business", Status: entity.StatusProposed,
		}); err != nil {
			return err
		}
		return errSentinel
	})
	if !errors.Is(err, errSentinel) {
		t.Fatalf("transaction error = %v", err)
	}
	if _, err := repo.GetRequirement(ctx, "tenant", "req"); !errors.Is(err, entity.ErrRequirementNotFound) {
		t.Fatalf("rolled-back requirement error = %v", err)
	}
}
