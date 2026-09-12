/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package fixture

import "github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"

// LaboratoryFlowCapability is an industry test fixture (lab equipment flow).
// Industry words appear only here — not in production packages.
func LaboratoryFlowCapability() entity.SemanticPayload {
	return entity.SemanticPayload{
		Name:                  "ListLaboratoryDevices",
		Description:           "List laboratory devices for a work cell",
		CapabilityKind:        entity.KindQuery,
		BusinessModelRevision: 1,
		InputSchema: entity.LogicalSchema{Fields: []entity.LogicalField{
			{LogicalKey: "work_cell_id", LogicalType: "STRING", Required: true},
		}},
		OutputSchema: entity.LogicalSchema{Fields: []entity.LogicalField{
			{LogicalKey: "device_id", LogicalType: "STRING"},
			{LogicalKey: "device_status", LogicalType: "STRING"},
		}},
		Preconditions: []entity.Precondition{
			{ID: "pc_cell", Predicate: entity.PredicateExists, LogicalKey: "work_cell_id"},
		},
		Effects: []entity.Effect{
			{ID: "ef_read", Kind: entity.EffectReadOnly, Description: "No laboratory mutation"},
		},
		DataContractBindings: []entity.DataContractBinding{
			{DataContractID: "dc_lab", DataContractRevisionID: "dcr_lab_1", DataContractVersion: 1},
		},
		QueryOperation:    entity.QueryOpList,
		OutputCardinality: entity.CardinalityMany,
	}
}

// LaboratoryCommandCapability is a COMMAND intent fixture (approve calibration).
func LaboratoryCommandCapability() entity.SemanticPayload {
	return entity.SemanticPayload{
		Name:                  "ApproveCalibration",
		Description:           "Approve laboratory calibration intent",
		CapabilityKind:        entity.KindCommand,
		BusinessModelRevision: 1,
		InputSchema: entity.LogicalSchema{Fields: []entity.LogicalField{
			{LogicalKey: "calibration_id", LogicalType: "STRING", Required: true},
		}},
		OutputSchema: entity.LogicalSchema{Fields: []entity.LogicalField{
			{LogicalKey: "approval_status", LogicalType: "STRING"},
		}},
		Effects: []entity.Effect{
			{ID: "ef_approve", Kind: entity.EffectIntent, Description: "Calibration approved"},
		},
		DataContractBindings: []entity.DataContractBinding{},
	}
}
