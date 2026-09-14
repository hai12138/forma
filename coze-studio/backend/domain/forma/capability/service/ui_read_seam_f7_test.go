/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	assetentity "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/fixture"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/repository"
)

// countingAssets spies ListCapabilityAssetsByTenant vs GetCapabilityAsset for F7 batch-read seams.
type countingAssets struct {
	inner     *MemoryAssetProjection
	listCalls atomic.Int64
	getCalls  atomic.Int64
}

func (c *countingAssets) CreateCapabilityAsset(ctx context.Context, asset *assetentity.AssetRef) error {
	return c.inner.CreateCapabilityAsset(ctx, asset)
}
func (c *countingAssets) UpdateCapabilityProjection(ctx context.Context, tenantID, assetID, name, semanticVersion, contentDigest string, status assetentity.AssetStatus) error {
	return c.inner.UpdateCapabilityProjection(ctx, tenantID, assetID, name, semanticVersion, contentDigest, status)
}
func (c *countingAssets) GetCapabilityAsset(ctx context.Context, tenantID, assetID string) (*assetentity.AssetRef, error) {
	c.getCalls.Add(1)
	return c.inner.GetCapabilityAsset(ctx, tenantID, assetID)
}
func (c *countingAssets) ListCapabilityAssetsByTenant(ctx context.Context, tenantID string) ([]*assetentity.AssetRef, error) {
	c.listCalls.Add(1)
	return c.inner.ListCapabilityAssetsByTenant(ctx, tenantID)
}

var _ AssetProjection = (*countingAssets)(nil)

// f7AssetsRootUoW exposes a spy AssetProjection on the non-tx root view while keeping owned Memory UoW txns.
type f7AssetsRootUoW struct {
	inner  *MemoryUnitOfWork
	assets *countingAssets
}

func (u *f7AssetsRootUoW) Root() repository.CapabilityRepository { return u.inner.Root() }
func (u *f7AssetsRootUoW) Assets() AssetProjection              { return u.assets }
func (u *f7AssetsRootUoW) WithinTransaction(ctx context.Context, fn func(tx CapabilityTx) error) error {
	return u.inner.WithinTransaction(ctx, fn)
}

var _ CapabilityUnitOfWork = (*f7AssetsRootUoW)(nil)

func newF7SpyService(gen ProposalGenerator) (CapabilityService, repository.CapabilityRepository, *countingAssets, *MemoryAssetProjection, *FakeBusinessModelPort, *FakeContractPort) {
	mem := NewMemoryUnitOfWork()
	spy := &countingAssets{inner: mem.AssetsView()}
	uow := &f7AssetsRootUoW{inner: mem, assets: spy}
	bm := NewFakeBusinessModelPort()
	contracts := NewFakeContractPort()
	svc := NewCapabilityService(&Components{UoW: uow, Generator: gen, Business: bm, Contract: contracts})
	return svc, mem.Root(), spy, mem.AssetsView(), bm, contracts
}

func requireConsistencyOrConflict(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	require.True(t, errors.Is(err, entity.ErrConsistency) || errors.Is(err, entity.ErrConflict),
		"want ErrConsistency or ErrConflict, got %v", err)
}

func assertPersistedProjectionMatchesPure(t *testing.T, assets *MemoryAssetProjection, cap *entity.BusinessCapability, revs []*entity.BusinessCapabilityRevision) {
	t.Helper()
	want, err := ProjectCapabilityAssetRef(cap, revs)
	require.NoError(t, err)
	got, err := assets.GetCapabilityAsset(context.Background(), cap.TenantID, cap.CapabilityID)
	require.NoError(t, err)
	require.Equal(t, assetentity.AssetKindCapability, got.Kind)
	require.Equal(t, cap.CapabilityID, got.AssetID)
	require.Equal(t, want.Name, got.Name)
	require.Equal(t, want.SemanticVersion, got.SemanticVersion)
	require.Equal(t, want.ContentDigest, got.ContentDigest)
	require.Equal(t, want.Status, got.Status)
}

