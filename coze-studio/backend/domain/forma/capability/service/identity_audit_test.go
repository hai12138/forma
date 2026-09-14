/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/fixture"
)

func TestPrincipalActorIDAndSeparateOwnerID(t *testing.T) {
	svc, repo, assets, _, _ := newTestService(nil)
	principal := "user_opaque_principal_xyz"
	owner := int64(4242)
	cap, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: principal, OwnerID: owner,
		CapabilityID: "cap-principal", Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	require.Equal(t, principal, cap.CreatedBy)
	require.Equal(t, principal, rev.CreatedBy)

	asset, err := assets.GetCapabilityAsset(context.Background(), "t1", "cap-principal")
	require.NoError(t, err)
	require.Equal(t, owner, asset.OwnerID)
	require.Equal(t, owner, asset.CreatedBy)

	decs, err := repo.ListDecisionsByCapability(context.Background(), "t1", "cap-principal")
	require.NoError(t, err)
	require.NotEmpty(t, decs)
	require.Equal(t, principal, decs[0].ActorPrincipalID)
}

func TestManualCreateMissingOwnerIDNoPartialWrite(t *testing.T) {
	svc, repo, assets, _, _ := newTestService(nil)
	_, _, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: "principal-alice", OwnerID: 0,
		CapabilityID: "cap-missing-owner", Payload: fixture.LaboratoryFlowCapability(),
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	_, err = repo.GetCapability(context.Background(), "t1", "cap-missing-owner")
	require.ErrorIs(t, err, entity.ErrNotFound)
	_, err = assets.GetCapabilityAsset(context.Background(), "t1", "cap-missing-owner")
	require.ErrorIs(t, err, entity.ErrNotFound)
	listed, err := svc.ListCapabilities(context.Background(), "t1", "biz-lab")
	require.NoError(t, err)
	require.Empty(t, listed)
}

func TestConfirmFirstCreateMissingOwnerIDNoPartialWrite(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, repo, assets, _, _ := newTestService(gen)
	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "confirm-no-owner", ActorID: testActor,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	propID := res.Proposals[0].ProposalID

	_, err = svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: propID, ActorID: testActor, OwnerID: 0, CapabilityID: "cap-no-owner",
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	_, err = repo.GetCapability(context.Background(), "t1", "cap-no-owner")
	require.ErrorIs(t, err, entity.ErrNotFound)
	_, err = assets.GetCapabilityAsset(context.Background(), "t1", "cap-no-owner")
	require.ErrorIs(t, err, entity.ErrNotFound)
	prop, err := repo.GetProposal(context.Background(), "t1", propID)
	require.NoError(t, err)
	require.Equal(t, entity.ProposalProposed, prop.Status)
}

func TestConfirmFirstCreateIdentityPrincipalAndOwner(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, repo, assets, _, _ := newTestService(gen)
	principal := "principal-confirm-alice"
	owner := int64(7777) // CozeUserID

	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "confirm-id-ok", ActorID: principal,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	propID := res.Proposals[0].ProposalID

	rev, err := svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: propID, ActorID: principal, OwnerID: owner,
		CapabilityID: "cap-confirm-id", Reason: "review trade secret policy",
	})
	require.NoError(t, err)
	require.Equal(t, principal, rev.CreatedBy)

	cap, err := repo.GetCapability(context.Background(), "t1", "cap-confirm-id")
	require.NoError(t, err)
	require.Equal(t, principal, cap.CreatedBy)

	asset, err := assets.GetCapabilityAsset(context.Background(), "t1", "cap-confirm-id")
	require.NoError(t, err)
	require.Equal(t, owner, asset.OwnerID)
	require.Equal(t, owner, asset.CreatedBy)

	decs, err := repo.ListDecisionsByCapability(context.Background(), "t1", "cap-confirm-id")
	require.NoError(t, err)
	require.NotEmpty(t, decs)
	require.Equal(t, principal, decs[0].ActorPrincipalID)
	require.Equal(t, entity.DecisionConfirm, decs[0].Action)
}

