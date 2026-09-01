/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	businessentity "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	businesssvc "github.com/coze-dev/coze-studio/backend/domain/forma/business/service"
	"github.com/coze-dev/coze-studio/backend/domain/forma/analyst/entity"
)

func BuildProposalPatch(assertions []*entity.BusinessAssertion) *entity.SemanticModelPatch {
	patch := &entity.SemanticModelPatch{Operations: []entity.PatchOperation{}}
	seenNodes := map[string]bool{}

	for _, a := range assertions {
		if a == nil || a.Status != entity.AssertionConfirmed {
			continue
		}
		srcIDs := []string{a.AssertionID}

		switch a.AssertionType {
		case entity.AssertionActorExists, entity.AssertionBusinessObjectExists,
			entity.AssertionProcessExists, entity.AssertionEventExists,
			entity.AssertionSystemExists, entity.AssertionPolicyExists:
			nodeID := sanitizeID(a.SubjectRef)
			if nodeID == "item" {
				nodeID = sanitizeID(a.ObjectValue)
			}
			key := string(a.AssertionType) + ":" + nodeID
			if seenNodes[key] {
				continue
			}
			seenNodes[key] = true
			patch.Operations = append(patch.Operations, entity.PatchOperation{
				Op: PatchAddNode,
				Node: &businessentity.SemanticNode{
					ID:           "node_" + nodeID,
					Type:         assertionNodeType(a.AssertionType),
					Name:         a.ObjectValue,
					Description:  a.Predicate,
					SourceMarker: businessentity.SourceAIGenerated,
				},
				SourceAssertionIDs: srcIDs,
			})

		case entity.AssertionRelationExists:
			parts := splitRelation(a.ObjectValue)
			if len(parts) == 2 {
				patch.Operations = append(patch.Operations, entity.PatchOperation{
					Op: PatchAddEdge,
					Edge: &businessentity.SemanticEdge{
						ID:           "edge_" + sanitizeID(parts[0]+"_"+parts[1]),
						Source:       "node_" + sanitizeID(parts[0]),
						Target:       "node_" + sanitizeID(parts[1]),
						Type:         businessentity.EdgeRelatesTo,
						Label:        a.Predicate,
						SourceMarker: businessentity.SourceAIGenerated,
					},
					SourceAssertionIDs: srcIDs,
				})
			}

		case entity.AssertionStateExists:
			objRef := sanitizeID(a.SubjectRef)
			stateID := sanitizeID(a.ObjectValue)
			patch.Operations = append(patch.Operations, entity.PatchOperation{
				Op: PatchAddState,
				State: &businessentity.BusinessState{
					ID:           "state_" + objRef + "_" + stateID,
					ObjectRef:    "node_" + objRef,
					Name:         a.ObjectValue,
					Description:  a.Predicate,
					SourceMarker: businessentity.SourceAIGenerated,
				},
				SourceAssertionIDs: srcIDs,
			})

		case entity.AssertionBusinessRule:
			ruleID := sanitizeID(a.SubjectRef)
			patch.Operations = append(patch.Operations, entity.PatchOperation{
				Op: PatchAddRule,
				Rule: &businessentity.BusinessRule{
					ID:           "rule_" + ruleID,
					Name:         a.ObjectValue,
					Description:  a.Predicate,
					Expression:   a.ObjectValue,
					SourceMarker: businessentity.SourceAIGenerated,
				},
				SourceAssertionIDs: srcIDs,
			})
		}
	}
	return patch
}

func splitRelation(s string) []string {
	for _, sep := range []string{"→", "->", "—", "-", "到"} {
		if parts := splitOnce(s, sep); len(parts) == 2 {
			return parts
		}
	}
	return nil
}

func splitOnce(s, sep string) []string {
	idx := -1
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil
	}
	return []string{strings.TrimSpace(s[:idx]), strings.TrimSpace(s[idx+len(sep):])]}
}

