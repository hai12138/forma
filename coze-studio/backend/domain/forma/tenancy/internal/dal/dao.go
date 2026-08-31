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

// ---------------------------------------------------------------------------
// Principal
// ---------------------------------------------------------------------------

type PrincipalDAO struct {
	db *gorm.DB
}

func NewPrincipalDAO(db *gorm.DB) *PrincipalDAO {
	return &PrincipalDAO{db: db}
}

func (d *PrincipalDAO) Create(ctx context.Context, p *entity.Principal) error {
	model := toPrincipalModel(p)
	if err := d.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	p.ID = model.ID
	return nil
}

func (d *PrincipalDAO) GetByPrincipalID(ctx context.Context, principalID string) (*entity.Principal, error) {
	var model PrincipalModel
	err := d.db.WithContext(ctx).Where("principal_id = ?", principalID).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toPrincipalEntity(&model), nil
}

func (d *PrincipalDAO) GetByCozeUserID(ctx context.Context, cozeUserID int64) (*entity.Principal, error) {
	var model PrincipalModel
	err := d.db.WithContext(ctx).Where("coze_user_id = ?", cozeUserID).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toPrincipalEntity(&model), nil
}

func (d *PrincipalDAO) GetByProviderSubject(ctx context.Context, provider, externalSubject string) (*entity.Principal, error) {
	var model PrincipalModel
	err := d.db.WithContext(ctx).
		Where("provider = ? AND external_subject = ?", provider, externalSubject).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toPrincipalEntity(&model), nil
}

