/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/business/repository"
)

type BusinessService interface {
	InitBusiness(ctx context.Context, tenantID, businessID, assetID, createdBy string, initial *entity.SemanticModel, changeSummary string) (*entity.BusinessModel, *entity.BusinessModelRevision, *entity.BusinessModelLayout, error)
	Get(ctx context.Context, tenantID, businessID string) (*entity.BusinessModel, error)
	List(ctx context.Context, tenantID string) ([]*entity.BusinessModel, error)
	GetModel(ctx context.Context, tenantID, businessID string) (*entity.BusinessModel, *entity.SemanticModel, *entity.BusinessModelRevision, error)
	SaveModel(ctx context.Context, tenantID, businessID, createdBy string, expectedRevision int32, model *entity.SemanticModel, changeSummary string) (*entity.BusinessModelRevision, bool, error)
	ListRevisions(ctx context.Context, tenantID, businessID string) ([]*entity.BusinessModelRevision, error)
	GetRevision(ctx context.Context, tenantID, businessID string, rev int32) (*entity.BusinessModelRevision, *entity.SemanticModel, error)
	Diff(ctx context.Context, tenantID, businessID string, from, to int32) (*entity.BusinessModelDiff, *entity.BusinessImpactSummary, error)
	GetLayout(ctx context.Context, tenantID, businessID string) (*entity.BusinessModelLayout, *entity.ViewLayout, error)
	SaveLayout(ctx context.Context, tenantID, businessID, updatedBy string, expectedLayoutRevision int32, basedOnModelRevision int32, layout *entity.ViewLayout) (*entity.BusinessModelLayout, error)
}

type Components struct {
	Repo repository.BusinessRepository
}

type businessServiceImpl struct {
	repo repository.BusinessRepository
}

func NewBusinessService(c *Components) BusinessService {
	return &businessServiceImpl{repo: c.Repo}
}

func emptySemantic() *entity.SemanticModel {
	return &entity.SemanticModel{
		SchemaVersion: entity.SemanticSchemaVersion,
		Nodes:         []entity.SemanticNode{},
		Edges:         []entity.SemanticEdge{},
		Rules:         []entity.BusinessRule{},
		States:        []entity.BusinessState{},
	}
}

func emptyLayout() *entity.ViewLayout {
	return &entity.ViewLayout{
		NodePositions: map[string]entity.NodePosition{},
		Zoom:          1,
		Viewport:      entity.Viewport{},
		Mode:          "manual",
	}
}

