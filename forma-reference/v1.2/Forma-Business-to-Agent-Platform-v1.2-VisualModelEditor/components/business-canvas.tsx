'use client';
import {
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react';
import {
  UserRound,
  Box,
  GitBranch,
  CircleCheck,
  ShieldCheck,
  Server,
  Bot,
  Layers,
  MousePointer2,
  Hand,
  Maximize,
  Minimize,
  Download,
  Plus,
  Minus,
  Scan,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  layoutGraph,
  nodeTypes,
  type BusinessGraph,
  type BusinessNode,
} from '@/lib/business-canvas';
import './business-canvas.css';

const icons = {
  role: UserRound,
  entity: Box,
  process: GitBranch,
  state: CircleCheck,
  rule: ShieldCheck,
  external: Server,
  agent: Bot,
  application: Layers,
};
export function BusinessCanvas({
  graph,
  title = 'Business Canvas',
  compact = false,
  onSelect,
  onNavigate,
  actions,
}: {
  graph: BusinessGraph;
  title?: string;
  compact?: boolean;
  onSelect?: (node: BusinessNode) => void;
  onNavigate?: (route: string) => void;
  actions?: ReactNode;
}) {
  const root = useRef<HTMLElement>(null);
  const surface = useRef<HTMLDivElement>(null);
  const id = useId().replace(/:/g, '');
  const nodes = useMemo(() => layoutGraph(graph), [graph]);
  const [size, setSize] = useState({ width: 900, height: 560 });
  const [zoom, setZoom] = useState(1);
  const [pan, setPan] = useState({ x: 0, y: 0 });
  const [tool, setTool] = useState<'select' | 'pan'>('select');
  const [selected, setSelected] = useState<string | null>(null);
  const [message, setMessage] = useState('');
  const [fullscreen, setFullscreen] = useState(false);
  const drag = useRef<{ x: number; y: number; px: number; py: number } | null>(
    null,
  );
  const bounds = {
    left: Math.min(0, ...nodes.map((n) => n.position.x)),
    top: Math.min(0, ...nodes.map((n) => n.position.y)),
    right: Math.max(190, ...nodes.map((n) => n.position.x + 190)),
    bottom: Math.max(88, ...nodes.map((n) => n.position.y + 88)),
  };
  const fit = Math.min(
    1.15,
    (size.width - 110) / (bounds.right - bounds.left),
    (size.height - 64) / (bounds.bottom - bounds.top),
  );
  const scale = Math.max(0.05, fit) * zoom;
  const tx = (size.width - (bounds.right + bounds.left) * scale) / 2 + pan.x;
  const ty = (size.height - (bounds.bottom + bounds.top) * scale) / 2 + pan.y;
  const active = nodes.find((n) => n.id === selected);
  useEffect(() => {
    if (!surface.current) return;
    const observer = new ResizeObserver(([entry]) =>
      setSize({
        width: entry.contentRect.width,
        height: entry.contentRect.height,
      }),
    );
    observer.observe(surface.current);
    return () => observer.disconnect();
  }, []);
  useEffect(() => {
    const sync = () =>
      setFullscreen(document.fullscreenElement === root.current);
    document.addEventListener('fullscreenchange', sync);
    return () => document.removeEventListener('fullscreenchange', sync);
  }, []);
  const reset = () => {
    setZoom(1);
    setPan({ x: 0, y: 0 });
  };
  return (
    <section
      ref={root}
      className={`business-canvas ${compact ? 'bc-compact' : ''}`}
      aria-label={title}
    >
      <header className="bc-header">
        <div>
          <span className="bc-live-dot" />
          <b>{title}</b>
          <small>
            {nodes.length} 节点 · {graph.edges.length} 关系
          </small>
        </div>
        <div className="bc-actions">
          {actions}
          <Button
            size="sm"
            variant="outline"
            onClick={() =>
              setMessage(
                '导出入口已预留：PNG / SVG / JSON；当前原型尚未生成文件。',
              )
            }
          >
            <Download size={14} />
            导出
          </Button>
          <Button
            size="sm"
            variant="outline"
            aria-label={fullscreen ? '退出全屏' : '全屏'}
            onClick={async () => {
              try {
                if (document.fullscreenElement === root.current)
                  await document.exitFullscreen();
                else await root.current?.requestFullscreen();
              } catch {
                setMessage('当前环境不支持全屏，请使用浏览器窗口最大化。');
              }
            }}
          >
            {fullscreen ? <Minimize size={15} /> : <Maximize size={15} />}
          </Button>
        </div>
      </header>
      <div className="bc-surface" ref={surface}>
        <div className="bc-tools" role="toolbar" aria-label="画布工具">
          <Button
            variant="ghost"
            size="icon"
            aria-label="选择节点"
            aria-pressed={tool === 'select'}
            onClick={() => setTool('select')}
          >
            <MousePointer2 size={17} />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            aria-label="平移画布"
            aria-pressed={tool === 'pan'}
            onClick={() => setTool('pan')}
          >
            <Hand size={17} />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            aria-label="适配视图"
            onClick={reset}
          >
            <Scan size={17} />
          </Button>
        </div>
        {/* The SVG is a keyboard-operable canvas; semantic buttons inside it remain individually focusable. */}
        {/* oxlint-disable jsx-a11y/no-noninteractive-element-interactions, jsx-a11y/no-noninteractive-tabindex */}
        <svg
          role="application"
          tabIndex={0}
          onKeyDown={(e) => {
            if (e.target !== e.currentTarget) return;
            if (e.key === 'Escape') setSelected(null);
            if (e.key === '0') reset();
            if (
              ['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown'].includes(
                e.key,
              )
            ) {
              e.preventDefault();
              setPan((p) => ({
                x:
                  p.x +
                  (e.key === 'ArrowLeft'
                    ? 30
                    : e.key === 'ArrowRight'
                      ? -30
                      : 0),
                y:
                  p.y +
                  (e.key === 'ArrowUp' ? 30 : e.key === 'ArrowDown' ? -30 : 0),
              }));
            }
          }}
          width="100%"
          height="100%"
          className={tool === 'pan' ? 'bc-panning' : ''}
          aria-label={`${title} 节点与带标签关系`}
          onPointerDown={(e) => {
            if (tool !== 'pan') return;
            drag.current = { x: e.clientX, y: e.clientY, px: pan.x, py: pan.y };
            e.currentTarget.setPointerCapture(e.pointerId);
          }}
          onPointerMove={(e) => {
            if (drag.current)
              setPan({
                x: drag.current.px + e.clientX - drag.current.x,
                y: drag.current.py + e.clientY - drag.current.y,
              });
          }}
          onPointerUp={() => {
            drag.current = null;
          }}
          onPointerCancel={() => {
            drag.current = null;
          }}
          onClick={(e) => {
            if (e.target === e.currentTarget) setSelected(null);
          }}
        >
          <defs>
            <marker
              id={`${id}-arrow`}
              viewBox="0 0 10 10"
              refX="9"
              refY="5"
              markerWidth="7"
              markerHeight="7"
              orient="auto-start-reverse"
            >
              <path
                d="M 1 1 L 9 5 L 1 9"
                fill="none"
                stroke="#98a2b3"
                strokeWidth="1.3"
              />
            </marker>
          </defs>
          <g transform={`translate(${tx},${ty}) scale(${scale})`}>
            {graph.edges.map((e) => {
              const a = nodes.find((n) => n.id === e.source),
                b = nodes.find((n) => n.id === e.target);
              if (!a || !b) return null;
              const horizontal =
                Math.abs(b.position.x - a.position.x) >=
                Math.abs(b.position.y - a.position.y);
              const sign = horizontal
                ? Math.sign(b.position.x - a.position.x) || 1
                : Math.sign(b.position.y - a.position.y) || 1;
              const x1 = a.position.x + 95 + (horizontal ? sign * 95 : 0),
                y1 = a.position.y + 44 + (horizontal ? 0 : sign * 44);
              const x2 = b.position.x + 95 - (horizontal ? sign * 95 : 0),
                y2 = b.position.y + 44 - (horizontal ? 0 : sign * 44);
              const mx = (x1 + x2) / 2,
                my = (y1 + y2) / 2;
              const path = horizontal
                ? `M${x1},${y1} C${mx},${y1} ${mx},${y2} ${x2},${y2}`
                : `M${x1},${y1} C${x1},${my} ${x2},${my} ${x2},${y2}`;
              return (
                <g key={e.id} className="bc-edge">
                  <path
                    d={path}
                    fill="none"
                    stroke="#a7afbd"
                    strokeWidth="1.25"
                    strokeDasharray={e.dashed ? '5 4' : undefined}
                    markerEnd={`url(#${id}-arrow)`}
                  />
                  <text x={mx} y={my - 8} textAnchor="middle">
                    {e.label}
                  </text>
                </g>
              );
            })}
            {nodes.map((n) => {
              const style = nodeTypes[n.type];
              const Icon = icons[n.type];
              return (
                <foreignObject
                  key={n.id}
                  x={n.position.x}
                  y={n.position.y}
                  width="190"
                  height="88"
                >
                  <button
                    className={`bc-node ${selected === n.id ? 'bc-selected' : ''}`}
                    style={
                      {
                        '--bc-color': style.color,
                        '--bc-bg': style.background,
                      } as React.CSSProperties
                    }
                    aria-pressed={selected === n.id}
                    aria-label={`${style.label}：${n.label}`}
                    title={n.description}
                    onClick={() => {
                      if (tool === 'pan') return;
                      setSelected(n.id);
                      onSelect?.(n);
                    }}
                  >
                    <Icon size={22} />
                    <span>
                      <small>{style.label}</small>
                      <b>{n.label}</b>
                      <em>{n.description}</em>
                    </span>
                  </button>
                </foreignObject>
              );
            })}
          </g>
        </svg>
        {/* oxlint-enable jsx-a11y/no-noninteractive-element-interactions, jsx-a11y/no-noninteractive-tabindex */}
        {!nodes.length && (
          <p className="bc-empty">暂无业务模型，请先添加业务事实。</p>
        )}
      </div>
      <footer className="bc-footer">
        <div className="bc-legend">
          {Object.entries(nodeTypes)
            .filter(([type]) => nodes.some((n) => n.type === type))
            .map(([type, style]) => (
              <span key={type}>
                <i style={{ background: style.color }} />
                {style.label}
              </span>
            ))}
        </div>
        <div className="bc-zoom">
          <Button
            size="icon"
            variant="ghost"
            aria-label="缩小"
            disabled={zoom <= 0.4}
            onClick={() => setZoom((z) => Math.max(0.4, z - 0.2))}
          >
            <Minus size={15} />
          </Button>
          <span aria-live="polite">{Math.round(scale * 100)}%</span>
          <Button
            size="icon"
            variant="ghost"
            aria-label="放大"
            disabled={zoom >= 3}
            onClick={() => setZoom((z) => Math.min(3, z + 0.2))}
          >
            <Plus size={15} />
          </Button>
          <Button
            size="icon"
            variant="ghost"
            aria-label="重置并适配视图"
            onClick={reset}
          >
            <Scan size={15} />
          </Button>
        </div>
      </footer>
      {active && (
        <div className="bc-detail">
          <span style={{ color: nodeTypes[active.type].color }}>
            {nodeTypes[active.type].label}
          </span>
          <b>{active.label}</b>
          <p>{active.description}</p>
          {active.route && onNavigate && (
            <Button
              size="sm"
              variant="outline"
              onClick={() => onNavigate(active.route!)}
            >
              打开资产 →
            </Button>
          )}
          <Button size="sm" variant="ghost" onClick={() => setSelected(null)}>
            取消选中
          </Button>
        </div>
      )}
      {message && (
        <output className="bc-message">
          {message}
          <Button size="sm" variant="ghost" onClick={() => setMessage('')}>
            关闭
          </Button>
        </output>
      )}
    </section>
  );
}
