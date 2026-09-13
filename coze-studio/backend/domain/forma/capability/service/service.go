/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	assetentity "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/repository"
)

// AssetProjection writes Capability AssetRef headers (Create + projection update).
type AssetProjection interface {
	CreateCapabilityAsset(ctx context.Context, asset *assetentity.AssetRef) error
	UpdateCapabilityProjection(ctx context.Context, tenantID, assetID, name, semanticVersion, contentDigest string, status assetentity.AssetStatus) error
	GetCapabilityAsset(ctx context.Context, tenantID, assetID string) (*assetentity.AssetRef, error)
}

// Clock abstracts time for tests.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

// CapabilityService is the domain façade for S5-G2.
type CapabilityService interface {
	ManualCreate(ctx context.Context, in *ManualCreateInput) (*entity.BusinessCapability, *entity.BusinessCapabilityRevision, error)
	DeriveRevision(ctx context.Context, in *DeriveInput) (*entity.BusinessCapabilityRevision, *entity.CapabilityDecision, error)
	StartAnalysis(ctx context.Context, in *StartAnalysisInput) (*AnalysisResult, error)
	GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*entity.CapabilityAnalysisRun, error)
	RetryFailedAnalysis(ctx context.Context, tenantID, analysisRunID, actorID string) (*AnalysisResult, error)
	ConfirmProposal(ctx context.Context, in *ConfirmInput) (*entity.BusinessCapabilityRevision, error)
	EditConfirmProposal(ctx context.Context, in *EditConfirmInput) (*entity.BusinessCapabilityRevision, error)
	RejectProposal(ctx context.Context, in *RejectInput) (*entity.CapabilityDecision, error)
	GetCapability(ctx context.Context, tenantID, capabilityID string) (*entity.BusinessCapability, error)
	ListCapabilities(ctx context.Context, tenantID, businessID string) ([]*entity.BusinessCapability, error)
	GetRevision(ctx context.Context, tenantID, revisionID string) (*entity.BusinessCapabilityRevision, error)
	ListRevisions(ctx context.Context, tenantID, capabilityID string) ([]*entity.BusinessCapabilityRevision, error)
	GetProposal(ctx context.Context, tenantID, proposalID string) (*entity.CapabilityProposal, error)
	ListValidations(ctx context.Context, tenantID, revisionID string) ([]*entity.CapabilityValidationResult, error)
	ListDecisions(ctx context.Context, tenantID, capabilityID string) ([]*entity.CapabilityDecision, error)
	Validate(ctx context.Context, tenantID, revisionID, actorID string) (*entity.BusinessCapabilityRevision, *entity.CapabilityValidationResult, error)
	Activate(ctx context.Context, tenantID, revisionID, actorID, reason string) (*entity.BusinessCapabilityRevision, error)
	MarkStale(ctx context.Context, tenantID, revisionID, actorID, reason string) (*entity.BusinessCapabilityRevision, error)
	Deprecate(ctx context.Context, tenantID, revisionID, actorID, reason string) (*entity.BusinessCapabilityRevision, error)
}

// Components wires CapabilityService dependencies.
type Components struct {
	// UoW is REQUIRED. Production: NewGormUnitOfWork(db). Tests: NewMemoryUnitOfWork().
	UoW       CapabilityUnitOfWork
	Generator ProposalGenerator
	Clock     Clock
	Business  BusinessModelValidationPort
	Contract  ContractValidationPort
}

type capabilityService struct {
	uow       CapabilityUnitOfWork
	generator ProposalGenerator
	clock     Clock
	business  BusinessModelValidationPort
	contract  ContractValidationPort
}

func NewCapabilityService(c *Components) CapabilityService {
	clk := Clock(realClock{})
	if c != nil && c.Clock != nil {
		clk = c.Clock
	}
	var uow CapabilityUnitOfWork
	var gen ProposalGenerator
	var business BusinessModelValidationPort
	var contract ContractValidationPort
	if c != nil {
		uow = c.UoW
		gen = c.Generator
		business = c.Business
		contract = c.Contract
	}
	return &capabilityService{uow: uow, generator: gen, clock: clk, business: business, contract: contract}
}

func (s *capabilityService) configured() bool {
	return s.uow != nil
}

func (s *capabilityService) portsConfigured() bool {
	return s.business != nil && s.contract != nil
}

func (s *capabilityService) root() repository.CapabilityRepository {
	return s.uow.Root()
}

func newID(prefix string) string {
	return prefix + "_" + uuid.NewString()
}

type ManualCreateInput struct {
	TenantID              string
	BusinessID            string
	CapabilityID          string // optional; allocated if empty
	ActorID               string
	OwnerID               int64
	Payload               entity.SemanticPayload
}

type DeriveInput struct {
	TenantID         string
	CapabilityID     string
	SourceRevisionID string
	ClientRequestID  string
	ActorID          string
	Reason           string
	Action           entity.DecisionAction // EDIT or DERIVE
	Payload          entity.SemanticPayload
}

type StartAnalysisInput struct {
	TenantID              string
	BusinessID            string
	BusinessModelRevision int32
	ClientRequestID       string
	ActorID               string
	Analysis              entity.AnalysisRequest
}

type AnalysisResult struct {
	Run          *entity.CapabilityAnalysisRun
	Proposals    []*entity.CapabilityProposal
	OwnedExecute bool
}

type ConfirmInput struct {
	TenantID        string
	ProposalID      string
	ActorID         string
	Reason          string
	ClientRequestID string
	CapabilityID    string // optional bind / first-create id
}

type EditConfirmInput struct {
	TenantID         string
	ProposalID       string
	ActorID          string
	Reason           string
	ClientRequestID  string
	CapabilityID     string
	EffectivePayload entity.SemanticPayload
}

type RejectInput struct {
	TenantID        string
	ProposalID      string
	ActorID         string
	Reason          string
	ClientRequestID string
}

func (s *capabilityService) now() time.Time {
	if s.clock == nil {
		return time.Now().UTC()
	}
	return s.clock.Now()
}

func (s *capabilityService) ManualCreate(ctx context.Context, in *ManualCreateInput) (*entity.BusinessCapability, *entity.BusinessCapabilityRevision, error) {
	if !s.configured() {
		return nil, nil, entity.ErrNotConfigured
	}
	if in == nil || strings.TrimSpace(in.TenantID) == "" || strings.TrimSpace(in.BusinessID) == "" || strings.TrimSpace(in.ActorID) == "" {
		return nil, nil, entity.ErrInvalidPayload
	}
	if err := ValidateMaterializationPayload(in.Payload); err != nil {
		return nil, nil, err
	}
	parsedOwner, err := parseOwnerID(in.ActorID)
	if err != nil {
		return nil, nil, err
	}
	ownerID := in.OwnerID
	if ownerID == 0 {
		ownerID = parsedOwner
	}
	capID := strings.TrimSpace(in.CapabilityID)
	if capID == "" {
		capID = newID("cap")
	}
	now := s.now()
	var outCap *entity.BusinessCapability
	var outRev *entity.BusinessCapabilityRevision
	err = s.uow.WithinTransaction(ctx, func(tx CapabilityTx) error {
		cap := &entity.BusinessCapability{
			CapabilityID: capID, TenantID: in.TenantID, BusinessID: in.BusinessID,
			AggregateGeneration: 0, CreatedBy: in.ActorID, CreatedAt: now, UpdatedAt: now,
		}
		if err := tx.Repo().CreateCapability(ctx, cap); err != nil {
			return err
		}
		digest, err := DecisionPayloadDigest(in.Payload)
		if err != nil {
			return err
		}
		revID := newID("crev")
		rev := revisionFromPayload(in.TenantID, in.BusinessID, capID, revID, 1, entity.SourceManualCreated, in.Payload, in.ActorID, now)
		proj, err := ProjectCapabilityAssetRef(cap, []*entity.BusinessCapabilityRevision{rev})
		if err != nil {
			return err
		}
		asset := &assetentity.AssetRef{
			TenantID: in.TenantID, AssetID: capID, Kind: assetentity.AssetKindCapability,
			Name: proj.Name, SemanticVersion: proj.SemanticVersion, Revision: 1, SchemaVersion: "1.0",
			Status: proj.Status, OwnerID: ownerID, CreatedBy: ownerID,
			ContentDigest: proj.ContentDigest, CreatedAt: now, UpdatedAt: now,
		}
		if err := tx.Assets().CreateCapabilityAsset(ctx, asset); err != nil {
			return err
		}
		if err := tx.Repo().CreateRevision(ctx, rev); err != nil {
			return err
		}
		dec := &entity.CapabilityDecision{
			DecisionID: newID("cdec"), TenantID: in.TenantID, BusinessID: in.BusinessID,
			CapabilityID: capID, TargetRevisionID: revID, Action: entity.DecisionCreate,
			PayloadDigest: digest, ActorPrincipalID: in.ActorID, CreatedAt: now,
		}
		if err := tx.Repo().CreateDecision(ctx, dec); err != nil {
			return err
		}
		if err := projectAndUpdateAssets(ctx, tx.Assets(), cap, []*entity.BusinessCapabilityRevision{rev}); err != nil {
			return err
		}
		outCap = cap
		outRev = rev
		return nil
	})
	return outCap, outRev, err
}

