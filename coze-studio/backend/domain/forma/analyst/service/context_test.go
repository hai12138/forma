/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"fmt"
	"strings"
	"testing"

	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	"github.com/stretchr/testify/require"
)

func TestBuildContextExcludesSystemPolicyFromText(t *testing.T) {
	input := &ContextInput{
		CurrentModel: &businessentity.SemanticModel{
			Nodes: []businessentity.SemanticNode{{ID: "n1"}},
		},
		OpenGaps: []*entity.AnalystGap{{
			GapID:    "gap1",
			Question: "谁可以关闭工单？",
			Status:   entity.GapOpen,
		}},
	}
	manifest, text := BuildContext(input)
	require.NotContains(t, text, "Never auto-confirm")
	require.Contains(t, text, "Business Model snapshot")
	require.Contains(t, text, "谁可以关闭工单")
	require.Contains(t, manifest.IncludedItems, "gap:gap1")
}

func TestBuildContextBudgetTruncation(t *testing.T) {
	input := &ContextInput{
		Confirmed: make([]*entity.BusinessAssertion, 0, maxConfirmedAssert),
	}
	for i := 0; i < maxConfirmedAssert; i++ {
		input.Confirmed = append(input.Confirmed, &entity.BusinessAssertion{
			AssertionID:   fmt.Sprintf("a%d", i),
			AssertionType: entity.AssertionProperty,
			SubjectRef:    "subject",
			Predicate:     "predicate",
			ObjectValue:   strings.Repeat("确认事实", 400),
		})
	}
	manifest, text := BuildContext(input)
	require.LessOrEqual(t, manifest.TokenEstimate, contextBudgetTokens)
	require.Contains(t, manifest.ExcludedItems, "context_truncated")
	require.LessOrEqual(t, len(text), contextBudgetTokens*4+64)
}

func TestPlanNextQuestionFocusedGapPriority(t *testing.T) {
	input := &ContextInput{
		FocusGapID: "gap_focus",
		OpenGaps: []*entity.AnalystGap{
			{GapID: "gap_other", Question: "其他问题", Status: entity.GapOpen},
			{GapID: "gap_focus", Question: "聚焦问题", Status: entity.GapOpen},
		},
	}
	plan := PlanNextQuestion(input)
	require.Equal(t, "聚焦问题", plan.Question)
	require.Equal(t, "resolve_gap:gap_focus", plan.Goal)
	require.Equal(t, 0, plan.Priority)
}

func TestFormatExtractionUserPromptTags(t *testing.T) {
	out := FormatExtractionUserPrompt("Business Model snapshot: 1 nodes", "turn_1", "管理员可以关闭")
	require.Contains(t, out, "<context>")
	require.Contains(t, out, "Business Model snapshot")
	require.Contains(t, out, "<current_user_turn turn_id=\"turn_1\">")
	require.Contains(t, out, "管理员可以关闭")
}
