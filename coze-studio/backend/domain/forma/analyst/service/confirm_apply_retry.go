/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
	analystrepo "github.com/coze-dev/coze-studio/backend/domain/forma/analyst/repository"
	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	businessrepo "github.com/coze-dev/coze-studio/backend/domain/forma/business/repository"
	businesssvc "github.com/coze-dev/coze-studio/backend/domain/forma/business/service"
)

func (s *analystServiceImpl) ConfirmAssertion(ctx context.Context, tenantID, businessID, assertionID, actorID, comment string, edit *AssertionEdit) (*entity.BusinessAssertion, error) {
	if edit != nil {
		if err := ValidateAssertionEdit(edit); err != nil {
			return nil, err
		}
	}
	var result *entity.BusinessAssertion
	err := s.repo.Transaction(ctx, func(txRepo analystrepo.AnalystRepository) error {
		a, err := txRepo.GetAssertion(ctx, tenantID, assertionID)
		if err != nil {
			return err
		}
		if a == nil || a.BusinessID != businessID {
			return entity.ErrAssertionNotFound
		}
		if a.Status == entity.AssertionConfirmed || a.Status == entity.AssertionRejected {
			return entity.ErrAssertionAlreadyDecided
		}
		evIDs, err := txRepo.ListEvidenceIDsForAssertion(ctx, tenantID, assertionID)
		if err != nil {
			return err
		}
		if len(evIDs) == 0 {
			return entity.ErrAssertionEvidenceRequired
		}
		now := time.Now().UTC()
		targetID := assertionID
		target := a

		if edit != nil {
			a.Status = entity.AssertionSuperseded
			a.UpdatedAt = now
			if err := txRepo.UpdateAssertion(ctx, a); err != nil {
				return err
			}
			manual := &entity.BusinessAssertion{
				AssertionID:            newID("assert"),
				TenantID:               tenantID,
				BusinessID:             businessID,
				SessionID:              a.SessionID,
				AssertionType:          edit.AssertionType,
				SubjectRef:             edit.SubjectRef,
				Predicate:              edit.Predicate,
				ObjectValue:            edit.ObjectValue,
				StructuredValue:        a.StructuredValue,
				Confidence:             a.Confidence,
				Status:                 entity.AssertionConfirmed,
				SourceMarker:           businessentity.SourceManualModified,
				DerivedFromAssertionID: a.AssertionID,
				CreatedBy:              actorID,
				CreatedAt:              now,
				UpdatedAt:              now,
			}
			if err := txRepo.CreateAssertion(ctx, manual); err != nil {
				return err
			}
			if err := copyEvidenceRefs(ctx, txRepo, tenantID, assertionID, manual.AssertionID, now); err != nil {
				return err
			}
			targetID = manual.AssertionID
			target = manual
			target.EvidenceIDs, _ = txRepo.ListEvidenceIDsForAssertion(ctx, tenantID, targetID)
		} else {
			target.Status = entity.AssertionConfirmed
			target.UpdatedAt = now
			if err := txRepo.UpdateAssertion(ctx, target); err != nil {
				return err
			}
			target.EvidenceIDs = evIDs
		}

		conf := &entity.BusinessConfirmation{
			ConfirmationID: newID("confm"),
			TenantID:       tenantID,
			BusinessID:     businessID,
			AssertionID:    targetID,
			Decision:       entity.DecisionConfirm,
			Comment:        comment,
			DecidedBy:      actorID,
			DecidedAt:      now,
		}
		if err := txRepo.CreateConfirmation(ctx, conf); err != nil {
			return err
		}
		if err := s.resolveConflictsForAssertions(ctx, txRepo, tenantID, businessID, now); err != nil {
			return err
		}
		result = target
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.tryResolveFocusedGap(ctx, tenantID, businessID, result.SessionID, result)
	_ = s.audit.RecordAnalystAudit(ctx, tenantID, actorID, "ASSERTION_CONFIRMED", result.AssertionID, "")
	return result, nil
}

func (s *analystServiceImpl) RejectAssertion(ctx context.Context, tenantID, businessID, assertionID, actorID, comment string) (*entity.BusinessAssertion, error) {
	var result *entity.BusinessAssertion
	err := s.repo.Transaction(ctx, func(txRepo analystrepo.AnalystRepository) error {
		a, err := txRepo.GetAssertion(ctx, tenantID, assertionID)
		if err != nil {
			return err
		}
		if a == nil || a.BusinessID != businessID {
			return entity.ErrAssertionNotFound
		}
		if a.Status == entity.AssertionConfirmed || a.Status == entity.AssertionRejected {
			return entity.ErrAssertionAlreadyDecided
		}
		now := time.Now().UTC()
		a.Status = entity.AssertionRejected
		a.UpdatedAt = now
		if err := txRepo.UpdateAssertion(ctx, a); err != nil {
			return err
		}
		conf := &entity.BusinessConfirmation{
			ConfirmationID: newID("confm"),
			TenantID:       tenantID,
			BusinessID:     businessID,
			AssertionID:    assertionID,
			Decision:       entity.DecisionReject,
			Comment:        comment,
			DecidedBy:      actorID,
			DecidedAt:      now,
		}
		if err := txRepo.CreateConfirmation(ctx, conf); err != nil {
			return err
		}
		if err := s.resolveConflictsForAssertions(ctx, txRepo, tenantID, businessID, now); err != nil {
			return err
		}
		a.EvidenceIDs, _ = txRepo.ListEvidenceIDsForAssertion(ctx, tenantID, assertionID)
		result = a
		return nil
	})
	if err != nil {
		return nil, err
	}
	_ = s.audit.RecordAnalystAudit(ctx, tenantID, actorID, "ASSERTION_REJECTED", assertionID, "")
	return result, nil
}

func (s *analystServiceImpl) ApplyProposal(ctx context.Context, tenantID, businessID, proposalID, actorID string) (*businessentity.BusinessModelRevision, error) {
	proposal, err := s.repo.GetProposal(ctx, tenantID, proposalID)
	if err != nil {
		return nil, err
	}
	if proposal == nil || proposal.BusinessID != businessID {
		return nil, entity.ErrProposalNotFound
	}
	if proposal.Status == entity.ProposalApplied {
		return nil, entity.ErrProposalAlreadyApplied
	}
	master, _, _, err := s.businessSVC.GetModel(ctx, tenantID, businessID)
	if err != nil {
		return nil, err
	}
	if master == nil {
		return nil, entity.ErrNotFound
	}
	if master.CurrentRevision != proposal.BaseRevision {
		_, _ = s.repo.MarkProposalStaleIfReady(ctx, tenantID, proposalID, time.Now().UTC())
		return nil, entity.ErrProposalStale
	}
	if s.db == nil || s.businessRepo == nil {
		return nil, entity.ErrNotConfigured
	}

	var rev *businessentity.BusinessModelRevision
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ar := analystrepo.NewAnalystRepository(tx)
		br := businessrepo.NewBusinessRepository(tx)
		r, err := s.applyProposalWithRepos(ctx, ar, br, tenantID, businessID, proposalID, actorID)
		if err != nil {
			return err
		}
		rev = r
		return nil
	})
	if err != nil {
		if errors.Is(err, entity.ErrProposalStale) {
			_, _ = s.repo.MarkProposalStaleIfReady(ctx, tenantID, proposalID, time.Now().UTC())
		}
		return nil, err
	}
	return rev, nil
}

