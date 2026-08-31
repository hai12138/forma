/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package integration

import "context"

type AgentDescribeResult struct {
	Available bool   `json:"available"`
	Message   string `json:"message"`
}

type AgentHealthResult struct {
	Healthy bool   `json:"healthy"`
	Message string `json:"message"`
}

// CozeAgentAdapter is the anti-corruption boundary to Coze Agent capabilities.
// Forma domain must depend on this interface, not on Coze agent repositories.
type CozeAgentAdapter interface {
	Describe(ctx context.Context) (*AgentDescribeResult, error)
	Health(ctx context.Context) (*AgentHealthResult, error)
}

// FormaCozeIntegration groups Coze integration adapters for Forma.
type FormaCozeIntegration interface {
	Agent() CozeAgentAdapter
}
