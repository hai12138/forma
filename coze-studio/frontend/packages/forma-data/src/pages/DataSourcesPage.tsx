import { useCallback, useEffect, useMemo, useState } from 'react';
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
import { safeMutate, sanitizedErrorMessage } from '../utils/errors';
import { isEditor } from '../utils/roles';
import {
  adaptersForSourceType,
  buildCreateConnectionBody,
  defaultAdapterForSourceType,
  defaultPortForAdapter,
  SOURCE_TYPES,
} from '../utils/source-connection';
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
  const [sourceType, setSourceType] = useState<(typeof SOURCE_TYPES)[number]>('RELATIONAL_DATABASE');
  const [credId, setCredId] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const resp = await client.listDataSources();
      setSources(resp.data ?? []);
    } catch (err) {
      setError(sanitizedErrorMessage(err));
      setSources([]);
    } finally {
      setLoading(false);
    }
  }, [client]);

  useEffect(() => {
    void load();
  }, [load, currentTenant?.tenant_id]);

  const create = () => {
    if (!name.trim() || busy) return;
    setBusy(true);
    void safeMutate(async () => {
      await client.createDataSource({ name: name.trim(), source_type: sourceType });
      setName('');
      await load();
    }, setError).finally(() => setBusy(false));
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
              <select
                value={sourceType}
                onChange={e => setSourceType(e.target.value as (typeof SOURCE_TYPES)[number])}
                data-testid="source-type-select"
              >
                {SOURCE_TYPES.map(t => (
                  <option key={t} value={t}>
                    {t}
                  </option>
                ))}
              </select>
            </div>
            <button
              className="forma-btn forma-btn-primary"
              type="button"
              data-testid="create-source"
              disabled={busy}
              onClick={create}
            >
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
                    onClick={() =>
                      void safeMutate(async () => {
                        await client.archiveDataSource(s.source_id);
                        await load();
                      }, setError)
                    }
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
  const [environment, setEnvironment] = useState('DEV');
  const [adapterType, setAdapterType] = useState('MYSQL');
  const [host, setHost] = useState('localhost');
  const [port, setPort] = useState('3306');
  const [database, setDatabase] = useState('demo');
  const [username, setUsername] = useState('u');
  const [baseUrl, setBaseUrl] = useState('https://api.example/v1');
  const [openapiUrl, setOpenapiUrl] = useState('');
  const [credRef, setCredRef] = useState('');
  const [busy, setBusy] = useState(false);

  const sourceType = source?.source_type ?? 'RELATIONAL_DATABASE';
  const adapterOptions = useMemo(() => adaptersForSourceType(sourceType), [sourceType]);
  const isHttp = sourceType === 'HTTP_API';

  useEffect(() => {
    if (!adapterOptions.includes(adapterType)) {
      const next = defaultAdapterForSourceType(sourceType);
      setAdapterType(next);
      if (next === 'MYSQL' || next === 'POSTGRESQL') {
        setPort(String(defaultPortForAdapter(next)));
      }
    }
  }, [adapterOptions, adapterType, sourceType]);

  const reload = useCallback(async () => {
    if (!sourceId) return;
    try {
      const [s, c, a] = await Promise.all([
        client.getDataSource(sourceId),
        client.listDataConnections(sourceId),
        client.listDataAssets(sourceId),
      ]);
      setSource(s.data && !Array.isArray(s.data) ? s.data : null);
      setConnections(c.data ?? []);
      setAssets(a.data ?? []);
    } catch (err) {
      setError(sanitizedErrorMessage(err));
    }
  }, [client, sourceId]);

  useEffect(() => {
    void reload();
  }, [reload]);

  const createConnection = () => {
    if (busy || !source) return;
    setBusy(true);
    void safeMutate(async () => {
      const body = buildCreateConnectionBody({
        name: connName,
        environment,
        sourceType: source.source_type,
        adapterType,
        form: isHttp
          ? { base_url: baseUrl, openapi_url: openapiUrl }
          : { host, port, database, username },
        credential_ref_id: credRef || undefined,
      });
      await client.createDataConnection(sourceId, body);
      await reload();
    }, setError).finally(() => setBusy(false));
  };

  const capture = (connectionId: string, assetId: string) => {
    void safeMutate(async () => {
      const resp = await client.captureDataSchema(sourceId, connectionId, assetId);
      const id = resp.data.snapshot_id;
      setCapturedIds(prev => (prev.includes(id) ? prev : [...prev, id]));
      setSnapshot(resp.data);
      setTab('schema');
    }, setError);
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
            <div className="forma-panel" style={{ marginBottom: 12 }} data-testid="create-connection-form">
              <div className="forma-form-row">
                <label>连接名称</label>
                <input
                  value={connName}
                  onChange={e => setConnName(e.target.value)}
                  data-testid="connection-name"
                />
              </div>
              <div className="forma-form-row">
                <label>环境</label>
                <select
                  value={environment}
                  onChange={e => setEnvironment(e.target.value)}
                  data-testid="connection-environment"
                >
                  <option value="DEV">DEV</option>
                  <option value="TEST">TEST</option>
                  <option value="PROD">PROD</option>
                </select>
              </div>
              <div className="forma-form-row">
                <label>适配器</label>
                <select
                  value={adapterType}
                  onChange={e => {
                    setAdapterType(e.target.value);
                    if (e.target.value === 'MYSQL' || e.target.value === 'POSTGRESQL') {
                      setPort(String(defaultPortForAdapter(e.target.value)));
                    }
                  }}
                  data-testid="connection-adapter"
                >
                  {adapterOptions.map(a => (
                    <option key={a} value={a}>
                      {a}
                    </option>
                  ))}
                </select>
              </div>
              {isHttp ? (
                <>
                  <div className="forma-form-row">
                    <label>Base URL</label>
                    <input
                      value={baseUrl}
                      onChange={e => setBaseUrl(e.target.value)}
                      data-testid="connection-base-url"
                    />
                  </div>
                  <div className="forma-form-row">
                    <label>OpenAPI URL</label>
                    <input
                      value={openapiUrl}
                      onChange={e => setOpenapiUrl(e.target.value)}
                      data-testid="connection-openapi-url"
                    />
                  </div>
                </>
              ) : (
                <>
                  <div className="forma-form-row">
                    <label>Host</label>
                    <input value={host} onChange={e => setHost(e.target.value)} data-testid="connection-host" />
                  </div>
                  <div className="forma-form-row">
                    <label>Port</label>
                    <input value={port} onChange={e => setPort(e.target.value)} data-testid="connection-port" />
                  </div>
                  <div className="forma-form-row">
                    <label>Database</label>
                    <input
                      value={database}
                      onChange={e => setDatabase(e.target.value)}
                      data-testid="connection-database"
                    />
                  </div>
                  <div className="forma-form-row">
                    <label>Username</label>
                    <input
                      value={username}
                      onChange={e => setUsername(e.target.value)}
                      data-testid="connection-username"
                    />
                  </div>
                </>
              )}
              <div className="forma-form-row">
                <label>凭证引用 ID</label>
                <input value={credRef} onChange={e => setCredRef(e.target.value)} />
              </div>
              <button
                className="forma-btn forma-btn-primary"
                type="button"
                data-testid="create-connection"
                disabled={busy}
                onClick={createConnection}
              >
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
                      onClick={() =>
                        void safeMutate(async () => {
                          await client.testDataConnection(sourceId, c.connection_id);
                          await reload();
                        }, setError)
                      }
                    >
                      测试连接
                    </button>
                    <button
                      className="forma-btn"
                      type="button"
                      onClick={() =>
                        void safeMutate(async () => {
                          await client.discoverDataAssets(sourceId, c.connection_id);
                          await reload();
                        }, setError)
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
                            onClick={() => capture(c.connection_id, a.asset_id)}
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
