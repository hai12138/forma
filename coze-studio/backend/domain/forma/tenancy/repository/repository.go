/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/internal/dal"
)

type PrincipalRepository interface {
	Create(ctx context.Context, p *entity.Principal) error
	GetByPrincipalID(ctx context.Context, principalID string) (*entity.Principal, error)
	GetByCozeUserID(ctx context.Context, cozeUserID int64) (*entity.Principal, error)
	GetByProviderSubject(ctx context.Context, provider, externalSubject string) (*entity.Principal, error)
}

type TenantRepository interface {
	Create(ctx context.Context, t *entity.Tenant) error
	GetByTenantID(ctx context.Context, tenantID string) (*entity.Tenant, error)
	Update(ctx context.Context, t *entity.Tenant, expectedRevision int32) error
	ListByPrincipalID(ctx context.Context, principalID string) ([]*entity.Tenant, error)
}

type MembershipRepository interface {
	Create(ctx context.Context, m *entity.Membership) error
	Get(ctx context.Context, tenantID, principalID string) (*entity.Membership, error)
	UpdateRole(ctx context.Context, tenantID, principalID string, role entity.MembershipRole, expectedRevision int32) (*entity.Membership, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*entity.Membership, error)
	ListByPrincipal(ctx context.Context, principalID string) ([]*entity.Membership, error)
	SoftRemove(ctx context.Context, tenantID, principalID string) error
}

type SpaceRefRepository interface {
	Create(ctx context.Context, ref *entity.TenantSpaceRef) error
	ListByTenant(ctx context.Context, tenantID string) ([]*entity.TenantSpaceRef, error)
	GetActiveBySpaceID(ctx context.Context, cozeSpaceID int64) (*entity.TenantSpaceRef, error)
	Deactivate(ctx context.Context, tenantID string, cozeSpaceID int64) error
}

type AuditRepository interface {
	Create(ctx context.Context, e *entity.AuditEvent) error
}

func NewPrincipalRepository(db *gorm.DB) PrincipalRepository {
	return dal.NewPrincipalDAO(db)
}

func NewTenantRepository(db *gorm.DB) TenantRepository {
	return dal.NewTenantDAO(db)
}

func NewMembershipRepository(db *gorm.DB) MembershipRepository {
	return dal.NewMembershipDAO(db)
}

func NewSpaceRefRepository(db *gorm.DB) SpaceRefRepository {
	return dal.NewSpaceRefDAO(db)
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return dal.NewAuditDAO(db)
}