func TestF7_ListCapabilities_ExactlyOneTenantAssetBatch(t *testing.T) {
	svc, _, spy, _, _, _ := newF7SpyService(nil)
	payload := fixture.LaboratoryCommandCapability()
	for i, id := range []string{"cap-batch-a", "cap-batch-b", "cap-batch-c"} {
		_, _, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
			TenantID: "t1", BusinessID: "biz-lab", CapabilityID: id,
			ActorID: testActor, OwnerID: testOwnerID, Payload: payload,
		})
		require.NoError(t, err, "create %d", i)
	}
	spy.listCalls.Store(0)
	spy.getCalls.Store(0)

	// Batch seam lives on ListCapabilityAssetsByTenant (app ListCapabilities calls it once).
	assets, err := svc.ListCapabilityAssetsByTenant(context.Background(), "t1")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(assets), 3)
	require.Equal(t, int64(1), spy.listCalls.Load(), "ListCapabilityAssetsByTenant must be one tenant batch")
	require.Equal(t, int64(0), spy.getCalls.Load(), "tenant batch must never N+1 GetCapabilityAsset")

	caps, err := svc.ListCapabilities(context.Background(), "t1", "biz-lab")
	require.NoError(t, err)
	require.Len(t, caps, 3)
	require.Equal(t, int64(1), spy.listCalls.Load(), "ListCapabilities must not issue extra AssetRef batch reads")
}

func TestF7_ProjectCapabilityAssetRef_StatusMatrixPersisted(t *testing.T) {
	svc, repo, _, assets, bm, contracts := newF7SpyService(nil)
	payload := fixture.LaboratoryCommandCapability()

	// Draft → DRAFT
	cap, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", CapabilityID: "cap-status",
		ActorID: testActor, OwnerID: testOwnerID, Payload: payload,
	})
	require.NoError(t, err)
	assertPersistedProjectionMatchesPure(t, assets, cap, []*entity.BusinessCapabilityRevision{rev})
	a, err := assets.GetCapabilityAsset(context.Background(), "t1", "cap-status")
	require.NoError(t, err)
	require.Equal(t, assetentity.AssetStatusDraft, a.Status)

	// Validated → VERIFIED
	seedPortsForRevision(bm, contracts, rev)
	_, _, err = svc.Validate(context.Background(), "t1", rev.RevisionID, testActor)
	require.NoError(t, err)
	cap, err = repo.GetCapability(context.Background(), "t1", "cap-status")
	require.NoError(t, err)
	revs, err := repo.ListRevisions(context.Background(), "t1", "cap-status")
	require.NoError(t, err)
	assertPersistedProjectionMatchesPure(t, assets, cap, revs)
	a, err = assets.GetCapabilityAsset(context.Background(), "t1", "cap-status")
	require.NoError(t, err)
	require.Equal(t, assetentity.AssetStatusVerified, a.Status)

	// Active → RELEASED
	active, err := svc.Activate(context.Background(), "t1", rev.RevisionID, testActor, "go-live")
	require.NoError(t, err)
	require.Equal(t, entity.RevisionActive, active.Status)
	cap, err = repo.GetCapability(context.Background(), "t1", "cap-status")
	require.NoError(t, err)
	revs, err = repo.ListRevisions(context.Background(), "t1", "cap-status")
	require.NoError(t, err)
	assertPersistedProjectionMatchesPure(t, assets, cap, revs)
	a, err = assets.GetCapabilityAsset(context.Background(), "t1", "cap-status")
	require.NoError(t, err)
	require.Equal(t, assetentity.AssetStatusReleased, a.Status)

	// Deprecated → DEPRECATED
	_, err = svc.Deprecate(context.Background(), "t1", rev.RevisionID, testActor, "retire")
	require.NoError(t, err)
	cap, err = repo.GetCapability(context.Background(), "t1", "cap-status")
	require.NoError(t, err)
	revs, err = repo.ListRevisions(context.Background(), "t1", "cap-status")
	require.NoError(t, err)
	assertPersistedProjectionMatchesPure(t, assets, cap, revs)
	a, err = assets.GetCapabilityAsset(context.Background(), "t1", "cap-status")
	require.NoError(t, err)
	require.Equal(t, assetentity.AssetStatusDeprecated, a.Status)

	// Stale → IN_REVIEW via ProjectCapabilityAssetRef + persisted projection update
	// (MarkStale is fail-closed without impact evidence; seed STALE via repo + projectAndUpdateAssets.)
	capS, revS, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", CapabilityID: "cap-stale",
		ActorID: testActor, OwnerID: testOwnerID, Payload: payload,
	})
	require.NoError(t, err)
	forceStatusForTest(t, repo, "t1", revS.RevisionID, entity.RevisionDraft, entity.RevisionStale)
	revS.Status = entity.RevisionStale
	require.NoError(t, projectAndUpdateAssets(context.Background(), assets, capS, []*entity.BusinessCapabilityRevision{revS}))
	wantStale, err := ProjectCapabilityAssetRef(capS, []*entity.BusinessCapabilityRevision{revS})
	require.NoError(t, err)
	require.Equal(t, assetentity.AssetStatusInReview, wantStale.Status)
	aS, err := assets.GetCapabilityAsset(context.Background(), "t1", "cap-stale")
	require.NoError(t, err)
	require.Equal(t, assetentity.AssetStatusInReview, aS.Status)
	require.Equal(t, wantStale.Name, aS.Name)
	require.Equal(t, wantStale.ContentDigest, aS.ContentDigest)
}

