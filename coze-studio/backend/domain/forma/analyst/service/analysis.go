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
)

func (s *analystServiceImpl) createUserTurnWithEvidence(
	ctx context.Context,
	tenantID, businessID, sessionID, content, clientRequestID, actorID string,
) (*entity.AnalystTurn, *entity.BusinessEvidence, bool, error) {
	var userTurn *entity.AnalystTurn
	var evidence *entity.BusinessEvidence
	createdNew := false
	err := s.repo.Transaction(ctx, func(txRepo repository.AnalystRepository) error {
		session, err := txRepo.GetSessionForUpdate(ctx, tenantID, sessionID)
		if err != nil {
			return err
		}
		if session == nil {
			return entity.ErrSessionNotFound
		}
		if session.NextTurnSequence <= 0 {
			session.NextTurnSequence = 1
		}

		clientReq := clientRequestID

		if clientRequestID != "" {
			existing, gErr := txRepo.GetTurnByClientRequestID(ctx, tenantID, sessionID, clientRequestID)
			if gErr != nil {
				return gErr
			}
			if existing != nil {
				userTurn = existing
				evList, _ := txRepo.ListEvidence(ctx, tenantID, businessID)
				for _, e := range evList {
					if e != nil && e.TurnID == existing.TurnID {
						evidence = e
						break
					}
				}
				return nil
			}
		}

		now := time.Now().UTC()
		userSeq := session.NextTurnSequence
		analystSeq := session.NextTurnSequence + 1
		session.NextTurnSequence += 2
		session.UpdatedAt = now
		if err := txRepo.UpdateSession(ctx, session); err != nil {
			return err
		}

		turnID := newID("turn")
		if clientReq == "" {
			clientReq = turnID
		}
		ut := &entity.AnalystTurn{
			TurnID:                turnID,
			TenantID:              tenantID,
			SessionID:             sessionID,
			Sequence:              userSeq,
			Speaker:               entity.SpeakerUser,
			Content:               content,
			ContentType:           entity.ContentText,
			ClientRequestID:       clientReq,
			ReservedReplySequence: analystSeq,
			AnalysisStatus:        entity.AnalysisPending,
			CreatedAt:             now,
		}
		if err := txRepo.CreateTurn(ctx, ut); err != nil {
			return err
		}
		ev := &entity.BusinessEvidence{
			EvidenceID:    newID("evid"),
			TenantID:      tenantID,
			BusinessID:    businessID,
			SessionID:     sessionID,
			TurnID:        ut.TurnID,
			SourceType:    entity.EvidenceInterviewTurn,
			SourceRef:     ut.TurnID,
			Quote:         content,
			ContentDigest: digestString(content),
			CreatedBy:     actorID,
			CreatedAt:     now,
		}
		if err := txRepo.CreateEvidence(ctx, ev); err != nil {
			return err
		}
		userTurn = ut
		evidence = ev
		createdNew = true
		return nil
	})
	if err != nil {
		if clientRequestID != "" {
			existing, gErr := s.repo.GetTurnByClientRequestID(ctx, tenantID, sessionID, clientRequestID)
			if gErr == nil && existing != nil {
				evList, _ := s.repo.ListEvidence(ctx, tenantID, businessID)
				for _, e := range evList {
					if e != nil && e.TurnID == existing.TurnID {
						return existing, e, false, nil
					}
				}
				return existing, nil, false, nil
			}
		}
		return nil, nil, false, err
	}
	return userTurn, evidence, createdNew, nil
}

func (s *analystServiceImpl) runAnalysisForUserTurn(
	ctx context.Context,
	tenantID, businessID, sessionID string,
	userTurn *entity.AnalystTurn,
	evidence *entity.BusinessEvidence,
	actorID string,
) (*entity.TurnSubmissionResult, error) {
	if userTurn.AnalysisStatus == entity.AnalysisCompleted {
		return s.buildIdempotentResult(ctx, tenantID, businessID, sessionID, userTurn)
	}
	if userTurn.AnalysisStatus == entity.AnalysisResponseFailed {
		return s.runGenerationOnly(ctx, tenantID, businessID, sessionID, userTurn, evidence, actorID)
	}
	return s.runFullAnalysis(ctx, tenantID, businessID, sessionID, userTurn, evidence, actorID)
}

