/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma_test

import (
	"go/build"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Ensures Forma domain packages do not import Coze agent or user repository implementations.
func TestFormaDomainDoesNotImportCozeAgentOrUserRepository(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	root := filepath.Dir(filename)
	forbidden := []string{
		"github.com/coze-dev/coze-studio/backend/domain/agent",
		"github.com/coze-dev/coze-studio/backend/domain/user/internal",
		"github.com/coze-dev/coze-studio/backend/domain/user/repository",
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		src := string(content)
		for _, imp := range forbidden {
			require.NotContains(t, src, imp, "forbidden import in %s", path)
		}
		return nil
	})
	require.NoError(t, err)
}

func TestIntegrationUsesCrossDomainNotDomainAgent(t *testing.T) {
	pkg, err := build.Import(
		"github.com/coze-dev/coze-studio/backend/crossdomain/forma/integration",
		".",
		build.ImportComment,
	)
	require.NoError(t, err)
	foundAgentCross := false
	foundUserCross := false
	for _, imp := range pkg.Imports {
		if imp == "github.com/coze-dev/coze-studio/backend/crossdomain/agent" {
			foundAgentCross = true
		}
		if imp == "github.com/coze-dev/coze-studio/backend/crossdomain/user" {
			foundUserCross = true
		}
		require.NotEqual(t, "github.com/coze-dev/coze-studio/backend/domain/agent/singleagent/repository", imp)
		require.NotEqual(t, "github.com/coze-dev/coze-studio/backend/domain/user/internal/dal", imp)
		require.False(t, strings.HasPrefix(imp, "github.com/coze-dev/coze-studio/backend/domain/user/"),
			"integration must not import domain/user packages directly: %s", imp)
	}
	require.True(t, foundAgentCross)
	require.True(t, foundUserCross)
}

func TestTenancyDoesNotImportDomainUser(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	tenancyRoot := filepath.Join(filepath.Dir(filename), "tenancy")
	err := filepath.Walk(tenancyRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		require.NotContains(t, string(content), "github.com/coze-dev/coze-studio/backend/domain/user",
			"tenancy must not import domain/user: %s", path)
		require.NotContains(t, string(content), "github.com/coze-dev/coze-studio/backend/domain/agent",
			"tenancy must not import domain/agent: %s", path)
		return nil
	})
	require.NoError(t, err)
}

func TestDataDomainDoesNotImportCozeRepositoriesOrProviderSDKs(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	dataRoot := filepath.Join(filepath.Dir(filename), "data")
	forbidden := []string{
		"github.com/coze-dev/coze-studio/backend/domain/agent",
		"github.com/coze-dev/coze-studio/backend/domain/user/internal",
		"github.com/coze-dev/coze-studio/backend/domain/user/repository",
		"deepseek",
		"openai",
		"qwen",
		"volcengine/ark",
		"volcengine-go-sdk",
	}
	err := filepath.Walk(dataRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		for _, imp := range forbidden {
			require.NotContains(t, strings.ToLower(string(content)), strings.ToLower(imp),
				"data domain contains forbidden dependency in %s", path)
		}
		return nil
	})
	require.NoError(t, err)
}
