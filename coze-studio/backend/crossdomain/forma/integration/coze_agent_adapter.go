/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package integration

import (
	"context"

	crossagent "github.com/coze-dev/coze-studio/backend/crossdomain/agent"
)

type cozeAgentAdapter struct{}

func NewCozeAgentAdapter() CozeAgentAdapter {
	return &cozeAgentAdapter{}
}

func (a *cozeAgentAdapter) Describe(_ context.Context) (*AgentDescribeResult, error) {
	svc := crossagent.DefaultSVC()
	if svc == nil {
		return &AgentDescribeResult{
			Available: false,
			Message:   "coze agent crossdomain service not initialized",
		}, nil
	}
	return &AgentDescribeResult{
		Available: true,
		Message:   "coze agent adapter ready via crossdomain/agent",
	}, nil
}

func (a *cozeAgentAdapter) Health(_ context.Context) (*AgentHealthResult, error) {
	svc := crossagent.DefaultSVC()
	if svc == nil {
		return &AgentHealthResult{
			Healthy: false,
			Message: "coze agent crossdomain service not initialized",
		}, nil
	}
	return &AgentHealthResult{
		Healthy: true,
		Message: "coze agent integration healthy",
	}, nil
}

type formaCozeIntegration struct {
	agent CozeAgentAdapter
}

func NewFormaCozeIntegration(agent CozeAgentAdapter) FormaCozeIntegration {
	if agent == nil {
		agent = NewCozeAgentAdapter()
	}
	return &formaCozeIntegration{agent: agent}
}

func (f *formaCozeIntegration) Agent() CozeAgentAdapter {
	return f.agent
}

var defaultIntegration FormaCozeIntegration

func DefaultIntegration() FormaCozeIntegration {
	if defaultIntegration == nil {
		defaultIntegration = NewFormaCozeIntegration(NewCozeAgentAdapter())
	}
	return defaultIntegration
}

func SetDefaultIntegration(i FormaCozeIntegration) {
	defaultIntegration = i
}
