/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"encoding/json"

	"github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
)

func DiffSemanticModels(fromRev, toRev int32, from, to *entity.SemanticModel) *entity.BusinessModelDiff {
	d := &entity.BusinessModelDiff{FromRevision: fromRev, ToRevision: toRev}
	if from == nil {
		from = &entity.SemanticModel{}
	}
	if to == nil {
		to = &entity.SemanticModel{}
	}
	d.Nodes = diffMaps(indexNodes(from.Nodes), indexNodes(to.Nodes))
	d.Edges = diffMaps(indexEdges(from.Edges), indexEdges(to.Edges))
	d.Rules = diffMaps(indexRules(from.Rules), indexRules(to.Rules))
	d.States = diffMaps(indexStates(from.States), indexStates(to.States))
	return d
}

func ImpactFromDiff(d *entity.BusinessModelDiff) *entity.BusinessImpactSummary {
	if d == nil {
		return &entity.BusinessImpactSummary{}
	}
	changed := len(d.Nodes.Added)+len(d.Nodes.Removed)+len(d.Nodes.Modified)+
		len(d.Edges.Added)+len(d.Edges.Removed)+len(d.Edges.Modified)+
		len(d.Rules.Added)+len(d.Rules.Removed)+len(d.Rules.Modified)+
		len(d.States.Added)+len(d.States.Removed)+len(d.States.Modified) > 0
	nodes := append(append(append([]string{}, d.Nodes.Added...), d.Nodes.Removed...), d.Nodes.Modified...)
	rules := append(append(append([]string{}, d.Rules.Added...), d.Rules.Removed...), d.Rules.Modified...)
	states := append(append(append([]string{}, d.States.Added...), d.States.Removed...), d.States.Modified...)
	return &entity.BusinessImpactSummary{
		SemanticChanged:  changed,
		AffectedNodeIDs:  nodes,
		AffectedRuleIDs:  rules,
		AffectedStateIDs: states,
	}
}

func indexNodes(xs []entity.SemanticNode) map[string]string {
	out := map[string]string{}
	for _, x := range xs {
		b, _ := json.Marshal(x)
		out[x.ID] = string(b)
	}
	return out
}

func indexEdges(xs []entity.SemanticEdge) map[string]string {
	out := map[string]string{}
	for _, x := range xs {
		b, _ := json.Marshal(x)
		out[x.ID] = string(b)
	}
	return out
}

func indexRules(xs []entity.BusinessRule) map[string]string {
	out := map[string]string{}
	for _, x := range xs {
		b, _ := json.Marshal(x)
		out[x.ID] = string(b)
	}
	return out
}

func indexStates(xs []entity.BusinessState) map[string]string {
	out := map[string]string{}
	for _, x := range xs {
		b, _ := json.Marshal(x)
		out[x.ID] = string(b)
	}
	return out
}

func diffMaps(a, b map[string]string) entity.ElementDiff {
	var d entity.ElementDiff
	for id, av := range a {
		bv, ok := b[id]
		if !ok {
			d.Removed = append(d.Removed, id)
			continue
		}
		if av != bv {
			d.Modified = append(d.Modified, id)
		}
	}
	for id := range b {
		if _, ok := a[id]; !ok {
			d.Added = append(d.Added, id)
		}
	}
	return d
}