func (s *capabilityService) DeriveRevision(ctx context.Context, in *DeriveInput) (*entity.BusinessCapabilityRevision, *entity.CapabilityDecision, error) {
	if !s.configured() {
		return nil, nil, entity.ErrNotConfigured
	}
	if in == nil || in.TenantID == "" || in.CapabilityID == "" || in.SourceRevisionID == "" || in.ClientRequestID == "" || in.ActorID == "" {
		return nil, nil, entity.ErrInvalidPayload
	}
	action := in.Action
	if action == "" {
		action = entity.DecisionDerive
	}
	if action != entity.DecisionEdit && action != entity.DecisionDerive {
		return nil, nil, entity.ErrInvalidPayload
	}
	if err := ValidateMaterializationPayload(in.Payload); err != nil {
		return nil, nil, err
	}
	digest, err := DecisionPayloadDigest(in.Payload)
	if err != nil {
		return nil, nil, err
	}
	if existing, err := s.root().GetDecisionByDeriveKey(ctx, in.TenantID, in.CapabilityID, in.SourceRevisionID, in.ClientRequestID); err == nil {
		if existing.PayloadDigest != digest {
			return nil, nil, entity.ErrIdempotencyConflict
		}
		rev, err := s.root().GetRevision(ctx, in.TenantID, existing.TargetRevisionID)
		return rev, existing, err
	} else if !errors.Is(err, entity.ErrDecisionNotFound) {
		return nil, nil, err
	}

	var outRev *entity.BusinessCapabilityRevision
	var outDec *entity.CapabilityDecision
	err = s.uow.WithinTransaction(ctx, func(tx CapabilityTx) error {
		cap, err := tx.Repo().GetCapabilityForUpdate(ctx, in.TenantID, in.CapabilityID)
		if err != nil {
			return err
		}
		src, err := tx.Repo().GetRevisionForUpdate(ctx, in.TenantID, in.SourceRevisionID)
		if err != nil {
			return err
		}
		if src.CapabilityID != in.CapabilityID || src.TenantID != in.TenantID {
			return entity.ErrCrossTenant
		}
		if existing, err := tx.Repo().GetDecisionByDeriveKey(ctx, in.TenantID, in.CapabilityID, in.SourceRevisionID, in.ClientRequestID); err == nil {
			if existing.PayloadDigest != digest {
				return entity.ErrIdempotencyConflict
			}
			rev, err := tx.Repo().GetRevision(ctx, in.TenantID, existing.TargetRevisionID)
			if err != nil {
				return err
			}
			outRev, outDec = rev, existing
			return nil
		} else if !errors.Is(err, entity.ErrDecisionNotFound) {
			return err
		}
		max, err := tx.Repo().MaxVersion(ctx, in.TenantID, in.CapabilityID)
		if err != nil {
			return err
		}
		now := s.now()
		revID := newID("crev")
		rev := revisionFromPayload(in.TenantID, cap.BusinessID, in.CapabilityID, revID, max+1, entity.SourceDerivedEdit, in.Payload, in.ActorID, now)
		rev.DerivedFromRevisionID = in.SourceRevisionID
		if err := tx.Repo().CreateRevision(ctx, rev); err != nil {
			return err
		}
		dec := &entity.CapabilityDecision{
			DecisionID: newID("cdec"), TenantID: in.TenantID, BusinessID: cap.BusinessID,
			CapabilityID: in.CapabilityID, SourceRevisionID: in.SourceRevisionID, TargetRevisionID: revID,
			Action: action, PayloadDigest: digest, ClientRequestID: in.ClientRequestID,
			ActorPrincipalID: in.ActorID, Reason: in.Reason, CreatedAt: now,
		}
		if err := tx.Repo().CreateDecision(ctx, dec); err != nil {
			return err
		}
		revs, err := tx.Repo().ListRevisions(ctx, in.TenantID, in.CapabilityID)
		if err != nil {
			return err
		}
		if err := projectAndUpdateAssets(ctx, tx.Assets(), cap, revs); err != nil {
			return err
		}
		outRev, outDec = rev, dec
		return nil
	})
	return outRev, outDec, err
}

func (s *capabilityService) StartAnalysis(ctx context.Context, in *StartAnalysisInput) (*AnalysisResult, error) {
	if !s.configured() || s.generator == nil {
		return nil, entity.ErrNotConfigured
	}
	if in == nil || in.TenantID == "" || in.BusinessID == "" || in.ClientRequestID == "" || in.BusinessModelRevision <= 0 {
		return nil, entity.ErrInvalidPayload
	}
	if strings.TrimSpace(in.ActorID) == "" {
		return nil, entity.ErrInvalidPayload
	}
	analysis := in.Analysis
	analysis.BusinessModelRevision = in.BusinessModelRevision
	if err := ValidateAnalysisRequest(analysis); err != nil {
		return nil, err
	}
	digest, err := AnalysisRequestDigest(analysis)
	if err != nil {
		return nil, err
	}
	reqJSON, err := json.Marshal(analysis)
	if err != nil {
		return nil, entity.ErrInvalidPayload
	}
	now := s.now()
	run := &entity.CapabilityAnalysisRun{
		AnalysisRunID: newID("carun"), TenantID: in.TenantID, BusinessID: in.BusinessID,
		BusinessModelRevision: in.BusinessModelRevision, ClientRequestID: in.ClientRequestID,
		RequestDigest: digest, Status: entity.AnalysisPending, Attempt: 1,
		RequestJSON: string(reqJSON), CreatedBy: in.ActorID, CreatedAt: now, UpdatedAt: now,
	}
	seedExecutionLease(run, now)
	var existing *entity.CapabilityAnalysisRun
	var created bool
	err = s.uow.WithinTransaction(ctx, func(tx CapabilityTx) error {
		var cerr error
		existing, created, cerr = tx.Repo().CreateOrClaimAnalysisRun(ctx, run)
		if cerr != nil {
			return MapRepoError(cerr)
		}
		if created {
			att := &entity.CapabilityAnalysisAttempt{
				AttemptID: newID("caatt"), AnalysisRunID: existing.AnalysisRunID, TenantID: existing.TenantID,
				Attempt: existing.Attempt, ActorPrincipalID: in.ActorID,
				TriggerKind: entity.AttemptTriggerFirst, ResultStatus: entity.AttemptResultPending, CreatedAt: now,
			}
			return MapRepoError(tx.Repo().CreateAnalysisAttempt(ctx, att))
		}
		return nil
	})
	if err != nil {
		return nil, MapRepoError(err)
	}
	if !created {
		return s.handleExistingAnalysis(ctx, existing, analysis, in.ActorID)
	}
	return s.executeAnalysis(ctx, existing, analysis)
}

