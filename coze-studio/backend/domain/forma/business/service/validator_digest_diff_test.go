/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package service

import (
	"testing"

	"github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"
	"github.com/stretchr/testify/require"
)

func sampleModel() *entity.SemanticModel {
	return &entity.SemanticModel{
		SchemaVersion: entity.SemanticSchemaVersion,
		Nodes: []entity.SemanticNode{
			{ID: "n1", Type: entity.NodeActor, Name: "报修人", SourceMarker: entity.SourceManualModified},
			{ID: "n2", Type: entity.NodeBusinessObject, Name: "维修工单", SourceMarker: entity.SourceManualModified},
		},
		Edges: []entity.SemanticEdge{
			{ID: "e1", Source: "n1", Target: "n2", Type: entity.EdgeCreates, Label: "创建", SourceMarker: entity.SourceManualModified},
		},
		Rules: []entity.BusinessRule{
			{ID: "r1", Name: "关闭权限", Expression: "has_close_permission", AppliesTo: []string{"n2"}, SourceMarker: entity.SourceManualModified},
		},
		States: []entity.BusinessState{
			{ID: "s1", ObjectRef: "n2", Name: "待受理", Initial: true, SourceMarker: entity.SourceManualModified},
		},
	}
}

func TestValidateOK(t *testing.T) {
	require.NoError(t, ValidateSemanticModel(sampleModel()))
}

func TestValidateDuplicateNode(t *testing.T) {
	m := sampleModel()
	m.Nodes = append(m.Nodes, entity.SemanticNode{ID: "n1", Type: entity.NodeProcess, Name: "dup", SourceMarker: entity.SourceManualModified})
	err := ValidateSemanticModel(m)
	require.ErrorIs(t, err, entity.ErrInvalidModel)
}

func TestValidateDanglingEdge(t *testing.T) {
	m := sampleModel()
	m.Edges[0].Target = "missing"
	err := ValidateSemanticModel(m)
	require.ErrorIs(t, err, entity.ErrInvalidRelation)
}

func TestDigestDeterministic(t *testing.T) {
	a := sampleModel()
	b := sampleModel()
	// reverse node order
	b.Nodes[0], b.Nodes[1] = b.Nodes[1], b.Nodes[0]
	da, _, err := ContentDigest(a)
	require.NoError(t, err)
	db, _, err := ContentDigest(b)
	require.NoError(t, err)
	require.Equal(t, da, db)
}

func TestDiffElements(t *testing.T) {
	from := sampleModel()
	to := sampleModel()
	to.Nodes = append(to.Nodes, entity.SemanticNode{ID: "n3", Type: entity.NodeProcess, Name: "受理", SourceMarker: entity.SourceManualModified})
	to.Nodes[0].Name = "报修客户"
	to.Edges = nil
	d := DiffSemanticModels(1, 2, from, to)
	require.Contains(t, d.Nodes.Added, "n3")
	require.Contains(t, d.Nodes.Modified, "n1")
	require.Contains(t, d.Edges.Removed, "e1")
	imp := ImpactFromDiff(d)
	require.True(t, imp.SemanticChanged)
}

func TestRejectAgentApplicationNodes(t *testing.T) {
	m := sampleModel()
	m.Nodes = append(m.Nodes, entity.SemanticNode{ID: "agent1", Type: entity.NodeAgentAlias, Name: "Agent", SourceMarker: entity.SourceManualModified})
	require.ErrorIs(t, ValidateSemanticModel(m), entity.ErrInvalidModel)

	m2 := sampleModel()
	m2.Nodes = append(m2.Nodes, entity.SemanticNode{ID: "app1", Type: entity.NodeApplicationAlias, Name: "App", SourceMarker: entity.SourceManualModified})
	require.ErrorIs(t, ValidateSemanticModel(m2), entity.ErrInvalidModel)
}

func TestRejectStateRuleAsNodeType(t *testing.T) {
	m := sampleModel()
	m.Nodes = append(m.Nodes, entity.SemanticNode{ID: "st_as_node", Type: entity.NodeStateAlias, Name: "状态", SourceMarker: entity.SourceManualModified})
	require.ErrorIs(t, ValidateSemanticModel(m), entity.ErrInvalidModel)

	m2 := sampleModel()
	m2.Nodes = append(m2.Nodes, entity.SemanticNode{ID: "rule_as_node", Type: entity.NodeRuleAlias, Name: "规则", SourceMarker: entity.SourceManualModified})
	require.ErrorIs(t, ValidateSemanticModel(m2), entity.ErrInvalidModel)
}

