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
	TenantID         string
	CozeSpaceID      int64
	Purpose          entity.SpacePurpose
	ActorPrincipalID string
	RequestID        string
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
	UpdateTenant(ctx context.Context, tenant *entity.Tenant, expectedRevision int32, actorPrincipalID, requestID string) (*entity.Tenant, error)
	ListTenantsForPrincipal(ctx context.Context, principalID string) ([]*entity.Tenant, error)

	AddMember(ctx context.Context, req *AddMemberRequest) (*entity.Membership, error)
	UpdateMemberRole(ctx context.Context, tenantID, principalID string, role entity.MembershipRole, expectedRevision int32, actorPrincipalID, requestID string) (*entity.Membership, error)
	RemoveMember(ctx context.Context, tenantID, principalID, actorPrincipalID, requestID string) error
	ListMembers(ctx context.Context, tenantID string) ([]*entity.Membership, error)
	GetMembership(ctx context.Context, tenantID, principalID string) (*entity.Membership, error)
	ListMembershipsForPrincipal(ctx context.Context, principalID string) ([]*entity.Membership, error)

	BindSpace(ctx context.Context, req *BindSpaceRequest) (*entity.TenantSpaceRef, error)
	UnbindSpace(ctx context.Context, tenantID string, cozeSpaceID int64, actorPrincipalID, requestID string) error
	ListSpaces(ctx context.Context, tenantID string) ([]*entity.TenantSpaceRef, error)

	Bootstrap(ctx context.Context, cozeUserID int64, displayName string, defaultSpaceID int64) (*BootstrapResult, error)
	RecordAudit(ctx context.Context, event *entity.AuditEvent) error

	// Platform admin
	GetPlatformRole(ctx context.Context, principalID string) (*entity.FormaPlatformRoleAssignment, error)
	SetPlatformRole(ctx context.Context, principalID string, role entity.PlatformRole) error
	SetPasswordChangeRequired(ctx context.Context, principalID string, required bool) error
	ListPlatformRoles(ctx context.Context) ([]*entity.FormaPlatformRoleAssignment, error)
	CountActiveSuperAdmins(ctx context.Context) (int, error)
	SuspendPrincipal(ctx context.Context, principalID string) error
	ActivatePrincipal(ctx context.Context, principalID string) error
	GetPrincipalByID(ctx context.Context, principalID string) (*entity.Principal, error)
	ListAllPrincipals(ctx context.Context) ([]*entity.Principal, error)
}

type Components struct {
	PrincipalRepo    repository.PrincipalRepository
	TenantRepo       repository.TenantRepository
	MembershipRepo   repository.MembershipRepository
	SpaceRefRepo     repository.SpaceRefRepository
	AuditRepo        repository.AuditRepository
	PlatformRoleRepo repository.PlatformRoleRepository
}

type tenancyServiceImpl struct {
	principalRepo    repository.PrincipalRepository
	tenantRepo       repository.TenantRepository
	membershipRepo   repository.MembershipRepository
	spaceRefRepo     repository.SpaceRefRepository
	auditRepo        repository.AuditRepository
	platformRoleRepo repository.PlatformRoleRepository
}

