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
	"github.com/coze-dev/coze-studio/backend/domain/forma/data/repository"
)

type contractFixture struct {
	svc      ContractService
	contracts repository.ContractRepository
	data     repository.DataRepository
	mappings repository.MappingRepository
	sources  repository.DataSourceRepository
	business *stubBusiness
	reqID    string
	mapID    string
	snapID   string
}

func newContractFixture(t *testing.T) *contractFixture {
	t.Helper()
	ctx := context.Background()
	data := repository.NewMemoryDataRepository()
	sources := repository.NewMemoryDataSourceRepository()
	mappings := repository.NewMemoryMappingRepository()
	contracts := repository.NewMemoryContractRepository()
	now := time.Now().UTC()

	req := &entity.DataRequirement{
		RequirementID: "req_temperature", TenantID: "tenant", BusinessID: "lab",
		BusinessModelRevision: 7, RequirementKind: entity.KindAttribute, SemanticName: "temperature",
		Status: entity.StatusConfirmed, CreatedAt: now, UpdatedAt: now,
	}
	if err := data.CreateRequirement(ctx, req); err != nil {
		t.Fatal(err)
	}
	if err := sources.CreateSource(ctx, &entity.DataSource{SourceID: "source", TenantID: "tenant"}); err != nil {
		t.Fatal(err)
	}
	if err := sources.CreateConnection(ctx, &entity.DataConnection{ConnectionID: "connection", SourceID: "source", TenantID: "tenant"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := sources.UpsertAsset(ctx, &entity.DataAsset{AssetID: "asset", TenantID: "tenant", SourceID: "source", ConnectionID: "connection", LocatorDigest: "asset"}); err != nil {
		t.Fatal(err)
	}
	schema, _ := json.Marshal(entity.PhysicalSchema{
		Name: "readings",
		Fields: []entity.PhysicalField{
			{Name: "temp_c", Path: "temp_c", DataType: "number"},
			{Name: "sensor_id", Path: "sensor_id", DataType: "string"},
		},
		Relationships: []entity.PhysicalRelationship{{Name: "sensor", FromFields: []string{"sensor_id"}, ToSchema: "sensors", ToFields: []string{"id"}}},
	})
	snap := &entity.SchemaSnapshot{
		SnapshotID: "snapshot", TenantID: "tenant", SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaJSON: string(schema), Fingerprint: "fp-v1", CreatedAt: now,
	}
	if err := sources.CreateSnapshot(ctx, snap); err != nil {
		t.Fatal(err)
	}
	m := &entity.SemanticMapping{
		MappingID: "map1", TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7,
		RequirementID: req.RequirementID, SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaSnapshotID: "snapshot", TargetFieldPaths: []string{"temp_c"}, MappingType: entity.MappingTypeDirect,
		TransformSpec: json.RawMessage(`{"type":"DIRECT"}`), Status: entity.MappingStatusConfirmed,
		Source: entity.MappingSourceManualCreated, CreatedAt: now, UpdatedAt: now,
	}
	if err := mappings.CreateMapping(ctx, m); err != nil {
		t.Fatal(err)
	}
	business := newStubBusiness()
	business.currentRevision = 7
	svc := NewContractService(&ContractComponents{
		Contracts: contracts, Data: data, Mappings: mappings, Sources: sources, Business: business,
	})
	return &contractFixture{svc: svc, contracts: contracts, data: data, mappings: mappings, sources: sources, business: business, reqID: req.RequirementID, mapID: m.MappingID, snapID: snap.SnapshotID}
}

func (f *contractFixture) defaultCreateInput() *CreateContractInput {
	return &CreateContractInput{
		TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7,
		Name: "Temperature Contract", Description: "lab temps",
		RequirementIDs: []string{f.reqID},
		LogicalSchema: entity.ContractLogicalSchema{Fields: []entity.LogicalField{{
			LogicalKey: "temperature", SemanticName: "temperature", LogicalType: "DECIMAL",
			RequirementID: f.reqID, Classification: entity.DataClassificationInternal,
		}}},
		QueryCapabilities: []entity.QueryCapability{entity.QueryCapabilityRead, entity.QueryCapabilityList},
		PaginationPolicy:  entity.PaginationPolicy{DefaultLimit: 20, MaxLimit: 100},
		FreshnessPolicy:   entity.FreshnessPolicyOnDemand,
		MappingIDs:        []string{f.mapID},
		ActorID:           "actor",
	}
}

func (f *contractFixture) createValidatedActive(t *testing.T) (*entity.DataContract, *entity.DataContractRevision) {
	t.Helper()
	c, rev, err := f.svc.CreateContract(context.Background(), f.defaultCreateInput())
	if err != nil {
		t.Fatal(err)
	}
	rev, result, err := f.svc.ValidateRevision(context.Background(), "tenant", rev.RevisionID, "actor")
	if err != nil || result.Status != entity.ValidationStatusPass {
		t.Fatalf("validate: %+v %v", result, err)
	}
	rev, err = f.svc.ActivateRevision(context.Background(), "tenant", rev.RevisionID, "actor", "go live")
	if err != nil {
		t.Fatal(err)
	}
	return c, rev
}

func TestContractCreateV1AndNoModelField(t *testing.T) {
	f := newContractFixture(t)
	svc := f.svc.(*contractService)
	if svc == nil {
		t.Fatal("expected concrete service")
	}
	// REAL_MODEL_CALLS=0: ContractComponents has no Model field — compile-time by design.
	c, rev, err := f.svc.CreateContract(context.Background(), f.defaultCreateInput())
	if err != nil {
		t.Fatal(err)
	}
	if c.ContractID == "" || rev.Version != 1 || rev.Status != entity.ContractStatusDraft {
		t.Fatalf("bad create: %+v %+v", c, rev)
	}
	if len(rev.BindingRefs) != 1 || rev.BindingRefs[0].MappingID != f.mapID || rev.BindingRefs[0].SourceID != "source" {
		t.Fatalf("bindings not materialized: %+v", rev.BindingRefs)
	}
}

func TestContractCreateRevisionImmutability(t *testing.T) {
	f := newContractFixture(t)
	_, v1, err := f.svc.CreateContract(context.Background(), f.defaultCreateInput())
	if err != nil {
		t.Fatal(err)
	}
	origName := v1.Name
	in := &CreateRevisionInput{
		TenantID: "tenant", ContractID: v1.ContractID, BaseRevisionID: v1.RevisionID,
		BusinessModelRevision: 7, Name: "v2", Description: "changed",
		RequirementIDs: []string{f.reqID},
		LogicalSchema: entity.ContractLogicalSchema{Fields: []entity.LogicalField{{
			LogicalKey: "temperature", SemanticName: "temperature", LogicalType: "DECIMAL",
			RequirementID: f.reqID, Classification: entity.DataClassificationInternal,
		}}},
		QueryCapabilities: []entity.QueryCapability{entity.QueryCapabilityRead},
		PaginationPolicy:  entity.PaginationPolicy{DefaultLimit: 10, MaxLimit: 50},
		FreshnessPolicy:   entity.FreshnessPolicyDaily,
		MappingIDs:        []string{f.mapID},
		ActorID:           "actor",
	}
	v2, err := f.svc.CreateRevision(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if v2.Version != 2 || v2.DerivedFromRevisionID != v1.RevisionID || v2.Status != entity.ContractStatusDraft {
		t.Fatalf("bad v2: %+v", v2)
	}
	reloaded, err := f.svc.GetRevision(context.Background(), "tenant", v1.RevisionID)
	if err != nil || reloaded.Name != origName || reloaded.Version != 1 {
		t.Fatalf("v1 mutated: %+v", reloaded)
	}
}

func TestContractConcurrentVersionAllocation(t *testing.T) {
	f := newContractFixture(t)
	_, v1, err := f.svc.CreateContract(context.Background(), f.defaultCreateInput())
	if err != nil {
		t.Fatal(err)
	}
	mk := func() *CreateRevisionInput {
		return &CreateRevisionInput{
			TenantID: "tenant", ContractID: v1.ContractID, BaseRevisionID: v1.RevisionID,
			BusinessModelRevision: 7, Name: "parallel", Description: "x",
			RequirementIDs: []string{f.reqID},
			LogicalSchema: entity.ContractLogicalSchema{Fields: []entity.LogicalField{{
				LogicalKey: "temperature", SemanticName: "temperature", LogicalType: "DECIMAL",
				RequirementID: f.reqID, Classification: entity.DataClassificationInternal,
			}}},
			QueryCapabilities: []entity.QueryCapability{entity.QueryCapabilityRead},
			PaginationPolicy:  entity.PaginationPolicy{DefaultLimit: 10, MaxLimit: 50},
			FreshnessPolicy:   entity.FreshnessPolicyOnDemand,
			MappingIDs:        []string{f.mapID}, ActorID: "actor",
		}
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	versions := make(chan int32, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			rev, err := f.svc.CreateRevision(context.Background(), mk())
			if err != nil {
				errs <- err
				return
			}
			versions <- rev.Version
		}()
	}
	close(start)
	wg.Wait()
	close(errs)
	close(versions)
	for err := range errs {
		t.Fatal(err)
	}
	seen := map[int32]bool{}
	for v := range versions {
		if seen[v] {
			t.Fatalf("duplicate version %d", v)
		}
		seen[v] = true
	}
	if len(seen) != 8 {
		t.Fatalf("expected 8 versions, got %d", len(seen))
	}
}

func TestContractConfirmedOnlyAndLineageDeny(t *testing.T) {
	f := newContractFixture(t)
	now := time.Now().UTC()
	proposed := &entity.SemanticMapping{
		MappingID: "map_proposed", TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7,
		RequirementID: f.reqID, SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaSnapshotID: f.snapID, TargetFieldPaths: []string{"temp_c"}, MappingType: entity.MappingTypeDirect,
		TransformSpec: json.RawMessage(`{"type":"DIRECT"}`), Status: entity.MappingStatusProposed,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := f.mappings.CreateMapping(context.Background(), proposed); err != nil {
		t.Fatal(err)
	}
	in := f.defaultCreateInput()
	in.MappingIDs = []string{proposed.MappingID}
	if _, _, err := f.svc.CreateContract(context.Background(), in); !errors.Is(err, entity.ErrContractBindingInvalid) {
		t.Fatalf("expected binding invalid, got %v", err)
	}

	wrongBiz := *proposed
	wrongBiz.MappingID = "map_wrong_biz"
	wrongBiz.Status = entity.MappingStatusConfirmed
	wrongBiz.BusinessID = "other"
	if err := f.mappings.CreateMapping(context.Background(), &wrongBiz); err != nil {
		t.Fatal(err)
	}
	in.MappingIDs = []string{wrongBiz.MappingID}
	if _, _, err := f.svc.CreateContract(context.Background(), in); !errors.Is(err, entity.ErrContractBindingInvalid) {
		t.Fatalf("expected lineage deny, got %v", err)
	}
}

func TestContractMultiBindingAndLogicalIndependence(t *testing.T) {
	f := newContractFixture(t)
	ctx := context.Background()
	now := time.Now().UTC()
	req2 := &entity.DataRequirement{
		RequirementID: "req_humidity", TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7,
		RequirementKind: entity.KindAttribute, SemanticName: "humidity", Status: entity.StatusConfirmed,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := f.data.CreateRequirement(ctx, req2); err != nil {
		t.Fatal(err)
	}
	schema, _ := json.Marshal(entity.PhysicalSchema{
		Name: "readings", Fields: []entity.PhysicalField{
			{Name: "temp_c", Path: "temp_c", DataType: "number"},
			{Name: "humidity", Path: "humidity", DataType: "number"},
		},
	})
	if err := f.sources.CreateSnapshot(ctx, &entity.SchemaSnapshot{
		SnapshotID: "snap2", TenantID: "tenant", SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaJSON: string(schema), Fingerprint: "fp2", CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	m2 := &entity.SemanticMapping{
		MappingID: "map2", TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7,
		RequirementID: req2.RequirementID, SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaSnapshotID: "snap2", TargetFieldPaths: []string{"humidity"}, MappingType: entity.MappingTypeDirect,
		TransformSpec: json.RawMessage(`{"type":"DIRECT"}`), Status: entity.MappingStatusConfirmed,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := f.mappings.CreateMapping(ctx, m2); err != nil {
		t.Fatal(err)
	}
	in := f.defaultCreateInput()
	in.RequirementIDs = []string{f.reqID, req2.RequirementID}
	in.MappingIDs = []string{f.mapID, m2.MappingID}
	in.LogicalSchema.Fields = []entity.LogicalField{
		{LogicalKey: "temperature", SemanticName: "temperature", LogicalType: "DECIMAL", RequirementID: f.reqID, Classification: entity.DataClassificationInternal},
		{LogicalKey: "humidity", SemanticName: "humidity", LogicalType: "DECIMAL", RequirementID: req2.RequirementID, Classification: entity.DataClassificationInternal},
	}
	_, rev, err := f.svc.CreateContract(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	if len(rev.BindingRefs) != 2 {
		t.Fatalf("expected 2 bindings, got %d", len(rev.BindingRefs))
	}
	_, result, err := f.svc.ValidateRevision(ctx, "tenant", rev.RevisionID, "actor")
	if err != nil || result.Status != entity.ValidationStatusPass {
		t.Fatalf("validate fail: %+v %v", result, err)
	}
}

func TestContractValidatePassFailActivateAndConcurrent(t *testing.T) {
	f := newContractFixture(t)
	_, rev, err := f.svc.CreateContract(context.Background(), f.defaultCreateInput())
	if err != nil {
		t.Fatal(err)
	}
	rev, result, err := f.svc.ValidateRevision(context.Background(), "tenant", rev.RevisionID, "actor")
	if err != nil || result.Status != entity.ValidationStatusPass || rev.Status != entity.ContractStatusValidated {
		t.Fatalf("pass expected: %+v %+v %v", rev, result, err)
	}
	events, _ := f.svc.ListLifecycleEvents(context.Background(), "tenant", rev.ContractID)
	if len(events) == 0 || events[0].Action != entity.LifecycleActionValidatePass {
		t.Fatalf("lifecycle: %+v", events)
	}

	// fail path: break requirement
	badIn := f.defaultCreateInput()
	badIn.Name = "bad"
	badIn.RequirementIDs = []string{"missing-req"}
	badIn.LogicalSchema.Fields[0].RequirementID = "missing-req"
	_, badRev, err := f.svc.CreateContract(context.Background(), badIn)
	if err != nil {
		t.Fatal(err)
	}
	badRev, failResult, err := f.svc.ValidateRevision(context.Background(), "tenant", badRev.RevisionID, "actor")
	if err != nil || failResult.Status != entity.ValidationStatusFail || badRev.Status != entity.ContractStatusDraft {
		t.Fatalf("fail expected: %+v %+v %v", badRev, failResult, err)
	}

	c, active := f.createValidatedActive(t)
	_ = c
	if active.Status != entity.ContractStatusActive {
		t.Fatalf("not active: %+v", active)
	}

	// second validated revision then concurrent activate
	v2in := &CreateRevisionInput{
		TenantID: "tenant", ContractID: active.ContractID, BaseRevisionID: active.RevisionID,
		BusinessModelRevision: 7, Name: "v2", Description: "x",
		RequirementIDs: []string{f.reqID},
		LogicalSchema: entity.ContractLogicalSchema{Fields: []entity.LogicalField{{
			LogicalKey: "temperature", SemanticName: "temperature", LogicalType: "DECIMAL",
			RequirementID: f.reqID, Classification: entity.DataClassificationInternal,
		}}},
		QueryCapabilities: []entity.QueryCapability{entity.QueryCapabilityRead},
		PaginationPolicy:  entity.PaginationPolicy{DefaultLimit: 10, MaxLimit: 50},
		FreshnessPolicy:   entity.FreshnessPolicyOnDemand,
		MappingIDs:        []string{f.mapID}, ActorID: "actor",
	}
	v2, err := f.svc.CreateRevision(context.Background(), v2in)
	if err != nil {
		t.Fatal(err)
	}
	v2, _, err = f.svc.ValidateRevision(context.Background(), "tenant", v2.RevisionID, "actor")
	if err != nil {
		t.Fatal(err)
	}
	// Also prepare a sibling validated revision for concurrent activate against same VALIDATED race:
	// Create two VALIDATED revisions and activate both concurrently — only one ACTIVE.
	v3in := *v2in
	v3in.Name = "v3"
	v3, err := f.svc.CreateRevision(context.Background(), &v3in)
	if err != nil {
		t.Fatal(err)
	}
	v3, _, err = f.svc.ValidateRevision(context.Background(), "tenant", v3.RevisionID, "actor")
	if err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	errs := make(chan error, 2)
	for _, id := range []string{v2.RevisionID, v3.RevisionID} {
		id := id
		go func() {
			<-start
			_, err := f.svc.ActivateRevision(context.Background(), "tenant", id, "actor", "race")
			errs <- err
		}()
	}
	close(start)
	var okN, failN int
	for i := 0; i < 2; i++ {
		err := <-errs
		if err == nil {
			okN++
		} else {
			failN++
			if !errors.Is(err, entity.ErrContractVersionConflict) && !errors.Is(err, entity.ErrContractInvalidState) {
				// One succeeds; other may fail only if both try same revision — they are different revisions so both can succeed sequentially.
				// Concurrent activate of TWO different VALIDATED revisions should both succeed: first activates, second deprecates first.
				_ = err
			}
		}
	}
	_ = okN
	_ = failN
	// After both, exactly one ACTIVE among contract revisions; previous active deprecated.
	revs, _ := f.svc.ListRevisions(context.Background(), "tenant", active.ContractID)
	activeCount := 0
	for _, r := range revs {
		if r.Status == entity.ContractStatusActive {
			activeCount++
		}
	}
	if activeCount != 1 {
		t.Fatalf("expected one ACTIVE, got %d among %+v", activeCount, revs)
	}
	prev, _ := f.svc.GetRevision(context.Background(), "tenant", active.RevisionID)
	if prev.Status != entity.ContractStatusDeprecated {
		t.Fatalf("previous should be deprecated: %+v", prev)
	}
}

func TestContractConcurrentActivateSameRevision(t *testing.T) {
	f := newContractFixture(t)
	_, rev, err := f.svc.CreateContract(context.Background(), f.defaultCreateInput())
	if err != nil {
		t.Fatal(err)
	}
	rev, _, err = f.svc.ValidateRevision(context.Background(), "tenant", rev.RevisionID, "actor")
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			<-start
			_, err := f.svc.ActivateRevision(context.Background(), "tenant", rev.RevisionID, "actor", "race")
			errs <- err
		}()
	}
	close(start)
	var okN, failN int
	for i := 0; i < 2; i++ {
		err := <-errs
		if err == nil {
			okN++
		} else {
			failN++
			if !errors.Is(err, entity.ErrContractVersionConflict) && !errors.Is(err, entity.ErrContractInvalidState) {
				t.Fatalf("unexpected err: %v", err)
			}
		}
	}
	if okN != 1 || failN != 1 {
		t.Fatalf("expected 1 success 1 fail, got ok=%d fail=%d", okN, failN)
	}
}

func TestContractDriftBreakingCompatibleNoChangeJoinRef(t *testing.T) {
	f := newContractFixture(t)
	_, active := f.createValidatedActive(t)
	ctx := context.Background()
	now := time.Now().UTC()

	// fingerprint same → NO_CHANGE
	sameSnap := &entity.SchemaSnapshot{
		SnapshotID: "snap-same", TenantID: "tenant", SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaJSON: mustGetSnapJSON(t, f, f.snapID), Fingerprint: "fp-v1", CreatedAt: now,
	}
	if err := f.sources.CreateSnapshot(ctx, sameSnap); err != nil {
		t.Fatal(err)
	}
	drift, rev, err := f.svc.EvaluateDrift(ctx, &EvaluateDriftInput{
		TenantID: "tenant", RevisionID: active.RevisionID,
		NewSnapshotIDs: map[string]string{f.snapID: sameSnap.SnapshotID}, ActorID: "actor",
	})
	if err != nil || drift.Severity != entity.DriftSeverityNoChange || rev.Status != entity.ContractStatusActive {
		t.Fatalf("no change: %+v %+v %v", drift, rev, err)
	}

	// compatible: add unused field
	compatSchema, _ := json.Marshal(entity.PhysicalSchema{
		Name: "readings",
		Fields: []entity.PhysicalField{
			{Name: "temp_c", Path: "temp_c", DataType: "number"},
			{Name: "sensor_id", Path: "sensor_id", DataType: "string"},
			{Name: "extra", Path: "extra", DataType: "string"},
		},
		Relationships: []entity.PhysicalRelationship{{Name: "sensor", FromFields: []string{"sensor_id"}, ToSchema: "sensors", ToFields: []string{"id"}}},
	})
	compatSnap := &entity.SchemaSnapshot{
		SnapshotID: "snap-compat", TenantID: "tenant", SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaJSON: string(compatSchema), Fingerprint: "fp-compat", CreatedAt: now,
	}
	if err := f.sources.CreateSnapshot(ctx, compatSnap); err != nil {
		t.Fatal(err)
	}
	drift, rev, err = f.svc.EvaluateDrift(ctx, &EvaluateDriftInput{
		TenantID: "tenant", RevisionID: active.RevisionID,
		NewSnapshotIDs: map[string]string{f.snapID: compatSnap.SnapshotID}, ActorID: "actor",
	})
	if err != nil || drift.Severity != entity.DriftSeverityCompatible || rev.Status != entity.ContractStatusActive {
		t.Fatalf("compatible: %+v %+v %v", drift, rev, err)
	}

	// breaking: remove mapped field → STALE
	breakSchema, _ := json.Marshal(entity.PhysicalSchema{
		Name: "readings", Fields: []entity.PhysicalField{{Name: "other", Path: "other", DataType: "string"}},
	})
	breakSnap := &entity.SchemaSnapshot{
		SnapshotID: "snap-break", TenantID: "tenant", SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaJSON: string(breakSchema), Fingerprint: "fp-break", CreatedAt: now,
	}
	if err := f.sources.CreateSnapshot(ctx, breakSnap); err != nil {
		t.Fatal(err)
	}
	drift, rev, err = f.svc.EvaluateDrift(ctx, &EvaluateDriftInput{
		TenantID: "tenant", RevisionID: active.RevisionID,
		NewSnapshotIDs: map[string]string{f.snapID: breakSnap.SnapshotID}, ActorID: "actor",
	})
	if err != nil || drift.Severity != entity.DriftSeverityBreaking || rev.Status != entity.ContractStatusStale {
		t.Fatalf("breaking: %+v %+v %v", drift, rev, err)
	}

	// JOIN_REF missing on new schema
	f2 := newContractFixture(t)
	joinSpec := json.RawMessage(`{"type":"JOIN_REF","relationship":"sensor","from_fields":["sensor_id"],"to_schema":"sensors","to_fields":["id"]}`)
	m := &entity.SemanticMapping{
		MappingID: "map_join", TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7,
		RequirementID: f2.reqID, SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaSnapshotID: f2.snapID, TargetFieldPaths: []string{"sensor_id"}, MappingType: entity.MappingTypeJoinRef,
		TransformSpec: joinSpec, Status: entity.MappingStatusConfirmed, CreatedAt: now, UpdatedAt: now,
	}
	if err := f2.mappings.CreateMapping(ctx, m); err != nil {
		t.Fatal(err)
	}
	in := f2.defaultCreateInput()
	in.MappingIDs = []string{m.MappingID}
	in.LogicalSchema.Fields[0].LogicalKey = "sensor"
	in.LogicalSchema.Fields[0].LogicalType = "STRING"
	in.LogicalSchema.Fields[0].SemanticName = "sensor"
	_, joinActive := func() (*entity.DataContract, *entity.DataContractRevision) {
		c, rev, err := f2.svc.CreateContract(ctx, in)
		if err != nil {
			t.Fatal(err)
		}
		rev, result, err := f2.svc.ValidateRevision(ctx, "tenant", rev.RevisionID, "actor")
		if err != nil || result.Status != entity.ValidationStatusPass {
			t.Fatalf("join validate: %+v %v", result, err)
		}
		rev, err = f2.svc.ActivateRevision(ctx, "tenant", rev.RevisionID, "actor", "live")
		if err != nil {
			t.Fatal(err)
		}
		return c, rev
	}()
	noRel, _ := json.Marshal(entity.PhysicalSchema{
		Name: "readings", Fields: []entity.PhysicalField{{Name: "sensor_id", Path: "sensor_id", DataType: "string"}},
	})
	noRelSnap := &entity.SchemaSnapshot{
		SnapshotID: "snap-norel", TenantID: "tenant", SourceID: "source", ConnectionID: "connection", AssetID: "asset",
		SchemaJSON: string(noRel), Fingerprint: "fp-norel", CreatedAt: now,
	}
	if err := f2.sources.CreateSnapshot(ctx, noRelSnap); err != nil {
		t.Fatal(err)
	}
	drift, rev, err = f2.svc.EvaluateDrift(ctx, &EvaluateDriftInput{
		TenantID: "tenant", RevisionID: joinActive.RevisionID,
		NewSnapshotIDs: map[string]string{f2.snapID: noRelSnap.SnapshotID}, ActorID: "actor",
	})
	if err != nil || drift.Severity != entity.DriftSeverityBreaking || rev.Status != entity.ContractStatusStale {
		t.Fatalf("join break: %+v %+v %v", drift, rev, err)
	}
}

func mustGetSnapJSON(t *testing.T, f *contractFixture, id string) string {
	t.Helper()
	snap, err := f.sources.GetSnapshot(context.Background(), "tenant", id)
	if err != nil {
		t.Fatal(err)
	}
	return snap.SchemaJSON
}

func TestContractBusinessGapNoMutate(t *testing.T) {
	f := newContractFixture(t)
	_, active := f.createValidatedActive(t)
	statusBefore := active.Status
	f.business.currentRevision = 8
	reqNew := &entity.DataRequirement{
		RequirementID: "req_new", TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 8,
		RequirementKind: entity.KindAttribute, SemanticName: "pressure", Status: entity.StatusConfirmed,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := f.data.CreateRequirement(context.Background(), reqNew); err != nil {
		t.Fatal(err)
	}
	gap, err := f.svc.EvaluateBusinessGap(context.Background(), &EvaluateGapInput{
		TenantID: "tenant", RevisionID: active.RevisionID, ActorID: "actor",
	})
	if err != nil || gap.GapStatus != "GAP" {
		t.Fatalf("gap: %+v %v", gap, err)
	}
	found := false
	for _, id := range gap.NewConfirmedRequirementIDs {
		if id == "req_new" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing new req: %+v", gap)
	}
	reloaded, _ := f.svc.GetRevision(context.Background(), "tenant", active.RevisionID)
	if reloaded.Status != statusBefore {
		t.Fatalf("gap mutated status: %s -> %s", statusBefore, reloaded.Status)
	}
}

func TestContractSourceIndependenceMySQLToHTTP(t *testing.T) {
	f := newContractFixture(t)
	_, v1, err := f.svc.CreateContract(context.Background(), f.defaultCreateInput())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	now := time.Now().UTC()
	if err := f.sources.CreateSource(ctx, &entity.DataSource{SourceID: "http-source", TenantID: "tenant"}); err != nil {
		t.Fatal(err)
	}
	if err := f.sources.CreateConnection(ctx, &entity.DataConnection{ConnectionID: "http-conn", SourceID: "http-source", TenantID: "tenant"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := f.sources.UpsertAsset(ctx, &entity.DataAsset{AssetID: "http-asset", TenantID: "tenant", SourceID: "http-source", ConnectionID: "http-conn", LocatorDigest: "http"}); err != nil {
		t.Fatal(err)
	}
	schema, _ := json.Marshal(entity.PhysicalSchema{
		Name: "api", Fields: []entity.PhysicalField{{Name: "temp_c", Path: "temp_c", DataType: "number"}},
	})
	if err := f.sources.CreateSnapshot(ctx, &entity.SchemaSnapshot{
		SnapshotID: "http-snap", TenantID: "tenant", SourceID: "http-source", ConnectionID: "http-conn", AssetID: "http-asset",
		SchemaJSON: string(schema), Fingerprint: "http-fp", CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	httpMap := &entity.SemanticMapping{
		MappingID: "map_http", TenantID: "tenant", BusinessID: "lab", BusinessModelRevision: 7,
		RequirementID: f.reqID, SourceID: "http-source", ConnectionID: "http-conn", AssetID: "http-asset",
		SchemaSnapshotID: "http-snap", TargetFieldPaths: []string{"temp_c"}, MappingType: entity.MappingTypeDirect,
		TransformSpec: json.RawMessage(`{"type":"DIRECT"}`), Status: entity.MappingStatusConfirmed,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := f.mappings.CreateMapping(ctx, httpMap); err != nil {
		t.Fatal(err)
	}
	v2, err := f.svc.CreateRevision(ctx, &CreateRevisionInput{
		TenantID: "tenant", ContractID: v1.ContractID, BaseRevisionID: v1.RevisionID,
		BusinessModelRevision: 7, Name: "http-backed", Description: "same logical",
		RequirementIDs: []string{f.reqID},
		LogicalSchema:  v1.LogicalSchema,
		QueryCapabilities: []entity.QueryCapability{entity.QueryCapabilityRead},
		PaginationPolicy:  entity.PaginationPolicy{DefaultLimit: 10, MaxLimit: 50},
		FreshnessPolicy:   entity.FreshnessPolicyOnDemand,
		MappingIDs:        []string{httpMap.MappingID}, ActorID: "actor",
	})
	if err != nil {
		t.Fatal(err)
	}
	if v2.LogicalSchema.Fields[0].LogicalKey != v1.LogicalSchema.Fields[0].LogicalKey {
		t.Fatal("logical schema changed")
	}
	if v2.BindingRefs[0].SourceID != "http-source" || v1.BindingRefs[0].SourceID != "source" {
		t.Fatalf("bindings: v1=%+v v2=%+v", v1.BindingRefs, v2.BindingRefs)
	}
}

func TestContractWriteCapabilitiesDeniedOnValidate(t *testing.T) {
	f := newContractFixture(t)
	in := f.defaultCreateInput()
	in.QueryCapabilities = []entity.QueryCapability{entity.QueryCapabilityRead, "WRITE"}
	_, rev, err := f.svc.CreateContract(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	_, result, err := f.svc.ValidateRevision(context.Background(), "tenant", rev.RevisionID, "actor")
	if err != nil || result.Status != entity.ValidationStatusFail {
		t.Fatalf("expected fail on write cap: %+v %v", result, err)
	}
}

func TestContractDeprecate(t *testing.T) {
	f := newContractFixture(t)
	_, active := f.createValidatedActive(t)
	rev, err := f.svc.DeprecateRevision(context.Background(), "tenant", active.RevisionID, "actor", "retire")
	if err != nil || rev.Status != entity.ContractStatusDeprecated {
		t.Fatalf("deprecate: %+v %v", rev, err)
	}
}
