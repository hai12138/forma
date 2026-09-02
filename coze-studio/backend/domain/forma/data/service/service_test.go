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

	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	dataentity "github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/data/repository"
)

const (
	testTenant = "tenant-a"
	labBiz     = "lab"
	poBiz      = "procurement"
)

type stubBusiness struct {
	mu               sync.Mutex
	getRevisionCalls int
	currentRevision  int32
}

func newStubBusiness() *stubBusiness { return &stubBusiness{currentRevision: 1} }

func (s *stubBusiness) semantic(businessID string) *businessentity.SemanticModel {
	switch businessID {
	case poBiz:
		return &businessentity.SemanticModel{
			SchemaVersion: businessentity.SemanticSchemaVersion,
			Nodes: []businessentity.SemanticNode{
				{ID: "node_po", Type: businessentity.NodeBusinessObject, Name: "Purchase"},
				{ID: "node_approver", Type: businessentity.NodeActor, Name: "Approver"},
			},
			Rules: []businessentity.BusinessRule{{ID: "rule_po_threshold", Name: "Threshold"}},
		}
	default:
		return &businessentity.SemanticModel{
			SchemaVersion: businessentity.SemanticSchemaVersion,
			Nodes: []businessentity.SemanticNode{
				{ID: "node_lab_sample", Type: businessentity.NodeBusinessObject, Name: "Sample"},
				{ID: "node_lab_tech", Type: businessentity.NodeActor, Name: "Technician"},
			},
			Rules: []businessentity.BusinessRule{{ID: "rule_lab_hold", Name: "Hold"}},
		}
	}
}

