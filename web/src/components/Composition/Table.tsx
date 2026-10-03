import { mdiChevronDown, mdiChevronUp, mdiClose } from "@mdi/js";
import {
  MouseEvent as ReactMouseEvent,
  ReactNode,
  isValidElement,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { SortState } from "../../types/utils.types";
import Card from "./Card";
import Flex from "./Flex";
import Icon from "./Icon";
import Paginator from "./Paginator";
import Shimmer from "./Shimmer";

type DataColumn<Row> = {
  kind?: "data";
  header: string;
  render: (row: Row) => ReactNode;
  sortKey?: string;
  sortValue?: (row: Row) => string | number | Date;
  maxWidth?: string;
  // Seed this column at its header width instead of its content width (still resizable afterwards).
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
  // Called when the backend reports fewer pages than the current one (e.g. a stale deep link); defaults to onPageChange.
  onPageOutOfRange?: (lastPage: number) => void;
}

type RowWithId = { id: string | number };

type BaseTableProps<Row extends RowWithId> = {
  columns: Column<Row>[];
  data: Row[];
  loading?: boolean;
  perPage?: number;
  // Stable id used to persist per-column widths in localStorage. Omit to keep widths only for the session.
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
  // Only ever called on data-column headers, which always render a .th-label.
  const label = th.querySelector<HTMLElement>(".th-label")!;
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

function saveColumnWidths(tableId: string | undefined, widths: Record<string, number>, validKeys: string[]) {
  if (!tableId) {
    return;
  }
  // Persist only the current columns, so renamed/removed ones don't leave orphan entries.
  const pruned: Record<string, number> = {};
  for (const key of validKeys) {
    if (widths[key] != null) {
      pruned[key] = widths[key];
    }
  }
  try {
    localStorage.setItem(WIDTHS_PREFIX + tableId, JSON.stringify(pruned));
  } catch {
    // ignore quota / serialization errors
  }
}

function useControllableState<T>(
  isControlled: boolean,
  controlledValue: T,
  onControlledChange: ((value: T) => void) | undefined,
  initial: T
): [T, (value: T) => void] {
  const [internal, setInternal] = useState(initial);
  const setValue = (next: T) => {
    onControlledChange?.(next);
    if (!isControlled) {
      setInternal(next);
    }
  };
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

  // Callers rebuild the columns array inline on every render, so these memos key
  // on the column signature instead of the array identity or they would never hit.
  const colSignature = columns.map(columnKey).join("|");

  // Render each row to searchable text once per dataset (not once per keystroke), and only while searching.
  // Texts are kept per column so the query must match within a single column.
  const searching = !serverMode && !!query;
  const searchableText = useMemo(() => {
    if (!searching) {
      return [];
    }
    return data.map(row =>
      columns.filter(column => column.kind !== "actions").map(column => extractText(column.render(row)).toLowerCase())
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searching, data, colSignature]);

  const filteredData = useMemo(() => {
    if (!searching) {
      return data ?? [];
    }
    const q = query.toLowerCase();
    return data.filter((_, index) => searchableText[index].some(text => text.includes(q))) ?? [];
  }, [searching, query, data, searchableText]);

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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [serverMode, filteredData, sort, colSignature]);

  const numPages = useMemo(() => {
    if (serverMode) {
      return totalPages ?? 0;
    }
    const pages = Math.ceil(filteredData.length / perPage);
    return Number.isNaN(pages) ? 0 : pages;
  }, [serverMode, totalPages, filteredData.length, perPage]);

  // If the dataset shrinks below the current page, clamp instead of rendering an empty page.
  const clampedPage = serverMode ? page : Math.min(page, Math.max(numPages, PAGE_FLOOR));

  // Slicing is separate from sorting so paging through results doesn't re-sort the whole dataset.
  const visibleData = useMemo(() => {
    if (serverMode) {
      return sortedData;
    }
    return sortedData.slice(perPage * (clampedPage - PAGE_FLOOR), perPage * clampedPage);
  }, [serverMode, sortedData, clampedPage, perPage]);

  return { filteredData, visibleData, numPages, page: clampedPage };
}

function useResizableColumns<Row>(columns: Column<Row>[], tableId?: string, ready?: boolean) {
  const [widths, setWidths] = useState<Record<string, number>>(() => loadColumnWidths(tableId));
  const containerRef = useRef<HTMLDivElement>(null);
  const tableRef = useRef<HTMLTableElement>(null);
  const headerRefs = useRef<Record<string, HTMLTableCellElement | null>>({});
  const colRefs = useRef<Record<string, HTMLTableColElement | null>>({});
  const headerRefSetters = useRef<Record<string, (el: HTMLTableCellElement | null) => void>>({});
  const colRefSetters = useRef<Record<string, (el: HTMLTableColElement | null) => void>>({});
  const draggedRef = useRef(false);
  const endDragRef = useRef<(() => void) | null>(null);

  // Unmounting mid-drag would otherwise leave the document listeners and body styles in place.
  useEffect(() => () => endDragRef.current?.(), []);

  // Seed widths only once real rows are present: an empty body would size columns from the headers alone.
  const colSignature = columns.map(columnKey).join("|");
  useLayoutEffect(() => {
    if (!ready) {
      return;
    }
    const dataKeys = columns.filter(column => column.kind !== "actions").map(columnKey);
    const isFreshTable = dataKeys.every(key => widths[key] == null);
    const next = { ...widths };
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
    if (!changed) {
      return;
    }
    if (isFreshTable) {
      // One-time stretch so a table with no stored widths fills its container.
      const containerWidth = containerRef.current?.clientWidth ?? 0;
      const sumAll = columns.reduce((sum, column) => sum + (next[columnKey(column)] ?? 0), 0);
      const sumData = dataKeys.reduce((sum, key) => sum + (next[key] ?? 0), 0);
      if (containerWidth > sumAll && sumData > 0 && dataKeys.length > 0) {
        const targetData = containerWidth - (sumAll - sumData);
        let used = 0;
        dataKeys.forEach((key, index) => {
          next[key] =
            index === dataKeys.length - 1 ? targetData - used : Math.round(targetData * (next[key] / sumData));
          used += next[key];
        });
      }
    }
    setWidths(next);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [colSignature, ready]);

  const allWidthsKnown = columns.every(column => widths[columnKey(column)] != null);
  const totalWidth = allWidthsKnown ? columns.reduce((sum, column) => sum + widths[columnKey(column)], 0) : undefined;

  // Stable ref callbacks per key so refs aren't detached and re-attached on every render.
  const setHeaderRef = (key: string) => {
    if (!headerRefSetters.current[key]) {
      headerRefSetters.current[key] = el => {
        headerRefs.current[key] = el;
      };
    }
    return headerRefSetters.current[key];
  };
  const setColRef = (key: string) => {
    if (!colRefSetters.current[key]) {
      colRefSetters.current[key] = el => {
        colRefs.current[key] = el;
      };
    }
    return colRefSetters.current[key];
  };

  const startResize = (event: ReactMouseEvent, key: string) => {
    // Before widths are seeded (empty table) a drag would write a bogus tiny
    // width on the table element that React never rewrites; ignore it.
    if (totalWidth == null) {
      return;
    }
    event.preventDefault();
    // The handle sits directly inside its data-column <th>.
    const th = (event.currentTarget as HTMLElement).parentElement!;
    const startWidth = widths[key];
    // A column can't shrink below its header content, otherwise the header would overlap the next column.
    const minWidth = Math.max(MIN_COLUMN_WIDTH, headerContentWidth(th));
    const startX = event.clientX;
    const startTotal = totalWidth;
    let finalWidth = startWidth;
    draggedRef.current = false;

    const onMove = (moveEvent: MouseEvent) => {
      // A mouseup outside the window is never delivered; end the drag on the first buttonless move.
      if (moveEvent.buttons === 0) {
        onUp();
        return;
      }
      if (moveEvent.clientX !== startX) {
        draggedRef.current = true;
      }
      finalWidth = Math.max(minWidth, Math.round(startWidth + moveEvent.clientX - startX));
      // Resize the <col> imperatively so mousemove doesn't re-render the table.
      const colEl = colRefs.current[key];
      if (colEl) {
        colEl.style.width = `${finalWidth}px`;
      }
      if (tableRef.current) {
        tableRef.current.style.width = `${startTotal - startWidth + finalWidth}px`;
      }
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
      // A click without movement would commit identical widths and re-render for nothing.
      if (!draggedRef.current) {
        return;
      }
      const next = { ...widths, [key]: finalWidth };
      saveColumnWidths(tableId, next, columns.map(columnKey));
      setWidths(next);
      // Let the click that follows mouseup read the drag flag (to skip sorting) before it is reset.
      setTimeout(() => {
        draggedRef.current = false;
      }, 0);
    };

    endDragRef.current = cleanup;
    document.body.style.userSelect = "none";
    document.body.style.cursor = "col-resize";
    document.addEventListener("mousemove", onMove);
    document.addEventListener("mouseup", onUp);
  };

  return {
    widths,
    allWidthsKnown,
    totalWidth,
    containerRef,
    tableRef,
    setHeaderRef,
    setColRef,
    startResize,
    isResizing: () => draggedRef.current,
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
  // Depend on the extracted text, not the ReactNode: render() returns a fresh
  // element identity every render, which would force a layout read per cell.
  const text = extractText(content);
  useLayoutEffect(() => {
    const el = ref.current;
    setTruncated(!!el && el.scrollWidth > el.clientWidth);
  }, [text, fixed, maxWidth, columnWidth]);
  return (
    <td className="text-nowrap">
      <div
        ref={ref}
        className="truncate-cell"
        style={!fixed && maxWidth != null ? { width: maxWidth } : undefined}
        title={truncated ? text || undefined : undefined}
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

  const {
    filteredData,
    visibleData,
    numPages,
    page: currentPage,
  } = useTableData({
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

  // Keyed on data so it only reacts to a fresh backend response: totalPages is stale while a fetch is in flight
  // (e.g. navigating back to a URL with another page size), and acting on it then would redirect wrongly.
  useEffect(() => {
    if (serverMode && numPages > 0 && currentPage > numPages) {
      (pagination.onPageOutOfRange ?? pagination.onPageChange)(numPages);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data]);

  const resize = useResizableColumns(columns, tableId, hasRows);

  const totalRows = pagination?.totalRows ?? filteredData.length;

  const sortKeyOf = (column: DataColumn<Row>) => (serverMode ? column.sortKey : (column.sortKey ?? column.header));
  const isSortable = (column: DataColumn<Row>) => !serverMode || !!column.sortKey;

  const handleSearch = (value: string) => {
    // Controlled search resets the page through the parent; only reset it ourselves in client mode.
    if (!search) {
      setPage(PAGE_FLOOR);
    }
    setQuery(value);
  };

  const handlePerPage = (value: number) => {
    // Controlled pagination resets the page through the parent; only reset it ourselves in client mode.
    if (!pagination) {
      setPage(PAGE_FLOOR);
    }
    setPerPage(value);
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

  // Empty column inserted before the actions column; it absorbs leftover container
  // width so resizing one column never changes the others.
  const actionsIndex = columns.findIndex(column => column.kind === "actions");
  const fillerIndex = actionsIndex === -1 ? columns.length : actionsIndex;
  const withFiller = (cells: ReactNode[], filler: ReactNode) => [
    ...cells.slice(0, fillerIndex),
    filler,
    ...cells.slice(fillerIndex),
  ];

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
      <div className="grid gap-2 pb-2">
        <div className="overflow-x-auto" ref={resize.containerRef}>
          <table
            ref={resize.tableRef}
            className="resizable-table"
            style={
              resize.allWidthsKnown
                ? { tableLayout: "fixed", width: resize.totalWidth, minWidth: "100%" }
                : hasRows
                  ? // Measuring pass: size to natural content so columns seed at their real width.
                    { tableLayout: "auto", width: "max-content" }
                  : { tableLayout: "auto", width: "100%" }
            }
          >
            <colgroup>
              {withFiller(
                columns.map((column, idx) => {
                  const key = columnKey(column);
                  const width = resize.widths[key];
                  return (
                    <col key={`col-${idx}`} ref={resize.setColRef(key)} style={width != null ? { width } : undefined} />
                  );
                }),
                <col key="col-filler" />
              )}
            </colgroup>
            <thead>
              <tr>
                {withFiller(
                  columns.map((column, idx) => {
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
                  }),
                  <th key="h-filler" aria-hidden className="!p-0" />
                )}
              </tr>
            </thead>
            <tbody>
              {loading ? (
                Array.from({ length: Math.min(perPage, 5) }).map((_, i) => (
                  <tr key={`s-${i}`}>
                    {withFiller(
                      columns.map((_, j) => (
                        <td key={`s-${i}-${j}`}>
                          <Shimmer />
                        </td>
                      )),
                      <td key={`s-${i}-filler`} aria-hidden data-filler className="!p-0" />
                    )}
                  </tr>
                ))
              ) : visibleData.length === 0 ? (
                <tr>
                  <td
                    colSpan={columns.length + 1}
                    className="border-t-[1px] border-[color:var(--border-primary)] text-center font-thin italic opacity-50"
                  >
                    No results available
                  </td>
                </tr>
              ) : (
                visibleData.map(row => (
                  <tr key={row.id}>
                    {withFiller(
                      columns.map((column, j) => {
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
                            columnWidth={resize.widths[column.header]}
                          />
                        );
                      }),
                      <td key="filler" aria-hidden data-filler className="!p-0" />
                    )}
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
        <Paginator
          currentPage={currentPage}
          numPages={numPages}
          perPage={perPage}
          totalRows={totalRows}
          setCurrentPage={setPage}
          setPerPage={handlePerPage}
        />
      </div>
    </Card>
  );
}
