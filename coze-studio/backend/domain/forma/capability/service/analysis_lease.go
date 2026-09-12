/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
)

const analysisExecutionLease = 5 * time.Minute

func leaseExpiryFrom(now time.Time) time.Time {
	return now.Add(analysisExecutionLease)
}

func hasActiveAnalysisLease(run *entity.CapabilityAnalysisRun, now time.Time) bool {
	if run == nil || run.Status != entity.AnalysisPending {
		return false
	}
	if run.LeaseExpiresAt == nil {
		return false
	}
	return now.Before(*run.LeaseExpiresAt)
}

func analysisLeaseExpired(run *entity.CapabilityAnalysisRun, now time.Time) bool {
	if run == nil || run.Status != entity.AnalysisPending {
		return false
	}
	if run.LeaseExpiresAt == nil {
		return true
	}
	return !now.Before(*run.LeaseExpiresAt)
}

func seedExecutionLease(run *entity.CapabilityAnalysisRun, now time.Time) {
	if run == nil {
		return
	}
	if run.Attempt == 0 {
		run.Attempt = 1
	}
	if run.ExecutionClaimedAt == nil {
		run.ExecutionClaimedAt = &now
	}
	if run.LeaseExpiresAt == nil {
		exp := leaseExpiryFrom(now)
		run.LeaseExpiresAt = &exp
	}
}
