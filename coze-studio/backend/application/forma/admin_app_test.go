/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma_test

import (
	"context"
	"fmt"
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
	userservice "github.com/coze-dev/coze-studio/backend/domain/user/service"
	"github.com/coze-dev/coze-studio/backend/pkg/ctxcache"
	"github.com/coze-dev/coze-studio/backend/types/consts"
)

type memUserDomain struct {
	mu    sync.Mutex
	byID  map[int64]*userentity.User
	email map[string]int64
	seq   int64
}

func newMemUserDomain() *memUserDomain {
	return &memUserDomain{
		byID:  make(map[int64]*userentity.User),
		email: make(map[string]int64),
	}
}

func (m *memUserDomain) Create(_ context.Context, req *userservice.CreateUserRequest) (*userentity.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.email[req.Email]; ok {
		return nil, fmt.Errorf("email already exist code=700000001")
	}
	m.seq++
	u := &userentity.User{
		UserID:     m.seq,
		Name:       req.Name,
		Email:      req.Email,
		SessionKey: fmt.Sprintf("sess-%d", m.seq),
		CreatedAt:  time.Now().UnixMilli(),
		UpdatedAt:  time.Now().UnixMilli(),
	}
	m.byID[u.UserID] = u
	m.email[req.Email] = u.UserID
	cp := *u
	return &cp, nil
}

func (m *memUserDomain) Login(_ context.Context, email, password string) (*userentity.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.email[email]
	if !ok {
		return nil, fmt.Errorf("invalid credentials")
	}
	u := m.byID[id]
	u.SessionKey = fmt.Sprintf("sess-%d-%d", id, time.Now().UnixNano())
	_ = password
	cp := *u
	return &cp, nil
}

func (m *memUserDomain) Logout(_ context.Context, userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if u, ok := m.byID[userID]; ok {
		u.SessionKey = ""
	}
	return nil
}

func (m *memUserDomain) ResetPassword(_ context.Context, email, password string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.email[email]
	if !ok {
		return fmt.Errorf("user not found")
	}
	u := m.byID[id]
	u.SessionKey = ""
	_ = password
	return nil
}

func (m *memUserDomain) GetUserInfo(_ context.Context, userID int64) (*userentity.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byID[userID]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	cp := *u
	return &cp, nil
}