func TestF7_ListGet_FailClosedMissingAssetRef(t *testing.T) {
	svc, _, _, assets, _, _ := newF7SpyService(nil)
	payload := fixture.LaboratoryCommandCapability()
	_, _, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", CapabilityID: "cap-missing",
		ActorID: testActor, OwnerID: testOwnerID, Payload: payload,
	})
	require.NoError(t, err)

	assets.mu.Lock()
	delete(assets.assets, assetKey("t1", "cap-missing"))
	assets.mu.Unlock()

	cap, err := svc.GetCapability(context.Background(), "t1", "cap-missing")
	require.NoError(t, err)
	got, err := svc.GetCapabilityAsset(context.Background(), "t1", "cap-missing")
	require.ErrorIs(t, err, entity.ErrNotFound)
	require.Nil(t, got)
	requireConsistencyOrConflict(t, validateCapabilityAssetProjection("t1", cap, nil))

	listed, err := svc.ListCapabilityAssetsByTenant(context.Background(), "t1")
	require.NoError(t, err)
	byID, err := indexCapabilityAssetsByID(listed)
	require.NoError(t, err)
	requireConsistencyOrConflict(t, validateCapabilityAssetProjection("t1", cap, byID[cap.CapabilityID]))
}

func TestF7_ListGet_FailClosedWrongKind(t *testing.T) {
	svc, _, _, assets, _, _ := newF7SpyService(nil)
	payload := fixture.LaboratoryCommandCapability()
	_, _, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", CapabilityID: "cap-kind",
		ActorID: testActor, OwnerID: testOwnerID, Payload: payload,
	})
	require.NoError(t, err)

	assets.mu.Lock()
	assets.assets[assetKey("t1", "cap-kind")].Kind = assetentity.AssetKindBusiness
	assets.mu.Unlock()

	cap, err := svc.GetCapability(context.Background(), "t1", "cap-kind")
	require.NoError(t, err)
	asset, err := svc.GetCapabilityAsset(context.Background(), "t1", "cap-kind")
	require.NoError(t, err)
	requireConsistencyOrConflict(t, validateCapabilityAssetProjection("t1", cap, asset))

	listed, err := svc.ListCapabilityAssetsByTenant(context.Background(), "t1")
	require.NoError(t, err)
	// Wrong kind is excluded from CAPABILITY index — listed capability then missing → fail closed.
	byID, err := indexCapabilityAssetsByID(listed)
	require.NoError(t, err)
	requireConsistencyOrConflict(t, validateCapabilityAssetProjection("t1", cap, byID[cap.CapabilityID]))
}

func TestF7_ListGet_FailClosedWrongAssetID(t *testing.T) {
	svc, _, _, assets, _, _ := newF7SpyService(nil)
	payload := fixture.LaboratoryCommandCapability()
	_, _, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", CapabilityID: "cap-aid",
		ActorID: testActor, OwnerID: testOwnerID, Payload: payload,
	})
	require.NoError(t, err)

	assets.mu.Lock()
	assets.assets[assetKey("t1", "cap-aid")].AssetID = "not-cap-aid"
	assets.mu.Unlock()

	cap, err := svc.GetCapability(context.Background(), "t1", "cap-aid")
	require.NoError(t, err)
	asset, err := svc.GetCapabilityAsset(context.Background(), "t1", "cap-aid")
	require.NoError(t, err)
	requireConsistencyOrConflict(t, validateCapabilityAssetProjection("t1", cap, asset))

	listed, err := svc.ListCapabilityAssetsByTenant(context.Background(), "t1")
	require.NoError(t, err)
	byID, err := indexCapabilityAssetsByID(listed)
	require.NoError(t, err)
	requireConsistencyOrConflict(t, validateCapabilityAssetProjection("t1", cap, byID[cap.CapabilityID]))
}

