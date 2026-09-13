/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package integration

import (
	"context"
	"errors"

	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	businesssvc "github.com/coze-dev/coze-studio/backend/domain/forma/business/service"
	capentity "github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	capsvc "github.com/coze-dev/coze-studio/backend/domain/forma/capability/service"
	dataentity "github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
	datasvc "github.com/coze-dev/coze-studio/backend/domain/forma/data/service"
)

// BusinessModelValidationAdapter wraps BusinessService for Capability Validate (fingerprint only).
type BusinessModelValidationAdapter struct {
	Business businesssvc.BusinessService
}

func NewBusinessModelValidationAdapter(business businesssvc.BusinessService) *BusinessModelValidationAdapter {
	return &BusinessModelValidationAdapter{Business: business}
}

func (a *BusinessModelValidationAdapter) GetBusinessModelRevision(ctx context.Context, tenantID, businessID string, revision int32) (*capsvc.BusinessModelRevisionEvidence, error) {
	if a == nil || a.Business == nil {
		return nil, capentity.ErrPortsNotConfigured
	}
	// Fail-closed: Capability may only pin the Business master's *current* revision.
	// Capability row locks do not cover the Business aggregate — re-read master each call.
	master, err := a.Business.Get(ctx, tenantID, businessID)
	if err != nil {
		return nil, mapBusinessValidationError(err)
	}
	if master == nil {
		return nil, capentity.ErrBusinessModelNotFound
	}
	if master.TenantID != tenantID || master.BusinessID != businessID {
		return nil, capentity.ErrConsistency
	}
	if master.CurrentRevision != revision {
		return nil, capentity.ErrBusinessModelNotFound
	}
	rev, _, err := a.Business.GetRevision(ctx, tenantID, businessID, revision)
	if err != nil {
		return nil, mapBusinessValidationError(err)
	}
	if rev == nil {
		return nil, capentity.ErrBusinessModelNotFound
	}
	if rev.TenantID != tenantID || rev.BusinessID != businessID || rev.RevisionNo != revision {
		return nil, capentity.ErrConsistency
	}
	return &capsvc.BusinessModelRevisionEvidence{
		TenantID:      tenantID,
		BusinessID:    businessID,
		Revision:      rev.RevisionNo,
		ContentDigest: rev.ContentDigest,
	}, nil
}

func mapBusinessValidationError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, businessentity.ErrNotFound), errors.Is(err, businessentity.ErrRevisionNotFound):
		return capentity.ErrBusinessModelNotFound
	default:
		return capentity.ErrConsistency
	}
}

// ContractValidationAdapter wraps ContractService for Capability Validate (logical descriptor only).
type ContractValidationAdapter struct {
	Contracts datasvc.ContractService
}

func NewContractValidationAdapter(contracts datasvc.ContractService) *ContractValidationAdapter {
	return &ContractValidationAdapter{Contracts: contracts}
}

func (a *ContractValidationAdapter) GetActiveContractLogicalDescriptor(ctx context.Context, tenantID, businessID, contractID string) (*capsvc.ContractLogicalDescriptor, error) {
	if a == nil || a.Contracts == nil {
		return nil, capentity.ErrPortsNotConfigured
	}
	contract, err := a.Contracts.GetContract(ctx, tenantID, contractID)
	if err != nil {
		return nil, mapContractValidationError(err)
	}
	if contract.BusinessID != businessID {
		return nil, capentity.ErrCrossTenant
	}
	desc, err := a.Contracts.GetActiveContractDescriptor(ctx, tenantID, contractID)
	if err != nil {
		return nil, mapContractValidationError(err)
	}
	return mapContractDescriptor(tenantID, contract.BusinessID, desc), nil
}

func mapContractDescriptor(tenantID, businessID string, desc *dataentity.DataContractDescriptor) *capsvc.ContractLogicalDescriptor {
	if desc == nil {
		return nil
	}
	fields := make([]capsvc.ContractLogicalField, 0, len(desc.LogicalSchema.Fields))
	for _, f := range desc.LogicalSchema.Fields {
		fields = append(fields, capsvc.ContractLogicalField{
			LogicalKey: f.LogicalKey, LogicalType: f.LogicalType, Nullable: f.Nullable,
		})
	}
	caps := make([]string, 0, len(desc.QueryCapabilities))
	for _, c := range desc.QueryCapabilities {
		caps = append(caps, string(c))
	}
	filters := make([]capsvc.ContractFilterFieldSpec, 0, len(desc.FilterSchema.Fields))
	for _, f := range desc.FilterSchema.Fields {
		ops := make([]string, 0, len(f.Operators))
		for _, op := range f.Operators {
			ops = append(ops, string(op))
		}
		filters = append(filters, capsvc.ContractFilterFieldSpec{LogicalKey: f.LogicalKey, Operators: ops})
	}
	sorts := make([]capsvc.ContractSortFieldSpec, 0, len(desc.SortSchema.Fields))
	for _, f := range desc.SortSchema.Fields {
		dirs := make([]string, 0, len(f.Directions))
		for _, d := range f.Directions {
			dirs = append(dirs, string(d))
		}
		sorts = append(sorts, capsvc.ContractSortFieldSpec{LogicalKey: f.LogicalKey, Directions: dirs})
	}
	classification := map[string]string{}
	for k, v := range desc.Classification {
		classification[k] = string(v)
	}
	return &capsvc.ContractLogicalDescriptor{
		TenantID: tenantID, BusinessID: businessID,
		ContractID: desc.ContractID, RevisionID: desc.RevisionID, Version: desc.Version,
		BusinessModelRevision: desc.BusinessModelRevision,
		Status:                string(desc.Status),
		LogicalSchema:         fields,
		QueryCapabilities:     caps,
		FilterSchema:          filters,
		SortSchema:            sorts,
		PaginationPolicy: capsvc.ContractPaginationPolicy{
			DefaultLimit: desc.PaginationPolicy.DefaultLimit,
			MaxLimit:     desc.PaginationPolicy.MaxLimit,
		},
		FreshnessPolicy: string(desc.FreshnessPolicy),
		Classification:  classification,
		AccessPolicyRef: desc.AccessPolicyRef,
	}
}

func mapContractValidationError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, dataentity.ErrContractNotFound), errors.Is(err, dataentity.ErrContractRevisionNotFound):
		return capentity.ErrContractNotFound
	case errors.Is(err, dataentity.ErrContractNotActive):
		return capentity.ErrContractNotActive
	default:
		return capentity.ErrConsistency
	}
}

var _ capsvc.BusinessModelValidationPort = (*BusinessModelValidationAdapter)(nil)
var _ capsvc.ContractValidationPort = (*ContractValidationAdapter)(nil)
