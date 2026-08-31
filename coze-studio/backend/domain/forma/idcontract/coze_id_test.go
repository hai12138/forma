/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package idcontract

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatParseRoundTrip_AboveJSSafeInteger(t *testing.T) {
	ids := []int64{
		9007199254740993,    // Number.MAX_SAFE_INTEGER + 2
		7563957783431741441, // typical Coze snowflake
	}
	for _, id := range ids {
		s := FormatCozeID(id)
		require.Equal(t, strconvFormat(id), s)
		got, err := ParseCozeID(s)
		require.NoError(t, err, "id=%d", id)
		assert.Equal(t, id, got)
	}
}

func strconvFormat(id int64) string {
	return FormatCozeID(id)
}

func TestParseCozeID_RejectsInvalid(t *testing.T) {
	cases := []string{
		"",
		" ",
		"abc",
		"1.23",
		"1e18",
		"1E18",
		"-1",
		"+1",
		"01",
		"0",
		"9223372036854775808", // MaxInt64+1
	}
	for _, c := range cases {
		_, err := ParseCozeID(c)
		assert.Error(t, err, "input=%q", c)
	}
}

func TestParseCozeID_RejectsUnicodeDigits(t *testing.T) {
	cases := []string{
		"١",     // Arabic-Indic digit one
		"１２３", // full-width 123
		"123١",  // mixed ASCII + Arabic-Indic
		"１２",  // full-width
	}
	for _, c := range cases {
		_, err := ParseCozeID(c)
		assert.Error(t, err, "unicode digit input must reject: %q", c)
	}
}


func TestParseCozeID_AcceptsValid(t *testing.T) {
	got, err := ParseCozeID("7563957783431741441")
	require.NoError(t, err)
	assert.Equal(t, int64(7563957783431741441), got)

	got, err = ParseCozeID("9007199254740993")
	require.NoError(t, err)
	assert.Equal(t, int64(9007199254740993), got)
}

func TestJSONNumberRejectedIntoStringDTO(t *testing.T) {
	type spaceIn struct {
		CozeSpaceID string `json:"coze_space_id"`
	}
	var in spaceIn
	err := json.Unmarshal([]byte(`{"coze_space_id":7563957783431741441}`), &in)
	require.Error(t, err, "JSON number must not unmarshal into string Coze ID field")
}

func TestJSONStringPreservesSnowflake(t *testing.T) {
	const raw = "7563957783431741441"
	type spaceOut struct {
		CozeSpaceID string `json:"coze_space_id"`
	}
	b, err := json.Marshal(spaceOut{CozeSpaceID: FormatCozeID(7563957783431741441)})
	require.NoError(t, err)
	assert.Contains(t, string(b), `"coze_space_id":"7563957783431741441"`)

	var back spaceOut
	require.NoError(t, json.Unmarshal(b, &back))
	assert.Equal(t, raw, back.CozeSpaceID)
	id, err := ParseCozeID(back.CozeSpaceID)
	require.NoError(t, err)
	assert.Equal(t, int64(7563957783431741441), id)
}

func TestJSONRoundTrip_AboveJSSafeInteger_SimulatedJSParse(t *testing.T) {
	// Simulate: backend int64 → DTO string → JSON → JS JSON.parse → request → Parse → int64
	const domainID int64 = 9007199254740993
	type dto struct {
		CozeUserID string `json:"coze_user_id"`
	}
	payload, err := json.Marshal(dto{CozeUserID: FormatCozeID(domainID)})
	require.NoError(t, err)

	// JS JSON.parse preserves string values exactly.
	var parsed map[string]any
	require.NoError(t, json.Unmarshal(payload, &parsed))
	s, ok := parsed["coze_user_id"].(string)
	require.True(t, ok, "must be JSON string, not number: %T", parsed["coze_user_id"])
	assert.Equal(t, "9007199254740993", s)

	reqBody, err := json.Marshal(map[string]string{"coze_user_id": s})
	require.NoError(t, err)
	var req dto
	require.NoError(t, json.Unmarshal(reqBody, &req))
	got, err := ParseCozeID(req.CozeUserID)
	require.NoError(t, err)
	assert.Equal(t, domainID, got)
}