type duplicatingListAssets struct {
	inner *MemoryAssetProjection
}

func (d *duplicatingListAssets) CreateCapabilityAsset(ctx context.Context, asset *assetentity.AssetRef) error {
	return d.inner.CreateCapabilityAsset(ctx, asset)
}
func (d *duplicatingListAssets) UpdateCapabilityProjection(ctx context.Context, tenantID, assetID, name, semanticVersion, contentDigest string, status assetentity.AssetStatus) error {
	return d.inner.UpdateCapabilityProjection(ctx, tenantID, assetID, name, semanticVersion, contentDigest, status)
}
func (d *duplicatingListAssets) GetCapabilityAsset(ctx context.Context, tenantID, assetID string) (*assetentity.AssetRef, error) {
	return d.inner.GetCapabilityAsset(ctx, tenantID, assetID)
}
func (d *duplicatingListAssets) ListCapabilityAssetsByTenant(ctx context.Context, tenantID string) ([]*assetentity.AssetRef, error) {
	rows, err := d.inner.ListCapabilityAssetsByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return rows, nil
	}
	cp := *rows[0]
	return append(append([]*assetentity.AssetRef{}, rows...), &cp), nil
}

var _ AssetProjection = (*duplicatingListAssets)(nil)

type f7DupListUoW struct {
	inner  *MemoryUnitOfWork
	assets AssetProjection
}

func (u *f7DupListUoW) Root() repository.CapabilityRepository { return u.inner.Root() }
func (u *f7DupListUoW) Assets() AssetProjection              { return u.assets }
func (u *f7DupListUoW) WithinTransaction(ctx context.Context, fn func(tx CapabilityTx) error) error {
	return u.inner.WithinTransaction(ctx, fn)
}

var _ CapabilityUnitOfWork = (*f7DupListUoW)(nil)

func TestF7_List_FailClosedDuplicateAssetIDInBatch(t *testing.T) {
	mem := NewMemoryUnitOfWork()
	poison := &duplicatingListAssets{inner: mem.AssetsView()}
	uow := &f7DupListUoW{inner: mem, assets: poison}
	bm := NewFakeBusinessModelPort()
	contracts := NewFakeContractPort()
	svc := NewCapabilityService(&Components{UoW: uow, Business: bm, Contract: contracts})

	payload := fixture.LaboratoryCommandCapability()
	_, _, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", CapabilityID: "cap-dup",
		ActorID: testActor, OwnerID: testOwnerID, Payload: payload,
	})
	require.NoError(t, err)

	listed, err := svc.ListCapabilityAssetsByTenant(context.Background(), "t1")
	require.NoError(t, err)
	_, err = indexCapabilityAssetsByID(listed)
	requireConsistencyOrConflict(t, err)
}

func TestF7_Create_ProjectionDeterministicViaProjectCapabilityAssetRef(t *testing.T) {
	svc, _, _, assets, _, _ := newF7SpyService(nil)
	payload := fixture.LaboratoryCommandCapability()
	cap, rev, err := svc.ManualCreate(context.Background(), &ManualCreateInput{
		TenantID: "t1", BusinessID: "biz-lab", CapabilityID: "cap-proj",
		ActorID: testActor, OwnerID: testOwnerID, Payload: payload,
	})
	require.NoError(t, err)
	want, err := ProjectCapabilityAssetRef(cap, []*entity.BusinessCapabilityRevision{rev})
	require.NoError(t, err)
	got, err := assets.GetCapabilityAsset(context.Background(), "t1", "cap-proj")
	require.NoError(t, err)
	require.Equal(t, want.Name, got.Name)
	require.Equal(t, want.SemanticVersion, got.SemanticVersion)
	require.Equal(t, want.ContentDigest, got.ContentDigest)
	require.Equal(t, want.Status, got.Status)
	require.Equal(t, assetentity.AssetStatusDraft, got.Status)
}

