/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package fixture

import "github.com/coze-dev/coze-studio/backend/domain/forma/capability/entity"

// ProcurementApprovalCapability is an industry test fixture (procurement).
// Industry words appear only here — not in production packages.
func ProcurementApprovalCapability() entity.SemanticPayload {
	return entity.SemanticPayload{
		Name:                  "SubmitProcurementApproval",
		Description:           "Submit procurement order for approval",
		CapabilityKind:        entity.KindCommand,
		BusinessModelRevision: 2,
		InputSchema: entity.LogicalSchema{Fields: []entity.LogicalField{
			{LogicalKey: "purchase_order_id", LogicalType: "STRING", Required: true},
			{LogicalKey: "amount", LogicalType: "DECIMAL", Required: true},
		}},
		OutputSchema: entity.LogicalSchema{Fields: []entity.LogicalField{
			{LogicalKey: "approval_request_id", LogicalType: "STRING"},
		}},
		Preconditions: []entity.Precondition{
			{ID: "pc_amount", Predicate: "GT", LogicalKey: "amount", Operator: "GT", Comparand: 0},
		},
		Effects: []entity.Effect{
			{ID: "ef_submit", Kind: "INTENT", Description: "Procurement approval submitted"},
		},
		DataContractBindings: []entity.DataContractBinding{},
	}
}

// ProcurementQueryCapability lists pending procurement approvals.
func ProcurementQueryCapability() entity.SemanticPayload {
	return entity.SemanticPayload{
		Name:                  "ListPendingProcurementApprovals",
		Description:           "List pending procurement approval requests",
		CapabilityKind:        entity.KindQuery,
		BusinessModelRevision: 2,
		InputSchema: entity.LogicalSchema{Fields: []entity.LogicalField{
			{LogicalKey: "buyer_org_id", LogicalType: "STRING", Required: true},
		}},
		OutputSchema: entity.LogicalSchema{Fields: []entity.LogicalField{
			{LogicalKey: "approval_request_id", LogicalType: "STRING"},
			{LogicalKey: "purchase_order_id", LogicalType: "STRING"},
		}},
		DataContractBindings: []entity.DataContractBinding{
			{DataContractID: "dc_proc", DataContractRevisionID: "dcr_proc_1", DataContractVersion: 1},
		},
		QueryOperation:    entity.QueryOpFilter,
		OutputCardinality: entity.CardinalityMany,
	}
}
