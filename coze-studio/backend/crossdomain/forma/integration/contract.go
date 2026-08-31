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

type SpaceDescribeResult struct {
	SpaceID     int64  `json:"space_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     int64  `json:"owner_id"`
	Available   bool   `json:"available"`
	Message     string `json:"message"`
}

// CozeSpaceAdapter is the anti-corruption boundary to Coze Space ACL.
// Implementations must use crossdomain/user — never domain/user/internal/dal.
type CozeSpaceAdapter interface {
	DescribeSpace(ctx context.Context, spaceID int64) (*SpaceDescribeResult, error)
	ValidateSpaceAccess(ctx context.Context, cozeUserID, spaceID int64) error
}

// FormaCozeIntegration groups Coze integration adapters for Forma.
type FormaCozeIntegration interface {
	Agent() CozeAgentAdapter
	Space() CozeSpaceAdapter
}
