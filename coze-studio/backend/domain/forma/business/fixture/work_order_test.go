/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package fixture_test

import (
	"testing"

	"github.com/coze-dev/coze-studio/backend/domain/forma/business/fixture"
	"github.com/coze-dev/coze-studio/backend/domain/forma/business/service"
	"github.com/stretchr/testify/require"
)

func TestWorkOrderFixtureValidates(t *testing.T) {
	m := fixture.WorkOrderRepairSemanticModel()
	require.NoError(t, service.ValidateSemanticModel(m))
	d1, _, err := service.ContentDigest(m)
	require.NoError(t, err)
	d2, _, err := service.ContentDigest(m)
	require.NoError(t, err)
	require.Equal(t, d1, d2)
}
