'use client';
import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  type ReactNode,
} from 'react';
import { initialState, type State, invalidate } from './domain';
import {
  applyVisualModel,
  createVisualModel,
  editSemantic,
} from './visual-model';
type Store = {
  state: State;
  update: (patch: Partial<State>, action: string, changed?: boolean) => void;
  notify: (text: string) => void;
  toast: string;
  reset: () => void;
};
const Context = createContext<Store | null>(null);
const KEY = 'forma-prototype-v1';
export function StoreProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<State>(initialState);
  const [ready, setReady] = useState(false);
  const [toast, setToast] = useState('');
  useEffect(() => {
    const task = setTimeout(() => {
      try {
        const saved = localStorage.getItem(KEY);
        if (saved) {
          const parsed = JSON.parse(saved);
          if (
            parsed?.agents &&
            parsed?.application &&
            parsed?.approvals &&
            typeof parsed.revision === 'number'
          )
            setState({ ...initialState, ...parsed });
        }
      } catch {
        setToast('无法读取本地记录，已载入演示数据。');
      }
      setReady(true);
    }, 0);
    return () => clearTimeout(task);
  }, []);
  useEffect(() => {
    if (ready)
      try {
        localStorage.setItem(KEY, JSON.stringify(state));
      } catch {
        queueMicrotask(() =>
          setToast('本机存储不可用，修改仅保留在当前会话。'),
        );
      }
  }, [state, ready]);
  useEffect(() => {
    if (!toast) return;
    const t = setTimeout(() => setToast(''), 4200);
    return () => clearTimeout(t);
  }, [toast]);
  const update = useCallback(
    (patch: Partial<State>, action: string, changed = false) => {
      setState((s) => {
        let next = changed ? invalidate(s, patch) : { ...s, ...patch };
        if (patch.visualModel) next = applyVisualModel(s, patch.visualModel);
        else if (patch.rule !== undefined && patch.rule !== s.rule) {
          const model = s.visualModel ?? createVisualModel(s.rule);
          const old = model.semantic_model.nodes.find((n) => n.id === 'rule');
          const nodes = old
            ? model.semantic_model.nodes.map((n) =>
                n.id === 'rule' ? { ...n, description: patch.rule } : n,
              )
            : [
                ...model.semantic_model.nodes,
                {
                  id: 'rule',
                  type: 'rule' as const,
                  label: '派单规则',
                  description: patch.rule,
                  source: 'manual_modified' as const,
                },
              ];
          const edited = editSemantic(
            model,
            { ...model.semantic_model, nodes },
            '更新候选规则',
          );
          next = applyVisualModel({ ...s, ...patch }, edited);
        }
        return {
          ...next,
          audit: [
            { at: new Date().toLocaleTimeString('zh-CN'), action },
            ...s.audit,
          ].slice(0, 100),
        };
      });
      setToast(action);
    },
    [],
  );
  return (
    <Context.Provider
      value={{
        state,
        update,
        toast,
        notify: setToast,
        reset: () => {
          setState(structuredClone(initialState));
          setToast('演示数据已重置');
        },
      }}
    >
      {children}
    </Context.Provider>
  );
}
export function useStore() {
  const context = useContext(Context);
  if (!context) throw Error('StoreProvider missing');
  return context;
}
