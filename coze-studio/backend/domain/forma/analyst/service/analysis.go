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
) (*entity.AnalystTurn, *entity.BusinessEvidence, error) {
	var userTurn *entity.AnalystTurn
	var evidence *entity.BusinessEvidence
	err := s.repo.Transaction(ctx, func(txRepo repository.AnalystRepository) error {
		if _, err := txRepo.GetSessionForUpdate(ctx, tenantID, sessionID); err != nil {
			return err
		}
		now := time.Now().UTC()
		seq, err := txRepo.MaxTurnSequence(ctx, tenantID, sessionID)
		if err != nil {
			return err
		}
		turnID := newID("turn")
		clientReq := clientRequestID
		if clientReq == "" {
			clientReq = turnID
		}
		ut := &entity.AnalystTurn{
			TurnID:          turnID,
			TenantID:        tenantID,
			SessionID:       sessionID,
			Sequence:        seq + 1,
			Speaker:         entity.SpeakerUser,
			Content:         content,
			ContentType:     entity.ContentText,
			ClientRequestID: clientReq,
			AnalysisStatus:  entity.AnalysisPending,
			CreatedAt:       now,
		}
		if err := txRepo.CreateTurn(ctx, ut); err != nil {
			if isDuplicateKeyErr(err) && clientRequestID != "" {
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
		return nil
	})
	if err != nil {
		if clientRequestID != "" {
			existing, gErr := s.repo.GetTurnByClientRequestID(ctx, tenantID, sessionID, clientRequestID)
			if gErr == nil && existing != nil {
				evList, _ := s.repo.ListEvidence(ctx, tenantID, businessID)
				for _, e := range evList {
					if e.TurnID == existing.TurnID {
						return existing, e, nil
					}
				}
				return existing, nil, nil
			}
		}
		return nil, nil, err
	}
	return userTurn, evidence, nil
}

func (s *analystServiceImpl) runAnalysisForUserTurn(
	ctx context.Context,
	tenantID, businessID, sessionID string,
	userTurn *entity.AnalystTurn,
	evidence *entity.BusinessEvidence,
	actorID string,
) (*entity.TurnSubmissionResult, error) {
	now := time.Now().UTC()
	ctxInput, manifest, err := s.buildContextInput(ctx, tenantID, businessID, sessionID)
	if err != nil {
		return nil, err
	}
	_, contextText := BuildContext(ctxInput)
	manifest.IncludedItems = append(manifest.IncludedItems, contextText != "")

	extractReqID := newID("mreq")
	startExtract := time.Now()
	extractReq := &ExtractionRequest{
		RequestID:       extractReqID,
		TenantID:        tenantID,
		BusinessID:      businessID,
		SessionID:       sessionID,
		ContextManifest: manifest,
		SystemPolicy:    analystSystemPolicy,
		UserTurnContent: userTurn.Content,
		UserTurnID:      userTurn.TurnID,
	}
	extraction, modelErr := s.model.ExtractAssertions(ctx, extractReq)
	extractLatency := int32(time.Since(startExtract).Milliseconds())

	result := &entity.TurnSubmissionResult{
		UserTurn:        userTurn,
		Evidence:        evidence,
		ContextManifest: manifest,
	}

	if modelErr != nil {
		_ = s.repo.UpdateTurnAnalysis(ctx, tenantID, userTurn.TurnID, entity.AnalysisFailed, extractReqID)
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
		userTurn.AnalysisStatus = entity.AnalysisFailed
		return result, nil
	}

	_ = s.repo.CreateModelCall(ctx, &entity.ModelCallRecord{
		RequestID:  extractReqID,
		TenantID:   tenantID,
		BusinessID: businessID,
		SessionID:  sessionID,
		Operation:  "ExtractAssertions",
		LatencyMs:  extractLatency,
		Success:    true,
		CreatedAt:  now,
	})

	assertions, conflicts, gaps, err := s.persistExtraction(ctx, tenantID, businessID, sessionID, actorID, evidence, extraction, now)
	if err != nil {
		_ = s.repo.UpdateTurnAnalysis(ctx, tenantID, userTurn.TurnID, entity.AnalysisFailed, extractReqID)
		result.ModelFailed = true
		result.ModelError = err.Error()
		userTurn.AnalysisStatus = entity.AnalysisFailed
		return result, nil
	}
	_ = s.repo.UpdateTurnAnalysis(ctx, tenantID, userTurn.TurnID, entity.AnalysisCompleted, extractReqID)
	userTurn.AnalysisStatus = entity.AnalysisCompleted

	plan := PlanNextQuestion(ctxInput)
	genReqID := newID("mreq")
	planReq := &InterviewTurnRequest{
		RequestID:       genReqID,
		TenantID:        tenantID,
		BusinessID:      businessID,
		SessionID:       sessionID,
		ContextManifest: manifest,
		SystemPolicy:    analystSystemPolicy,
		UserMessage:     userTurn.Content,
		NextQuestion:    plan,
	}
	startGen := time.Now()
	turnResp, turnErr := s.model.GenerateInterviewTurn(ctx, planReq)
	genLatency := int32(time.Since(startGen).Milliseconds())

	if turnErr != nil {
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

	analystTurn := &entity.AnalystTurn{
		TurnID:         newID("turn"),
		TenantID:       tenantID,
		SessionID:      sessionID,
		Sequence:       userTurn.Sequence + 1,
		Speaker:        entity.SpeakerAnalyst,
		Content:        analystContent,
		ContentType:    entity.ContentText,
		ModelRequestID: genReqID,
		AnalysisStatus: entity.AnalysisCompleted,
		CreatedAt:      now,
	}
	if err := s.repo.CreateTurn(ctx, analystTurn); err != nil {
		return nil, err
	}

	result.Assertions = assertions
	result.Conflicts = conflicts
	result.Gaps = gaps
	result.AnalystTurn = analystTurn
	result.NextQuestion = plan
	return result, nil
}