func (s *capabilityService) handleExistingAnalysis(ctx context.Context, existing *entity.CapabilityAnalysisRun, analysis entity.AnalysisRequest, actorID string) (*AnalysisResult, error) {
	now := s.now()
	if existing.Status == entity.AnalysisPending && analysisLeaseExpired(existing, now) {
		priorAttempt := existing.Attempt
		var claimed *entity.CapabilityAnalysisRun
		var owned bool
		err := s.uow.WithinTransaction(ctx, func(tx CapabilityTx) error {
			var cerr error
			claimed, owned, cerr = tx.Repo().ClaimExpiredPendingExecution(ctx, existing.TenantID, existing.AnalysisRunID, existing.Attempt, now)
			if cerr != nil {
				return MapRepoError(cerr)
			}
			if owned {
				if serr := tx.Repo().SupersedeAnalysisAttempt(ctx, claimed.TenantID, claimed.AnalysisRunID, priorAttempt); serr != nil {
					return MapRepoError(serr)
				}
				att := &entity.CapabilityAnalysisAttempt{
					AttemptID: newID("caatt"), AnalysisRunID: claimed.AnalysisRunID, TenantID: claimed.TenantID,
					Attempt: claimed.Attempt, ActorPrincipalID: actorID,
					TriggerKind: entity.AttemptTriggerLeaseTakeover, ResultStatus: entity.AttemptResultPending, CreatedAt: now,
				}
				return MapRepoError(tx.Repo().CreateAnalysisAttempt(ctx, att))
			}
			return nil
		})
		if err != nil {
			return nil, MapRepoError(err)
		}
		if owned {
			persisted, loadErr := s.loadValidatedPersistedAnalysisRequest(ctx, claimed)
			if loadErr != nil {
				return nil, loadErr
			}
			return s.executeAnalysis(ctx, claimed, persisted) // NOT caller analysis
		}
		existing = claimed
	}
	props, err := s.root().ListProposalsByAnalysisRun(ctx, existing.TenantID, existing.AnalysisRunID)
	if err != nil {
		return nil, MapRepoError(err)
	}
	if existing == nil {
		return nil, entity.ErrConsistency
	}
	return &AnalysisResult{Run: existing, Proposals: props, OwnedExecute: false}, nil
}

// loadValidatedPersistedAnalysisRequest unmarshals and validates RequestJSON; marks failed on any error.
// Never calls the generator.
func (s *capabilityService) loadValidatedPersistedAnalysisRequest(ctx context.Context, run *entity.CapabilityAnalysisRun) (entity.AnalysisRequest, error) {
	var empty entity.AnalysisRequest
	if run == nil {
		return empty, entity.ErrConsistency
	}
	fail := func(code string, retErr error) (entity.AnalysisRequest, error) {
		failErr := s.uow.WithinTransaction(ctx, func(tx CapabilityTx) error {
			if markErr := tx.Repo().MarkAnalysisFailed(ctx, run.TenantID, run.AnalysisRunID, code, run.Attempt); markErr != nil {
				return MapRepoError(markErr)
			}
			if attErr := tx.Repo().CompleteAnalysisAttempt(ctx, run.TenantID, run.AnalysisRunID, run.Attempt, entity.AttemptResultFailed, code); attErr != nil {
				return MapRepoError(attErr)
			}
			return nil
		})
		if failErr != nil {
			return empty, MapRepoError(failErr)
		}
		return empty, retErr
	}
	var req entity.AnalysisRequest
	if err := json.Unmarshal([]byte(run.RequestJSON), &req); err != nil {
		return fail("FORMA_CAPABILITY_INVALID_REQUEST", entity.ErrConsistency)
	}
	if err := ValidateAnalysisRequest(req); err != nil {
		return fail("FORMA_CAPABILITY_INVALID_REQUEST", entity.ErrInvalidPayload)
	}
	digest, err := AnalysisRequestDigest(req)
	if err != nil || digest != run.RequestDigest {
		return fail("FORMA_CAPABILITY_INVALID_REQUEST", entity.ErrConsistency)
	}
	return req, nil
}

func (s *capabilityService) executeAnalysis(ctx context.Context, run *entity.CapabilityAnalysisRun, analysis entity.AnalysisRequest) (*AnalysisResult, error) {
	if run == nil {
		return nil, entity.ErrConsistency
	}
	attempt := run.Attempt
	res, err := s.generator.Generate(ctx, GenerateRequest{
		TenantID: run.TenantID, BusinessID: run.BusinessID, BusinessModelRevision: run.BusinessModelRevision,
		AnalysisRunID: run.AnalysisRunID, Analysis: analysis,
	})
	if err != nil {
		code := SanitizedAnalysisErrorCode(err)
		if markErr := s.markAnalysisFailedWithAttempt(ctx, run.TenantID, run.AnalysisRunID, code, attempt); markErr != nil {
			return nil, SanitizeAnalysisError(markErr)
		}
		failed, getErr := s.root().GetAnalysisRun(ctx, run.TenantID, run.AnalysisRunID)
		if getErr != nil {
			return nil, SanitizeAnalysisError(getErr)
		}
		return &AnalysisResult{Run: failed, OwnedExecute: true}, entity.ErrAnalysisFailed
	}
	now := s.now()
	var created []*entity.CapabilityProposal
	err = s.uow.WithinTransaction(ctx, func(tx CapabilityTx) error {
		for _, payload := range res.Proposals {
			if err := ValidateMaterializationPayload(payload); err != nil {
				return err
			}
			p := &entity.CapabilityProposal{
				ProposalID: newID("cprop"), TenantID: run.TenantID, BusinessID: run.BusinessID,
				AnalysisRunID: run.AnalysisRunID, Status: entity.ProposalProposed,
				Payload: payload, CreatedAt: now,
			}
			if err := tx.Repo().CreateProposal(ctx, p); err != nil {
				return MapRepoError(err)
			}
			created = append(created, p)
		}
		if err := tx.Repo().MarkAnalysisSucceeded(ctx, run.TenantID, run.AnalysisRunID, res.ModelRef, attempt); err != nil {
			return MapRepoError(err)
		}
		return MapRepoError(tx.Repo().CompleteAnalysisAttempt(ctx, run.TenantID, run.AnalysisRunID, attempt, entity.AttemptResultSucceeded, ""))
	})
	if err != nil {
		if errors.Is(err, entity.ErrStaleGeneration) {
			final, getErr := s.root().GetAnalysisRun(ctx, run.TenantID, run.AnalysisRunID)
			if getErr != nil {
				return nil, MapRepoError(getErr)
			}
			props, listErr := s.root().ListProposalsByAnalysisRun(ctx, run.TenantID, run.AnalysisRunID)
			if listErr != nil {
				return nil, MapRepoError(listErr)
			}
			return &AnalysisResult{Run: final, Proposals: props, OwnedExecute: true}, entity.ErrStaleGeneration
		}
		code := SanitizedAnalysisErrorCode(err)
		if markErr := s.markAnalysisFailedWithAttempt(ctx, run.TenantID, run.AnalysisRunID, code, attempt); markErr != nil {
			return nil, SanitizeAnalysisError(markErr)
		}
		failed, getErr := s.root().GetAnalysisRun(ctx, run.TenantID, run.AnalysisRunID)
		if getErr != nil {
			return nil, SanitizeAnalysisError(getErr)
		}
		return &AnalysisResult{Run: failed, OwnedExecute: true}, SanitizeAnalysisError(err)
	}
	final, getErr := s.root().GetAnalysisRun(ctx, run.TenantID, run.AnalysisRunID)
	if getErr != nil {
		return nil, MapRepoError(getErr)
	}
	return &AnalysisResult{Run: final, Proposals: created, OwnedExecute: true}, nil
}