func (s *analystServiceImpl) applyProposalWithRepos(
	ctx context.Context,
	ar analystrepo.AnalystRepository,
	br businessrepo.BusinessRepository,
	tenantID, businessID, proposalID, actorID string,
) (*businessentity.BusinessModelRevision, error) {
	proposal, err := ar.GetProposalForUpdate(ctx, tenantID, proposalID)
	if err != nil {
		return nil, err
	}
	if proposal == nil || proposal.BusinessID != businessID {
		return nil, entity.ErrProposalNotFound
	}
	if proposal.Status == entity.ProposalApplied {
		return nil, entity.ErrProposalAlreadyApplied
	}
	if proposal.Status == entity.ProposalRejected || proposal.Status == entity.ProposalStale {
		return nil, entity.ErrProposalInvalid
	}
	if proposal.Status != entity.ProposalReadyForReview {
		return nil, entity.ErrProposalInvalid
	}

	master, err := br.GetMaster(ctx, tenantID, businessID)
	if err != nil {
		return nil, err
	}
	if master == nil {
		return nil, entity.ErrNotFound
	}
	if master.CurrentRevision != proposal.BaseRevision {
		return nil, entity.ErrProposalStale
	}
	curRev, err := br.GetRevision(ctx, tenantID, businessID, master.CurrentRevision)
	if err != nil {
		return nil, err
	}
	if curRev == nil {
		return nil, businessentity.ErrRevisionNotFound
	}
	model, err := businesssvc.ParseSemanticJSON(curRev.SemanticModelJSON)
	if err != nil {
		return nil, err
	}
	newModel, err := ApplyPatch(model, proposal.Patch)
	if err != nil {
		return nil, err
	}
	newModel.EvidenceRefs = collectEvidenceRefs(ctx, ar, tenantID, businessID, proposal.AssertionIDs)
	newModel.AssertionRefs = append([]string(nil), proposal.AssertionIDs...)

	summary := fmt.Sprintf("Apply analyst proposal %s", proposal.ProposalID)
	rev, noChange, err := businesssvc.SaveModelRevision(ctx, br, tenantID, businessID, actorID, proposal.BaseRevision, newModel, summary)
	if err != nil {
		return nil, err
	}
	if noChange {
		return nil, entity.ErrProposalInvalid
	}
	now := time.Now().UTC()
	if err := ar.UpdateProposalStatus(ctx, tenantID, proposalID, entity.ProposalApplied, now); err != nil {
		return nil, err
	}
	if err := ar.CreateProvenance(ctx, &entity.RevisionProvenance{
		TenantID:     tenantID,
		BusinessID:   businessID,
		RevisionNo:   rev.RevisionNo,
		ProposalID:   proposal.ProposalID,
		AssertionIDs: proposal.AssertionIDs,
		CreatedAt:    now,
	}); err != nil {
		return nil, err
	}
	_ = s.audit.RecordAnalystAudit(ctx, tenantID, actorID, "PROPOSAL_APPLIED", proposalID, "")
	return rev, nil
}

