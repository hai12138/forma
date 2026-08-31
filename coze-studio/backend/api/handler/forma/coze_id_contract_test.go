/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	"github.com/coze-dev/coze-studio/backend/domain/forma/idcontract"
	tenancyentity "github.com/coze-dev/coze-studio/backend/domain/forma/tenancy/entity"
)

func TestBindSpaceInput_RejectsJSONNumber(t *testing.T) {
	var in formaapp.BindSpaceInput
	err := json.Unmarshal([]byte(`{"coze_space_id":7563957783431741441,"purpose":"DEFAULT"}`), &in)
	require.Error(t, err)
}

func TestBindSpaceInput_AcceptsStringSnowflake(t *testing.T) {
	var in formaapp.BindSpaceInput
	err := json.Unmarshal([]byte(`{"coze_space_id":"7563957783431741441","purpose":"DEFAULT"}`), &in)
	require.NoError(t, err)
	id, err := idcontract.ParseCozeID(in.CozeSpaceID)
	require.NoError(t, err)
	assert.Equal(t, int64(7563957783431741441), id)
	assert.Equal(t, tenancyentity.SpacePurposeDefault, in.Purpose)
}

func TestBindSpaceInput_RejectStrings(t *testing.T) {
	bad := []string{`"1.23"`, `"1e18"`, `"-1"`, `"abc"`, `"9223372036854775808"`}
	for _, raw := range bad {
		var in formaapp.BindSpaceInput
		payload := `{"coze_space_id":` + raw + `,"purpose":"DEFAULT"}`
		require.NoError(t, json.Unmarshal([]byte(payload), &in))
		_, err := idcontract.ParseCozeID(in.CozeSpaceID)
		assert.Error(t, err, "payload=%s", payload)
	}
}

func TestTenantSpaceDTO_JSONIsString(t *testing.T) {
	dto := &formaapp.TenantSpaceDTO{
		TenantID:    "ten_x",
		CozeSpaceID: idcontract.FormatCozeID(9007199254740993),
		Purpose:     "DEFAULT",
		Status:      "ACTIVE",
	}
	b, err := json.Marshal(dto)
	require.NoError(t, err)
	assert.Contains(t, string(b), `"coze_space_id":"9007199254740993"`)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	_, isString := m["coze_space_id"].(string)
	assert.True(t, isString)
}

func TestPrincipalDTO_CozeUserIDString(t *testing.T) {
	dto := &formaapp.PrincipalDTO{CozeUserID: idcontract.FormatCozeID(7563957783431741441)}
	b, err := json.Marshal(dto)
	require.NoError(t, err)
	assert.Contains(t, string(b), `"coze_user_id":"7563957783431741441"`)
}

func TestBootstrapInput_OptionalEmpty(t *testing.T) {
	var in formaapp.BootstrapInput
	require.NoError(t, json.Unmarshal([]byte(`{}`), &in))
	id, err := idcontract.ParseOptionalCozeID(in.DefaultSpaceID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), id)
}
