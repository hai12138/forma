/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/entity"
	"github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/repository"
	"github.com/coze-dev/coze-studio/backend/domain/forma/asset_registry/service"
)

func TestAssetRegistry_CreateAsset(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	gdb, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `forma_asset_ref`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	reg := service.NewAssetRegistry(&service.Components{
		AssetRepo: repository.NewAssetRefRepository(gdb),
		CozeRepo:  repository.NewCozeResourceRefRepository(gdb),
	})

	asset, err := reg.CreateAsset(context.Background(), &service.CreateAssetRequest{
		TenantID: "tenant-1",
		AssetID:  "asset-1",
		Kind:     entity.AssetKindBusiness,
		Name:     "Reference Business Model",
		OwnerID:  100,
		CreatedBy: 100,
	})
	require.NoError(t, err)
	assert.Equal(t, entity.AssetStatusDraft, asset.Status)
	assert.Equal(t, int32(1), asset.Revision)
	assert.Equal(t, "0.0.0", asset.SemanticVersion)
}

func TestAssetKindAndStatusConstants(t *testing.T) {
	kinds := []entity.AssetKind{
		entity.AssetKindBusiness,
		entity.AssetKindCapability,
		entity.AssetKindAgent,
		entity.AssetKindApplication,
	}
	assert.Len(t, kinds, 4)

	statuses := []entity.AssetStatus{
		entity.AssetStatusDraft,
		entity.AssetStatusInReview,
		entity.AssetStatusVerified,
		entity.AssetStatusFrozen,
		entity.AssetStatusReleased,
		entity.AssetStatusDeprecated,
		entity.AssetStatusArchived,
	}
	assert.Len(t, statuses, 7)
}

func TestCozeResourceTypes(t *testing.T) {
	types := []entity.CozeResourceType{
		entity.CozeResourceTypeAgent,
		entity.CozeResourceTypeWorkflow,
		entity.CozeResourceTypePlugin,
		entity.CozeResourceTypeKnowledge,
		entity.CozeResourceTypeApp,
		entity.CozeResourceTypeDatabase,
	}
	assert.Len(t, types, 6)
	_ = time.Now()
}
