import { useEffect, useId, useRef, useState, type CSSProperties, type RefObject } from 'react';

import type {
  FormaBusinessRule,
  FormaBusinessState,
  FormaSemanticEdge,
  FormaSemanticModel,
  FormaSemanticNode,
  FormaViewLayout,
} from '@forma/api-client';

import { computeAutoLayout } from '../auto-layout';
import { EDGE_TYPES } from '../canonical';
import {
  analyzeNodeDeleteImpact,
  applyLayoutChange,
  applySemanticChange,
  collectCanvasItems,
  deleteNodeWithDependencies,
  isEdgeEndpoint,
  isLayoutDirty,
  isSemanticDirty,
  redo,
  undo,
  type EditBuffer,
} from '../edit-buffer';
import { ADDABLE_NODE_TYPES, styleFor } from '../node-style';

const NODE_W = 190;
const NODE_H = 88;

type Props = {
  buffer: EditBuffer;
  onBufferChange: (next: EditBuffer) => void;
  businessName: string;
  onSaveModel: () => void;
  onSaveLayout: () => void;
  savingModel?: boolean;
  savingLayout?: boolean;
  onOpenRevisions: () => void;
  onOpenDiff: () => void;
  message?: string;
  /** Historical revision view — canvas + props read-only. */
  readOnly?: boolean;
  viewingRevision?: number | null;
  onBackToCurrent?: () => void;
};