func (s *analystServiceImpl) runFullAnalysis(
	ctx context.Context,
	tenantID, businessID, sessionID string,
	userTurn *entity.AnalystTurn,
	evidence *entity.BusinessEvidence,
	actorID string,
) (*entity.TurnSubmissionResult, error) {
	now := time.Now().UTC()
	ctxInput, manifest, contextText, err := s.buildContextInput(ctx, tenantID, businessID, sessionID)
	if err != nil {
		return nil, err
	}
	if contextText != "" {
		manifest.IncludedItems = append(manifest.IncludedItems, "rendered_context")
	}

	extractReqID := newID("mreq")
	analysisClaim := analysisClaimID(extractReqID)
	_ = s.repo.UpdateTurnAnalysis(ctx, tenantID, userTurn.TurnID, entity.AnalysisPending, analysisClaim)
	userTurn.ModelRequestID = analysisClaim
	startExtract := time.Now()
	extractReq := &ExtractionRequest{
		RequestID:       extractReqID,
		TenantID:        tenantID,
		BusinessID:      businessID,
		SessionID:       sessionID,
		ContextManifest: manifest,
		ContextText:     contextText,
		SystemPolicy:    analystSystemPolicy,
		UserTurnContent: userTurn.Content,
		UserTurnID:      userTurn.TurnID,
	}
	outcome, modelErr := s.model.ExtractAssertions(ctx, extractReq)
	extractLatency := int32(time.Since(startExtract).Milliseconds())

	result := &entity.TurnSubmissionResult{
		UserTurn:        userTurn,
		Evidence:        evidence,
		ContextManifest: manifest,
	}

	if modelErr != nil {
		_ = s.repo.UpdateTurnAnalysis(ctx, tenantID, userTurn.TurnID, entity.AnalysisExtractionFailed, extractReqID)
		_ = s.repo.CreateModelCall(ctx, &entity.ModelCallRecord{
			RequestID:    extractReqID,
			TenantID:     tenantID,
			BusinessID:   businessID,
			SessionID:    sessionID,
			Operation:    "ExtractAssertions",
			LatencyMs:    extractLatency,
			Success:      false,
			ErrorMessage: modelErr.Error(),
			CreatedAt:    now,
		})
		result.ModelFailed = true
		result.ModelError = modelErr.Error()
		userTurn.AnalysisStatus = entity.AnalysisExtractionFailed
		return result, nil
	}

	extractModelRef := ""
	inTok, outTok := int32(0), int32(0)
	if outcome != nil {
		extractModelRef = outcome.ModelRef
		inTok = outcome.InputTokens
		outTok = outcome.OutputTokens
	}
	_ = s.repo.CreateModelCall(ctx, &entity.ModelCallRecord{
		RequestID:    extractReqID,
		TenantID:     tenantID,
		BusinessID:   businessID,
		SessionID:    sessionID,
		Operation:    "ExtractAssertions",
		ModelRef:     extractModelRef,
		LatencyMs:    extractLatency,
		Success:      true,
		InputTokens:  inTok,
		OutputTokens: outTok,
		CreatedAt:    now,
	})

	extraction := outcome.Result
	assertions, conflicts, gaps, err := s.persistExtraction(ctx, tenantID, businessID, sessionID, actorID, evidence, extraction, now)
	if err != nil {
		_ = s.repo.UpdateTurnAnalysis(ctx, tenantID, userTurn.TurnID, entity.AnalysisExtractionFailed, extractReqID)
		result.ModelFailed = true
		result.ModelError = err.Error()
		userTurn.AnalysisStatus = entity.AnalysisExtractionFailed
		return result, nil
	}

	return s.completeGeneration(ctx, tenantID, businessID, sessionID, userTurn, evidence, actorID, manifest, contextText, ctxInput, assertions, conflicts, gaps, result)
}

func (s *analystServiceImpl) runGenerationOnly(
	ctx context.Context,
	tenantID, businessID, sessionID string,
	userTurn *entity.AnalystTurn,
	evidence *entity.BusinessEvidence,
	actorID string,
) (*entity.TurnSubmissionResult, error) {
	ctxInput, manifest, contextText, err := s.buildContextInput(ctx, tenantID, businessID, sessionID)
	if err != nil {
		return nil, err
	}
	assertions, err := s.loadAssertionsForEvidence(ctx, tenantID, businessID, evidence)
	if err != nil {
		return nil, err
	}
	conflicts, _ := s.repo.ListConflicts(ctx, tenantID, businessID)
	gaps, _ := s.repo.ListGaps(ctx, tenantID, businessID)
	result := &entity.TurnSubmissionResult{
		UserTurn:        userTurn,
		Evidence:        evidence,
		ContextManifest: manifest,
		Assertions:      assertions,
		Conflicts:       conflicts,
		Gaps:            gaps,
	}
	return s.completeGeneration(ctx, tenantID, businessID, sessionID, userTurn, evidence, actorID, manifest, contextText, ctxInput, assertions, conflicts, gaps, result)
}

