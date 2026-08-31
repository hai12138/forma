import { createFormaApiClient, type FormaBaselineData } from '@forma/api-client';
import { useCallback, useEffect, useMemo, useState } from 'react';

export function useFormaBaseline() {
  const client = useMemo(() => createFormaApiClient(), []);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [baseline, setBaseline] = useState<FormaBaselineData | null>(null);

  const reload = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const resp = await client.baseline();
      setBaseline(resp.data);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load baseline');
      setBaseline(null);
    } finally {
      setLoading(false);
    }
  }, [client]);

  useEffect(() => {
    void reload();
  }, [reload]);

  return { loading, error, baseline, reload };
}
