import {
  createContext,
  createElement,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
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

export type SessionState =
  | 'loading'
  | 'ready'
  | 'unauthenticated'
  | 'forbidden'
  | 'suspended'
  | 'empty'
  | 'authenticated_no_tenant'
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
  refresh: () => Promise<SessionState>;
  bootstrap: () => Promise<void>;
  clearLocalSession: () => void;
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
  const unauthorizedNavRef = useRef(false);
  const wasReadyRef = useRef(false);

  const handleUnauthorized = useCallback(() => {
    // Only treat mid-session 401 as expiry — initial unauthenticated /me must not flash "expired".
    if (!wasReadyRef.current) return;
    if (unauthorizedNavRef.current) return;
    if (typeof window === 'undefined') return;
    const path = `${window.location.pathname}${window.location.search}`;
    if (path.startsWith('/login')) return;
    unauthorizedNavRef.current = true;
    wasReadyRef.current = false;
    writeStoredTenantId(undefined);
    setTenantId(undefined);
    setMe(null);
    setAssetCounts(null);
    setState('unauthenticated');
    const returnTo = encodeURIComponent(path.startsWith('/') && !path.startsWith('//') ? path : '/');
    window.location.assign(`/login?expired=1&returnTo=${returnTo}`);
  }, []);

  const client = useMemo(
    () =>
      createFormaApiClient({
        getTenantId: () => tenantId,
        onUnauthorized: handleUnauthorized,
      }),
    [tenantId, cacheEpoch, handleUnauthorized],
  );

  const load = useCallback(async (): Promise<SessionState> => {
    setState('loading');
    setError(null);
    unauthorizedNavRef.current = false;
    try {
      const meResp = await client.me();
      const data = meResp.data;

      if (!data.tenants?.length) {
        setMe(data);
        setAssetCounts(null);
        wasReadyRef.current = false;
        const next: SessionState = 'empty';
        setState(next);
        return next;
      }

      let selected =
        data.tenants.find(t => t.tenant_id === tenantId) ||
        data.current_tenant ||
        data.tenants[0];

      if (selected?.status === 'SUSPENDED') {
        setMe(data);
        setState('suspended');
        setError('FORMA_TENANT_SUSPENDED');
        return 'suspended';
      }

      if (selected && selected.tenant_id !== tenantId) {
        setTenantId(selected.tenant_id);
        writeStoredTenantId(selected.tenant_id);
      }

      setMe({ ...data, current_tenant: selected });

      try {
        const counts = await createFormaApiClient({
          getTenantId: () => selected?.tenant_id,
          onUnauthorized: handleUnauthorized,
        }).assetCounts();
        setAssetCounts(counts.data);
      } catch {
        setAssetCounts({ business: 0, capability: 0, agent: 0, application: 0 });
      }

      setState('ready');
      wasReadyRef.current = true;
      return 'ready';
    } catch (err) {
      wasReadyRef.current = false;
      if (err instanceof FormaApiError) {
        if (err.code === 'UNAUTHORIZED') {
          setMe(null);
          setAssetCounts(null);
          setState('unauthenticated');
          setError(null);
          return 'unauthenticated';
        }
        if (err.code === 'FORBIDDEN') {
          const next: SessionState =
            err.errorKey === 'FORMA_TENANT_SUSPENDED' ? 'suspended' : 'forbidden';
          setState(next);
          setError(err.errorKey || err.message);
          return next;
        }
        if (err.code === 'NETWORK_ERROR') {
          setState('network_error');
          setError(err.message);
          return 'network_error';
        }
      }
      setState('network_error');
      setError(err instanceof Error ? err.message : 'Unknown error');
      return 'network_error';
    }
  }, [client, tenantId, handleUnauthorized]);

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

  const clearLocalSession = useCallback(() => {
    wasReadyRef.current = false;
    writeStoredTenantId(undefined);
    setTenantId(undefined);
    setMe(null);
    setAssetCounts(null);
    setState('unauthenticated');
    setCacheEpoch(v => v + 1);
  }, []);

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
    clearLocalSession,
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
