/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package entity

import (
	"time"
)

const SemanticSchemaVersion = "2.0"

// SourceMarker provenance for semantic elements.
type SourceMarker string

const (
	SourceAIGenerated    SourceMarker = "AI_GENERATED"
	SourceManualModified SourceMarker = "MANUAL_MODIFIED"
)

// NodeType — canonical Business Model types only (persisted SoT).
// v1.2 aliases (role/entity/process/external) may be converted by import/FE adapters
// before persistence. agent/application/state/rule must never be persisted as NodeType.
type NodeType string

const (
	NodeActor          NodeType = "ACTOR"
	NodeBusinessObject NodeType = "BUSINESS_OBJECT"
	NodeProcess        NodeType = "PROCESS"
	NodeEvent          NodeType = "EVENT"
	NodeDecision       NodeType = "DECISION"
	NodeSystem         NodeType = "SYSTEM"
	NodePolicy         NodeType = "POLICY"

	// Rejected / non-canonical markers (validator rejects; not stored).
	NodeRoleAlias        NodeType = "role"
	NodeEntityAlias      NodeType = "entity"
	NodeProcessAlias     NodeType = "process"
	NodeStateAlias       NodeType = "state"
	NodeRuleAlias        NodeType = "rule"
	NodeExternalAlias    NodeType = "external"
	NodeAgentAlias       NodeType = "agent"
	NodeApplicationAlias NodeType = "application"
)

type EdgeType string

const (
	EdgePerforms       EdgeType = "PERFORMS"
	EdgeCreates        EdgeType = "CREATES"
	EdgeUpdates        EdgeType = "UPDATES"
	EdgeTriggers       EdgeType = "TRIGGERS"
	EdgeRequires       EdgeType = "REQUIRES"
	EdgeDependsOn      EdgeType = "DEPENDS_ON"
	EdgeTransitionsTo  EdgeType = "TRANSITIONS_TO"
	EdgeRelatesTo      EdgeType = "RELATES_TO"
)

type SemanticNode struct {
	ID           string         `json:"id"`
	Type         NodeType       `json:"type"`
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	Properties   map[string]any `json:"properties,omitempty"`
	SourceMarker SourceMarker   `json:"source_marker"`
}

type SemanticEdge struct {
	ID           string         `json:"id"`
	Source       string         `json:"source"`
	Target       string         `json:"target"`
	Type         EdgeType       `json:"type"`
	Label        string         `json:"label,omitempty"`
	Description  string         `json:"description,omitempty"`
	Properties   map[string]any `json:"properties,omitempty"`
	SourceMarker SourceMarker   `json:"source_marker"`
}

type BusinessRule struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	Expression   string         `json:"expression,omitempty"`
	AppliesTo    []string       `json:"applies_to,omitempty"`
	Severity     string         `json:"severity,omitempty"`
	SourceMarker SourceMarker   `json:"source_marker"`
	Properties   map[string]any `json:"properties,omitempty"`
}

type BusinessState struct {
	ID           string         `json:"id"`
	ObjectRef    string         `json:"object_ref"`
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	Initial      bool           `json:"initial,omitempty"`
	Terminal     bool           `json:"terminal,omitempty"`
	SourceMarker SourceMarker   `json:"source_marker"`
	Properties   map[string]any `json:"properties,omitempty"`
}

// SemanticModel is the Business Model Source of Truth payload (no layout).
type SemanticModel struct {
	SchemaVersion string           `json:"schema_version"`
	Nodes         []SemanticNode   `json:"nodes"`
	Edges         []SemanticEdge   `json:"edges"`
	Rules         []BusinessRule   `json:"rules"`
	States        []BusinessState  `json:"states"`
	EvidenceRefs  []string         `json:"evidence_refs,omitempty"`
	AssertionRefs []string         `json:"assertion_refs,omitempty"`
}

// ViewLayout is presentation-only; never part of semantic digest/revision.
type ViewLayout struct {
	NodePositions map[string]NodePosition `json:"node_positions"`
	Zoom          float64                 `json:"zoom"`
	Viewport      Viewport                `json:"viewport"`
	Mode          string                  `json:"mode,omitempty"` // auto | manual
	Groups        [][]string              `json:"groups,omitempty"`
	Collapsed     map[string]bool         `json:"collapsed,omitempty"`
	Canvas        map[string]any          `json:"canvas,omitempty"`
}

type NodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Viewport struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// BusinessModel master — business_id == asset_id (ADR-013).
type BusinessModel struct {
	ID             int64
	TenantID       string
	BusinessID     string // = asset_id
	AssetID        string
	CurrentRevision int32
	SchemaVersion  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type BusinessModelRevision struct {
	ID              int64
	TenantID        string
	BusinessID      string
	RevisionNo      int32
	BaseRevisionNo  int32
	SchemaVersion   string
	SemanticModelJSON string
	ContentDigest   string
	ChangeSummary   string
	CreatedBy       string
	CreatedAt       time.Time
}

type BusinessModelLayout struct {
	ID                   int64
	TenantID             string
	BusinessID           string
	LayoutRevision       int32
	BasedOnModelRevision int32
	LayoutJSON           string
	UpdatedBy            string
	UpdatedAt            time.Time
}

type ElementDiff struct {
	Added    []string `json:"added"`
	Removed  []string `json:"removed"`
	Modified []string `json:"modified"`
}

type BusinessModelDiff struct {
	FromRevision int32       `json:"from_revision"`
	ToRevision   int32       `json:"to_revision"`
	Nodes        ElementDiff `json:"nodes"`
	Edges        ElementDiff `json:"edges"`
	Rules        ElementDiff `json:"rules"`
	States       ElementDiff `json:"states"`
}

type BusinessImpactSummary struct {
	SemanticChanged  bool     `json:"semantic_changed"`
	AffectedNodeIDs  []string `json:"affected_node_ids"`
	AffectedRuleIDs  []string `json:"affected_rule_ids"`
	AffectedStateIDs []string `json:"affected_state_ids"`
}
