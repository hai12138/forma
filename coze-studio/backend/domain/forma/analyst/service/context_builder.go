/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"fmt"
	"strings"

	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
)

const (
	maxRecentTurns      = 8
	maxEvidenceExcerpts = 6
	maxConfirmedAssert  = 12
	contextBudgetTokens = 4000
)

type ContextInput struct {
	CurrentModel   *businessentity.SemanticModel
	Confirmed      []*entity.BusinessAssertion
	Proposed       []*entity.BusinessAssertion
	OpenConflicts  []*entity.AssertionConflict
	OpenGaps       []*entity.AnalystGap
	RecentTurns    []*entity.AnalystTurn
	EvidenceByTurn map[string]*entity.BusinessEvidence
}

func BuildContext(input *ContextInput) (*entity.ContextManifest, string) {
	if input == nil {
		return &entity.ContextManifest{}, analystSystemPolicy
	}
	var parts []string
	manifest := &entity.ContextManifest{
		IncludedItems: []string{},
		ExcludedItems: []string{},
	}

	parts = append(parts, analystSystemPolicy)

	if input.CurrentModel != nil {
		summary := fmt.Sprintf("Business Model snapshot: %d nodes, %d edges, %d states, %d rules",
			len(input.CurrentModel.Nodes), len(input.CurrentModel.Edges),
			len(input.CurrentModel.States), len(input.CurrentModel.Rules))
		parts = append(parts, summary)
		manifest.IncludedItems = append(manifest.IncludedItems, "business_model_snapshot")
	} else {
		manifest.ExcludedItems = append(manifest.ExcludedItems, "business_model_snapshot")
	}

	for _, c := range input.OpenConflicts {
		if c != nil && c.Status == entity.ConflictOpen {
			parts = append(parts, fmt.Sprintf("CONFLICT: %s %s — assertion %s vs %s",
				c.SubjectRef, c.Predicate, c.AssertionIDA, c.AssertionIDB))
			manifest.IncludedItems = append(manifest.IncludedItems, "conflict:"+c.ConflictID)
		}
	}
	if len(input.OpenConflicts) == 0 {
		manifest.ExcludedItems = append(manifest.ExcludedItems, "conflicts")
	}

	for _, g := range input.OpenGaps {
		if g != nil && g.Status == entity.GapOpen {
			parts = append(parts, fmt.Sprintf("GAP: %s", g.Question))
			manifest.IncludedItems = append(manifest.IncludedItems, "gap:"+g.GapID)
		}
	}
	if len(input.OpenGaps) == 0 {
		manifest.ExcludedItems = append(manifest.ExcludedItems, "gaps")
	}

	turnCount := len(input.RecentTurns)
	start := 0
	if turnCount > maxRecentTurns {
		start = turnCount - maxRecentTurns
		manifest.ExcludedItems = append(manifest.ExcludedItems, fmt.Sprintf("turns:older_%d", start))
	}
	for i := start; i < turnCount; i++ {
		t := input.RecentTurns[i]
		if t == nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("TURN[%s/%s]: %s", t.Speaker, t.TurnID, truncate(t.Content, 200)))
		manifest.IncludedItems = append(manifest.IncludedItems, "turn:"+t.TurnID)
	}

	confCount := 0
	for _, a := range input.Confirmed {
		if a == nil || confCount >= maxConfirmedAssert {
			break
		}
		parts = append(parts, fmt.Sprintf("CONFIRMED: %s %s %s=%s", a.AssertionType, a.SubjectRef, a.Predicate, a.ObjectValue))
		manifest.IncludedItems = append(manifest.IncludedItems, "confirmed:"+a.AssertionID)
		confCount++
	}

	evCount := 0
	for _, t := range input.RecentTurns {
		if t == nil || input.EvidenceByTurn == nil {
			continue
		}
		if ev, ok := input.EvidenceByTurn[t.TurnID]; ok && ev != nil && evCount < maxEvidenceExcerpts {
			parts = append(parts, fmt.Sprintf("EVIDENCE[%s]: %s", ev.EvidenceID, truncate(ev.Quote, 150)))
			manifest.IncludedItems = append(manifest.IncludedItems, "evidence:"+ev.EvidenceID)
			evCount++
		}
	}

	contextText := strings.Join(parts, "\n")
	manifest.TokenEstimate = len(contextText) / 4
	if manifest.TokenEstimate > contextBudgetTokens {
		manifest.ExcludedItems = append(manifest.ExcludedItems, "context_truncated")
		contextText = truncate(contextText, contextBudgetTokens*4)
	}
	return manifest, contextText
}

func PlanNextQuestion(input *ContextInput) *entity.NextQuestionPlan {
	if input == nil {
		return &entity.NextQuestionPlan{
			Question: "请描述您的核心业务流程和参与角色。",
			Goal:     "discover_actors_and_processes",
			Priority: 1,
		}
	}
	for _, g := range input.OpenGaps {
		if g != nil && g.Status == entity.GapOpen {
			return &entity.NextQuestionPlan{
				Question:        g.Question,
				Goal:            "resolve_gap:" + g.GapID,
				RelatedElements: g.RelatedAssertionIDs,
				Priority:        1,
			}
		}
	}
	for _, c := range input.OpenConflicts {
		if c != nil && c.Status == entity.ConflictOpen {
			return &entity.NextQuestionPlan{
				Question:        fmt.Sprintf("关于「%s %s」存在冲突观点，请澄清正确业务规则。", c.SubjectRef, c.Predicate),
				Goal:            "resolve_conflict:" + c.ConflictID,
				RelatedElements: []string{c.AssertionIDA, c.AssertionIDB},
				Priority:        2,
			}
		}
	}
	if input.CurrentModel != nil && len(input.CurrentModel.Nodes) == 0 {
		return &entity.NextQuestionPlan{
			Question: "请介绍业务中的主要角色和业务对象（例如：谁创建什么、处理什么）。",
			Goal:     "discover_core_elements",
			Priority: 3,
		}
	}
	return &entity.NextQuestionPlan{
		Question: "还有哪些业务规则、状态流转或例外情况需要补充？",
		Goal:     "discover_rules_and_states",
		Priority: 4,
	}
}
