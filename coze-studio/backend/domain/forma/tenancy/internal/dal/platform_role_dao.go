/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package dal

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
)

type PlatformRoleDAO struct {
	db *gorm.DB
}

func NewPlatformRoleDAO(db *gorm.DB) *PlatformRoleDAO {
	return &PlatformRoleDAO{db: db}
}

func (d *PlatformRoleDAO) GetByPrincipalID(ctx context.Context, principalID string) (*entity.FormaPlatformRoleAssignment, error) {
	var model PlatformRoleModel
	err := d.db.WithContext(ctx).Where("principal_id = ?", principalID).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toPlatformRoleEntity(&model), nil
}

func (d *PlatformRoleDAO) Create(ctx context.Context, assignment *entity.FormaPlatformRoleAssignment) error {
	model := toPlatformRoleModel(assignment)
	if err := d.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	assignment.ID = model.ID
	return nil
}

func (d *PlatformRoleDAO) Update(ctx context.Context, assignment *entity.FormaPlatformRoleAssignment) error {
	now := time.Now().UTC()
	result := d.db.WithContext(ctx).Model(&PlatformRoleModel{}).
		Where("principal_id = ?", assignment.PrincipalID).
		Updates(map[string]interface{}{
			"role":                     string(assignment.Role),
			"password_change_required": boolToInt8(assignment.PasswordChangeRequired),
			"updated_at":              now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	assignment.UpdatedAt = now
	return nil
}

func (d *PlatformRoleDAO) ListSuperAdmins(ctx context.Context) ([]*entity.FormaPlatformRoleAssignment, error) {
	var models []PlatformRoleModel
	err := d.db.WithContext(ctx).Where("role = ?", string(entity.PlatformRoleSuperAdmin)).Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]*entity.FormaPlatformRoleAssignment, 0, len(models))
	for i := range models {
		out = append(out, toPlatformRoleEntity(&models[i]))
	}
	return out, nil
}

func (d *PlatformRoleDAO) ListAll(ctx context.Context) ([]*entity.FormaPlatformRoleAssignment, error) {
	var models []PlatformRoleModel
	err := d.db.WithContext(ctx).Order("created_at ASC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]*entity.FormaPlatformRoleAssignment, 0, len(models))
	for i := range models {
		out = append(out, toPlatformRoleEntity(&models[i]))
	}
	return out, nil
}

func (d *PlatformRoleDAO) CountActiveSuperAdmins(ctx context.Context, activePrincipalIDs []string) (int, error) {
	if len(activePrincipalIDs) == 0 {
		return 0, nil
	}
	var count int64
	err := d.db.WithContext(ctx).Model(&PlatformRoleModel{}).
		Where("role = ? AND principal_id IN ?", string(entity.PlatformRoleSuperAdmin), activePrincipalIDs).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func toPlatformRoleModel(a *entity.FormaPlatformRoleAssignment) *PlatformRoleModel {
	return &PlatformRoleModel{
		ID:                     a.ID,
		PrincipalID:            a.PrincipalID,
		Role:                   string(a.Role),
		PasswordChangeRequired: boolToInt8(a.PasswordChangeRequired),
		CreatedAt:              a.CreatedAt,
		UpdatedAt:              a.UpdatedAt,
	}
}

func toPlatformRoleEntity(m *PlatformRoleModel) *entity.FormaPlatformRoleAssignment {
	return &entity.FormaPlatformRoleAssignment{
		ID:                     m.ID,
		PrincipalID:            m.PrincipalID,
		Role:                   entity.PlatformRole(m.Role),
		PasswordChangeRequired: m.PasswordChangeRequired != 0,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
	}
}

func boolToInt8(b bool) int8 {
	if b {
		return 1
	}
	return 0
}
