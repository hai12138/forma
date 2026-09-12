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

	"github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"
)

// CapabilityContentDigest digests only behavioral semantic fields (§3.4.2).
// Excludes source/provenance/ids/status/version/audit. Provenance-neutral.
func CapabilityContentDigest(payload entity.SemanticPayload) (string, error) {
	raw, err := canonicalSemanticJSON(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

// DecisionPayloadDigest digests the effective materialization / derive body.
// Excludes actor, timestamps, and client_request_id.
func DecisionPayloadDigest(payload entity.SemanticPayload) (string, error) {
	return CapabilityContentDigest(payload)
}

// AnalysisRequestDigest digests analysis request fields (§9.5).
func AnalysisRequestDigest(req entity.AnalysisRequest) (string, error) {
	raw, err := canonicalAnalysisRequestJSON(req)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func canonicalSemanticJSON(payload entity.SemanticPayload) ([]byte, error) {
	norm := normalizeSemanticPayload(payload)
	return json.Marshal(norm)
}

func canonicalAnalysisRequestJSON(req entity.AnalysisRequest) ([]byte, error) {
	pins := append([]entity.DataContractPin(nil), req.DataContractPins...)
	sort.Slice(pins, func(i, j int) bool {
		if pins[i].DataContractID != pins[j].DataContractID {
			return pins[i].DataContractID < pins[j].DataContractID
		}
		return pins[i].DataContractVersion < pins[j].DataContractVersion
	})
	refs := append([]string(nil), req.RequirementRefs...)
	sort.Strings(refs)
	if refs == nil {
		refs = []string{}
	}
	doc := map[string]any{
		"business_model_revision": req.BusinessModelRevision,
		"data_contract_pins":      pins,
		"requirement_refs":        refs,
	}
	return json.Marshal(doc)
}

func normalizeSemanticPayload(p entity.SemanticPayload) map[string]any {
	preconds := append([]entity.Precondition(nil), p.Preconditions...)
	sort.Slice(preconds, func(i, j int) bool { return preconds[i].ID < preconds[j].ID })
	effects := append([]entity.Effect(nil), p.Effects...)
	sort.Slice(effects, func(i, j int) bool { return effects[i].ID < effects[j].ID })
	bindings := append([]entity.DataContractBinding(nil), p.DataContractBindings...)
	sort.Slice(bindings, func(i, j int) bool {
		if bindings[i].DataContractID != bindings[j].DataContractID {
			return bindings[i].DataContractID < bindings[j].DataContractID
		}
		return bindings[i].DataContractRevisionID < bindings[j].DataContractRevisionID
	})
	for i := range bindings {
		maps := append([]entity.LogicalFieldMapping(nil), bindings[i].LogicalFieldMappings...)
		sort.Slice(maps, func(a, b int) bool {
			if maps[a].CapabilityLogicalKey != maps[b].CapabilityLogicalKey {
				return maps[a].CapabilityLogicalKey < maps[b].CapabilityLogicalKey
			}
			return maps[a].ContractLogicalKey < maps[b].ContractLogicalKey
		})
		if maps == nil {
			maps = []entity.LogicalFieldMapping{}
		}
		bindings[i].LogicalFieldMappings = maps
	}
	inFields := append([]entity.LogicalField(nil), p.InputSchema.Fields...)
	outFields := append([]entity.LogicalField(nil), p.OutputSchema.Fields...)
	// Field order is ordered (caller obligation order) — preserved, not sorted.
	if inFields == nil {
		inFields = []entity.LogicalField{}
	}
	if outFields == nil {
		outFields = []entity.LogicalField{}
	}
	if preconds == nil {
		preconds = []entity.Precondition{}
	}
	if effects == nil {
		effects = []entity.Effect{}
	}
	if bindings == nil {
		bindings = []entity.DataContractBinding{}
	}

	doc := map[string]any{
		"name":                    p.Name,
		"description":             p.Description,
		"capability_kind":         string(p.CapabilityKind),
		"business_model_revision": p.BusinessModelRevision,
		"input_schema":            map[string]any{"fields": inFields},
		"output_schema":           map[string]any{"fields": outFields},
		"preconditions":           normalizePreconditions(preconds),
		"effects":                 effects,
		"data_contract_bindings":  bindings,
	}
	if p.CapabilityKind == entity.KindQuery {
		doc["query_operation"] = string(p.QueryOperation)
		doc["output_cardinality"] = string(p.OutputCardinality)
	}
	return doc
}

func normalizePreconditions(in []entity.Precondition) []map[string]any {
	out := make([]map[string]any, 0, len(in))
	for _, p := range in {
		m := map[string]any{
			"id":          p.ID,
			"predicate":   p.Predicate,
			"logical_key": p.LogicalKey,
			"description": p.Description,
		}
		if p.Comparand != nil {
			m["comparand"] = p.Comparand
		}
		out = append(out, m)
	}
	return out
}

// MustCapabilityContentDigest panics on error — tests only helpers should avoid; service uses error return.
func MustCapabilityContentDigest(payload entity.SemanticPayload) string {
	d, err := CapabilityContentDigest(payload)
	if err != nil {
		panic(fmt.Sprintf("digest: %v", err))
	}
	return d
}
