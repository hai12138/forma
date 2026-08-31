/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	tenancysvc "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/service"
	userentity "github.com/coze-dev/coze-studio/backend/domain/user/entity"
	"github.com/coze-dev/coze-studio/backend/pkg/ctxcache"
	"github.com/coze-dev/coze-studio/backend/types/consts"
)

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

func (r *memPrincipalRepo) GetByProviderSubject(_ context.Context, provider, subject string) (*entity.Principal, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.byID {
		if p.Provider == provider && p.ExternalSubject == subject {
			cp := *p
			return &cp, nil
		}
	}
	return nil, nil
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
	if !ok {
		return nil, nil
	}
	cp := *t
	return &cp, nil
}

func (r *memTenantRepo) Update(_ context.Context, t *entity.Tenant, expectedRevision int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cur, ok := r.byID[t.TenantID]
	if !ok {
		return entity.ErrNotFound
	}
	if cur.Revision != expectedRevision {
		return entity.ErrRevisionConflict
	}
	cp := *t
	cp.Revision = expectedRevision + 1
	r.byID[t.TenantID] = &cp
	t.Revision = cp.Revision
	return nil
}

func (r *memTenantRepo) ListByPrincipalID(_ context.Context, _ string) ([]*entity.Tenant, error) {
	return nil, nil
}

type memMembershipRepo struct {
	mu   sync.Mutex
	rows []*entity.Membership
	seq  int64
}

func newMemMembershipRepo() *memMembershipRepo { return &memMembershipRepo{} }

func (r *memMembershipRepo) Create(_ context.Context, m *entity.Membership) error {
	r.mu.Lock()
	defer r.mu.Unlock()
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

func newMemSpaceRefRepo() *memSpaceRefRepo { return &memSpaceRefRepo{} }

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

type memAuditRepo struct{}

func (r *memAuditRepo) Create(_ context.Context, _ *entity.AuditEvent) error { return nil }

func withSession(userID int64, email string) context.Context {
	ctx := ctxcache.Init(context.Background())
	ctxcache.Store(ctx, consts.SessionDataKeyInCtx, &userentity.Session{
		UserID:    userID,
		UserEmail: email,
	})
	return ctx
}

func newAppService() *formaapp.ApplicationService {
	return &formaapp.ApplicationService{
		TenancySVC: tenancysvc.NewTenancyService(&tenancysvc.Components{
			PrincipalRepo:  newMemPrincipalRepo(),
			TenantRepo:     newMemTenantRepo(),
			MembershipRepo: newMemMembershipRepo(),
			SpaceRefRepo:   newMemSpaceRefRepo(),
			AuditRepo:      &memAuditRepo{},
		}),
	}
}

func TestResolveTenantContext_ForgedHeaderForbidden(t *testing.T) {
	svc := newAppService()
	ctx := withSession(100, "alice@example.com")

	boot, err := svc.TenancySVC.Bootstrap(ctx, 100, "Alice", 9001)
	require.NoError(t, err)
	require.NotNil(t, boot)

	// Legitimate header works.
	tc, err := svc.ResolveTenantContext(ctx, boot.Tenant.TenantID)
	require.NoError(t, err)
	assert.Equal(t, boot.Tenant.TenantID, tc.TenantID)
	assert.Contains(t, tc.AllowedSpaceIDs, int64(9001))

	// Forged tenant header for another tenant → forbidden.
	_, err = svc.ResolveTenantContext(ctx, "ten_forged_other")
	require.Error(t, err)
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
	assert.Equal(t, formaerrors.KeyTenantForbidden, fe.Key)
	assert.Equal(t, formaerrors.CodeTenantForbidden, fe.Code)
}

func TestResolveTenantContext_Unauthenticated(t *testing.T) {
	svc := newAppService()
	_, err := svc.ResolveTenantContext(context.Background(), "")
	require.Error(t, err)
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
	assert.Equal(t, formaerrors.KeyUnauthenticated, fe.Key)
}

func TestResolveTenantContext_EmptyHeaderPicksFirst(t *testing.T) {
	svc := newAppService()
	ctx := withSession(200, "bob@example.com")
	boot, err := svc.TenancySVC.Bootstrap(ctx, 200, "Bob", 42)
	require.NoError(t, err)

	tc, err := svc.ResolveTenantContext(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, boot.Tenant.TenantID, tc.TenantID)
}
