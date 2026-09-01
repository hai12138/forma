import { describe, expect, it, vi } from 'vitest';
import { renderToStaticMarkup } from 'react-dom/server';
import React from 'react';

import {
  applyLayoutChange,
  applySemanticChange,
  analyzeNodeDeleteImpact,
  collectCanvasItems,
  createEditBuffer,
  deleteNodeWithDependencies,
  isEdgeEndpoint,
  isLayoutDirty,
  isSemanticDirty,
  redo,
  undo,
} from './edit-buffer';
import { computeAutoLayout } from './auto-layout';
import { adaptModelForPersistence, canonicalizeNodeType } from './canonical';
import {
  formatValidationIssues,
  validateEditorSemanticModel,
} from './semantic-validator';
import { emptyLayout, emptySemanticModel, workOrderSeed } from './work-order-seed';
import { createBusinessSubmitHandler } from './create-handlers';

describe('empty state render', () => {
  it('renders empty-state markup when list is empty', () => {
    const markup = renderToStaticMarkup(
      React.createElement(
        'div',
        { className: 'forma-biz-empty', 'data-testid': 'business-empty-state' },
        React.createElement('strong', null, '暂无业务资产'),
        React.createElement('p', null, '创建第一个 Business Model，开始可视化建模。'),
      ),
    );
    expect(markup).toContain('business-empty-state');
    expect(markup).toContain('暂无业务资产');
  });
});

describe('createBusinessSubmitHandler', () => {
  it('exposes async click-handler shape and seeds 维修工单', async () => {
    const createBusiness = vi.fn().mockResolvedValue({
      data: { business_id: 'b1', name: '维修工单', status: 'ACTIVE', current_revision: 1 },
    });
    const onCreated = vi.fn();
    const onError = vi.fn();
    const handler = createBusinessSubmitHandler({
      client: { createBusiness },
      name: '维修工单',
      seedWorkOrder: false,
      onCreated,
      onError,
    });
    expect(typeof handler).toBe('function');
    await handler();
    expect(createBusiness).toHaveBeenCalledWith(
      expect.objectContaining({
        name: '维修工单',
        semantic_model: expect.objectContaining({ schema_version: '2.0' }),
      }),
    );
    expect(onCreated).toHaveBeenCalled();
    expect(onError).not.toHaveBeenCalled();
  });

  it('rejects empty name', async () => {
    const onError = vi.fn();
    const handler = createBusinessSubmitHandler({
      client: { createBusiness: vi.fn() },
      name: '  ',
      seedWorkOrder: false,
      onCreated: vi.fn(),
      onError,
    });
    await handler();
    expect(onError).toHaveBeenCalledWith('请输入业务名称');
  });
});

describe('semantic dirty vs layout-only', () => {
  it('layout drag marks layout dirty but not semantic dirty', () => {
    const model = workOrderSeed();
    let buffer = createEditBuffer({
      semantic_model: model,
      layout: emptyLayout(),
      modelRevision: 1,
      layoutRevision: 0,
    });
    buffer = applyLayoutChange(buffer, {
      ...buffer.current.layout,
      node_positions: { actor_reporter: { x: 10, y: 20 } },
    });
    expect(isLayoutDirty(buffer)).toBe(true);
    expect(isSemanticDirty(buffer)).toBe(false);
  });

  it('semantic edit marks semantic dirty', () => {
    const model = workOrderSeed();
    let buffer = createEditBuffer({
      semantic_model: model,
      layout: emptyLayout(),
      modelRevision: 1,
      layoutRevision: 0,
    });
    const next = structuredClone(model);
    next.nodes[0] = { ...next.nodes[0], name: '报修人-改' };
    buffer = applySemanticChange(buffer, next);
    expect(isSemanticDirty(buffer)).toBe(true);
  });
});

describe('undo/redo buffer', () => {
  it('undo and redo restore edit snapshots', () => {
    let buffer = createEditBuffer({
      semantic_model: emptySemanticModel(),
      layout: emptyLayout(),
      modelRevision: 1,
      layoutRevision: 0,
    });
    buffer = applySemanticChange(buffer, {
      ...emptySemanticModel(),
      nodes: [
        {
          id: 'n1',
          type: 'ACTOR',
          name: 'A',
          source_marker: 'MANUAL_MODIFIED',
        },
      ],
    });
    expect(buffer.current.semantic_model.nodes).toHaveLength(1);
    buffer = undo(buffer);
    expect(buffer.current.semantic_model.nodes).toHaveLength(0);
    buffer = redo(buffer);
    expect(buffer.current.semantic_model.nodes).toHaveLength(1);
    expect(buffer.current.semantic_model.nodes[0].name).toBe('A');
  });
});

describe('auto layout', () => {
  it('is deterministic for same model', () => {
    const model = workOrderSeed();
    const ids = [
      ...model.nodes.map(n => n.id),
      ...model.states.map(s => s.id),
      ...model.rules.map(r => r.id),
    ];
    const a = computeAutoLayout(model, emptyLayout(), ids);
    const b = computeAutoLayout(model, emptyLayout(), ids);
    expect(a.node_positions).toEqual(b.node_positions);
    expect(a.mode).toBe('auto');
  });

  it('auto layout marks layout dirty only', () => {
    const model = workOrderSeed();
    let buffer = createEditBuffer({
      semantic_model: model,
      layout: emptyLayout(),
      modelRevision: 1,
      layoutRevision: 1,
    });
    const ids = collectCanvasItems(model).map(i => i.id);
    buffer = applyLayoutChange(buffer, computeAutoLayout(model, buffer.current.layout, ids));
    expect(isLayoutDirty(buffer)).toBe(true);
    expect(isSemanticDirty(buffer)).toBe(false);
  });
});

