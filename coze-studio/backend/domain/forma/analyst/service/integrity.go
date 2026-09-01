/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"strings"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/repository"
	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
)

func canonicalConflictPair(a, b string) (string, string) {
	if a < b {
		return a, b
	}
	return b, a
}

func isDuplicateKeyErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "1062")
}

func (s *analystServiceImpl) upsertConflict(
	ctx context.Context,
	repo repository.AnalystRepository,
	tenantID, businessID, sessionID string,
	aID, bID, subject, predicate string,
	now time.Time,
) (*entity.AssertionConflict, error) {
	idA, idB := canonicalConflictPair(aID, bID)
	existing, err := repo.GetConflictByPair(ctx, tenantID, businessID, idA, idB)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	c := &entity.AssertionConflict{
		ConflictID:   newID("conf"),
		TenantID:     tenantID,
		BusinessID:   businessID,
		SessionID:    sessionID,
		AssertionIDA: idA,
		AssertionIDB: idB,
		SubjectRef:   subject,
		Predicate:    predicate,
		Status:       entity.ConflictOpen,
		CreatedAt:    now,
	}
	if err := repo.CreateConflict(ctx, c); err != nil {
		if isDuplicateKeyErr(err) {
			return repo.GetConflictByPair(ctx, tenantID, businessID, idA, idB)
		}
		return nil, err
	}
	return c, nil
}

func (s *analystServiceImpl) resolveConflictsForAssertions(
	ctx context.Context,
	repo repository.AnalystRepository,
	tenantID, businessID string,
	now time.Time,
) error {
	conflicts, err := repo.ListConflicts(ctx, tenantID, businessID)
	if err != nil {
		return err
	}
	for _, c := range conflicts {
		if c == nil || c.Status != entity.ConflictOpen {
			continue
		}
		a, err := repo.GetAssertion(ctx, tenantID, c.AssertionIDA)
		if err != nil {
			return err
		}
		b, err := repo.GetAssertion(ctx, tenantID, c.AssertionIDB)
		if err != nil {
			return err
		}
		if a == nil || b == nil {
			continue
		}
		resolved := false
		if a.Status == entity.AssertionConfirmed && b.Status == entity.AssertionRejected {
			resolved = true
		}
		if b.Status == entity.AssertionConfirmed && a.Status == entity.AssertionRejected {
			resolved = true
		}
		if a.Status == entity.AssertionRejected && b.Status == entity.AssertionRejected {
			resolved = true
		}
		if a.Status == entity.AssertionSuperseded || b.Status == entity.AssertionSuperseded {
			if a.Status != entity.AssertionProposed && b.Status != entity.AssertionProposed {
				resolved = true
			}
		}
		if resolved {
			if err := repo.UpdateConflictStatus(ctx, tenantID, c.ConflictID, entity.ConflictResolved, now); err != nil {
				return err
			}
		}
	}
	return nil
}

func filterAssertionsBySession(assertions []*entity.BusinessAssertion, sessionID string, confirmedOnly bool) []*entity.BusinessAssertion {
	var out []*entity.BusinessAssertion
	for _, a := range assertions {
		if a == nil || a.SessionID != sessionID {
			continue
		}
		if confirmedOnly && a.Status != entity.AssertionConfirmed {
			continue
		}
		out = append(out, a)
	}
	return out
}

func copyEvidenceRefs(ctx context.Context, repo repository.AnalystRepository, tenantID, fromAssertionID, toAssertionID string, at time.Time) error {
	ids, err := repo.ListEvidenceIDsForAssertion(ctx, tenantID, fromAssertionID)
	if err != nil {
		return err
	}
	for _, eid := range ids {
		if err := repo.CreateAssertionEvidenceRef(ctx, tenantID, toAssertionID, eid, at); err != nil && !isDuplicateKeyErr(err) {
			return err
		}
	}
	return nil
}

// ProposalPreviewResult bundles proposal preview artifacts.
type ProposalPreviewResult struct {
	Proposal        *entity.BusinessModelProposal
	ProposedModel   *businessentity.SemanticModel
	CurrentRevision int32
	ValidationValid bool
	ValidationError string
	Diff            *businessentity.BusinessModelDiff
	Impact          *businessentity.BusinessImpactSummary
}
