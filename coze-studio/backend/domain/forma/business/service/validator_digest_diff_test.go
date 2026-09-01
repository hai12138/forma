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