func TestEditConfirmFirstCreateIdentityPrincipalAndOwner(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc, repo, assets, _, _ := newTestService(gen)
	principal := "principal-editconfirm-bob"
	owner := int64(8888) // CozeUserID

	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: "editconfirm-id-ok", ActorID: principal,
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	propID := res.Proposals[0].ProposalID

	edited := fixture.LaboratoryFlowCapability()
	edited.Name = "EditedLaboratoryFlow"
	rev, err := svc.EditConfirmProposal(context.Background(), &EditConfirmInput{
		TenantID: "t1", ProposalID: propID, ActorID: principal, OwnerID: owner,
		CapabilityID: "cap-editconfirm-id", Reason: "password reset workflow approved",
		EffectivePayload: edited,
	})
	require.NoError(t, err)
	require.Equal(t, principal, rev.CreatedBy)
	require.Equal(t, "EditedLaboratoryFlow", rev.Name)

	cap, err := repo.GetCapability(context.Background(), "t1", "cap-editconfirm-id")
	require.NoError(t, err)
	require.Equal(t, principal, cap.CreatedBy)

	asset, err := assets.GetCapabilityAsset(context.Background(), "t1", "cap-editconfirm-id")
	require.NoError(t, err)
	require.Equal(t, owner, asset.OwnerID)
	require.Equal(t, owner, asset.CreatedBy)

	decs, err := repo.ListDecisionsByCapability(context.Background(), "t1", "cap-editconfirm-id")
	require.NoError(t, err)
	require.NotEmpty(t, decs)
	require.Equal(t, principal, decs[0].ActorPrincipalID)
	require.Equal(t, entity.DecisionEditConfirm, decs[0].Action)
}

func TestValidateAuditMetadataAllowsBusinessFalsePositives(t *testing.T) {
	require.NoError(t, ValidateAuditMetadata("", ""))
	require.NoError(t, ValidateAuditMetadata("approved by owner", "req-123"))

	for _, reason := range []string{
		"review trade secret policy",
		"tokenization completed",
		"cookie policy approved",
		"password reset workflow approved",
	} {
		require.NoError(t, ValidateAuditMetadata(reason, ""), reason)
		require.NoError(t, ValidateAuditMetadata(reason, "client-req-ok"), reason)
	}
}

