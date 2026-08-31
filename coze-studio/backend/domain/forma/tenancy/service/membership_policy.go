/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	formaerrors "github.com/coze-dev/coze-studio/backend/domain/forma/errors"
	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
)

// MembershipPolicy encodes OWNER/ADMIN/MEMBER/VIEWER authorization for membership mutations.
type MembershipPolicy struct{}

// ValidRole reports whether role is one of OWNER/ADMIN/MEMBER/VIEWER.
func ValidRole(role entity.MembershipRole) bool {
	switch role {
	case entity.RoleOwner, entity.RoleAdmin, entity.RoleMember, entity.RoleViewer:
		return true
	default:
		return false
	}
}

// CanAddMember returns nil when actorRole may add a member with newRole.
func (MembershipPolicy) CanAddMember(actorRole, newRole entity.MembershipRole) error {
	if !ValidRole(newRole) {
		return entity.ErrInvalidRole
	}
	switch actorRole {
	case entity.RoleOwner:
		return nil
	case entity.RoleAdmin:
		if newRole == entity.RoleOwner {
			return formaerrors.MembershipForbidden("ADMIN cannot add OWNER")
		}
		return nil
	default:
		return formaerrors.MembershipForbidden("MEMBER or VIEWER cannot manage membership")
	}
}

// CanChangeRole returns nil when actor may change target's role to newRole.
// Last-owner demotion is enforced by the domain service (needs live owner count).
func (MembershipPolicy) CanChangeRole(
	actorRole entity.MembershipRole,
	actorPrincipalID, targetPrincipalID string,
	targetCurrentRole, newRole entity.MembershipRole,
	primaryOwnerPrincipalID string,
) error {
	if !ValidRole(newRole) {
		return entity.ErrInvalidRole
	}
	switch actorRole {
	case entity.RoleMember, entity.RoleViewer:
		return formaerrors.MembershipForbidden("MEMBER or VIEWER cannot manage membership")
	case entity.RoleAdmin:
		if targetCurrentRole == entity.RoleOwner {
			return formaerrors.MembershipForbidden("ADMIN cannot modify OWNER")
		}
		if newRole == entity.RoleOwner {
			return formaerrors.MembershipForbidden("ADMIN cannot promote to OWNER")
		}
		return nil
	case entity.RoleOwner:
		if targetPrincipalID == primaryOwnerPrincipalID && newRole != entity.RoleOwner {
			return entity.ErrPrimaryOwnerImmutable
		}
		return nil
	default:
		return formaerrors.MembershipForbidden("insufficient membership role")
	}
}

// CanRemoveMember returns nil when actor may remove the target membership.
func (MembershipPolicy) CanRemoveMember(
	actorRole entity.MembershipRole,
	actorPrincipalID, targetPrincipalID string,
	targetRole entity.MembershipRole,
	primaryOwnerPrincipalID string,
	activeOwnerCount int,
) error {
	switch actorRole {
	case entity.RoleMember, entity.RoleViewer:
		return formaerrors.MembershipForbidden("MEMBER or VIEWER cannot manage membership")
	case entity.RoleAdmin:
		if targetRole == entity.RoleOwner {
			return formaerrors.MembershipForbidden("ADMIN cannot modify OWNER")
		}
		return nil
	case entity.RoleOwner:
		if targetPrincipalID == primaryOwnerPrincipalID {
			return entity.ErrPrimaryOwnerImmutable
		}
		if targetRole == entity.RoleOwner && activeOwnerCount <= 1 {
			return entity.ErrLastOwner
		}
		return nil
	default:
		return formaerrors.MembershipForbidden("insufficient membership role")
	}
}
