/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

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
	GetRevision(ctx context.Context, tenantID, revisionID string) (*entity.BusinessCapabilityRevision, error)
	ListRevisions(ctx context.Context, tenantID, capabilityID string) ([]*entity.BusinessCapabilityRevision, error)
	Validate(ctx context.Context, tenantID, revisionID, actorID string) (*entity.BusinessCapabilityRevision, error)
	Activate(ctx context.Context, tenantID, revisionID, actorID, reason string) (*entity.BusinessCapabilityRevision, error)
	MarkStale(ctx context.Context, tenantID, revisionID, actorID, reason string) (*entity.BusinessCapabilityRevision, error)
	Deprecate(ctx context.Context, tenantID, revisionID, actorID, reason string) (*entity.BusinessCapabilityRevision, error)
}

// Components wires CapabilityService dependencies.
type Components struct {
	Repo      repository.CapabilityRepository
	Assets    AssetProjection
	Generator ProposalGenerator
	Clock     Clock
	// DB when set binds Capability repo + AssetRef projection onto one GORM transaction.
	DB *gorm.DB
}

type capabilityService struct {
	repo      repository.CapabilityRepository
	assets    AssetProjection
	generator ProposalGenerator
	clock     Clock
	db        *gorm.DB
}

func NewCapabilityService(c *Components) CapabilityService {
	clk := Clock(realClock{})
	if c != nil && c.Clock != nil {
		clk = c.Clock
	}
	var repo repository.CapabilityRepository
	var assets AssetProjection
	var gen ProposalGenerator
	var db *gorm.DB
	if c != nil {
		repo = c.Repo
		assets = c.Assets
		gen = c.Generator
		db = c.DB
	}
	return &capabilityService{repo: repo, assets: assets, generator: gen, clock: clk, db: db}
}

func (s *capabilityService) configured() bool {
	return s.repo != nil && s.assets != nil
}