func toPrincipalModel(p *entity.Principal) *PrincipalModel {
	return &PrincipalModel{
		ID:              p.ID,
		PrincipalID:     p.PrincipalID,
		PrincipalType:   string(p.PrincipalType),
		Provider:        p.Provider,
		ExternalSubject: p.ExternalSubject,
		CozeUserID:      p.CozeUserID,
		DisplayName:     p.DisplayName,
		Status:          string(p.Status),
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

func toPrincipalEntity(m *PrincipalModel) *entity.Principal {
	return &entity.Principal{
		ID:              m.ID,
		PrincipalID:     m.PrincipalID,
		PrincipalType:   entity.PrincipalType(m.PrincipalType),
		Provider:        m.Provider,
		ExternalSubject: m.ExternalSubject,
		CozeUserID:      m.CozeUserID,
		DisplayName:     m.DisplayName,
		Status:          entity.PrincipalStatus(m.Status),
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

// ---------------------------------------------------------------------------
// Tenant
// ---------------------------------------------------------------------------

type TenantDAO struct {
	db *gorm.DB
}

func NewTenantDAO(db *gorm.DB) *TenantDAO {
	return &TenantDAO{db: db}
}

func (d *TenantDAO) Create(ctx context.Context, t *entity.Tenant) error {
	model := toTenantModel(t)
	if err := d.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	t.ID = model.ID
	return nil
}

func (d *TenantDAO) GetByTenantID(ctx context.Context, tenantID string) (*entity.Tenant, error) {
	var model TenantModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toTenantEntity(&model), nil
}

// Update applies fields with optimistic lock on revision.
// On success, t.Revision is incremented to expectedRevision+1.
func (d *TenantDAO) Update(ctx context.Context, t *entity.Tenant, expectedRevision int32) error {
	now := time.Now().UTC()
	result := d.db.WithContext(ctx).Model(&TenantModel{}).
		Where("tenant_id = ? AND revision = ? AND deleted_at IS NULL", t.TenantID, expectedRevision).
		Updates(map[string]interface{}{
			"tenant_key":         t.TenantKey,
			"name":               t.Name,
			"display_name":       t.DisplayName,
			"status":             string(t.Status),
			"owner_principal_id": t.OwnerPrincipalID,
			"revision":           expectedRevision + 1,
			"updated_at":         now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entity.ErrRevisionConflict
	}
	t.Revision = expectedRevision + 1
	t.UpdatedAt = now
	return nil
}

func (d *TenantDAO) ListByPrincipalID(ctx context.Context, principalID string) ([]*entity.Tenant, error) {
	var models []TenantModel
	err := d.db.WithContext(ctx).
		Table("forma_tenant AS t").
		Joins("INNER JOIN forma_tenant_membership AS m ON m.tenant_id = t.tenant_id").
		Where("m.principal_id = ? AND m.status = ? AND t.deleted_at IS NULL", principalID, string(entity.MembershipActive)).
		Order("t.updated_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]*entity.Tenant, 0, len(models))
	for i := range models {
		out = append(out, toTenantEntity(&models[i]))
	}
	return out, nil
}

func toTenantModel(t *entity.Tenant) *TenantModel {
	return &TenantModel{
		ID:               t.ID,
		TenantID:         t.TenantID,
		TenantKey:        t.TenantKey,
		Name:             t.Name,
		DisplayName:      t.DisplayName,
		Status:           string(t.Status),
		OwnerPrincipalID: t.OwnerPrincipalID,
		Revision:         t.Revision,
		CreatedAt:        t.CreatedAt,
		UpdatedAt:        t.UpdatedAt,
		DeletedAt:        t.DeletedAt,
	}
}

func toTenantEntity(m *TenantModel) *entity.Tenant {
	return &entity.Tenant{
		ID:               m.ID,
		TenantID:         m.TenantID,
		TenantKey:        m.TenantKey,
		Name:             m.Name,
		DisplayName:      m.DisplayName,
		Status:           entity.TenantStatus(m.Status),
		OwnerPrincipalID: m.OwnerPrincipalID,
		Revision:         m.Revision,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
		DeletedAt:        m.DeletedAt,
	}
}

// ---------------------------------------------------------------------------
// Membership
// ---------------------------------------------------------------------------

type MembershipDAO struct {
	db *gorm.DB
}

func NewMembershipDAO(db *gorm.DB) *MembershipDAO {
	return &MembershipDAO{db: db}
}

func (d *MembershipDAO) Create(ctx context.Context, m *entity.Membership) error {
	model := toMembershipModel(m)
	if err := d.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	m.ID = model.ID
	return nil
}

func (d *MembershipDAO) Get(ctx context.Context, tenantID, principalID string) (*entity.Membership, error) {
	var model MembershipModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ? AND principal_id = ?", tenantID, principalID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toMembershipEntity(&model), nil
}

func (d *MembershipDAO) UpdateRole(ctx context.Context, tenantID, principalID string, role entity.MembershipRole, expectedRevision int32) (*entity.Membership, error) {
	now := time.Now().UTC()
	result := d.db.WithContext(ctx).Model(&MembershipModel{}).
		Where("tenant_id = ? AND principal_id = ? AND revision = ?", tenantID, principalID, expectedRevision).
		Updates(map[string]interface{}{
			"role":       string(role),
			"revision":   expectedRevision + 1,
			"updated_at": now,
		})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, entity.ErrRevisionConflict
	}
	return d.Get(ctx, tenantID, principalID)
}

func (d *MembershipDAO) ListByTenant(ctx context.Context, tenantID string) ([]*entity.Membership, error) {
	var models []MembershipModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("joined_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]*entity.Membership, 0, len(models))
	for i := range models {
		out = append(out, toMembershipEntity(&models[i]))
	}
	return out, nil
}

func (d *MembershipDAO) ListByPrincipal(ctx context.Context, principalID string) ([]*entity.Membership, error) {
	var models []MembershipModel
	err := d.db.WithContext(ctx).
		Where("principal_id = ?", principalID).
		Order("joined_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]*entity.Membership, 0, len(models))
	for i := range models {
		out = append(out, toMembershipEntity(&models[i]))
	}
	return out, nil
}

func (d *MembershipDAO) SoftRemove(ctx context.Context, tenantID, principalID string) error {
	now := time.Now().UTC()
	result := d.db.WithContext(ctx).Model(&MembershipModel{}).
		Where("tenant_id = ? AND principal_id = ?", tenantID, principalID).
		Updates(map[string]interface{}{
			"status":     string(entity.MembershipRemoved),
			"updated_at": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func toMembershipModel(m *entity.Membership) *MembershipModel {
	return &MembershipModel{
		ID:          m.ID,
		TenantID:    m.TenantID,
		PrincipalID: m.PrincipalID,
		Role:        string(m.Role),
		Status:      string(m.Status),
		Revision:    m.Revision,
		JoinedAt:    m.JoinedAt,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func toMembershipEntity(m *MembershipModel) *entity.Membership {
	return &entity.Membership{
		ID:          m.ID,
		TenantID:    m.TenantID,
		PrincipalID: m.PrincipalID,
		Role:        entity.MembershipRole(m.Role),
		Status:      entity.MembershipStatus(m.Status),
		Revision:    m.Revision,
		JoinedAt:    m.JoinedAt,
		CreatedBy:   m.CreatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ---------------------------------------------------------------------------
// SpaceRef
// ---------------------------------------------------------------------------

type SpaceRefDAO struct {
	db *gorm.DB
}

func NewSpaceRefDAO(db *gorm.DB) *SpaceRefDAO {
	return &SpaceRefDAO{db: db}
}

func (d *SpaceRefDAO) GetBySpaceID(ctx context.Context, cozeSpaceID int64) (*entity.TenantSpaceRef, error) {
	var model TenantSpaceRefModel
	err := d.db.WithContext(ctx).
		Where("coze_space_id = ?", cozeSpaceID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toSpaceRefEntity(&model), nil
}

func (d *SpaceRefDAO) UpsertBind(ctx context.Context, ref *entity.TenantSpaceRef) error {
	if ref == nil || ref.CozeSpaceID == 0 {
		return entity.ErrNotFound
	}
	now := time.Now().UTC()
	existing, err := d.GetBySpaceID(ctx, ref.CozeSpaceID)
	if err != nil {
		return err
	}
	if existing == nil {
		if ref.CreatedAt.IsZero() {
			ref.CreatedAt = now
		}
		ref.UpdatedAt = now
		if ref.Status == "" {
			ref.Status = entity.SpaceRefActive
		}
		model := toSpaceRefModel(ref)
		if err := d.db.WithContext(ctx).Create(model).Error; err != nil {
			return err
		}
		ref.ID = model.ID
		return nil
	}
	result := d.db.WithContext(ctx).Model(&TenantSpaceRefModel{}).
		Where("coze_space_id = ?", ref.CozeSpaceID).
		Updates(map[string]interface{}{
			"tenant_id":  ref.TenantID,
			"purpose":    string(ref.Purpose),
			"status":     string(entity.SpaceRefActive),
			"updated_at": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	ref.ID = existing.ID
	ref.Status = entity.SpaceRefActive
	ref.CreatedAt = existing.CreatedAt
	ref.UpdatedAt = now
	return nil
}

func (d *SpaceRefDAO) ListByTenant(ctx context.Context, tenantID string) ([]*entity.TenantSpaceRef, error) {
	var models []TenantSpaceRefModel
	err := d.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]*entity.TenantSpaceRef, 0, len(models))
	for i := range models {
		out = append(out, toSpaceRefEntity(&models[i]))
	}
	return out, nil
}

func (d *SpaceRefDAO) Deactivate(ctx context.Context, tenantID string, cozeSpaceID int64) error {
	now := time.Now().UTC()
	result := d.db.WithContext(ctx).Model(&TenantSpaceRefModel{}).
		Where("tenant_id = ? AND coze_space_id = ?", tenantID, cozeSpaceID).
		Updates(map[string]interface{}{
			"status":     string(entity.SpaceRefInactive),
			"updated_at": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func toSpaceRefModel(r *entity.TenantSpaceRef) *TenantSpaceRefModel {
	return &TenantSpaceRefModel{
		ID:          r.ID,
		TenantID:    r.TenantID,
		CozeSpaceID: r.CozeSpaceID,
		Purpose:     string(r.Purpose),
		Status:      string(r.Status),
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func toSpaceRefEntity(m *TenantSpaceRefModel) *entity.TenantSpaceRef {
	return &entity.TenantSpaceRef{
		ID:          m.ID,
		TenantID:    m.TenantID,
		CozeSpaceID: m.CozeSpaceID,
		Purpose:     entity.SpacePurpose(m.Purpose),
		Status:      entity.SpaceRefStatus(m.Status),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ---------------------------------------------------------------------------
// Audit
// ---------------------------------------------------------------------------

type AuditDAO struct {
	db *gorm.DB
}

func NewAuditDAO(db *gorm.DB) *AuditDAO {
	return &AuditDAO{db: db}
}

func (d *AuditDAO) Create(ctx context.Context, e *entity.AuditEvent) error {
	model := toAuditModel(e)
	if err := d.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	e.ID = model.ID
	return nil
}

func toAuditModel(e *entity.AuditEvent) *AuditEventModel {
	return &AuditEventModel{
		ID:          e.ID,
		TenantID:    e.TenantID,
		PrincipalID: e.PrincipalID,
		Action:      e.Action,
		Resource:    e.Resource,
		RequestID:   e.RequestID,
		CreatedAt:   e.CreatedAt,
	}
}