func (s *stubBusiness) InitBusiness(context.Context, string, string, string, string, *businessentity.SemanticModel, string) (*businessentity.BusinessModel, *businessentity.BusinessModelRevision, *businessentity.BusinessModelLayout, error) {
	return nil, nil, nil, businessentity.ErrNotFound
}
func (s *stubBusiness) Get(_ context.Context, tenantID, businessID string) (*businessentity.BusinessModel, error) {
	if tenantID == "" || businessID == "" {
		return nil, businessentity.ErrNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return &businessentity.BusinessModel{TenantID: tenantID, BusinessID: businessID, CurrentRevision: s.currentRevision}, nil
}
func (s *stubBusiness) List(context.Context, string) ([]*businessentity.BusinessModel, error) {
	return nil, businessentity.ErrNotFound
}
func (s *stubBusiness) GetModel(ctx context.Context, tenantID, businessID string) (*businessentity.BusinessModel, *businessentity.SemanticModel, *businessentity.BusinessModelRevision, error) {
	m, err := s.Get(ctx, tenantID, businessID)
	if err != nil {
		return nil, nil, nil, err
	}
	rev, sm, err := s.GetRevision(ctx, tenantID, businessID, m.CurrentRevision)
	return m, sm, rev, err
}
func (s *stubBusiness) SaveModel(context.Context, string, string, string, int32, *businessentity.SemanticModel, string) (*businessentity.BusinessModelRevision, bool, error) {
	return nil, false, businessentity.ErrNotFound
}
func (s *stubBusiness) ListRevisions(context.Context, string, string) ([]*businessentity.BusinessModelRevision, error) {
	return nil, businessentity.ErrNotFound
}
func (s *stubBusiness) GetRevision(_ context.Context, tenantID, businessID string, revision int32) (*businessentity.BusinessModelRevision, *businessentity.SemanticModel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.getRevisionCalls++
	if tenantID == "" || revision != s.currentRevision {
		return nil, nil, businessentity.ErrRevisionNotFound
	}
	return &businessentity.BusinessModelRevision{
		TenantID: tenantID, BusinessID: businessID, RevisionNo: revision,
	}, s.semantic(businessID), nil
}
func (s *stubBusiness) Diff(context.Context, string, string, int32, int32) (*businessentity.BusinessModelDiff, *businessentity.BusinessImpactSummary, error) {
	return nil, nil, businessentity.ErrNotFound
}
func (s *stubBusiness) GetLayout(context.Context, string, string) (*businessentity.BusinessModelLayout, *businessentity.ViewLayout, error) {
	return nil, nil, businessentity.ErrNotFound
}
func (s *stubBusiness) SaveLayout(context.Context, string, string, string, int32, int32, *businessentity.ViewLayout) (*businessentity.BusinessModelLayout, error) {
	return nil, businessentity.ErrNotFound
}

func labProposal() dataentity.DataRequirementProposal {
	return dataentity.DataRequirementProposal{
		RequirementKind: dataentity.KindEntity, SemanticName: "sample_record",
		BusinessElementRefs: []string{"node_lab_sample", "rule_lab_hold"},
	}
}

func poProposal() dataentity.DataRequirementProposal {
	return dataentity.DataRequirementProposal{
		RequirementKind: dataentity.KindMetric, SemanticName: "approval_threshold",
		BusinessElementRefs: []string{"node_po", "node_approver", "rule_po_threshold"},
	}
}

func newService(model *FakeFormaDataModel) (DataService, repository.DataRepository, *stubBusiness) {
	repo := repository.NewMemoryDataRepository()
	business := newStubBusiness()
	return NewDataService(&Components{Repo: repo, BusinessSVC: business, Model: model}), repo, business
}

func analyzeInput(businessID, requestID string) *AnalyzeInput {
	return &AnalyzeInput{
		TenantID: testTenant, BusinessID: businessID, BusinessModelRevision: 1,
		ClientRequestID: requestID, ActorID: "analyst",
	}
}

func successfulModel(proposal dataentity.DataRequirementProposal) *FakeFormaDataModel {
	return &FakeFormaDataModel{AnalyzeFn: func(context.Context, *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
		return &AnalyzeDataRequirementsResponse{ModelRef: "fixture-model", Proposals: []dataentity.DataRequirementProposal{proposal}}, nil
	}}
}

func TestRequestDigestUsesStableAnalyzeInputJSON(t *testing.T) {
	input := analyzeInput(labBiz, "digest")
	if got, want := requestDigest(input), DigestJSON(input); got != want {
		t.Fatalf("request digest=%q want stable JSON digest=%q", got, want)
	}
	changed := *input
	changed.ActorID = "another-actor"
	if requestDigest(&changed) == requestDigest(input) {
		t.Fatal("request payload change did not change digest")
	}
}

func TestAnalyzeAcceptsOnlyProposedAIOutput(t *testing.T) {
	t.Run("domain persists AI output as proposed", func(t *testing.T) {
		svc, _, _ := newService(successfulModel(labProposal()))
		result, err := svc.AnalyzeDataRequirements(context.Background(), analyzeInput(labBiz, "proposed"))
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Requirements) != 1 || result.Requirements[0].Status != dataentity.StatusProposed || result.Requirements[0].Source != dataentity.SourceAIGenerated {
			t.Fatalf("requirements = %+v", result.Requirements)
		}
	})

	for _, forbidden := range []dataentity.RequirementStatus{dataentity.StatusConfirmed, dataentity.StatusRejected} {
		t.Run("model cannot set "+string(forbidden), func(t *testing.T) {
			proposal := labProposal()
			proposal.Status = forbidden
			svc, repo, _ := newService(successfulModel(proposal))
			result, err := svc.AnalyzeDataRequirements(context.Background(), analyzeInput(labBiz, "forbidden-"+string(forbidden)))
			if !errors.Is(err, dataentity.ErrInvalidProposal) || result.Run.Status != dataentity.AnalysisFailed {
				t.Fatalf("result=%+v error=%v", result, err)
			}
			reqs, _ := repo.ListRequirementsByRevision(context.Background(), testTenant, labBiz, 1)
			if len(reqs) != 0 {
				t.Fatalf("persisted forbidden proposals: %+v", reqs)
			}
		})
	}
}

