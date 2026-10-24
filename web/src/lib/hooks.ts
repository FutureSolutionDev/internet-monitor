import { useEffect, useState } from "react";

export function useNow(ms = 1000) {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const id = window.setInterval(() => setNow(Date.now()), ms);
    return () => window.clearInterval(id);
  }, [ms]);
  return now;
}

export function useAsync<T>(load: (signal: AbortSignal) => Promise<T>, deps: unknown[], scope = "") {
  const [state, setState] = useState<{ data: T | null; error: unknown; loading: boolean; scope: string }>({
    data: null,
    error: null,
    loading: true,
    scope,
  });
  const [nonce, setNonce] = useState(0);
  // biome-ignore lint/correctness/useExhaustiveDependencies: callers pass explicit deps
  useEffect(() => {
    const ctrl = new AbortController();
    setState((s) => ({ ...s, loading: true, error: null }));
    load(ctrl.signal).then(
      (data) => {
        if (!ctrl.signal.aborted) setState({ data, error: null, loading: false, scope });
      },
      (error: unknown) => {
        if (!ctrl.signal.aborted)
          setState((s) => ({ data: s.scope === scope ? s.data : null, error, loading: false, scope }));
      },
    );
    return () => ctrl.abort();
  }, [...deps, scope, nonce]);
  const current = state.scope === scope;
  return {
    data: current ? state.data : null,
    error: current ? state.error : null,
    loading: state.loading || !current,
    reload: () => setNonce((n) => n + 1),
  };
}