export function VisualModelEditor({
  buffer,
  onBufferChange,
  businessName,
  onSaveModel,
  onSaveLayout,
  savingModel,
  savingLayout,
  onOpenRevisions,
  onOpenDiff,
  message,
  readOnly = false,
  viewingRevision = null,
  onBackToCurrent,
}: Props) {
  const root = useRef<HTMLElement>(null);
  const surface = useRef<HTMLDivElement>(null);
  const nameInputRef = useRef<HTMLInputElement>(null);
  const marker = useId().replace(/:/g, '');
  const [selected, setSelected] = useState<string | null>(null);
  const [tool, setTool] = useState<'select' | 'pan'>('select');
  const [connection, setConnection] = useState<string | null>(null);
  const [preview, setPreview] = useState<FormaViewLayout | null>(null);
  const [fullscreen, setFullscreen] = useState(false);
  const [focusName, setFocusName] = useState(false);
  const [hint, setHint] = useState(
    message ?? '拖拽调整布局；编辑属性保存语义；布局与语义分别保存。',
  );
  const drag = useRef<{
    kind: 'node' | 'pan' | 'edge';
    id?: string;
    x: number;
    y: number;
    layout: FormaViewLayout;
    moved: boolean;
  } | null>(null);

  const model = buffer.current.semantic_model;
  const layout = preview ?? buffer.current.layout;
  const items = collectCanvasItems(model);
  const semanticDirty = isSemanticDirty(buffer);
  const layoutDirty = isLayoutDirty(buffer);
  const layoutMode = layout.mode === 'auto' ? 'auto' : 'manual';

  useEffect(() => {
    const sync = () => setFullscreen(document.fullscreenElement === root.current);
    document.addEventListener('fullscreenchange', sync);
    return () => document.removeEventListener('fullscreenchange', sync);
  }, []);

  useEffect(() => {
    if (message) setHint(message);
  }, [message]);

  useEffect(() => {
    if (focusName && nameInputRef.current) {
      nameInputRef.current.focus();
      nameInputRef.current.select();
      setFocusName(false);
    }
  }, [focusName, selected]);

  const setLayout = (next: FormaViewLayout, note: string) => {
    if (readOnly) return;
    onBufferChange(applyLayoutChange(buffer, next));
    setHint(note);
  };

  const setSemantic = (next: FormaSemanticModel, note: string, lay?: FormaViewLayout) => {
    if (readOnly) return;
    onBufferChange(applySemanticChange(buffer, next, lay));
    setHint(note);
  };

  const fit = () => {
    const points = Object.values(buffer.current.layout.node_positions);
    if (!points.length) return;
    const left = Math.min(...points.map(p => p.x));
    const top = Math.min(...points.map(p => p.y));
    const right = Math.max(...points.map(p => p.x + NODE_W));
    const bottom = Math.max(...points.map(p => p.y + NODE_H));
    const width = surface.current?.clientWidth ?? 900;
    const height = surface.current?.clientHeight ?? 560;
    const zoom = Math.max(
      0.05,
      Math.min(1.5, (width - 80) / (right - left), (height - 80) / (bottom - top)),
    );
    setLayout(
      {
        ...buffer.current.layout,
        zoom,
        viewport: {
          x: (width - (right + left) * zoom) / 2,
          y: (height - (bottom + top) * zoom) / 2,
        },
      },
      '适配视图',
    );
  };

  const runAutoLayout = () => {
    if (readOnly) return;
    const next = computeAutoLayout(
      model,
      buffer.current.layout,
      items.map(i => i.id),
    );
    setLayout(next, 'Auto Layout · 仅布局脏（语义未变）');
  };

  const connect = (from: string, target: string) => {
    if (readOnly) return;
    if (!isEdgeEndpoint(model, from) || !isEdgeEndpoint(model, target)) {
      setHint('关系只能连接节点或状态（规则不是端点）。');
      setConnection(null);
      return;
    }
    if (from === target) {
      setHint('请选择另一个节点作为关系终点。');
      return;
    }
    const id = `e_${Date.now().toString(36)}`;
    setSemantic(
      {
        ...model,
        edges: [
          ...model.edges,
          {
            id,
            source: from,
            target,
            type: 'RELATES_TO',
            label: '关联',
            source_marker: 'MANUAL_MODIFIED',
          },
        ],
      },
      '创建关系 · 未保存语义变更',
    );
    setSelected(id);
    setConnection(null);
  };

  const removeSelected = () => {
    if (readOnly || !selected) return;
    if (model.nodes.some(n => n.id === selected)) {
      const impact = analyzeNodeDeleteImpact(model, selected);
      if (impact.edgeCount || impact.stateCount || impact.ruleRefCount) {
        const ok = window.confirm(
          `删除该业务元素将同时影响：\n` +
            `${impact.edgeCount} 个关系\n` +
            `${impact.stateCount} 个状态（将删除）\n` +
            `${impact.ruleRefCount} 个规则引用（仅移除引用，规则保留）`,
        );
        if (!ok) return;
      }
      const next = deleteNodeWithDependencies(model, selected);
      const positions = { ...buffer.current.layout.node_positions };
      delete positions[selected];
      for (const sid of impact.dependentStateIds) delete positions[sid];
      setSemantic(next, '删除节点（依赖级联）', {
        ...buffer.current.layout,
        node_positions: positions,
      });
    } else if (model.edges.some(e => e.id === selected)) {
      setSemantic(
        { ...model, edges: model.edges.filter(e => e.id !== selected) },
        '删除关系',
      );
    } else if (model.states.some(s => s.id === selected)) {
      setSemantic(
        {
          ...model,
          states: model.states.filter(s => s.id !== selected),
          edges: model.edges.filter(e => e.source !== selected && e.target !== selected),
        },
        '删除状态',
      );
    } else if (model.rules.some(r => r.id === selected)) {
      setSemantic(
        { ...model, rules: model.rules.filter(r => r.id !== selected) },
        '删除规则',
      );
    }
    setSelected(null);
  };

  const addNode = (type: string) => {
    if (readOnly) return;
    const id = `n_${Date.now().toString(36)}`;
    const st = styleFor(type);
    const pos = {
      x: 80 + items.length * 24,
      y: 80 + items.length * 16,
    };
    setSemantic(
      {
        ...model,
        nodes: [
          ...model.nodes,
          {
            id,
            type,
            name: `新${st.label}`,
            source_marker: 'MANUAL_MODIFIED',
          },
        ],
      },
      '新增节点 · 未保存语义变更',
      {
        ...buffer.current.layout,
        mode: 'manual',
        node_positions: { ...buffer.current.layout.node_positions, [id]: pos },
      },
    );
    setSelected(id);
  };

  const selectedNode = model.nodes.find(n => n.id === selected);
  const selectedEdge = model.edges.find(e => e.id === selected);
  const selectedState = model.states.find(s => s.id === selected);
  const selectedRule = model.rules.find(r => r.id === selected);

  return (
    <section
      className="forma-vme"
      ref={root}
      aria-label="Visual Business Model Editor"
      data-readonly={readOnly ? 'true' : 'false'}
      data-layout-mode={layoutMode}
      data-semantic-dirty={semanticDirty ? 'true' : 'false'}
      data-layout-dirty={layoutDirty ? 'true' : 'false'}
      data-model-revision={buffer.modelRevision}
      data-layout-revision={buffer.layoutRevision}
      onKeyDown={e => {
        if (readOnly) return;
        if ((e.target as HTMLElement).closest('input,textarea,select')) return;
        if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'z') {
          e.preventDefault();
          onBufferChange(e.shiftKey ? redo(buffer) : undo(buffer));
        }
        if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'y') {
          e.preventDefault();
          onBufferChange(redo(buffer));
        }
        if (e.key === 'Escape') {
          setConnection(null);
          setSelected(null);
          setPreview(null);
          drag.current = null;
        }
        if (e.key === 'Delete' || e.key === 'Backspace') {
          e.preventDefault();
          removeSelected();
        }
      }}
    >
      {readOnly && viewingRevision != null && (
        <div
          className="forma-vme-readonly-banner"
          data-testid="revision-readonly-banner"
          style={{
            background: '#fff8e6',
            borderBottom: '1px solid #f0d48a',
            padding: '8px 16px',
            display: 'flex',
            gap: 12,
            alignItems: 'center',
          }}
        >
          <strong>Viewing revision r{viewingRevision}</strong>
          <span>Read only</span>
          <button type="button" className="forma-vme-btn primary" onClick={onBackToCurrent}>
            Back to Current
          </button>
        </div>
      )}

      <header className="forma-vme-header">
        <div className="forma-vme-header-title">
          <span className="forma-vme-dot" />
          <b>{businessName}</b>
          <small>
            r{buffer.modelRevision} · lr{buffer.layoutRevision} · {items.length} 节点 /{' '}
            {model.edges.length} 关系 · {layoutMode}
          </small>
          {semanticDirty && (
            <span className="forma-vme-dirty" data-testid="semantic-dirty">
              未保存语义变更
            </span>
          )}
          {!semanticDirty && layoutDirty && (
            <span className="forma-vme-dirty" data-testid="layout-dirty">
              布局未保存
            </span>
          )}
        </div>
        <div className="forma-vme-actions">
          {!readOnly && (
            <>
              <button
                type="button"
                className="forma-vme-btn primary"
                data-testid="save-model"
                disabled={!semanticDirty || savingModel}
                onClick={onSaveModel}
              >
                {savingModel ? '保存中…' : 'Save Model'}
              </button>
              <button
                type="button"
                className="forma-vme-btn"
                data-testid="save-layout"
                disabled={!layoutDirty || savingLayout}
                onClick={onSaveLayout}
              >
                {savingLayout ? '保存中…' : 'Save Layout'}
              </button>
              <button
                type="button"
                className="forma-vme-btn"
                data-testid="auto-layout"
                onClick={runAutoLayout}
              >
                Auto Layout
              </button>
              <button
                type="button"
                className="forma-vme-btn"
                aria-pressed={layoutMode === 'manual'}
                onClick={() =>
                  setLayout({ ...layout, mode: 'manual' }, 'Manual layout mode')
                }
              >
                Manual
              </button>
              <button
                type="button"
                className="forma-vme-btn"
                disabled={!buffer.past.length}
                onClick={() => onBufferChange(undo(buffer))}
              >
                Undo
              </button>
              <button
                type="button"
                className="forma-vme-btn"
                disabled={!buffer.future.length}
                onClick={() => onBufferChange(redo(buffer))}
              >
                Redo
              </button>
            </>
          )}
          <button
            type="button"
            className="forma-vme-btn"
            onClick={() =>
              setLayout(
                { ...layout, zoom: Math.min(2, layout.zoom * 1.15) },
                '放大',
              )
            }
          >
            Zoom +
          </button>
          <button
            type="button"
            className="forma-vme-btn"
            onClick={() =>
              setLayout(
                { ...layout, zoom: Math.max(0.2, layout.zoom / 1.15) },
                '缩小',
              )
            }
          >
            Zoom −
          </button>
          <button type="button" className="forma-vme-btn" onClick={fit}>
            Fit View
          </button>
          <button
            type="button"
            className="forma-vme-btn"
            onClick={async () => {
              try {
                if (document.fullscreenElement) await document.exitFullscreen();
                else await root.current?.requestFullscreen();
              } catch {
                setHint('当前环境不支持全屏');
              }
            }}
          >
            {fullscreen ? 'Exit Fullscreen' : 'Fullscreen'}
          </button>
          {!readOnly && (
            <>
              <button type="button" className="forma-vme-btn" onClick={onOpenRevisions}>
                Revisions
              </button>
              <button type="button" className="forma-vme-btn" onClick={onOpenDiff}>
                Diff
              </button>
            </>
          )}
        </div>
      </header>

      {!readOnly && (
        <div className="forma-vme-toolbar" role="toolbar" aria-label="模型编辑工具">
          {ADDABLE_NODE_TYPES.map(type => (
            <button
              key={type}
              type="button"
              className="forma-vme-btn"
              onClick={() => addNode(type)}
            >
              ＋{styleFor(type).label}
            </button>
          ))}
          <span className="forma-vme-sep" />
          <button
            type="button"
            className="forma-vme-btn"
            data-testid="add-state"
            disabled={model.nodes.length === 0}
            title={
              model.nodes.length === 0
                ? '请先创建一个业务元素，再为其定义状态。'
                : '新增状态'
            }
            onClick={() => {
              if (model.nodes.length === 0) {
                setHint('请先创建一个业务元素，再为其定义状态。');
                return;
              }
              const id = `st_${Date.now().toString(36)}`;
              setSemantic(
                {
                  ...model,
                  states: [
                    ...model.states,
                    {
                      id,
                      object_ref: model.nodes[0].id,
                      name: '新状态',
                      source_marker: 'MANUAL_MODIFIED',
                    },
                  ],
                },
                '新增状态',
                {
                  ...buffer.current.layout,
                  node_positions: {
                    ...buffer.current.layout.node_positions,
                    [id]: { x: 100, y: 400 },
                  },
                },
              );
              setSelected(id);
            }}
          >
            ＋状态
          </button>
          <button
            type="button"
            className="forma-vme-btn"
            data-testid="add-rule"
            onClick={() => {
              const id = `rule_${Date.now().toString(36)}`;
              setSemantic(
                {
                  ...model,
                  rules: [
                    ...model.rules,
                    {
                      id,
                      name: '新规则',
                      applies_to: model.nodes[0] ? [model.nodes[0].id] : [],
                      source_marker: 'MANUAL_MODIFIED',
                    },
                  ],
                },
                '新增规则',
                {
                  ...buffer.current.layout,
                  node_positions: {
                    ...buffer.current.layout.node_positions,
                    [id]: { x: 100, y: 520 },
                  },
                },
              );
              setSelected(id);
            }}
          >
            ＋规则
          </button>
        </div>
      )}

      <div className="forma-vme-body">
        <div
          className={`forma-vme-surface${tool === 'pan' ? ' forma-vme-panning' : ''}`}
          ref={surface}
        >
          <div className="forma-vme-tools">
            <button
              type="button"
              className="forma-vme-btn"
              aria-pressed={tool === 'select'}
              onClick={() => setTool('select')}
            >
              选择
            </button>
            <button
              type="button"
              className="forma-vme-btn"
              aria-pressed={tool === 'pan'}
              onClick={() => setTool('pan')}
            >
              平移
            </button>
            <button type="button" className="forma-vme-btn" onClick={fit}>
              适配
            </button>
          </div>

          <svg
            role="application"
            aria-label="可编辑业务模型画布"
            tabIndex={0}
            onPointerDown={e => {
              if (readOnly || e.button !== 0) return;
              const target = (e.target as Element).closest('[data-node]');
              const id = target?.getAttribute('data-node') ?? undefined;
              if ((e.target as Element).closest('[data-edge]')) return;
              const kind = (e.target as Element).closest('[data-handle]')
                ? 'edge'
                : tool === 'pan' || !id
                  ? 'pan'
                  : 'node';
              drag.current = {
                kind,
                id,
                x: e.clientX,
                y: e.clientY,
                layout: structuredClone(buffer.current.layout),
                moved: false,
              };
              if (kind === 'edge' && id) {
                setConnection(id);
                setHint('松开到目标节点/状态，或点击目标建立关系。');
              }
              e.currentTarget.setPointerCapture(e.pointerId);
            }}
            onPointerMove={e => {
              if (readOnly) return;
              const d = drag.current;
              if (!d) return;
              const dx = e.clientX - d.x;
              const dy = e.clientY - d.y;
              if (Math.abs(dx) + Math.abs(dy) < 3) return;
              d.moved = true;
              if (d.kind === 'pan') {
                setPreview({
                  ...d.layout,
                  viewport: {
                    x: d.layout.viewport.x + dx,
                    y: d.layout.viewport.y + dy,
                  },
                });
              }
              if (d.kind === 'node' && d.id) {
                const p = d.layout.node_positions[d.id] ?? { x: 0, y: 0 };
                setPreview({
                  ...d.layout,
                  mode: 'manual',
                  node_positions: {
                    ...d.layout.node_positions,
                    [d.id]: {
                      x: p.x + dx / d.layout.zoom,
                      y: p.y + dy / d.layout.zoom,
                    },
                  },
                });
              }
            }}
            onPointerUp={e => {
              if (readOnly) return;
              const d = drag.current;
              if (!d) return;
              const target = document
                .elementFromPoint(e.clientX, e.clientY)
                ?.closest('[data-node]')
                ?.getAttribute('data-node');
              if (d.kind === 'edge' && d.id && target && target !== d.id) {
                connect(d.id, target);
              } else if (d.moved && preview) {
                setLayout(
                  preview,
                  d.kind === 'node' ? '移动节点 · 仅更新布局' : '平移画布',
                );
              } else if (!d.moved && d.kind === 'node' && d.id) {
                if (connection) connect(connection, d.id);
                else setSelected(d.id);
              }
              drag.current = null;
              setPreview(null);
            }}
            onPointerCancel={() => {
              drag.current = null;
              setPreview(null);
              setConnection(null);
            }}
          >
            <defs>
              <marker
                id={marker}
                viewBox="0 0 10 10"
                refX="9"
                refY="5"
                markerWidth="7"
                markerHeight="7"
                orient="auto-start-reverse"
              >
                <path d="M1 1 L9 5 L1 9" fill="none" stroke="#8d99ad" />
              </marker>
            </defs>
            <g
              transform={`translate(${layout.viewport.x},${layout.viewport.y}) scale(${layout.zoom})`}
            >
              {model.edges.map(edge => {
                const a = layout.node_positions[edge.source];
                const b = layout.node_positions[edge.target];
                if (!a || !b) return null;
                const x1 = a.x + NODE_W;
                const y1 = a.y + NODE_H / 2;
                const x2 = b.x;
                const y2 = b.y + NODE_H / 2;
                const mx = (x1 + x2) / 2;
                const path = `M${x1},${y1} C${mx},${y1} ${mx},${y2} ${x2},${y2}`;
                return (
                  <g
                    key={edge.id}
                    data-edge={edge.id}
                    role="button"
                    tabIndex={0}
                    aria-label={`关系 ${edge.label || edge.type}`}
                    onClick={() => setSelected(edge.id)}
                    onKeyDown={ev => {
                      if (ev.key === 'Enter') setSelected(edge.id);
                    }}
                    style={{ cursor: 'pointer' }}
                  >
                    <path
                      d={path}
                      fill="none"
                      stroke={selected === edge.id ? '#2860dd' : '#8d99ad'}
                      strokeWidth={selected === edge.id ? 2.5 : 1.5}
                      markerEnd={`url(#${marker})`}
                    />
                    <text x={mx} y={(y1 + y2) / 2 - 6} textAnchor="middle">
                      {edge.label || edge.type}
                    </text>
                  </g>
                );
              })}

              {items.map(item => {
                const pos = layout.node_positions[item.id] ?? { x: 0, y: 0 };
                const st = styleFor(item.type);
                const canConnect = item.kind !== 'rule';
                return (
                  <foreignObject
                    key={item.id}
                    x={pos.x}
                    y={pos.y}
                    width={NODE_W}
                    height={NODE_H}
                  >
                    <div
                      data-node={item.id}
                      data-kind={item.kind}
                      className={`forma-vme-node${selected === item.id ? ' selected' : ''}`}
                      style={
                        {
                          '--vme-color': st.color,
                          '--vme-bg': st.background,
                        } as CSSProperties
                      }
                      onDoubleClick={e => {
                        if (readOnly || item.kind !== 'node') return;
                        e.stopPropagation();
                        setSelected(item.id);
                        setFocusName(true);
                        setHint('编辑名称 · 仅缓冲（未保存不产生修订）');
                      }}
                    >
                      <span>
                        <small>{st.label}</small>
                        <b>{item.name}</b>
                        <em>{item.description || item.id}</em>
                      </span>
                      {canConnect && !readOnly && (
                        <div
                          data-handle={item.id}
                          style={{
                            marginLeft: 'auto',
                            width: 10,
                            height: 10,
                            borderRadius: '50%',
                            background: st.color,
                            flexShrink: 0,
                          }}
                          title="拖出以建立关系"
                        />
                      )}
                    </div>
                  </foreignObject>
                );
              })}
            </g>
          </svg>

          {!items.length && (
            <div
              style={{
                position: 'absolute',
                inset: '40% 10%',
                textAlign: 'center',
                color: '#8b93a3',
              }}
            >
              空模型 — 从工具栏添加节点
            </div>
          )}
        </div>

        <aside className="forma-vme-props">
          <PropertyPanel
            readOnly={readOnly}
            nameInputRef={nameInputRef}
            model={model}
            selectedNode={selectedNode}
            selectedEdge={selectedEdge}
            selectedState={selectedState}
            selectedRule={selectedRule}
            onUpdateNode={node => {
              setSemantic(
                {
                  ...model,
                  nodes: model.nodes.map(n => (n.id === node.id ? node : n)),
                },
                '更新节点属性',
              );
            }}
            onUpdateEdge={(edge, note) => {
              if (!edge.label?.trim()) {
                setHint('关系标签不能为空');
                return;
              }
              setSemantic(
                {
                  ...model,
                  edges: model.edges.map(e => (e.id === edge.id ? edge : e)),
                },
                note ?? '更新关系属性',
              );
            }}
            onUpdateState={state => {
              setSemantic(
                {
                  ...model,
                  states: model.states.map(s => (s.id === state.id ? state : s)),
                },
                '更新状态属性',
              );
            }}
            onUpdateRule={rule => {
              setSemantic(
                {
                  ...model,
                  rules: model.rules.map(r => (r.id === rule.id ? rule : r)),
                },
                '更新规则属性',
              );
            }}
            onHint={setHint}
            onDelete={removeSelected}
          />
        </aside>
      </div>

      <div className="forma-vme-hint">
        <span>{hint}</span>
        <span>{Math.round(layout.zoom * 100)}%</span>
      </div>

      <footer className="forma-vme-footer">
        <div className="forma-vme-legend">
          {(['ACTOR', 'BUSINESS_OBJECT', 'PROCESS', 'EVENT', 'STATE', 'RULE'] as const).map(
            t => {
              const st = styleFor(t);
              return (
                <span key={t}>
                  <i style={{ background: st.color }} />
                  {st.label}
                </span>
              );
            },
          )}
        </div>
      </footer>
    </section>
  );
}

