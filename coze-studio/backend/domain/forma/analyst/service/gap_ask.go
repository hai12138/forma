/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/repository"
)

func (s *analystServiceImpl) AskGap(ctx context.Context, tenantID, businessID, sessionID, gapID, actorID string) (*entity.GapAskResult, error) {
	var gap *entity.AnalystGap
	gaps, err := s.repo.ListGaps(ctx, tenantID, businessID)
	if err != nil {
		return nil, err
	}
	for _, g := range gaps {
		if g != nil && g.GapID == gapID {
			gap = g
			break
		}
	}
	if gap == nil || gap.SessionID != sessionID {
		return nil, entity.ErrGapNotFound
	}
	if gap.Status != entity.GapOpen {
		return nil, entity.ErrGapNotFound
	}

	var analystTurn *entity.AnalystTurn
	err = s.repo.Transaction(ctx, func(txRepo repository.AnalystRepository) error {
		session, err := txRepo.GetSessionForUpdate(ctx, tenantID, sessionID)
		if err != nil {
			return err
		}
		if session == nil || session.BusinessID != businessID {
			return entity.ErrSessionNotFound
		}
		if session.NextTurnSequence <= 0 {
			session.NextTurnSequence = 1
		}

		clientReq := fmt.Sprintf("gap_ask:%s", gapID)
		existing, gErr := txRepo.GetTurnByClientRequestID(ctx, tenantID, sessionID, clientReq)
		if gErr != nil {
			return gErr
		}
		if existing != nil {
			analystTurn = existing
			session.FocusGapID = gapID
			session.UpdatedAt = time.Now().UTC()
			return txRepo.UpdateSession(ctx, session)
		}

		now := time.Now().UTC()
		analystSeq := session.NextTurnSequence
		session.NextTurnSequence++
		session.FocusGapID = gapID
		session.UpdatedAt = now
		if err := txRepo.UpdateSession(ctx, session); err != nil {
			return err
		}

		turnID := newID("turn")
		at := &entity.AnalystTurn{
			TurnID:          turnID,
			TenantID:        tenantID,
			SessionID:       sessionID,
			Sequence:        analystSeq,
			Speaker:         entity.SpeakerAnalyst,
			Content:         gap.Question,
			ContentType:     entity.ContentText,
			ClientRequestID: clientReq,
			AnalysisStatus:  entity.AnalysisCompleted,
			CreatedAt:       now,
		}
		if err := txRepo.CreateTurn(ctx, at); err != nil {
			return err
		}
		analystTurn = at
		_ = s.audit.RecordAnalystAudit(ctx, tenantID, actorID, "GAP_ASK", gapID, turnID)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &entity.GapAskResult{
		AnalystTurn: analystTurn,
		Gap:         gap,
	}, nil
}

func (s *analystServiceImpl) tryResolveFocusedGap(ctx context.Context, tenantID, businessID, sessionID string, assertion *entity.BusinessAssertion) {
	if assertion == nil {
		return
	}
	session, err := s.repo.GetSession(ctx, tenantID, sessionID)
	if err != nil || session == nil || session.FocusGapID == "" {
		return
	}
	gaps, err := s.repo.ListGaps(ctx, tenantID, businessID)
	if err != nil {
		return
	}
	var gap *entity.AnalystGap
	for _, g := range gaps {
		if g != nil && g.GapID == session.FocusGapID && g.Status == entity.GapOpen {
			gap = g
			break
		}
	}
	if gap == nil {
		return
	}
	now := time.Now().UTC()
	_ = s.repo.UpdateGapStatus(ctx, tenantID, gap.GapID, entity.GapResolved, now)
	session.FocusGapID = ""
	session.UpdatedAt = now
	_ = s.repo.UpdateSession(ctx, session)
}
