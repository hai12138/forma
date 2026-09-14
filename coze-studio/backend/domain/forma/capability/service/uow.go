/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"errors"

	assetentity "github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/repository"
	"gorm.io/gorm"
)

// CapabilityUnitOfWork coordinates Cap DAO + AssetProjection atomicity.
type CapabilityUnitOfWork interface {
	WithinTransaction(ctx context.Context, fn func(tx CapabilityTx) error) error
	Root() repository.CapabilityRepository
	// Assets returns the non-transactional root AssetProjection view.
	Assets() AssetProjection
}

// CapabilityTx is the transactional view of Cap repo + assets sharing one logical txn.
type CapabilityTx interface {
	Repo() repository.CapabilityRepository
	Assets() AssetProjection
}

type capabilityTx struct {
	repo   repository.CapabilityRepository
	assets AssetProjection
}

func (t *capabilityTx) Repo() repository.CapabilityRepository { return t.repo }
func (t *capabilityTx) Assets() AssetProjection               { return t.assets }

// gormUoW shares one *gorm.DB txn for Cap DAO + GormAssetProjection.
type gormUoW struct {
	db   *gorm.DB
	root repository.CapabilityRepository
}

// NewGormUnitOfWork binds Capability repo + AssetRef projection onto one GORM transaction.
func NewGormUnitOfWork(db *gorm.DB) CapabilityUnitOfWork {
	return &gormUoW{db: db, root: repository.NewCapabilityRepository(db)}
}

func (u *gormUoW) Root() repository.CapabilityRepository { return u.root }

func (u *gormUoW) Assets() AssetProjection { return NewGormAssetProjection(u.db) }

func (u *gormUoW) WithinTransaction(ctx context.Context, fn func(tx CapabilityTx) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&capabilityTx{
			repo:   repository.NewCapabilityRepository(tx),
			assets: NewGormAssetProjection(tx),
		})
	})
}

// MemoryUoWOptions injects failures for tests without losing ownership of repo/assets.
type MemoryUoWOptions struct {
	FailAssetCreate      bool
	FailAssetUpdate      bool
	FailCreateValidation bool
	FailCommit           bool // after successful callback, before commit — restore both
	OnCommitFail         func()
	OnAssetFail          func()
	// OnBeforeListRevisions runs before each ListRevisions in the UoW txn (race/fence tests).
	OnBeforeListRevisions func()
}

// MemoryUnitOfWork owns both the in-memory Cap repo and AssetProjection.
type MemoryUnitOfWork struct {
	repo      repository.CapabilityRepository
	memAssets *MemoryAssetProjection
	opts      MemoryUoWOptions
}

// NewMemoryUnitOfWork returns a UoW that owns memRepo + memAssets.
func NewMemoryUnitOfWork() *MemoryUnitOfWork {
	return NewMemoryUnitOfWorkWithOptions(MemoryUoWOptions{})
}

// NewMemoryUnitOfWorkWithOptions returns a UoW with optional failure injection seams.
func NewMemoryUnitOfWorkWithOptions(opts MemoryUoWOptions) *MemoryUnitOfWork {
	return &MemoryUnitOfWork{
		repo:      repository.NewMemoryCapabilityRepository(),
		memAssets: NewMemoryAssetProjection(),
		opts:      opts,
	}
}

func (u *MemoryUnitOfWork) Root() repository.CapabilityRepository { return u.repo }

func (u *MemoryUnitOfWork) Assets() AssetProjection { return u.memAssets }

// AssetsView exposes the owned asset projection for test inspection.
func (u *MemoryUnitOfWork) AssetsView() *MemoryAssetProjection { return u.memAssets }

// SetFailAssetUpdate toggles asset update failure injection (tests).
func (u *MemoryUnitOfWork) SetFailAssetUpdate(v bool) {
	u.opts.FailAssetUpdate = v
}

// SetFailCreateValidation toggles CreateValidationResult failure injection (tests).
func (u *MemoryUnitOfWork) SetFailCreateValidation(v bool) {
	u.opts.FailCreateValidation = v
}