func (s *analystServiceImpl) completeGeneration(
	ctx context.Context,
	tenantID, businessID, sessionID string,
	userTurn *entity.AnalystTurn,
	evidence *entity.BusinessEvidence,
	actorID string,
	manifest *entity.ContextManifest,
	contextText string,
	ctxInput *ContextInput,
	assertions []*entity.BusinessAssertion,
	conflicts []*entity.AssertionConflict,
	gaps []*entity.AnalystGap,
	result *entity.TurnSubmissionResult,
) (*entity.TurnSubmissionResult, error) {
	now := time.Now().UTC()
	plan := PlanNextQuestion(ctxInput)
	genReqID := newID("mreq")
	planReq := &InterviewTurnRequest{
		RequestID:       genReqID,
		TenantID:        tenantID,
		BusinessID:      businessID,
		SessionID:       sessionID,
		ContextManifest: manifest,
		ContextText:     contextText,
		SystemPolicy:    analystSystemPolicy,
		UserMessage:     userTurn.Content,
		NextQuestion:    plan,
	}
	startGen := time.Now()
	turnResp, turnErr := s.model.GenerateInterviewTurn(ctx, planReq)
	genLatency := int32(time.Since(startGen).Milliseconds())

	if turnErr != nil {
		_ = s.repo.UpdateTurnAnalysis(ctx, tenantID, userTurn.TurnID, entity.AnalysisResponseFailed, genReqID)
		_ = s.repo.CreateModelCall(ctx, &entity.ModelCallRecord{
			RequestID:    genReqID,
			TenantID:     tenantID,
			BusinessID:   businessID,
			SessionID:    sessionID,
			Operation:    "GenerateInterviewTurn",
			LatencyMs:    genLatency,
			Success:      false,
			ErrorMessage: turnErr.Error(),
			CreatedAt:    now,
		})
		result.ModelFailed = true
		result.ModelError = turnErr.Error()
		result.Assertions = assertions
		result.Conflicts = conflicts
		result.Gaps = gaps
		result.NextQuestion = plan
		userTurn.AnalysisStatus = entity.AnalysisResponseFailed
		return result, nil
	}

	modelRef := ""
	inTok, outTok := int32(0), int32(0)
	analystContent := "我已分析您的描述并提取了业务事实，请在右侧查看并确认。"
	if turnResp != nil {
		modelRef = turnResp.ModelRef
		inTok = turnResp.InputTokens
		outTok = turnResp.OutputTokens
		if turnResp.Content != "" {
			analystContent = turnResp.Content
		}
	}
	_ = s.repo.CreateModelCall(ctx, &entity.ModelCallRecord{
		RequestID:    genReqID,
		TenantID:     tenantID,
		BusinessID:   businessID,
		SessionID:    sessionID,
		Operation:    "GenerateInterviewTurn",
		ModelRef:     modelRef,
		LatencyMs:    genLatency,
		Success:      true,
		InputTokens:  inTok,
		OutputTokens: outTok,
		CreatedAt:    now,
	})

	existingAnalyst, _ := s.repo.GetTurnByReplyTo(ctx, tenantID, sessionID, userTurn.TurnID)
	if existingAnalyst != nil {
		result.Assertions = assertions
		result.Conflicts = conflicts
		result.Gaps = gaps
		result.AnalystTurn = existingAnalyst
		result.NextQuestion = plan
		userTurn.AnalysisStatus = entity.AnalysisCompleted
		return result, nil
	}

	analystSeq := userTurn.ReservedReplySequence
	if analystSeq <= 0 {
		analystSeq = userTurn.Sequence + 1
	}
	analystTurnID := newID("turn")
	analystTurn := &entity.AnalystTurn{
		TurnID:          analystTurnID,
		TenantID:        tenantID,
		SessionID:       sessionID,
		Sequence:        analystSeq,
		Speaker:         entity.SpeakerAnalyst,
		Content:         analystContent,
		ContentType:     entity.ContentText,
		ClientRequestID: analystTurnID,
		ReplyToTurnID:   userTurn.TurnID,
		ModelRequestID:  genReqID,
		AnalysisStatus:  entity.AnalysisCompleted,
		CreatedAt:       now,
	}
	if err := s.repo.CreateTurn(ctx, analystTurn); err != nil {
		return nil, err
	}

	_ = s.repo.UpdateTurnAnalysis(ctx, tenantID, userTurn.TurnID, entity.AnalysisCompleted, genReqID)
	userTurn.AnalysisStatus = entity.AnalysisCompleted

	result.Assertions = assertions
	result.Conflicts = conflicts
	result.Gaps = gaps
	result.AnalystTurn = analystTurn
	result.NextQuestion = plan
	return result, nil
}

func (s *analystServiceImpl) loadAssertionsForEvidence(
	ctx context.Context,
	tenantID, businessID string,
	evidence *entity.BusinessEvidence,
) ([]*entity.BusinessAssertion, error) {
	if evidence == nil {
		return nil, nil
	}
	ids, err := s.repo.ListAssertionIDsForEvidence(ctx, tenantID, evidence.EvidenceID)
	if err != nil {
		return nil, err
	}
	var out []*entity.BusinessAssertion
	for _, id := range ids {
		a, err := s.repo.GetAssertion(ctx, tenantID, id)
		if err != nil {
			return nil, err
		}
		if a != nil && a.BusinessID == businessID {
			a.EvidenceIDs, _ = s.repo.ListEvidenceIDsForAssertion(ctx, tenantID, id)
			out = append(out, a)
		}
	}
	return out, nil
}
