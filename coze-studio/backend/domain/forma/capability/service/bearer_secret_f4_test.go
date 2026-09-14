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

func TestF4BearerPunctuationAndBareOfRejected(t *testing.T) {
	rejected := []string{
		"Bearer /abc",
		"Bearer +abc",
		"Bearer ~abc",
		`Bearer "abc"`,
		"Bearer of",
		"Bearer !secret",
		"Bearer @token",
		"Bearer #hash",
		"Bearer $val",
		"Bearer %pct",
		"Bearer &amp",
		"Bearer *star",
		"Bearer (paren",
		"Bearer [bracket",
		"Bearer {brace",
		"Bearer |pipe",
		"Bearer :colon",
		"Bearer ;semi",
		"Bearer <lt",
		"Bearer >gt",
		"Bearer ,comma",
		"Bearer ?q",
		"Bearer abc/def",
		"authorization: Bearer /abc",
	}
	for _, s := range rejected {
		require.True(t, containsSecret(s), "must reject Bearer credential %q", s)
		require.ErrorIs(t, ValidateAuditMetadata(s, ""), entity.ErrInvalidPayload, s)
	}
}

func TestF4BearerBusinessPhraseAllowed(t *testing.T) {
	// F5 tightened Bearer safe-phrase allowlist to exact "bearer of responsibility" only.
	allowed := []string{
		"bearer of responsibility",
		"BEARER OF RESPONSIBILITY",
		"review trade secret policy",
		"tokenization completed",
		"cookie policy approved",
		"password reset workflow approved",
	}
	for _, s := range allowed {
		require.False(t, containsSecret(s), "must allow business phrase %q", s)
		require.NoError(t, ValidateAuditMetadata(s, ""), s)
		require.NoError(t, ValidateAuditMetadata(s, "client-req-ok"), s)
	}
}

func TestF4BearerMixedPhraseAndCredentialRejected(t *testing.T) {
	mixed := []string{
		"bearer of responsibility Bearer /abc",
		"Bearer of duty Bearer +abc",
		"bearer of trust Bearer ~xyz",
		`bearer of responsibility Bearer "abc"`,
		"Bearer of role Bearer of",
		"ok: bearer of responsibility; bad: Bearer short1",
	}
	for _, s := range mixed {
		require.True(t, containsSecret(s), "must reject mixed phrase+credential %q", s)
		require.ErrorIs(t, ValidateAuditMetadata(s, ""), entity.ErrInvalidPayload, s)
	}
}

func TestF4AssignmentPatternsStillRejected(t *testing.T) {
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
		require.True(t, containsSecret(s), "must not weaken F3 assignment/Bearer rejection for %q", s)
		require.ErrorIs(t, ValidateAuditMetadata(s, ""), entity.ErrInvalidPayload, s)
	}
}

func TestF4IllegalBearerReasonNoPartialWrite(t *testing.T) {
	svc, repo, _, bmPort, contractPort := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, OwnerID: testOwnerID,
		CapabilityID: "cap-f4-bearer", Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	beforeRevs, err := repo.ListRevisions(context.Background(), "t1", "cap-f4-bearer")
	require.NoError(t, err)
	beforeDecs, err := repo.ListDecisionsByCapability(context.Background(), "t1", "cap-f4-bearer")
	require.NoError(t, err)

	_, _, err = svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: "cap-f4-bearer", SourceRevisionID: rev.RevisionID,
		ClientRequestID: "derive-f4-ok", ActorID: testActor, Reason: "Bearer /abc",
		Action: entity.DecisionDerive, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	afterRevs, err := repo.ListRevisions(context.Background(), "t1", "cap-f4-bearer")
	require.NoError(t, err)
	afterDecs, err := repo.ListDecisionsByCapability(context.Background(), "t1", "cap-f4-bearer")
	require.NoError(t, err)
	require.Equal(t, len(beforeRevs), len(afterRevs))
	require.Equal(t, len(beforeDecs), len(afterDecs))

	seedPASSValidationForTest(t, repo, bmPort, contractPort, rev)
	forceStatusForTest(t, repo, "t1", rev.RevisionID, entity.RevisionDraft, entity.RevisionValidated)
	_, err = svc.Activate(context.Background(), "t1", rev.RevisionID, testActor, "Bearer of")
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
	got, err := repo.GetRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Equal(t, entity.RevisionValidated, got.Status)
}

func TestF4IllegalBearerConfirmNoPartialWrite(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, repo, assets, _, _ := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "confirm-f4-start",
		ActorID: testActor, Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	propID := res.Proposals[0].ProposalID

	_, err = svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: propID, ActorID: testActor, OwnerID: testOwnerID,
		CapabilityID: "cap-f4-confirm", Reason: "Bearer ~abc", ClientRequestID: "confirm-f4-ok",
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	_, err = repo.GetCapability(context.Background(), "t1", "cap-f4-confirm")
	require.ErrorIs(t, err, entity.ErrNotFound)
	_, err = assets.GetCapabilityAsset(context.Background(), "t1", "cap-f4-confirm")
	require.ErrorIs(t, err, entity.ErrNotFound)
	prop, err := repo.GetProposal(context.Background(), "t1", propID)
	require.NoError(t, err)
	require.Equal(t, entity.ProposalProposed, prop.Status)
}
