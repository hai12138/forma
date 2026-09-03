/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
)

func TestNormalizeAccount_LocalAlias(t *testing.T) {
	n, err := formaapp.NormalizeAccount("admin")
	require.NoError(t, err)
	assert.Equal(t, "admin", n.Account)
	assert.Equal(t, "admin@forma.local", n.Email)
	assert.True(t, n.IsLocalAlias)

	n, err = formaapp.NormalizeAccount("User01")
	require.NoError(t, err)
	assert.Equal(t, "user01", n.Account)
	assert.Equal(t, "user01@forma.local", n.Email)
}

func TestNormalizeAccount_LocalEmailCollapsesToAlias(t *testing.T) {
	a, err := formaapp.NormalizeAccount("user01")
	require.NoError(t, err)
	b, err := formaapp.NormalizeAccount("user01@forma.local")
	require.NoError(t, err)
	assert.Equal(t, a.Email, b.Email)
	assert.Equal(t, a.Account, b.Account)
	assert.True(t, a.IsLocalAlias)
	assert.True(t, b.IsLocalAlias)
}

func TestNormalizeAccount_ExternalEmail(t *testing.T) {
	n, err := formaapp.NormalizeAccount("Email@Example.com")
	require.NoError(t, err)
	assert.Equal(t, "email@example.com", n.Account)
	assert.Equal(t, "email@example.com", n.Email)
	assert.False(t, n.IsLocalAlias)
}

func TestNormalizeAccount_RejectsInvalid(t *testing.T) {
	cases := []string{
		"",
		" ",
		"user name",
		"user/01",
		"user\\01",
		"@forma.local",
		"a@b@c.com",
		"bad!",
	}
	for _, c := range cases {
		_, err := formaapp.NormalizeAccount(c)
		require.Error(t, err, "expected error for %q", c)
	}
}

func TestAccountFromEmail(t *testing.T) {
	assert.Equal(t, "admin", formaapp.AccountFromEmail("admin@forma.local"))
	assert.Equal(t, "user@example.com", formaapp.AccountFromEmail("user@example.com"))
}

func TestValidatePasswordPolicy(t *testing.T) {
	require.Error(t, formaapp.ValidatePasswordForTest("short"))
	require.Error(t, formaapp.ValidatePasswordForTest("admin123"))
	require.NoError(t, formaapp.ValidatePasswordForTest("Admin123456!"))
}
