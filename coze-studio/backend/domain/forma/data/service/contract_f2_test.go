/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
)

// TestF2StaleDeprecatePreservesNewActivePointer is the exact lifecycle regression for G4-F2:
// v1 ACTIVE → BREAKING STALE (pointer cleared) → activate v2 → Deprecate(v1) must leave v2 ACTIVE.
func TestF2StaleDeprecatePreservesNewActivePointer(t *testing.T) {
	f := newContractFixture(t)
	c, v1 := f.createValidatedActive(t)

	now := time.Now().UTC()
	breakSchema, _ := json.Marshal(entity.PhysicalSchema{
		Name: "readings", Fields: []entity.PhysicalField{{Name: "other", Path: "other", DataType: "string"}},
	})
	breakSnap := &entity.SchemaSnapshot{
		SnapshotID: "f2-break", TenantID: "tenant", SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaJSON: string(breakSchema), Fingerprint: "fp-f2-break", CreatedAt: now,
	}
	if err := f.sources.CreateSnapshot(context.Background(), breakSnap); err != nil {
		t.Fatal(err)
	}
	drift, rev, err := f.svc.EvaluateDrift(context.Background(), &EvaluateDriftInput{
		TenantID: "tenant", RevisionID: v1.RevisionID,
		NewSnapshotIDs: map[string]string{f.snapID: breakSnap.SnapshotID}, ActorID: "actor",
	})
	if err != nil || drift.Severity != entity.DriftSeverityBreaking || rev.Status != entity.ContractStatusStale {
		t.Fatalf("breaking drift: %+v %+v %v", drift, rev, err)
	}
	contract, _ := f.svc.GetContract(context.Background(), "tenant", c.ContractID)
	if contract.ActiveRevisionID != "" {
		t.Fatalf("expected cleared pointer after STALE, got %q", contract.ActiveRevisionID)
	}
	if _, err := f.svc.GetActiveContractDescriptor(context.Background(), "tenant", c.ContractID); err == nil {
		t.Fatal("expected CONTRACT_NOT_ACTIVE after STALE")
	}

	v2, err := f.svc.CreateRevision(context.Background(), &CreateRevisionInput{
		TenantID: "tenant", ContractID: c.ContractID, BaseRevisionID: v1.RevisionID,
		BusinessModelRevision: 7, Name: "v2", Description: "recover",
		RequirementIDs: []string{f.reqID},
		LogicalSchema: entity.ContractLogicalSchema{Fields: []entity.LogicalField{{
			LogicalKey: "temperature", SemanticName: "temperature", LogicalType: "DECIMAL",
			RequirementID: f.reqID, Classification: entity.DataClassificationInternal,
		}}},
		QueryCapabilities: []entity.QueryCapability{entity.QueryCapabilityRead},
		PaginationPolicy:  entity.PaginationPolicy{DefaultLimit: 10, MaxLimit: 50},
		FreshnessPolicy:   entity.FreshnessPolicyOnDemand,
		MappingIDs:        []string{f.mapID}, ActorID: "actor",
	})
	if err != nil {
		t.Fatal(err)
	}
	v2, result, err := f.svc.ValidateRevision(context.Background(), "tenant", v2.RevisionID, "actor")
	if err != nil || result.Status != entity.ValidationStatusPass {
		t.Fatalf("validate v2: %+v %v", result, err)
	}
	v2, err = f.svc.ActivateRevision(context.Background(), "tenant", v2.RevisionID, "actor", "recover")
	if err != nil || v2.Status != entity.ContractStatusActive {
		t.Fatalf("activate v2: %+v %v", v2, err)
	}
	contract, _ = f.svc.GetContract(context.Background(), "tenant", c.ContractID)
	if contract.ActiveRevisionID != v2.RevisionID {
		t.Fatalf("pointer after v2 activate: %q want %q", contract.ActiveRevisionID, v2.RevisionID)
	}
	desc, err := f.svc.GetActiveContractDescriptor(context.Background(), "tenant", c.ContractID)
	if err != nil || desc == nil || desc.RevisionID != v2.RevisionID {
		t.Fatalf("active descriptor after v2: %+v %v", desc, err)
	}

	// Deprecate historical STALE v1 while v2 is ACTIVE — must not clear v2 pointer.
	v1Gone, err := f.svc.DeprecateRevision(context.Background(), "tenant", v1.RevisionID, "actor", "retire-stale")
	if err != nil {
		t.Fatalf("deprecate stale v1 must succeed: %v", err)
	}
	if v1Gone.Status != entity.ContractStatusDeprecated {
		t.Fatalf("v1 status: %s", v1Gone.Status)
	}
	v2Still, err := f.svc.GetRevision(context.Background(), "tenant", v2.RevisionID)
	if err != nil || v2Still.Status != entity.ContractStatusActive {
		t.Fatalf("v2 must remain ACTIVE: %+v %v", v2Still, err)
	}
	contract, _ = f.svc.GetContract(context.Background(), "tenant", c.ContractID)
	if contract.ActiveRevisionID != v2.RevisionID {
		t.Fatalf("active pointer must remain v2 after deprecating stale v1: got %q", contract.ActiveRevisionID)
	}
	desc, err = f.svc.GetActiveContractDescriptor(context.Background(), "tenant", c.ContractID)
	if err != nil || desc.RevisionID != v2.RevisionID {
		t.Fatalf("descriptor must still be v2: %+v %v", desc, err)
	}
}

func TestF2DeprecateStaleWithEmptyPointer(t *testing.T) {
	f := newContractFixture(t)
	c, v1 := f.createValidatedActive(t)
	now := time.Now().UTC()
	breakSchema, _ := json.Marshal(entity.PhysicalSchema{
		Name: "readings", Fields: []entity.PhysicalField{{Name: "other", Path: "other", DataType: "string"}},
	})
	breakSnap := &entity.SchemaSnapshot{
		SnapshotID: "f2-empty", TenantID: "tenant", SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaJSON: string(breakSchema), Fingerprint: "fp-f2-empty", CreatedAt: now,
	}
	if err := f.sources.CreateSnapshot(context.Background(), breakSnap); err != nil {
		t.Fatal(err)
	}
	if _, _, err := f.svc.EvaluateDrift(context.Background(), &EvaluateDriftInput{
		TenantID: "tenant", RevisionID: v1.RevisionID,
		NewSnapshotIDs: map[string]string{f.snapID: breakSnap.SnapshotID}, ActorID: "actor",
	}); err != nil {
		t.Fatal(err)
	}
	contract, _ := f.svc.GetContract(context.Background(), "tenant", c.ContractID)
	if contract.ActiveRevisionID != "" {
		t.Fatalf("expected empty pointer, got %q", contract.ActiveRevisionID)
	}
	gone, err := f.svc.DeprecateRevision(context.Background(), "tenant", v1.RevisionID, "actor", "retire-stale-empty")
	if err != nil || gone.Status != entity.ContractStatusDeprecated {
		t.Fatalf("deprecate stale with empty pointer: %+v %v", gone, err)
	}
	contract, _ = f.svc.GetContract(context.Background(), "tenant", c.ContractID)
	if contract.ActiveRevisionID != "" {
		t.Fatalf("pointer must stay empty: %q", contract.ActiveRevisionID)
	}
}
