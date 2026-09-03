/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/service"
)

// ---------------------------------------------------------------------------
// In-memory fakes
// ---------------------------------------------------------------------------

type memPrincipalRepo struct {
	mu   sync.Mutex
	byID map[string]*entity.Principal
	seq  int64
}

func newMemPrincipalRepo() *memPrincipalRepo {
	return &memPrincipalRepo{byID: make(map[string]*entity.Principal)}
}

func (r *memPrincipalRepo) Create(_ context.Context, p *entity.Principal) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.byID {
		if existing.Provider == p.Provider && existing.ExternalSubject == p.ExternalSubject {
			return fmt.Errorf("duplicate provider subject")
		}
		if existing.CozeUserID == p.CozeUserID {
			return fmt.Errorf("duplicate coze user")
		}
	}
	r.seq++
	cp := *p
	cp.ID = r.seq
	r.byID[cp.PrincipalID] = &cp
	p.ID = cp.ID
	return nil
}

func (r *memPrincipalRepo) GetByPrincipalID(_ context.Context, principalID string) (*entity.Principal, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.byID[principalID]
	if !ok {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}

func (r *memPrincipalRepo) GetByCozeUserID(_ context.Context, cozeUserID int64) (*entity.Principal, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.byID {
		if p.CozeUserID == cozeUserID {
			cp := *p
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *memPrincipalRepo) GetByProviderSubject(_ context.Context, provider, externalSubject string) (*entity.Principal, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.byID {
		if p.Provider == provider && p.ExternalSubject == externalSubject {
			cp := *p
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *memPrincipalRepo) UpdateStatus(_ context.Context, principalID string, status entity.PrincipalStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.byID[principalID]
	if !ok {
		return entity.ErrNotFound
	}
	p.Status = status
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *memPrincipalRepo) ListAll(_ context.Context) ([]*entity.Principal, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*entity.Principal, 0, len(r.byID))
	for _, p := range r.byID {
		cp := *p
		out = append(out, &cp)
	}
	return out, nil
}

type memPlatformRoleRepo struct {
	mu   sync.Mutex
	byID map[string]*entity.FormaPlatformRoleAssignment
	seq  int64
}

func newMemPlatformRoleRepo() *memPlatformRoleRepo {
	return &memPlatformRoleRepo{byID: make(map[string]*entity.FormaPlatformRoleAssignment)}
}

func (r *memPlatformRoleRepo) GetByPrincipalID(_ context.Context, principalID string) (*entity.FormaPlatformRoleAssignment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.byID[principalID]
	if !ok {
		return nil, nil
	}
	cp := *a
	return &cp, nil
}

func (r *memPlatformRoleRepo) Create(_ context.Context, assignment *entity.FormaPlatformRoleAssignment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[assignment.PrincipalID]; ok {
		return fmt.Errorf("duplicate platform role")
	}
	r.seq++
	cp := *assignment
	cp.ID = r.seq
	r.byID[cp.PrincipalID] = &cp
	assignment.ID = cp.ID
	return nil
}

func (r *memPlatformRoleRepo) Update(_ context.Context, assignment *entity.FormaPlatformRoleAssignment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cur, ok := r.byID[assignment.PrincipalID]
	if !ok {
		return entity.ErrNotFound
	}
	cp := *assignment
	cp.ID = cur.ID
	cp.UpdatedAt = time.Now().UTC()
	r.byID[cp.PrincipalID] = &cp
	assignment.UpdatedAt = cp.UpdatedAt
	return nil
}

func (r *memPlatformRoleRepo) ListSuperAdmins(_ context.Context) ([]*entity.FormaPlatformRoleAssignment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*entity.FormaPlatformRoleAssignment
	for _, a := range r.byID {
		if a.Role == entity.PlatformRoleSuperAdmin {
			cp := *a
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *memPlatformRoleRepo) ListAll(_ context.Context) ([]*entity.FormaPlatformRoleAssignment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*entity.FormaPlatformRoleAssignment, 0, len(r.byID))
	for _, a := range r.byID {
		cp := *a
		out = append(out, &cp)
	}
	return out, nil
}

func (r *memPlatformRoleRepo) CountActiveSuperAdmins(_ context.Context, activePrincipalIDs []string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	active := make(map[string]struct{}, len(activePrincipalIDs))
	for _, id := range activePrincipalIDs {
		active[id] = struct{}{}
	}
	n := 0
	for _, a := range r.byID {
		if a.Role == entity.PlatformRoleSuperAdmin {
			if _, ok := active[a.PrincipalID]; ok {
				n++
			}
		}
	}
	return n, nil
}

type memTenantRepo struct {
	mu   sync.Mutex
	byID map[string]*entity.Tenant
	seq  int64
}

func newMemTenantRepo() *memTenantRepo {
	return &memTenantRepo{byID: make(map[string]*entity.Tenant)}
}

func (r *memTenantRepo) Create(_ context.Context, t *entity.Tenant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.byID {
		if existing.TenantKey == t.TenantKey {
			return fmt.Errorf("duplicate tenant key")
		}
	}
	r.seq++
	cp := *t
	cp.ID = r.seq
	r.byID[cp.TenantID] = &cp
	t.ID = cp.ID
	return nil
}

func (r *memTenantRepo) GetByTenantID(_ context.Context, tenantID string) (*entity.Tenant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.byID[tenantID]
	if !ok || t.DeletedAt != nil {
		return nil, nil
	}
	cp := *t
	return &cp, nil
}

func (r *memTenantRepo) Update(_ context.Context, t *entity.Tenant, expectedRevision int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cur, ok := r.byID[t.TenantID]
	if !ok || cur.DeletedAt != nil {
		return entity.ErrNotFound
	}
	if cur.Revision != expectedRevision {
		return entity.ErrRevisionConflict
	}
	cp := *t
	cp.ID = cur.ID
	cp.Revision = expectedRevision + 1
	cp.UpdatedAt = time.Now().UTC()
	r.byID[t.TenantID] = &cp
	t.Revision = cp.Revision
	t.UpdatedAt = cp.UpdatedAt
	return nil
}

func (r *memTenantRepo) ListByPrincipalID(ctx context.Context, principalID string) ([]*entity.Tenant, error) {
	// Populated via membership join in real DAO; tests use membership list + Get.
	_ = ctx
	_ = principalID
	return nil, nil
}

type memMembershipRepo struct {
	mu   sync.Mutex
	rows []*entity.Membership
	seq  int64
}

func newMemMembershipRepo() *memMembershipRepo {
	return &memMembershipRepo{}
}

func (r *memMembershipRepo) Create(_ context.Context, m *entity.Membership) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.rows {
		if existing.TenantID == m.TenantID && existing.PrincipalID == m.PrincipalID {
			return fmt.Errorf("duplicate membership")
		}
	}
	r.seq++
	cp := *m
	cp.ID = r.seq
	r.rows = append(r.rows, &cp)
	m.ID = cp.ID
	return nil
}

func (r *memMembershipRepo) Get(_ context.Context, tenantID, principalID string) (*entity.Membership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, m := range r.rows {
		if m.TenantID == tenantID && m.PrincipalID == principalID {
			cp := *m
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *memMembershipRepo) UpdateRole(_ context.Context, tenantID, principalID string, role entity.MembershipRole, expectedRevision int32) (*entity.Membership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, m := range r.rows {
		if m.TenantID == tenantID && m.PrincipalID == principalID {
			if m.Revision != expectedRevision {
				return nil, entity.ErrRevisionConflict
			}
			m.Role = role
			m.Revision = expectedRevision + 1
			m.UpdatedAt = time.Now().UTC()
			cp := *m
			return &cp, nil
		}
	}
	return nil, entity.ErrNotFound
}

func (r *memMembershipRepo) ListByTenant(_ context.Context, tenantID string) ([]*entity.Membership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*entity.Membership
	for _, m := range r.rows {
		if m.TenantID == tenantID {
			cp := *m
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *memMembershipRepo) ListByPrincipal(_ context.Context, principalID string) ([]*entity.Membership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*entity.Membership
	for _, m := range r.rows {
		if m.PrincipalID == principalID {
			cp := *m
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *memMembershipRepo) SoftRemove(_ context.Context, tenantID, principalID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, m := range r.rows {
		if m.TenantID == tenantID && m.PrincipalID == principalID {
			m.Status = entity.MembershipRemoved
			m.UpdatedAt = time.Now().UTC()
			return nil
		}
	}
	return entity.ErrNotFound
}

type memSpaceRefRepo struct {
	mu   sync.Mutex
	rows []*entity.TenantSpaceRef
	seq  int64
}

func newMemSpaceRefRepo() *memSpaceRefRepo {
	return &memSpaceRefRepo{}
}

func (r *memSpaceRefRepo) GetBySpaceID(_ context.Context, cozeSpaceID int64) (*entity.TenantSpaceRef, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, sp := range r.rows {
		if sp.CozeSpaceID == cozeSpaceID {
			cp := *sp
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *memSpaceRefRepo) UpsertBind(_ context.Context, ref *entity.TenantSpaceRef) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	for _, sp := range r.rows {
		if sp.CozeSpaceID == ref.CozeSpaceID {
			sp.TenantID = ref.TenantID
			sp.Purpose = ref.Purpose
			sp.Status = entity.SpaceRefActive
			sp.UpdatedAt = now
			ref.ID = sp.ID
			ref.Status = entity.SpaceRefActive
			ref.CreatedAt = sp.CreatedAt
			ref.UpdatedAt = now
			return nil
		}
	}
	r.seq++
	cp := *ref
	cp.ID = r.seq
	cp.Status = entity.SpaceRefActive
	if cp.CreatedAt.IsZero() {
		cp.CreatedAt = now
	}
	cp.UpdatedAt = now
	r.rows = append(r.rows, &cp)
	ref.ID = cp.ID
	ref.Status = cp.Status
	ref.CreatedAt = cp.CreatedAt
	ref.UpdatedAt = cp.UpdatedAt
	return nil
}

func (r *memSpaceRefRepo) ListByTenant(_ context.Context, tenantID string) ([]*entity.TenantSpaceRef, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*entity.TenantSpaceRef
	for _, sp := range r.rows {
		if sp.TenantID == tenantID {
			cp := *sp
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *memSpaceRefRepo) Deactivate(_ context.Context, tenantID string, cozeSpaceID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, sp := range r.rows {
		if sp.TenantID == tenantID && sp.CozeSpaceID == cozeSpaceID {
			sp.Status = entity.SpaceRefInactive
			sp.UpdatedAt = time.Now().UTC()
			return nil
		}
	}
	return entity.ErrNotFound
}

type memAuditRepo struct {
	mu   sync.Mutex
	rows []*entity.AuditEvent
	seq  int64
}

func newMemAuditRepo() *memAuditRepo {
	return &memAuditRepo{}
}

func (r *memAuditRepo) Create(_ context.Context, e *entity.AuditEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	cp := *e
	cp.ID = r.seq
	r.rows = append(r.rows, &cp)
	e.ID = cp.ID
	return nil
}

func newTestService() service.TenancyService {
	return service.NewTenancyService(&service.Components{
		PrincipalRepo:    newMemPrincipalRepo(),
		TenantRepo:       newMemTenantRepo(),
		MembershipRepo:   newMemMembershipRepo(),
		SpaceRefRepo:     newMemSpaceRefRepo(),
		AuditRepo:        newMemAuditRepo(),
		PlatformRoleRepo: newMemPlatformRoleRepo(),
	})
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestBootstrap_Idempotent(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	first, err := svc.Bootstrap(ctx, 42, "Alice", 1001)
	require.NoError(t, err)
	require.NotNil(t, first)
	assert.True(t, first.Created)
	assert.Equal(t, entity.RoleOwner, first.Membership.Role)
	assert.Equal(t, entity.SpacePurposeDefault, first.SpaceRef.Purpose)
	assert.Equal(t, int64(1001), first.SpaceRef.CozeSpaceID)
	assert.Equal(t, fmt.Sprintf("personal_%d", 42), first.Tenant.TenantKey)

	second, err := svc.Bootstrap(ctx, 42, "Alice", 1001)
	require.NoError(t, err)
	require.NotNil(t, second)
	assert.False(t, second.Created)
	assert.Equal(t, first.Principal.PrincipalID, second.Principal.PrincipalID)
	assert.Equal(t, first.Tenant.TenantID, second.Tenant.TenantID)
	assert.Equal(t, first.Membership.PrincipalID, second.Membership.PrincipalID)
}

func TestUpdateTenant_RevisionConflict(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	principal, err := svc.ResolveOrCreatePrincipal(ctx, 7, "Bob")
	require.NoError(t, err)

	tenant, err := svc.CreateTenant(ctx, &service.CreateTenantRequest{
		Name:             "Acme",
		OwnerPrincipalID: principal.PrincipalID,
	})
	require.NoError(t, err)
	require.Equal(t, int32(1), tenant.Revision)

	tenant.DisplayName = "Acme Corp"
	updated, err := svc.UpdateTenant(ctx, tenant, 1, principal.PrincipalID, "req-1")
	require.NoError(t, err)
	assert.Equal(t, int32(2), updated.Revision)

	tenant.DisplayName = "Stale"
	_, err = svc.UpdateTenant(ctx, tenant, 1, principal.PrincipalID, "req-1")
	require.ErrorIs(t, err, entity.ErrRevisionConflict)
}

func TestUpdateMemberRole_RevisionConflict(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	owner, err := svc.ResolveOrCreatePrincipal(ctx, 1, "Owner")
	require.NoError(t, err)
	member, err := svc.ResolveOrCreatePrincipal(ctx, 2, "Member")
	require.NoError(t, err)

	tenant, err := svc.CreateTenant(ctx, &service.CreateTenantRequest{
		Name:             "Team",
		OwnerPrincipalID: owner.PrincipalID,
	})
	require.NoError(t, err)

	m, err := svc.AddMember(ctx, &service.AddMemberRequest{
		TenantID:    tenant.TenantID,
		PrincipalID: member.PrincipalID,
		Role:        entity.RoleMember,
		CreatedBy:   owner.PrincipalID,
	})
	require.NoError(t, err)
	require.Equal(t, int32(1), m.Revision)

	updated, err := svc.UpdateMemberRole(ctx, tenant.TenantID, member.PrincipalID, entity.RoleAdmin, 1, owner.PrincipalID, "req-1")
	require.NoError(t, err)
	assert.Equal(t, entity.RoleAdmin, updated.Role)
	assert.Equal(t, int32(2), updated.Revision)

	_, err = svc.UpdateMemberRole(ctx, tenant.TenantID, member.PrincipalID, entity.RoleViewer, 1, owner.PrincipalID, "req-1")
	require.ErrorIs(t, err, entity.ErrRevisionConflict)
}

func TestPlatformRole_SetAndPasswordChangeRequired(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	principal, err := svc.ResolveOrCreatePrincipal(ctx, 9001, "admin")
	require.NoError(t, err)

	require.NoError(t, svc.SetPlatformRole(ctx, principal.PrincipalID, entity.PlatformRoleSuperAdmin))
	require.NoError(t, svc.SetPasswordChangeRequired(ctx, principal.PrincipalID, true))

	role, err := svc.GetPlatformRole(ctx, principal.PrincipalID)
	require.NoError(t, err)
	require.NotNil(t, role)
	assert.Equal(t, entity.PlatformRoleSuperAdmin, role.Role)
	assert.True(t, role.PasswordChangeRequired)

	require.NoError(t, svc.SetPasswordChangeRequired(ctx, principal.PrincipalID, false))
	role, err = svc.GetPlatformRole(ctx, principal.PrincipalID)
	require.NoError(t, err)
	assert.False(t, role.PasswordChangeRequired)
}

func TestPlatformRole_LastSuperAdminProtection(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	admin, err := svc.ResolveOrCreatePrincipal(ctx, 1, "admin")
	require.NoError(t, err)
	require.NoError(t, svc.SetPlatformRole(ctx, admin.PrincipalID, entity.PlatformRoleSuperAdmin))

	count, err := svc.CountActiveSuperAdmins(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	require.NoError(t, svc.SuspendPrincipal(ctx, admin.PrincipalID))
	count, err = svc.CountActiveSuperAdmins(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	require.NoError(t, svc.ActivatePrincipal(ctx, admin.PrincipalID))
	count, err = svc.CountActiveSuperAdmins(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestPrincipal_ListAllAndStatus(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	p1, err := svc.ResolveOrCreatePrincipal(ctx, 11, "one")
	require.NoError(t, err)
	p2, err := svc.ResolveOrCreatePrincipal(ctx, 12, "two")
	require.NoError(t, err)

	all, err := svc.ListAllPrincipals(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(all), 2)

	require.NoError(t, svc.SuspendPrincipal(ctx, p2.PrincipalID))
	got, err := svc.GetPrincipalByID(ctx, p2.PrincipalID)
	require.NoError(t, err)
	assert.Equal(t, entity.PrincipalStatusSuspended, got.Status)

	require.NoError(t, svc.ActivatePrincipal(ctx, p2.PrincipalID))
	got, err = svc.GetPrincipalByID(ctx, p1.PrincipalID)
	require.NoError(t, err)
	assert.Equal(t, entity.PrincipalStatusActive, got.Status)
}