func (s *capabilityService) markAnalysisFailedWithAttempt(ctx context.Context, tenantID, analysisRunID, errorCode string, attempt int32) error {
	return s.uow.WithinTransaction(ctx, func(tx CapabilityTx) error {
		if err := tx.Repo().MarkAnalysisFailed(ctx, tenantID, analysisRunID, errorCode, attempt); err != nil {
			return MapRepoError(err)
		}
		return MapRepoError(tx.Repo().CompleteAnalysisAttempt(ctx, tenantID, analysisRunID, attempt, entity.AttemptResultFailed, errorCode))
	})
}

func (s *capabilityService) GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*entity.CapabilityAnalysisRun, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	run, err := s.root().GetAnalysisRun(ctx, tenantID, analysisRunID)
	return run, MapRepoError(err)
}

func (s *capabilityService) RetryFailedAnalysis(ctx context.Context, tenantID, analysisRunID, actorID string) (*AnalysisResult, error) {
	if !s.configured() || s.generator == nil {
		return nil, entity.ErrNotConfigured
	}
	if strings.TrimSpace(actorID) == "" {
		return nil, entity.ErrInvalidPayload
	}
	run, err := s.root().GetAnalysisRun(ctx, tenantID, analysisRunID)
	if err != nil {
		return nil, MapRepoError(err)
	}
	if run.Status != entity.AnalysisFailed {
		return nil, entity.ErrAnalysisNotFailed
	}
	var attempt int32
	var ok bool
	now := s.now()
	err = s.uow.WithinTransaction(ctx, func(tx CapabilityTx) error {
		var cerr error
		ok, attempt, cerr = tx.Repo().ClaimAnalysisRetry(ctx, tenantID, analysisRunID, actorID)
		if cerr != nil {
			return MapRepoError(cerr)
		}
		if !ok {
			return entity.ErrAnalysisNotFailed
		}
		att := &entity.CapabilityAnalysisAttempt{
			AttemptID: newID("caatt"), AnalysisRunID: analysisRunID, TenantID: tenantID,
			Attempt: attempt, ActorPrincipalID: actorID,
			TriggerKind: entity.AttemptTriggerRetry, ResultStatus: entity.AttemptResultPending, CreatedAt: now,
		}
		return MapRepoError(tx.Repo().CreateAnalysisAttempt(ctx, att))
	})
	if err != nil {
		return nil, MapRepoError(err)
	}
	run, err = s.root().GetAnalysisRun(ctx, tenantID, analysisRunID)
	if err != nil {
		return nil, MapRepoError(err)
	}
	if run == nil {
		return nil, entity.ErrConsistency
	}
	run.Attempt = attempt
	analysis, loadErr := s.loadValidatedPersistedAnalysisRequest(ctx, run)
	if loadErr != nil {
		return nil, loadErr
	}
	return s.executeAnalysis(ctx, run, analysis)
}

func (s *capabilityService) ConfirmProposal(ctx context.Context, in *ConfirmInput) (*entity.BusinessCapabilityRevision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	if in == nil || in.TenantID == "" || in.ProposalID == "" || in.ActorID == "" {
		return nil, entity.ErrInvalidPayload
	}
	return s.materializeProposal(ctx, in.TenantID, in.ProposalID, in.ActorID, in.Reason, in.ClientRequestID, in.CapabilityID, entity.DecisionConfirm, nil)
}

func (s *capabilityService) EditConfirmProposal(ctx context.Context, in *EditConfirmInput) (*entity.BusinessCapabilityRevision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	if in == nil || in.TenantID == "" || in.ProposalID == "" || in.ActorID == "" {
		return nil, entity.ErrInvalidPayload
	}
	if err := ValidateMaterializationPayload(in.EffectivePayload); err != nil {
		return nil, err
	}
	return s.materializeProposal(ctx, in.TenantID, in.ProposalID, in.ActorID, in.Reason, in.ClientRequestID, in.CapabilityID, entity.DecisionEditConfirm, &in.EffectivePayload)
}

func assertProposalCapabilityBinding(prop *entity.CapabilityProposal, capabilityID string) error {
	if prop == nil {
		return entity.ErrConsistency
	}
	caller := strings.TrimSpace(capabilityID)
	if prop.CapabilityID != "" && caller != "" && caller != prop.CapabilityID {
		return entity.ErrConflict
	}
	return nil
}

func (s *capabilityService) materializeProposal(ctx context.Context, tenantID, proposalID, actorID, reason, clientRequestID, capabilityID string, action entity.DecisionAction, editPayload *entity.SemanticPayload) (*entity.BusinessCapabilityRevision, error) {
	var out *entity.BusinessCapabilityRevision
	err := s.uow.WithinTransaction(ctx, func(tx CapabilityTx) error {
		prop, err := tx.Repo().GetProposalForUpdate(ctx, tenantID, proposalID)
		if err != nil {
			return MapRepoError(err)
		}
		// Binding check BEFORE status switch — covers CONFIRMED / EDIT_CONFIRMED replay mismatch.
		if err := assertProposalCapabilityBinding(prop, capabilityID); err != nil {
			return err
		}
		var effective entity.SemanticPayload
		switch action {
		case entity.DecisionConfirm:
			effective = prop.Payload
		case entity.DecisionEditConfirm:
			effective = *editPayload
		default:
			return entity.ErrInvalidPayload
		}
		digest, err := DecisionPayloadDigest(effective)
		if err != nil {
			return err
		}
		switch prop.Status {
		case entity.ProposalConfirmed:
			if action != entity.DecisionConfirm {
				return entity.ErrConflict
			}
			return s.replayTerminal(ctx, tx, prop, digest, &out)
		case entity.ProposalEditConfirmed:
			if action != entity.DecisionEditConfirm {
				return entity.ErrConflict
			}
			return s.replayTerminal(ctx, tx, prop, digest, &out)
		case entity.ProposalRejected:
			return entity.ErrConflict
		case entity.ProposalProposed:
			// first materialization
		default:
			return entity.ErrConflict
		}

		if prop.CapabilityID != "" {
			capabilityID = prop.CapabilityID
		}

		capID := strings.TrimSpace(capabilityID)
		if capID == "" {
			capID = strings.TrimSpace(prop.CapabilityID)
		}
		now := s.now()
		terminalStatus := entity.ProposalConfirmed
		if action == entity.DecisionEditConfirm {
			terminalStatus = entity.ProposalEditConfirmed
		}

		if err := ValidateMaterializationPayload(effective); err != nil {
			return err
		}

		if capID != "" {
			if _, err := tx.Repo().GetCapability(ctx, tenantID, capID); err == nil {
				return s.materializeExisting(ctx, tx, prop, capID, effective, digest, action, terminalStatus, actorID, reason, clientRequestID, now, &out)
			} else if !errors.Is(err, entity.ErrNotFound) {
				return err
			}
		}
		if capID == "" {
			capID = newID("cap")
		}
		return s.materializeFirstCreate(ctx, tx, prop, capID, effective, digest, action, terminalStatus, actorID, reason, clientRequestID, now, &out)
	})
	return out, err
}

func (s *capabilityService) replayTerminal(ctx context.Context, tx CapabilityTx, prop *entity.CapabilityProposal, digest string, out **entity.BusinessCapabilityRevision) error {
	dec, err := tx.Repo().GetDecisionByProposal(ctx, prop.TenantID, prop.ProposalID)
	if err != nil {
		return entity.ErrConsistency
	}
	if dec.PayloadDigest != digest {
		return entity.ErrIdempotencyConflict
	}
	if prop.MaterializedRevisionID == "" {
		return entity.ErrConsistency
	}
	rev, err := tx.Repo().GetRevision(ctx, prop.TenantID, prop.MaterializedRevisionID)
	if err != nil {
		return entity.ErrConsistency
	}
	*out = rev
	return nil
}

