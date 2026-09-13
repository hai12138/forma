/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"strconv"
	"sync"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
)

// FakeBusinessModelPort is an in-memory BM validation port for tests.
// Mirrors BusinessModelValidationAdapter: only the master's CurrentRevision is readable.
type FakeBusinessModelPort struct {
	mu      sync.Mutex
	byKey   map[string]*BusinessModelRevisionEvidence // tenant|business|rev
	current map[string]int32                          // tenant|business -> CurrentRevision
}

func NewFakeBusinessModelPort() *FakeBusinessModelPort {
	return &FakeBusinessModelPort{
		byKey:   map[string]*BusinessModelRevisionEvidence{},
		current: map[string]int32{},
	}
}

func (f *FakeBusinessModelPort) Put(ev *BusinessModelRevisionEvidence) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byKey[fakeBMKey(ev.TenantID, ev.BusinessID, ev.Revision)] = ev
	ck := fakeBMCurrentKey(ev.TenantID, ev.BusinessID)
	if cur, ok := f.current[ck]; !ok || ev.Revision >= cur {
		f.current[ck] = ev.Revision
	}
}

// SetCurrentRevision advances/pins the Business master CurrentRevision without deleting old evidence rows.
func (f *FakeBusinessModelPort) SetCurrentRevision(tenantID, businessID string, revision int32) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.current[fakeBMCurrentKey(tenantID, businessID)] = revision
}

func (f *FakeBusinessModelPort) GetBusinessModelRevision(_ context.Context, tenantID, businessID string, revision int32) (*BusinessModelRevisionEvidence, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	cur, ok := f.current[fakeBMCurrentKey(tenantID, businessID)]
	if !ok || cur != revision {
		return nil, entity.ErrBusinessModelNotFound
	}
	ev, ok := f.byKey[fakeBMKey(tenantID, businessID, revision)]
	if !ok {
		return nil, entity.ErrBusinessModelNotFound
	}
	if ev.TenantID != tenantID || ev.BusinessID != businessID || ev.Revision != revision {
		return nil, entity.ErrConsistency
	}
	cp := *ev
	return &cp, nil
}

func fakeBMKey(tenantID, businessID string, rev int32) string {
	return tenantID + "\x00" + businessID + "\x00" + strconv.FormatInt(int64(rev), 10)
}

func fakeBMCurrentKey(tenantID, businessID string) string {
	return tenantID + "\x00" + businessID
}

// FakeContractPort is an in-memory contract validation port for tests.
type FakeContractPort struct {
	mu    sync.Mutex
	byKey map[string]*ContractLogicalDescriptor // tenant|business|contract
}

func NewFakeContractPort() *FakeContractPort {
	return &FakeContractPort{byKey: map[string]*ContractLogicalDescriptor{}}
}

func (f *FakeContractPort) Put(desc *ContractLogicalDescriptor) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byKey[fakeContractKey(desc.TenantID, desc.BusinessID, desc.ContractID)] = desc
}

func (f *FakeContractPort) Remove(tenantID, businessID, contractID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.byKey, fakeContractKey(tenantID, businessID, contractID))
}

func (f *FakeContractPort) GetActiveContractLogicalDescriptor(_ context.Context, tenantID, businessID, contractID string) (*ContractLogicalDescriptor, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	desc, ok := f.byKey[fakeContractKey(tenantID, businessID, contractID)]
	if !ok {
		return nil, entity.ErrContractNotFound
	}
	if desc.BusinessID != businessID || desc.TenantID != tenantID {
		return nil, entity.ErrCrossTenant
	}
	if desc.Status != "ACTIVE" {
		return nil, entity.ErrContractNotActive
	}
	return cloneContractDescriptor(desc), nil
}

func cloneContractDescriptor(desc *ContractLogicalDescriptor) *ContractLogicalDescriptor {
	if desc == nil {
		return nil
	}
	cp := *desc
	cp.LogicalSchema = append([]ContractLogicalField(nil), desc.LogicalSchema...)
	cp.QueryCapabilities = append([]string(nil), desc.QueryCapabilities...)
	cp.FilterSchema = append([]ContractFilterFieldSpec(nil), desc.FilterSchema...)
	for i := range cp.FilterSchema {
		cp.FilterSchema[i].Operators = append([]string(nil), desc.FilterSchema[i].Operators...)
	}
	cp.SortSchema = append([]ContractSortFieldSpec(nil), desc.SortSchema...)
	for i := range cp.SortSchema {
		cp.SortSchema[i].Directions = append([]string(nil), desc.SortSchema[i].Directions...)
	}
	if desc.Classification != nil {
		cp.Classification = map[string]string{}
		for k, v := range desc.Classification {
			cp.Classification[k] = v
		}
	}
	return &cp
}

func fakeContractKey(tenantID, businessID, contractID string) string {
	return tenantID + "\x00" + businessID + "\x00" + contractID
}
