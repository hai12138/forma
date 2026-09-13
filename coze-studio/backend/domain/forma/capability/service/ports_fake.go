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
type FakeBusinessModelPort struct {
	mu    sync.Mutex
	byKey map[string]*BusinessModelRevisionEvidence // tenant|business|rev
}

func NewFakeBusinessModelPort() *FakeBusinessModelPort {
	return &FakeBusinessModelPort{byKey: map[string]*BusinessModelRevisionEvidence{}}
}

func (f *FakeBusinessModelPort) Put(ev *BusinessModelRevisionEvidence) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byKey[fakeBMKey(ev.TenantID, ev.BusinessID, ev.Revision)] = ev
}

func (f *FakeBusinessModelPort) GetBusinessModelRevision(_ context.Context, tenantID, businessID string, revision int32) (*BusinessModelRevisionEvidence, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	ev, ok := f.byKey[fakeBMKey(tenantID, businessID, revision)]
	if !ok {
		return nil, entity.ErrBusinessModelNotFound
	}
	cp := *ev
	return &cp, nil
}

func fakeBMKey(tenantID, businessID string, rev int32) string {
	return tenantID + "\x00" + businessID + "\x00" + strconv.FormatInt(int64(rev), 10)
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
	cp := *desc
	cp.LogicalSchema = append([]ContractLogicalField(nil), desc.LogicalSchema...)
	cp.QueryCapabilities = append([]string(nil), desc.QueryCapabilities...)
	cp.FilterSchema = append([]ContractFilterFieldSpec(nil), desc.FilterSchema...)
	for i := range cp.FilterSchema {
		cp.FilterSchema[i].Operators = append([]string(nil), desc.FilterSchema[i].Operators...)
	}
	return &cp, nil
}

func fakeContractKey(tenantID, businessID, contractID string) string {
	return tenantID + "\x00" + businessID + "\x00" + contractID
}
