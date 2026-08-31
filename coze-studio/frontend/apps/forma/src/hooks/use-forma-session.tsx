import {
  createContext,
  createElement,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react';

import {
  createFormaApiClient,
  FormaApiError,
  type FormaApiClient,
  type FormaAssetCounts,
  type FormaMeData,
  type FormaTenant,
} from '@forma/api-client';

const TENANT_STORAGE_KEY = 'forma.selectedTenantId';

type SessionState =
  | 'loading'
  | 'ready'
  | 'unauthenticated'
  | 'forbidden'
  | 'suspended'
  | 'empty'
  | 'network_error';

interface FormaSessionValue {
  state: SessionState;
  error: string | null;
  me: FormaMeData | null;
  tenants: FormaTenant[];
  currentTenant: FormaTenant | null;
  assetCounts: FormaAssetCounts | null;
  client: FormaApiClient;
  switchTenant: (tenantId: string) => Promise<void>;
  refresh: () => Promise<void>;
  bootstrap: () => Promise<void>;
}

const FormaSessionContext = createContext<FormaSessionValue | null>(null);

function readStoredTenantId(): string | undefined {
  try {
    return sessionStorage.getItem(TENANT_STORAGE_KEY) || undefined;
  } catch {
    return undefined;
  }
}

function writeStoredTenantId(tenantId: string | undefined) {
  try {
    if (!tenantId) {
      sessionStorage.removeItem(TENANT_STORAGE_KEY);
    } else {
      sessionStorage.setItem(TENANT_STORAGE_KEY, tenantId);
    }
  } catch {
    // ignore
  }
}

export function FormaSessionProvider({ children }: { children: ReactNode }) {
  const [tenantId, setTenantId] = useState<string | undefined>(() => readStoredTenantId());
  const [state, setState] = useState<SessionState>('loading');
  const [error, setError] = useState<string | null>(null);
  const [me, setMe] = useState<FormaMeData | null>(null);
  const [assetCounts, setAssetCounts] = useState<FormaAssetCounts | null>(null);
  const [cacheEpoch, setCacheEpoch] = useState(0);

  const client = useMemo(
    () =>
      createFormaApiClient({
        getTenantId: () => tenantId,
      }),
    [tenantId, cacheEpoch],
  );

  const load = useCallback(async () => {
    setState('loading');
    setError(null);
    try {
      let meResp = await client.me();
      let data = meResp.data;

      if ((!data.tenants || data.tenants.length === 0) && !data.current_tenant) {
        try {
          await client.bootstrap({});
          meResp = await client.me();
          data = meResp.data;
        } catch {
          // bootstrap may fail without session spaces; surface empty state
        }
      }

      if (!data.tenants?.length) {
        setMe(data);
        setAssetCounts(null);
        setState('empty');
        return;
      }

      let selected =
        data.tenants.find(t => t.tenant_id === tenantId) ||
        data.current_tenant ||
        data.tenants[0];

      if (selected?.status === 'SUSPENDED') {
        setMe(data);
        setState('suspended');
        setError('FORMA_TENANT_SUSPENDED');
        return;
      }

      if (selected && selected.tenant_id !== tenantId) {
        setTenantId(selected.tenant_id);
        writeStoredTenantId(selected.tenant_id);
      }

      setMe({ ...data, current_tenant: selected });

      try {
        const counts = await createFormaApiClient({
          getTenantId: () => selected?.tenant_id,
        }).assetCounts();
        setAssetCounts(counts.data);
      } catch {
        setAssetCounts({ business: 0, capability: 0, agent: 0, application: 0 });
      }

      setState('ready');
    } catch (err) {
      if (err instanceof FormaApiError) {
        if (err.code === 'UNAUTHORIZED') {
          setState('unauthenticated');
          setError(err.message);
          return;
        }
        if (err.code === 'FORBIDDEN') {
          setState(err.errorKey === 'FORMA_TENANT_SUSPENDED' ? 'suspended' : 'forbidden');
          setError(err.errorKey || err.message);
          return;
        }
        if (err.code === 'NETWORK_ERROR') {
          setState('network_error');
          setError(err.message);
          return;
        }
      }
      setState('network_error');
      setError(err instanceof Error ? err.message : 'Unknown error');
    }
  }, [client, tenantId]);

  useEffect(() => {
    void load();
  }, [load]);

  const switchTenant = useCallback(async (nextTenantId: string) => {
    writeStoredTenantId(nextTenantId);
    setTenantId(nextTenantId);
    setAssetCounts(null);
    setCacheEpoch(v => v + 1);
  }, []);

  const bootstrap = useCallback(async () => {
    await client.bootstrap({});
    setCacheEpoch(v => v + 1);
  }, [client]);

  const value: FormaSessionValue = {
    state,
    error,
    me,
    tenants: me?.tenants ?? [],
    currentTenant: me?.current_tenant ?? null,
    assetCounts,
    client,
    switchTenant,
    refresh: load,
    bootstrap,
  };

  return createElement(FormaSessionContext.Provider, { value }, children);
}

export function useFormaSession() {
  const ctx = useContext(FormaSessionContext);
  if (!ctx) {
    throw new Error('useFormaSession must be used within FormaSessionProvider');
  }
  return ctx;
}
