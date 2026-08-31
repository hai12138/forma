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

// Ensures Forma domain packages do not import Coze agent repository implementations.
func TestFormaDomainDoesNotImportCozeAgentRepository(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	root := filepath.Dir(filename)
	forbidden := []string{
		"github.com/coze-dev/coze-studio/backend/domain/agent",
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
	foundCross := false
	for _, imp := range pkg.Imports {
		if imp == "github.com/coze-dev/coze-studio/backend/crossdomain/agent" {
			foundCross = true
		}
		require.NotEqual(t, "github.com/coze-dev/coze-studio/backend/domain/agent/singleagent/repository", imp)
	}
	require.True(t, foundCross)
}
