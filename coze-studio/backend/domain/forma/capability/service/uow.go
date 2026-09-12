/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/repository"
	"gorm.io/gorm"
)

// CapabilityUnitOfWork coordinates Cap DAO + AssetProjection atomicity.
type CapabilityUnitOfWork interface {
	WithinTransaction(ctx context.Context, fn func(tx CapabilityTx) error) error
	// Root returns a repository bound to the same store (auto-commit / single ops outside explicit txn).
	Root() repository.CapabilityRepository
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

func (u *gormUoW) WithinTransaction(ctx context.Context, fn func(tx CapabilityTx) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&capabilityTx{
			repo:   repository.NewCapabilityRepository(tx),
			assets: NewGormAssetProjection(tx),
		})
	})
}

// memoryUoW snap/restores BOTH repo and assets on failure; commits only on success.
type memoryUoW struct {
	repo      repository.CapabilityRepository
	memAssets *MemoryAssetProjection
	assets    AssetProjection
}

// NewMemoryUnitOfWork returns an in-memory UoW and the asset projection for inspection.
func NewMemoryUnitOfWork() (CapabilityUnitOfWork, *MemoryAssetProjection) {
	repo := repository.NewMemoryCapabilityRepository()
	assets := NewMemoryAssetProjection()
	return &memoryUoW{repo: repo, memAssets: assets, assets: assets}, assets
}

// NewMemoryUnitOfWorkWith wires a custom root (e.g. test wrappers) and optional asset view
// over the same MemoryAssetProjection used for snap/restore.
// Root MUST be the same backing store as used by WithinTransaction (documented contract).
func NewMemoryUnitOfWorkWith(repo repository.CapabilityRepository, memAssets *MemoryAssetProjection, assets AssetProjection) CapabilityUnitOfWork {
	if assets == nil {
		assets = memAssets
	}
	return &memoryUoW{repo: repo, memAssets: memAssets, assets: assets}
}

func (u *memoryUoW) Root() repository.CapabilityRepository { return u.repo }

func (u *memoryUoW) WithinTransaction(ctx context.Context, fn func(tx CapabilityTx) error) error {
	// Snapshot assets inside repo.Transaction so concurrent losers restore to a snap
	// that already includes winners' commits (repo lock serializes Memory UoW).
	return u.repo.Transaction(ctx, func(txRepo repository.CapabilityRepository) error {
		assetSnap := u.memAssets.snapshot()
		err := fn(&capabilityTx{repo: txRepo, assets: u.assets})
		if err != nil {
			u.memAssets.restore(assetSnap)
		}
		return err
	})
}