func (s *capabilityService) materializeExisting(ctx context.Context, tx CapabilityTx, prop *entity.CapabilityProposal, capID string, effective entity.SemanticPayload, digest string, action entity.DecisionAction, terminal entity.ProposalStatus, actorID, reason, clientRequestID string, now time.Time, out **entity.BusinessCapabilityRevision) error {
	cap, err := tx.Repo().GetCapabilityForUpdate(ctx, prop.TenantID, capID)
	if err != nil {
		return err
	}
	if cap.BusinessID != prop.BusinessID {
		return entity.ErrCrossTenant
	}
	max, err := tx.Repo().MaxVersion(ctx, prop.TenantID, capID)
	if err != nil {
		return err
	}
	revID := newID("crev")
	rev := revisionFromPayload(prop.TenantID, prop.BusinessID, capID, revID, max+1, entity.SourceAIProposal, effective, actorID, now)
	rev.AnalysisRunID = prop.AnalysisRunID
	rev.ProposalID = prop.ProposalID
	if err := tx.Repo().CreateRevision(ctx, rev); err != nil {
		return err
	}
	dec := &entity.CapabilityDecision{
		DecisionID: newID("cdec"), TenantID: prop.TenantID, BusinessID: prop.BusinessID,
		CapabilityID: capID, ProposalID: prop.ProposalID, TargetRevisionID: revID,
		Action: action, PayloadDigest: digest, ClientRequestID: clientRequestID,
		ActorPrincipalID: actorID, Reason: reason, CreatedAt: now,
	}
	if err := tx.Repo().CreateDecision(ctx, dec); err != nil {
		return err
	}
	if err := tx.Repo().TerminalizeProposal(ctx, prop.TenantID, prop.ProposalID, terminal, revID, capID); err != nil {
		return err
	}
	revs, err := tx.Repo().ListRevisions(ctx, prop.TenantID, capID)
	if err != nil {
		return err
	}
	if err := projectAndUpdateAssets(ctx, tx.Assets(), cap, revs); err != nil {
		return err
	}
	*out = rev
	return nil
}

func (s *capabilityService) materializeFirstCreate(ctx context.Context, tx CapabilityTx, prop *entity.CapabilityProposal, capID string, effective entity.SemanticPayload, digest string, action entity.DecisionAction, terminal entity.ProposalStatus, actorID, reason, clientRequestID string, now time.Time, out **entity.BusinessCapabilityRevision) error {
	ownerID, err := parseOwnerID(actorID)
	if err != nil {
		return err
	}
	cap := &entity.BusinessCapability{
		CapabilityID: capID, TenantID: prop.TenantID, BusinessID: prop.BusinessID,
		AggregateGeneration: 0, CreatedBy: actorID, CreatedAt: now, UpdatedAt: now,
	}
	if err := tx.Repo().CreateCapability(ctx, cap); err != nil {
		return err
	}
	revID := newID("crev")
	rev := revisionFromPayload(prop.TenantID, prop.BusinessID, capID, revID, 1, entity.SourceAIProposal, effective, actorID, now)
	rev.AnalysisRunID = prop.AnalysisRunID
	rev.ProposalID = prop.ProposalID
	proj, err := ProjectCapabilityAssetRef(cap, []*entity.BusinessCapabilityRevision{rev})
	if err != nil {
		return err
	}
	asset := &assetentity.AssetRef{
		TenantID: prop.TenantID, AssetID: capID, Kind: assetentity.AssetKindCapability,
		Name: proj.Name, SemanticVersion: proj.SemanticVersion, Revision: 1, SchemaVersion: "1.0",
		Status: proj.Status, OwnerID: ownerID, CreatedBy: ownerID, ContentDigest: proj.ContentDigest, CreatedAt: now, UpdatedAt: now,
	}
	if err := tx.Assets().CreateCapabilityAsset(ctx, asset); err != nil {
		return err
	}
	if err := tx.Repo().CreateRevision(ctx, rev); err != nil {
		return err
	}
	dec := &entity.CapabilityDecision{
		DecisionID: newID("cdec"), TenantID: prop.TenantID, BusinessID: prop.BusinessID,
		CapabilityID: capID, ProposalID: prop.ProposalID, TargetRevisionID: revID,
		Action: action, PayloadDigest: digest, ClientRequestID: clientRequestID,
		ActorPrincipalID: actorID, Reason: reason, CreatedAt: now,
	}
	if err := tx.Repo().CreateDecision(ctx, dec); err != nil {
		return err
	}
	if err := tx.Repo().TerminalizeProposal(ctx, prop.TenantID, prop.ProposalID, terminal, revID, capID); err != nil {
		return err
	}
	if err := projectAndUpdateAssets(ctx, tx.Assets(), cap, []*entity.BusinessCapabilityRevision{rev}); err != nil {
		return err
	}
	*out = rev
	return nil
}

func (s *capabilityService) RejectProposal(ctx context.Context, in *RejectInput) (*entity.CapabilityDecision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	if in == nil || in.TenantID == "" || in.ProposalID == "" || in.ActorID == "" {
		return nil, entity.ErrInvalidPayload
	}
	var out *entity.CapabilityDecision
	err := s.uow.WithinTransaction(ctx, func(tx CapabilityTx) error {
		prop, err := tx.Repo().GetProposalForUpdate(ctx, in.TenantID, in.ProposalID)
		if err != nil {
			return err
		}
		switch prop.Status {
		case entity.ProposalRejected:
			dec, err := tx.Repo().GetDecisionByProposal(ctx, in.TenantID, in.ProposalID)
			if err != nil {
				return entity.ErrConsistency
			}
			out = dec
			return nil
		case entity.ProposalConfirmed, entity.ProposalEditConfirmed:
			return entity.ErrConflict
		case entity.ProposalProposed:
		default:
			return entity.ErrConflict
		}
		now := s.now()
		dec := &entity.CapabilityDecision{
			DecisionID: newID("cdec"), TenantID: prop.TenantID, BusinessID: prop.BusinessID,
			CapabilityID: prop.CapabilityID, ProposalID: prop.ProposalID,
			Action: entity.DecisionReject, ActorPrincipalID: in.ActorID, Reason: in.Reason,
			ClientRequestID: in.ClientRequestID, CreatedAt: now,
		}
		if err := tx.Repo().CreateDecision(ctx, dec); err != nil {
			return err
		}
		if err := tx.Repo().TerminalizeProposal(ctx, in.TenantID, in.ProposalID, entity.ProposalRejected, "", prop.CapabilityID); err != nil {
			return err
		}
		out = dec
		return nil
	})
	return out, err
}

func (s *capabilityService) GetCapability(ctx context.Context, tenantID, capabilityID string) (*entity.BusinessCapability, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	return s.root().GetCapability(ctx, tenantID, capabilityID)
}

func (s *capabilityService) ListCapabilities(ctx context.Context, tenantID, businessID string) ([]*entity.BusinessCapability, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	if strings.TrimSpace(tenantID) == "" || strings.TrimSpace(businessID) == "" {
		return nil, entity.ErrInvalidPayload
	}
	return s.root().ListCapabilitiesByBusiness(ctx, tenantID, businessID)
}

func (s *capabilityService) GetRevision(ctx context.Context, tenantID, revisionID string) (*entity.BusinessCapabilityRevision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	return s.root().GetRevision(ctx, tenantID, revisionID)
}

func (s *capabilityService) ListRevisions(ctx context.Context, tenantID, capabilityID string) ([]*entity.BusinessCapabilityRevision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	return s.root().ListRevisions(ctx, tenantID, capabilityID)
}

func (s *capabilityService) GetProposal(ctx context.Context, tenantID, proposalID string) (*entity.CapabilityProposal, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	return s.root().GetProposal(ctx, tenantID, proposalID)
}

func (s *capabilityService) ListValidations(ctx context.Context, tenantID, revisionID string) ([]*entity.CapabilityValidationResult, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	return s.root().ListValidationsByRevision(ctx, tenantID, revisionID)
}

func (s *capabilityService) ListDecisions(ctx context.Context, tenantID, capabilityID string) ([]*entity.CapabilityDecision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	return s.root().ListDecisionsByCapability(ctx, tenantID, capabilityID)
}