func (s *businessServiceImpl) InitBusiness(
	ctx context.Context,
	tenantID, businessID, assetID, createdBy string,
	initial *entity.SemanticModel,
	changeSummary string,
) (*entity.BusinessModel, *entity.BusinessModelRevision, *entity.BusinessModelLayout, error) {
	if tenantID == "" || businessID == "" || assetID == "" {
		return nil, nil, nil, fmt.Errorf("%w: tenant/business/asset required", entity.ErrInvalidModel)
	}
	if initial == nil {
		initial = emptySemantic()
	}
	if initial.SchemaVersion == "" {
		initial.SchemaVersion = entity.SemanticSchemaVersion
	}
	if err := ValidateSemanticModel(initial); err != nil {
		return nil, nil, nil, err
	}
	digest, raw, err := ContentDigest(initial)
	if err != nil {
		return nil, nil, nil, err
	}
	now := time.Now().UTC()
	if changeSummary == "" {
		changeSummary = "initial"
	}
	master := &entity.BusinessModel{
		TenantID:        tenantID,
		BusinessID:      businessID,
		AssetID:         assetID,
		CurrentRevision: 1,
		SchemaVersion:   initial.SchemaVersion,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	rev := &entity.BusinessModelRevision{
		TenantID:          tenantID,
		BusinessID:        businessID,
		RevisionNo:        1,
		BaseRevisionNo:    0,
		SchemaVersion:     initial.SchemaVersion,
		SemanticModelJSON: string(raw),
		ContentDigest:     digest,
		ChangeSummary:     changeSummary,
		CreatedBy:         createdBy,
		CreatedAt:         now,
	}
	layoutPayload, _ := json.Marshal(emptyLayout())
	layout := &entity.BusinessModelLayout{
		TenantID:             tenantID,
		BusinessID:           businessID,
		LayoutRevision:       1,
		BasedOnModelRevision: 1,
		LayoutJSON:           string(layoutPayload),
		UpdatedBy:            createdBy,
		UpdatedAt:            now,
	}
	if err := s.repo.CreateMaster(ctx, master); err != nil {
		return nil, nil, nil, err
	}
	if err := s.repo.CreateRevision(ctx, rev); err != nil {
		return nil, nil, nil, err
	}
	if err := s.repo.UpsertLayout(ctx, layout); err != nil {
		return nil, nil, nil, err
	}
	return master, rev, layout, nil
}

func (s *businessServiceImpl) Get(ctx context.Context, tenantID, businessID string) (*entity.BusinessModel, error) {
	m, err := s.repo.GetMaster(ctx, tenantID, businessID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, entity.ErrNotFound
	}
	return m, nil
}

func (s *businessServiceImpl) List(ctx context.Context, tenantID string) ([]*entity.BusinessModel, error) {
	return s.repo.ListMasters(ctx, tenantID)
}

func (s *businessServiceImpl) GetModel(ctx context.Context, tenantID, businessID string) (*entity.BusinessModel, *entity.SemanticModel, *entity.BusinessModelRevision, error) {
	m, err := s.Get(ctx, tenantID, businessID)
	if err != nil {
		return nil, nil, nil, err
	}
	rev, err := s.repo.GetRevision(ctx, tenantID, businessID, m.CurrentRevision)
	if err != nil {
		return nil, nil, nil, err
	}
	if rev == nil {
		return nil, nil, nil, entity.ErrRevisionNotFound
	}
	sm, err := parseSemantic(rev.SemanticModelJSON)
	if err != nil {
		return nil, nil, nil, err
	}
	return m, sm, rev, nil
}

func (s *businessServiceImpl) SaveModel(
	ctx context.Context,
	tenantID, businessID, createdBy string,
	expectedRevision int32,
	model *entity.SemanticModel,
	changeSummary string,
) (*entity.BusinessModelRevision, bool, error) {
	master, err := s.Get(ctx, tenantID, businessID)
	if err != nil {
		return nil, false, err
	}
	if expectedRevision != master.CurrentRevision {
		return nil, false, entity.ErrRevisionConflict
	}
	if model == nil {
		return nil, false, fmt.Errorf("%w: semantic_model required", entity.ErrInvalidModel)
	}
	if model.SchemaVersion == "" {
		model.SchemaVersion = entity.SemanticSchemaVersion
	}
	if err := ValidateSemanticModel(model); err != nil {
		return nil, false, err
	}
	digest, raw, err := ContentDigest(model)
	if err != nil {
		return nil, false, err
	}
	cur, err := s.repo.GetRevision(ctx, tenantID, businessID, master.CurrentRevision)
	if err != nil {
		return nil, false, err
	}
	if cur != nil && cur.ContentDigest == digest {
		return cur, true, nil // no change
	}
	next := master.CurrentRevision + 1
	now := time.Now().UTC()
	if changeSummary == "" {
		changeSummary = "semantic update"
	}
	rev := &entity.BusinessModelRevision{
		TenantID:          tenantID,
		BusinessID:        businessID,
		RevisionNo:        next,
		BaseRevisionNo:    master.CurrentRevision,
		SchemaVersion:     model.SchemaVersion,
		SemanticModelJSON: string(raw),
		ContentDigest:     digest,
		ChangeSummary:     changeSummary,
		CreatedBy:         createdBy,
		CreatedAt:         now,
	}
	err = s.repo.Transaction(ctx, func(txRepo repository.BusinessRepository) error {
		ok, err := txRepo.CASBumpRevision(ctx, tenantID, businessID, expectedRevision, next)
		if err != nil {
			return err
		}
		if !ok {
			return entity.ErrRevisionConflict
		}
		return txRepo.CreateRevision(ctx, rev)
	})
	if err != nil {
		return nil, false, err
	}
	return rev, false, nil
}

func (s *businessServiceImpl) ListRevisions(ctx context.Context, tenantID, businessID string) ([]*entity.BusinessModelRevision, error) {
	if _, err := s.Get(ctx, tenantID, businessID); err != nil {
		return nil, err
	}
	return s.repo.ListRevisions(ctx, tenantID, businessID)
}

func (s *businessServiceImpl) GetRevision(ctx context.Context, tenantID, businessID string, revNo int32) (*entity.BusinessModelRevision, *entity.SemanticModel, error) {
	if _, err := s.Get(ctx, tenantID, businessID); err != nil {
		return nil, nil, err
	}
	rev, err := s.repo.GetRevision(ctx, tenantID, businessID, revNo)
	if err != nil {
		return nil, nil, err
	}
	if rev == nil {
		return nil, nil, entity.ErrRevisionNotFound
	}
	sm, err := parseSemantic(rev.SemanticModelJSON)
	if err != nil {
		return nil, nil, err
	}
	return rev, sm, nil
}

func (s *businessServiceImpl) Diff(ctx context.Context, tenantID, businessID string, from, to int32) (*entity.BusinessModelDiff, *entity.BusinessImpactSummary, error) {
	_, fromSM, err := s.GetRevision(ctx, tenantID, businessID, from)
	if err != nil {
		return nil, nil, err
	}
	_, toSM, err := s.GetRevision(ctx, tenantID, businessID, to)
	if err != nil {
		return nil, nil, err
	}
	d := DiffSemanticModels(from, to, fromSM, toSM)
	return d, ImpactFromDiff(d), nil
}

func (s *businessServiceImpl) GetLayout(ctx context.Context, tenantID, businessID string) (*entity.BusinessModelLayout, *entity.ViewLayout, error) {
	if _, err := s.Get(ctx, tenantID, businessID); err != nil {
		return nil, nil, err
	}
	row, err := s.repo.GetLayout(ctx, tenantID, businessID)
	if err != nil {
		return nil, nil, err
	}
	if row == nil {
		return nil, nil, entity.ErrNotFound
	}
	vl, err := parseLayout(row.LayoutJSON)
	if err != nil {
		return nil, nil, err
	}
	return row, vl, nil
}

func (s *businessServiceImpl) SaveLayout(
	ctx context.Context,
	tenantID, businessID, updatedBy string,
	expectedLayoutRevision int32,
	basedOnModelRevision int32,
	layout *entity.ViewLayout,
) (*entity.BusinessModelLayout, error) {
	master, err := s.Get(ctx, tenantID, businessID)
	if err != nil {
		return nil, err
	}
	cur, err := s.repo.GetLayout(ctx, tenantID, businessID)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return nil, entity.ErrNotFound
	}
	if expectedLayoutRevision != cur.LayoutRevision {
		return nil, entity.ErrLayoutConflict
	}
	if layout == nil {
		return nil, fmt.Errorf("%w: layout required", entity.ErrInvalidModel)
	}
	if basedOnModelRevision == 0 {
		basedOnModelRevision = master.CurrentRevision
	}
	raw, err := json.Marshal(layout)
	if err != nil {
		return nil, err
	}
	next := cur.LayoutRevision + 1
	ok, err := s.repo.CASBumpLayout(ctx, tenantID, businessID, expectedLayoutRevision, next, basedOnModelRevision, string(raw), updatedBy)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, entity.ErrLayoutConflict
	}
	_ = s.repo.TouchUpdatedAt(ctx, tenantID, businessID)
	return s.repo.GetLayout(ctx, tenantID, businessID)
}