function SourceMarkerBadge({ marker }: { marker?: string }) {
  const ai = marker === 'AI_GENERATED';
  return (
    <div className={`forma-vme-source ${ai ? 'ai' : 'manual'}`}>
      Source: {marker || 'MANUAL_MODIFIED'}
    </div>
  );
}

function PropertyPanel({
  readOnly,
  nameInputRef,
  model,
  selectedNode,
  selectedEdge,
  selectedState,
  selectedRule,
  onUpdateNode,
  onUpdateEdge,
  onUpdateState,
  onUpdateRule,
  onHint,
  onDelete,
}: {
  readOnly?: boolean;
  nameInputRef?: RefObject<HTMLInputElement>;
  model: FormaSemanticModel;
  selectedNode?: FormaSemanticNode;
  selectedEdge?: FormaSemanticEdge;
  selectedState?: FormaBusinessState;
  selectedRule?: FormaBusinessRule;
  onUpdateNode: (n: FormaSemanticNode) => void;
  onUpdateEdge: (e: FormaSemanticEdge, note?: string) => void;
  onUpdateState: (s: FormaBusinessState) => void;
  onUpdateRule: (r: FormaBusinessRule) => void;
  onHint: (msg: string) => void;
  onDelete: () => void;
}) {
  if (selectedNode) {
    return (
      <form
        onSubmit={e => e.preventDefault()}
        onBlur={e => {
          if (readOnly) return;
          const form = e.currentTarget;
          const nameInput = form.elements.namedItem('name') as HTMLInputElement;
          const name = nameInput.value;
          const type = (form.elements.namedItem('type') as HTMLSelectElement).value;
          const description = (form.elements.namedItem('description') as HTMLTextAreaElement)
            .value;
          if (!name.trim()) {
            onHint('节点名称不能为空');
            nameInput.value = selectedNode.name;
            return;
          }
          if (
            name !== selectedNode.name ||
            type !== selectedNode.type ||
            description !== (selectedNode.description || '')
          ) {
            onUpdateNode({ ...selectedNode, name: name.trim(), type, description });
          }
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between' }}>
          <strong>节点属性</strong>
          <small>{selectedNode.id}</small>
        </div>
        <SourceMarkerBadge marker={selectedNode.source_marker} />
        <label>
          名称
          <input
            name="name"
            ref={nameInputRef}
            defaultValue={selectedNode.name}
            key={selectedNode.id + '-n'}
            disabled={readOnly}
            data-testid="node-name-input"
            required
          />
        </label>
        <label>
          类型
          <select
            name="type"
            defaultValue={selectedNode.type}
            key={selectedNode.id + '-t'}
            disabled={readOnly}
          >
            {ADDABLE_NODE_TYPES.map(t => (
              <option key={t} value={t}>
                {styleFor(t).label}
              </option>
            ))}
          </select>
        </label>
        <label>
          描述
          <textarea
            name="description"
            rows={3}
            defaultValue={selectedNode.description || ''}
            key={selectedNode.id + '-d'}
            disabled={readOnly}
          />
        </label>
        {!readOnly && (
          <button type="button" className="forma-vme-btn danger" onClick={onDelete}>
            删除
          </button>
        )}
      </form>
    );
  }

  if (selectedEdge) {
    return (
      <form
        onSubmit={e => e.preventDefault()}
        onBlur={e => {
          if (readOnly) return;
          const form = e.currentTarget;
          const labelInput = form.elements.namedItem('label') as HTMLInputElement;
          const label = labelInput.value;
          const type = (form.elements.namedItem('type') as HTMLSelectElement).value;
          if (!label.trim()) {
            onHint('关系标签不能为空');
            labelInput.value = selectedEdge.label || '关联';
            return;
          }
          if (label !== (selectedEdge.label || '') || type !== selectedEdge.type) {
            onUpdateEdge({ ...selectedEdge, label: label.trim(), type });
          }
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between' }}>
          <strong>关系属性</strong>
          <small>{selectedEdge.id}</small>
        </div>
        <SourceMarkerBadge marker={selectedEdge.source_marker} />
        <label>
          标签（必填）
          <input
            name="label"
            defaultValue={selectedEdge.label || ''}
            key={selectedEdge.id + '-l'}
            required
            disabled={readOnly}
            data-testid="edge-label-input"
          />
        </label>
        <label>
          类型
          <select
            name="type"
            defaultValue={selectedEdge.type}
            key={selectedEdge.id + '-t'}
            disabled={readOnly}
            data-testid="edge-type-select"
          >
            {EDGE_TYPES.map(t => (
              <option key={t} value={t}>
                {t}
              </option>
            ))}
          </select>
        </label>
        <p style={{ color: '#758196' }}>
          {selectedEdge.source} → {selectedEdge.target}
        </p>
        {!readOnly && (
          <button type="button" className="forma-vme-btn danger" onClick={onDelete}>
            删除
          </button>
        )}
      </form>
    );
  }

  if (selectedState) {
    const refValid = model.nodes.some(n => n.id === selectedState.object_ref);
    return (
      <form
        onSubmit={e => e.preventDefault()}
        onBlur={e => {
          if (readOnly) return;
          const form = e.currentTarget;
          const nameInput = form.elements.namedItem('name') as HTMLInputElement;
          const name = nameInput.value;
          const object_ref = (form.elements.namedItem('object_ref') as HTMLSelectElement).value;
          if (!name.trim()) {
            onHint('状态名称不能为空');
            nameInput.value = selectedState.name;
            return;
          }
          if (!object_ref || !model.nodes.some(n => n.id === object_ref)) {
            onHint('请选择合法的业务元素作为状态对象');
            return;
          }
          if (name !== selectedState.name || object_ref !== selectedState.object_ref) {
            onUpdateState({ ...selectedState, name: name.trim(), object_ref });
          }
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between' }}>
          <strong>状态属性</strong>
          <small>{selectedState.id}</small>
        </div>
        <SourceMarkerBadge marker={selectedState.source_marker} />
        <label>
          名称
          <input
            name="name"
            defaultValue={selectedState.name}
            key={selectedState.id + '-n'}
            disabled={readOnly}
            data-testid="state-name-input"
            required
          />
        </label>
        <label>
          对象
          {readOnly && !refValid ? (
            <p data-testid="invalid-object-ref" style={{ color: '#c0392b' }}>
              Invalid reference: {selectedState.object_ref || '(empty)'}
            </p>
          ) : (
            <select
              name="object_ref"
              defaultValue={refValid ? selectedState.object_ref : model.nodes[0]?.id}
              key={selectedState.id + '-o'}
              disabled={readOnly}
              data-testid="state-object-ref-select"
            >
              {model.nodes.map(n => (
                <option key={n.id} value={n.id}>
                  {n.name}
                </option>
              ))}
            </select>
          )}
        </label>
        {!readOnly && (
          <button type="button" className="forma-vme-btn danger" onClick={onDelete}>
            删除
          </button>
        )}
      </form>
    );
  }

  if (selectedRule) {
    const appliesCandidates = [
      ...model.nodes.map(n => ({
        id: n.id,
        label: `${n.name} · ${styleFor(n.type).label}`,
        kind: 'node' as const,
      })),
      ...model.states.map(s => ({
        id: s.id,
        label: `${s.name} · 状态`,
        kind: 'state' as const,
      })),
    ];
    return (
      <form
        onSubmit={e => e.preventDefault()}
        onBlur={e => {
          if (readOnly) return;
          const form = e.currentTarget;
          const nameInput = form.elements.namedItem('name') as HTMLInputElement;
          const name = nameInput.value;
          const expression = (form.elements.namedItem('expression') as HTMLTextAreaElement)
            .value;
          const description = (form.elements.namedItem('description') as HTMLTextAreaElement)
            .value;
          const checks = form.querySelectorAll<HTMLInputElement>(
            'input[name="applies_to"]:checked',
          );
          const applies = [...checks].map(c => c.value);
          if (!name.trim()) {
            onHint('规则名称不能为空');
            nameInput.value = selectedRule.name;
            return;
          }
          if (
            name !== selectedRule.name ||
            expression !== (selectedRule.expression || '') ||
            description !== (selectedRule.description || '') ||
            JSON.stringify(applies) !== JSON.stringify(selectedRule.applies_to ?? [])
          ) {
            onUpdateRule({
              ...selectedRule,
              name: name.trim(),
              expression,
              description,
              applies_to: applies,
            });
          }
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between' }}>
          <strong>规则属性</strong>
          <small>{selectedRule.id}</small>
        </div>
        <SourceMarkerBadge marker={selectedRule.source_marker} />
        <label>
          名称
          <input
            name="name"
            defaultValue={selectedRule.name}
            key={selectedRule.id + '-n'}
            disabled={readOnly}
            data-testid="rule-name-input"
            required
          />
        </label>
        <fieldset data-testid="rule-applies-to" style={{ border: '1px solid #e4e8f0', padding: 8 }}>
          <legend>适用于</legend>
          {appliesCandidates.length === 0 && (
            <p style={{ color: '#758196' }}>暂无可选业务元素/状态</p>
          )}
          {appliesCandidates.map(c => (
            <label
              key={c.id}
              style={{ display: 'flex', gap: 8, alignItems: 'center', marginBottom: 4 }}
            >
              <input
                type="checkbox"
                name="applies_to"
                value={c.id}
                defaultChecked={(selectedRule.applies_to ?? []).includes(c.id)}
                disabled={readOnly}
              />
              <span>{c.label}</span>
            </label>
          ))}
        </fieldset>
        <label>
          表达式
          <textarea
            name="expression"
            rows={3}
            defaultValue={selectedRule.expression || ''}
            key={selectedRule.id + '-e'}
            disabled={readOnly}
          />
        </label>
        <label>
          描述
          <textarea
            name="description"
            rows={2}
            defaultValue={selectedRule.description || ''}
            key={selectedRule.id + '-d'}
            disabled={readOnly}
          />
        </label>
        {!readOnly && (
          <button type="button" className="forma-vme-btn danger" onClick={onDelete}>
            删除
          </button>
        )}
      </form>
    );
  }

  return (
    <div>
      <strong>属性面板</strong>
      <p style={{ color: '#758196', lineHeight: 1.7 }}>
        选中节点、关系、状态或规则以编辑。Source Marker（AI_GENERATED /
        MANUAL_MODIFIED）。规则通过 applies_to 关联，不是关系端点。
      </p>
    </div>
  );
}