func (s *capabilityService) Validate(ctx context.Context, tenantID, revisionID, actorID string) (*entity.BusinessCapabilityRevision, *entity.CapabilityValidationResult, error) {
	if !s.configured() {
		return nil, nil, entity.ErrNotConfigured
	}
	if strings.TrimSpace(actorID) == "" || strings.TrimSpace(tenantID) == "" || strings.TrimSpace(revisionID) == "" {
		return nil, nil, entity.ErrInvalidPayload
	}
	if !s.portsConfigured() {
		return nil, nil, entity.ErrPortsNotConfigured
	}

	// Existence / early deny peek only — digests rebuilt under lock from locked revision.
	if _, err := s.root().GetRevision(ctx, tenantID, revisionID); err != nil {
		return nil, nil, err
	}

	var outRev *entity.BusinessCapabilityRevision
	var outResult *entity.CapabilityValidationResult
	err := s.uow.WithinTransaction(ctx, func(tx CapabilityTx) error {
		cap, rev, err := s.lockCapAndRev(ctx, tx, tenantID, revisionID)
		if err != nil {
			return err
		}
		if cap.TenantID != tenantID || rev.TenantID != tenantID {
			return entity.ErrCrossTenant
		}
		if rev.CapabilityID != cap.CapabilityID || rev.BusinessID != cap.BusinessID {
			return entity.ErrConsistency
		}

		ev, err := s.buildValidationEvidence(ctx, rev)
		if err != nil {
			return err
		}

		if rev.Status == entity.RevisionValidated {
			existing, getErr := tx.Repo().GetValidationByEvidence(ctx, tenantID, revisionID, ev.evidenceDigest)
			if getErr == nil && existing.Status == entity.ValidationPass &&
				existing.RevisionContentDigest == ev.revDigest {
				// Replay path: fence before returning so BM/Contract drift cannot silently reuse PASS.
				if _, fenceErr := s.requireFinalEvidenceFence(ctx, rev, ev); fenceErr != nil {
					return fenceErr
				}
				outRev = rev
				outResult = existing
				return nil
			}
			return entity.ErrMissingValidationEvidence
		}
		if rev.Status != entity.RevisionDraft {
			return entity.ErrIllegalTransition
		}
		if !AllowTransition(entity.RevisionDraft, entity.RevisionValidated) {
			return entity.ErrIllegalTransition
		}
		if err := s.requireValidateProvenance(ctx, tx, rev); err != nil {
			return err
		}
		if err := ValidateMaterializationPayload(rev.ToSemanticPayload()); err != nil {
			return err
		}

		issues := validateCompatibility(rev, ev.bmEvidence, ev.descriptors)
		// Final fence before any ValidationResult create or revision status write.
		// Capability locks do not cover Business/Contract aggregates.
		ev, err = s.requireFinalEvidenceFence(ctx, rev, ev)
		if err != nil {
			return err
		}
		now := s.now()
		result := &entity.CapabilityValidationResult{
			ValidationID:               newID("cval"),
			TenantID:                   tenantID,
			BusinessID:                 rev.BusinessID,
			CapabilityID:               rev.CapabilityID,
			RevisionID:                 revisionID,
			RevisionContentDigest:      ev.revDigest,
			BusinessModelRevision:      rev.BusinessModelRevision,
			BusinessModelContentDigest: ev.bmDigest,
			ContractEvidenceDigest:     ev.contractDigest,
			EvidenceDigest:             ev.evidenceDigest,
			IssueCodes:                 issues,
			ValidatedBy:                actorID,
			ValidatedAt:                now,
			CreatedAt:                  now,
		}
		if len(issues) > 0 {
			result.Status = entity.ValidationFail
			created, createErr := s.createOrGetValidation(ctx, tx, result)
			if createErr != nil {
				return createErr
			}
			outResult = created
			outRev = rev
			// Commit FAIL rows — never return ErrValidationFailed from the txn callback.
			return nil
		}
		result.Status = entity.ValidationPass
		created, createErr := s.createOrGetValidation(ctx, tx, result)
		if createErr != nil {
			return createErr
		}
		ok, err := tx.Repo().UpdateRevisionStatus(ctx, tenantID, revisionID, entity.RevisionDraft, entity.RevisionValidated)
		if err != nil || !ok {
			return entity.ErrIllegalTransition
		}
		rev.Status = entity.RevisionValidated
		revs, err := tx.Repo().ListRevisions(ctx, tenantID, cap.CapabilityID)
		if err != nil {
			return err
		}
		if err := projectAndUpdateAssets(ctx, tx.Assets(), cap, revs); err != nil {
			return err
		}
		outRev = rev
		outResult = created
		return nil
	})
	if err != nil {
		return outRev, outResult, err
	}
	if outResult != nil && outResult.Status == entity.ValidationFail {
		return outRev, outResult, entity.ErrValidationFailed
	}
	return outRev, outResult, nil
}

func (s *capabilityService) createOrGetValidation(ctx context.Context, tx CapabilityTx, result *entity.CapabilityValidationResult) (*entity.CapabilityValidationResult, error) {
	err := tx.Repo().CreateValidationResult(ctx, result)
	if err == nil {
		return result, nil
	}
	if !errors.Is(err, entity.ErrConflict) {
		return nil, err
	}
	existing, getErr := tx.Repo().GetValidationByEvidence(ctx, result.TenantID, result.RevisionID, result.EvidenceDigest)
	if getErr != nil {
		return nil, entity.ErrConflict
	}
	return existing, nil
}

type evidenceBundle struct {
	bmEvidence     *BusinessModelRevisionEvidence
	descriptors    map[string]*ContractLogicalDescriptor
	revDigest      string
	contractDigest string
	bmDigest       string
	evidenceDigest string
}

// buildValidationEvidence fetches BM/contract ports and digests from the locked revision.
// Shared by Validate and Activate so final writes/gates never rely solely on a pre-UoW snapshot.
func (s *capabilityService) buildValidationEvidence(ctx context.Context, rev *entity.BusinessCapabilityRevision) (*evidenceBundle, error) {
	bmEvidence, descriptors, err := s.fetchValidationEvidence(ctx, rev)
	if err != nil {
		return nil, err
	}
	revDigest, err := CapabilityContentDigest(rev.ToSemanticPayload())
	if err != nil {
		return nil, entity.ErrConsistency
	}
	descList := make([]*ContractLogicalDescriptor, 0, len(descriptors))
	for _, d := range descriptors {
		descList = append(descList, d)
	}
	contractDigest, err := ContractEvidenceDigest(descList)
	if err != nil {
		return nil, entity.ErrConsistency
	}
	bmDigest := ""
	if bmEvidence != nil {
		bmDigest = bmEvidence.ContentDigest
	}
	evidenceDigest, err := ValidationEvidenceDigest(rev.BusinessModelRevision, bmDigest, contractDigest)
	if err != nil {
		return nil, entity.ErrConsistency
	}
	return &evidenceBundle{
		bmEvidence:     bmEvidence,
		descriptors:    descriptors,
		revDigest:      revDigest,
		contractDigest: contractDigest,
		bmDigest:       bmDigest,
		evidenceDigest: evidenceDigest,
	}, nil
}

// requireFinalEvidenceFence re-reads BM/Contract evidence and requires byte-identical digests
// before Create ValidationResult, revision status mutation, or Activate's first status write.
// Capability row locks do not cover Business/Contract aggregates.
func (s *capabilityService) requireFinalEvidenceFence(ctx context.Context, rev *entity.BusinessCapabilityRevision, first *evidenceBundle) (*evidenceBundle, error) {
	if first == nil {
		return nil, entity.ErrConsistency
	}
	second, err := s.buildValidationEvidence(ctx, rev)
	if err != nil {
		return nil, err
	}
	if !evidenceBundleDigestsEqual(first, second) {
		return nil, entity.ErrConsistency
	}
	return second, nil
}

