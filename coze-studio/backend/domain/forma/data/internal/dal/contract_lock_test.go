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
)

// TestGetContractForUpdateUsesRowLock proves ContractDAO issues SELECT ... FOR UPDATE.
func TestGetContractForUpdateUsesRowLock(t *testing.T) {
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
		"id", "contract_id", "tenant_id", "business_id", "active_revision_id", "created_by", "created_at", "updated_at",
	}).AddRow(1, "ctr1", "tenant", "lab", "", "actor", now, now)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `forma_data_contract` WHERE tenant_id = ? AND contract_id = ? ORDER BY `forma_data_contract`.`id` LIMIT ? FOR UPDATE",
	)).WithArgs("tenant", "ctr1", 1).WillReturnRows(rows)

	dao := NewContractDAO(gdb)
	got, err := dao.GetContractForUpdate(context.Background(), "tenant", "ctr1")
	require.NoError(t, err)
	require.Equal(t, "ctr1", got.ContractID)
	require.NoError(t, mock.ExpectationsWereMet())
}
