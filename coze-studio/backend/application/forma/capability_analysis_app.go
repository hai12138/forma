/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"

	capentity "github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	capsvc "github.com/coze-dev/coze-studio/backend/domain/forma/capability/service"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
)

// RetryCapabilityAnalysisInput is the optional body for explicit analysis retry.
type RetryCapabilityAnalysisInput struct {
	Reason string `json:"reason,omitempty"`
}

// StartCapabilityAnalysis starts (or replays) a capability analysis run.
// When the domain persists a FAILED run (any domain error with status FAILED),
// this method returns the FAILED DTO with a nil error so the handler can respond HTTP 200.
func (s *ApplicationService) StartCapabilityAnalysis(ctx context.Context, businessID string, in *StartCapabilityAnalysisInput) (*StartCapabilityAnalysisResponse, error) {
	tc, err := s.requireCapabilityAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, formaerrors.CapabilityInvalidPayload("analysis input required")
	}
	if err := s.requireCapabilityBusiness(ctx, tc.TenantID, businessID); err != nil {
		return nil, err
	}
	analysis := in.Analysis
	if analysis.BusinessModelRevision == 0 {
		analysis.BusinessModelRevision = in.BusinessModelRevision
	}
	res, err := s.CapabilitySVC.StartAnalysis(ctx, &capsvc.StartAnalysisInput{
		TenantID:              tc.TenantID,
		BusinessID:            businessID,
		BusinessModelRevision: analysis.BusinessModelRevision,
		ClientRequestID:       in.ClientRequestID,
		ActorID:               tc.PrincipalID,
		Analysis:              analysis,
	})
	if err != nil {
		if keepFailedAnalysisDTO(res) {
			s.recordCapabilityAudit(ctx, tc, "capability.analyze", res.Run.AnalysisRunID)
			return toStartAnalysisResponse(res), nil
		}
		return nil, formaerrors.MapDomainError(err)
	}
	s.recordCapabilityAudit(ctx, tc, "capability.analyze", res.Run.AnalysisRunID)
	return toStartAnalysisResponse(res), nil
}

// GetCapabilityAnalysis returns an analysis run scoped to tenant+business.
func (s *ApplicationService) GetCapabilityAnalysis(ctx context.Context, businessID, analysisRunID string) (*CapabilityAnalysisRunDTO, error) {
	tc, err := s.requireCapabilityRead(ctx)
	if err != nil {
		return nil, err
	}
	run, err := s.requireCapabilityAnalysis(ctx, tc.TenantID, businessID, analysisRunID)
	if err != nil {
		return nil, err
	}
	return capabilityAnalysisDTO(run), nil
}

// RetryCapabilityAnalysis explicitly retries a FAILED analysis run (no auto-retry on Start replay).
// Same FAILED-visibility contract as Start: persisted FAILED result returns HTTP-200-ready DTO.
func (s *ApplicationService) RetryCapabilityAnalysis(ctx context.Context, businessID, analysisRunID string, in *RetryCapabilityAnalysisInput) (*StartCapabilityAnalysisResponse, error) {
	tc, err := s.requireCapabilityAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.requireCapabilityBusiness(ctx, tc.TenantID, businessID); err != nil {
		return nil, err
	}
	if _, err := s.requireCapabilityAnalysis(ctx, tc.TenantID, businessID, analysisRunID); err != nil {
		return nil, err
	}
	if in != nil {
		if err := validateCapabilityRetryReason(in.Reason); err != nil {
			return nil, err
		}
	}
	// Domain RetryFailedAnalysis takes actor PrincipalID; optional reason is validated here for early reject.
	res, err := s.CapabilitySVC.RetryFailedAnalysis(ctx, tc.TenantID, analysisRunID, tc.PrincipalID)
	if err != nil {
		if keepFailedAnalysisDTO(res) {
			s.recordCapabilityAudit(ctx, tc, "capability.analyze.retry", res.Run.AnalysisRunID)
			return toStartAnalysisResponse(res), nil
		}
		return nil, formaerrors.MapDomainError(err)
	}
	s.recordCapabilityAudit(ctx, tc, "capability.analyze.retry", res.Run.AnalysisRunID)
	return toStartAnalysisResponse(res), nil
}

func validateCapabilityRetryReason(reason string) error {
	if err := capsvc.ValidateAuditMetadata(reason, ""); err != nil {
		return formaerrors.MapDomainError(err)
	}
	return nil
}

// keepFailedAnalysisDTO is true when the domain persisted a FAILED run that must surface as a success DTO
// (nil app error / HTTP 200 path). Status is authoritative — do not require errors.Is(ErrAnalysisFailed),
// because invalid proposals return ErrInvalidPayload after the run is marked FAILED.
func keepFailedAnalysisDTO(res *capsvc.AnalysisResult) bool {
	return res != nil && res.Run != nil && res.Run.Status == capentity.AnalysisFailed
}
