/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
	tenancysvc "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/service"
	userservice "github.com/coze-dev/coze-studio/backend/domain/user/service"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
)

const (
	DefaultBootstrapAdminUsername = "admin"
	DefaultBootstrapAdminEmail    = "admin@" + FormaLocalEmailDomain
	DefaultBootstrapAdminPassword = "admin123"
	DefaultBootstrapTenantName    = "Forma Workspace"
)

func bootstrapAdminUsername() string {
	if v := os.Getenv("FORMA_BOOTSTRAP_ADMIN_USERNAME"); v != "" {
		return v
	}
	return DefaultBootstrapAdminUsername
}

func bootstrapAdminEmail() string {
	if v := os.Getenv("FORMA_BOOTSTRAP_ADMIN_EMAIL"); v != "" {
		return v
	}
	return DefaultBootstrapAdminEmail
}

func bootstrapAdminPassword() string {
	if v := os.Getenv("FORMA_BOOTSTRAP_ADMIN_PASSWORD"); v != "" {
		return v
	}
	return DefaultBootstrapAdminPassword
}

// BootstrapDefaultAdmin creates the default platform super-admin if it does not yet exist.
// It is idempotent: repeated calls are safe.
func (s *ApplicationService) BootstrapDefaultAdmin(ctx context.Context) error {
	if s.UserDomainSVC == nil {
		return fmt.Errorf("UserDomainSVC not set; call SetUserDomainSVC first")
	}

	email := bootstrapAdminEmail()
	password := bootstrapAdminPassword()
	displayName := bootstrapAdminUsername()

	// Try to create the Coze user. If it already exists, try to log in.
	cozeUser, err := s.UserDomainSVC.Create(ctx, &userservice.CreateUserRequest{
		Email:    email,
		Password: password,
		Name:     displayName,
	})

	if err != nil {
		if isEmailExistsError(err) {
			// User exists — try login with default password
			cozeUser, err = s.UserDomainSVC.Login(ctx, email, password)
			if err != nil {
				// Password was changed or other issue — admin already bootstrapped
				logs.Infof("bootstrap admin already exists (login with default password failed, likely already changed)")
				return nil
			}
		} else {
			return fmt.Errorf("bootstrap admin create failed: %w", err)
		}
	}

	// Ensure Forma principal exists
	principal, err := s.TenancySVC.ResolveOrCreatePrincipal(ctx, cozeUser.UserID, displayName)
	if err != nil {
		return fmt.Errorf("bootstrap admin principal: %w", err)
	}

	// Set SUPER_ADMIN role (idempotent)
	existing, _ := s.TenancySVC.GetPlatformRole(ctx, principal.PrincipalID)
	if existing == nil {
		if err := s.TenancySVC.SetPlatformRole(ctx, principal.PrincipalID, entity.PlatformRoleSuperAdmin); err != nil {
			return fmt.Errorf("bootstrap admin set role: %w", err)
		}
		if err := s.TenancySVC.SetPasswordChangeRequired(ctx, principal.PrincipalID, true); err != nil {
			return fmt.Errorf("bootstrap admin set password_change_required: %w", err)
		}
	}

	// Ensure default workspace tenant
	tenants, _ := s.TenancySVC.ListTenantsForPrincipal(ctx, principal.PrincipalID)
	if len(tenants) == 0 {
		tenant, tErr := s.TenancySVC.CreateTenant(ctx, &tenancysvc.CreateTenantRequest{
			Name:             DefaultBootstrapTenantName,
			DisplayName:      DefaultBootstrapTenantName,
			OwnerPrincipalID: principal.PrincipalID,
		})
		if tErr != nil {
			return fmt.Errorf("bootstrap admin tenant: %w", tErr)
		}
		_, _ = s.TenancySVC.AddMember(ctx, &tenancysvc.AddMemberRequest{
			TenantID:    tenant.TenantID,
			PrincipalID: principal.PrincipalID,
			Role:        entity.RoleOwner,
			CreatedBy:   principal.PrincipalID,
		})
	}

	// Record audit
	_ = s.TenancySVC.RecordAudit(ctx, &entity.AuditEvent{
		PrincipalID: principal.PrincipalID,
		Action:      entity.AuditAdminBootstrap,
		Resource:    principal.PrincipalID,
		CreatedAt:   time.Now().UTC(),
	})

	if password == DefaultBootstrapAdminPassword {
		logs.Warnf("[SECURITY] Forma bootstrap admin using default password — change immediately in production!")
	}

	logs.Infof("bootstrap admin ready: principal=%s", principal.PrincipalID)
	return nil
}

// isEmailExistsError checks if the error indicates the email already exists.
func isEmailExistsError(err error) bool {
	if err == nil {
		return false
	}
	// The Coze user domain returns errorx with code 700000001 for duplicate email.
	// Check the error message as a fallback since we can't import errno directly
	// without coupling, but we also check the code.
	msg := err.Error()
	return strings.Contains(msg, "700000001") || strings.Contains(msg, "email already exist")
}