// withinTx runs fn with Capability + AssetProjection sharing one logical transaction.
func (s *capabilityService) withinTx(ctx context.Context, fn func(tx repository.CapabilityRepository, assets AssetProjection) error) error {
	if s.db != nil {
		return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			return fn(repository.NewCapabilityRepository(tx), NewGormAssetProjection(tx))
		})
	}
	if mem, ok := s.assets.(*MemoryAssetProjection); ok {
		return s.repo.Transaction(ctx, func(txRepo repository.CapabilityRepository) error {
			snap := mem.snapshot()
			err := fn(txRepo, mem)
			if err != nil {
				mem.restore(snap)
			}
			return err
		})
	}
	return s.repo.Transaction(ctx, func(txRepo repository.CapabilityRepository) error {
		return fn(txRepo, s.assets)
	})
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
	if err := validateSemanticBasics(in.Payload); err != nil {
		return nil, nil, err
	}
	capID := strings.TrimSpace(in.CapabilityID)
	if capID == "" {
		capID = newID("cap")
	}
	now := s.now()
	var outCap *entity.BusinessCapability
	var outRev *entity.BusinessCapabilityRevision
	err := s.withinTx(ctx, func(tx repository.CapabilityRepository, assets AssetProjection) error {
		cap := &entity.BusinessCapability{
			CapabilityID: capID, TenantID: in.TenantID, BusinessID: in.BusinessID,
			CreatedBy: in.ActorID, CreatedAt: now, UpdatedAt: now,
		}
		if err := tx.CreateCapability(ctx, cap); err != nil {
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
			Status: proj.Status, OwnerID: in.OwnerID, CreatedBy: parseActorInt(in.ActorID),
			ContentDigest: proj.ContentDigest, CreatedAt: now, UpdatedAt: now,
		}
		if err := assets.CreateCapabilityAsset(ctx, asset); err != nil {
			return err
		}
		if err := tx.CreateRevision(ctx, rev); err != nil {
			return err
		}
		dec := &entity.CapabilityDecision{
			DecisionID: newID("cdec"), TenantID: in.TenantID, BusinessID: in.BusinessID,
			CapabilityID: capID, TargetRevisionID: revID, Action: entity.DecisionCreate,
			PayloadDigest: digest, ActorPrincipalID: in.ActorID, CreatedAt: now,
		}
		if err := tx.CreateDecision(ctx, dec); err != nil {
			return err
		}
		if err := projectAndUpdateAssets(ctx, assets, cap, []*entity.BusinessCapabilityRevision{rev}); err != nil {
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
	digest, err := DecisionPayloadDigest(in.Payload)
	if err != nil {
		return nil, nil, err
	}
	if existing, err := s.repo.GetDecisionByDeriveKey(ctx, in.TenantID, in.CapabilityID, in.SourceRevisionID, in.ClientRequestID); err == nil {
		if existing.PayloadDigest != digest {
			return nil, nil, entity.ErrIdempotencyConflict
		}
		rev, err := s.repo.GetRevision(ctx, in.TenantID, existing.TargetRevisionID)
		return rev, existing, err
	} else if !errors.Is(err, entity.ErrDecisionNotFound) {
		return nil, nil, err
	}

	var outRev *entity.BusinessCapabilityRevision
	var outDec *entity.CapabilityDecision
	err = s.withinTx(ctx, func(tx repository.CapabilityRepository, assets AssetProjection) error {
		cap, err := tx.GetCapabilityForUpdate(ctx, in.TenantID, in.CapabilityID)
		if err != nil {
			return err
		}
		src, err := tx.GetRevisionForUpdate(ctx, in.TenantID, in.SourceRevisionID)
		if err != nil {
			return err
		}
		if src.CapabilityID != in.CapabilityID || src.TenantID != in.TenantID {
			return entity.ErrCrossTenant
		}
		if existing, err := tx.GetDecisionByDeriveKey(ctx, in.TenantID, in.CapabilityID, in.SourceRevisionID, in.ClientRequestID); err == nil {
			if existing.PayloadDigest != digest {
				return entity.ErrIdempotencyConflict
			}
			rev, err := tx.GetRevision(ctx, in.TenantID, existing.TargetRevisionID)
			if err != nil {
				return err
			}
			outRev, outDec = rev, existing
			return nil
		} else if !errors.Is(err, entity.ErrDecisionNotFound) {
			return err
		}
		max, err := tx.MaxVersion(ctx, in.TenantID, in.CapabilityID)
		if err != nil {
			return err
		}
		now := s.now()
		revID := newID("crev")
		rev := revisionFromPayload(in.TenantID, cap.BusinessID, in.CapabilityID, revID, max+1, entity.SourceDerivedEdit, in.Payload, in.ActorID, now)
		rev.DerivedFromRevisionID = in.SourceRevisionID
		if err := tx.CreateRevision(ctx, rev); err != nil {
			return err
		}
		dec := &entity.CapabilityDecision{
			DecisionID: newID("cdec"), TenantID: in.TenantID, BusinessID: cap.BusinessID,
			CapabilityID: in.CapabilityID, SourceRevisionID: in.SourceRevisionID, TargetRevisionID: revID,
			Action: action, PayloadDigest: digest, ClientRequestID: in.ClientRequestID,
			ActorPrincipalID: in.ActorID, Reason: in.Reason, CreatedAt: now,
		}
		if err := tx.CreateDecision(ctx, dec); err != nil {
			return err
		}
		revs, err := tx.ListRevisions(ctx, in.TenantID, in.CapabilityID)
		if err != nil {
			return err
		}
		if err := projectAndUpdateAssets(ctx, assets, cap, revs); err != nil {
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
	analysis := in.Analysis
	analysis.BusinessModelRevision = in.BusinessModelRevision
	digest, err := AnalysisRequestDigest(analysis)
	if err != nil {
		return nil, err
	}
	reqJSON, _ := json.Marshal(analysis)
	now := s.now()
	run := &entity.CapabilityAnalysisRun{
		AnalysisRunID: newID("carun"), TenantID: in.TenantID, BusinessID: in.BusinessID,
		BusinessModelRevision: in.BusinessModelRevision, ClientRequestID: in.ClientRequestID,
		RequestDigest: digest, Status: entity.AnalysisPending, Attempt: 1,
		RequestJSON: string(reqJSON), CreatedBy: in.ActorID, CreatedAt: now, UpdatedAt: now,
	}
	seedExecutionLease(run, now)
	existing, created, err := s.repo.CreateOrClaimAnalysisRun(ctx, run)
	if err != nil {
		return nil, err
	}
	if !created {
		return s.handleExistingAnalysis(ctx, existing, analysis)
	}
	return s.executeAnalysis(ctx, existing, analysis)
}

func (s *capabilityService) handleExistingAnalysis(ctx context.Context, existing *entity.CapabilityAnalysisRun, analysis entity.AnalysisRequest) (*AnalysisResult, error) {
	props, _ := s.repo.ListProposalsByAnalysisRun(ctx, existing.TenantID, existing.AnalysisRunID)
	now := s.now()
	if existing.Status == entity.AnalysisPending && analysisLeaseExpired(existing, now) {
		claimed, owned, err := s.repo.ClaimExpiredPendingExecution(ctx, existing.TenantID, existing.AnalysisRunID, existing.Attempt, now)
		if err != nil {
			return nil, err
		}
		if owned {
			return s.executeAnalysis(ctx, claimed, analysis)
		}
	}
	return &AnalysisResult{Run: existing, Proposals: props, OwnedExecute: false}, nil
}

func (s *capabilityService) executeAnalysis(ctx context.Context, run *entity.CapabilityAnalysisRun, analysis entity.AnalysisRequest) (*AnalysisResult, error) {
	attempt := run.Attempt
	res, err := s.generator.Generate(ctx, GenerateRequest{
		TenantID: run.TenantID, BusinessID: run.BusinessID, BusinessModelRevision: run.BusinessModelRevision,
		AnalysisRunID: run.AnalysisRunID, Analysis: analysis,
	})
	if err != nil {
		_ = s.repo.MarkAnalysisFailed(ctx, run.TenantID, run.AnalysisRunID, "FORMA_CAPABILITY_MODEL_FAILED", attempt)
		failed, _ := s.repo.GetAnalysisRun(ctx, run.TenantID, run.AnalysisRunID)
		return &AnalysisResult{Run: failed, OwnedExecute: true}, err
	}
	now := s.now()
	var created []*entity.CapabilityProposal
	err = s.withinTx(ctx, func(tx repository.CapabilityRepository, assets AssetProjection) error {
		for _, payload := range res.Proposals {
			p := &entity.CapabilityProposal{
				ProposalID: newID("cprop"), TenantID: run.TenantID, BusinessID: run.BusinessID,
				AnalysisRunID: run.AnalysisRunID, Status: entity.ProposalProposed,
				Payload: payload, CreatedAt: now,
			}
			if err := tx.CreateProposal(ctx, p); err != nil {
				return err
			}
			created = append(created, p)
		}
		return tx.MarkAnalysisSucceeded(ctx, run.TenantID, run.AnalysisRunID, res.ModelRef, attempt)
	})
	if err != nil {
		if errors.Is(err, entity.ErrStaleGeneration) {
			final, _ := s.repo.GetAnalysisRun(ctx, run.TenantID, run.AnalysisRunID)
			props, _ := s.repo.ListProposalsByAnalysisRun(ctx, run.TenantID, run.AnalysisRunID)
			return &AnalysisResult{Run: final, Proposals: props, OwnedExecute: true}, entity.ErrStaleGeneration
		}
		_ = s.repo.MarkAnalysisFailed(ctx, run.TenantID, run.AnalysisRunID, "FORMA_CAPABILITY_PERSIST_FAILED", attempt)
		failed, _ := s.repo.GetAnalysisRun(ctx, run.TenantID, run.AnalysisRunID)
		return &AnalysisResult{Run: failed, OwnedExecute: true}, err
	}
	final, _ := s.repo.GetAnalysisRun(ctx, run.TenantID, run.AnalysisRunID)
	return &AnalysisResult{Run: final, Proposals: created, OwnedExecute: true}, nil
}

func (s *capabilityService) GetAnalysisRun(ctx context.Context, tenantID, analysisRunID string) (*entity.CapabilityAnalysisRun, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	return s.repo.GetAnalysisRun(ctx, tenantID, analysisRunID)
}

func (s *capabilityService) RetryFailedAnalysis(ctx context.Context, tenantID, analysisRunID, actorID string) (*AnalysisResult, error) {
	if !s.configured() || s.generator == nil {
		return nil, entity.ErrNotConfigured
	}
	run, err := s.repo.GetAnalysisRun(ctx, tenantID, analysisRunID)
	if err != nil {
		return nil, err
	}
	if run.Status != entity.AnalysisFailed {
		return nil, entity.ErrAnalysisNotFailed
	}
	ok, attempt, err := s.repo.ClaimAnalysisRetry(ctx, tenantID, analysisRunID, actorID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, entity.ErrAnalysisNotFailed
	}
	run, _ = s.repo.GetAnalysisRun(ctx, tenantID, analysisRunID)
	run.Attempt = attempt
	var analysis entity.AnalysisRequest
	if json.Unmarshal([]byte(run.RequestJSON), &analysis) != nil {
		_ = s.repo.MarkAnalysisFailed(ctx, tenantID, analysisRunID, "FORMA_CAPABILITY_INVALID_REQUEST", attempt)
		return nil, entity.ErrInvalidPayload
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
	if err := validateSemanticBasics(in.EffectivePayload); err != nil {
		return nil, err
	}
	return s.materializeProposal(ctx, in.TenantID, in.ProposalID, in.ActorID, in.Reason, in.ClientRequestID, in.CapabilityID, entity.DecisionEditConfirm, &in.EffectivePayload)
}

func (s *capabilityService) materializeProposal(ctx context.Context, tenantID, proposalID, actorID, reason, clientRequestID, capabilityID string, action entity.DecisionAction, editPayload *entity.SemanticPayload) (*entity.BusinessCapabilityRevision, error) {
	var out *entity.BusinessCapabilityRevision
	err := s.withinTx(ctx, func(tx repository.CapabilityRepository, assets AssetProjection) error {
		prop, err := tx.GetProposalForUpdate(ctx, tenantID, proposalID)
		if err != nil {
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

		capID := strings.TrimSpace(capabilityID)
		if capID == "" {
			capID = strings.TrimSpace(prop.CapabilityID)
		}
		now := s.now()
		terminalStatus := entity.ProposalConfirmed
		if action == entity.DecisionEditConfirm {
			terminalStatus = entity.ProposalEditConfirmed
		}

		if capID != "" {
			if _, err := tx.GetCapability(ctx, tenantID, capID); err == nil {
				return s.materializeExisting(ctx, tx, assets, prop, capID, effective, digest, action, terminalStatus, actorID, reason, clientRequestID, now, &out)
			} else if !errors.Is(err, entity.ErrNotFound) {
				return err
			}
		}
		if capID == "" {
			capID = newID("cap")
		}
		return s.materializeFirstCreate(ctx, tx, assets, prop, capID, effective, digest, action, terminalStatus, actorID, reason, clientRequestID, now, &out)
	})
	return out, err
}

func (s *capabilityService) replayTerminal(ctx context.Context, tx repository.CapabilityRepository, prop *entity.CapabilityProposal, digest string, out **entity.BusinessCapabilityRevision) error {
	dec, err := tx.GetDecisionByProposal(ctx, prop.TenantID, prop.ProposalID)
	if err != nil {
		return entity.ErrConsistency
	}
	if dec.PayloadDigest != digest {
		return entity.ErrIdempotencyConflict
	}
	if prop.MaterializedRevisionID == "" {
		return entity.ErrConsistency
	}
	rev, err := tx.GetRevision(ctx, prop.TenantID, prop.MaterializedRevisionID)
	if err != nil {
		return entity.ErrConsistency
	}
	*out = rev
	return nil
}

func (s *capabilityService) materializeExisting(ctx context.Context, tx repository.CapabilityRepository, assets AssetProjection, prop *entity.CapabilityProposal, capID string, effective entity.SemanticPayload, digest string, action entity.DecisionAction, terminal entity.ProposalStatus, actorID, reason, clientRequestID string, now time.Time, out **entity.BusinessCapabilityRevision) error {
	cap, err := tx.GetCapabilityForUpdate(ctx, prop.TenantID, capID)
	if err != nil {
		return err
	}
	if cap.BusinessID != prop.BusinessID {
		return entity.ErrCrossTenant
	}
	max, err := tx.MaxVersion(ctx, prop.TenantID, capID)
	if err != nil {
		return err
	}
	revID := newID("crev")
	rev := revisionFromPayload(prop.TenantID, prop.BusinessID, capID, revID, max+1, entity.SourceAIProposal, effective, actorID, now)
	rev.AnalysisRunID = prop.AnalysisRunID
	rev.ProposalID = prop.ProposalID
	if err := tx.CreateRevision(ctx, rev); err != nil {
		return err
	}
	dec := &entity.CapabilityDecision{
		DecisionID: newID("cdec"), TenantID: prop.TenantID, BusinessID: prop.BusinessID,
		CapabilityID: capID, ProposalID: prop.ProposalID, TargetRevisionID: revID,
		Action: action, PayloadDigest: digest, ClientRequestID: clientRequestID,
		ActorPrincipalID: actorID, Reason: reason, CreatedAt: now,
	}
	if err := tx.CreateDecision(ctx, dec); err != nil {
		return err
	}
	if err := tx.TerminalizeProposal(ctx, prop.TenantID, prop.ProposalID, terminal, revID, capID); err != nil {
		return err
	}
	revs, err := tx.ListRevisions(ctx, prop.TenantID, capID)
	if err != nil {
		return err
	}
	if err := projectAndUpdateAssets(ctx, assets, cap, revs); err != nil {
		return err
	}
	*out = rev
	return nil
}

func (s *capabilityService) materializeFirstCreate(ctx context.Context, tx repository.CapabilityRepository, assets AssetProjection, prop *entity.CapabilityProposal, capID string, effective entity.SemanticPayload, digest string, action entity.DecisionAction, terminal entity.ProposalStatus, actorID, reason, clientRequestID string, now time.Time, out **entity.BusinessCapabilityRevision) error {
	cap := &entity.BusinessCapability{
		CapabilityID: capID, TenantID: prop.TenantID, BusinessID: prop.BusinessID,
		CreatedBy: actorID, CreatedAt: now, UpdatedAt: now,
	}
	if err := tx.CreateCapability(ctx, cap); err != nil {
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
		Status: proj.Status, ContentDigest: proj.ContentDigest, CreatedAt: now, UpdatedAt: now,
	}
	if err := assets.CreateCapabilityAsset(ctx, asset); err != nil {
		return err
	}
	if err := tx.CreateRevision(ctx, rev); err != nil {
		return err
	}
	dec := &entity.CapabilityDecision{
		DecisionID: newID("cdec"), TenantID: prop.TenantID, BusinessID: prop.BusinessID,
		CapabilityID: capID, ProposalID: prop.ProposalID, TargetRevisionID: revID,
		Action: action, PayloadDigest: digest, ClientRequestID: clientRequestID,
		ActorPrincipalID: actorID, Reason: reason, CreatedAt: now,
	}
	if err := tx.CreateDecision(ctx, dec); err != nil {
		return err
	}
	if err := tx.TerminalizeProposal(ctx, prop.TenantID, prop.ProposalID, terminal, revID, capID); err != nil {
		return err
	}
	if err := projectAndUpdateAssets(ctx, assets, cap, []*entity.BusinessCapabilityRevision{rev}); err != nil {
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
	err := s.withinTx(ctx, func(tx repository.CapabilityRepository, assets AssetProjection) error {
		prop, err := tx.GetProposalForUpdate(ctx, in.TenantID, in.ProposalID)
		if err != nil {
			return err
		}
		switch prop.Status {
		case entity.ProposalRejected:
			dec, err := tx.GetDecisionByProposal(ctx, in.TenantID, in.ProposalID)
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
		if err := tx.CreateDecision(ctx, dec); err != nil {
			return err
		}
		if err := tx.TerminalizeProposal(ctx, in.TenantID, in.ProposalID, entity.ProposalRejected, "", prop.CapabilityID); err != nil {
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
	return s.repo.GetCapability(ctx, tenantID, capabilityID)
}

func (s *capabilityService) GetRevision(ctx context.Context, tenantID, revisionID string) (*entity.BusinessCapabilityRevision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	return s.repo.GetRevision(ctx, tenantID, revisionID)
}

func (s *capabilityService) ListRevisions(ctx context.Context, tenantID, capabilityID string) ([]*entity.BusinessCapabilityRevision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	return s.repo.ListRevisions(ctx, tenantID, capabilityID)
}

func (s *capabilityService) Validate(ctx context.Context, tenantID, revisionID, actorID string) (*entity.BusinessCapabilityRevision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	var out *entity.BusinessCapabilityRevision
	err := s.withinTx(ctx, func(tx repository.CapabilityRepository, assets AssetProjection) error {
		cap, rev, err := s.lockCapAndRev(ctx, tx, tenantID, revisionID)
		if err != nil {
			return err
		}
		if rev.Status != entity.RevisionDraft {
			return entity.ErrIllegalTransition
		}
		if !AllowTransition(rev.Status, entity.RevisionValidated) {
			return entity.ErrIllegalTransition
		}
		if err := s.requireValidateProvenance(ctx, tx, rev); err != nil {
			return err
		}
		ok, err := tx.UpdateRevisionStatus(ctx, tenantID, revisionID, entity.RevisionDraft, entity.RevisionValidated)
		if err != nil || !ok {
			return entity.ErrIllegalTransition
		}
		rev.Status = entity.RevisionValidated
		revs, err := tx.ListRevisions(ctx, tenantID, cap.CapabilityID)
		if err != nil {
			return err
		}
		if err := projectAndUpdateAssets(ctx, assets, cap, revs); err != nil {
			return err
		}
		out = rev
		_ = actorID
		return nil
	})
	return out, err
}

func (s *capabilityService) Activate(ctx context.Context, tenantID, revisionID, actorID, reason string) (*entity.BusinessCapabilityRevision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	var out *entity.BusinessCapabilityRevision
	err := s.withinTx(ctx, func(tx repository.CapabilityRepository, assets AssetProjection) error {
		cap, rev, err := s.lockCapAndRev(ctx, tx, tenantID, revisionID)
		if err != nil {
			return err
		}
		if rev.Status != entity.RevisionValidated {
			return entity.ErrIllegalTransition
		}
		if !AllowTransition(rev.Status, entity.RevisionActive) {
			return entity.ErrIllegalTransition
		}
		revs, err := tx.ListRevisions(ctx, tenantID, cap.CapabilityID)
		if err != nil {
			return err
		}
		for _, r := range revs {
			if r.Status == entity.RevisionActive && r.RevisionID != revisionID {
				ok, err := tx.UpdateRevisionStatus(ctx, tenantID, r.RevisionID, entity.RevisionActive, entity.RevisionDeprecated)
				if err != nil || !ok {
					return entity.ErrActiveConflict
				}
			}
		}
		ok, err := tx.UpdateRevisionStatus(ctx, tenantID, revisionID, entity.RevisionValidated, entity.RevisionActive)
		if err != nil || !ok {
			return entity.ErrActiveConflict
		}
		if err := tx.UpdateActiveRevisionID(ctx, tenantID, cap.CapabilityID, revisionID); err != nil {
			return err
		}
		cap.ActiveRevisionID = revisionID
		revs, err = tx.ListRevisions(ctx, tenantID, cap.CapabilityID)
		if err != nil {
			return err
		}
		if err := projectAndUpdateAssets(ctx, assets, cap, revs); err != nil {
			return err
		}
		rev.Status = entity.RevisionActive
		out = rev
		_ = actorID
		_ = reason
		return nil
	})
	return out, err
}

func (s *capabilityService) MarkStale(ctx context.Context, tenantID, revisionID, actorID, reason string) (*entity.BusinessCapabilityRevision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	var out *entity.BusinessCapabilityRevision
	err := s.withinTx(ctx, func(tx repository.CapabilityRepository, assets AssetProjection) error {
		cap, rev, err := s.lockCapAndRev(ctx, tx, tenantID, revisionID)
		if err != nil {
			return err
		}
		if rev.Status != entity.RevisionActive || !AllowTransition(rev.Status, entity.RevisionStale) {
			return entity.ErrIllegalTransition
		}
		ok, err := tx.UpdateRevisionStatus(ctx, tenantID, revisionID, entity.RevisionActive, entity.RevisionStale)
		if err != nil || !ok {
			return entity.ErrIllegalTransition
		}
		if cap.ActiveRevisionID == revisionID {
			if err := tx.UpdateActiveRevisionID(ctx, tenantID, cap.CapabilityID, ""); err != nil {
				return err
			}
			cap.ActiveRevisionID = ""
		}
		rev.Status = entity.RevisionStale
		revs, err := tx.ListRevisions(ctx, tenantID, cap.CapabilityID)
		if err != nil {
			return err
		}
		if err := projectAndUpdateAssets(ctx, assets, cap, revs); err != nil {
			return err
		}
		out = rev
		_ = actorID
		_ = reason
		return nil
	})
	return out, err
}

func (s *capabilityService) Deprecate(ctx context.Context, tenantID, revisionID, actorID, reason string) (*entity.BusinessCapabilityRevision, error) {
	if !s.configured() {
		return nil, entity.ErrNotConfigured
	}
	var out *entity.BusinessCapabilityRevision
	err := s.withinTx(ctx, func(tx repository.CapabilityRepository, assets AssetProjection) error {
		cap, rev, err := s.lockCapAndRev(ctx, tx, tenantID, revisionID)
		if err != nil {
			return err
		}
		from := rev.Status
		if !AllowTransition(from, entity.RevisionDeprecated) {
			return entity.ErrIllegalTransition
		}
		ok, err := tx.UpdateRevisionStatus(ctx, tenantID, revisionID, from, entity.RevisionDeprecated)
		if err != nil || !ok {
			return entity.ErrIllegalTransition
		}
		if cap.ActiveRevisionID == revisionID {
			if err := tx.UpdateActiveRevisionID(ctx, tenantID, cap.CapabilityID, ""); err != nil {
				return err
			}
			cap.ActiveRevisionID = ""
		}
		rev.Status = entity.RevisionDeprecated
		revs, err := tx.ListRevisions(ctx, tenantID, cap.CapabilityID)
		if err != nil {
			return err
		}
		if err := projectAndUpdateAssets(ctx, assets, cap, revs); err != nil {
			return err
		}
		out = rev
		_ = actorID
		_ = reason
		return nil
	})
	return out, err
}

func (s *capabilityService) lockCapAndRev(ctx context.Context, tx repository.CapabilityRepository, tenantID, revisionID string) (*entity.BusinessCapability, *entity.BusinessCapabilityRevision, error) {
	rev, err := tx.GetRevision(ctx, tenantID, revisionID)
	if err != nil {
		return nil, nil, err
	}
	cap, err := tx.GetCapabilityForUpdate(ctx, tenantID, rev.CapabilityID)
	if err != nil {
		return nil, nil, err
	}
	rev, err = tx.GetRevisionForUpdate(ctx, tenantID, revisionID)
	if err != nil {
		return nil, nil, err
	}
	return cap, rev, nil
}

func (s *capabilityService) requireValidateProvenance(ctx context.Context, tx repository.CapabilityRepository, rev *entity.BusinessCapabilityRevision) error {
	switch rev.Source {
	case entity.SourceAIProposal:
		if rev.ProposalID == "" {
			return entity.ErrMissingProvenance
		}
		dec, err := tx.GetDecisionByProposal(ctx, rev.TenantID, rev.ProposalID)
		if err != nil {
			return entity.ErrMissingProvenance
		}
		if dec.Action != entity.DecisionConfirm && dec.Action != entity.DecisionEditConfirm {
			return entity.ErrMissingProvenance
		}
		return nil
	case entity.SourceManualCreated:
		decs, err := tx.ListDecisionsByCapability(ctx, rev.TenantID, rev.CapabilityID)
		if err != nil {
			return err
		}
		for _, d := range decs {
			if d.Action == entity.DecisionCreate && d.TargetRevisionID == rev.RevisionID {
				return nil
			}
		}
		return entity.ErrMissingProvenance
	case entity.SourceDerivedEdit:
		decs, err := tx.ListDecisionsByCapability(ctx, rev.TenantID, rev.CapabilityID)
		if err != nil {
			return err
		}
		for _, d := range decs {
			if (d.Action == entity.DecisionEdit || d.Action == entity.DecisionDerive) &&
				d.TargetRevisionID == rev.RevisionID && d.SourceRevisionID != "" {
				return nil
			}
		}
		return entity.ErrMissingProvenance
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

func validateSemanticBasics(p entity.SemanticPayload) error {
	if strings.TrimSpace(p.Name) == "" || p.CapabilityKind == "" || p.BusinessModelRevision <= 0 {
		return entity.ErrInvalidPayload
	}
	switch p.CapabilityKind {
	case entity.KindQuery, entity.KindCommand:
	default:
		return entity.ErrInvalidPayload
	}
	return nil
}

func parseActorInt(actor string) int64 {
	var n int64
	_, _ = fmt.Sscanf(actor, "%d", &n)
	return n
}
