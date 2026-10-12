import { parseAsInteger, parseAsString, parseAsStringEnum, useQueryStates } from "nuqs";
import { useEffect, useMemo, useRef } from "react";
import { useLocation } from "react-router";
import { SortState } from "../types/utils.types";

type Options = {
  defaultLimit?: number;
  defaultSort?: SortState;
};

const tableQueryOptions = { history: "push" as const, shallow: false };

export function useTableUrlState({ defaultLimit = 25, defaultSort }: Options = {}) {
  const parsers = useMemo(
    () => ({
      query: parseAsString.withDefault("").withOptions({ throttleMs: 400, clearOnDefault: false }),
      page: parseAsInteger.withDefault(1).withOptions({ clearOnDefault: false }),
      limit: parseAsInteger.withDefault(defaultLimit).withOptions({ clearOnDefault: false }),
      sort_field: parseAsString,
      sort_order: parseAsStringEnum(["asc", "desc"] as const),
    }),
    [defaultLimit]
  );

  const [params, setParams] = useQueryStates(parsers, tableQueryOptions);
  const location = useLocation();

  useEffect(() => {
    setParams(
      {
        query: params.query,
        page: params.page,
        limit: params.limit,
        ...(params.sort_field == null && defaultSort
          ? { sort_field: defaultSort.key, sort_order: defaultSort.order }
          : {}),
      },
      { history: "replace" }
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Derived from the same location the page fetches on, so the fetch never fires against the still-bare URL.
  // Every key the mount effect writes must be present, or a partial deep link (e.g. only ?limit=) fetches twice.
  // sort_field is required only until the first sync: afterwards clearing the sort legitimately removes it.
  const urlParams = new URLSearchParams(location.search);
  const syncedRef = useRef(false);
  const ready =
    ["query", "page", "limit"].every(key => urlParams.has(key)) &&
    (syncedRef.current || !defaultSort || urlParams.has("sort_field"));
  if (ready) {
    syncedRef.current = true;
  }

  return {
    ready,
    search: {
      value: params.query,
      onChange: (q: string) => setParams({ query: q, page: 1 }),
    },
    sort: params.sort_field ? { key: params.sort_field, order: params.sort_order ?? "asc" } : undefined,
    onSortChange: (s: SortState | undefined) =>
      setParams({ sort_field: s?.key ?? null, sort_order: s?.order ?? null, page: 1 }),
    pagination: {
      page: params.page,
      perPage: params.limit,
      onPageChange: (p: number) => setParams({ page: p }),
      // Replace, not push: otherwise Back returns to the out-of-range page and gets redirected again.
      onPageOutOfRange: (lastPage: number) => setParams({ page: lastPage }, { history: "replace" }),
      onPerPageChange: (l: number) => setParams({ limit: l, page: 1 }),
    },
    // Restore defaults rather than clearing: a bare URL makes the backend fall back to its own page size/sort.
    reset: () =>
      setParams({
        query: "",
        page: 1,
        limit: defaultLimit,
        sort_field: defaultSort?.key ?? null,
        sort_order: defaultSort?.order ?? null,
      }),
  };
}