func TestCanonicalizeLegacyAliases(t *testing.T) {
	m := &entity.SemanticModel{
		SchemaVersion: entity.SemanticSchemaVersion,
		Nodes: []entity.SemanticNode{
			{ID: "r1", Type: entity.NodeRoleAlias, Name: "角色", SourceMarker: entity.SourceManualModified},
			{ID: "e1", Type: entity.NodeEntityAlias, Name: "实体", SourceMarker: entity.SourceManualModified},
			{ID: "p1", Type: entity.NodeProcessAlias, Name: "流程", SourceMarker: entity.SourceManualModified},
			{ID: "x1", Type: entity.NodeExternalAlias, Name: "外部", SourceMarker: entity.SourceManualModified},
		},
		Edges:  []entity.SemanticEdge{},
		Rules:  []entity.BusinessRule{},
		States: []entity.BusinessState{},
	}
	require.NoError(t, ValidateSemanticModel(m))
	require.Equal(t, entity.NodeActor, m.Nodes[0].Type)
	require.Equal(t, entity.NodeBusinessObject, m.Nodes[1].Type)
	require.Equal(t, entity.NodeProcess, m.Nodes[2].Type)
	require.Equal(t, entity.NodeSystem, m.Nodes[3].Type)
}

func TestRejectInvalidSourceMarker(t *testing.T) {
	m := sampleModel()
	m.Nodes[0].SourceMarker = "unknown"
	require.ErrorIs(t, ValidateSemanticModel(m), entity.ErrInvalidModel)
	m.Nodes[0].SourceMarker = "manual"
	require.ErrorIs(t, ValidateSemanticModel(m), entity.ErrInvalidModel)
	m.Nodes[0].SourceMarker = "ai"
	require.ErrorIs(t, ValidateSemanticModel(m), entity.ErrInvalidModel)
	m.Nodes[0].SourceMarker = "foo"
	require.ErrorIs(t, ValidateSemanticModel(m), entity.ErrInvalidModel)
}

func TestEmptySourceMarkerDefaults(t *testing.T) {
	m := sampleModel()
	m.Nodes[0].SourceMarker = ""
	require.NoError(t, ValidateSemanticModel(m))
	require.Equal(t, entity.SourceManualModified, m.Nodes[0].SourceMarker)
}

func TestRejectEmptyEdgeLabel(t *testing.T) {
	m := sampleModel()
	m.Edges[0].Label = "  "
	require.ErrorIs(t, ValidateSemanticModel(m), entity.ErrInvalidModel)
	m.Edges[0].Label = ""
	require.ErrorIs(t, ValidateSemanticModel(m), entity.ErrInvalidModel)
}

func TestDiffDeterministicSorted(t *testing.T) {
	from := sampleModel()
	to := sampleModel()
	to.Nodes = append(to.Nodes,
		entity.SemanticNode{ID: "n9", Type: entity.NodeProcess, Name: "A", SourceMarker: entity.SourceManualModified},
		entity.SemanticNode{ID: "n3", Type: entity.NodeEvent, Name: "B", SourceMarker: entity.SourceManualModified},
		entity.SemanticNode{ID: "n5", Type: entity.NodeSystem, Name: "C", SourceMarker: entity.SourceManualModified},
	)
	to.Nodes[0].Name = "改"
	d1 := DiffSemanticModels(1, 2, from, to)
	d2 := DiffSemanticModels(1, 2, from, to)
	require.Equal(t, d1, d2)
	require.Equal(t, []string{"n3", "n5", "n9"}, d1.Nodes.Added)
	imp1 := ImpactFromDiff(d1)
	imp2 := ImpactFromDiff(d2)
	require.Equal(t, imp1, imp2)
	require.Equal(t, []string{"n1", "n3", "n5", "n9"}, imp1.AffectedNodeIDs)
}

func TestRejectGlobalSemanticIDCollision(t *testing.T) {
	m := sampleModel()
	m.States = append(m.States, entity.BusinessState{
		ID: "n1", ObjectRef: "n2", Name: "碰撞", SourceMarker: entity.SourceManualModified,
	})
	require.ErrorIs(t, ValidateSemanticModel(m), entity.ErrInvalidModel)

	m2 := sampleModel()
	m2.Rules = append(m2.Rules, entity.BusinessRule{
		ID: "e1", Name: "撞边", SourceMarker: entity.SourceManualModified,
	})
	require.ErrorIs(t, ValidateSemanticModel(m2), entity.ErrInvalidModel)

	m3 := sampleModel()
	m3.Edges = append(m3.Edges, entity.SemanticEdge{
		ID: "r1", Source: "n1", Target: "n2", Type: entity.EdgeRelatesTo, Label: "x",
		SourceMarker: entity.SourceManualModified,
	})
	require.ErrorIs(t, ValidateSemanticModel(m3), entity.ErrInvalidModel)
}
