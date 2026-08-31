/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/repository"
)

const (
	providerCoze = "coze"
)

type CreateTenantRequest struct {
	TenantKey        string
	Name             string
	DisplayName      string
	OwnerPrincipalID string
}

type AddMemberRequest struct {
	TenantID    string
	PrincipalID string
	Role        entity.MembershipRole
	CreatedBy   string
}

type BindSpaceRequest struct {
	TenantID    string
	CozeSpaceID int64
	Purpose     entity.SpacePurpose
}

type BootstrapResult struct {
	Principal  *entity.Principal
	Tenant     *entity.Tenant
	Membership *entity.Membership
	SpaceRef   *entity.TenantSpaceRef
	Created    bool
}

type TenancyService interface {
	ResolveOrCreatePrincipal(ctx context.Context, cozeUserID int64, displayName string) (*entity.Principal, error)

	CreateTenant(ctx context.Context, req *CreateTenantRequest) (*entity.Tenant, error)
	GetTenant(ctx context.Context, tenantID string) (*entity.Tenant, error)
	UpdateTenant(ctx context.Context, tenant *entity.Tenant, expectedRevision int32) (*entity.Tenant, error)
	ListTenantsForPrincipal(ctx context.Context, principalID string) ([]*entity.Tenant, error)

	AddMember(ctx context.Context, req *AddMemberRequest) (*entity.Membership, error)
	UpdateMemberRole(ctx context.Context, tenantID, principalID string, role entity.MembershipRole, expectedRevision int32) (*entity.Membership, error)
	RemoveMember(ctx context.Context, tenantID, principalID string) error
	ListMembers(ctx context.Context, tenantID string) ([]*entity.Membership, error)
	GetMembership(ctx context.Context, tenantID, principalID string) (*entity.Membership, error)
	ListMembershipsForPrincipal(ctx context.Context, principalID string) ([]*entity.Membership, error)

	BindSpace(ctx context.Context, req *BindSpaceRequest) (*entity.TenantSpaceRef, error)
	UnbindSpace(ctx context.Context, tenantID string, cozeSpaceID int64) error
	ListSpaces(ctx context.Context, tenantID string) ([]*entity.TenantSpaceRef, error)

	Bootstrap(ctx context.Context, cozeUserID int64, displayName string, defaultSpaceID int64) (*BootstrapResult, error)
	RecordAudit(ctx context.Context, event *entity.AuditEvent) error
}

type Components struct {
	PrincipalRepo  repository.PrincipalRepository
	TenantRepo     repository.TenantRepository
	MembershipRepo repository.MembershipRepository
	SpaceRefRepo   repository.SpaceRefRepository
	AuditRepo      repository.AuditRepository
}

type tenancyServiceImpl struct {
	principalRepo  repository.PrincipalRepository
	tenantRepo     repository.TenantRepository
	membershipRepo repository.MembershipRepository
	spaceRefRepo   repository.SpaceRefRepository
	auditRepo      repository.AuditRepository
}

func NewTenancyService(c *Components) TenancyService {
	return &tenancyServiceImpl{
		principalRepo:  c.PrincipalRepo,
		tenantRepo:     c.TenantRepo,
		membershipRepo: c.MembershipRepo,
		spaceRefRepo:   c.SpaceRefRepo,
		auditRepo:      c.AuditRepo,
	}
}

