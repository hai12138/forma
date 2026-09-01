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

	"github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
)

// CanonicalSerialize produces deterministic JSON (sorted object keys via encoding/json map sort + structured rebuild).
func CanonicalSerialize(m *entity.SemanticModel) ([]byte, error) {
	if m == nil {
		return nil, fmt.Errorf("nil model")
	}
	normalizeSemantic(m)
	// Rebuild with sorted slices by ID so order of arrays does not affect digest.
	cp := *m
	cp.Nodes = append([]entity.SemanticNode(nil), m.Nodes...)
	cp.Edges = append([]entity.SemanticEdge(nil), m.Edges...)
	cp.Rules = append([]entity.BusinessRule(nil), m.Rules...)
	cp.States = append([]entity.BusinessState(nil), m.States...)
	sort.Slice(cp.Nodes, func(i, j int) bool { return cp.Nodes[i].ID < cp.Nodes[j].ID })
	sort.Slice(cp.Edges, func(i, j int) bool { return cp.Edges[i].ID < cp.Edges[j].ID })
	sort.Slice(cp.Rules, func(i, j int) bool { return cp.Rules[i].ID < cp.Rules[j].ID })
	sort.Slice(cp.States, func(i, j int) bool { return cp.States[i].ID < cp.States[j].ID })
	for i := range cp.Nodes {
		cp.Nodes[i].Properties = sortedAnyMap(cp.Nodes[i].Properties)
	}
	for i := range cp.Edges {
		cp.Edges[i].Properties = sortedAnyMap(cp.Edges[i].Properties)
	}
	for i := range cp.Rules {
		cp.Rules[i].Properties = sortedAnyMap(cp.Rules[i].Properties)
		cp.Rules[i].AppliesTo = append([]string(nil), cp.Rules[i].AppliesTo...)
		sort.Strings(cp.Rules[i].AppliesTo)
	}
	for i := range cp.States {
		cp.States[i].Properties = sortedAnyMap(cp.States[i].Properties)
	}
	cp.EvidenceRefs = append([]string(nil), cp.EvidenceRefs...)
	cp.AssertionRefs = append([]string(nil), cp.AssertionRefs...)
	sort.Strings(cp.EvidenceRefs)
	sort.Strings(cp.AssertionRefs)

	return json.Marshal(cp)
}

func sortedAnyMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	keys := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make(map[string]any, len(in))
	for _, k := range keys {
		out[k] = in[k]
	}
	return out
}

func ContentDigest(m *entity.SemanticModel) (string, []byte, error) {
	raw, err := CanonicalSerialize(m)
	if err != nil {
		return "", nil, err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), raw, nil
}
