import { parseAsInteger, parseAsString, parseAsStringEnum, useQueryStates } from "nuqs";
import { useEffect, useMemo } from "react";
import { useLocation } from "react-router";

type SortState = { key: string; order: "asc" | "desc" };

type Options = {
  namespace?: string;
  defaultLimit?: number;
  defaultSort?: SortState;
};

const tableQueryOptions = { history: "push" as const, shallow: false };

export function useTableUrlState({ namespace = "", defaultLimit = 25, defaultSort }: Options = {}) {
  const keys = useMemo(() => {
    const prefix = namespace ? `${namespace}_` : "";
    return {
      query: `${prefix}query`,
      page: `${prefix}page`,
      limit: `${prefix}limit`,
      sort_field: `${prefix}sort_field`,
      sort_order: `${prefix}sort_order`,
    };
  }, [namespace]);

  const parsers = useMemo(
    () => ({
      [keys.query]: parseAsString.withDefault("").withOptions({ throttleMs: 400, clearOnDefault: false }),
      [keys.page]: parseAsInteger.withDefault(1).withOptions({ clearOnDefault: false }),
      [keys.limit]: parseAsInteger.withDefault(defaultLimit).withOptions({ clearOnDefault: false }),
      [keys.sort_field]: parseAsString,
      [keys.sort_order]: parseAsStringEnum(["asc", "desc"] as const),
    }),
    [keys, defaultLimit]
  );

  const [params, setParams] = useQueryStates(parsers, tableQueryOptions);
  const location = useLocation();

  useEffect(() => {
    const next: Record<string, string | number> = {
      [keys.query]: params[keys.query] as string,
      [keys.page]: params[keys.page] as number,
      [keys.limit]: params[keys.limit] as number,
    };
    if (params[keys.sort_field] == null && defaultSort) {
      next[keys.sort_field] = defaultSort.key;
      next[keys.sort_order] = defaultSort.order;
    }
    setParams(next, { history: "replace" });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Derived from the same location the page fetches on, so the fetch never fires against the still-bare URL.
  const ready = new URLSearchParams(location.search).has(keys.limit);

  const queryValue = (params[keys.query] as string | null) ?? "";
  const sortField = params[keys.sort_field] as string | null;
  const sortOrder = params[keys.sort_order] as "asc" | "desc" | null;
  const page = (params[keys.page] as number | null) ?? 1;
  const perPage = (params[keys.limit] as number | null) ?? defaultLimit;

  return {
    ready,
    search: {
      value: queryValue,
      onChange: (q: string) => setParams({ [keys.query]: q, [keys.page]: 1 }),
    },
    sort: sortField ? { key: sortField, order: sortOrder ?? "asc" } : undefined,
    onSortChange: (s: SortState | undefined) =>
      setParams({ [keys.sort_field]: s?.key ?? null, [keys.sort_order]: s?.order ?? null, [keys.page]: 1 }),
    pagination: {
      page,
      perPage,
      onPageChange: (p: number) => setParams({ [keys.page]: p }),
      onPerPageChange: (l: number) => setParams({ [keys.limit]: l, [keys.page]: 1 }),
    },
    // Restore defaults rather than clearing: a bare URL makes the backend fall back to its own page size/sort.
    reset: () =>
      setParams({
        [keys.query]: "",
        [keys.page]: 1,
        [keys.limit]: defaultLimit,
        [keys.sort_field]: defaultSort?.key ?? null,
        [keys.sort_order]: defaultSort?.order ?? null,
      }),
  };
}
