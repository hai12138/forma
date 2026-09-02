import { useOutletContext } from 'react-router-dom';

import type { DataPlaneOutletContext } from '../components/DataPlaneShell';

export function useDataPlaneContext(): DataPlaneOutletContext {
  return useOutletContext<DataPlaneOutletContext>();
}