func TestAnalyzeFailureAndAtomicSuccess(t *testing.T) {
	tests := []struct {
		name      string
		model     *FakeFormaDataModel
		wantError error
		wantCount int
		wantState dataentity.AnalysisRunStatus
	}{
		{
			name: "invalid element refs fail without requirements",
			model: successfulModel(dataentity.DataRequirementProposal{
				RequirementKind: dataentity.KindEntity, SemanticName: "bad",
				BusinessElementRefs: []string{"node_missing"},
			}),
			wantError: dataentity.ErrBusinessElementRefInvalid, wantState: dataentity.AnalysisFailed,
		},
		{
			name: "model failure fails without requirements",
			model: &FakeFormaDataModel{AnalyzeFn: func(context.Context, *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
				return nil, errors.New("provider unavailable")
			}},
			wantError: dataentity.ErrModelFailed, wantState: dataentity.AnalysisFailed,
		},
		{
			name: "success atomically persists run and requirements",
			model: &FakeFormaDataModel{AnalyzeFn: func(context.Context, *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
				return &AnalyzeDataRequirementsResponse{ModelRef: "model", Proposals: []dataentity.DataRequirementProposal{labProposal(), labProposal()}}, nil
			}},
			wantCount: 2, wantState: dataentity.AnalysisSucceeded,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo, _ := newService(tc.model)
			result, err := svc.AnalyzeDataRequirements(context.Background(), analyzeInput(labBiz, tc.name))
			if !errors.Is(err, tc.wantError) {
				t.Fatalf("error=%v want=%v", err, tc.wantError)
			}
			if result.Run.Status != tc.wantState {
				t.Fatalf("run status=%s", result.Run.Status)
			}
			reqs, listErr := repo.ListRequirementsByRevision(context.Background(), testTenant, labBiz, 1)
			if listErr != nil || len(reqs) != tc.wantCount {
				t.Fatalf("requirements=%d error=%v", len(reqs), listErr)
			}
		})
	}
}

func TestAnalyzeIdempotencySequentialAndConcurrent(t *testing.T) {
	t.Run("sequential calls execute model once", func(t *testing.T) {
		model := successfulModel(labProposal())
		svc, _, _ := newService(model)
		input := analyzeInput(labBiz, "same-sequential")
		first, err := svc.AnalyzeDataRequirements(context.Background(), input)
		if err != nil {
			t.Fatal(err)
		}
		second, err := svc.AnalyzeDataRequirements(context.Background(), input)
		if err != nil {
			t.Fatal(err)
		}
		if model.Calls != 1 || first.Run.AnalysisRunID != second.Run.AnalysisRunID || second.OwnedExecute {
			t.Fatalf("calls=%d first=%+v second=%+v", model.Calls, first, second)
		}
	})

	t.Run("concurrent calls have one execution owner", func(t *testing.T) {
		entered := make(chan struct{})
		release := make(chan struct{})
		model := &FakeFormaDataModel{AnalyzeFn: func(context.Context, *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
			close(entered)
			<-release
			return &AnalyzeDataRequirementsResponse{ModelRef: "model", Proposals: []dataentity.DataRequirementProposal{labProposal()}}, nil
		}}
		svc, _, _ := newService(model)
		input := analyzeInput(labBiz, "same-concurrent")
		var wg sync.WaitGroup
		errs := make(chan error, 2)
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.AnalyzeDataRequirements(context.Background(), input)
			errs <- err
		}()
		<-entered
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.AnalyzeDataRequirements(context.Background(), input)
			errs <- err
		}()
		close(release)
		wg.Wait()
		close(errs)
		for err := range errs {
			if err != nil {
				t.Fatal(err)
			}
		}
		if model.Calls != 1 {
			t.Fatalf("model calls=%d", model.Calls)
		}
	})
}

