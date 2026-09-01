/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/business/repository"
	"github.com/stretchr/testify/require"
)

type memRepo struct {
	mu        sync.Mutex
	masters   map[string]*entity.BusinessModel
	revisions map[string]map[int32]*entity.BusinessModelRevision
	layouts   map[string]*entity.BusinessModelLayout
}

func newMem() *memRepo {
	return &memRepo{
		masters:   map[string]*entity.BusinessModel{},
		revisions: map[string]map[int32]*entity.BusinessModelRevision{},
		layouts:   map[string]*entity.BusinessModelLayout{},
	}
}

func key(tenant, biz string) string { return tenant + "|" + biz }

func (m *memRepo) CreateMaster(_ context.Context, model *entity.BusinessModel) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *model
	m.masters[key(model.TenantID, model.BusinessID)] = &cp
	return nil
}
func (m *memRepo) GetMaster(_ context.Context, tenantID, businessID string) (*entity.BusinessModel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := m.masters[key(tenantID, businessID)]
	if v == nil {
		return nil, nil
	}
	cp := *v
	return &cp, nil
}
func (m *memRepo) ListMasters(_ context.Context, tenantID string) ([]*entity.BusinessModel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*entity.BusinessModel
	for _, v := range m.masters {
		if v.TenantID == tenantID {
			cp := *v
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (m *memRepo) CASBumpRevision(_ context.Context, tenantID, businessID string, expected, next int32) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := m.masters[key(tenantID, businessID)]
	if v == nil || v.CurrentRevision != expected {
		return false, nil
	}
	v.CurrentRevision = next
	v.UpdatedAt = time.Now().UTC()
	return true, nil
}
func (m *memRepo) TouchUpdatedAt(_ context.Context, tenantID, businessID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if v := m.masters[key(tenantID, businessID)]; v != nil {
		v.UpdatedAt = time.Now().UTC()
	}
	return nil
}
func (m *memRepo) CreateRevision(_ context.Context, r *entity.BusinessModelRevision) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := key(r.TenantID, r.BusinessID)
	if m.revisions[k] == nil {
		m.revisions[k] = map[int32]*entity.BusinessModelRevision{}
	}
	cp := *r
	m.revisions[k][r.RevisionNo] = &cp
	return nil
}
func (m *memRepo) GetRevision(_ context.Context, tenantID, businessID string, rev int32) (*entity.BusinessModelRevision, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := m.revisions[key(tenantID, businessID)][rev]
	if v == nil {
		return nil, nil
	}
	cp := *v
	return &cp, nil
}
func (m *memRepo) ListRevisions(_ context.Context, tenantID, businessID string) ([]*entity.BusinessModelRevision, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*entity.BusinessModelRevision
	for _, v := range m.revisions[key(tenantID, businessID)] {
		cp := *v
		out = append(out, &cp)
	}
	return out, nil
}
func (m *memRepo) UpsertLayout(_ context.Context, l *entity.BusinessModelLayout) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *l
	m.layouts[key(l.TenantID, l.BusinessID)] = &cp
	return nil
}
func (m *memRepo) GetLayout(_ context.Context, tenantID, businessID string) (*entity.BusinessModelLayout, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := m.layouts[key(tenantID, businessID)]
	if v == nil {
		return nil, nil
	}
	cp := *v
	return &cp, nil
}
func (m *memRepo) CASBumpLayout(_ context.Context, tenantID, businessID string, expected, next int32, basedOn int32, layoutJSON, updatedBy string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := m.layouts[key(tenantID, businessID)]
	if v == nil || v.LayoutRevision != expected {
		return false, nil
	}
	v.LayoutRevision = next
	v.BasedOnModelRevision = basedOn
	v.LayoutJSON = layoutJSON
	v.UpdatedBy = updatedBy
	v.UpdatedAt = time.Now().UTC()
	return true, nil
}
func (m *memRepo) Transaction(ctx context.Context, fn func(txRepo repository.BusinessRepository) error) error {
	return fn(m)
}

var _ repository.BusinessRepository = (*memRepo)(nil)

