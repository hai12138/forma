/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/coze-dev/coze-studio/backend/domain/forma/data/entity"
)

func TestF1ActivateDriftStaleClearPointerChain(t *testing.T) {
	f := newContractFixture(t)
	c, v1 := f.createValidatedActive(t)
	contract, err := f.svc.GetContract(context.Background(), "tenant", c.ContractID)
	if err != nil || contract.ActiveRevisionID != v1.RevisionID {
		t.Fatalf("pointer after activate: %+v %v", contract, err)
	}

	now := time.Now().UTC()
	breakSchema, _ := json.Marshal(entity.PhysicalSchema{
		Name: "readings", Fields: []entity.PhysicalField{{Name: "other", Path: "other", DataType: "string"}},
	})
	breakSnap := &entity.SchemaSnapshot{
		SnapshotID: "f1-break", TenantID: "tenant", SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaJSON: string(breakSchema), Fingerprint: "fp-f1-break", CreatedAt: now,
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
	contract, _ = f.svc.GetContract(context.Background(), "tenant", c.ContractID)
	if contract.ActiveRevisionID != "" {
		t.Fatalf("expected cleared pointer after STALE, got %q", contract.ActiveRevisionID)
	}

	// Activate v2 from a fresh validated revision
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
		t.Fatalf("pointer after v2 activate: %q", contract.ActiveRevisionID)
	}

	v2, err = f.svc.DeprecateRevision(context.Background(), "tenant", v2.RevisionID, "actor", "retire")
	if err != nil || v2.Status != entity.ContractStatusDeprecated {
		t.Fatalf("deprecate v2: %+v %v", v2, err)
	}
	contract, _ = f.svc.GetContract(context.Background(), "tenant", c.ContractID)
	if contract.ActiveRevisionID != "" {
		t.Fatalf("expected cleared pointer after deprecate, got %q", contract.ActiveRevisionID)
	}
}

