/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
	businesssvc "github.com/coze-dev/coze-studio/backend/domain/forma/business/service"
	"github.com/stretchr/testify/require"
)

type recordingAnalystModel struct {
	DeterministicFakeModel
	mu             sync.Mutex
	extractCalls   []*ExtractionRequest
	interviewCalls []*InterviewTurnRequest
}

func (m *recordingAnalystModel) ExtractAssertions(ctx context.Context, req *ExtractionRequest) (*ExtractionOutcome, error) {
	m.mu.Lock()
	cp := *req
	m.extractCalls = append(m.extractCalls, &cp)
	m.mu.Unlock()
	return m.DeterministicFakeModel.ExtractAssertions(ctx, req)
}

func (m *recordingAnalystModel) GenerateInterviewTurn(ctx context.Context, req *InterviewTurnRequest) (*InterviewTurnResponse, error) {
	m.mu.Lock()
	cp := *req
	m.interviewCalls = append(m.interviewCalls, &cp)
	m.mu.Unlock()
	return m.DeterministicFakeModel.GenerateInterviewTurn(ctx, req)
}

func (m *recordingAnalystModel) lastExtract() *ExtractionRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.extractCalls) == 0 {
		return nil
	}
	return m.extractCalls[len(m.extractCalls)-1]
}

func TestMultiTurnContextTextReachesModel(t *testing.T) {
	ar := newMemAnalystRepo()
	br := newBusinessTestMem()
	model := &recordingAnalystModel{}
	svc := NewAnalystService(&Components{
		Repo:         ar,
		BusinessSVC:  businesssvc.NewBusinessService(&businesssvc.Components{Repo: br}),
		BusinessRepo: br,
		Model:        model,
	})
	ctx := context.Background()
	seedBusiness(ctx, br, "t1", "b1")
	sess, _ := svc.CreateSession(ctx, "t1", "b1", "test", "p1", entity.PolicyDevelopment)

	_, err := svc.SubmitTurn(ctx, "t1", "b1", sess.SessionID,
		"维修人员处理完成后，工单还需要有人关闭。", "cr_ctx_1", "p1")
	require.NoError(t, err)

	_, err = svc.SubmitTurn(ctx, "t1", "b1", sess.SessionID, "管理员。", "cr_ctx_2", "p1")
	require.NoError(t, err)

	req := model.lastExtract()
	require.NotNil(t, req)
	require.NotEmpty(t, req.ContextText)
	require.Contains(t, req.ContextText, "TURN[")
	require.Contains(t, req.ContextText, "关闭")
}

func TestGapAskCreatesAnalystTurnWithoutEvidence(t *testing.T) {
	ar := newMemAnalystRepo()
	br := newBusinessTestMem()
	svc := newTestAnalystSvc(ar, br)
	ctx := context.Background()
	seedBusiness(ctx, br, "t1", "b1")
	sess, _ := svc.CreateSession(ctx, "t1", "b1", "test", "p1", entity.PolicyDevelopment)

	now := time.Now().UTC()
	gap := &entity.AnalystGap{
		GapID:      "gap_close",
		TenantID:   "t1",
		BusinessID: "b1",
		SessionID:  sess.SessionID,
		GapType:    "INFORMATION",
		Question:   "谁可以关闭工单？",
		Status:     entity.GapOpen,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	require.NoError(t, ar.CreateGap(ctx, gap))

	evBefore := len(ar.evidence)
	res, err := svc.AskGap(ctx, "t1", "b1", sess.SessionID, "gap_close", "p1")
	require.NoError(t, err)
	require.NotNil(t, res.AnalystTurn)
	require.Equal(t, entity.SpeakerAnalyst, res.AnalystTurn.Speaker)
	require.Equal(t, "谁可以关闭工单？", res.AnalystTurn.Content)
	require.Equal(t, len(ar.evidence), evBefore)

	updated, _ := ar.GetSession(ctx, "t1", sess.SessionID)
	require.Equal(t, "gap_close", updated.FocusGapID)
}

func TestGapResolvedOnConfirmWhenFocused(t *testing.T) {
	ar := newMemAnalystRepo()
	br := newBusinessTestMem()
	svc := newTestAnalystSvc(ar, br)
	ctx := context.Background()
	seedBusiness(ctx, br, "t1", "b1")
	sess, _ := svc.CreateSession(ctx, "t1", "b1", "test", "p1", entity.PolicyDevelopment)

	now := time.Now().UTC()
	require.NoError(t, ar.CreateGap(ctx, &entity.AnalystGap{
		GapID: "gap_close", TenantID: "t1", BusinessID: "b1", SessionID: sess.SessionID,
		GapType: "INFORMATION", Question: "谁可以关闭工单？", Status: entity.GapOpen,
		CreatedAt: now, UpdatedAt: now,
	}))
	_, _ = svc.AskGap(ctx, "t1", "b1", sess.SessionID, "gap_close", "p1")

	submit, err := svc.SubmitTurn(ctx, "t1", "b1", sess.SessionID, "管理员可以关闭。", "cr_gap_answer", "p1")
	require.NoError(t, err)
	require.NotEmpty(t, submit.Assertions)

	target := submit.Assertions[0]
	for _, a := range submit.Assertions {
		if a.AssertionType == entity.AssertionActorExists || a.AssertionType == entity.AssertionBusinessRule {
			target = a
			break
		}
	}
	_, err = svc.ConfirmAssertion(ctx, "t1", "b1", target.AssertionID, "p1", "confirmed", nil)
	require.NoError(t, err)

	gaps, _ := ar.ListGaps(ctx, "t1", "b1")
	for _, g := range gaps {
		if g.GapID == "gap_close" {
			require.Equal(t, entity.GapResolved, g.Status)
		}
	}
	updated, _ := ar.GetSession(ctx, "t1", sess.SessionID)
	require.Empty(t, updated.FocusGapID)
}

func TestGapAskIdempotentClientRequest(t *testing.T) {
	ar := newMemAnalystRepo()
	br := newBusinessTestMem()
	svc := newTestAnalystSvc(ar, br)
	ctx := context.Background()
	seedBusiness(ctx, br, "t1", "b1")
	sess, _ := svc.CreateSession(ctx, "t1", "b1", "test", "p1", entity.PolicyDevelopment)
	now := time.Now().UTC()
	require.NoError(t, ar.CreateGap(ctx, &entity.AnalystGap{
		GapID: "gap1", TenantID: "t1", BusinessID: "b1", SessionID: sess.SessionID,
		Question: "Q?", Status: entity.GapOpen, CreatedAt: now, UpdatedAt: now,
	}))

	r1, err := svc.AskGap(ctx, "t1", "b1", sess.SessionID, "gap1", "p1")
	require.NoError(t, err)
	r2, err := svc.AskGap(ctx, "t1", "b1", sess.SessionID, "gap1", "p1")
	require.NoError(t, err)
	require.Equal(t, r1.AnalystTurn.TurnID, r2.AnalystTurn.TurnID)
}