func TestCreateRevisionAndNoChange(t *testing.T) {
	svc := NewBusinessService(&Components{Repo: newMem()})
	ctx := context.Background()
	_, _, _, err := svc.InitBusiness(ctx, "ten_a", "biz_1", "biz_1", "prin_1", sampleModel(), "seed")
	require.NoError(t, err)

	m2 := sampleModel()
	m2.Nodes[0].Name = "报修客户"
	rev, noChange, err := svc.SaveModel(ctx, "ten_a", "biz_1", "prin_1", 1, m2, "rename")
	require.NoError(t, err)
	require.False(t, noChange)
	require.Equal(t, int32(2), rev.RevisionNo)

	rev2, noChange2, err := svc.SaveModel(ctx, "ten_a", "biz_1", "prin_1", 2, m2, "again")
	require.NoError(t, err)
	require.True(t, noChange2)
	require.Equal(t, int32(2), rev2.RevisionNo)
}

func TestVersionConflict(t *testing.T) {
	svc := NewBusinessService(&Components{Repo: newMem()})
	ctx := context.Background()
	_, _, _, err := svc.InitBusiness(ctx, "ten_a", "biz_1", "biz_1", "prin_1", sampleModel(), "seed")
	require.NoError(t, err)

	m2 := sampleModel()
	m2.Nodes = append(m2.Nodes, entity.SemanticNode{ID: "nx", Type: entity.NodeEvent, Name: "故障", SourceMarker: entity.SourceManualModified})
	_, _, err = svc.SaveModel(ctx, "ten_a", "biz_1", "a", 1, m2, "a")
	require.NoError(t, err)

	m3 := sampleModel()
	m3.Nodes[0].Description = "stale"
	_, _, err = svc.SaveModel(ctx, "ten_a", "biz_1", "b", 1, m3, "stale")
	require.ErrorIs(t, err, entity.ErrRevisionConflict)
}

func TestLayoutDoesNotBumpSemantic(t *testing.T) {
	svc := NewBusinessService(&Components{Repo: newMem()})
	ctx := context.Background()
	master, _, _, err := svc.InitBusiness(ctx, "ten_a", "biz_1", "biz_1", "prin_1", sampleModel(), "seed")
	require.NoError(t, err)
	require.Equal(t, int32(1), master.CurrentRevision)

	layout := &entity.ViewLayout{
		NodePositions: map[string]entity.NodePosition{"n1": {X: 10, Y: 20}},
		Zoom:          1.2,
		Mode:          "manual",
	}
	l, err := svc.SaveLayout(ctx, "ten_a", "biz_1", "prin_1", 1, 1, layout)
	require.NoError(t, err)
	require.Equal(t, int32(2), l.LayoutRevision)

	m, _, _, err := svc.GetModel(ctx, "ten_a", "biz_1")
	require.NoError(t, err)
	require.Equal(t, int32(1), m.CurrentRevision)
}

func TestTenantIsolationGet(t *testing.T) {
	svc := NewBusinessService(&Components{Repo: newMem()})
	ctx := context.Background()
	_, _, _, err := svc.InitBusiness(ctx, "ten_a", "biz_1", "biz_1", "prin_1", sampleModel(), "seed")
	require.NoError(t, err)
	_, err = svc.Get(ctx, "ten_b", "biz_1")
	require.ErrorIs(t, err, entity.ErrNotFound)
}

func TestLayoutConflict(t *testing.T) {
	svc := NewBusinessService(&Components{Repo: newMem()})
	ctx := context.Background()
	_, _, _, err := svc.InitBusiness(ctx, "ten_a", "biz_1", "biz_1", "prin_1", sampleModel(), "seed")
	require.NoError(t, err)
	layout := &entity.ViewLayout{NodePositions: map[string]entity.NodePosition{}, Zoom: 1, Mode: "manual"}
	_, err = svc.SaveLayout(ctx, "ten_a", "biz_1", "a", 1, 1, layout)
	require.NoError(t, err)
	_, err = svc.SaveLayout(ctx, "ten_a", "biz_1", "b", 1, 1, layout)
	require.ErrorIs(t, err, entity.ErrLayoutConflict)
}
