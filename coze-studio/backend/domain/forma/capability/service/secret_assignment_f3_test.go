/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/fixture"
)

func TestF3AssignmentPatternsRejectNonEmptyRHS(t *testing.T) {
	assignments := []string{
		"token=abc",
		"cookie=sessionid",
		"secret=s3cr3t",
		"session=sid123",
		"client_secret=cs_value",
		"access_token=at_value",
		"refresh_token=rt_value",
		"authorization: Bearer x",
		"authorization=Bearerx",
		"Authorization: tok",
		"Bearer x",
		"Bearer ab",
		"Bearer short1",
	}
	for _, s := range assignments {
		require.True(t, containsSecret(s), "must reject assignment/Bearer form %q", s)
		require.ErrorIs(t, ValidateAuditMetadata(s, ""), entity.ErrInvalidPayload, s)
	}
}

func TestF3BusinessFalsePositivesAllowed(t *testing.T) {
	allowed := []string{
		"review trade secret policy",
		"tokenization completed",
		"cookie policy approved",
		"password reset workflow approved",
		"bearer of responsibility",
	}
	for _, s := range allowed {
		require.False(t, containsSecret(s), "must allow business phrase %q", s)
		require.NoError(t, ValidateAuditMetadata(s, ""), s)
		require.NoError(t, ValidateAuditMetadata(s, "client-req-ok"), s)
	}
}

func TestF3IllegalReasonNoPartialWrite(t *testing.T) {
	svc, repo, _, bmPort, contractPort := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, OwnerID: testOwnerID,
		CapabilityID: "cap-f3-reason", Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	beforeRevs, err := repo.ListRevisions(context.Background(), "t1", "cap-f3-reason")
	require.NoError(t, err)
	beforeDecs, err := repo.ListDecisionsByCapability(context.Background(), "t1", "cap-f3-reason")
	require.NoError(t, err)

	_, _, err = svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: "cap-f3-reason", SourceRevisionID: rev.RevisionID,
		ClientRequestID: "derive-f3-ok", ActorID: testActor, Reason: "token=leaked_value",
		Action: entity.DecisionDerive, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	afterRevs, err := repo.ListRevisions(context.Background(), "t1", "cap-f3-reason")
	require.NoError(t, err)
	afterDecs, err := repo.ListDecisionsByCapability(context.Background(), "t1", "cap-f3-reason")
	require.NoError(t, err)
	require.Equal(t, len(beforeRevs), len(afterRevs))
	require.Equal(t, len(beforeDecs), len(afterDecs))

	seedPASSValidationForTest(t, repo, bmPort, contractPort, rev)
	forceStatusForTest(t, repo, "t1", rev.RevisionID, entity.RevisionDraft, entity.RevisionValidated)
	_, err = svc.Activate(context.Background(), "t1", rev.RevisionID, testActor, "secret=leaked")
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
	got, err := repo.GetRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Equal(t, entity.RevisionValidated, got.Status)
}

func TestF3IllegalClientRequestIDNoPartialWrite(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, _, _, _, _ := newTestService(gen)
	badCRID := "ghp_abcdefghijklmnopqrstuvwxyz12"
	_, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: badCRID,
		ActorID: testActor, Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
	require.Equal(t, 0, gen.CallCount())

	_, err = svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "token=x",
		ActorID: testActor, Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
	require.Equal(t, 0, gen.CallCount())
}
