import { useCallback, useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';

import type {
  FormaApiClient,
  FormaBusiness,
  FormaBusinessRevision,
  FormaDiffResponse,
  FormaSemanticModel,
  FormaTenant,
  FormaViewLayout,
} from '@forma/api-client';
import { FormaApiError } from '@forma/api-client';

import { adaptModelForPersistence } from '../canonical';
import { VisualModelEditor } from '../components/VisualModelEditor';
import {
  createEditBuffer,
  isLayoutDirty,
  isSemanticDirty,
  resetLayoutBaseline,
  resetSemanticBaseline,
  type EditBuffer,
} from '../edit-buffer';
import { emptyLayout, emptySemanticModel, workOrderDefaultLayout } from '../work-order-seed';
import '../styles/editor.css';

export interface BusinessEditorPageProps {
  client: FormaApiClient;
  currentTenant: FormaTenant | null;
}

export function BusinessEditorPage({ client, currentTenant }: BusinessEditorPageProps) {
  const { businessId = '' } = useParams<{ businessId: string }>();
  const [business, setBusiness] = useState<FormaBusiness | null>(null);
  const [buffer, setBuffer] = useState<EditBuffer | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [savingModel, setSavingModel] = useState(false);
  const [savingLayout, setSavingLayout] = useState(false);
  const [message, setMessage] = useState<string | undefined>();
  const [showRevisions, setShowRevisions] = useState(false);
  const [showDiff, setShowDiff] = useState(false);
  const [revisions, setRevisions] = useState<FormaBusinessRevision[]>([]);
  const [diffFrom, setDiffFrom] = useState(1);
  const [diffTo, setDiffTo] = useState(1);
  const [diffResult, setDiffResult] = useState<FormaDiffResponse | null>(null);
  const [viewingRevision, setViewingRevision] = useState<number | null>(null);
  const [historicalBuffer, setHistoricalBuffer] = useState<EditBuffer | null>(null);
  const [liveBufferSnapshot, setLiveBufferSnapshot] = useState<EditBuffer | null>(null);

  const normalizeSemantic = (semanticRaw: FormaSemanticModel): FormaSemanticModel => ({
    ...semanticRaw,
    nodes: semanticRaw.nodes ?? [],
    edges: semanticRaw.edges ?? [],
    rules: semanticRaw.rules ?? [],
    states: semanticRaw.states ?? [],
  });

  const ensureLayoutPositions = (
    semantic: FormaSemanticModel,
    lay: FormaViewLayout,
  ): FormaViewLayout => {
    const positions = { ...(lay.node_positions ?? {}) };
    let i = 0;
    for (const id of [
      ...semantic.nodes.map(n => n.id),
      ...semantic.states.map(s => s.id),
      ...semantic.rules.map(r => r.id),
    ]) {
      if (!positions[id]) {
        positions[id] = { x: 40 + (i % 4) * 220, y: 40 + Math.floor(i / 4) * 120 };
      }
      i += 1;
    }
    return { ...lay, node_positions: positions };
  };

  const load = useCallback(async () => {
    if (!businessId) return;
    setLoading(true);
    setError(null);
    try {
      const [biz, model, layout] = await Promise.all([
        client.getBusiness(businessId),
        client.getBusinessModel(businessId),
        client.getBusinessLayout(businessId).catch(() => null),
      ]);
      setBusiness(biz.data);
      const semantic = normalizeSemantic(model.data.semantic_model ?? emptySemanticModel());
      const lay =
        layout?.data.layout ??
        (semantic.nodes.length ? workOrderDefaultLayout() : emptyLayout());
      setBuffer(
        createEditBuffer({
          semantic_model: semantic,
          layout: ensureLayoutPositions(semantic, lay),
          modelRevision: model.data.current_revision,
          layoutRevision: layout?.data.layout_revision ?? 1,
        }),
      );
      setViewingRevision(null);
      setHistoricalBuffer(null);
    } catch (err) {
      setError(err instanceof FormaApiError ? err.message : '加载失败');
      setBuffer(null);
    } finally {
      setLoading(false);
    }
  }, [client, businessId, currentTenant?.tenant_id]);

  useEffect(() => {
    void load();
  }, [load]);

  const saveModel = async () => {
    if (!buffer || !businessId || !isSemanticDirty(buffer)) return;
    setSavingModel(true);
    setError(null);
    try {
      const semantic_model = adaptModelForPersistence(buffer.current.semantic_model);
      const resp = await client.putBusinessModel(businessId, {
        expected_revision: buffer.modelRevision,
        semantic_model,
        change_summary: 'Visual editor save',
      });
      setBuffer(
        resetSemanticBaseline(buffer, resp.data.current_revision, resp.data.semantic_model),
      );
      setBusiness(b =>
        b
          ? { ...b, current_revision: resp.data.current_revision, updated_at: new Date().toISOString() }
          : b,
      );
      setMessage(
        resp.data.no_change
          ? '语义无变化，修订未递增'
          : `已保存语义模型 r${resp.data.current_revision}`,
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存模型失败');
    } finally {
      setSavingModel(false);
    }
  };

  const saveLayout = async () => {
    if (!buffer || !businessId || !isLayoutDirty(buffer)) return;
    setSavingLayout(true);
    setError(null);
    try {
      const resp = await client.putBusinessLayout(businessId, {
        expected_layout_revision: buffer.layoutRevision,
        based_on_model_revision: buffer.modelRevision,
        layout: buffer.current.layout,
      });
      setBuffer(resetLayoutBaseline(buffer, resp.data.layout_revision, resp.data.layout));
      setMessage(`已保存布局 lr${resp.data.layout_revision}（未变更语义修订）`);
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存布局失败');
    } finally {
      setSavingLayout(false);
    }
  };

  const openRevisions = async () => {
    setShowRevisions(true);
    setShowDiff(false);
    try {
      const resp = await client.listBusinessRevisions(businessId);
      setRevisions(resp.data ?? []);
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载修订失败');
    }
  };

  const openHistoricalRevision = async (revNo: number) => {
    if (!buffer) return;
    try {
      const resp = await client.getBusinessRevision(businessId, revNo);
      const semantic = normalizeSemantic(resp.data.semantic_model);
      setLiveBufferSnapshot(buffer);
      setHistoricalBuffer(
        createEditBuffer({
          semantic_model: semantic,
          layout: ensureLayoutPositions(semantic, buffer.current.layout),
          modelRevision: revNo,
          layoutRevision: buffer.layoutRevision,
        }),
      );
      setViewingRevision(revNo);
      setShowRevisions(false);
      setMessage(`Viewing revision r${revNo} · Read only`);
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载历史修订失败');
    }
  };

  const backToCurrent = () => {
    if (liveBufferSnapshot) {
      setBuffer(liveBufferSnapshot);
    }
    setHistoricalBuffer(null);
    setViewingRevision(null);
    setLiveBufferSnapshot(null);
    setMessage('已返回当前编辑缓冲');
  };

  const openDiff = async () => {
    setShowDiff(true);
    setShowRevisions(false);
    const to = buffer?.modelRevision ?? 1;
    const from = Math.max(1, to - 1);
    setDiffFrom(from);
    setDiffTo(to);
    try {
      const revs = await client.listBusinessRevisions(businessId);
      setRevisions(revs.data ?? []);
      if (to > 1) {
        const resp = await client.diffBusiness(businessId, from, to);
        setDiffResult(resp.data);
      } else {
        setDiffResult(null);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Diff 失败');
    }
  };

  const runDiff = async () => {
    try {
      const resp = await client.diffBusiness(businessId, diffFrom, diffTo);
      setDiffResult(resp.data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Diff 失败');
    }
  };

  if (loading) {
    return (
      <div className="forma-panel">
        <p className="forma-placeholder">加载业务模型…</p>
      </div>
    );
  }

  if (error && !buffer) {
    return (
      <div className="forma-panel">
        <p className="forma-error">{error}</p>
        <Link to="/business">返回列表</Link>
      </div>
    );
  }

  const activeBuffer = viewingRevision != null ? historicalBuffer : buffer;
  if (!activeBuffer || !business) {
    return (
      <div className="forma-panel">
        <p className="forma-placeholder">未找到业务</p>
        <Link to="/business">返回列表</Link>
      </div>
    );
  }

  return (
    <div>
      <div style={{ marginBottom: 12, display: 'flex', gap: 12, alignItems: 'center' }}>
        <Link to="/business" className="forma-vme-btn">
          ← 业务列表
        </Link>
        {currentTenant && (
          <span className="forma-placeholder">{currentTenant.display_name}</span>
        )}
        {error && <span className="forma-error">{error}</span>}
      </div>

      <VisualModelEditor
        buffer={activeBuffer}
        onBufferChange={viewingRevision != null ? () => undefined : setBuffer}
        businessName={business.name}
        onSaveModel={() => void saveModel()}
        onSaveLayout={() => void saveLayout()}
        savingModel={savingModel}
        savingLayout={savingLayout}
        onOpenRevisions={() => void openRevisions()}
        onOpenDiff={() => void openDiff()}
        message={message}
        readOnly={viewingRevision != null}
        viewingRevision={viewingRevision}
        onBackToCurrent={backToCurrent}
      />

      {showRevisions && (
        <div className="forma-biz-side-panel" data-testid="revisions-panel">
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <h3>修订历史</h3>
            <button type="button" className="forma-vme-btn" onClick={() => setShowRevisions(false)}>
              关闭
            </button>
          </div>
          {revisions.length === 0 && <p className="forma-placeholder">暂无修订记录</p>}
          <ul style={{ margin: 0, paddingLeft: 18, fontSize: 12 }}>
            {revisions.map(r => (
              <li key={r.revision_no} style={{ marginBottom: 8 }}>
                <button
                  type="button"
                  className="forma-vme-btn"
                  data-testid={`revision-r${r.revision_no}`}
                  onClick={() => void openHistoricalRevision(r.revision_no)}
                >
                  r{r.revision_no}
                </button>{' '}
                · {r.change_summary || '—'} · {new Date(r.created_at).toLocaleString()}
                <br />
                <span className="forma-placeholder">digest {r.content_digest.slice(0, 12)}…</span>
              </li>
            ))}
          </ul>
        </div>
      )}

      {showDiff && (
        <div className="forma-biz-side-panel" data-testid="diff-panel">
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 12 }}>
            <h3>Diff</h3>
            <button type="button" className="forma-vme-btn" onClick={() => setShowDiff(false)}>
              关闭
            </button>
          </div>
          <div style={{ display: 'flex', gap: 8, alignItems: 'center', marginBottom: 12 }}>
            <label>
              from{' '}
              <select
                value={diffFrom}
                onChange={e => setDiffFrom(Number(e.target.value))}
              >
                {revisions.map(r => (
                  <option key={r.revision_no} value={r.revision_no}>
                    r{r.revision_no}
                  </option>
                ))}
              </select>
            </label>
            <label>
              to{' '}
              <select value={diffTo} onChange={e => setDiffTo(Number(e.target.value))}>
                {revisions.map(r => (
                  <option key={r.revision_no} value={r.revision_no}>
                    r{r.revision_no}
                  </option>
                ))}
              </select>
            </label>
            <button type="button" className="forma-vme-btn primary" onClick={() => void runDiff()}>
              Compare
            </button>
          </div>
          {diffResult ? (
            <DiffView data={diffResult} />
          ) : (
            <p className="forma-placeholder">选择两个修订进行对比</p>
          )}
        </div>
      )}
    </div>
  );
}

function DiffView({ data }: { data: FormaDiffResponse }) {
  const sections = [
    ['Nodes', data.diff.nodes],
    ['Edges', data.diff.edges],
    ['Rules', data.diff.rules],
    ['States', data.diff.states],
  ] as const;

  return (
    <div>
      <p className="forma-placeholder">
        r{data.diff.from_revision} → r{data.diff.to_revision}
        {data.impact.semantic_changed ? ' · semantic changed' : ' · no semantic change'}
      </p>
      {sections.map(([title, block]) => (
        <div key={title} className="forma-biz-diff-block">
          <b>{title}</b>
          <ul>
            {block.added.map(id => (
              <li key={`a-${id}`} className="forma-biz-diff-added">
                Added: {id}
              </li>
            ))}
            {block.removed.map(id => (
              <li key={`r-${id}`} className="forma-biz-diff-removed">
                Removed: {id}
              </li>
            ))}
            {block.modified.map(id => (
              <li key={`m-${id}`} className="forma-biz-diff-modified">
                Modified: {id}
              </li>
            ))}
            {!block.added.length && !block.removed.length && !block.modified.length && (
              <li className="forma-placeholder">无变更</li>
            )}
          </ul>
        </div>
      ))}
    </div>
  );
}