func (m *memUserDomain) UpdateAvatar(context.Context, int64, string, []byte) (string, error) {
	return "", nil
}
func (m *memUserDomain) UpdateProfile(context.Context, *userservice.UpdateProfileRequest) error {
	return nil
}
func (m *memUserDomain) ValidateProfileUpdate(context.Context, *userservice.ValidateProfileUpdateRequest) (*userservice.ValidateProfileUpdateResponse, error) {
	return &userservice.ValidateProfileUpdateResponse{Code: userservice.ValidateSuccess}, nil
}
func (m *memUserDomain) GetUserProfiles(ctx context.Context, userID int64) (*userentity.User, error) {
	return m.GetUserInfo(ctx, userID)
}
func (m *memUserDomain) MGetUserProfiles(context.Context, []int64) ([]*userentity.User, error) {
	return nil, nil
}
func (m *memUserDomain) ValidateSession(context.Context, string) (*userentity.Session, bool, error) {
	return nil, false, nil
}
func (m *memUserDomain) GetUserSpaceList(context.Context, int64) ([]*userentity.Space, error) {
	return nil, nil
}
func (m *memUserDomain) GetUserSpaceBySpaceID(context.Context, []int64) ([]*userentity.Space, error) {
	return nil, nil
}
func (m *memUserDomain) GetSaasUserInfo(context.Context) (*userentity.SaasUserData, error) {
	return nil, nil
}
func (m *memUserDomain) GetUserBenefit(context.Context) (*userentity.UserBenefit, error) {
	return nil, nil
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
	if _, ok := r.byID[assignment.PrincipalID]; !ok {
		return entity.ErrNotFound
	}
	cp := *assignment
	cp.UpdatedAt = time.Now().UTC()
	r.byID[cp.PrincipalID] = &cp
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
	set := map[string]struct{}{}
	for _, id := range activePrincipalIDs {
		set[id] = struct{}{}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, a := range r.byID {
		if a.Role == entity.PlatformRoleSuperAdmin {
			if _, ok := set[a.PrincipalID]; ok {
				n++
			}
		}
	}
	return n, nil
}

func withAdminSession(userID int64, email string) context.Context {
	ctx := ctxcache.Init(context.Background())
	ctxcache.Store(ctx, consts.SessionDataKeyInCtx, &userentity.Session{
		UserID:    userID,
		UserEmail: email,
	})
	return ctx
}

func newAdminAppService(users *memUserDomain) *formaapp.ApplicationService {
	svc := &formaapp.ApplicationService{
		TenancySVC: tenancysvc.NewTenancyService(&tenancysvc.Components{
			PrincipalRepo:    newMemPrincipalRepo(),
			TenantRepo:       newMemTenantRepo(),
			MembershipRepo:   newMemMembershipRepo(),
			SpaceRefRepo:     newMemSpaceRefRepo(),
			AuditRepo:        &memAuditRepo{},
			PlatformRoleRepo: newMemPlatformRoleRepo(),
		}),
	}
	svc.SetUserDomainSVC(users)
	return svc
}

func TestBootstrapDefaultAdmin_Idempotent(t *testing.T) {
	users := newMemUserDomain()
	svc := newAdminAppService(users)

	require.NoError(t, svc.BootstrapDefaultAdmin(context.Background()))
	require.NoError(t, svc.BootstrapDefaultAdmin(context.Background()))

	assert.Equal(t, 1, len(users.email))
	assert.Contains(t, users.email, "admin@forma.local")

	adminUser, err := users.GetUserInfo(context.Background(), users.email["admin@forma.local"])
	require.NoError(t, err)

	ctx := withAdminSession(adminUser.UserID, adminUser.Email)
	principal, err := svc.TenancySVC.ResolveOrCreatePrincipal(ctx, adminUser.UserID, "admin")
	require.NoError(t, err)
	role, err := svc.TenancySVC.GetPlatformRole(ctx, principal.PrincipalID)
	require.NoError(t, err)
	require.NotNil(t, role)
	assert.Equal(t, entity.PlatformRoleSuperAdmin, role.Role)
	assert.True(t, role.PasswordChangeRequired)
}

func TestAdminCreateUser_AndDisableEnable(t *testing.T) {
	users := newMemUserDomain()
	svc := newAdminAppService(users)
	require.NoError(t, svc.BootstrapDefaultAdmin(context.Background()))

	adminID := users.email["admin@forma.local"]
	adminCtx := withAdminSession(adminID, "admin@forma.local")

	created, err := svc.AdminCreateUser(adminCtx, &formaapp.AdminCreateUserInput{
		Account:     "user01",
		DisplayName: "User One",
		Password:    "User01Init!",
	})
	require.NoError(t, err)
	require.NotNil(t, created.User)
	assert.Equal(t, "user01", created.User.Account)
	assert.Equal(t, "User01Init!", created.InitialPassword)
	assert.True(t, created.User.PasswordChangeRequired)

	// Collision: user01@forma.local must map to same identity and fail as exists.
	_, err = svc.AdminCreateUser(adminCtx, &formaapp.AdminCreateUserInput{
		Account:  "user01@forma.local",
		Password: "User01Init!",
	})
	require.Error(t, err)

	principalID := created.User.PrincipalID
	require.NoError(t, svc.AdminDisableUser(adminCtx, principalID))
	got, err := svc.AdminGetUser(adminCtx, principalID)
	require.NoError(t, err)
	assert.Equal(t, string(entity.PrincipalStatusSuspended), got.Status)

	require.NoError(t, svc.AdminEnableUser(adminCtx, principalID))
	got, err = svc.AdminGetUser(adminCtx, principalID)
	require.NoError(t, err)
	assert.Equal(t, string(entity.PrincipalStatusActive), got.Status)

	require.NoError(t, svc.AdminResetPassword(adminCtx, principalID, &formaapp.AdminResetPasswordInput{
		NewPassword: "User01Reset!",
	}))
	role, err := svc.TenancySVC.GetPlatformRole(adminCtx, principalID)
	require.NoError(t, err)
	assert.True(t, role.PasswordChangeRequired)
}

func TestAdminDisable_LastSuperAdminDenied(t *testing.T) {
	users := newMemUserDomain()
	svc := newAdminAppService(users)
	require.NoError(t, svc.BootstrapDefaultAdmin(context.Background()))

	adminID := users.email["admin@forma.local"]
	adminCtx := withAdminSession(adminID, "admin@forma.local")
	principal, err := svc.TenancySVC.ResolveOrCreatePrincipal(adminCtx, adminID, "admin")
	require.NoError(t, err)

	err = svc.AdminDisableUser(adminCtx, principal.PrincipalID)
	require.Error(t, err)
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
	assert.Equal(t, formaerrors.KeyAdminLastSuperAdmin, fe.Key)
}

func TestAdminChangePassword_ClearsRequiredFlag(t *testing.T) {
	users := newMemUserDomain()
	svc := newAdminAppService(users)
	require.NoError(t, svc.BootstrapDefaultAdmin(context.Background()))

	adminID := users.email["admin@forma.local"]
	adminCtx := withAdminSession(adminID, "admin@forma.local")
	principal, err := svc.TenancySVC.ResolveOrCreatePrincipal(adminCtx, adminID, "admin")
	require.NoError(t, err)

	require.NoError(t, svc.AdminChangePassword(adminCtx, "admin123", "Admin123456!"))
	role, err := svc.TenancySVC.GetPlatformRole(adminCtx, principal.PrincipalID)
	require.NoError(t, err)
	assert.False(t, role.PasswordChangeRequired)

	err = svc.AdminChangePassword(adminCtx, "Admin123456!", "admin123")
	require.Error(t, err)
}

func TestNonAdminCannotCreateUser(t *testing.T) {
	users := newMemUserDomain()
	svc := newAdminAppService(users)
	require.NoError(t, svc.BootstrapDefaultAdmin(context.Background()))

	member, err := users.Create(context.Background(), &userservice.CreateUserRequest{
		Email:    "member@example.com",
		Password: "Member123!",
		Name:     "member",
	})
	require.NoError(t, err)
	_, err = svc.TenancySVC.ResolveOrCreatePrincipal(context.Background(), member.UserID, "member")
	require.NoError(t, err)

	memberCtx := withAdminSession(member.UserID, member.Email)
	_, err = svc.AdminCreateUser(memberCtx, &formaapp.AdminCreateUserInput{
		Account:  "x",
		Password: "Member123!",
	})
	require.Error(t, err)
	fe, ok := formaerrors.AsFormaError(err)
	require.True(t, ok)
	assert.Equal(t, formaerrors.KeyAdminForbidden, fe.Key)
}