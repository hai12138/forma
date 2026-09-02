/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestProductionDatasourceDomainIsIndustryAgnostic(t *testing.T) {
	root := filepath.Clean("..")
	forbidden := regexp.MustCompile(`(?i)\bwork[_ -]?order\b|\brepair\s+order\b`)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if strings.Contains(strings.ToLower(filepath.ToSlash(path)), "/testdata") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if match := forbidden.Find(content); match != nil {
			t.Errorf("%s contains industry-specific term %q", path, match)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