func TestRetryFailedAnalysis(t *testing.T) {
	attempt := 0
	model := &FakeFormaDataModel{AnalyzeFn: func(context.Context, *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
		attempt++
		if attempt == 1 {
			return nil, errors.New("temporary")
		}
		return &AnalyzeDataRequirementsResponse{ModelRef: "model", Proposals: []dataentity.DataRequirementProposal{labProposal()}}, nil
	}}
	svc, _, _ := newService(model)
	failed, err := svc.AnalyzeDataRequirements(context.Background(), analyzeInput(labBiz, "retry"))
	if !errors.Is(err, dataentity.ErrModelFailed) {
		t.Fatalf("initial error=%v", err)
	}
	result, err := svc.RetryFailedAnalysis(context.Background(), testTenant, failed.Run.AnalysisRunID, "human")
	if err != nil || result.Run.Status != dataentity.AnalysisSucceeded || len(result.Requirements) != 1 || model.Calls != 2 {
		t.Fatalf("result=%+v calls=%d error=%v", result, model.Calls, err)
	}
}

func createManual(t *testing.T, svc DataService, tenantID, requestName string) *dataentity.DataRequirement {
	t.Helper()
	req, err := svc.CreateManualRequirement(context.Background(), &ManualCreateInput{
		TenantID: tenantID, BusinessID: labBiz, BusinessModelRevision: 1, ActorID: "human",
		RequirementKind: dataentity.KindEntity, SemanticName: requestName,
		BusinessElementRefs: []string{"node_lab_sample"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if req.Status != dataentity.StatusProposed || req.Source != dataentity.SourceManualCreated {
		t.Fatalf("manual requirement=%+v", req)
	}
	return req
}

func TestHumanDecisionsAreAtomicImmutableAndAuditable(t *testing.T) {
	t.Run("confirm", func(t *testing.T) {
		svc, _, _ := newService(successfulModel(labProposal()))
		req := createManual(t, svc, testTenant, "confirm")
		confirmed, decision, err := svc.ConfirmRequirement(context.Background(), testTenant, req.RequirementID, "reviewer", "valid")
		if err != nil || confirmed.Status != dataentity.StatusConfirmed || decision.Action != dataentity.DecisionConfirm {
			t.Fatalf("requirement=%+v decision=%+v error=%v", confirmed, decision, err)
		}
		decisions, err := svc.ListDecisions(context.Background(), testTenant, req.RequirementID)
		if err != nil || len(decisions) != 1 {
			t.Fatalf("decisions=%+v error=%v", decisions, err)
		}
	})

	t.Run("reject", func(t *testing.T) {
		svc, _, _ := newService(successfulModel(labProposal()))
		req := createManual(t, svc, testTenant, "reject")
		rejected, decision, err := svc.RejectRequirement(context.Background(), testTenant, req.RequirementID, "reviewer", "invalid")
		if err != nil || rejected.Status != dataentity.StatusRejected || decision.Action != dataentity.DecisionReject {
			t.Fatalf("requirement=%+v decision=%+v error=%v", rejected, decision, err)
		}
	})

	t.Run("edit confirm preserves lineage", func(t *testing.T) {
		svc, _, _ := newService(successfulModel(labProposal()))
		req := createManual(t, svc, testTenant, "original")
		original, replacement, decision, err := svc.EditConfirmRequirement(context.Background(), &EditConfirmInput{
			TenantID: testTenant, SourceRequirementID: req.RequirementID, ActorID: "reviewer",
			Reason: "refined", SemanticName: "replacement",
		})
		if err != nil || original.Status != dataentity.StatusSuperseded ||
			replacement.Status != dataentity.StatusConfirmed || replacement.Source != dataentity.SourceManualModified ||
			replacement.DerivedFromRequirementID != original.RequirementID ||
			decision.Action != dataentity.DecisionEditConfirm || decision.TargetRequirementID != replacement.RequirementID {
			t.Fatalf("original=%+v replacement=%+v decision=%+v error=%v", original, replacement, decision, err)
		}
		decisions, _ := svc.ListDecisions(context.Background(), testTenant, replacement.RequirementID)
		if len(decisions) != 1 {
			t.Fatalf("replacement decisions=%+v", decisions)
		}
	})

	t.Run("decision source is immutable", func(t *testing.T) {
		_, repo, _ := newService(successfulModel(labProposal()))
		first := &dataentity.DataRequirementDecision{DecisionID: "one", TenantID: testTenant, SourceRequirementID: "source"}
		second := &dataentity.DataRequirementDecision{DecisionID: "two", TenantID: testTenant, SourceRequirementID: "source"}
		if err := repo.CreateDecision(context.Background(), first); err != nil {
			t.Fatal(err)
		}
		if err := repo.CreateDecision(context.Background(), second); !errors.Is(err, dataentity.ErrRequirementAlreadyDecided) {
			t.Fatalf("duplicate decision error=%v", err)
		}
	})
}

func TestConcurrentConfirmVersusRejectOnlyOneWins(t *testing.T) {
	svc, _, _ := newService(successfulModel(labProposal()))
	req := createManual(t, svc, testTenant, "race")
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _, err := svc.ConfirmRequirement(context.Background(), testTenant, req.RequirementID, "a", "")
		errs <- err
	}()
	go func() {
		defer wg.Done()
		_, _, err := svc.RejectRequirement(context.Background(), testTenant, req.RequirementID, "b", "")
		errs <- err
	}()
	wg.Wait()
	close(errs)
	successes := 0
	for err := range errs {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("successful decisions=%d", successes)
	}
	decisions, _ := svc.ListDecisions(context.Background(), testTenant, req.RequirementID)
	if len(decisions) != 1 {
		t.Fatalf("decisions=%+v", decisions)
	}
}

func TestTenantIsolationAndPinnedBusinessRevision(t *testing.T) {
	model := successfulModel(labProposal())
	svc, _, business := newService(model)
	req := createManual(t, svc, testTenant, "isolated")
	if _, err := svc.GetRequirement(context.Background(), "tenant-b", req.RequirementID); !errors.Is(err, dataentity.ErrRequirementNotFound) {
		t.Fatalf("cross-tenant error=%v", err)
	}
	before, err := business.Get(context.Background(), testTenant, labBiz)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AnalyzeDataRequirements(context.Background(), analyzeInput(labBiz, "pinned")); err != nil {
		t.Fatal(err)
	}
	after, err := business.Get(context.Background(), testTenant, labBiz)
	if err != nil {
		t.Fatal(err)
	}
	if before.CurrentRevision != 1 || after.CurrentRevision != before.CurrentRevision {
		t.Fatalf("business revision changed: before=%d after=%d", before.CurrentRevision, after.CurrentRevision)
	}
	business.mu.Lock()
	calls := business.getRevisionCalls
	business.mu.Unlock()
	if calls == 0 {
		t.Fatal("analysis did not load pinned revision")
	}
}

func TestSameServiceAnalyzesMultipleBusinessFixtures(t *testing.T) {
	model := &FakeFormaDataModel{AnalyzeFn: func(_ context.Context, req *AnalyzeDataRequirementsRequest) (*AnalyzeDataRequirementsResponse, error) {
		proposal := labProposal()
		if req.BusinessID == poBiz {
			proposal = poProposal()
		}
		return &AnalyzeDataRequirementsResponse{ModelRef: "fixture", Proposals: []dataentity.DataRequirementProposal{proposal}}, nil
	}}
	svc, _, _ := newService(model)
	for _, businessID := range []string{labBiz, poBiz} {
		result, err := svc.AnalyzeDataRequirements(context.Background(), analyzeInput(businessID, "fixture-"+businessID))
		if err != nil || result.Run.Status != dataentity.AnalysisSucceeded || len(result.Requirements) != 1 {
			t.Fatalf("%s result=%+v error=%v", businessID, result, err)
		}
	}
}