func (s *analystServiceImpl) RetryTurnAnalysis(ctx context.Context, tenantID, businessID, sessionID, turnID, actorID string) (*entity.TurnSubmissionResult, error) {
	turn, err := s.repo.GetTurn(ctx, tenantID, turnID)
	if err != nil {
		return nil, err
	}
	if turn == nil || turn.SessionID != sessionID || turn.Speaker != entity.SpeakerUser {
		return nil, entity.ErrInvalidTurn
	}
	switch turn.AnalysisStatus {
	case entity.AnalysisExtractionFailed, entity.AnalysisFailed, entity.AnalysisPending:
		// full pipeline
	case entity.AnalysisResponseFailed:
		// generation only
	default:
		return nil, entity.ErrInvalidTurn
	}
	evList, err := s.repo.ListEvidence(ctx, tenantID, businessID)
	if err != nil {
		return nil, err
	}
	var evidence *entity.BusinessEvidence
	for _, e := range evList {
		if e != nil && e.TurnID == turnID {
			evidence = e
			break
		}
	}
	if evidence == nil {
		return nil, entity.ErrInvalidTurn
	}
	if turn.AnalysisStatus == entity.AnalysisResponseFailed {
		return s.runGenerationOnly(ctx, tenantID, businessID, sessionID, turn, evidence, actorID)
	}
	return s.runFullAnalysis(ctx, tenantID, businessID, sessionID, turn, evidence, actorID)
}

func (s *analystServiceImpl) GetProposalPreview(ctx context.Context, tenantID, proposalID string) (*ProposalPreviewResult, error) {
	proposal, err := s.GetProposal(ctx, tenantID, proposalID)
	if err != nil {
		return nil, err
	}
	master, model, _, err := s.businessSVC.GetModel(ctx, tenantID, proposal.BusinessID)
	if err != nil {
		return nil, err
	}
	out := &ProposalPreviewResult{
		Proposal:        proposal,
		CurrentRevision: master.CurrentRevision,
	}
	if master.CurrentRevision != proposal.BaseRevision {
		out.ValidationValid = false
		out.ValidationError = "proposal stale: base revision mismatch"
		return out, nil
	}
	proposed, err := ApplyPatch(model, proposal.Patch)
	if err != nil {
		out.ValidationValid = false
		out.ValidationError = err.Error()
		return out, nil
	}
	out.ProposedModel = proposed
	out.ValidationValid = true
	nextRev := master.CurrentRevision + 1
	out.Diff = businesssvc.DiffSemanticModels(master.CurrentRevision, nextRev, model, proposed)
	out.Impact = businesssvc.ImpactFromDiff(out.Diff)
	return out, nil
}