// SetOnBeforeListRevisions installs a hook invoked before each ListRevisions in UoW txns (tests).
func (u *MemoryUnitOfWork) SetOnBeforeListRevisions(fn func()) {
	u.opts.OnBeforeListRevisions = fn
}

func (u *MemoryUnitOfWork) WithinTransaction(ctx context.Context, fn func(tx CapabilityTx) error) error {
	// Snapshot assets inside repo.Transaction so concurrent losers restore to a snap
	// that already includes winners' commits (repo lock serializes Memory UoW).
	return u.repo.Transaction(ctx, func(txRepo repository.CapabilityRepository) error {
		assetSnap := u.memAssets.snapshot()
		assets := AssetProjection(u.memAssets)
		if u.opts.FailAssetCreate || u.opts.FailAssetUpdate {
			assets = &txnFailingAssets{
				inner:      u.memAssets,
				failCreate: u.opts.FailAssetCreate,
				failUpdate: u.opts.FailAssetUpdate,
				onFail:     u.opts.OnAssetFail,
			}
		}
		repo := txRepo
		if u.opts.FailCreateValidation {
			repo = &failCreateValidationRepo{CapabilityRepository: repo}
		}
		if u.opts.OnBeforeListRevisions != nil {
			repo = &listRevisionsHookRepo{CapabilityRepository: repo, before: u.opts.OnBeforeListRevisions}
		}
		err := fn(&capabilityTx{repo: repo, assets: assets})
		if err != nil {
			u.memAssets.restore(assetSnap)
			return err
		}
		if u.opts.FailCommit {
			u.memAssets.restore(assetSnap)
			if u.opts.OnCommitFail != nil {
				u.opts.OnCommitFail()
			}
			return entity.ErrUoWCommitFailed
		}
		return nil
	})
}

// failCreateValidationRepo injects CreateValidationResult failures for UoW tests.
type failCreateValidationRepo struct {
	repository.CapabilityRepository
}

func (r *failCreateValidationRepo) CreateValidationResult(ctx context.Context, v *entity.CapabilityValidationResult) error {
	return entity.ErrConsistency
}

// listRevisionsHookRepo runs a test hook before ListRevisions (Activate fence race).
type listRevisionsHookRepo struct {
	repository.CapabilityRepository
	before func()
}

func (r *listRevisionsHookRepo) ListRevisions(ctx context.Context, tenantID, capabilityID string) ([]*entity.BusinessCapabilityRevision, error) {
	if r.before != nil {
		r.before()
	}
	return r.CapabilityRepository.ListRevisions(ctx, tenantID, capabilityID)
}

// txnFailingAssets wraps owned MemoryAssetProjection for a single transaction only.
type txnFailingAssets struct {
	inner      *MemoryAssetProjection
	failCreate bool
	failUpdate bool
	onFail     func()
}

func (f *txnFailingAssets) CreateCapabilityAsset(ctx context.Context, asset *assetentity.AssetRef) error {
	if f.failCreate {
		if f.onFail != nil {
			f.onFail()
		}
		return errors.New("injected asset create failure")
	}
	return f.inner.CreateCapabilityAsset(ctx, asset)
}

func (f *txnFailingAssets) UpdateCapabilityProjection(ctx context.Context, tenantID, assetID, name, semanticVersion, contentDigest string, status assetentity.AssetStatus) error {
	if f.failUpdate {
		if f.onFail != nil {
			f.onFail()
		}
		return errors.New("injected asset update failure")
	}
	return f.inner.UpdateCapabilityProjection(ctx, tenantID, assetID, name, semanticVersion, contentDigest, status)
}

func (f *txnFailingAssets) GetCapabilityAsset(ctx context.Context, tenantID, assetID string) (*assetentity.AssetRef, error) {
	return f.inner.GetCapabilityAsset(ctx, tenantID, assetID)
}

func (f *txnFailingAssets) ListCapabilityAssetsByTenant(ctx context.Context, tenantID string) ([]*assetentity.AssetRef, error) {
	return f.inner.ListCapabilityAssetsByTenant(ctx, tenantID)
}

var _ CapabilityUnitOfWork = (*MemoryUnitOfWork)(nil)
var _ AssetProjection = (*txnFailingAssets)(nil)