func evidenceBundleDigestsEqual(a, b *evidenceBundle) bool {
	if a == nil || b == nil {
		return false
	}
	return a.revDigest == b.revDigest &&
		a.bmDigest == b.bmDigest &&
		a.contractDigest == b.contractDigest &&
		a.evidenceDigest == b.evidenceDigest
}

func (s *capabilityService) fetchValidationEvidence(ctx context.Context, rev *entity.BusinessCapabilityRevision) (*BusinessModelRevisionEvidence, map[string]*ContractLogicalDescriptor, error) {
	bm, err := s.business.GetBusinessModelRevision(ctx, rev.TenantID, rev.BusinessID, rev.BusinessModelRevision)
	if err != nil {
		if errors.Is(err, entity.ErrBusinessModelNotFound) || errors.Is(err, entity.ErrNotFound) ||
			errors.Is(err, entity.ErrCrossTenant) {
			bm = nil
		} else {
			return nil, nil, mapValidationPortError(err)
		}
	}
	descriptors := map[string]*ContractLogicalDescriptor{}
	for _, b := range rev.DataContractBindings {
		desc, err := s.contract.GetActiveContractLogicalDescriptor(ctx, rev.TenantID, rev.BusinessID, b.DataContractID)
		if err != nil {
			// Missing/inactive/cross isolation → validation FAIL issue codes (nil descriptor).
			if errors.Is(err, entity.ErrContractNotFound) || errors.Is(err, entity.ErrContractNotActive) ||
				errors.Is(err, entity.ErrCrossTenant) || errors.Is(err, entity.ErrNotFound) {
				continue
			}
			return nil, nil, mapValidationPortError(err)
		}
		descriptors[b.DataContractID] = desc
	}
	return bm, descriptors, nil
}

func mapValidationPortError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, entity.ErrBusinessModelNotFound),
		errors.Is(err, entity.ErrContractNotFound),
		errors.Is(err, entity.ErrContractNotActive),
		errors.Is(err, entity.ErrCrossTenant),
		errors.Is(err, entity.ErrNotFound),
		errors.Is(err, entity.ErrPortsNotConfigured),
		errors.Is(err, entity.ErrNotConfigured),
		errors.Is(err, entity.ErrConsistency),
		errors.Is(err, entity.ErrForbidden),
		errors.Is(err, entity.ErrInvalidPayload):
		return err
	default:
		return entity.ErrConsistency
	}
}

func (s *capabilityService) Activate(ctx context.Context, tenantID, revisionID, actorID, reason string) (*entity.BusinessCapabilityRevision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	if !s.portsConfigured() {
		return nil, entity.ErrPortsNotConfigured
	}
	// Optimistic read outside the lock for CAS expected values only.
	revPeek, err := s.root().GetRevision(ctx, tenantID, revisionID)
	if err != nil {
		return nil, err
	}
	capPeek, err := s.root().GetCapability(ctx, tenantID, revPeek.CapabilityID)
	if err != nil {
		return nil, err
	}
	expectedActive := capPeek.ActiveRevisionID
	expectedGen := capPeek.AggregateGeneration

	var out *entity.BusinessCapabilityRevision
	err = s.uow.WithinTransaction(ctx, func(tx CapabilityTx) error {
		cap, rev, err := s.lockCapAndRev(ctx, tx, tenantID, revisionID)
		if err != nil {
			return err
		}
		if cap.ActiveRevisionID != expectedActive || cap.AggregateGeneration != expectedGen {
			return entity.ErrActiveConflict
		}
		if rev.Status != entity.RevisionValidated {
			return entity.ErrIllegalTransition
		}
		if !AllowTransition(rev.Status, entity.RevisionActive) {
			return entity.ErrIllegalTransition
		}
		ev, err := s.buildValidationEvidence(ctx, rev)
		if err != nil {
			return err
		}
		pass, err := tx.Repo().GetValidationByEvidence(ctx, tenantID, revisionID, ev.evidenceDigest)
		if err != nil || pass.Status != entity.ValidationPass ||
			pass.RevisionContentDigest != ev.revDigest ||
			pass.BusinessModelRevision != rev.BusinessModelRevision ||
			pass.BusinessModelContentDigest != ev.bmDigest {
			return entity.ErrMissingValidationEvidence
		}
		// Fresh digests under lock — old PASS must fail if descriptors drifted.
		if pass.EvidenceDigest != ev.evidenceDigest || pass.ContractEvidenceDigest != ev.contractDigest {
			return entity.ErrMissingValidationEvidence
		}
		// List revisions before the final fence so no repo/port I/O sits between
		// fence success and the first UpdateRevisionStatus (S5-G3-F3).
		revs, err := tx.Repo().ListRevisions(ctx, tenantID, cap.CapabilityID)
		if err != nil {
			return err
		}
		// Final fence before first Activate status write (Capability lock ≠ BM/Contract lock).
		ev, err = s.requireFinalEvidenceFence(ctx, rev, ev)
		if err != nil {
			return err
		}
		if pass.EvidenceDigest != ev.evidenceDigest || pass.ContractEvidenceDigest != ev.contractDigest ||
			pass.RevisionContentDigest != ev.revDigest || pass.BusinessModelContentDigest != ev.bmDigest {
			return entity.ErrMissingValidationEvidence
		}
		// Between fence and first status write: memory checks only (no port/repo/DAO calls).
		for _, r := range revs {
			if r.Status == entity.RevisionActive && r.RevisionID != revisionID {
				ok, err := tx.Repo().UpdateRevisionStatus(ctx, tenantID, r.RevisionID, entity.RevisionActive, entity.RevisionDeprecated)
				if err != nil || !ok {
					return entity.ErrActiveConflict
				}
			}
		}
		ok, err := tx.Repo().UpdateRevisionStatus(ctx, tenantID, revisionID, entity.RevisionValidated, entity.RevisionActive)
		if err != nil || !ok {
			return entity.ErrActiveConflict
		}
		if err := tx.Repo().UpdateActiveRevisionID(ctx, tenantID, cap.CapabilityID, revisionID); err != nil {
			return err
		}
		bumped, err := tx.Repo().CASBumpAggregateGeneration(ctx, tenantID, cap.CapabilityID, expectedGen)
		if err != nil {
			return err
		}
		if !bumped {
			return entity.ErrActiveConflict
		}
		dec := &entity.CapabilityDecision{
			DecisionID: newID("cdec"), TenantID: tenantID, BusinessID: cap.BusinessID,
			CapabilityID: cap.CapabilityID, TargetRevisionID: revisionID,
			Action: entity.DecisionActivate, ActorPrincipalID: actorID, Reason: reason, CreatedAt: s.now(),
		}
		if err := tx.Repo().CreateDecision(ctx, dec); err != nil {
			return err
		}
		cap.ActiveRevisionID = revisionID
		cap.AggregateGeneration = expectedGen + 1
		revs, err = tx.Repo().ListRevisions(ctx, tenantID, cap.CapabilityID)
		if err != nil {
			return err
		}
		if err := projectAndUpdateAssets(ctx, tx.Assets(), cap, revs); err != nil {
			return err
		}
		rev.Status = entity.RevisionActive
		out = rev
		return nil
	})
	return out, err
}

func (s *capabilityService) MarkStale(ctx context.Context, tenantID, revisionID, actorID, reason string) (*entity.BusinessCapabilityRevision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	_ = tenantID
	_ = revisionID
	_ = actorID
	_ = reason
	// Fail-closed until impact evidence exists (G3+).
	return nil, entity.ErrMissingImpactEvidence
}