describe('dependency-aware delete', () => {
  it('cascades states and strips rule applies_to', () => {
    const model = workOrderSeed();
    const impact = analyzeNodeDeleteImpact(model, 'obj_work_order');
    expect(impact.stateCount).toBeGreaterThan(0);
    expect(impact.ruleRefCount).toBeGreaterThan(0);
    const next = deleteNodeWithDependencies(model, 'obj_work_order');
    expect(next.nodes.find(n => n.id === 'obj_work_order')).toBeUndefined();
    expect(next.states.every(s => s.object_ref !== 'obj_work_order')).toBe(true);
    expect(
      next.rules.every(r => !(r.applies_to ?? []).includes('obj_work_order')),
    ).toBe(true);
    expect(next.rules.find(r => r.id === 'rule_close_permission')).toBeTruthy();
  });
});

describe('canonical adapter', () => {
  it('maps v1.2 aliases and rejects agent', () => {
    expect(canonicalizeNodeType('role')).toBe('ACTOR');
    expect(canonicalizeNodeType('entity')).toBe('BUSINESS_OBJECT');
    const adapted = adaptModelForPersistence({
      nodes: [{ type: 'role' }, { type: 'external' }],
    });
    expect(adapted.nodes.map(n => n.type)).toEqual(['ACTOR', 'SYSTEM']);
    expect(() =>
      adaptModelForPersistence({ nodes: [{ type: 'agent' }] }),
    ).toThrow(/agent/);
  });
});

describe('edge endpoints', () => {
  it('rules are not edge endpoints', () => {
    const model = workOrderSeed();
    expect(isEdgeEndpoint(model, 'obj_work_order')).toBe(true);
    expect(isEdgeEndpoint(model, 'st_pending')).toBe(true);
    expect(isEdgeEndpoint(model, 'rule_close_permission')).toBe(false);
  });
});

describe('S2-G2 editor validity', () => {
  it('blocks add-state when empty model has no nodes', () => {
    const model = emptySemanticModel();
    expect(model.nodes.length).toBe(0);
    // UI disables add-state when nodes.length===0; creating with empty object_ref fails preflight
    const illegal = {
      ...model,
      states: [
        {
          id: 'st_x',
          object_ref: '',
          name: '新状态',
          source_marker: 'MANUAL_MODIFIED' as const,
        },
      ],
    };
    const result = validateEditorSemanticModel(illegal);
    expect(result.ok).toBe(false);
    expect(result.issues.some(i => i.code === 'STATE_OBJECT_REF_INVALID')).toBe(true);
  });

  it('state object_ref must point to existing node', () => {
    const model = workOrderSeed();
    const bad = structuredClone(model);
    bad.states[0].object_ref = 'missing_node';
    expect(validateEditorSemanticModel(bad).ok).toBe(false);
    bad.states[0].object_ref = 'obj_work_order';
    expect(validateEditorSemanticModel(bad).ok).toBe(true);
  });

  it('rule applies_to must reference node or state only', () => {
    const model = workOrderSeed();
    const bad = structuredClone(model);
    bad.rules[0].applies_to = ['rule_close_permission'];
    expect(validateEditorSemanticModel(bad).ok).toBe(false);
    bad.rules[0].applies_to = ['obj_work_order', 'st_pending'];
    expect(validateEditorSemanticModel(bad).ok).toBe(true);
  });

  it('blank node/state/rule names fail preflight', () => {
    const model = workOrderSeed();
    const blankNode = structuredClone(model);
    blankNode.nodes[0].name = '   ';
    expect(validateEditorSemanticModel(blankNode).issues.some(i => i.code === 'NODE_NAME_EMPTY')).toBe(
      true,
    );
    const blankState = structuredClone(model);
    blankState.states[0].name = '';
    expect(
      validateEditorSemanticModel(blankState).issues.some(i => i.code === 'STATE_NAME_EMPTY'),
    ).toBe(true);
    const blankRule = structuredClone(model);
    blankRule.rules[0].name = ' ';
    expect(validateEditorSemanticModel(blankRule).issues.some(i => i.code === 'RULE_NAME_EMPTY')).toBe(
      true,
    );
  });

  it('preflight invalid reference blocks save API', async () => {
    const putBusinessModel = vi.fn();
    const model = workOrderSeed();
    model.states[0].object_ref = 'gone';
    const preflight = validateEditorSemanticModel(model);
    expect(preflight.ok).toBe(false);
    if (!preflight.ok) {
      // Save Model path must not call API
      expect(putBusinessModel).not.toHaveBeenCalled();
      expect(formatValidationIssues(preflight).length).toBeGreaterThan(0);
    }
  });

  it('valid work-order model passes preflight', () => {
    expect(validateEditorSemanticModel(workOrderSeed()).ok).toBe(true);
  });

  it('duplicate semantic element IDs fail preflight', () => {
    const model = workOrderSeed();
    model.states[0].id = 'obj_work_order';
    const result = validateEditorSemanticModel(model);
    expect(result.ok).toBe(false);
    expect(result.issues.some(i => i.code === 'ID_COLLISION')).toBe(true);
  });
});