func (s *tenancyServiceImpl) ResolveOrCreatePrincipal(ctx context.Context, cozeUserID int64, displayName string) (*entity.Principal, error) {
	if cozeUserID == 0 {
		return nil, fmt.Errorf("coze_user_id is required")
	}
	subject := strconv.FormatInt(cozeUserID, 10)

	existing, err := s.principalRepo.GetByProviderSubject(ctx, providerCoze, subject)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	byUser, err := s.principalRepo.GetByCozeUserID(ctx, cozeUserID)
	if err != nil {
		return nil, err
	}
	if byUser != nil {
		return byUser, nil
	}

	now := time.Now().UTC()
	p := &entity.Principal{
		PrincipalID:     "prin_" + uuid.NewString(),
		PrincipalType:   entity.PrincipalTypeUser,
		Provider:        providerCoze,
		ExternalSubject: subject,
		CozeUserID:      cozeUserID,
		DisplayName:     displayName,
		Status:          entity.PrincipalStatusActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.principalRepo.Create(ctx, p); err != nil {
		// Race: another request may have created the same mapping.
		again, getErr := s.principalRepo.GetByProviderSubject(ctx, providerCoze, subject)
		if getErr != nil {
			return nil, err
		}
		if again != nil {
			return again, nil
		}
		return nil, err
	}
	return p, nil
}

func (s *tenancyServiceImpl) CreateTenant(ctx context.Context, req *CreateTenantRequest) (*entity.Tenant, error) {
	if req == nil || req.OwnerPrincipalID == "" || req.Name == "" {
		return nil, fmt.Errorf("owner_principal_id and name are required")
	}
	now := time.Now().UTC()
	tenantKey := req.TenantKey
	if tenantKey == "" {
		tenantKey = "tenant_" + uuid.NewString()
	}
	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Name
	}
	t := &entity.Tenant{
		TenantID:         "ten_" + uuid.NewString(),
		TenantKey:        tenantKey,
		Name:             req.Name,
		DisplayName:      displayName,
		Status:           entity.TenantStatusActive,
		OwnerPrincipalID: req.OwnerPrincipalID,
		Revision:         1,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.tenantRepo.Create(ctx, t); err != nil {
		return nil, err
	}
	_ = s.RecordAudit(ctx, &entity.AuditEvent{
		TenantID:    t.TenantID,
		PrincipalID: req.OwnerPrincipalID,
		Action:      entity.AuditTenantCreated,
		Resource:    t.TenantID,
		CreatedAt:   now,
	})
	return t, nil
}

func (s *tenancyServiceImpl) GetTenant(ctx context.Context, tenantID string) (*entity.Tenant, error) {
	return s.tenantRepo.GetByTenantID(ctx, tenantID)
}

func (s *tenancyServiceImpl) UpdateTenant(ctx context.Context, tenant *entity.Tenant, expectedRevision int32) (*entity.Tenant, error) {
	if tenant == nil || tenant.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if err := s.tenantRepo.Update(ctx, tenant, expectedRevision); err != nil {
		return nil, err
	}
	_ = s.RecordAudit(ctx, &entity.AuditEvent{
		TenantID:    tenant.TenantID,
		PrincipalID: tenant.OwnerPrincipalID,
		Action:      entity.AuditTenantUpdated,
		Resource:    tenant.TenantID,
		CreatedAt:   time.Now().UTC(),
	})
	return tenant, nil
}

func (s *tenancyServiceImpl) ListTenantsForPrincipal(ctx context.Context, principalID string) ([]*entity.Tenant, error) {
	return s.tenantRepo.ListByPrincipalID(ctx, principalID)
}

func (s *tenancyServiceImpl) AddMember(ctx context.Context, req *AddMemberRequest) (*entity.Membership, error) {
	if req == nil || req.TenantID == "" || req.PrincipalID == "" {
		return nil, fmt.Errorf("tenant_id and principal_id are required")
	}
	role := req.Role
	if role == "" {
		role = entity.RoleMember
	}
	now := time.Now().UTC()
	m := &entity.Membership{
		TenantID:    req.TenantID,
		PrincipalID: req.PrincipalID,
		Role:        role,
		Status:      entity.MembershipActive,
		Revision:    1,
		JoinedAt:    now,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.membershipRepo.Create(ctx, m); err != nil {
		return nil, err
	}
	_ = s.RecordAudit(ctx, &entity.AuditEvent{
		TenantID:    req.TenantID,
		PrincipalID: req.CreatedBy,
		Action:      entity.AuditMemberAdded,
		Resource:    req.PrincipalID,
		CreatedAt:   now,
	})
	return m, nil
}

func (s *tenancyServiceImpl) UpdateMemberRole(ctx context.Context, tenantID, principalID string, role entity.MembershipRole, expectedRevision int32) (*entity.Membership, error) {
	if tenantID == "" || principalID == "" {
		return nil, fmt.Errorf("tenant_id and principal_id are required")
	}
	m, err := s.membershipRepo.UpdateRole(ctx, tenantID, principalID, role, expectedRevision)
	if err != nil {
		return nil, err
	}
	_ = s.RecordAudit(ctx, &entity.AuditEvent{
		TenantID:    tenantID,
		PrincipalID: principalID,
		Action:      entity.AuditMemberRoleChanged,
		Resource:    string(role),
		CreatedAt:   time.Now().UTC(),
	})
	return m, nil
}

func (s *tenancyServiceImpl) RemoveMember(ctx context.Context, tenantID, principalID string) error {
	if err := s.membershipRepo.SoftRemove(ctx, tenantID, principalID); err != nil {
		return err
	}
	_ = s.RecordAudit(ctx, &entity.AuditEvent{
		TenantID:    tenantID,
		PrincipalID: principalID,
		Action:      entity.AuditMemberRemoved,
		Resource:    principalID,
		CreatedAt:   time.Now().UTC(),
	})
	return nil
}

func (s *tenancyServiceImpl) ListMembers(ctx context.Context, tenantID string) ([]*entity.Membership, error) {
	return s.membershipRepo.ListByTenant(ctx, tenantID)
}

func (s *tenancyServiceImpl) GetMembership(ctx context.Context, tenantID, principalID string) (*entity.Membership, error) {
	return s.membershipRepo.Get(ctx, tenantID, principalID)
}

func (s *tenancyServiceImpl) ListMembershipsForPrincipal(ctx context.Context, principalID string) ([]*entity.Membership, error) {
	return s.membershipRepo.ListByPrincipal(ctx, principalID)
}

func (s *tenancyServiceImpl) BindSpace(ctx context.Context, req *BindSpaceRequest) (*entity.TenantSpaceRef, error) {
	if req == nil || req.TenantID == "" || req.CozeSpaceID == 0 {
		return nil, fmt.Errorf("tenant_id and coze_space_id are required")
	}
	purpose := req.Purpose
	if purpose == "" {
		purpose = entity.SpacePurposeDefault
	}
	now := time.Now().UTC()
	ref := &entity.TenantSpaceRef{
		TenantID:    req.TenantID,
		CozeSpaceID: req.CozeSpaceID,
		Purpose:     purpose,
		Status:      entity.SpaceRefActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.spaceRefRepo.Create(ctx, ref); err != nil {
		return nil, err
	}
	_ = s.RecordAudit(ctx, &entity.AuditEvent{
		TenantID:  req.TenantID,
		Action:    entity.AuditSpaceBound,
		Resource:  strconv.FormatInt(req.CozeSpaceID, 10),
		CreatedAt: now,
	})
	return ref, nil
}

func (s *tenancyServiceImpl) UnbindSpace(ctx context.Context, tenantID string, cozeSpaceID int64) error {
	if err := s.spaceRefRepo.Deactivate(ctx, tenantID, cozeSpaceID); err != nil {
		return err
	}
	_ = s.RecordAudit(ctx, &entity.AuditEvent{
		TenantID:  tenantID,
		Action:    entity.AuditSpaceUnbound,
		Resource:  strconv.FormatInt(cozeSpaceID, 10),
		CreatedAt: time.Now().UTC(),
	})
	return nil
}

func (s *tenancyServiceImpl) ListSpaces(ctx context.Context, tenantID string) ([]*entity.TenantSpaceRef, error) {
	return s.spaceRefRepo.ListByTenant(ctx, tenantID)
}

func (s *tenancyServiceImpl) Bootstrap(ctx context.Context, cozeUserID int64, displayName string, defaultSpaceID int64) (*BootstrapResult, error) {
	principal, err := s.ResolveOrCreatePrincipal(ctx, cozeUserID, displayName)
	if err != nil {
		return nil, err
	}

	personalKey := personalTenantKey(cozeUserID)
	memberships, err := s.membershipRepo.ListByPrincipal(ctx, principal.PrincipalID)
	if err != nil {
		return nil, err
	}
	for _, m := range memberships {
		if m.Role != entity.RoleOwner || m.Status != entity.MembershipActive {
			continue
		}
		tenant, getErr := s.tenantRepo.GetByTenantID(ctx, m.TenantID)
		if getErr != nil {
			return nil, getErr
		}
		if tenant == nil || tenant.Status != entity.TenantStatusActive {
			continue
		}
		if tenant.TenantKey != personalKey && !strings.HasPrefix(tenant.TenantKey, "personal_") {
			continue
		}
		// Prefer exact personal key match; also accept any personal_* owned as OWNER.
		if tenant.TenantKey != personalKey && tenant.OwnerPrincipalID != principal.PrincipalID {
			continue
		}
		spaces, listErr := s.spaceRefRepo.ListByTenant(ctx, tenant.TenantID)
		if listErr != nil {
			return nil, listErr
		}
		var spaceRef *entity.TenantSpaceRef
		for _, sp := range spaces {
			if sp.Status == entity.SpaceRefActive && sp.Purpose == entity.SpacePurposeDefault {
				spaceRef = sp
				break
			}
		}
		if spaceRef == nil && len(spaces) > 0 {
			spaceRef = spaces[0]
		}
		return &BootstrapResult{
			Principal:  principal,
			Tenant:     tenant,
			Membership: m,
			SpaceRef:   spaceRef,
			Created:    false,
		}, nil
	}

	name := displayName
	if name == "" {
		name = "Personal Workspace"
	} else {
		name = name + "'s Workspace"
	}

	tenant, err := s.CreateTenant(ctx, &CreateTenantRequest{
		TenantKey:        personalKey,
		Name:             name,
		DisplayName:      name,
		OwnerPrincipalID: principal.PrincipalID,
	})
	if err != nil {
		return nil, err
	}

	membership, err := s.AddMember(ctx, &AddMemberRequest{
		TenantID:    tenant.TenantID,
		PrincipalID: principal.PrincipalID,
		Role:        entity.RoleOwner,
		CreatedBy:   principal.PrincipalID,
	})
	if err != nil {
		return nil, err
	}

	var spaceRef *entity.TenantSpaceRef
	if defaultSpaceID > 0 {
		spaceRef, err = s.BindSpace(ctx, &BindSpaceRequest{
			TenantID:    tenant.TenantID,
			CozeSpaceID: defaultSpaceID,
			Purpose:     entity.SpacePurposeDefault,
		})
		if err != nil {
			return nil, err
		}
	}

	return &BootstrapResult{
		Principal:  principal,
		Tenant:     tenant,
		Membership: membership,
		SpaceRef:   spaceRef,
		Created:    true,
	}, nil
}

func (s *tenancyServiceImpl) RecordAudit(ctx context.Context, event *entity.AuditEvent) error {
	if event == nil || event.Action == "" {
		return fmt.Errorf("action is required")
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	return s.auditRepo.Create(ctx, event)
}

func personalTenantKey(cozeUserID int64) string {
	return fmt.Sprintf("personal_%d", cozeUserID)
}