func ProposalDigest(patch *entity.SemanticModelPatch, baseRev int32, assertionIDs []string) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("base:%d", baseRev)))
	sorted := append([]string(nil), assertionIDs...)
	sort.Strings(sorted)
	for _, id := range sorted {
		h.Write([]byte(id))
	}
	if patch != nil {
		b, _ := json.Marshal(patch)
		h.Write(b)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ApplyPatch(model *businessentity.SemanticModel, patch *entity.SemanticModelPatch) (*businessentity.SemanticModel, error) {
	if model == nil {
		model = &businessentity.SemanticModel{SchemaVersion: businessentity.SemanticSchemaVersion}
	}
	cp := cloneSemantic(model)
	if patch == nil {
		return cp, nil
	}

	for _, op := range patch.Operations {
		switch op.Op {
		case PatchAddNode:
			if op.Node != nil {
				cp.Nodes = append(cp.Nodes, *op.Node)
			}
		case PatchUpdateNode:
			for i := range cp.Nodes {
				if cp.Nodes[i].ID == op.TargetID && op.Node != nil {
					cp.Nodes[i] = *op.Node
				}
			}
		case PatchRemoveNode:
			cp.Nodes = filterNodes(cp.Nodes, op.TargetID)
		case PatchAddEdge:
			if op.Edge != nil {
				cp.Edges = append(cp.Edges, *op.Edge)
			}
		case PatchUpdateEdge:
			for i := range cp.Edges {
				if cp.Edges[i].ID == op.TargetID && op.Edge != nil {
					cp.Edges[i] = *op.Edge
				}
			}
		case PatchRemoveEdge:
			cp.Edges = filterEdges(cp.Edges, op.TargetID)
		case PatchAddState:
			if op.State != nil {
				cp.States = append(cp.States, *op.State)
			}
		case PatchUpdateState:
			for i := range cp.States {
				if cp.States[i].ID == op.TargetID && op.State != nil {
					cp.States[i] = *op.State
				}
			}
		case PatchRemoveState:
			cp.States = filterStates(cp.States, op.TargetID)
		case PatchAddRule:
			if op.Rule != nil {
				cp.Rules = append(cp.Rules, *op.Rule)
			}
		case PatchUpdateRule:
			for i := range cp.Rules {
				if cp.Rules[i].ID == op.TargetID && op.Rule != nil {
					cp.Rules[i] = *op.Rule
				}
			}
		case PatchRemoveRule:
			cp.Rules = filterRules(cp.Rules, op.TargetID)
		}
	}

	if err := businesssvc.ValidateSemanticModel(cp); err != nil {
		return nil, fmt.Errorf("%w: %v", entity.ErrProposalInvalid, err)
	}
	return cp, nil
}

func cloneSemantic(m *businessentity.SemanticModel) *businessentity.SemanticModel {
	b, _ := json.Marshal(m)
	var cp businessentity.SemanticModel
	_ = json.Unmarshal(b, &cp)
	return &cp
}

func filterNodes(nodes []businessentity.SemanticNode, id string) []businessentity.SemanticNode {
	out := make([]businessentity.SemanticNode, 0, len(nodes))
	for _, n := range nodes {
		if n.ID != id {
			out = append(out, n)
		}
	}
	return out
}

func filterEdges(edges []businessentity.SemanticEdge, id string) []businessentity.SemanticEdge {
	out := make([]businessentity.SemanticEdge, 0, len(edges))
	for _, e := range edges {
		if e.ID != id {
			out = append(out, e)
		}
	}
	return out
}

func filterStates(states []businessentity.BusinessState, id string) []businessentity.BusinessState {
	out := make([]businessentity.BusinessState, 0, len(states))
	for _, s := range states {
		if s.ID != id {
			out = append(out, s)
		}
	}
	return out
}

func filterRules(rules []businessentity.BusinessRule, id string) []businessentity.BusinessRule {
	out := make([]businessentity.BusinessRule, 0, len(rules))
	for _, r := range rules {
		if r.ID != id {
			out = append(out, r)
		}
	}
	return out
}

// Re-export patch op constants for patch_apply
const (
	PatchAddNode     = entity.PatchAddNode
	PatchUpdateNode  = entity.PatchUpdateNode
	PatchRemoveNode  = entity.PatchRemoveNode
	PatchAddEdge     = entity.PatchAddEdge
	PatchUpdateEdge  = entity.PatchUpdateEdge
	PatchRemoveEdge  = entity.PatchRemoveEdge
	PatchAddState    = entity.PatchAddState
	PatchUpdateState = entity.PatchUpdateState
	PatchRemoveState = entity.PatchRemoveState
	PatchAddRule     = entity.PatchAddRule
	PatchUpdateRule  = entity.PatchUpdateRule
	PatchRemoveRule  = entity.PatchRemoveRule
)
