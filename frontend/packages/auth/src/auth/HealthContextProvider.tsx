import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { useDispatch, useSelector } from "react-redux";
import { baseApi } from "@moh-sso/api";

import { selectAuthenticated, selectAuthLoaded } from "./auth.selectors";
import type { EffectiveHealthContext } from "./auth.types";
import {
  useMyHealthContextsQuery,
  useSelectHealthContextMutation,
} from "../api/health-context.api";

export const ACTIVE_HEALTH_CONTEXT_STORAGE_KEY = "moh.activeHealthContextId";

type HealthContextValue = {
  contexts: EffectiveHealthContext[];
  activeContext?: EffectiveHealthContext;
  isLoading: boolean;
  isUpdating: boolean;
  selectContext: (contextId: string) => Promise<void>;
};

const HealthContext = createContext<HealthContextValue | undefined>(undefined);

export function HealthContextProvider({ children }: { children: ReactNode }) {
  const dispatch = useDispatch();
  const authenticated = useSelector(selectAuthenticated);
  const authLoaded = useSelector(selectAuthLoaded);
  const [selectedContextID, setSelectedContextID] = useState<string | null>(() =>
    typeof window === "undefined"
      ? null
      : window.localStorage.getItem(ACTIVE_HEALTH_CONTEXT_STORAGE_KEY),
  );
  const {
    data: contexts = [],
    isLoading,
    isSuccess,
  } = useMyHealthContextsQuery(undefined, {
    skip: !authenticated,
  });
  const [selectRemote, { isLoading: isUpdating }] = useSelectHealthContextMutation();

  const activeContext = useMemo(() => {
    return (
      contexts.find((context) => context.id === selectedContextID) ??
      contexts.find((context) => context.isActive) ??
      contexts.find((context) => context.isDefault) ??
      contexts[0]
    );
  }, [contexts, selectedContextID]);

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }

    if (authLoaded && !authenticated) {
      window.localStorage.removeItem(ACTIVE_HEALTH_CONTEXT_STORAGE_KEY);
      setSelectedContextID(null);
      return;
    }

    if (authenticated && isSuccess && contexts.length === 0) {
      window.localStorage.removeItem(ACTIVE_HEALTH_CONTEXT_STORAGE_KEY);
      setSelectedContextID(null);
      return;
    }

    if (activeContext) {
      window.localStorage.setItem(ACTIVE_HEALTH_CONTEXT_STORAGE_KEY, activeContext.id);
      if (selectedContextID !== activeContext.id) {
        setSelectedContextID(activeContext.id);
      }
      window.dispatchEvent(
        new CustomEvent("moh:health-context-changed", { detail: activeContext }),
      );
    }
  }, [
    activeContext,
    authenticated,
    authLoaded,
    contexts.length,
    isSuccess,
    selectedContextID,
  ]);

  useEffect(() => {
    const handleStorage = (event: StorageEvent) => {
      if (event.key === ACTIVE_HEALTH_CONTEXT_STORAGE_KEY) {
        setSelectedContextID(event.newValue);
        dispatch(baseApi.util.resetApiState());
      }
    };
    window.addEventListener("storage", handleStorage);
    return () => window.removeEventListener("storage", handleStorage);
  }, [dispatch]);

  const selectContext = useCallback(
    async (contextId: string) => {
      await selectRemote(contextId).unwrap();
      window.localStorage.setItem(ACTIVE_HEALTH_CONTEXT_STORAGE_KEY, contextId);
      setSelectedContextID(contextId);
      dispatch(baseApi.util.resetApiState());
      const selected = contexts.find((context) => context.id === contextId);
      window.dispatchEvent(
        new CustomEvent("moh:health-context-changed", { detail: selected }),
      );
    },
    [contexts, dispatch, selectRemote],
  );

  const value = useMemo(
    () => ({ contexts, activeContext, isLoading, isUpdating, selectContext }),
    [activeContext, contexts, isLoading, isUpdating, selectContext],
  );

  return <HealthContext.Provider value={value}>{children}</HealthContext.Provider>;
}

export function useHealthContext(): HealthContextValue {
  const value = useContext(HealthContext);
  if (!value) {
    throw new Error("useHealthContext must be used inside HealthContextProvider");
  }
  return value;
}

export function useAvailableHealthContexts(): EffectiveHealthContext[] {
  return useHealthContext().contexts;
}

export function useActiveHealthContext(): EffectiveHealthContext | undefined {
  return useHealthContext().activeContext;
}

export function useRequireHealthContext(): HealthContextValue & { hasContext: boolean } {
  const value = useHealthContext();
  return { ...value, hasContext: Boolean(value.activeContext) };
}
