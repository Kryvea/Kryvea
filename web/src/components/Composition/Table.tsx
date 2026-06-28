import { mdiChevronDown, mdiChevronUp, mdiClose } from "@mdi/js";
import {
  MouseEvent as ReactMouseEvent,
  ReactNode,
  isValidElement,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import Card from "./Card";
import Flex from "./Flex";
import Icon from "./Icon";
import Paginator from "./Paginator";
import Shimmer from "./Shimmer";

type SortState = { key: string; order: "asc" | "desc" };

type DataColumn<Row> = {
  kind?: "data";
  header: string;
  render: (row: Row) => ReactNode;
  sortKey?: string;
  sortable?: boolean;
  sortValue?: (row: Row) => string | number | Date;
  maxWidth?: string;
  /** Seed this column at its header width instead of its content width (still resizable afterwards). */
  fitHeader?: boolean;
};

type ActionsColumn<Row> = {
  kind: "actions";
  render: (row: Row) => ReactNode;
};

export type Column<Row> = DataColumn<Row> | ActionsColumn<Row>;

interface SearchProps {
  value: string;
  onChange: (q: string) => void;
}

interface PaginationProps {
  page: number;
  perPage: number;
  totalPages?: number;
  totalRows?: number;
  onPageChange: (page: number) => void;
  onPerPageChange: (perPage: number) => void;
}

type RowWithId = { id: string | number };

type BaseTableProps<Row extends RowWithId> = {
  columns: Column<Row>[];
  data: Row[];
  loading?: boolean;
  rowKey?: (row: Row, index: number) => string | number;
  perPage?: number;
  /** Stable id used to persist per-column widths in localStorage. Omit to keep widths only for the session. */
  tableId?: string;
};

type ControlledTableProps = {
  search: SearchProps;
  pagination: PaginationProps;
  sort: SortState | undefined;
  onSortChange: (next: SortState | undefined) => void;
  defaultSort?: never;
};

type UncontrolledTableProps = {
  search?: undefined;
  pagination?: undefined;
  sort?: undefined;
  onSortChange?: undefined;
  defaultSort?: SortState;
};

type TableProps<Row extends RowWithId> = BaseTableProps<Row> & (ControlledTableProps | UncontrolledTableProps);

const PAGE_FLOOR = 1;
const WIDTHS_PREFIX = "kryvea:table-widths:";
const MIN_COLUMN_WIDTH = 48;
const ACTIONS_KEY = "__actions__";

// Width persistence keys columns by header, so data-column headers must be unique within a table.
const columnKey = <Row,>(column: Column<Row>) => (column.kind === "actions" ? ACTIONS_KEY : column.header);

function headerContentWidth(th: HTMLElement): number {
  const label = th.querySelector<HTMLElement>(".th-label");
  if (!label) {
    return Math.round(th.getBoundingClientRect().width);
  }
  const style = getComputedStyle(th);
  return Math.ceil(label.scrollWidth + parseFloat(style.paddingLeft) + parseFloat(style.paddingRight));
}

function extractText(node: ReactNode): string {
  if (node == null || typeof node === "boolean") {
    return "";
  }
  if (typeof node === "string" || typeof node === "number") {
    return String(node);
  }
  if (Array.isArray(node)) {
    return node.map(extractText).join(" ");
  }
  if (isValidElement<{ children?: ReactNode }>(node)) {
    return extractText(node.props.children);
  }
  return "";
}

function compareValues(a: string | number | Date, b: string | number | Date): number {
  if (typeof a === "string" && typeof b === "string") {
    return a.localeCompare(b, undefined, { sensitivity: "base", numeric: true });
  }
  if (a < b) {
    return -1;
  }
  return a > b ? 1 : 0;
}

function cycleSort(prev: SortState | undefined, key: string): SortState | undefined {
  if (!prev || prev.key !== key) {
    return { key, order: "asc" };
  }
  return prev.order === "asc" ? { key, order: "desc" } : undefined;
}

function loadColumnWidths(tableId?: string): Record<string, number> {
  if (!tableId) {
    return {};
  }
  try {
    const parsed = JSON.parse(localStorage.getItem(WIDTHS_PREFIX + tableId) ?? "{}") as Record<string, unknown>;
    const widths: Record<string, number> = {};
    for (const [key, value] of Object.entries(parsed)) {
      // Drop anything that isn't a usable width; a NaN/legacy entry would otherwise collapse the whole table.
      if (typeof value === "number" && Number.isFinite(value) && value > 0) {
        widths[key] = value;
      }
    }
    return widths;
  } catch {
    return {};
  }
}

function saveColumnWidths(tableId: string | undefined, widths: Record<string, number>) {
  if (!tableId) {
    return;
  }
  try {
    localStorage.setItem(WIDTHS_PREFIX + tableId, JSON.stringify(widths));
  } catch {
    /* ignore quota / serialization errors */
  }
}

function useControllableState<T>(
  isControlled: boolean,
  controlledValue: T,
  onControlledChange: ((value: T) => void) | undefined,
  initial: T
): [T, (value: T) => void] {
  const [internal, setInternal] = useState(initial);
  const setValue = useCallback(
    (next: T) => {
      onControlledChange?.(next);
      if (!isControlled) {
        setInternal(next);
      }
    },
    [isControlled, onControlledChange]
  );
  return [isControlled ? controlledValue : internal, setValue];
}

function useTableData<Row>(params: {
  columns: Column<Row>[];
  data: Row[];
  serverMode: boolean;
  query: string;
  sort: SortState | undefined;
  page: number;
  perPage: number;
  totalPages?: number;
}) {
  const { columns, data, serverMode, query, sort, page, perPage, totalPages } = params;

  const filteredData = useMemo(() => {
    if (serverMode || !query) {
      return data;
    }
    const q = query.toLowerCase();
    return data.filter(row =>
      columns.some(column => column.kind !== "actions" && extractText(column.render(row)).toLowerCase().includes(q))
    );
  }, [serverMode, query, data, columns]);

  const sortedData = useMemo(() => {
    if (serverMode || !sort) {
      return filteredData;
    }
    const column = columns.find(c => c.kind !== "actions" && (c.sortKey ?? c.header) === sort.key) as
      | DataColumn<Row>
      | undefined;
    if (!column) {
      return filteredData;
    }
    const valueOf = column.sortValue ?? ((row: Row) => extractText(column.render(row)));
    const direction = sort.order === "asc" ? 1 : -1;
    // Compute each sort key once (Schwartzian transform) rather than re-deriving it on every comparison.
    return filteredData
      .map(row => ({ row, sortValue: valueOf(row) }))
      .sort((a, b) => compareValues(a.sortValue, b.sortValue) * direction)
      .map(entry => entry.row);
  }, [serverMode, filteredData, sort, columns]);

  // Slicing is separate from sorting so paging through results doesn't re-sort the whole dataset.
  const visibleData = useMemo(() => {
    if (serverMode) {
      return sortedData;
    }
    return sortedData.slice(perPage * (page - PAGE_FLOOR), perPage * page);
  }, [serverMode, sortedData, page, perPage]);

  const numPages = useMemo(() => {
    if (serverMode) {
      return totalPages ?? 0;
    }
    const pages = Math.ceil(filteredData.length / perPage);
    return Number.isNaN(pages) ? 0 : pages;
  }, [serverMode, totalPages, filteredData.length, perPage]);

  return { filteredData, visibleData, numPages };
}

function useResizableColumns<Row>(columns: Column<Row>[], tableId?: string, ready?: boolean) {
  const [baseWidths, setBaseWidths] = useState<Record<string, number>>(() => loadColumnWidths(tableId));
  const [containerWidth, setContainerWidth] = useState(0);
  const [dragging, setDragging] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const headerRefs = useRef<Record<string, HTMLTableCellElement | null>>({});
  const refSetters = useRef<Record<string, (el: HTMLTableCellElement | null) => void>>({});
  const draggingRef = useRef(false);
  const endDragRef = useRef<(() => void) | null>(null);

  useLayoutEffect(() => {
    const el = containerRef.current;
    if (!el) {
      return;
    }
    const measure = () => setContainerWidth(el.clientWidth);
    measure();
    const observer = new ResizeObserver(measure);
    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  // Unmounting mid-drag would otherwise leave the document listeners and body styles in place.
  useEffect(() => () => endDragRef.current?.(), []);

  // Seed widths only once real rows are present: an empty body would size columns from the headers alone.
  const colSignature = columns.map(columnKey).join("|");
  useLayoutEffect(() => {
    if (!ready) {
      return;
    }
    setBaseWidths(prev => {
      const next = { ...prev };
      let changed = false;
      for (const column of columns) {
        const key = columnKey(column);
        const el = headerRefs.current[key];
        if (next[key] == null && el) {
          next[key] =
            column.kind !== "actions" && column.fitHeader
              ? headerContentWidth(el)
              : Math.round(el.getBoundingClientRect().width);
          changed = true;
        }
      }
      return changed ? next : prev;
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [colSignature, ready]);

  const allWidthsKnown = columns.every(column => baseWidths[columnKey(column)] != null);

  // Leftover space is spread across data columns proportionally to their width (AG Grid "size columns to fit").
  // Integer widths only, so 1px borders never land on a fractional edge; the last data column absorbs the rounding.
  const columnWidths = useMemo(() => {
    if (!allWidthsKnown || dragging) {
      return baseWidths;
    }
    const dataKeys = columns.filter(column => column.kind !== "actions").map(columnKey);
    const sumAll = columns.reduce((sum, column) => sum + baseWidths[columnKey(column)], 0);
    const sumData = dataKeys.reduce((sum, key) => sum + baseWidths[key], 0);
    const targetData = containerWidth - (sumAll - sumData);
    if (containerWidth <= sumAll || sumData <= 0 || dataKeys.length === 0) {
      return baseWidths;
    }
    const next = { ...baseWidths };
    let used = 0;
    dataKeys.forEach((key, index) => {
      next[key] = index === dataKeys.length - 1 ? targetData - used : Math.round(targetData * (baseWidths[key] / sumData));
      used += next[key];
    });
    return next;
  }, [allWidthsKnown, dragging, columns, baseWidths, containerWidth]);

  const totalWidth = allWidthsKnown
    ? columns.reduce((sum, column) => sum + columnWidths[columnKey(column)], 0)
    : undefined;

  // Stable ref callback per column key, so the header refs aren't detached and re-attached on every render.
  const setHeaderRef = (key: string) => {
    if (!refSetters.current[key]) {
      refSetters.current[key] = el => {
        headerRefs.current[key] = el;
      };
    }
    return refSetters.current[key];
  };

  const startResize = (event: ReactMouseEvent, key: string) => {
    event.preventDefault();
    // Freeze every column at its rendered (post-fill) width so entering drag-mode doesn't jump the layout.
    setBaseWidths(prev => {
      const next = { ...prev };
      for (const column of columns) {
        const key = columnKey(column);
        const el = headerRefs.current[key];
        if (el) {
          next[key] = Math.round(el.getBoundingClientRect().width);
        }
      }
      return next;
    });
    setDragging(true);
    const startX = event.clientX;
    const th = (event.currentTarget as HTMLElement).parentElement;
    const startWidth = th ? Math.round(th.getBoundingClientRect().width) : MIN_COLUMN_WIDTH;
    // A column can't shrink below its header content, otherwise the header would overlap the next column.
    const minWidth = Math.max(MIN_COLUMN_WIDTH, th ? headerContentWidth(th) : MIN_COLUMN_WIDTH);
    draggingRef.current = false;

    const onMove = (moveEvent: MouseEvent) => {
      if (moveEvent.clientX !== startX) {
        draggingRef.current = true;
      }
      setBaseWidths(prev => ({
        ...prev,
        [key]: Math.max(minWidth, Math.round(startWidth + moveEvent.clientX - startX)),
      }));
    };

    const cleanup = () => {
      document.removeEventListener("mousemove", onMove);
      document.removeEventListener("mouseup", onUp);
      document.body.style.userSelect = "";
      document.body.style.cursor = "";
      endDragRef.current = null;
    };

    const onUp = () => {
      cleanup();
      setDragging(false);
      setBaseWidths(prev => {
        saveColumnWidths(tableId, prev);
        return prev;
      });
      // Let the click that follows mouseup read the drag flag (to skip sorting) before it is reset.
      setTimeout(() => {
        draggingRef.current = false;
      }, 0);
    };

    endDragRef.current = cleanup;
    document.body.style.userSelect = "none";
    document.body.style.cursor = "col-resize";
    document.addEventListener("mousemove", onMove);
    document.addEventListener("mouseup", onUp);
  };

  return {
    columnWidths,
    allWidthsKnown,
    totalWidth,
    containerRef,
    setHeaderRef,
    startResize,
    isResizing: () => draggingRef.current,
  };
}

// Clipped to the column width; the tooltip is shown only when the text is actually truncated.
function TruncatingCell({
  content,
  fixed,
  maxWidth,
  columnWidth,
}: {
  content: ReactNode;
  fixed: boolean;
  maxWidth?: string;
  columnWidth?: number;
}) {
  const ref = useRef<HTMLDivElement>(null);
  const [truncated, setTruncated] = useState(false);
  useLayoutEffect(() => {
    const el = ref.current;
    setTruncated(!!el && el.scrollWidth > el.clientWidth);
  }, [content, fixed, maxWidth, columnWidth]);
  return (
    <td className="text-nowrap">
      <div
        ref={ref}
        className="truncate-cell"
        // During the measuring pass `maxWidth` sets the column's width, so it seeds at exactly that size.
        style={!fixed && maxWidth != null ? { width: maxWidth } : undefined}
        title={truncated ? extractText(content) || undefined : undefined}
      >
        {content}
      </div>
    </td>
  );
}

export default function Table<Row extends RowWithId>({
  columns,
  data,
  loading,
  perPage: perPageInitial = 5,
  rowKey,
  tableId,
  search,
  pagination,
  sort,
  onSortChange,
  defaultSort,
}: TableProps<Row>) {
  const serverMode = pagination?.totalRows !== undefined;

  const [query, setQuery] = useControllableState(!!search, search?.value ?? "", search?.onChange, "");
  const [page, setPage] = useControllableState(
    !!pagination,
    pagination?.page ?? PAGE_FLOOR,
    pagination?.onPageChange,
    PAGE_FLOOR
  );
  const [perPage, setPerPage] = useControllableState(
    !!pagination,
    pagination?.perPage ?? perPageInitial,
    pagination?.onPerPageChange,
    perPageInitial
  );
  const [activeSort, setActiveSort] = useControllableState(!!onSortChange, sort, onSortChange, defaultSort);

  const { filteredData, visibleData, numPages } = useTableData({
    columns,
    data,
    serverMode,
    query,
    sort: activeSort,
    page,
    perPage,
    totalPages: pagination?.totalPages,
  });
  const hasRows = !loading && visibleData.length > 0;
  const resize = useResizableColumns(columns, tableId, hasRows);

  const pagesList = useMemo(() => Array.from({ length: numPages }, (_, i) => i + PAGE_FLOOR), [numPages]);

  const sortKeyOf = (column: DataColumn<Row>) => (serverMode ? column.sortKey : (column.sortKey ?? column.header));
  const isSortable = (column: DataColumn<Row>) => (serverMode ? !!column.sortKey : column.sortable !== false);

  const handleSearch = (value: string) => {
    // Controlled search resets the page through the parent; only reset it ourselves in client mode.
    if (!search) {
      setPage(PAGE_FLOOR);
    }
    setQuery(value);
  };

  const handlePerPage = (value: number) => {
    setPerPage(value);
    setPage(PAGE_FLOOR);
  };

  const handleHeaderClick = (column: DataColumn<Row>) => {
    if (resize.isResizing()) {
      return;
    }
    const key = sortKeyOf(column);
    if (key) {
      setActiveSort(cycleSort(activeSort, key));
    }
  };

  return (
    <Card className="!relative !gap-0 !p-0">
      <Flex className="px-2 pt-1" items="center">
        <input
          className="w-full rounded-t-2xl bg-transparent focus:border-transparent"
          placeholder="Search"
          type="text"
          value={query}
          onChange={e => handleSearch(e.target.value)}
        />
        {query !== "" && (
          <span onClick={() => handleSearch("")}>
            <Icon className="text-[color:--text-secondary] hover:opacity-50" path={mdiClose} size={18} />
          </span>
        )}
      </Flex>
      <div className="grid gap-2">
        <div className="overflow-x-auto" ref={resize.containerRef}>
          <table
            className="resizable-table"
            style={
              resize.allWidthsKnown
                ? // Fixed layout: columnWidths fill the container (min-width 100%) or overflow it; width matches their sum.
                  { tableLayout: "fixed", width: resize.totalWidth, minWidth: "100%" }
                : hasRows
                  ? // Measuring pass: size to the natural content so columns seed at their real width, not squeezed.
                    { tableLayout: "auto", width: "max-content" }
                  : // Empty / loading: nothing to measure yet, just fill the container.
                    { tableLayout: "auto", width: "100%" }
            }
          >
            <colgroup>
              {columns.map((column, idx) => {
                const width = resize.columnWidths[columnKey(column)];
                return <col key={`col-${idx}`} style={width != null ? { width } : undefined} />;
              })}
            </colgroup>
            {columns.length > 0 && (
              <thead>
                <tr>
                  {columns.map((column, idx) => {
                    if (column.kind === "actions") {
                      return (
                        <th
                          key={`h-${idx}`}
                          ref={resize.setHeaderRef(ACTIONS_KEY)}
                          style={{ width: "1%", whiteSpace: "nowrap" }}
                        />
                      );
                    }
                    const sortable = isSortable(column);
                    const isCurrent = sortable && activeSort?.key === sortKeyOf(column);
                    return (
                      <th
                        key={`h-${idx}`}
                        ref={resize.setHeaderRef(column.header)}
                        className={`align-middle ${sortable ? "cursor-pointer hover:opacity-60" : ""}`}
                        onClick={sortable ? () => handleHeaderClick(column) : undefined}
                      >
                        <span className="th-label">
                          {column.header}
                          {sortable && (
                            <Icon
                              className={isCurrent ? "" : "opacity-0"}
                              path={activeSort?.order === "asc" ? mdiChevronUp : mdiChevronDown}
                              viewBox="0 0 18 18"
                            />
                          )}
                        </span>
                        <span
                          className="resize-handle"
                          onMouseDown={e => resize.startResize(e, column.header)}
                          onClick={e => e.stopPropagation()}
                        />
                      </th>
                    );
                  })}
                </tr>
              </thead>
            )}
            <tbody>
              {loading ? (
                Array.from({ length: Math.min(perPage, 5) }).map((_, i) => (
                  <tr key={`s-${i}`}>
                    {columns.map((_, j) => (
                      <td key={`s-${i}-${j}`}>
                        <Shimmer />
                      </td>
                    ))}
                  </tr>
                ))
              ) : visibleData.length === 0 ? (
                <tr>
                  <td
                    colSpan={columns.length || 1}
                    className="border-t-[1px] border-[color:var(--border-primary)] text-center font-thin italic opacity-50"
                  >
                    No results available
                  </td>
                </tr>
              ) : (
                visibleData.map((row, i) => (
                  <tr key={rowKey ? rowKey(row, i) : row.id}>
                    {columns.map((column, j) => {
                      if (column.kind === "actions") {
                        return (
                          <td key={j} className="sticky right-0" data-buttons-cell>
                            {column.render(row)}
                          </td>
                        );
                      }
                      return (
                        <TruncatingCell
                          key={j}
                          content={column.render(row)}
                          fixed={resize.allWidthsKnown}
                          maxWidth={column.maxWidth}
                          columnWidth={resize.columnWidths[column.header]}
                        />
                      );
                    })}
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
        <div>
          <Paginator
            currentPage={page}
            perPage={perPage}
            pagesList={pagesList}
            filteredData={filteredData}
            backendTotalRows={pagination?.totalRows}
            setCurrentPage={setPage}
            setPerPage={handlePerPage}
          />
        </div>
        <div /> {/* Empty element just to even the last element gap */}
      </div>
    </Card>
  );
}
