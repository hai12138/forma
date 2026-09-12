/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"context"
	"sync"

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
)

// ProposalGenerator produces AI proposals for an analysis run. No real model SDKs.
type ProposalGenerator interface {
	Generate(ctx context.Context, req GenerateRequest) (*GenerateResult, error)
}

// GenerateRequest is the domain input for proposal generation.
type GenerateRequest struct {
	TenantID              string
	BusinessID            string
	BusinessModelRevision int32
	AnalysisRunID         string
	Analysis              entity.AnalysisRequest
}

// GenerateResult is deterministic fake / adapter output.
type GenerateResult struct {
	ModelRef  string
	Proposals []entity.SemanticPayload
}

// DeterministicFakeGenerator is a test double with call counting (REAL_MODEL_CALLS=0).
type DeterministicFakeGenerator struct {
	mu        sync.Mutex
	Calls     int
	ModelRef  string
	Proposals []entity.SemanticPayload
	Err       error
	GenerateFn func(context.Context, GenerateRequest) (*GenerateResult, error)
}

func (g *DeterministicFakeGenerator) Generate(ctx context.Context, req GenerateRequest) (*GenerateResult, error) {
	g.mu.Lock()
	g.Calls++
	fn := g.GenerateFn
	err := g.Err
	ref := g.ModelRef
	props := append([]entity.SemanticPayload(nil), g.Proposals...)
	g.mu.Unlock()
	if fn != nil {
		return fn(ctx, req)
	}
	if err != nil {
		return nil, err
	}
	if ref == "" {
		ref = "fake-model"
	}
	return &GenerateResult{ModelRef: ref, Proposals: props}, nil
}

func (g *DeterministicFakeGenerator) CallCount() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.Calls
}