func parseSemantic(raw string) (*entity.SemanticModel, error) {
	var sm entity.SemanticModel
	if err := json.Unmarshal([]byte(raw), &sm); err != nil {
		return nil, fmt.Errorf("%w: %v", entity.ErrInvalidModel, err)
	}
	normalizeSemantic(&sm)
	return &sm, nil
}

func normalizeSemantic(m *entity.SemanticModel) {
	if m.Nodes == nil {
		m.Nodes = []entity.SemanticNode{}
	}
	if m.Edges == nil {
		m.Edges = []entity.SemanticEdge{}
	}
	if m.Rules == nil {
		m.Rules = []entity.BusinessRule{}
	}
	if m.States == nil {
		m.States = []entity.BusinessState{}
	}
	if m.EvidenceRefs == nil {
		m.EvidenceRefs = []string{}
	}
	if m.AssertionRefs == nil {
		m.AssertionRefs = []string{}
	}
}

func parseLayout(raw string) (*entity.ViewLayout, error) {
	var vl entity.ViewLayout
	if err := json.Unmarshal([]byte(raw), &vl); err != nil {
		return nil, fmt.Errorf("%w: %v", entity.ErrInvalidModel, err)
	}
	if vl.NodePositions == nil {
		vl.NodePositions = map[string]entity.NodePosition{}
	}
	return &vl, nil
}