func TestValidateAuditMetadataRejectsCredentialShapes(t *testing.T) {
	require.ErrorIs(t, ValidateAuditMetadata("has\nnewline", ""), entity.ErrInvalidPayload)
	require.ErrorIs(t, ValidateAuditMetadata("has\ttab", ""), entity.ErrInvalidPayload)
	require.ErrorIs(t, ValidateAuditMetadata(string([]byte{0xff, 0xfe, 0xfd}), ""), entity.ErrInvalidPayload)

	longReason := strings.Repeat("a", maxAuditReasonLen+1)
	require.ErrorIs(t, ValidateAuditMetadata(longReason, ""), entity.ErrInvalidPayload)
	longCRID := strings.Repeat("b", maxAuditClientRequestIDLen+1)
	require.ErrorIs(t, ValidateAuditMetadata("", longCRID), entity.ErrInvalidPayload)

	// Free-text reason: assignment forms + credential shapes (not bare keywords).
	reasonSecrets := []string{
		"Bearer abcdefghijklmnop",
		"eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.sig",
		"token ghp_abcdefghijklmnopqrstuvwxyz12",
		"pat github_pat_abcdefghijklmnopqrstuvwxyz",
		"key sk-abcdefghijklmnopqrstuvwxyz12",
		"key sk-proj-abcdefghijklmnopQR",
		"slack xoxb-1234567890-abcdefghij",
		"-----BEGIN RSA PRIVATE KEY-----",
		"password=hunter2",
		"api_key=abc123",
		"Authorization: Bearer abcdefghijklmnop",
		strings.Repeat("A", 48),
		strings.Repeat("A", 64) + "==",
	}
	for _, secret := range reasonSecrets {
		require.ErrorIs(t, ValidateAuditMetadata(secret, ""), entity.ErrInvalidPayload, secret)
	}

	// client_request_id: opaque + credential-shape (no spaces / assignment forms that fail OpaqueID).
	opaqueSecrets := []string{
		"eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.sig",
		"ghp_abcdefghijklmnopqrstuvwxyz12",
		"github_pat_abcdefghijklmnopqrstuvwxyz",
		"sk-abcdefghijklmnopqrstuvwxyz12",
		"sk-proj-abcdefghijklmnopQR",
		"xoxb-1234567890-abcdefghij",
		strings.Repeat("A", 48),
	}
	for _, secret := range opaqueSecrets {
		require.NoError(t, ValidateOpaqueID(secret), "sample %q must pass ValidateOpaqueID", secret)
		require.True(t, containsCredentialShape(secret), "sample %q must hit credential-shape", secret)
		require.ErrorIs(t, ValidateAuditMetadata("", secret), entity.ErrInvalidPayload, secret)
	}

	require.ErrorIs(t, ValidateAuditMetadata("", "has space"), entity.ErrInvalidPayload)
	require.ErrorIs(t, ValidateAuditMetadata("", "password=x"), entity.ErrInvalidPayload)

	require.True(t, utf8.ValidString("ok"))
}

func TestAuditMetadataSecretRejectNoPartialWrites(t *testing.T) {
	svc, repo, _, bmPort, contractPort := newTestService(nil)
	_, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor, OwnerID: testOwnerID,
		CapabilityID: "cap-audit", Payload: fixture.LaboratoryFlowCapability(),
	})
	require.NoError(t, err)
	beforeRevs, err := repo.ListRevisions(context.Background(), "t1", "cap-audit")
	require.NoError(t, err)
	beforeDecs, err := repo.ListDecisionsByCapability(context.Background(), "t1", "cap-audit")
	require.NoError(t, err)

	_, _, err = svc.DeriveRevision(context.Background(), &DeriveInput{
		TenantID: "t1", CapabilityID: "cap-audit", SourceRevisionID: rev.RevisionID,
		ClientRequestID: "derive-ok", ActorID: testActor, Reason: "password=leaked_value",
		Action: entity.DecisionDerive, Payload: fixture.LaboratoryFlowCapability(),
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)

	afterRevs, err := repo.ListRevisions(context.Background(), "t1", "cap-audit")
	require.NoError(t, err)
	afterDecs, err := repo.ListDecisionsByCapability(context.Background(), "t1", "cap-audit")
	require.NoError(t, err)
	require.Equal(t, len(beforeRevs), len(afterRevs))
	require.Equal(t, len(beforeDecs), len(afterDecs))

	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryFlowCapability()}}
	svc2, _, _, _, _ := newTestService(gen)
	badCRID := "ghp_abcdefghijklmnopqrstuvwxyz12"
	_, err = svc2.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz", BusinessModelRevision: 1, ClientRequestID: badCRID,
		ActorID: testActor, Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
	require.Equal(t, 0, gen.CallCount())

	seedPASSValidationForTest(t, repo, bmPort, contractPort, rev)
	forceStatusForTest(t, repo, "t1", rev.RevisionID, entity.RevisionDraft, entity.RevisionValidated)
	_, err = svc.Activate(context.Background(), "t1", rev.RevisionID, testActor, "Authorization: Bearer abcdefghijklmnop")
	require.ErrorIs(t, err, entity.ErrInvalidPayload)
	got, err := repo.GetRevision(context.Background(), "t1", rev.RevisionID)
	require.NoError(t, err)
	require.Equal(t, entity.RevisionValidated, got.Status)
}
