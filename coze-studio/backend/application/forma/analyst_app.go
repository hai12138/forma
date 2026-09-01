/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"
	"time"

	analystentity "github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
	analystsvc "github.com/coze-dev/coze-studio/backend/domain/forma/analyst/service"
	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenantctx "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/context"
)

type AnalystSessionDTO struct {
	SessionID          string `json:"session_id"`
	BusinessID         string `json:"business_id"`
	Status             string `json:"status"`
	Title              string `json:"title"`
	ConfirmationPolicy string `json:"confirmation_policy"`
	CreatedBy          string `json:"created_by"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

type CreateAnalystSessionInput struct {
	Title              string `json:"title"`
	ConfirmationPolicy string `json:"confirmation_policy"`
}

type SubmitTurnInput struct {
	Content         string `json:"content"`
	ClientRequestID string `json:"client_request_id"`
}

type TurnDTO struct {
	TurnID          string `json:"turn_id"`
	Sequence        int32  `json:"sequence"`
	Speaker         string `json:"speaker"`
	Content         string `json:"content"`
	ContentType     string `json:"content_type"`
	AnalysisStatus  string `json:"analysis_status"`
	ClientRequestID string `json:"client_request_id,omitempty"`
	CreatedAt       string `json:"created_at"`
}

type EvidenceDTO struct {
	EvidenceID string `json:"evidence_id"`
	SessionID  string `json:"session_id"`
	TurnID     string `json:"turn_id"`
	SourceType string `json:"source_type"`
	Quote      string `json:"quote"`
	CreatedBy  string `json:"created_by"`
	CreatedAt  string `json:"created_at"`
}

type AssertionDTO struct {
	AssertionID     string         `json:"assertion_id"`
	SessionID       string         `json:"session_id"`
	AssertionType   string         `json:"assertion_type"`
	SubjectRef      string         `json:"subject_ref"`
	Predicate       string         `json:"predicate"`
	ObjectValue     string         `json:"object_value"`
	Confidence      float64        `json:"confidence"`
	Status          string         `json:"status"`
	SourceMarker    string         `json:"source_marker"`
	EvidenceIDs     []string       `json:"evidence_ids"`
	StructuredValue map[string]any `json:"structured_value,omitempty"`
	CreatedAt       string         `json:"created_at"`
	UpdatedAt       string         `json:"updated_at"`
}

type ConflictDTO struct {
	ConflictID   string `json:"conflict_id"`
	SessionID    string `json:"session_id"`
	AssertionIDA string `json:"assertion_id_a"`
	AssertionIDB string `json:"assertion_id_b"`
	SubjectRef   string `json:"subject_ref"`
	Predicate    string `json:"predicate"`
	Status       string `json:"status"`
}

type GapDTO struct {
	GapID               string   `json:"gap_id"`
	SessionID           string   `json:"session_id"`
	GapType             string   `json:"gap_type"`
	Question            string   `json:"question"`
	Status              string   `json:"status"`
	RelatedAssertionIDs []string `json:"related_assertion_ids,omitempty"`
}

type ProposalDTO struct {
	ProposalID    string                            `json:"proposal_id"`
	BusinessID    string                            `json:"business_id"`
	SessionID     string                            `json:"session_id"`
	BaseRevision  int32                             `json:"base_revision"`
	AssertionIDs  []string                          `json:"assertion_ids"`
	Patch         *analystentity.SemanticModelPatch `json:"patch"`
	Status        string                            `json:"status"`
	ContentDigest string                            `json:"content_digest"`
	CreatedAt     string                            `json:"created_at"`
}

type TurnSubmissionResponse struct {
	UserTurn     *TurnDTO                        `json:"user_turn"`
	AnalystTurn  *TurnDTO                        `json:"analyst_turn,omitempty"`
	Evidence     *EvidenceDTO                    `json:"evidence,omitempty"`
	Assertions   []*AssertionDTO                 `json:"assertions,omitempty"`
	Conflicts    []*ConflictDTO                  `json:"conflicts,omitempty"`
	Gaps         []*GapDTO                       `json:"gaps,omitempty"`
	ModelFailed  bool                            `json:"model_failed,omitempty"`
	ModelError   string                          `json:"model_error,omitempty"`
	NextQuestion *analystentity.NextQuestionPlan `json:"next_question,omitempty"`
}

type ConfirmAssertionInput struct {
	Comment string              `json:"comment"`
	Edit    *AssertionEditInput `json:"edit,omitempty"`
}

type AssertionEditInput struct {
	AssertionType string `json:"assertion_type"`
	SubjectRef    string `json:"subject_ref"`
	Predicate     string `json:"predicate"`
	ObjectValue   string `json:"object_value"`
}

type CreateProposalInput struct {
	SessionID    string   `json:"session_id"`
	AssertionIDs []string `json:"assertion_ids"`
}

type ApplyProposalResponse struct {
	RevisionNo int32  `json:"revision_no"`
	ProposalID string `json:"proposal_id"`
}

type ProposalPreviewResponse struct {
	Proposal        *ProposalDTO                          `json:"proposal"`
	CurrentRevision int32                                 `json:"current_revision"`
	ValidationValid bool                                  `json:"validation_valid"`
	ValidationError string                                `json:"validation_error,omitempty"`
	AssertionCount  int                                   `json:"assertion_count"`
	ProposedModel   *businessentity.SemanticModel         `json:"proposed_model,omitempty"`
	Diff            *businessentity.BusinessModelDiff     `json:"diff,omitempty"`
	Impact          *businessentity.BusinessImpactSummary `json:"impact,omitempty"`
}

func (s *ApplicationService) CreateAnalystSession(ctx context.Context, businessID string, in *CreateAnalystSessionInput) (*AnalystSessionDTO, error) {
	tc, err := s.requireAnalystTenant(ctx)
	if err != nil {
		return nil, err
	}
	title := ""
	policy := analystentity.PolicyDevelopment
	if in != nil {
		title = in.Title
		if in.ConfirmationPolicy != "" {
			policy = analystentity.ConfirmationPolicy(in.ConfirmationPolicy)
		}
	}
	session, err := s.AnalystSVC.CreateSession(ctx, tc.TenantID, businessID, title, tc.PrincipalID, policy)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return toSessionDTO(session), nil
}

func (s *ApplicationService) ListAnalystSessions(ctx context.Context, businessID string) ([]*AnalystSessionDTO, error) {
	tc, err := s.requireAnalystTenant(ctx)
	if err != nil {
		return nil, err
	}
	list, err := s.AnalystSVC.ListSessions(ctx, tc.TenantID, businessID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	out := make([]*AnalystSessionDTO, 0, len(list))
	for _, sess := range list {
		out = append(out, toSessionDTO(sess))
	}
	return out, nil
}

func (s *ApplicationService) GetAnalystSession(ctx context.Context, businessID, sessionID string) (*AnalystSessionDTO, error) {
	tc, err := s.requireAnalystTenant(ctx)
	if err != nil {
		return nil, err
	}
	session, err := s.AnalystSVC.GetSession(ctx, tc.TenantID, sessionID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	if session.BusinessID != businessID {
		return nil, formaerrors.AnalystSessionNotFound("session not found")
	}
	return toSessionDTO(session), nil
}

func (s *ApplicationService) SubmitAnalystTurn(ctx context.Context, businessID, sessionID string, in *SubmitTurnInput) (*TurnSubmissionResponse, error) {
	tc, err := s.requireAnalystTenant(ctx)
	if err != nil {
		return nil, err
	}
	content := ""
	clientReq := ""
	if in != nil {
		content = in.Content
		clientReq = in.ClientRequestID
	}
	result, err := s.AnalystSVC.SubmitTurn(ctx, tc.TenantID, businessID, sessionID, content, clientReq, tc.PrincipalID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return toTurnSubmissionResponse(result), nil
}

func (s *ApplicationService) ListAnalystTurns(ctx context.Context, businessID, sessionID string) ([]*TurnDTO, error) {
	tc, err := s.requireAnalystTenant(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetAnalystSession(ctx, businessID, sessionID); err != nil {
		return nil, err
	}
	turns, err := s.AnalystSVC.ListTurns(ctx, tc.TenantID, sessionID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	out := make([]*TurnDTO, 0, len(turns))
	for _, t := range turns {
		out = append(out, toTurnDTO(t))
	}
	return out, nil
}

func (s *ApplicationService) ListAssertions(ctx context.Context, businessID string) ([]*AssertionDTO, error) {
	tc, err := s.requireAnalystTenant(ctx)
	if err != nil {
		return nil, err
	}
	list, err := s.AnalystSVC.ListAssertions(ctx, tc.TenantID, businessID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	out := make([]*AssertionDTO, 0, len(list))
	for _, a := range list {
		out = append(out, toAssertionDTO(a))
	}
	return out, nil
}

func (s *ApplicationService) ListEvidence(ctx context.Context, businessID string) ([]*EvidenceDTO, error) {
	tc, err := s.requireAnalystTenant(ctx)
	if err != nil {
		return nil, err
	}
	list, err := s.AnalystSVC.ListEvidence(ctx, tc.TenantID, businessID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	out := make([]*EvidenceDTO, 0, len(list))
	for _, e := range list {
		out = append(out, toEvidenceDTO(e))
	}
	return out, nil
}

func (s *ApplicationService) ConfirmAssertion(ctx context.Context, businessID, assertionID string, in *ConfirmAssertionInput) (*AssertionDTO, error) {
	tc, err := s.requireAnalystConfirm(ctx)
	if err != nil {
		return nil, err
	}
	var edit *analystsvc.AssertionEdit
	if in != nil && in.Edit != nil {
		edit = &analystsvc.AssertionEdit{
			AssertionType: analystentity.AssertionType(in.Edit.AssertionType),
			SubjectRef:    in.Edit.SubjectRef,
			Predicate:     in.Edit.Predicate,
			ObjectValue:   in.Edit.ObjectValue,
		}
	}
	comment := ""
	if in != nil {
		comment = in.Comment
	}
	a, err := s.AnalystSVC.ConfirmAssertion(ctx, tc.TenantID, businessID, assertionID, tc.PrincipalID, comment, edit)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return toAssertionDTO(a), nil
}

func (s *ApplicationService) RejectAssertion(ctx context.Context, businessID, assertionID string, in *ConfirmAssertionInput) (*AssertionDTO, error) {
	tc, err := s.requireAnalystConfirm(ctx)
	if err != nil {
		return nil, err
	}
	comment := ""
	if in != nil {
		comment = in.Comment
	}
	a, err := s.AnalystSVC.RejectAssertion(ctx, tc.TenantID, businessID, assertionID, tc.PrincipalID, comment)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return toAssertionDTO(a), nil
}

func (s *ApplicationService) ListConflicts(ctx context.Context, businessID string) ([]*ConflictDTO, error) {
	tc, err := s.requireAnalystTenant(ctx)
	if err != nil {
		return nil, err
	}
	list, err := s.AnalystSVC.ListConflicts(ctx, tc.TenantID, businessID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	out := make([]*ConflictDTO, 0, len(list))
	for _, c := range list {
		out = append(out, toConflictDTO(c))
	}
	return out, nil
}

func (s *ApplicationService) ListGaps(ctx context.Context, businessID string) ([]*GapDTO, error) {
	tc, err := s.requireAnalystTenant(ctx)
	if err != nil {
		return nil, err
	}
	list, err := s.AnalystSVC.ListGaps(ctx, tc.TenantID, businessID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	out := make([]*GapDTO, 0, len(list))
	for _, g := range list {
		out = append(out, toGapDTO(g))
	}
	return out, nil
}

func (s *ApplicationService) CreateProposal(ctx context.Context, businessID string, in *CreateProposalInput) (*ProposalDTO, error) {
	tc, err := s.requireAnalystConfirm(ctx)
	if err != nil {
		return nil, err
	}
	sessionID := ""
	var assertionIDs []string
	if in != nil {
		sessionID = in.SessionID
		assertionIDs = in.AssertionIDs
	}
	p, err := s.AnalystSVC.CreateProposal(ctx, tc.TenantID, businessID, sessionID, tc.PrincipalID, assertionIDs)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return toProposalDTO(p), nil
}

func (s *ApplicationService) GetProposal(ctx context.Context, businessID, proposalID string) (*ProposalDTO, error) {
	tc, err := s.requireAnalystTenant(ctx)
	if err != nil {
		return nil, err
	}
	p, err := s.AnalystSVC.GetProposal(ctx, tc.TenantID, proposalID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	if p.BusinessID != businessID {
		return nil, formaerrors.AnalystProposalNotFound("proposal not found")
	}
	return toProposalDTO(p), nil
}

func (s *ApplicationService) ApplyProposal(ctx context.Context, businessID, proposalID string) (*ApplyProposalResponse, error) {
	tc, err := s.requireAnalystConfirm(ctx)
	if err != nil {
		return nil, err
	}
	rev, err := s.AnalystSVC.ApplyProposal(ctx, tc.TenantID, businessID, proposalID, tc.PrincipalID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return &ApplyProposalResponse{
		RevisionNo: rev.RevisionNo,
		ProposalID: proposalID,
	}, nil
}

func (s *ApplicationService) RetryAnalystTurnAnalysis(ctx context.Context, businessID, sessionID, turnID string) (*TurnSubmissionResponse, error) {
	tc, err := s.requireAnalystTenant(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetAnalystSession(ctx, businessID, sessionID); err != nil {
		return nil, err
	}
	result, err := s.AnalystSVC.RetryTurnAnalysis(ctx, tc.TenantID, businessID, sessionID, turnID, tc.PrincipalID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return toTurnSubmissionResponse(result), nil
}

func (s *ApplicationService) GetProposalPreview(ctx context.Context, businessID, proposalID string) (*ProposalPreviewResponse, error) {
	tc, err := s.requireAnalystTenant(ctx)
	if err != nil {
		return nil, err
	}
	p, err := s.AnalystSVC.GetProposal(ctx, tc.TenantID, proposalID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	if p.BusinessID != businessID {
		return nil, formaerrors.AnalystProposalNotFound("proposal not found")
	}
	preview, err := s.AnalystSVC.GetProposalPreview(ctx, tc.TenantID, proposalID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	out := &ProposalPreviewResponse{
		Proposal:        toProposalDTO(preview.Proposal),
		CurrentRevision: preview.CurrentRevision,
		ValidationValid: preview.ValidationValid,
		ValidationError: preview.ValidationError,
		AssertionCount:  len(preview.Proposal.AssertionIDs),
		ProposedModel:   preview.ProposedModel,
		Diff:            preview.Diff,
		Impact:          preview.Impact,
	}
	return out, nil
}

func (s *ApplicationService) requireAnalystTenant(ctx context.Context) (*tenantctx.TenantContext, error) {
	tc, ok := tenantctx.FromContext(ctx)
	if !ok || tc == nil || tc.TenantID == "" {
		return nil, formaerrors.TenantRequired("tenant context required")
	}
	if _, err := s.requireMemberOf(ctx, tc.TenantID); err != nil {
		return nil, err
	}
	if s.AnalystSVC == nil {
		return nil, formaerrors.Internal("analyst service not initialized")
	}
	return tc, nil
}

func (s *ApplicationService) requireAnalystConfirm(ctx context.Context) (*tenantctx.TenantContext, error) {
	tc, err := s.requireAnalystTenant(ctx)
	if err != nil {
		return nil, err
	}
	if !roleAtLeastAdmin(tc.MembershipRole) {
		return nil, formaerrors.AnalystForbidden("confirm/apply requires OWNER or ADMIN")
	}
	return tc, nil
}

func toSessionDTO(s *analystentity.AnalystSession) *AnalystSessionDTO {
	if s == nil {
		return nil
	}
	return &AnalystSessionDTO{
		SessionID:          s.SessionID,
		BusinessID:         s.BusinessID,
		Status:             string(s.Status),
		Title:              s.Title,
		ConfirmationPolicy: string(s.ConfirmationPolicy),
		CreatedBy:          s.CreatedBy,
		CreatedAt:          s.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:          s.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func toTurnDTO(t *analystentity.AnalystTurn) *TurnDTO {
	if t == nil {
		return nil
	}
	return &TurnDTO{
		TurnID:          t.TurnID,
		Sequence:        t.Sequence,
		Speaker:         string(t.Speaker),
		Content:         t.Content,
		ContentType:     string(t.ContentType),
		AnalysisStatus:  string(t.AnalysisStatus),
		ClientRequestID: t.ClientRequestID,
		CreatedAt:       t.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func toEvidenceDTO(e *analystentity.BusinessEvidence) *EvidenceDTO {
	if e == nil {
		return nil
	}
	return &EvidenceDTO{
		EvidenceID: e.EvidenceID,
		SessionID:  e.SessionID,
		TurnID:     e.TurnID,
		SourceType: string(e.SourceType),
		Quote:      e.Quote,
		CreatedBy:  e.CreatedBy,
		CreatedAt:  e.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func toAssertionDTO(a *analystentity.BusinessAssertion) *AssertionDTO {
	if a == nil {
		return nil
	}
	return &AssertionDTO{
		AssertionID:     a.AssertionID,
		SessionID:       a.SessionID,
		AssertionType:   string(a.AssertionType),
		SubjectRef:      a.SubjectRef,
		Predicate:       a.Predicate,
		ObjectValue:     a.ObjectValue,
		Confidence:      a.Confidence,
		Status:          string(a.Status),
		SourceMarker:    string(a.SourceMarker),
		EvidenceIDs:     a.EvidenceIDs,
		StructuredValue: a.StructuredValue,
		CreatedAt:       a.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:       a.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func toConflictDTO(c *analystentity.AssertionConflict) *ConflictDTO {
	if c == nil {
		return nil
	}
	return &ConflictDTO{
		ConflictID:   c.ConflictID,
		SessionID:    c.SessionID,
		AssertionIDA: c.AssertionIDA,
		AssertionIDB: c.AssertionIDB,
		SubjectRef:   c.SubjectRef,
		Predicate:    c.Predicate,
		Status:       string(c.Status),
	}
}

func toGapDTO(g *analystentity.AnalystGap) *GapDTO {
	if g == nil {
		return nil
	}
	return &GapDTO{
		GapID:               g.GapID,
		SessionID:           g.SessionID,
		GapType:             g.GapType,
		Question:            g.Question,
		Status:              string(g.Status),
		RelatedAssertionIDs: g.RelatedAssertionIDs,
	}
}

func toProposalDTO(p *analystentity.BusinessModelProposal) *ProposalDTO {
	if p == nil {
		return nil
	}
	return &ProposalDTO{
		ProposalID:    p.ProposalID,
		BusinessID:    p.BusinessID,
		SessionID:     p.SessionID,
		BaseRevision:  p.BaseRevision,
		AssertionIDs:  p.AssertionIDs,
		Patch:         p.Patch,
		Status:        string(p.Status),
		ContentDigest: p.ContentDigest,
		CreatedAt:     p.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func toTurnSubmissionResponse(r *analystentity.TurnSubmissionResult) *TurnSubmissionResponse {
	if r == nil {
		return nil
	}
	resp := &TurnSubmissionResponse{
		UserTurn:     toTurnDTO(r.UserTurn),
		AnalystTurn:  toTurnDTO(r.AnalystTurn),
		Evidence:     toEvidenceDTO(r.Evidence),
		ModelFailed:  r.ModelFailed,
		ModelError:   r.ModelError,
		NextQuestion: r.NextQuestion,
	}
	for _, a := range r.Assertions {
		resp.Assertions = append(resp.Assertions, toAssertionDTO(a))
	}
	for _, c := range r.Conflicts {
		resp.Conflicts = append(resp.Conflicts, toConflictDTO(c))
	}
	for _, g := range r.Gaps {
		resp.Gaps = append(resp.Gaps, toGapDTO(g))
	}
	return resp
}
