'use client';
// SVG graph nodes/edges use keyboard-operable roles; native HTML buttons cannot replace SVG groups.
/* oxlint-disable jsx-a11y/no-noninteractive-element-interactions, jsx-a11y/no-noninteractive-tabindex, jsx-a11y/prefer-tag-over-role */
import { useEffect, useId, useRef, useState } from 'react';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Textarea } from './ui/textarea';
import { nodeTypes, type BusinessNodeType } from '@/lib/business-canvas';
import {
  arrangeModel,
  commitModel,
  deleteNode,
  editSemantic,
  historyModel,
  type SemanticNode,
  type SemanticEdge,
  type VisualModel,
  type ViewLayout,
} from '@/lib/visual-model';
import './business-canvas.css';
import './visual-model-editor.css';

const types: BusinessNodeType[] = [
  'role',
  'entity',
  'process',
  'state',
  'rule',
  'external',
];
type Props = {
  value: VisualModel;
  onChange: (value: VisualModel, action: string) => void;
  onImpact: () => void;
};
export function VisualBusinessModelEditor({
  value,
  onChange,
  onImpact,
}: Props) {
  const root = useRef<HTMLElement>(null),
    surface = useRef<HTMLDivElement>(null);
  const marker = useId().replace(/:/g, '');
  const [selected, setSelected] = useState<string | null>(null);
  const [tool, setTool] = useState<'select' | 'pan'>('select');
  const [connection, setConnection] = useState<string | null>(null);
  const [preview, setPreview] = useState<ViewLayout | null>(null);
  const [message, setMessage] = useState(
    '拖拽调整布局；双击节点编辑；拖动右侧圆点或依次点击两个节点建立关系。',
  );
  const [fullscreen, setFullscreen] = useState(false);
  const drag = useRef<{
    kind: 'node' | 'pan' | 'edge';
    id?: string;
    x: number;
    y: number;
    layout: ViewLayout;
    moved: boolean;
  } | null>(null);
  const layout = preview ?? value.view_layout;
  const nodes = value.semantic_model.nodes,
    edges = value.semantic_model.edges;
  const node = nodes.find((n) => n.id === selected),
    edge = edges.find((e) => e.id === selected);
  useEffect(() => {
    const sync = () =>
      setFullscreen(document.fullscreenElement === root.current);
    document.addEventListener('fullscreenchange', sync);
    return () => document.removeEventListener('fullscreenchange', sync);
  }, []);
  const change = (next: VisualModel, action: string) => {
    if (next !== value) onChange(next, action);
  };
  const changeLayout = (next: ViewLayout, action: string) =>
    change(commitModel(value, { ...value, view_layout: next }, action), action);
  const fit = () => {
    const points = Object.values(value.view_layout.node_positions);
    const left = Math.min(0, ...points.map((p) => p.x)),
      top = Math.min(0, ...points.map((p) => p.y));
    const right = Math.max(190, ...points.map((p) => p.x + 190)),
      bottom = Math.max(88, ...points.map((p) => p.y + 88));
    const width = surface.current?.clientWidth ?? 900,
      height = surface.current?.clientHeight ?? 560;
    const zoom = Math.max(
      0.05,
      Math.min(
        1.5,
        (width - 80) / (right - left),
        (height - 80) / (bottom - top),
      ),
    );
    changeLayout(
      {
        ...value.view_layout,
        zoom,
        viewport: {
          x: (width - (right + left) * zoom) / 2,
          y: (height - (bottom + top) * zoom) / 2,
        },
      },
      '适配视图',
    );
  };
  const connect = (from: string, target: string) => {
    if (from === target) {
      setMessage('请选择另一个节点作为关系终点。');
      return;
    }
    const id = crypto.randomUUID();
    change(
      editSemantic(
        value,
        {
          ...value.semantic_model,
          edges: [
            ...edges,
            { id, from, target, label: '关联', source: 'manual_modified' },
          ],
        },
        '创建关系',
      ),
      '创建关系 · 进入 Impact Analysis',
    );
    setSelected(id);
    setConnection(null);
  };
  const remove = () => {
    if (node)
      change(
        deleteNode(value, node.id),
        '删除节点及关联关系 · 进入 Impact Analysis',
      );
    else if (edge)
      change(
        editSemantic(
          value,
          {
            ...value.semantic_model,
            edges: edges.filter((e) => e.id !== edge.id),
          },
          '删除关系',
        ),
        '删除关系 · 进入 Impact Analysis',
      );
    setSelected(null);
  };
  const origin = (source: string) =>
    source === 'AI_generated' ? 'AI 生成' : '人工修改';
  return (
    <section
      className="business-canvas vme"
      ref={root}
      aria-label="Visual Business Model Editor"
      onKeyDown={(e) => {
        if ((e.target as HTMLElement).closest('input,textarea,select')) return;
        if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'z') {
          e.preventDefault();
          change(
            historyModel(value, e.shiftKey ? 'redo' : 'undo'),
            e.shiftKey ? '重做' : '撤销',
          );
        }
        if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'y') {
          e.preventDefault();
          change(historyModel(value, 'redo'), '重做');
        }
        if (e.key === 'Escape') {
          setConnection(null);
          setSelected(null);
          setPreview(null);
          drag.current = null;
        }
        if (e.key === 'Delete' || e.key === 'Backspace') {
          e.preventDefault();
          remove();
        }
      }}
    >
      <header className="bc-header">
        <div>
          <span className="bc-live-dot" />
          <b>Visual Business Model Editor</b>
          <small>
            BM-001 · r{value.revision} · {nodes.length} 节点 / {edges.length}{' '}
            关系
          </small>
        </div>
        <div className="bc-actions">
          <Button
            size="sm"
            variant="outline"
            onClick={() => {
              const next = {
                ...value,
                saved_layout: structuredClone(value.view_layout),
              };
              onChange(next, '保存当前布局');
              setMessage(
                '布局保存请求已提交；若本机存储失败，平台会显示警告。',
              );
            }}
          >
            保存当前布局
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={async () => {
              try {
                if (document.fullscreenElement) await document.exitFullscreen();
                else await root.current?.requestFullscreen();
              } catch {
                setMessage('当前环境不支持全屏，请最大化浏览器窗口。');
              }
            }}
          >
            {fullscreen ? '退出全屏' : '全屏'}
          </Button>
        </div>
      </header>
      <div className="vme-toolbar" role="toolbar" aria-label="模型编辑工具">
        <Button
          size="sm"
          variant="outline"
          disabled={!value.past.length}
          onClick={() => change(historyModel(value, 'undo'), '撤销')}
        >
          撤销
        </Button>
        <Button
          size="sm"
          variant="outline"
          disabled={!value.future.length}
          onClick={() => change(historyModel(value, 'redo'), '重做')}
        >
          重做
        </Button>
        <span className="vme-separator" />
        {types.map((type) => (
          <Button
            key={type}
            size="sm"
            variant="ghost"
            onClick={() => {
              const id = crypto.randomUUID();
              change(
                editSemantic(
                  value,
                  {
                    ...value.semantic_model,
                    nodes: [
                      ...nodes,
                      {
                        id,
                        type,
                        label: `新${nodeTypes[type].label}`,
                        description: '',
                        source: 'manual_modified',
                      },
                    ],
                  },
                  '新增节点',
                ),
                '新增节点 · 进入 Impact Analysis',
              );
              setSelected(id);
            }}
          >
            ＋{nodeTypes[type].label}
          </Button>
        ))}
        <span className="vme-separator" />
        <Button
          size="sm"
          variant="outline"
          onClick={() => change(arrangeModel(value), '自动布局 · 仅更新布局')}
        >
          自动布局
        </Button>
        <Button
          size="sm"
          variant="outline"
          aria-pressed={layout.mode === 'manual'}
          onClick={() =>
            changeLayout({ ...layout, mode: 'manual' }, '切换手动布局')
          }
        >
          手动布局
        </Button>
        <Button
          size="sm"
          variant="outline"
          onClick={() => {
            change(arrangeModel(value), 'AI 重新布局（本地模拟）· 仅更新布局');
            setMessage(
              'AI 重新布局为本地确定性布局模拟，未连接 LLM；业务语义保持不变。',
            );
          }}
        >
          ✦ AI 重新布局
        </Button>
      </div>
      <div className="vme-body">
        <div className="bc-surface" ref={surface}>
          <div className="bc-tools">
            <Button
              size="sm"
              variant="ghost"
              aria-pressed={tool === 'select'}
              onClick={() => setTool('select')}
            >
              选择
            </Button>
            <Button
              size="sm"
              variant="ghost"
              aria-pressed={tool === 'pan'}
              onClick={() => setTool('pan')}
            >
              平移
            </Button>
            <Button size="sm" variant="ghost" onClick={fit}>
              适配
            </Button>
          </div>
          <svg
            width="100%"
            height="100%"
            role="application"
            aria-label="可编辑业务模型画布"
            tabIndex={0}
            onPointerDown={(e) => {
              if (e.button !== 0) return;
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
                layout: structuredClone(value.view_layout),
                moved: false,
              };
              if (kind === 'edge' && id) {
                setConnection(id);
                setMessage('松开到目标节点，或点击目标节点建立关系。');
              }
              e.currentTarget.setPointerCapture(e.pointerId);
            }}
            onPointerMove={(e) => {
              const d = drag.current;
              if (!d) return;
              const dx = e.clientX - d.x,
                dy = e.clientY - d.y;
              if (Math.abs(dx) + Math.abs(dy) < 3) return;
              d.moved = true;
              if (d.kind === 'pan')
                setPreview({
                  ...d.layout,
                  viewport: {
                    x: d.layout.viewport.x + dx,
                    y: d.layout.viewport.y + dy,
                  },
                });
              if (d.kind === 'node' && d.id) {
                const p = d.layout.node_positions[d.id];
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
            onPointerUp={(e) => {
              const d = drag.current;
              if (!d) return;
              const target = document
                .elementFromPoint(e.clientX, e.clientY)
                ?.closest('[data-node]')
                ?.getAttribute('data-node');
              if (d.kind === 'edge' && d.id && target && target !== d.id)
                connect(d.id, target);
              else if (d.moved && preview)
                changeLayout(
                  preview,
                  d.kind === 'node' ? '移动节点 · 仅更新布局' : '平移画布',
                );
              else if (!d.moved && d.kind === 'node' && d.id) {
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
              {edges.map((e) => {
                const a = layout.node_positions[e.from],
                  b = layout.node_positions[e.target];
                if (!a || !b) return null;
                const x1 = a.x + 190,
                  y1 = a.y + 44,
                  x2 = b.x,
                  y2 = b.y + 44,
                  mx = (x1 + x2) / 2;
                const path = `M${x1},${y1} C${mx},${y1} ${mx},${y2} ${x2},${y2}`;
                return (
                  <g
                    key={e.id}
                    data-edge={e.id}
                    className="bc-edge"
                    role="button"
                    tabIndex={0}
                    aria-label={`关系 ${e.label}`}
                    onClick={() => setSelected(e.id)}
                    onKeyDown={(event) => {
                      if (event.key === 'Enter') setSelected(e.id);
                    }}
                  >
                    <path
                      d={path}
                      stroke="transparent"
                      strokeWidth="18"
                      fill="none"
                      style={{ cursor: 'pointer' }}
                    />
                    <path
                      d={path}
                      stroke={selected === e.id ? '#3478f6' : '#a7afbd'}
                      strokeWidth={selected === e.id ? 2.5 : 1.5}
                      fill="none"
                      markerEnd={`url(#${marker})`}
                      strokeDasharray={e.dashed ? '5 4' : undefined}
                    />
                    <text x={mx} y={(y1 + y2) / 2 - 9} textAnchor="middle">
                      {e.label}
                    </text>
                  </g>
                );
              })}
              {nodes.map((n) => {
                const p = layout.node_positions[n.id] ?? { x: 0, y: 0 },
                  style = nodeTypes[n.type];
                return (
                  <g
                    key={n.id}
                    data-node={n.id}
                    transform={`translate(${p.x},${p.y})`}
                    tabIndex={0}
                    role="button"
                    aria-label={`节点 ${n.label}`}
                    onDoubleClick={() => {
                      setSelected(n.id);
                      setTimeout(
                        () =>
                          root.current
                            ?.querySelector<HTMLInputElement>(
                              '[aria-label="节点名称"]',
                            )
                            ?.focus(),
                        0,
                      );
                    }}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') {
                        if (connection) connect(connection, n.id);
                        else setSelected(n.id);
                      }
                    }}
                    style={{ cursor: tool === 'pan' ? 'grab' : 'move' }}
                  >
                    <rect
                      width="190"
                      height="88"
                      rx="12"
                      fill={style.background}
                      stroke={
                        selected === n.id || connection === n.id
                          ? style.color
                          : `${style.color}66`
                      }
                      strokeWidth={selected === n.id ? 2.5 : 1}
                    />
                    <text x="14" y="20" fill={style.color} fontSize="10">
                      {style.label} · {origin(n.source)}
                    </text>
                    <text
                      x="14"
                      y="43"
                      fill="#243045"
                      fontSize="13"
                      fontWeight="600"
                    >
                      {n.label.length > 15
                        ? n.label.slice(0, 15) + '…'
                        : n.label}
                    </text>
                    <text x="14" y="66" fill="#7a8597" fontSize="10">
                      {(n.description ?? '添加业务描述').slice(0, 20)}
                    </text>
                    <circle
                      data-handle="true"
                      cx="190"
                      cy="44"
                      r="7"
                      fill="white"
                      stroke={style.color}
                      strokeWidth="2"
                      style={{ cursor: 'crosshair' }}
                    >
                      <title>从此节点创建关系</title>
                    </circle>
                  </g>
                );
              })}
            </g>
          </svg>
          {!nodes.length && (
            <div className="bc-empty">空模型 · 从上方添加第一个节点</div>
          )}
        </div>
        <aside className="vme-properties" aria-label="属性面板">
          <div className="vme-panel-title">
            <b>{node ? '节点属性' : edge ? '关系属性' : '模型检查器'}</b>
            <small>{layout.mode === 'auto' ? '自动布局' : '手动布局'}</small>
          </div>
          {node ? (
            <NodeForm
              key={`${node.id}-${value.revision}`}
              node={node}
              onSave={(patch) =>
                change(
                  editSemantic(
                    value,
                    {
                      ...value.semantic_model,
                      nodes: nodes.map((n) =>
                        n.id === node.id ? { ...n, ...patch } : n,
                      ),
                    },
                    '编辑节点',
                  ),
                  '编辑节点 · 进入 Impact Analysis',
                )
              }
            />
          ) : edge ? (
            <EdgeForm
              key={`${edge.id}-${value.revision}`}
              edge={edge}
              onSave={(label) =>
                change(
                  editSemantic(
                    value,
                    {
                      ...value.semantic_model,
                      edges: edges.map((e) =>
                        e.id === edge.id ? { ...e, label } : e,
                      ),
                    },
                    '编辑关系语义',
                  ),
                  '编辑关系语义 · 进入 Impact Analysis',
                )
              }
            />
          ) : (
            <p>
              选择节点或关系查看属性。布局随操作保存在本机；“保存当前布局”可创建布局恢复点。
            </p>
          )}
          {(node || edge) && (
            <>
              <p className="vme-source">
                来源：{origin((node ?? edge)!.source)}
                <br />
                <code>source={(node ?? edge)!.source}</code>
              </p>
              {node && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    setConnection(node.id);
                    setMessage('点击目标节点建立关系；Escape 取消。');
                  }}
                >
                  从此节点创建关系
                </Button>
              )}
              <Button
                className="vme-delete"
                variant="outline"
                size="sm"
                onClick={remove}
              >
                {node ? '删除节点及关联关系' : '删除关系'}
              </Button>
            </>
          )}
          {connection && (
            <p className="vme-hint">
              正在连线：{nodes.find((n) => n.id === connection)?.label} →
              请选择目标{' '}
              <Button
                size="sm"
                variant="ghost"
                onClick={() => setConnection(null)}
              >
                取消
              </Button>
            </p>
          )}
          <div className="vme-governance">
            <b>语义与布局分离</b>
            <p>
              拖动 / 缩放 → 仅更新布局
              <br />
              业务修改 → 新 revision + 重新确认与评测
            </p>
            {value.impact && (
              <Button size="sm" variant="outline" onClick={onImpact}>
                Impact Analysis · r{value.impact.revision} 待处理 →
              </Button>
            )}
          </div>
          {value.saved_layout && (
            <Button
              size="sm"
              variant="ghost"
              onClick={() => {
                const saved = value.saved_layout!;
                const node_positions = Object.fromEntries(
                  nodes.map((n) => [
                    n.id,
                    saved.node_positions[n.id] ?? layout.node_positions[n.id],
                  ]),
                );
                changeLayout({ ...saved, node_positions }, '恢复已保存布局');
              }}
            >
              恢复已保存布局
            </Button>
          )}
        </aside>
      </div>
      <footer className="bc-footer">
        <div className="bc-legend">
          {types.map((t) => (
            <span key={t}>
              <i style={{ background: nodeTypes[t].color }} />
              {nodeTypes[t].label}
            </span>
          ))}
        </div>
        <div className="bc-zoom">
          <Button
            size="sm"
            variant="ghost"
            aria-label="缩小"
            onClick={() =>
              changeLayout(
                { ...layout, zoom: Math.max(0.05, layout.zoom / 1.2) },
                '缩小',
              )
            }
          >
            −
          </Button>
          <span>{Math.round(layout.zoom * 100)}%</span>
          <Button
            size="sm"
            variant="ghost"
            aria-label="放大"
            onClick={() =>
              changeLayout(
                { ...layout, zoom: Math.min(3, layout.zoom * 1.2) },
                '放大',
              )
            }
          >
            ＋
          </Button>
          <Button size="sm" variant="ghost" onClick={fit}>
            Fit View
          </Button>
        </div>
      </footer>
      <div className="bc-message" role="status">
        {message}
      </div>
    </section>
  );
}
function NodeForm({
  node,
  onSave,
}: {
  node: SemanticNode;
  onSave: (patch: Pick<SemanticNode, 'label' | 'description' | 'type'>) => void;
}) {
  const formId = useId();
  const [label, setLabel] = useState(node.label),
    [description, setDescription] = useState(node.description ?? ''),
    [type, setType] = useState(node.type);
  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        onSave({ label: label.trim(), description, type });
      }}
    >
      <label htmlFor={`${formId}-name`}>
        名称
        <Input
          id={`${formId}-name`}
          aria-label="节点名称"
          value={label}
          onChange={(e) => setLabel(e.target.value)}
          maxLength={100}
        />
      </label>
      <label htmlFor={`${formId}-description`}>
        描述
        <Textarea
          id={`${formId}-description`}
          aria-label="节点描述"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />
      </label>
      <label>
        节点类型
        <select
          aria-label="节点类型"
          value={type}
          onChange={(e) => setType(e.target.value as BusinessNodeType)}
        >
          {types.map((t) => (
            <option key={t} value={t}>
              {nodeTypes[t].label}
            </option>
          ))}
        </select>
      </label>
      <Button
        size="sm"
        type="submit"
        disabled={
          !label.trim() ||
          (label === node.label &&
            description === (node.description ?? '') &&
            type === node.type)
        }
      >
        保存节点属性
      </Button>
    </form>
  );
}
function EdgeForm({
  edge,
  onSave,
}: {
  edge: SemanticEdge;
  onSave: (label: string) => void;
}) {
  const formId = useId();
  const [label, setLabel] = useState(edge.label);
  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        onSave(label.trim());
      }}
    >
      <label htmlFor={`${formId}-label`}>
        关系语义 / 标签
        <Input
          id={`${formId}-label`}
          aria-label="关系标签"
          value={label}
          onChange={(e) => setLabel(e.target.value)}
          maxLength={100}
        />
      </label>
      <Button
        size="sm"
        type="submit"
        disabled={!label.trim() || label === edge.label}
      >
        保存关系语义
      </Button>
    </form>
  );
}
