import { useCallback, useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';

import type {
  FormaDataAsset,
  FormaDataConnection,
  FormaDataSource,
  FormaPhysicalSchema,
  FormaSchemaSnapshot,
} from '@forma/api-client';

import { EmptyState } from '../components/EmptyState';
import { SecretCredentialForm } from '../components/SecretCredentialForm';
import { StatusBadge } from '../components/StatusBadge';
import { isEditor } from '../utils/roles';
import { useDataPlaneContext } from './useDataPlaneContext';

function parseSchema(raw: FormaSchemaSnapshot['schema']): FormaPhysicalSchema | null {
  if (!raw || typeof raw !== 'object') return null;
  if ('fields' in raw) return raw as FormaPhysicalSchema;
  return null;
}

export function DataSourcesPage() {
  const { client, currentTenant, businessId } = useDataPlaneContext();
  const canEdit = isEditor(currentTenant?.role);
  const [sources, setSources] = useState<FormaDataSource[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [name, setName] = useState('');
  const [sourceType, setSourceType] = useState('EXTERNAL_SQL');
  const [credId, setCredId] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const resp = await client.listDataSources();
      setSources(resp.data ?? []);
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败');
      setSources([]);
    } finally {
      setLoading(false);
    }
  }, [client]);

  useEffect(() => {
    void load();
  }, [load, currentTenant?.tenant_id]);

  const create = async () => {
    if (!name.trim()) return;
    await client.createDataSource({ name: name.trim(), source_type: sourceType });
    setName('');
    await load();
  };

  const q = businessId ? `?businessId=${encodeURIComponent(businessId)}` : '';

  return (
    <div data-testid="data-sources-page">
      <div className="forma-data-toolbar" style={{ marginBottom: 12 }}>
        <h2 style={{ margin: 0 }}>数据源</h2>
      </div>
      {error ? <div className="forma-error">{error}</div> : null}
      {canEdit ? (
        <>
          <SecretCredentialForm
            client={client}
            canEdit={canEdit}
            onCreated={c => setCredId(c.credential_ref_id)}
            onError={setError}
          />
          {credId ? <p className="forma-muted">最近凭证引用：{credId}</p> : null}
          <div className="forma-panel" style={{ marginBottom: 12 }}>
            <div className="forma-form-row">
              <label>名称</label>
              <input value={name} onChange={e => setName(e.target.value)} data-testid="source-name" />
            </div>
            <div className="forma-form-row">
              <label>类型</label>
              <select value={sourceType} onChange={e => setSourceType(e.target.value)}>
                <option value="EXTERNAL_SQL">EXTERNAL_SQL</option>
                <option value="EXTERNAL_HTTP">EXTERNAL_HTTP</option>
              </select>
            </div>
            <button className="forma-btn forma-btn-primary" type="button" onClick={() => void create()}>
              创建数据源
            </button>
          </div>
        </>
      ) : null}
      {loading ? <div className="forma-muted">加载中…</div> : null}
      {!loading && sources.length === 0 ? (
        <EmptyState title="暂无数据源" hint="创建外部数据源并配置连接与凭证。" />
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {sources.map(s => (
            <div className="forma-card" key={s.source_id}>
              <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                <Link to={`../sources/${s.source_id}${q}`}>{s.name}</Link>
                <StatusBadge status={s.status} />
                <span className="forma-muted">{s.source_type}</span>
              </div>
              {canEdit && s.status !== 'ARCHIVED' ? (
                <div className="forma-card-actions">
                  <button
                    className="forma-btn"
                    type="button"
                    onClick={() => void client.archiveDataSource(s.source_id).then(load)}
                  >
                    归档
                  </button>
                </div>
              ) : null}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export function SourceDetailPage() {
  const { sourceId = '' } = useParams();
  const { client, currentTenant, businessId } = useDataPlaneContext();
  const canEdit = isEditor(currentTenant?.role);
  const [source, setSource] = useState<FormaDataSource | null>(null);
  const [connections, setConnections] = useState<FormaDataConnection[]>([]);
  const [assets, setAssets] = useState<FormaDataAsset[]>([]);
  const [tab, setTab] = useState<'connections' | 'schema'>('connections');
  const [capturedIds, setCapturedIds] = useState<string[]>([]);
  const [snapshot, setSnapshot] = useState<FormaSchemaSnapshot | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [connName, setConnName] = useState('default');
  const [credRef, setCredRef] = useState('');

  const reload = useCallback(async () => {
    if (!sourceId) return;
    try {
      const [s, c, a] = await Promise.all([
        client.getDataSource(sourceId),
        client.listDataConnections(sourceId),
        client.listDataAssets(sourceId),
      ]);
      setSource(s.data);
      setConnections(c.data ?? []);
      setAssets(a.data ?? []);
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败');
    }
  }, [client, sourceId]);

  useEffect(() => {
    void reload();
  }, [reload]);

  const createConnection = async () => {
    await client.createDataConnection(sourceId, {
      name: connName,
      environment: 'DEV',
      adapter_type: 'MYSQL',
      public_config: { host: 'localhost', port: 3306, database: 'demo', username: 'u' },
      credential_ref_id: credRef || undefined,
    });
    await reload();
  };

  const capture = async (connectionId: string, assetId: string) => {
    const resp = await client.captureDataSchema(sourceId, connectionId, assetId);
    const id = resp.data.snapshot_id;
    setCapturedIds(prev => (prev.includes(id) ? prev : [...prev, id]));
    setSnapshot(resp.data);
    setTab('schema');
  };

  const loadSnapshot = async (id: string) => {
    const resp = await client.getSchemaSnapshot(id);
    setSnapshot(resp.data);
  };

  const q = businessId ? `?businessId=${encodeURIComponent(businessId)}` : '';
  const schema = snapshot ? parseSchema(snapshot.schema) : null;

  if (!sourceId) return <EmptyState title="缺少数据源 ID" />;

  return (
    <div data-testid="source-detail-page">
      <div className="forma-data-toolbar">
        <Link to={`../sources${q}`}>← 返回</Link>
        <h2 style={{ margin: 0 }}>{source?.name ?? sourceId}</h2>
        {source ? <StatusBadge status={source.status} /> : null}
      </div>
      {error ? <div className="forma-error">{error}</div> : null}
      <div className="forma-tabs">
        <button
          type="button"
          className={`forma-tab${tab === 'connections' ? ' active' : ''}`}
          onClick={() => setTab('connections')}
        >
          连接与资产
        </button>
        <button
          type="button"
          className={`forma-tab${tab === 'schema' ? ' active' : ''}`}
          onClick={() => setTab('schema')}
          data-testid="schema-explorer-tab"
        >
          Schema Explorer
        </button>
      </div>
      {tab === 'connections' ? (
        <>
          {canEdit ? (
            <div className="forma-panel" style={{ marginBottom: 12 }}>
              <div className="forma-form-row">
                <label>连接名称</label>
                <input value={connName} onChange={e => setConnName(e.target.value)} />
              </div>
              <div className="forma-form-row">
                <label>凭证引用 ID</label>
                <input value={credRef} onChange={e => setCredRef(e.target.value)} />
              </div>
              <button className="forma-btn forma-btn-primary" type="button" onClick={() => void createConnection()}>
                创建连接
              </button>
            </div>
          ) : null}
          {connections.length === 0 ? (
            <EmptyState title="暂无连接" />
          ) : (
            connections.map(c => (
              <div className="forma-card" key={c.connection_id}>
                <div>
                  <strong>{c.name}</strong> · {c.adapter_type} · {c.environment}{' '}
                  <StatusBadge status={c.status} />
                </div>
                {canEdit ? (
                  <div className="forma-card-actions">
                    <button
                      className="forma-btn"
                      type="button"
                      onClick={() => void client.testDataConnection(sourceId, c.connection_id).then(reload)}
                    >
                      测试连接
                    </button>
                    <button
                      className="forma-btn"
                      type="button"
                      onClick={() =>
                        void client.discoverDataAssets(sourceId, c.connection_id).then(reload)
                      }
                    >
                      发现资产
                    </button>
                  </div>
                ) : null}
                <ul>
                  {assets
                    .filter(a => a.connection_id === c.connection_id)
                    .map(a => (
                      <li key={a.asset_id}>
                        {a.name} ({a.asset_type})
                        {canEdit ? (
                          <button
                            className="forma-btn"
                            type="button"
                            style={{ marginLeft: 8 }}
                            onClick={() => void capture(c.connection_id, a.asset_id)}
                          >
                            捕获 Schema
                          </button>
                        ) : null}
                      </li>
                    ))}
                </ul>
              </div>
            ))
          )}
        </>
      ) : (
        <div data-testid="schema-explorer">
          <p className="forma-muted">已捕获快照保存在本页状态（无需 ListSnapshots）。</p>
          {capturedIds.length === 0 ? (
            <EmptyState title="尚未捕获 Schema" hint="在连接页面对资产执行捕获。" />
          ) : (
            <div className="forma-form-row">
              <label>快照</label>
              <select
                value={snapshot?.snapshot_id ?? ''}
                onChange={e => void loadSnapshot(e.target.value)}
              >
                <option value="">选择快照…</option>
                {capturedIds.map(id => (
                  <option key={id} value={id}>
                    {id}
                  </option>
                ))}
              </select>
            </div>
          )}
          {schema ? (
            <div className="forma-panel">
              <strong>{schema.name}</strong>
              <ul>
                {schema.fields.map(f => (
                  <li key={f.path || f.name}>
                    {f.path || f.name} · {f.data_type}
                    {f.primary_key ? ' · PK' : ''}
                  </li>
                ))}
              </ul>
            </div>
          ) : null}
        </div>
      )}
    </div>
  );
}
