/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/repository"
	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
)

func (s *analystServiceImpl) persistExtraction(
	ctx context.Context,
	tenantID, businessID, sessionID, actorID string,
	evidence *entity.BusinessEvidence,
	extraction *entity.ExtractionResult,
	now time.Time,
) ([]*entity.BusinessAssertion, []*entity.AssertionConflict, []*entity.AnalystGap, error) {
	if extraction == nil {
		return nil, nil, nil, nil
	}
	var assertions []*entity.BusinessAssertion
	var outConflicts []*entity.AssertionConflict
	var gaps []*entity.AnalystGap
	err := s.repo.Transaction(ctx, func(txRepo repository.AnalystRepository) error {
		a, c, g, err := s.persistExtractionWithRepo(ctx, txRepo, tenantID, businessID, sessionID, actorID, evidence, extraction, now)
		if err != nil {
			return err
		}
		assertions = a
		outConflicts = c
		gaps = g
		return nil
	})
	return assertions, outConflicts, gaps, err
}

func (s *analystServiceImpl) persistExtractionWithRepo(
	ctx context.Context,
	repo repository.AnalystRepository,
	tenantID, businessID, sessionID, actorID string,
	evidence *entity.BusinessEvidence,
	extraction *entity.ExtractionResult,
	now time.Time,
) ([]*entity.BusinessAssertion, []*entity.AssertionConflict, []*entity.AnalystGap, error) {
	if extraction == nil {
		return nil, nil, nil, nil
	}
	var assertions []*entity.BusinessAssertion
	turnEvidenceMap := map[string]string{}
	if evidence != nil {
		turnEvidenceMap[evidence.TurnID] = evidence.EvidenceID
	}

	for _, ea := range extraction.Assertions {
		a := &entity.BusinessAssertion{
			AssertionID:     newID("assert"),
			TenantID:        tenantID,
			BusinessID:      businessID,
			SessionID:       sessionID,
			AssertionType:   ea.AssertionType,
			SubjectRef:      ea.SubjectRef,
			Predicate:       ea.Predicate,
			ObjectValue:     ea.ObjectValue,
			StructuredValue: ea.StructuredValue,
			Confidence:      ea.Confidence,
			Status:          entity.AssertionProposed,
			SourceMarker:    businessentity.SourceAIGenerated,
			CreatedBy:       actorID,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := repo.CreateAssertion(ctx, a); err != nil {
			return nil, nil, nil, err
		}
		linked := false
		for _, tid := range ea.EvidenceTurnIDs {
			eid := turnEvidenceMap[tid]
			if eid == "" && evidence != nil && evidence.TurnID == tid {
				eid = evidence.EvidenceID
			}
			if eid != "" {
				if err := repo.CreateAssertionEvidenceRef(ctx, tenantID, a.AssertionID, eid, now); err != nil {
					return nil, nil, nil, err
				}
				linked = true
			}
		}
		if !linked && evidence != nil {
			if err := repo.CreateAssertionEvidenceRef(ctx, tenantID, a.AssertionID, evidence.EvidenceID, now); err != nil {
				return nil, nil, nil, err
			}
		}
		a.EvidenceIDs, _ = repo.ListEvidenceIDsForAssertion(ctx, tenantID, a.AssertionID)
		assertions = append(assertions, a)
		_ = s.audit.RecordAnalystAudit(ctx, tenantID, actorID, "ASSERTION_CREATED", a.AssertionID, "")
	}

	for _, link := range extraction.EvidenceLinks {
		if link.AssertionIndex < 0 || link.AssertionIndex >= len(assertions) {
			continue
		}
		a := assertions[link.AssertionIndex]
		eid := turnEvidenceMap[link.EvidenceTurnID]
		if eid != "" {
			_ = repo.CreateAssertionEvidenceRef(ctx, tenantID, a.AssertionID, eid, now)
		}
	}

	var gaps []*entity.AnalystGap
	for _, g := range extraction.Gaps {
		gap := &entity.AnalystGap{
			GapID:      newID("gap"),
			TenantID:   tenantID,
			BusinessID: businessID,
			SessionID:  sessionID,
			GapType:    g.GapType,
			Question:   g.Question,
			Status:     entity.GapOpen,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := repo.CreateGap(ctx, gap); err != nil {
			return nil, nil, nil, err
		}
		gaps = append(gaps, gap)
	}

	allProposed := append([]*entity.BusinessAssertion{}, assertions...)
	existing, _ := loadAssertionsWithEvidenceFromRepo(ctx, repo, tenantID, businessID)
	for _, e := range existing {
		if e.Status == entity.AssertionProposed {
			allProposed = append(allProposed, e)
		}
	}
	var outConflicts []*entity.AssertionConflict
	detected := detectConflicts(allProposed, tenantID, businessID, sessionID, now)
	for _, c := range detected {
		uc, err := s.upsertConflict(ctx, repo, tenantID, businessID, sessionID, c.AssertionIDA, c.AssertionIDB, c.SubjectRef, c.Predicate, now)
		if err != nil {
			return nil, nil, nil, err
		}
		outConflicts = append(outConflicts, uc)
	}

	for _, mc := range extraction.Conflicts {
		if mc.AssertionIndexA < len(assertions) && mc.AssertionIndexB < len(assertions) {
			uc, err := s.upsertConflict(ctx, repo, tenantID, businessID, sessionID,
				assertions[mc.AssertionIndexA].AssertionID,
				assertions[mc.AssertionIndexB].AssertionID,
				assertions[mc.AssertionIndexA].SubjectRef,
				assertions[mc.AssertionIndexA].Predicate,
				now,
			)
			if err != nil {
				return nil, nil, nil, err
			}
			outConflicts = append(outConflicts, uc)
		}
	}

	return assertions, outConflicts, gaps, nil
}

func loadAssertionsWithEvidenceFromRepo(
	ctx context.Context,
	repo repository.AnalystRepository,
	tenantID, businessID string,
) ([]*entity.BusinessAssertion, error) {
	list, err := repo.ListAssertions(ctx, tenantID, businessID)
	if err != nil {
		return nil, err
	}
	for _, a := range list {
		ids, _ := repo.ListEvidenceIDsForAssertion(ctx, tenantID, a.AssertionID)
		a.EvidenceIDs = ids
	}
	return list, nil
}