func (s *capabilityService) Deprecate(ctx context.Context, tenantID, revisionID, actorID, reason string) (*entity.BusinessCapabilityRevision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	var out *entity.BusinessCapabilityRevision
	err := s.uow.WithinTransaction(ctx, func(tx CapabilityTx) error {
		cap, rev, err := s.lockCapAndRev(ctx, tx, tenantID, revisionID)
		if err != nil {
			return err
		}
		from := rev.Status
		if !AllowTransition(from, entity.RevisionDeprecated) {
			return entity.ErrIllegalTransition
		}
		ok, err := tx.Repo().UpdateRevisionStatus(ctx, tenantID, revisionID, from, entity.RevisionDeprecated)
		if err != nil || !ok {
			return entity.ErrIllegalTransition
		}
		if cap.ActiveRevisionID == revisionID {
			if err := tx.Repo().UpdateActiveRevisionID(ctx, tenantID, cap.CapabilityID, ""); err != nil {
				return err
			}
			cap.ActiveRevisionID = ""
		}
		dec := &entity.CapabilityDecision{
			DecisionID: newID("cdec"), TenantID: tenantID, BusinessID: cap.BusinessID,
			CapabilityID: cap.CapabilityID, TargetRevisionID: revisionID,
			Action: entity.DecisionDeprecate, ActorPrincipalID: actorID, Reason: reason, CreatedAt: s.now(),
		}
		if err := tx.Repo().CreateDecision(ctx, dec); err != nil {
			return err
		}
		rev.Status = entity.RevisionDeprecated
		revs, err := tx.Repo().ListRevisions(ctx, tenantID, cap.CapabilityID)
		if err != nil {
			return err
		}
		if err := projectAndUpdateAssets(ctx, tx.Assets(), cap, revs); err != nil {
			return err
		}
		out = rev
		return nil
	})
	return out, err
}

func (s *capabilityService) lockCapAndRev(ctx context.Context, tx CapabilityTx, tenantID, revisionID string) (*entity.BusinessCapability, *entity.BusinessCapabilityRevision, error) {
	rev, err := tx.Repo().GetRevision(ctx, tenantID, revisionID)
	if err != nil {
		return nil, nil, err
	}
	cap, err := tx.Repo().GetCapabilityForUpdate(ctx, tenantID, rev.CapabilityID)
	if err != nil {
		return nil, nil, err
	}
	rev, err = tx.Repo().GetRevisionForUpdate(ctx, tenantID, revisionID)
	if err != nil {
		return nil, nil, err
	}
	return cap, rev, nil
}

func (s *capabilityService) requireValidateProvenance(ctx context.Context, tx CapabilityTx, rev *entity.BusinessCapabilityRevision) error {
	switch rev.Source {
	case entity.SourceAIProposal:
		if rev.ProposalID == "" || rev.AnalysisRunID == "" {
			return entity.ErrMissingProvenance
		}
		prop, err := tx.Repo().GetProposal(ctx, rev.TenantID, rev.ProposalID)
		if err != nil {
			return entity.ErrMissingProvenance
		}
		if prop.TenantID != rev.TenantID || prop.BusinessID != rev.BusinessID {
			return entity.ErrMissingProvenance
		}
		if prop.Status != entity.ProposalConfirmed && prop.Status != entity.ProposalEditConfirmed {
			return entity.ErrMissingProvenance
		}
		if prop.CapabilityID == "" || prop.CapabilityID != rev.CapabilityID {
			return entity.ErrMissingProvenance
		}
		if prop.MaterializedRevisionID == "" || prop.MaterializedRevisionID != rev.RevisionID {
			return entity.ErrMissingProvenance
		}
		if prop.AnalysisRunID == "" || prop.AnalysisRunID != rev.AnalysisRunID {
			return entity.ErrMissingProvenance
		}
		dec, err := tx.Repo().GetDecisionByProposal(ctx, rev.TenantID, rev.ProposalID)
		if err != nil {
			return entity.ErrMissingProvenance
		}
		if dec.Action != entity.DecisionConfirm && dec.Action != entity.DecisionEditConfirm {
			return entity.ErrMissingProvenance
		}
		if dec.TenantID != rev.TenantID || dec.BusinessID != rev.BusinessID {
			return entity.ErrMissingProvenance
		}
		if dec.ProposalID != rev.ProposalID {
			return entity.ErrMissingProvenance
		}
		if dec.CapabilityID == "" || dec.CapabilityID != rev.CapabilityID {
			return entity.ErrMissingProvenance
		}
		if dec.TargetRevisionID == "" || dec.TargetRevisionID != rev.RevisionID {
			return entity.ErrMissingProvenance
		}
		run, err := tx.Repo().GetAnalysisRun(ctx, rev.TenantID, rev.AnalysisRunID)
		if err != nil {
			return entity.ErrMissingProvenance
		}
		if run.TenantID != rev.TenantID || run.BusinessID != rev.BusinessID || run.AnalysisRunID != rev.AnalysisRunID {
			return entity.ErrMissingProvenance
		}
		if run.BusinessModelRevision != rev.BusinessModelRevision {
			return entity.ErrMissingProvenance
		}
		if run.Status != entity.AnalysisSucceeded {
			return entity.ErrMissingProvenance
		}
		return nil
	case entity.SourceManualCreated:
		decs, err := tx.Repo().ListDecisionsByCapability(ctx, rev.TenantID, rev.CapabilityID)
		if err != nil {
			return err
		}
		for _, d := range decs {
			if d.Action == entity.DecisionCreate &&
				d.TenantID == rev.TenantID &&
				d.BusinessID == rev.BusinessID &&
				d.CapabilityID == rev.CapabilityID &&
				d.TargetRevisionID == rev.RevisionID {
				return nil
			}
		}
		return entity.ErrMissingProvenance
	case entity.SourceDerivedEdit:
		if rev.DerivedFromRevisionID == "" {
			return entity.ErrMissingProvenance
		}
		decs, err := tx.Repo().ListDecisionsByCapability(ctx, rev.TenantID, rev.CapabilityID)
		if err != nil {
			return err
		}
		matched := false
		for _, d := range decs {
			if (d.Action == entity.DecisionEdit || d.Action == entity.DecisionDerive) &&
				d.TenantID == rev.TenantID &&
				d.BusinessID == rev.BusinessID &&
				d.CapabilityID == rev.CapabilityID &&
				d.SourceRevisionID == rev.DerivedFromRevisionID &&
				d.TargetRevisionID == rev.RevisionID {
				matched = true
				break
			}
		}
		if !matched {
			return entity.ErrMissingProvenance
		}
		src, err := tx.Repo().GetRevision(ctx, rev.TenantID, rev.DerivedFromRevisionID)
		if err != nil {
			return entity.ErrMissingProvenance
		}
		if src.TenantID != rev.TenantID || src.BusinessID != rev.BusinessID || src.CapabilityID != rev.CapabilityID {
			return entity.ErrMissingProvenance
		}
		return nil
	default:
		return entity.ErrMissingProvenance
	}
}

func projectAndUpdateAssets(ctx context.Context, assets AssetProjection, cap *entity.BusinessCapability, revs []*entity.BusinessCapabilityRevision) error {
	proj, err := ProjectCapabilityAssetRef(cap, revs)
	if err != nil {
		return err
	}
	return assets.UpdateCapabilityProjection(ctx, cap.TenantID, cap.CapabilityID, proj.Name, proj.SemanticVersion, proj.ContentDigest, proj.Status)
}

func revisionFromPayload(tenantID, businessID, capID, revID string, version int32, source entity.Source, p entity.SemanticPayload, actor string, now time.Time) *entity.BusinessCapabilityRevision {
	return &entity.BusinessCapabilityRevision{
		RevisionID: revID, CapabilityID: capID, TenantID: tenantID, BusinessID: businessID,
		Version: version, Status: entity.RevisionDraft,
		Name: p.Name, Description: p.Description, BusinessModelRevision: p.BusinessModelRevision,
		CapabilityKind: p.CapabilityKind, InputSchema: p.InputSchema, OutputSchema: p.OutputSchema,
		Preconditions: append([]entity.Precondition(nil), p.Preconditions...),
		Effects: append([]entity.Effect(nil), p.Effects...),
		DataContractBindings: append([]entity.DataContractBinding(nil), p.DataContractBindings...),
		QueryOperation: p.QueryOperation, OutputCardinality: p.OutputCardinality,
		Source: source, CreatedBy: actor, CreatedAt: now,
	}
}
