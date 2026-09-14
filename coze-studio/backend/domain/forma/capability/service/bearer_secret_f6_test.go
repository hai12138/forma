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

// F6: safe phrase is an exact whole-string match after TrimSpace (any case).
// Mid-string / prefix / suffix / punctuation / repetition around the phrase MUST reject.
// Unicode whitespace after "Bearer" MUST reject (Go regexp \s is ASCII-only by default).
//
// Surrounding-whitespace ALLOW cases use ASCII spaces that strings.TrimSpace removes and that
// pass ValidateAuditMetadata (tab/LF are TrimSpace-able but rejected as control chars r < 0x20).

func TestF6ExactSafePhraseAllowed(t *testing.T) {
	allowed := []string{
		"bearer of responsibility",
		"BEARER OF RESPONSIBILITY",
		"  bearer of responsibility  ",
		"  BEARER OF RESPONSIBILITY  ",
	}
	for _, s := range allowed {
		t.Run(s, func(t *testing.T) {
			require.False(t, containsSecret(s), "must allow exact safe phrase %q", s)
			require.NoError(t, ValidateAuditMetadata(s, ""), s)
			require.NoError(t, ValidateAuditMetadata(s, "client-req-ok"), s)
		})
	}
}

func TestF6BoundaryAroundSafePhraseRejected(t *testing.T) {
	rejected := []string{
		"bearer of responsibility approved",
		"prefix bearer of responsibility",
		"bearer of responsibility;",
		"bearer of responsibility/abc",
		"bearer of responsibility abc123",
		"bearer of responsibility bearer of responsibility",
	}
	for _, s := range rejected {
		t.Run(s, func(t *testing.T) {
			require.True(t, containsSecret(s), "must reject non-exact safe phrase boundary %q", s)
			require.ErrorIs(t, ValidateAuditMetadata(s, ""), entity.ErrInvalidPayload, s)
		})
	}
}

func TestF6UnicodeWhitespaceBearerRejected(t *testing.T) {
	rejected := []string{
		"Bearer\u00A0abc",
		"Bearer\u2003abc",
		"Bearer\u202Fabc",
		"authorization: Bearer\u00A0abc",
	}
	for _, s := range rejected {
		t.Run(s, func(t *testing.T) {
			require.True(t, containsSecret(s), "must reject Unicode-whitespace Bearer credential %q", s)
			require.ErrorIs(t, ValidateAuditMetadata(s, ""), entity.ErrInvalidPayload, s)
		})
	}
}

func TestF6IllegalBearerReasonNoPartialWrite(t *testing.T) {
	// Pick one REJECT reason from the F6 boundary matrix.
	const rejectReason = "bearer of responsibility approved"

	svc, repo, _, bmPort, contractPort := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, OwnerID: testOwnerID,
		CapabilityID: "cap-f6-bearer", Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	beforeRevs, err := repo.ListRevisions(context.Background(), "t1", "cap-f6-bearer")
	require.NoError(t, err)
	beforeDecs, err := repo.ListDecisionsByCapability(context.Background(), "t1", "cap-f6-bearer")
	require.NoError(t, err)

	_, _, err = svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: "cap-f6-bearer", SourceRevisionID: rev.RevisionID,
		ClientRequestID: "derive-f6-ok", ActorID: testActor, Reason: rejectReason,
		Action: entity.DecisionDerive, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	afterRevs, err := repo.ListRevisions(context.Background(), "t1", "cap-f6-bearer")
	require.NoError(t, err)
	afterDecs, err := repo.ListDecisionsByCapability(context.Background(), "t1", "cap-f6-bearer")
	require.NoError(t, err)
	require.Equal(t, len(beforeRevs), len(afterRevs))
	require.Equal(t, len(beforeDecs), len(afterDecs))

	seedPASSValidationForTest(t, repo, bmPort, contractPort, rev)
	forceStatusForTest(t, repo, "t1", rev.RevisionID, entity.RevisionDraft, entity.RevisionValidated)
	_, err = svc.Activate(context.Background(), "t1", rev.RevisionID, testActor, rejectReason)
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
	got, err := repo.GetRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Equal(t, entity.RevisionValidated, got.Status)
}