func TestF7_ListProposalsByAnalysisRun_FacadeStableScoped(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryCommandCapability()}}
	svc, repo, _, _, _, _ := newF7SpyService(gen)

	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor,
		BusinessModelRevision: 1, ClientRequestID: "an-f7-1",
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	require.NotNil(t, res.Run)
	runID := res.Run.AnalysisRunID
	require.NotEmpty(t, res.Proposals)

	res2, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor,
		BusinessModelRevision: 1, ClientRequestID: "an-f7-2",
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	require.NotEqual(t, runID, res2.Run.AnalysisRunID)

	listed, err := svc.ListProposalsByAnalysisRun(context.Background(), "t1", runID)
	require.NoError(t, err)
	require.Len(t, listed, len(res.Proposals))
	for _, p := range listed {
		require.Equal(t, "t1", p.TenantID)
		require.Equal(t, "biz-lab", p.BusinessID)
		require.Equal(t, runID, p.AnalysisRunID)
	}
	listed2, err := svc.ListProposalsByAnalysisRun(context.Background(), "t1", runID)
	require.NoError(t, err)
	require.Equal(t, len(listed), len(listed2))
	for i := range listed {
		require.Equal(t, listed[i].ProposalID, listed2[i].ProposalID)
		require.Equal(t, listed[i].Status, listed2[i].Status)
	}

	before, err := repo.GetAnalysisRun(context.Background(), "t1", runID)
	require.NoError(t, err)
	_, err = svc.ListProposalsByAnalysisRun(context.Background(), "t1", runID)
	require.NoError(t, err)
	after, err := repo.GetAnalysisRun(context.Background(), "t1", runID)
	require.NoError(t, err)
	require.Equal(t, before.Status, after.Status)
	require.Equal(t, before.Attempt, after.Attempt)

	otherTenant, err := svc.ListProposalsByAnalysisRun(context.Background(), "t-other", runID)
	require.NoError(t, err)
	require.Empty(t, otherTenant)
}

func TestF7_ListProposalsByAnalysisRun_TerminalStatusesReadable(t *testing.T) {
	gen := &DeterministicFakeGenerator{Proposals: []entity.SemanticPayload{fixture.LaboratoryCommandCapability()}}
	svc, _, _, _, _, _ := newF7SpyService(gen)

	res, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor,
		BusinessModelRevision: 1, ClientRequestID: "an-term-1",
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	propID := res.Proposals[0].ProposalID
	require.Equal(t, entity.ProposalProposed, res.Proposals[0].Status)

	got, err := svc.GetProposal(context.Background(), "t1", propID)
	require.NoError(t, err)
	require.Equal(t, entity.ProposalProposed, got.Status)

	_, err = svc.ConfirmProposal(context.Background(), &ConfirmInput{
		TenantID: "t1", ProposalID: propID, ActorID: testActor, OwnerID: testOwnerID,
		ClientRequestID: "confirm-term", CapabilityID: "cap-from-prop",
	})
	require.NoError(t, err)
	got, err = svc.GetProposal(context.Background(), "t1", propID)
	require.NoError(t, err)
	require.Equal(t, entity.ProposalConfirmed, got.Status)

	listed, err := svc.ListProposalsByAnalysisRun(context.Background(), "t1", res.Run.AnalysisRunID)
	require.NoError(t, err)
	require.NotEmpty(t, listed)
	var found bool
	for _, p := range listed {
		if p.ProposalID == propID {
			found = true
			require.Equal(t, entity.ProposalConfirmed, p.Status)
		}
	}
	require.True(t, found)

	resR, err := svc.StartAnalysis(context.Background(), &StartAnalysisInput{
		TenantID: "t1", BusinessID: "biz-lab", ActorID: testActor,
		BusinessModelRevision: 1, ClientRequestID: "an-term-rej",
		Analysis: entity.AnalysisRequest{BusinessModelRevision: 1},
	})
	require.NoError(t, err)
	rejID := resR.Proposals[0].ProposalID
	_, err = svc.RejectProposal(context.Background(), &RejectInput{
		TenantID: "t1", ProposalID: rejID, ActorID: testActor, ClientRequestID: "rej-1",
	})
	require.NoError(t, err)
	got, err = svc.GetProposal(context.Background(), "t1", rejID)
	require.NoError(t, err)
	require.Equal(t, entity.ProposalRejected, got.Status)
	listedR, err := svc.ListProposalsByAnalysisRun(context.Background(), "t1", resR.Run.AnalysisRunID)
	require.NoError(t, err)
	require.Equal(t, entity.ProposalRejected, listedR[0].Status)
}
