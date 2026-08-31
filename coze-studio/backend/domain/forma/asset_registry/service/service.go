/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/repository"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
)

type CreateAssetRequest struct {
	TenantID        string
	AssetID         string
	Kind            entity.AssetKind
	Name            string
	SemanticVersion string
	SchemaVersion   string
	OwnerID         int64
	CreatedBy       int64
	ContentDigest   string
}

type BindCozeResourceRequest struct {
	TenantID         string
	AssetID          string
	AssetRevision    int32
	CozeResourceType entity.CozeResourceType
	CozeResourceID   int64
	CozeSpaceID      *int64
	CozeVersion      string
}

type AssetRegistry interface {
	CreateAsset(ctx context.Context, req *CreateAssetRequest) (*entity.AssetRef, error)
	GetAsset(ctx context.Context, tenantID, assetID string, revision int32) (*entity.AssetRef, error)
	ListAssetsByTenant(ctx context.Context, tenantID string) ([]*entity.AssetRef, error)
	BindCozeResource(ctx context.Context, req *BindCozeResourceRequest) (*entity.CozeResourceRef, error)
	ListCozeResources(ctx context.Context, tenantID, assetID string, revision int32) ([]*entity.CozeResourceRef, error)
}

type assetRegistryImpl struct {
	assetRepo repository.AssetRefRepository
	cozeRepo  repository.CozeResourceRefRepository
	idGen     idgen.IDGenerator
}

type Components struct {
	AssetRepo repository.AssetRefRepository
	CozeRepo  repository.CozeResourceRefRepository
	IDGen     idgen.IDGenerator
}

func NewAssetRegistry(c *Components) AssetRegistry {
	return &assetRegistryImpl{
		assetRepo: c.AssetRepo,
		cozeRepo:  c.CozeRepo,
		idGen:     c.IDGen,
	}
}

func (s *assetRegistryImpl) CreateAsset(ctx context.Context, req *CreateAssetRequest) (*entity.AssetRef, error) {
	if req.TenantID == "" || req.AssetID == "" || req.Name == "" {
		return nil, fmt.Errorf("tenant_id, asset_id and name are required")
	}
	if req.SemanticVersion == "" {
		req.SemanticVersion = "0.0.0"
	}
	if req.SchemaVersion == "" {
		req.SchemaVersion = "1.0"
	}
	now := time.Now().UTC()
	asset := &entity.AssetRef{
		TenantID:        req.TenantID,
		AssetID:         req.AssetID,
		Kind:            req.Kind,
		Name:            req.Name,
		SemanticVersion: req.SemanticVersion,
		Revision:        1,
		SchemaVersion:   req.SchemaVersion,
		Status:          entity.AssetStatusDraft,
		OwnerID:         req.OwnerID,
		CreatedBy:       req.CreatedBy,
		ContentDigest:   req.ContentDigest,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.assetRepo.Create(ctx, asset); err != nil {
		return nil, err
	}
	return asset, nil
}

func (s *assetRegistryImpl) GetAsset(ctx context.Context, tenantID, assetID string, revision int32) (*entity.AssetRef, error) {
	return s.assetRepo.GetByTenantAssetRevision(ctx, tenantID, assetID, revision)
}

func (s *assetRegistryImpl) ListAssetsByTenant(ctx context.Context, tenantID string) ([]*entity.AssetRef, error) {
	return s.assetRepo.ListByTenant(ctx, tenantID)
}

func (s *assetRegistryImpl) BindCozeResource(ctx context.Context, req *BindCozeResourceRequest) (*entity.CozeResourceRef, error) {
	if req.TenantID == "" || req.AssetID == "" || req.CozeResourceID == 0 {
		return nil, fmt.Errorf("tenant_id, asset_id and coze_resource_id are required")
	}
	now := time.Now().UTC()
	ref := &entity.CozeResourceRef{
		TenantID:         req.TenantID,
		AssetID:          req.AssetID,
		AssetRevision:    req.AssetRevision,
		CozeResourceType: req.CozeResourceType,
		CozeResourceID:   req.CozeResourceID,
		CozeSpaceID:      req.CozeSpaceID,
		CozeVersion:      req.CozeVersion,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.cozeRepo.Create(ctx, ref); err != nil {
		return nil, err
	}
	return ref, nil
}

func (s *assetRegistryImpl) ListCozeResources(ctx context.Context, tenantID, assetID string, revision int32) ([]*entity.CozeResourceRef, error) {
	return s.cozeRepo.ListByAsset(ctx, tenantID, assetID, revision)
}
