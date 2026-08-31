/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package forma_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/ut"

	formaRouter "github.com/coze-dev/coze-studio/backend/api/router/forma"
	formaapp "github.com/coze-dev/coze-studio/backend/application/forma"
	"github.com/coze-dev/coze-studio/backend/domain/forma/meta"
)

func TestLiveFormaMetaAPI(t *testing.T) {
	formaapp.ApplicationSVC = &formaapp.ApplicationService{}

	h := server.Default()
	formaRouter.Register(h)

	tests := []struct {
		path    string
		checkFn func(t *testing.T, data map[string]any)
	}{
		{
			path: "/api/forma/v1/health",
			checkFn: func(t *testing.T, data map[string]any) {
				if data["status"] != "ok" {
					t.Fatalf("expected status ok, got %v", data["status"])
				}
			},
		},
		{
			path: "/api/forma/v1/version",
			checkFn: func(t *testing.T, data map[string]any) {
				if data["forma_version"] != meta.FormaVersion {
					t.Fatalf("unexpected forma_version: %v", data["forma_version"])
				}
				if data["forma_schema_version"] != meta.FormaSchemaVersion {
					t.Fatalf("unexpected schema version: %v", data["forma_schema_version"])
				}
			},
		},
		{
			path: "/api/forma/v1/meta/baseline",
			checkFn: func(t *testing.T, data map[string]any) {
				if data["forma_baseline_tag"] != meta.FormaBaselineTag {
					t.Fatalf("unexpected baseline tag: %v", data["forma_baseline_tag"])
				}
				if data["coze_baseline_commit"] != meta.CozeBaselineCommit {
					t.Fatalf("unexpected coze baseline: %v", data["coze_baseline_commit"])
				}
				if data["runtime_foundation"] != "eino" {
					t.Fatalf("expected eino runtime, got %v", data["runtime_foundation"])
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			w := ut.PerformRequest(
				h.Engine,
				"GET",
				tc.path,
				nil,
				ut.Header{Key: "X-Request-ID", Value: "forma-g1-smoke"},
			)

			if w.Code != 200 {
				t.Fatalf("status %d body %s", w.Code, w.Body.String())
			}

			var body struct {
				Code      int32          `json:"code"`
				Msg       string         `json:"msg"`
				RequestID string         `json:"request_id"`
				Data      map[string]any `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Code != 0 || body.Msg != "ok" {
				t.Fatalf("unexpected envelope: %+v", body)
			}
			if body.RequestID == "" {
				t.Fatal("expected request_id in response")
			}
			tc.checkFn(t, body.Data)
		})
	}
}

func TestFormaMetaAPILiveTCP(t *testing.T) {
	formaapp.ApplicationSVC = &formaapp.ApplicationService{}
	h := server.Default(server.WithHostPorts("127.0.0.1:18888"))
	formaRouter.Register(h)

	go func() {
		if err := h.Run(); err != nil {
			t.Logf("hertz run ended: %v", err)
		}
	}()
	defer func() {
		_ = h.Shutdown(context.Background())
	}()

	waitForLive := func() error {
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			resp, err := http.Get("http://127.0.0.1:18888/api/forma/v1/health")
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return nil
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
		return http.ErrHandlerTimeout
	}
	if err := waitForLive(); err != nil {
		t.Fatalf("server did not become ready: %v", err)
	}

	resp, err := http.Get("http://127.0.0.1:18888/api/forma/v1/meta/baseline")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	var body struct {
		Code int32 `json:"code"`
		Data struct {
			FormaVersion       string `json:"forma_version"`
			CozeBaselineCommit string `json:"coze_baseline_commit"`
			RuntimeFoundation  string `json:"runtime_foundation"`
			FormaSchemaVersion string `json:"forma_schema_version"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Data.RuntimeFoundation != "eino" {
		t.Fatalf("expected eino, got %s", body.Data.RuntimeFoundation)
	}
	if body.Data.CozeBaselineCommit != meta.CozeBaselineCommit {
		t.Fatalf("unexpected baseline commit")
	}
	if body.Data.FormaSchemaVersion != meta.FormaSchemaVersion {
		t.Fatalf("unexpected schema version")
	}
}
