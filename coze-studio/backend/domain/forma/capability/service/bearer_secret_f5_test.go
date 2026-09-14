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

// F5: only the exact safe phrase "bearer of responsibility" (any case) is exempted.
// Open "bearer of <anything>" is NOT a safe exemption under the new policy.

func TestF5BearerSafePhraseOnlyAllowed(t *testing.T) {
	allowed := []string{
		"bearer of responsibility",
		"BEARER OF RESPONSIBILITY",
	}
	for _, s := range allowed {
		t.Run(s, func(t *testing.T) {
			require.False(t, containsSecret(s), "must allow exact safe phrase %q", s)
			require.NoError(t, ValidateAuditMetadata(s, ""), s)
			require.NoError(t, ValidateAuditMetadata(s, "client-req-ok"), s)
		})
	}
}

func TestF5BearerOpenOfPhraseRejected(t *testing.T) {
	rejected := []string{
		"Bearer of /abc",
		"Bearer of +abc",
		"Bearer of duty",
		"bearer of trust",
		"bearer of responsibility; Bearer of /abc",
		"bearer of responsibility Bearer x",
	}
	for _, s := range rejected {
		t.Run(s, func(t *testing.T) {
			require.True(t, containsSecret(s), "must reject non-safe Bearer phrase %q", s)
			require.ErrorIs(t, ValidateAuditMetadata(s, ""), entity.ErrInvalidPayload, s)
		})
	}
}

func TestF5IllegalBearerReasonNoPartialWrite(t *testing.T) {
	svc, repo, _, bmPort, contractPort := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, OwnerID: testOwnerID,
		CapabilityID: "cap-f5-bearer", Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	beforeRevs, err := repo.ListRevisions(context.Background(), "t1", "cap-f5-bearer")
	require.NoError(t, err)
	beforeDecs, err := repo.ListDecisionsByCapability(context.Background(), "t1", "cap-f5-bearer")
	require.NoError(t, err)

	_, _, err = svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: "cap-f5-bearer", SourceRevisionID: rev.RevisionID,
		ClientRequestID: "derive-f5-ok", ActorID: testActor, Reason: "Bearer of duty",
		Action: entity.DecisionDerive, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	afterRevs, err := repo.ListRevisions(context.Background(), "t1", "cap-f5-bearer")
	require.NoError(t, err)
	afterDecs, err := repo.ListDecisionsByCapability(context.Background(), "t1", "cap-f5-bearer")
	require.NoError(t, err)
	require.Equal(t, len(beforeRevs), len(afterRevs))
	require.Equal(t, len(beforeDecs), len(afterDecs))

	seedPASSValidationForTest(t, repo, bmPort, contractPort, rev)
	forceStatusForTest(t, repo, "t1", rev.RevisionID, entity.RevisionDraft, entity.RevisionValidated)
	_, err = svc.Activate(context.Background(), "t1", rev.RevisionID, testActor, "Bearer of /abc")
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
	got, err := repo.GetRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Equal(t, entity.RevisionValidated, got.Status)
}