func NewTenancyService(c *Components) TenancyService {
	return &tenancyServiceImpl{
		principalRepo:    c.PrincipalRepo,
		tenantRepo:       c.TenantRepo,
		membershipRepo:   c.MembershipRepo,
		spaceRefRepo:     c.SpaceRefRepo,
		auditRepo:        c.AuditRepo,
		platformRoleRepo: c.PlatformRoleRepo,
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

func (s *tenancyServiceImpl) UpdateTenant(ctx context.Context, tenant *entity.Tenant, expectedRevision int32, actorPrincipalID, requestID string) (*entity.Tenant, error) {
	if tenant == nil || tenant.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if err := s.tenantRepo.Update(ctx, tenant, expectedRevision); err != nil {
		return nil, err
	}
	_ = s.RecordAudit(ctx, &entity.AuditEvent{
		TenantID:    tenant.TenantID,
		PrincipalID: actorPrincipalID,
		Action:      entity.AuditTenantUpdated,
		Resource:    tenant.TenantID,
		RequestID:   requestID,
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
	if !ValidRole(role) {
		return nil, entity.ErrInvalidRole
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

func (s *tenancyServiceImpl) UpdateMemberRole(ctx context.Context, tenantID, principalID string, role entity.MembershipRole, expectedRevision int32, actorPrincipalID, requestID string) (*entity.Membership, error) {
	if tenantID == "" || principalID == "" {
		return nil, fmt.Errorf("tenant_id and principal_id are required")
	}
	if !ValidRole(role) {
		return nil, entity.ErrInvalidRole
	}

	current, err := s.membershipRepo.Get(ctx, tenantID, principalID)
	if err != nil {
		return nil, err
	}
	if current == nil || current.Status != entity.MembershipActive {
		return nil, entity.ErrNotFound
	}

	tenant, err := s.tenantRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, entity.ErrNotFound
	}

	if principalID == tenant.OwnerPrincipalID && role != entity.RoleOwner {
		return nil, entity.ErrPrimaryOwnerImmutable
	}

	if current.Role == entity.RoleOwner && role != entity.RoleOwner {
		owners, countErr := s.countActiveOwners(ctx, tenantID)
		if countErr != nil {
			return nil, countErr
		}
		if owners <= 1 {
			return nil, entity.ErrLastOwner
		}
	}

	m, err := s.membershipRepo.UpdateRole(ctx, tenantID, principalID, role, expectedRevision)
	if err != nil {
		return nil, err
	}
	_ = s.RecordAudit(ctx, &entity.AuditEvent{
		TenantID:    tenantID,
		PrincipalID: actorPrincipalID,
		Action:      entity.AuditMemberRoleChanged,
		Resource:    principalID,
		RequestID:   requestID,
		CreatedAt:   time.Now().UTC(),
	})
	return m, nil
}

func (s *tenancyServiceImpl) RemoveMember(ctx context.Context, tenantID, principalID, actorPrincipalID, requestID string) error {
	if tenantID == "" || principalID == "" {
		return fmt.Errorf("tenant_id and principal_id are required")
	}
	current, err := s.membershipRepo.Get(ctx, tenantID, principalID)
	if err != nil {
		return err
	}
	if current == nil || current.Status != entity.MembershipActive {
		return entity.ErrNotFound
	}

	tenant, err := s.tenantRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}
	if tenant == nil {
		return entity.ErrNotFound
	}
	if principalID == tenant.OwnerPrincipalID {
		return entity.ErrPrimaryOwnerImmutable
	}
	if current.Role == entity.RoleOwner {
		owners, countErr := s.countActiveOwners(ctx, tenantID)
		if countErr != nil {
			return countErr
		}
		if owners <= 1 {
			return entity.ErrLastOwner
		}
	}

	if err := s.membershipRepo.SoftRemove(ctx, tenantID, principalID); err != nil {
		return err
	}
	_ = s.RecordAudit(ctx, &entity.AuditEvent{
		TenantID:    tenantID,
		PrincipalID: actorPrincipalID,
		Action:      entity.AuditMemberRemoved,
		Resource:    principalID,
		RequestID:   requestID,
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

	existing, err := s.spaceRefRepo.GetBySpaceID(ctx, req.CozeSpaceID)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.Status == entity.SpaceRefActive && existing.TenantID != req.TenantID {
		return nil, entity.ErrSpaceOwned
	}

	ref := &entity.TenantSpaceRef{
		TenantID:    req.TenantID,
		CozeSpaceID: req.CozeSpaceID,
		Purpose:     purpose,
		Status:      entity.SpaceRefActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if existing != nil {
		ref.ID = existing.ID
		ref.CreatedAt = existing.CreatedAt
	}
	if err := s.spaceRefRepo.UpsertBind(ctx, ref); err != nil {
		return nil, err
	}
	_ = s.RecordAudit(ctx, &entity.AuditEvent{
		TenantID:    req.TenantID,
		PrincipalID: req.ActorPrincipalID,
		Action:      entity.AuditSpaceBound,
		Resource:    strconv.FormatInt(req.CozeSpaceID, 10),
		RequestID:   req.RequestID,
		CreatedAt:   now,
	})
	return ref, nil
}

func (s *tenancyServiceImpl) UnbindSpace(ctx context.Context, tenantID string, cozeSpaceID int64, actorPrincipalID, requestID string) error {
	if err := s.spaceRefRepo.Deactivate(ctx, tenantID, cozeSpaceID); err != nil {
		return err
	}
	_ = s.RecordAudit(ctx, &entity.AuditEvent{
		TenantID:    tenantID,
		PrincipalID: actorPrincipalID,
		Action:      entity.AuditSpaceUnbound,
		Resource:    strconv.FormatInt(cozeSpaceID, 10),
		RequestID:   requestID,
		CreatedAt:   time.Now().UTC(),
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
			TenantID:         tenant.TenantID,
			CozeSpaceID:      defaultSpaceID,
			Purpose:          entity.SpacePurposeDefault,
			ActorPrincipalID: principal.PrincipalID,
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

func (s *tenancyServiceImpl) countActiveOwners(ctx context.Context, tenantID string) (int, error) {
	members, err := s.membershipRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, m := range members {
		if m != nil && m.Status == entity.MembershipActive && m.Role == entity.RoleOwner {
			n++
		}
	}
	return n, nil
}

func (s *tenancyServiceImpl) GetPlatformRole(ctx context.Context, principalID string) (*entity.FormaPlatformRoleAssignment, error) {
	if s.platformRoleRepo == nil {
		return nil, nil
	}
	return s.platformRoleRepo.GetByPrincipalID(ctx, principalID)
}

func (s *tenancyServiceImpl) SetPlatformRole(ctx context.Context, principalID string, role entity.PlatformRole) error {
	if s.platformRoleRepo == nil {
		return fmt.Errorf("platform role repository not initialized")
	}
	now := time.Now().UTC()
	existing, err := s.platformRoleRepo.GetByPrincipalID(ctx, principalID)
	if err != nil {
		return err
	}
	if existing != nil {
		existing.Role = role
		return s.platformRoleRepo.Update(ctx, existing)
	}
	return s.platformRoleRepo.Create(ctx, &entity.FormaPlatformRoleAssignment{
		PrincipalID: principalID,
		Role:        role,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

func (s *tenancyServiceImpl) SetPasswordChangeRequired(ctx context.Context, principalID string, required bool) error {
	if s.platformRoleRepo == nil {
		return fmt.Errorf("platform role repository not initialized")
	}
	existing, err := s.platformRoleRepo.GetByPrincipalID(ctx, principalID)
	if err != nil {
		return err
	}
	if existing == nil {
		return entity.ErrNotFound
	}
	existing.PasswordChangeRequired = required
	return s.platformRoleRepo.Update(ctx, existing)
}

func (s *tenancyServiceImpl) ListPlatformRoles(ctx context.Context) ([]*entity.FormaPlatformRoleAssignment, error) {
	if s.platformRoleRepo == nil {
		return nil, nil
	}
	return s.platformRoleRepo.ListAll(ctx)
}

func (s *tenancyServiceImpl) CountActiveSuperAdmins(ctx context.Context) (int, error) {
	if s.platformRoleRepo == nil {
		return 0, nil
	}
	admins, err := s.platformRoleRepo.ListSuperAdmins(ctx)
	if err != nil {
		return 0, err
	}
	activeIDs := make([]string, 0, len(admins))
	for _, a := range admins {
		p, pErr := s.principalRepo.GetByPrincipalID(ctx, a.PrincipalID)
		if pErr != nil {
			return 0, pErr
		}
		if p != nil && p.Status == entity.PrincipalStatusActive {
			activeIDs = append(activeIDs, a.PrincipalID)
		}
	}
	return len(activeIDs), nil
}

func (s *tenancyServiceImpl) SuspendPrincipal(ctx context.Context, principalID string) error {
	return s.principalRepo.UpdateStatus(ctx, principalID, entity.PrincipalStatusSuspended)
}

func (s *tenancyServiceImpl) ActivatePrincipal(ctx context.Context, principalID string) error {
	return s.principalRepo.UpdateStatus(ctx, principalID, entity.PrincipalStatusActive)
}

func (s *tenancyServiceImpl) GetPrincipalByID(ctx context.Context, principalID string) (*entity.Principal, error) {
	return s.principalRepo.GetByPrincipalID(ctx, principalID)
}

func (s *tenancyServiceImpl) ListAllPrincipals(ctx context.Context) ([]*entity.Principal, error) {
	return s.principalRepo.ListAll(ctx)
}

func personalTenantKey(cozeUserID int64) string {
	return fmt.Sprintf("personal_%d", cozeUserID)
}
