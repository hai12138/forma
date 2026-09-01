import type { FormaSemanticModel, FormaViewLayout } from '@forma/api-client';

const M = 'MANUAL_MODIFIED' as const;

/** S2 reference seed: 维修工单 — formal types, MANUAL_MODIFIED. */
export function workOrderSeed(): FormaSemanticModel {
  return {
    schema_version: '2.0',
    nodes: [
      {
        id: 'actor_reporter',
        type: 'ACTOR',
        name: '报修人',
        description: '发起故障报修的人员',
        source_marker: M,
      },
      {
        id: 'actor_technician',
        type: 'ACTOR',
        name: '维修人员',
        description: '执行维修处理的人员',
        source_marker: M,
      },
      {
        id: 'obj_work_order',
        type: 'BUSINESS_OBJECT',
        name: '维修工单',
        description: '维修业务主对象',
        source_marker: M,
      },
      {
        id: 'obj_equipment',
        type: 'BUSINESS_OBJECT',
        name: '设备',
        description: '报修关联设备',
        source_marker: M,
      },
      {
        id: 'proc_submit',
        type: 'PROCESS',
        name: '提交工单',
        source_marker: M,
      },
      {
        id: 'proc_accept',
        type: 'PROCESS',
        name: '受理',
        source_marker: M,
      },
      {
        id: 'proc_repair',
        type: 'PROCESS',
        name: '维修处理',
        source_marker: M,
      },
      {
        id: 'proc_complete',
        type: 'PROCESS',
        name: '完工',
        source_marker: M,
      },
      {
        id: 'evt_fault',
        type: 'EVENT',
        name: '故障报修',
        source_marker: M,
      },
    ],
    edges: [
      {
        id: 'e_reporter_submit',
        source: 'actor_reporter',
        target: 'proc_submit',
        type: 'PERFORMS',
        label: '提交',
        source_marker: M,
      },
      {
        id: 'e_submit_creates',
        source: 'proc_submit',
        target: 'obj_work_order',
        type: 'CREATES',
        label: '创建',
        source_marker: M,
      },
      {
        id: 'e_fault_triggers',
        source: 'evt_fault',
        target: 'proc_submit',
        type: 'TRIGGERS',
        label: '触发',
        source_marker: M,
      },
      {
        id: 'e_order_equip',
        source: 'obj_work_order',
        target: 'obj_equipment',
        type: 'REQUIRES',
        label: '关联',
        source_marker: M,
      },
      {
        id: 'e_submit_accept',
        source: 'proc_submit',
        target: 'proc_accept',
        type: 'RELATES_TO',
        label: '下一步',
        source_marker: M,
      },
      {
        id: 'e_accept_repair',
        source: 'proc_accept',
        target: 'proc_repair',
        type: 'RELATES_TO',
        label: '下一步',
        source_marker: M,
      },
      {
        id: 'e_repair_complete',
        source: 'proc_repair',
        target: 'proc_complete',
        type: 'RELATES_TO',
        label: '下一步',
        source_marker: M,
      },
      {
        id: 'e_tech_repair',
        source: 'actor_technician',
        target: 'proc_repair',
        type: 'PERFORMS',
        label: '执行',
        source_marker: M,
      },
      {
        id: 'e_st_pending_progress',
        source: 'st_pending',
        target: 'st_in_progress',
        type: 'TRANSITIONS_TO',
        label: '受理',
        source_marker: M,
      },
      {
        id: 'e_st_progress_done',
        source: 'st_in_progress',
        target: 'st_done',
        type: 'TRANSITIONS_TO',
        label: '完工',
        source_marker: M,
      },
      {
        id: 'e_st_done_closed',
        source: 'st_done',
        target: 'st_closed',
        type: 'TRANSITIONS_TO',
        label: '关闭',
        source_marker: M,
      },
    ],
    rules: [
      {
        id: 'rule_close_permission',
        name: '关闭权限',
        description: '只有具备相应权限的人员可以关闭工单',
        expression: "actor.has_permission('work_order.close')",
        applies_to: ['obj_work_order', 'st_closed'],
        severity: 'error',
        source_marker: M,
      },
    ],
    states: [
      {
        id: 'st_pending',
        object_ref: 'obj_work_order',
        name: '待受理',
        initial: true,
        source_marker: M,
      },
      {
        id: 'st_in_progress',
        object_ref: 'obj_work_order',
        name: '处理中',
        source_marker: M,
      },
      {
        id: 'st_done',
        object_ref: 'obj_work_order',
        name: '已完成',
        source_marker: M,
      },
      {
        id: 'st_closed',
        object_ref: 'obj_work_order',
        name: '已关闭',
        terminal: true,
        source_marker: M,
      },
    ],
  };
}

export function workOrderDefaultLayout(): FormaViewLayout {
  return {
    node_positions: {
      actor_reporter: { x: 40, y: 40 },
      evt_fault: { x: 40, y: 160 },
      proc_submit: { x: 280, y: 100 },
      proc_accept: { x: 520, y: 100 },
      proc_repair: { x: 760, y: 100 },
      proc_complete: { x: 1000, y: 100 },
      actor_technician: { x: 760, y: 240 },
      obj_work_order: { x: 280, y: 280 },
      obj_equipment: { x: 520, y: 280 },
      st_pending: { x: 280, y: 420 },
      st_in_progress: { x: 520, y: 420 },
      st_done: { x: 760, y: 420 },
      st_closed: { x: 1000, y: 420 },
      rule_close_permission: { x: 1000, y: 280 },
    },
    zoom: 0.7,
    viewport: { x: 40, y: 40 },
    mode: 'manual',
    groups: [],
  };
}

export function emptySemanticModel(): FormaSemanticModel {
  return {
    schema_version: '2.0',
    nodes: [],
    edges: [],
    rules: [],
    states: [],
  };
}

export function emptyLayout(): FormaViewLayout {
  return {
    node_positions: {},
    zoom: 1,
    viewport: { x: 0, y: 0 },
    mode: 'manual',
    groups: [],
  };
}
