/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package dal

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
)

func TestGetCapabilityForUpdateUsesRowLock(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	gdb, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)

	now := time.Now().UTC()
	rows := sqlmock.NewRows([]string{
		"id", "capability_id", "tenant_id", "business_id", "active_revision_id", "aggregate_generation", "created_by", "created_at", "updated_at",
	}).AddRow(1, "cap1", "tenant", "biz", nil, 0, "actor", now, now)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `forma_business_capability` WHERE tenant_id = ? AND capability_id = ? ORDER BY `forma_business_capability`.`id` LIMIT ? FOR UPDATE",
	)).WithArgs("tenant", "cap1", 1).WillReturnRows(rows)

	dao := NewCapabilityDAO(gdb)
	got, err := dao.GetCapabilityForUpdate(context.Background(), "tenant", "cap1")
	require.NoError(t, err)
	require.Equal(t, "cap1", got.CapabilityID)
	require.Equal(t, int64(0), got.AggregateGeneration)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProposalForUpdateUsesRowLock(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	gdb, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)

	now := time.Now().UTC()
	rows := sqlmock.NewRows([]string{
		"id", "proposal_id", "tenant_id", "business_id", "analysis_run_id", "capability_id", "status", "payload_json", "materialized_revision_id", "created_at",
	}).AddRow(1, "p1", "tenant", "biz", "run", nil, "PROPOSED", `{"name":"n","capability_kind":"QUERY","business_model_revision":1}`, nil, now)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `forma_capability_proposal` WHERE tenant_id = ? AND proposal_id = ? ORDER BY `forma_capability_proposal`.`id` LIMIT ? FOR UPDATE",
	)).WithArgs("tenant", "p1", 1).WillReturnRows(rows)

	dao := NewCapabilityDAO(gdb)
	got, err := dao.GetProposalForUpdate(context.Background(), "tenant", "p1")
	require.NoError(t, err)
	require.Equal(t, "p1", got.ProposalID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProposalCorruptJSONFailClosed(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	gdb, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)

	now := time.Now().UTC()
	rows := sqlmock.NewRows([]string{
		"id", "proposal_id", "tenant_id", "business_id", "analysis_run_id", "capability_id", "status", "payload_json", "materialized_revision_id", "created_at",
	}).AddRow(1, "p1", "tenant", "biz", "run", nil, "PROPOSED", `{not-json`, nil, now)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `forma_capability_proposal` WHERE tenant_id = ? AND proposal_id = ? ORDER BY `forma_capability_proposal`.`id` LIMIT ?",
	)).WithArgs("tenant", "p1", 1).WillReturnRows(rows)

	dao := NewCapabilityDAO(gdb)
	_, err = dao.GetProposal(context.Background(), "tenant", "p1")
	require.ErrorIs(t, err, entity.ErrConsistency)
	require.NoError(t, mock.ExpectationsWereMet())
}