func TestF1ConcurrentActivateExactlyOneActive(t *testing.T) {
	f := newContractFixture(t)
	_, v1, err := f.svc.CreateContract(context.Background(), f.defaultCreateInput())
	if err != nil {
		t.Fatal(err)
	}
	mkValidated := func(name string) string {
		rev, err := f.svc.CreateRevision(context.Background(), &CreateRevisionInput{
			TenantID: "tenant", ContractID: v1.ContractID, BaseRevisionID: v1.RevisionID,
			BusinessModelRevision: 7, Name: name, Description: "race",
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
		rev, result, err := f.svc.ValidateRevision(context.Background(), "tenant", rev.RevisionID, "actor")
		if err != nil || result.Status != entity.ValidationStatusPass {
			t.Fatalf("validate %s: %+v %v", name, result, err)
		}
		return rev.RevisionID
	}
	a := mkValidated("a")
	b := mkValidated("b")
	start := make(chan struct{})
	var wg sync.WaitGroup
	for _, id := range []string{a, b} {
		id := id
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, _ = f.svc.ActivateRevision(context.Background(), "tenant", id, "actor", "race")
		}()
	}
	close(start)
	wg.Wait()

	contract, err := f.svc.GetContract(context.Background(), "tenant", v1.ContractID)
	if err != nil {
		t.Fatal(err)
	}
	winner := contract.ActiveRevisionID
	if winner != a && winner != b {
		t.Fatalf("pointer %q not one of candidates", winner)
	}
	activeCount := 0
	for _, id := range []string{a, b} {
		rev, _ := f.svc.GetRevision(context.Background(), "tenant", id)
		if rev.Status == entity.ContractStatusActive {
			activeCount++
			if rev.RevisionID != winner {
				t.Fatalf("ACTIVE %s != pointer %s", rev.RevisionID, winner)
			}
		}
	}
	if activeCount != 1 {
		t.Fatalf("expected exactly one ACTIVE, got %d", activeCount)
	}
	// Memory: GetContractForUpdate is GetContract under txn mutex.
	// DAL: GetContractForUpdate uses SELECT ... FOR UPDATE (clause.Locking UPDATE).
}

func TestF1DriftLineageWrongAsset(t *testing.T) {
	f := newContractFixture(t)
	_, active := f.createValidatedActive(t)
	now := time.Now().UTC()
	wrong := &entity.SchemaSnapshot{
		SnapshotID: "wrong-asset", TenantID: "tenant", SourceID: "source", ConnectionID: "connection", AssetID: "other-asset",
		SchemaJSON: mustGetSnapJSON(t, f, f.snapID), Fingerprint: "fp-wrong", CreatedAt: now,
	}
	if err := f.sources.CreateSnapshot(context.Background(), wrong); err != nil {
		t.Fatal(err)
	}
	_, _, err := f.svc.EvaluateDrift(context.Background(), &EvaluateDriftInput{
		TenantID: "tenant", RevisionID: active.RevisionID,
		NewSnapshotIDs: map[string]string{f.snapID: wrong.SnapshotID}, ActorID: "actor",
	})
	if !errors.Is(err, entity.ErrContractDriftInvalid) {
		t.Fatalf("expected ErrContractDriftInvalid, got %v", err)
	}

	okSnap := &entity.SchemaSnapshot{
		SnapshotID: "same-lineage", TenantID: "tenant", SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaJSON: mustGetSnapJSON(t, f, f.snapID), Fingerprint: "fp-v1", CreatedAt: now,
	}
	if err := f.sources.CreateSnapshot(context.Background(), okSnap); err != nil {
		t.Fatal(err)
	}
	drift, _, err := f.svc.EvaluateDrift(context.Background(), &EvaluateDriftInput{
		TenantID: "tenant", RevisionID: active.RevisionID,
		NewSnapshotIDs: map[string]string{f.snapID: okSnap.SnapshotID}, ActorID: "actor",
	})
	if err != nil || drift.Severity != entity.DriftSeverityNoChange {
		t.Fatalf("same lineage: %+v %v", drift, err)
	}
}

func TestF1TypeGuaranteeValidateAndResolve(t *testing.T) {
	f := newContractFixture(t)
	now := time.Now().UTC()

	// DIRECT string physical → logical DECIMAL FAIL
	strSchema, _ := json.Marshal(entity.PhysicalSchema{
		Name: "readings", Fields: []entity.PhysicalField{{Name: "temp_c", Path: "temp_c", DataType: "string"}},
	})
	if err := f.sources.CreateSnapshot(context.Background(), &entity.SchemaSnapshot{
		SnapshotID: "snap-str", TenantID: "tenant", SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaJSON: string(strSchema), Fingerprint: "fp-str", CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	strMap := &entity.SemanticMapping{
		MappingID: "map_str", TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7,
		RequirementID: f.reqID, SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaSnapshotID: "snap-str", TargetFieldPaths: []string{"temp_c"}, MappingType: entity.MappingTypeDirect,
		TransformSpec: json.RawMessage(`{"type":"DIRECT"}`), Status: entity.MappingStatusConfirmed,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := f.mappings.CreateMapping(context.Background(), strMap); err != nil {
		t.Fatal(err)
	}
	in := f.defaultCreateInput()
	in.MappingIDs = []string{strMap.MappingID}
	_, rev, err := f.svc.CreateContract(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	_, result, err := f.svc.ValidateRevision(context.Background(), "tenant", rev.RevisionID, "actor")
	if err != nil || result.Status != entity.ValidationStatusFail {
		t.Fatalf("expected type fail: %+v %v", result, err)
	}
	found := false
	for _, iss := range result.Errors {
		if iss.Code == "LOGICAL_TYPE_MISMATCH" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing LOGICAL_TYPE_MISMATCH: %+v", result.Errors)
	}

	// number → DECIMAL PASS (default fixture)
	f2 := newContractFixture(t)
	_, rev2, err := f2.svc.CreateContract(context.Background(), f2.defaultCreateInput())
	if err != nil {
		t.Fatal(err)
	}
	_, result2, err := f2.svc.ValidateRevision(context.Background(), "tenant", rev2.RevisionID, "actor")
	if err != nil || result2.Status != entity.ValidationStatusPass {
		t.Fatalf("number→DECIMAL: %+v %v", result2, err)
	}

	// CAST STRING→DECIMAL PASS
	f3 := newContractFixture(t)
	castSchema, _ := json.Marshal(entity.PhysicalSchema{
		Name: "readings", Fields: []entity.PhysicalField{{Name: "temp_c", Path: "temp_c", DataType: "string"}},
	})
	if err := f3.sources.CreateSnapshot(context.Background(), &entity.SchemaSnapshot{
		SnapshotID: "snap-cast", TenantID: "tenant", SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaJSON: string(castSchema), Fingerprint: "fp-cast", CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	castMap := &entity.SemanticMapping{
		MappingID: "map_cast", TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7,
		RequirementID: f3.reqID, SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaSnapshotID: "snap-cast", TargetFieldPaths: []string{"temp_c"}, MappingType: entity.MappingTypeCast,
		TransformSpec: json.RawMessage(`{"type":"CAST","from_type":"string","to_type":"DECIMAL"}`),
		Status: entity.MappingStatusConfirmed, CreatedAt: now, UpdatedAt: now,
	}
	if err := f3.mappings.CreateMapping(context.Background(), castMap); err != nil {
		t.Fatal(err)
	}
	in3 := f3.defaultCreateInput()
	in3.MappingIDs = []string{castMap.MappingID}
	_, rev3, err := f3.svc.CreateContract(context.Background(), in3)
	if err != nil {
		t.Fatal(err)
	}
	_, result3, err := f3.svc.ValidateRevision(context.Background(), "tenant", rev3.RevisionID, "actor")
	if err != nil || result3.Status != entity.ValidationStatusPass {
		t.Fatalf("CAST STRING→DECIMAL: %+v %v", result3, err)
	}

	got, err := ResolveMappingOutputContractType(castMap, &entity.PhysicalSchema{
		Fields: []entity.PhysicalField{{Name: "temp_c", Path: "temp_c", DataType: "string"}},
	})
	if err != nil || got != "DECIMAL" {
		t.Fatalf("resolve cast: %q %v", got, err)
	}
}

func TestF1NullabilityGuaranteeValidateAndDrift(t *testing.T) {
	f := newContractFixture(t)
	now := time.Now().UTC()
	schema, _ := json.Marshal(entity.PhysicalSchema{
		Name: "readings",
		Fields: []entity.PhysicalField{
			{Name: "temp_c", Path: "temp_c", DataType: "number", Nullable: true},
			{Name: "sensor_id", Path: "sensor_id", DataType: "string"},
		},
		Relationships: []entity.PhysicalRelationship{{Name: "sensor", FromFields: []string{"sensor_id"}, ToSchema: "sensors", ToFields: []string{"id"}}},
	})
	if err := f.sources.CreateSnapshot(context.Background(), &entity.SchemaSnapshot{
		SnapshotID: "snap-null", TenantID: "tenant", SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaJSON: string(schema), Fingerprint: "fp-null", CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	m := &entity.SemanticMapping{
		MappingID: "map_null", TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7,
		RequirementID: f.reqID, SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaSnapshotID: "snap-null", TargetFieldPaths: []string{"temp_c"}, MappingType: entity.MappingTypeDirect,
		TransformSpec: json.RawMessage(`{"type":"DIRECT"}`), Status: entity.MappingStatusConfirmed,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := f.mappings.CreateMapping(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	in := f.defaultCreateInput()
	in.MappingIDs = []string{m.MappingID}
	in.LogicalSchema.Fields[0].Nullable = false
	_, rev, err := f.svc.CreateContract(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	_, result, err := f.svc.ValidateRevision(context.Background(), "tenant", rev.RevisionID, "actor")
	if err != nil || result.Status != entity.ValidationStatusFail {
		t.Fatalf("expected nullability fail: %+v %v", result, err)
	}
	found := false
	for _, iss := range result.Errors {
		if iss.Code == "NULLABILITY_GUARANTEE_LOST" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing NULLABILITY_GUARANTEE_LOST: %+v", result.Errors)
	}

	// Drift: active non-null logical, new schema makes physical nullable → BREAKING STALE
	f2 := newContractFixture(t)
	c, active := f2.createValidatedActive(t)
	_ = c
	nullSchema, _ := json.Marshal(entity.PhysicalSchema{
		Name: "readings",
		Fields: []entity.PhysicalField{
			{Name: "temp_c", Path: "temp_c", DataType: "number", Nullable: true},
			{Name: "sensor_id", Path: "sensor_id", DataType: "string"},
		},
		Relationships: []entity.PhysicalRelationship{{Name: "sensor", FromFields: []string{"sensor_id"}, ToSchema: "sensors", ToFields: []string{"id"}}},
	})
	nullSnap := &entity.SchemaSnapshot{
		SnapshotID: "snap-drift-null", TenantID: "tenant", SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaJSON: string(nullSchema), Fingerprint: "fp-drift-null", CreatedAt: now,
	}
	if err := f2.sources.CreateSnapshot(context.Background(), nullSnap); err != nil {
		t.Fatal(err)
	}
	drift, rev2, err := f2.svc.EvaluateDrift(context.Background(), &EvaluateDriftInput{
		TenantID: "tenant", RevisionID: active.RevisionID,
		NewSnapshotIDs: map[string]string{f2.snapID: nullSnap.SnapshotID}, ActorID: "actor",
	})
	if err != nil || drift.Severity != entity.DriftSeverityBreaking || rev2.Status != entity.ContractStatusStale {
		t.Fatalf("nullability drift: %+v %+v %v", drift, rev2, err)
	}
	codes := map[string]bool{}
	for _, finding := range drift.Findings {
		codes[finding.Code] = true
	}
	if !codes["NULLABILITY_GUARANTEE_LOST"] {
		t.Fatalf("expected NULLABILITY_GUARANTEE_LOST: %+v", drift.Findings)
	}
}

func TestF1GapUnmappedOnlyCurrentConfirmed(t *testing.T) {
	f := newContractFixture(t)
	_, active := f.createValidatedActive(t)
	now := time.Now().UTC()
	f.business.currentRevision = 8

	reqA := &entity.DataRequirement{
		RequirementID: "req_a", TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 8,
		RequirementKind: entity.KindAttribute, SemanticName: "a", Status: entity.StatusConfirmed,
		CreatedAt: now, UpdatedAt: now,
	}
	reqB := &entity.DataRequirement{
		RequirementID: "req_b", TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 8,
		RequirementKind: entity.KindAttribute, SemanticName: "b", Status: entity.StatusConfirmed,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := f.data.CreateRequirement(context.Background(), reqA); err != nil {
		t.Fatal(err)
	}
	if err := f.data.CreateRequirement(context.Background(), reqB); err != nil {
		t.Fatal(err)
	}
	mapA := &entity.SemanticMapping{
		MappingID: "map_a_r8", TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 8,
		RequirementID: "req_a", SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaSnapshotID: f.snapID, TargetFieldPaths: []string{"temp_c"}, MappingType: entity.MappingTypeDirect,
		TransformSpec: json.RawMessage(`{"type":"DIRECT"}`), Status: entity.MappingStatusConfirmed,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := f.mappings.CreateMapping(context.Background(), mapA); err != nil {
		t.Fatal(err)
	}

	gap, err := f.svc.EvaluateBusinessGap(context.Background(), &EvaluateGapInput{
		TenantID: "tenant", RevisionID: active.RevisionID, ActorID: "actor",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(gap.UnmappedRequirementIDs) != 1 || gap.UnmappedRequirementIDs[0] != "req_b" {
		t.Fatalf("Unmapped want [req_b], got %+v", gap.UnmappedRequirementIDs)
	}
	for _, id := range gap.UnmappedRequirementIDs {
		if id == f.reqID {
			t.Fatalf("old pinned req %q must not appear in Unmapped", f.reqID)
		}
	}
}

func TestF1DescriptorNoPhysicalAndNotActive(t *testing.T) {
	f := newContractFixture(t)
	c, active := f.createValidatedActive(t)
	desc, err := f.svc.GetActiveContractDescriptor(context.Background(), "tenant", c.ContractID)
	if err != nil || desc == nil || desc.Status != entity.ContractStatusActive {
		t.Fatalf("descriptor: %+v %v", desc, err)
	}
	raw, err := json.Marshal(desc)
	if err != nil {
		t.Fatal(err)
	}
	lower := string(raw)
	for _, key := range []string{"source_id", "connection_id", "asset_id", "schema_snapshot_id", "binding_refs", "mapping_id"} {
		needle := `"` + key + `"`
		if containsJSONKey(lower, needle) {
			t.Fatalf("descriptor leaked physical key %s: %s", key, lower)
		}
	}

	_, err = f.svc.DeprecateRevision(context.Background(), "tenant", active.RevisionID, "actor", "retire")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.svc.GetActiveContractDescriptor(context.Background(), "tenant", c.ContractID)
	if !errors.Is(err, entity.ErrContractNotActive) {
		t.Fatalf("expected ErrContractNotActive after deprecate, got %v", err)
	}

	// STALE path
	f2 := newContractFixture(t)
	c2, active2 := f2.createValidatedActive(t)
	now := time.Now().UTC()
	breakSchema, _ := json.Marshal(entity.PhysicalSchema{
		Name: "readings", Fields: []entity.PhysicalField{{Name: "other", Path: "other", DataType: "string"}},
	})
	breakSnap := &entity.SchemaSnapshot{
		SnapshotID: "desc-break", TenantID: "tenant", SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaJSON: string(breakSchema), Fingerprint: "fp-desc-break", CreatedAt: now,
	}
	if err := f2.sources.CreateSnapshot(context.Background(), breakSnap); err != nil {
		t.Fatal(err)
	}
	_, _, err = f2.svc.EvaluateDrift(context.Background(), &EvaluateDriftInput{
		TenantID: "tenant", RevisionID: active2.RevisionID,
		NewSnapshotIDs: map[string]string{f2.snapID: breakSnap.SnapshotID}, ActorID: "actor",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = f2.svc.GetActiveContractDescriptor(context.Background(), "tenant", c2.ContractID)
	if !errors.Is(err, entity.ErrContractNotActive) {
		t.Fatalf("expected ErrContractNotActive when STALE, got %v", err)
	}
}

func containsJSONKey(raw, needle string) bool {
	return len(raw) > 0 && (jsonContains(raw, needle))
}

func jsonContains(raw, needle string) bool {
	return stringContainsFold(raw, needle)
}

func stringContainsFold(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})())
}

func TestF1DuplicateRequirementIDsDenied(t *testing.T) {
	f := newContractFixture(t)
	in := f.defaultCreateInput()
	in.RequirementIDs = []string{f.reqID, f.reqID}
	_, _, err := f.svc.CreateContract(context.Background(), in)
	if !errors.Is(err, entity.ErrContractInvalidPayload) {
		t.Fatalf("create duplicate: %v", err)
	}

	_, rev, err := f.svc.CreateContract(context.Background(), f.defaultCreateInput())
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.svc.CreateRevision(context.Background(), &CreateRevisionInput{
		TenantID: "tenant", ContractID: rev.ContractID, BaseRevisionID: rev.RevisionID,
		BusinessModelRevision: 7, Name: "dup", Description: "x",
		RequirementIDs: []string{f.reqID, f.reqID},
		LogicalSchema: entity.ContractLogicalSchema{Fields: []entity.LogicalField{{
			LogicalKey: "temperature", SemanticName: "temperature", LogicalType: "DECIMAL",
			RequirementID: f.reqID, Classification: entity.DataClassificationInternal,
		}}},
		QueryCapabilities: []entity.QueryCapability{entity.QueryCapabilityRead},
		PaginationPolicy:  entity.PaginationPolicy{DefaultLimit: 10, MaxLimit: 50},
		FreshnessPolicy:   entity.FreshnessPolicyOnDemand,
		MappingIDs:        []string{f.mapID}, ActorID: "actor",
	})
	if !errors.Is(err, entity.ErrContractInvalidPayload) {
		t.Fatalf("create revision duplicate: %v", err)
	}
}

func TestF1NormalizePhysicalDataType(t *testing.T) {
	cases := map[string]string{
		"varchar(64)": "STRING",
		"bigint":      "INTEGER",
		"number":      "DECIMAL",
		"bool":        "BOOLEAN",
		"date":        "DATE",
		"datetime":    "DATETIME",
		"timestamp":   "DATETIME",
		"time":        "TIME",
		"json":        "JSON",
	}
	for in, want := range cases {
		if got := NormalizePhysicalDataType(in); got != want {
			t.Fatalf("%q => %q want %q", in, got, want)
		}
	}
}
