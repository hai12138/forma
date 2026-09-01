/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"strings"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
)

const analysisPendingLease = 5 * time.Minute

const (
	analysisClaimPrefix = "analysis_claim:"
	retryClaimPrefix    = "retry_claim:"
)

func analysisClaimID(requestID string) string {
	return analysisClaimPrefix + requestID
}

func hasActiveAnalysisLease(turn *entity.AnalystTurn, now time.Time) bool {
	if turn == nil || turn.AnalysisStatus != entity.AnalysisPending {
		return false
	}
	if !strings.HasPrefix(turn.ModelRequestID, analysisClaimPrefix) {
		return false
	}
	return now.Sub(turn.CreatedAt) < analysisPendingLease
}

func analysisLeaseExpired(turn *entity.AnalystTurn, now time.Time) bool {
	if turn == nil || !strings.HasPrefix(turn.ModelRequestID, analysisClaimPrefix) {
		return true
	}
	return now.Sub(turn.CreatedAt) >= analysisPendingLease
}
