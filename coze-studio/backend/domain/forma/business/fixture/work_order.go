/*
 * Forma — Business-to-Agent Platform (proprietary)
 * Copyright (c) 2026 Forma. All rights reserved.
 */

package fixture

import "github.com/coze-dev/coze-studio/backend/domain/forma/business/entity"

// WorkOrderRepairSemanticModel is the S2 reference business (维修工单) — test/demo seed only.
// Not a platform core enum.
func WorkOrderRepairSemanticModel() *entity.SemanticModel {
	m := entity.SourceManualModified
	return &entity.SemanticModel{
		SchemaVersion: entity.SemanticSchemaVersion,
		Nodes: []entity.SemanticNode{
			{ID: "actor_reporter", Type: entity.NodeActor, Name: "报修人", Description: "发起故障报修的人员", SourceMarker: m},
			{ID: "actor_technician", Type: entity.NodeActor, Name: "维修人员", Description: "执行维修处理的人员", SourceMarker: m},
			{ID: "obj_work_order", Type: entity.NodeBusinessObject, Name: "维修工单", Description: "维修业务主对象", SourceMarker: m},
			{ID: "obj_equipment", Type: entity.NodeBusinessObject, Name: "设备", Description: "报修关联设备", SourceMarker: m},
			{ID: "proc_submit", Type: entity.NodeProcess, Name: "提交工单", SourceMarker: m},
			{ID: "proc_accept", Type: entity.NodeProcess, Name: "受理", SourceMarker: m},
			{ID: "proc_repair", Type: entity.NodeProcess, Name: "维修处理", SourceMarker: m},
			{ID: "proc_complete", Type: entity.NodeProcess, Name: "完工", SourceMarker: m},
			{ID: "evt_fault", Type: entity.NodeEvent, Name: "故障报修", SourceMarker: m},
		},
		Edges: []entity.SemanticEdge{
			{ID: "e_reporter_submit", Source: "actor_reporter", Target: "proc_submit", Type: entity.EdgePerforms, Label: "提交", SourceMarker: m},
			{ID: "e_submit_creates", Source: "proc_submit", Target: "obj_work_order", Type: entity.EdgeCreates, Label: "创建", SourceMarker: m},
			{ID: "e_fault_triggers", Source: "evt_fault", Target: "proc_submit", Type: entity.EdgeTriggers, Label: "触发", SourceMarker: m},
			{ID: "e_order_equip", Source: "obj_work_order", Target: "obj_equipment", Type: entity.EdgeRequires, Label: "关联", SourceMarker: m},
			{ID: "e_submit_accept", Source: "proc_submit", Target: "proc_accept", Type: entity.EdgeRelatesTo, Label: "下一步", SourceMarker: m},
			{ID: "e_accept_repair", Source: "proc_accept", Target: "proc_repair", Type: entity.EdgeRelatesTo, Label: "下一步", SourceMarker: m},
			{ID: "e_repair_complete", Source: "proc_repair", Target: "proc_complete", Type: entity.EdgeRelatesTo, Label: "下一步", SourceMarker: m},
			{ID: "e_tech_repair", Source: "actor_technician", Target: "proc_repair", Type: entity.EdgePerforms, Label: "执行", SourceMarker: m},
			{ID: "e_st_pending_progress", Source: "st_pending", Target: "st_in_progress", Type: entity.EdgeTransitionsTo, Label: "受理", SourceMarker: m},
			{ID: "e_st_progress_done", Source: "st_in_progress", Target: "st_done", Type: entity.EdgeTransitionsTo, Label: "完工", SourceMarker: m},
			{ID: "e_st_done_closed", Source: "st_done", Target: "st_closed", Type: entity.EdgeTransitionsTo, Label: "关闭", SourceMarker: m},
		},
		Rules: []entity.BusinessRule{
			{
				ID: "rule_close_permission", Name: "关闭权限",
				Description: "只有具备相应权限的人员可以关闭工单",
				Expression:  "actor.has_permission('work_order.close')",
				AppliesTo:   []string{"obj_work_order", "st_closed"},
				Severity:    "error",
				SourceMarker: m,
			},
		},
		States: []entity.BusinessState{
			{ID: "st_pending", ObjectRef: "obj_work_order", Name: "待受理", Initial: true, SourceMarker: m},
			{ID: "st_in_progress", ObjectRef: "obj_work_order", Name: "处理中", SourceMarker: m},
			{ID: "st_done", ObjectRef: "obj_work_order", Name: "已完成", SourceMarker: m},
			{ID: "st_closed", ObjectRef: "obj_work_order", Name: "已关闭", Terminal: true, SourceMarker: m},
		},
	}
}
