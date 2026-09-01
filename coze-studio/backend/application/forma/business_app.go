/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	assetentity "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
	assetrepo "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/repository"
	assetsvc "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/service"
	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	businessrepo "github.com/coze-dev/coze-studio/backend/domain/forma/business/repository"
	businesssvc "github.com/coze-dev/coze-studio/backend/domain/forma/business/service"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenantctx "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/context"
)

type BusinessDTO struct {
	BusinessID      string `json:"business_id"`
	AssetID         string `json:"asset_id"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	CurrentRevision int32  `json:"current_revision"`
	SchemaVersion   string `json:"schema_version"`
	UpdatedAt       string `json:"updated_at"`
	CreatedAt       string `json:"created_at"`
}

type CreateBusinessInput struct {
	Name          string                        `json:"name"`
	SemanticModel *businessentity.SemanticModel `json:"semantic_model"`
	ChangeSummary string                        `json:"change_summary"`
}

type PatchBusinessInput struct {
	Name string `json:"name"`
}

type PutModelInput struct {
	ExpectedRevision int32                         `json:"expected_revision"`
	SemanticModel    *businessentity.SemanticModel `json:"semantic_model"`
	ChangeSummary    string                        `json:"change_summary"`
}

type ModelResponse struct {
	BusinessID      string                        `json:"business_id"`
	CurrentRevision int32                         `json:"current_revision"`
	ContentDigest   string                        `json:"content_digest"`
	ChangeSummary   string                        `json:"change_summary"`
	NoChange        bool                          `json:"no_change,omitempty"`
	SemanticModel   *businessentity.SemanticModel `json:"semantic_model"`
}

type RevisionDTO struct {
	RevisionNo     int32  `json:"revision_no"`
	BaseRevisionNo int32  `json:"base_revision_no"`
	SchemaVersion  string `json:"schema_version"`
	ContentDigest  string `json:"content_digest"`
	ChangeSummary  string `json:"change_summary"`
	CreatedBy      string `json:"created_by"`
	CreatedAt      string `json:"created_at"`
}

type RevisionDetailResponse struct {
	Revision      *RevisionDTO                  `json:"revision"`
	SemanticModel *businessentity.SemanticModel `json:"semantic_model"`
	ReadOnly      bool                          `json:"read_only"`
}

type DiffResponse struct {
	Diff   *businessentity.BusinessModelDiff     `json:"diff"`
	Impact *businessentity.BusinessImpactSummary `json:"impact"`
}

type PutLayoutInput struct {
	ExpectedLayoutRevision int32                     `json:"expected_layout_revision"`
	BasedOnModelRevision   int32                     `json:"based_on_model_revision"`
	Layout                 *businessentity.ViewLayout `json:"layout"`
}

type LayoutResponse struct {
	BusinessID             string                     `json:"business_id"`
	LayoutRevision         int32                      `json:"layout_revision"`
	BasedOnModelRevision   int32                      `json:"based_on_model_revision"`
	Layout                 *businessentity.ViewLayout `json:"layout"`
}

func (s *ApplicationService) CreateBusiness(ctx context.Context, in *CreateBusinessInput) (*BusinessDTO, error) {
	tc, err := s.requireBusinessTenant(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil || in.Name == "" {
		return nil, formaerrors.BusinessInvalidModel("name required")
	}
	businessID := "biz_" + uuid.NewString()
	createdBy := tc.PrincipalID
	ownerID := tc.CozeUserID

	var dto *BusinessDTO
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ar := assetsvc.NewAssetRegistry(&assetsvc.Components{
			AssetRepo: assetrepo.NewAssetRefRepository(tx),
			CozeRepo:  assetrepo.NewCozeResourceRefRepository(tx),
			IDGen:     s.IDGen,
		})
		bs := businesssvc.NewBusinessService(&businesssvc.Components{
			Repo: businessrepo.NewBusinessRepository(tx),
		})
		asset, err := ar.CreateAsset(ctx, &assetsvc.CreateAssetRequest{
			TenantID:        tc.TenantID,
			AssetID:         businessID,
			Kind:            assetentity.AssetKindBusiness,
			Name:            in.Name,
			SemanticVersion: "0.1.0",
			SchemaVersion:   businessentity.SemanticSchemaVersion,
			OwnerID:         ownerID,
			CreatedBy:       ownerID,
		})
		if err != nil {
			return err
		}
		summary := in.ChangeSummary
		if summary == "" {
			summary = "create business"
		}
		master, _, _, err := bs.InitBusiness(ctx, tc.TenantID, businessID, asset.AssetID, createdBy, in.SemanticModel, summary)
		if err != nil {
			return err
		}
		dto = &BusinessDTO{
			BusinessID:      master.BusinessID,
			AssetID:         master.AssetID,
			Name:            asset.Name,
			Status:          string(asset.Status),
			CurrentRevision: master.CurrentRevision,
			SchemaVersion:   master.SchemaVersion,
			UpdatedAt:       master.UpdatedAt.UTC().Format(time.RFC3339Nano),
			CreatedAt:       master.CreatedAt.UTC().Format(time.RFC3339Nano),
		}
		return nil
	})
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return dto, nil
}

func (s *ApplicationService) ListBusinesses(ctx context.Context) ([]*BusinessDTO, error) {
	tc, err := s.requireBusinessTenant(ctx)
	if err != nil {
		return nil, err
	}
	masters, err := s.BusinessSVC.List(ctx, tc.TenantID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	assets, err := s.DomainSVC.ListAssetsByTenant(ctx, tc.TenantID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	byID := map[string]*assetentity.AssetRef{}
	for _, a := range assets {
		if a != nil && a.Kind == assetentity.AssetKindBusiness {
			byID[a.AssetID] = a
		}
	}
	out := make([]*BusinessDTO, 0, len(masters))
	for _, m := range masters {
		a := byID[m.AssetID]
		name := m.BusinessID
		status := string(assetentity.AssetStatusDraft)
		if a != nil {
			name = a.Name
			status = string(a.Status)
		}
		out = append(out, &BusinessDTO{
			BusinessID:      m.BusinessID,
			AssetID:         m.AssetID,
			Name:            name,
			Status:          status,
			CurrentRevision: m.CurrentRevision,
			SchemaVersion:   m.SchemaVersion,
			UpdatedAt:       m.UpdatedAt.UTC().Format(time.RFC3339Nano),
			CreatedAt:       m.CreatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	return out, nil
}

func (s *ApplicationService) GetBusiness(ctx context.Context, businessID string) (*BusinessDTO, error) {
	tc, err := s.requireBusinessTenant(ctx)
	if err != nil {
		return nil, err
	}
	m, err := s.BusinessSVC.Get(ctx, tc.TenantID, businessID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	asset, err := s.DomainSVC.GetAsset(ctx, tc.TenantID, businessID, 1)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	name := businessID
	status := string(assetentity.AssetStatusDraft)
	if asset != nil {
		name = asset.Name
		status = string(asset.Status)
	}
	return &BusinessDTO{
		BusinessID:      m.BusinessID,
		AssetID:         m.AssetID,
		Name:            name,
		Status:          status,
		CurrentRevision: m.CurrentRevision,
		SchemaVersion:   m.SchemaVersion,
		UpdatedAt:       m.UpdatedAt.UTC().Format(time.RFC3339Nano),
		CreatedAt:       m.CreatedAt.UTC().Format(time.RFC3339Nano),
	}, nil
}

func (s *ApplicationService) PatchBusiness(ctx context.Context, businessID string, in *PatchBusinessInput) (*BusinessDTO, error) {
	tc, err := s.requireBusinessTenant(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil || in.Name == "" {
		return nil, formaerrors.BusinessInvalidModel("name required")
	}
	if _, err := s.BusinessSVC.Get(ctx, tc.TenantID, businessID); err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	if _, err := s.DomainSVC.UpdateAssetName(ctx, tc.TenantID, businessID, 1, in.Name); err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return s.GetBusiness(ctx, businessID)
}

func (s *ApplicationService) ArchiveBusiness(ctx context.Context, businessID string) (*BusinessDTO, error) {
	tc, err := s.requireBusinessTenant(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.BusinessSVC.Get(ctx, tc.TenantID, businessID); err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	if _, err := s.DomainSVC.ArchiveAsset(ctx, tc.TenantID, businessID, 1); err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return s.GetBusiness(ctx, businessID)
}

func (s *ApplicationService) GetBusinessModel(ctx context.Context, businessID string) (*ModelResponse, error) {
	tc, err := s.requireBusinessTenant(ctx)
	if err != nil {
		return nil, err
	}
	_, sm, rev, err := s.BusinessSVC.GetModel(ctx, tc.TenantID, businessID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return &ModelResponse{
		BusinessID:      businessID,
		CurrentRevision: rev.RevisionNo,
		ContentDigest:   rev.ContentDigest,
		ChangeSummary:   rev.ChangeSummary,
		SemanticModel:   sm,
	}, nil
}

func (s *ApplicationService) PutBusinessModel(ctx context.Context, businessID string, in *PutModelInput) (*ModelResponse, error) {
	tc, err := s.requireBusinessTenant(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil || in.SemanticModel == nil {
		return nil, formaerrors.BusinessInvalidModel("semantic_model required")
	}
	rev, noChange, err := s.BusinessSVC.SaveModel(ctx, tc.TenantID, businessID, tc.PrincipalID, in.ExpectedRevision, in.SemanticModel, in.ChangeSummary)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	sm := in.SemanticModel
	if noChange {
		_, sm, _, _ = s.BusinessSVC.GetModel(ctx, tc.TenantID, businessID)
	}
	return &ModelResponse{
		BusinessID:      businessID,
		CurrentRevision: rev.RevisionNo,
		ContentDigest:   rev.ContentDigest,
		ChangeSummary:   rev.ChangeSummary,
		NoChange:        noChange,
		SemanticModel:   sm,
	}, nil
}

func (s *ApplicationService) ListBusinessRevisions(ctx context.Context, businessID string) ([]*RevisionDTO, error) {
	tc, err := s.requireBusinessTenant(ctx)
	if err != nil {
		return nil, err
	}
	revs, err := s.BusinessSVC.ListRevisions(ctx, tc.TenantID, businessID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	out := make([]*RevisionDTO, 0, len(revs))
	for _, r := range revs {
		out = append(out, toRevisionDTO(r))
	}
	return out, nil
}

func (s *ApplicationService) GetBusinessRevision(ctx context.Context, businessID string, revNo int32) (*RevisionDetailResponse, error) {
	tc, err := s.requireBusinessTenant(ctx)
	if err != nil {
		return nil, err
	}
	rev, sm, err := s.BusinessSVC.GetRevision(ctx, tc.TenantID, businessID, revNo)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	master, _ := s.BusinessSVC.Get(ctx, tc.TenantID, businessID)
	readOnly := master == nil || rev.RevisionNo != master.CurrentRevision
	return &RevisionDetailResponse{
		Revision:      toRevisionDTO(rev),
		SemanticModel: sm,
		ReadOnly:      readOnly,
	}, nil
}

func (s *ApplicationService) DiffBusiness(ctx context.Context, businessID string, from, to int32) (*DiffResponse, error) {
	tc, err := s.requireBusinessTenant(ctx)
	if err != nil {
		return nil, err
	}
	if from <= 0 || to <= 0 {
		return nil, formaerrors.BusinessInvalidModel("from and to revision required")
	}
	d, impact, err := s.BusinessSVC.Diff(ctx, tc.TenantID, businessID, from, to)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return &DiffResponse{Diff: d, Impact: impact}, nil
}

func (s *ApplicationService) GetBusinessLayout(ctx context.Context, businessID string) (*LayoutResponse, error) {
	tc, err := s.requireBusinessTenant(ctx)
	if err != nil {
		return nil, err
	}
	row, vl, err := s.BusinessSVC.GetLayout(ctx, tc.TenantID, businessID)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	return &LayoutResponse{
		BusinessID:           businessID,
		LayoutRevision:       row.LayoutRevision,
		BasedOnModelRevision: row.BasedOnModelRevision,
		Layout:               vl,
	}, nil
}

func (s *ApplicationService) PutBusinessLayout(ctx context.Context, businessID string, in *PutLayoutInput) (*LayoutResponse, error) {
	tc, err := s.requireBusinessTenant(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil || in.Layout == nil {
		return nil, formaerrors.BusinessInvalidModel("layout required")
	}
	row, err := s.BusinessSVC.SaveLayout(ctx, tc.TenantID, businessID, tc.PrincipalID, in.ExpectedLayoutRevision, in.BasedOnModelRevision, in.Layout)
	if err != nil {
		return nil, formaerrors.MapDomainError(err)
	}
	vl := in.Layout
	return &LayoutResponse{
		BusinessID:           businessID,
		LayoutRevision:       row.LayoutRevision,
		BasedOnModelRevision: row.BasedOnModelRevision,
		Layout:               vl,
	}, nil
}

func (s *ApplicationService) requireBusinessTenant(ctx context.Context) (*tenantctx.TenantContext, error) {
	tc, ok := tenantctx.FromContext(ctx)
	if !ok || tc == nil || tc.TenantID == "" {
		return nil, formaerrors.TenantRequired("tenant context required")
	}
	if _, err := s.requireMemberOf(ctx, tc.TenantID); err != nil {
		return nil, err
	}
	if s.BusinessSVC == nil || s.DB == nil {
		return nil, formaerrors.Internal("business service not initialized")
	}
	return tc, nil
}

func toRevisionDTO(r *businessentity.BusinessModelRevision) *RevisionDTO {
	return &RevisionDTO{
		RevisionNo:     r.RevisionNo,
		BaseRevisionNo: r.BaseRevisionNo,
		SchemaVersion:  r.SchemaVersion,
		ContentDigest:  r.ContentDigest,
		ChangeSummary:  r.ChangeSummary,
		CreatedBy:      r.CreatedBy,
		CreatedAt:      r.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}
