/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"
	"fmt"
	"strings"
	"time"

	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	tenancyentity "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	tenancysvc "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/service"
	userservice "github.com/coze-dev/coze-studio/backend/domain/user/service"
)

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type AdminUserDTO struct {
	PrincipalID            string                    `json:"principal_id"`
	Account                string                    `json:"account"`
	DisplayName            string                    `json:"display_name"`
	Status                 string                    `json:"status"`
	PlatformRole           string                    `json:"platform_role"`
	Workspaces             []*AdminUserWorkspaceDTO  `json:"workspaces"`
	PasswordChangeRequired bool                      `json:"password_change_required"`
	CreatedAt              string                    `json:"created_at"`
}

type AdminUserWorkspaceDTO struct {
	TenantID   string `json:"tenant_id"`
	TenantName string `json:"tenant_name"`
	Role       string `json:"role"`
}

type AdminCreateUserInput struct {
	Account     string `json:"account"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
	TenantID    string `json:"tenant_id,omitempty"`
	TenantRole  string `json:"tenant_role,omitempty"`
}

type AdminCreateUserResponse struct {
	User            *AdminUserDTO `json:"user"`
	InitialPassword string        `json:"initial_password"`
}

type AdminResetPasswordInput struct {
	NewPassword string `json:"new_password"`
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (s *ApplicationService) requireSuperAdmin(ctx context.Context) (*tenancyentity.Principal, error) {
	principal, err := s.currentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	role, err := s.TenancySVC.GetPlatformRole(ctx, principal.PrincipalID)
	if err != nil {
		return nil, formaerrors.Internal(err.Error())
	}
	if role == nil || role.Role != tenancyentity.PlatformRoleSuperAdmin {
		return nil, formaerrors.AdminForbidden("")
	}
	return principal, nil
}

func accountToEmail(account string) string {
	if strings.Contains(account, "@") {
		return account
	}
	return account + "@forma.local"
}

func emailToAccount(email string) string {
	if strings.HasSuffix(email, "@forma.local") {
		return strings.TrimSuffix(email, "@forma.local")
	}
	return email
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return formaerrors.AdminBadRequest("password must be at least 8 characters")
	}
	if password == DefaultBootstrapAdminPassword {
		return formaerrors.AdminBadRequest("password 'admin123' is not allowed as a new password")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Admin Application Methods
// ---------------------------------------------------------------------------

func (s *ApplicationService) AdminListUsers(ctx context.Context) ([]*AdminUserDTO, error) {
	if _, err := s.requireSuperAdmin(ctx); err != nil {
		return nil, err
	}

	principals, err := s.TenancySVC.ListAllPrincipals(ctx)
	if err != nil {
		return nil, formaerrors.Internal(err.Error())
	}

	roles, err := s.TenancySVC.ListPlatformRoles(ctx)
	if err != nil {
		return nil, formaerrors.Internal(err.Error())
	}
	roleMap := make(map[string]*tenancyentity.FormaPlatformRoleAssignment, len(roles))
	for _, r := range roles {
		roleMap[r.PrincipalID] = r
	}

	out := make([]*AdminUserDTO, 0, len(principals))
	for _, p := range principals {
		dto := s.buildAdminUserDTO(ctx, p, roleMap[p.PrincipalID])
		out = append(out, dto)
	}
	return out, nil
}

func (s *ApplicationService) AdminGetUser(ctx context.Context, principalID string) (*AdminUserDTO, error) {
	if _, err := s.requireSuperAdmin(ctx); err != nil {
		return nil, err
	}

	p, err := s.TenancySVC.GetPrincipalByID(ctx, principalID)
	if err != nil {
		return nil, formaerrors.Internal(err.Error())
	}
	if p == nil {
		return nil, formaerrors.AdminUserNotFound("")
	}

	role, _ := s.TenancySVC.GetPlatformRole(ctx, principalID)
	return s.buildAdminUserDTO(ctx, p, role), nil
}

func (s *ApplicationService) AdminCreateUser(ctx context.Context, in *AdminCreateUserInput) (*AdminCreateUserResponse, error) {
	admin, err := s.requireSuperAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil || in.Account == "" {
		return nil, formaerrors.AdminBadRequest("account is required")
	}
	if s.UserDomainSVC == nil {
		return nil, formaerrors.Internal("user service not available")
	}

	password := in.Password
	if password == "" {
		password = "Forma" + time.Now().Format("0102") + "!"
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	email := accountToEmail(in.Account)
	displayName := in.DisplayName
	if displayName == "" {
		displayName = in.Account
	}

	cozeUser, err := s.UserDomainSVC.Create(ctx, &userservice.CreateUserRequest{
		Email:    email,
		Password: password,
		Name:     displayName,
	})
	if err != nil {
		if isEmailExistsError(err) {
			return nil, formaerrors.AdminBadRequest("account already exists")
		}
		return nil, formaerrors.Internal(fmt.Sprintf("create user failed: %v", err))
	}

	principal, err := s.TenancySVC.ResolveOrCreatePrincipal(ctx, cozeUser.UserID, displayName)
	if err != nil {
		return nil, formaerrors.Internal(err.Error())
	}

	if err := s.TenancySVC.SetPlatformRole(ctx, principal.PrincipalID, tenancyentity.PlatformRoleUser); err != nil {
		return nil, formaerrors.Internal(err.Error())
	}
	_ = s.TenancySVC.SetPasswordChangeRequired(ctx, principal.PrincipalID, true)

	if in.TenantID != "" {
		tenantRole := tenancyentity.MembershipRole(in.TenantRole)
		if tenantRole == "" {
			tenantRole = tenancyentity.RoleMember
		}
		_, _ = s.TenancySVC.AddMember(ctx, &tenancysvc.AddMemberRequest{
			TenantID:    in.TenantID,
			PrincipalID: principal.PrincipalID,
			Role:        tenantRole,
			CreatedBy:   admin.PrincipalID,
		})
	}

	_ = s.TenancySVC.RecordAudit(ctx, &tenancyentity.AuditEvent{
		PrincipalID: admin.PrincipalID,
		Action:      tenancyentity.AuditAdminUserCreated,
		Resource:    principal.PrincipalID,
		CreatedAt:   time.Now().UTC(),
	})

	role, _ := s.TenancySVC.GetPlatformRole(ctx, principal.PrincipalID)
	dto := s.buildAdminUserDTO(ctx, &tenancyentity.Principal{
		PrincipalID: principal.PrincipalID,
		DisplayName: displayName,
		Status:      tenancyentity.PrincipalStatusActive,
		CreatedAt:   principal.CreatedAt,
	}, role)

	return &AdminCreateUserResponse{
		User:            dto,
		InitialPassword: password,
	}, nil
}

func (s *ApplicationService) AdminDisableUser(ctx context.Context, principalID string) error {
	admin, err := s.requireSuperAdmin(ctx)
	if err != nil {
		return err
	}

	p, err := s.TenancySVC.GetPrincipalByID(ctx, principalID)
	if err != nil {
		return formaerrors.Internal(err.Error())
	}
	if p == nil {
		return formaerrors.AdminUserNotFound("")
	}

	// Prevent disabling the last active super admin
	role, _ := s.TenancySVC.GetPlatformRole(ctx, principalID)
	if role != nil && role.Role == tenancyentity.PlatformRoleSuperAdmin {
		count, cErr := s.TenancySVC.CountActiveSuperAdmins(ctx)
		if cErr != nil {
			return formaerrors.Internal(cErr.Error())
		}
		if count <= 1 {
			return formaerrors.AdminLastSuperAdmin("")
		}
	}

	if err := s.TenancySVC.SuspendPrincipal(ctx, principalID); err != nil {
		return formaerrors.Internal(err.Error())
	}

	// Try to clear session by logging out the Coze user
	if s.UserDomainSVC != nil && p.CozeUserID > 0 {
		_ = s.UserDomainSVC.Logout(ctx, p.CozeUserID)
	}

	_ = s.TenancySVC.RecordAudit(ctx, &tenancyentity.AuditEvent{
		PrincipalID: admin.PrincipalID,
		Action:      tenancyentity.AuditAdminUserDisabled,
		Resource:    principalID,
		CreatedAt:   time.Now().UTC(),
	})

	return nil
}

func (s *ApplicationService) AdminEnableUser(ctx context.Context, principalID string) error {
	admin, err := s.requireSuperAdmin(ctx)
	if err != nil {
		return err
	}

	p, err := s.TenancySVC.GetPrincipalByID(ctx, principalID)
	if err != nil {
		return formaerrors.Internal(err.Error())
	}
	if p == nil {
		return formaerrors.AdminUserNotFound("")
	}

	if err := s.TenancySVC.ActivatePrincipal(ctx, principalID); err != nil {
		return formaerrors.Internal(err.Error())
	}

	_ = s.TenancySVC.RecordAudit(ctx, &tenancyentity.AuditEvent{
		PrincipalID: admin.PrincipalID,
		Action:      tenancyentity.AuditAdminUserEnabled,
		Resource:    principalID,
		CreatedAt:   time.Now().UTC(),
	})

	return nil
}

func (s *ApplicationService) AdminResetPassword(ctx context.Context, principalID string, in *AdminResetPasswordInput) error {
	admin, err := s.requireSuperAdmin(ctx)
	if err != nil {
		return err
	}
	if in == nil || in.NewPassword == "" {
		return formaerrors.AdminBadRequest("new_password is required")
	}
	if err := validatePassword(in.NewPassword); err != nil {
		return err
	}
	if s.UserDomainSVC == nil {
		return formaerrors.Internal("user service not available")
	}

	p, err := s.TenancySVC.GetPrincipalByID(ctx, principalID)
	if err != nil {
		return formaerrors.Internal(err.Error())
	}
	if p == nil {
		return formaerrors.AdminUserNotFound("")
	}

	// Find email via Coze user
	cozeUser, err := s.UserDomainSVC.GetUserInfo(ctx, p.CozeUserID)
	if err != nil {
		return formaerrors.Internal(fmt.Sprintf("get user info: %v", err))
	}

	if err := s.UserDomainSVC.ResetPassword(ctx, cozeUser.Email, in.NewPassword); err != nil {
		return formaerrors.Internal(fmt.Sprintf("reset password: %v", err))
	}

	_ = s.TenancySVC.SetPasswordChangeRequired(ctx, principalID, true)

	_ = s.TenancySVC.RecordAudit(ctx, &tenancyentity.AuditEvent{
		PrincipalID: admin.PrincipalID,
		Action:      tenancyentity.AuditAdminPasswordReset,
		Resource:    principalID,
		CreatedAt:   time.Now().UTC(),
	})

	return nil
}

func (s *ApplicationService) AdminChangePassword(ctx context.Context, currentPassword, newPassword string) error {
	principal, err := s.currentPrincipal(ctx)
	if err != nil {
		return err
	}
	if s.UserDomainSVC == nil {
		return formaerrors.Internal("user service not available")
	}
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	// Verify current password by attempting login
	cozeUser, err := s.UserDomainSVC.GetUserInfo(ctx, principal.CozeUserID)
	if err != nil {
		return formaerrors.Internal("failed to get user info")
	}
	_, err = s.UserDomainSVC.Login(ctx, cozeUser.Email, currentPassword)
	if err != nil {
		return formaerrors.Unauthenticated("current password is incorrect")
	}

	if err := s.UserDomainSVC.ResetPassword(ctx, cozeUser.Email, newPassword); err != nil {
		return formaerrors.Internal(fmt.Sprintf("reset password: %v", err))
	}

	// Clear password_change_required flag
	_ = s.TenancySVC.SetPasswordChangeRequired(ctx, principal.PrincipalID, false)

	_ = s.TenancySVC.RecordAudit(ctx, &tenancyentity.AuditEvent{
		PrincipalID: principal.PrincipalID,
		Action:      tenancyentity.AuditAdminPasswordChanged,
		Resource:    principal.PrincipalID,
		CreatedAt:   time.Now().UTC(),
	})

	return nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (s *ApplicationService) buildAdminUserDTO(ctx context.Context, p *tenancyentity.Principal, role *tenancyentity.FormaPlatformRoleAssignment) *AdminUserDTO {
	dto := &AdminUserDTO{
		PrincipalID: p.PrincipalID,
		Account:     p.DisplayName,
		DisplayName: p.DisplayName,
		Status:      string(p.Status),
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		Workspaces:  make([]*AdminUserWorkspaceDTO, 0),
	}

	if role != nil {
		dto.PlatformRole = string(role.Role)
		dto.PasswordChangeRequired = role.PasswordChangeRequired
	} else {
		dto.PlatformRole = string(tenancyentity.PlatformRoleUser)
	}

	// Resolve account from CozeUserID
	if s.UserDomainSVC != nil && p.CozeUserID > 0 {
		if cozeUser, err := s.UserDomainSVC.GetUserInfo(ctx, p.CozeUserID); err == nil && cozeUser != nil {
			dto.Account = emailToAccount(cozeUser.Email)
			if dto.DisplayName == "" {
				dto.DisplayName = cozeUser.Name
			}
		}
	}

	// Resolve workspaces
	memberships, _ := s.TenancySVC.ListMembershipsForPrincipal(ctx, p.PrincipalID)
	for _, m := range memberships {
		if m == nil || m.Status != tenancyentity.MembershipActive {
			continue
		}
		t, _ := s.TenancySVC.GetTenant(ctx, m.TenantID)
		name := m.TenantID
		if t != nil {
			name = t.DisplayName
		}
		dto.Workspaces = append(dto.Workspaces, &AdminUserWorkspaceDTO{
			TenantID:   m.TenantID,
			TenantName: name,
			Role:       string(m.Role),
		})
	}

	return dto
}
