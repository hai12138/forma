/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"fmt"
	"strings"

	"github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
)

var allowedNodeTypes = map[entity.NodeType]bool{
	entity.NodeActor: true, entity.NodeBusinessObject: true, entity.NodeProcess: true,
	entity.NodeEvent: true, entity.NodeDecision: true, entity.NodeSystem: true, entity.NodePolicy: true,
	entity.NodeRoleV12: true, entity.NodeEntityV12: true, entity.NodeProcessV12: true,
	entity.NodeStateV12: true, entity.NodeRuleV12: true, entity.NodeExternalV12: true,
	entity.NodeAgentV12: true, entity.NodeApplicationV12: true,
}

var allowedEdgeTypes = map[entity.EdgeType]bool{
	entity.EdgePerforms: true, entity.EdgeCreates: true, entity.EdgeUpdates: true,
	entity.EdgeTriggers: true, entity.EdgeRequires: true, entity.EdgeDependsOn: true,
	entity.EdgeTransitionsTo: true, entity.EdgeRelatesTo: true,
	"": true, // label-only edges from v1.2 allowed; treat as RELATES_TO
}

func ValidateSemanticModel(m *entity.SemanticModel) error {
	if m == nil {
		return fmt.Errorf("%w: model is nil", entity.ErrInvalidModel)
	}
	if m.SchemaVersion == "" {
		return fmt.Errorf("%w: schema_version required", entity.ErrInvalidModel)
	}
	if m.SchemaVersion != entity.SemanticSchemaVersion && m.SchemaVersion != "1.0" && m.SchemaVersion != "2.0" {
		return fmt.Errorf("%w: unsupported schema_version %q", entity.ErrInvalidModel, m.SchemaVersion)
	}
	if m.Nodes == nil {
		m.Nodes = []entity.SemanticNode{}
	}
	if m.Edges == nil {
		m.Edges = []entity.SemanticEdge{}
	}
	if m.Rules == nil {
		m.Rules = []entity.BusinessRule{}
	}
	if m.States == nil {
		m.States = []entity.BusinessState{}
	}

	nodeIDs := map[string]bool{}
	for i, n := range m.Nodes {
		id := strings.TrimSpace(n.ID)
		if id == "" {
			return fmt.Errorf("%w: node[%d] id empty", entity.ErrInvalidModel, i)
		}
		if nodeIDs[id] {
			return fmt.Errorf("%w: duplicate node id %q", entity.ErrInvalidModel, id)
		}
		nodeIDs[id] = true
		if strings.TrimSpace(n.Name) == "" {
			return fmt.Errorf("%w: node %q name empty", entity.ErrInvalidModel, id)
		}
		if !allowedNodeTypes[n.Type] {
			return fmt.Errorf("%w: unsupported node type %q", entity.ErrInvalidModel, n.Type)
		}
		if n.SourceMarker == "" {
			m.Nodes[i].SourceMarker = entity.SourceManualModified
		}
	}

	// Collect state IDs first so TRANSITIONS_TO may reference states as endpoints.
	stateIDs := map[string]bool{}
	for i, s := range m.States {
		id := strings.TrimSpace(s.ID)
		if id == "" {
			return fmt.Errorf("%w: state[%d] id empty", entity.ErrInvalidModel, i)
		}
		if stateIDs[id] || nodeIDs[id] {
			return fmt.Errorf("%w: duplicate state id %q", entity.ErrInvalidModel, id)
		}
		stateIDs[id] = true
		if strings.TrimSpace(s.Name) == "" {
			return fmt.Errorf("%w: state %q name empty", entity.ErrInvalidModel, id)
		}
		if s.ObjectRef == "" || !nodeIDs[s.ObjectRef] {
			return fmt.Errorf("%w: state %q object_ref missing/invalid", entity.ErrInvalidModel, id)
		}
		if s.SourceMarker == "" {
			m.States[i].SourceMarker = entity.SourceManualModified
		}
	}

	endpointOK := func(id string) bool { return nodeIDs[id] || stateIDs[id] }

	edgeIDs := map[string]bool{}
	for i, e := range m.Edges {
		id := strings.TrimSpace(e.ID)
		if id == "" {
			return fmt.Errorf("%w: edge[%d] id empty", entity.ErrInvalidModel, i)
		}
		if edgeIDs[id] {
			return fmt.Errorf("%w: duplicate edge id %q", entity.ErrInvalidModel, id)
		}
		edgeIDs[id] = true
		if !endpointOK(e.Source) || !endpointOK(e.Target) {
			return fmt.Errorf("%w: edge %q dangling endpoints", entity.ErrInvalidRelation, id)
		}
		if e.Source == e.Target && e.Type != entity.EdgeRelatesTo {
			return fmt.Errorf("%w: self-edge %q not allowed for type %q", entity.ErrInvalidRelation, id, e.Type)
		}
		if e.Type != "" && !allowedEdgeTypes[e.Type] {
			return fmt.Errorf("%w: unsupported edge type %q", entity.ErrInvalidModel, e.Type)
		}
		if e.Type == "" {
			m.Edges[i].Type = entity.EdgeRelatesTo
		}
		if e.SourceMarker == "" {
			m.Edges[i].SourceMarker = entity.SourceManualModified
		}
	}

	ruleIDs := map[string]bool{}
	for i, r := range m.Rules {
		id := strings.TrimSpace(r.ID)
		if id == "" {
			return fmt.Errorf("%w: rule[%d] id empty", entity.ErrInvalidModel, i)
		}
		if ruleIDs[id] {
			return fmt.Errorf("%w: duplicate rule id %q", entity.ErrInvalidModel, id)
		}
		ruleIDs[id] = true
		if strings.TrimSpace(r.Name) == "" {
			return fmt.Errorf("%w: rule %q name empty", entity.ErrInvalidModel, id)
		}
		for _, ref := range r.AppliesTo {
			if ref != "" && !endpointOK(ref) {
				return fmt.Errorf("%w: rule %q applies_to unknown ref %q", entity.ErrInvalidModel, id, ref)
			}
		}
		if r.SourceMarker == "" {
			m.Rules[i].SourceMarker = entity.SourceManualModified
		}
	}
	return nil
}
