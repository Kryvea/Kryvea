import { mdiChevronDown, mdiChevronUp, mdiClose } from "@mdi/js";
import { ReactNode, isValidElement, useMemo, useState } from "react";
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

type TableProps<Row extends RowWithId> = {
  columns: Column<Row>[];
  data: Row[];
  loading?: boolean;
  rowKey?: (row: Row, index: number) => string | number;
  perPage?: number;
  search?: SearchProps;
  pagination?: PaginationProps;
  sort?: SortState;
  onSortChange?: (next: SortState | undefined) => void;
  defaultSort?: SortState;
};

const PAGE_FLOOR = 1;

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

function cycleSort(prev: SortState | undefined, key: string): SortState | undefined {
  if (!prev || prev.key !== key) {
    return { key, order: "asc" };
  }
  if (prev.order === "asc") {
    return { key, order: "desc" };
  }
  return undefined;
}

export default function Table<Row extends RowWithId>({
  columns,
  data,
  loading,
  perPage: perPageInitial = 5,
  search,
  pagination,
  sort,
  onSortChange,
  defaultSort,
  rowKey,
}: TableProps<Row>) {
  const [filterText, setFilterText] = useState("");
  const [clientPage, setClientPage] = useState(PAGE_FLOOR);
  const [clientPerPage, setClientPerPage] = useState(perPageInitial);
  const [clientSort, setClientSort] = useState<SortState | undefined>(defaultSort);

  const isServerMode = pagination?.totalRows !== undefined;
  const effectiveSearch = search?.value ?? filterText;
  const effectivePage = pagination?.page ?? clientPage;
  const effectivePerPage = pagination?.perPage ?? clientPerPage;
  const effectiveSort = sort ?? clientSort;

  const sortKeyOf = (column: DataColumn<Row>) =>
    isServerMode ? column.sortKey : (column.sortKey ?? column.header);

  const isSortable = (column: DataColumn<Row>) =>
    isServerMode ? !!column.sortKey : column.sortable !== false;

  const sortColumn = useMemo(() => {
    if (!effectiveSort) {
      return undefined;
    }
    return columns.find(column => {
      if (column.kind === "actions") {
        return false;
      }
      return sortKeyOf(column) === effectiveSort.key;
    }) as DataColumn<Row> | undefined;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [columns, effectiveSort, isServerMode]);

  const filteredData = useMemo(() => {
    if (isServerMode) {
      return data;
    }
    const query = effectiveSearch.toLowerCase();
    if (!query) {
      return data;
    }
    return data.filter(row =>
      columns.some(column => column.kind !== "actions" && extractText(column.render(row)).toLowerCase().includes(query))
    );
  }, [data, columns, effectiveSearch, isServerMode]);

  const sortedData = useMemo(() => {
    if (isServerMode) {
      return filteredData;
    }
    if (!effectiveSort || !sortColumn) {
      return filteredData;
    }
    const direction = effectiveSort.order === "asc" ? 1 : -1;
    const getValue = sortColumn.sortValue ?? ((row: Row) => extractText(sortColumn.render(row)));
    return [...filteredData].sort((a, b) => {
      const valueA = getValue(a);
      const valueB = getValue(b);
      if (typeof valueA === "string" && typeof valueB === "string") {
        return valueA.localeCompare(valueB, undefined, { sensitivity: "base", numeric: true }) * direction;
      }
      if (valueA === valueB) return 0;
      return valueA > valueB ? direction : -direction;
    });
  }, [filteredData, effectiveSort, sortColumn, isServerMode]);

  const visibleData = useMemo(() => {
    if (isServerMode) {
      return sortedData;
    }
    return sortedData.slice(effectivePerPage * (effectivePage - PAGE_FLOOR), effectivePerPage * effectivePage);
  }, [sortedData, effectivePage, effectivePerPage, isServerMode]);

  const numPages = useMemo(() => {
    if (isServerMode) {
      return pagination!.totalPages ?? 0;
    }
    const totalPages = Math.ceil(filteredData.length / effectivePerPage);
    return isNaN(totalPages) ? 0 : totalPages;
  }, [isServerMode, pagination, filteredData.length, effectivePerPage]);

  const pagesList = useMemo(() => Array.from({ length: numPages }, (_, i) => i + PAGE_FLOOR), [numPages]);

  const handleSearchChange = (v: string) => {
    if (search) {
      search.onChange(v);
      return;
    }
    setClientPage(PAGE_FLOOR);
    setFilterText(v);
  };

  const handleHeaderClick = (column: DataColumn<Row>) => {
    const key = sortKeyOf(column);
    if (!key) return;
    const next = cycleSort(effectiveSort, key);
    if (onSortChange) {
      onSortChange(next);
      return;
    }
    setClientSort(next);
  };

  const indicatorKey = (column: DataColumn<Row>) => sortKeyOf(column);

  const setCurrentPage = (p: number) => {
    if (pagination) {
      pagination.onPageChange(p);
      return;
    }
    setClientPage(p);
  };

  const setPerPage = (n: number) => {
    if (pagination) {
      pagination.onPerPageChange(n);
      pagination.onPageChange(PAGE_FLOOR);
      return;
    }
    setClientPage(PAGE_FLOOR);
    setClientPerPage(n);
  };

  return (
    <Card className="!relative !gap-0 !p-0">
      <Flex className="px-2 pt-1" items="center">
        <input
          className="w-full rounded-t-2xl bg-transparent focus:border-transparent"
          placeholder="Search"
          type="text"
          value={effectiveSearch}
          onChange={e => handleSearchChange(e.target.value)}
        />
        {effectiveSearch !== "" && (
          <span onClick={() => handleSearchChange("")}>
            <Icon className="text-[color:--text-secondary] hover:opacity-50" path={mdiClose} size={18} />
          </span>
        )}
      </Flex>
      <div className="grid gap-2">
        <div className="overflow-x-auto">
          <table className="w-full">
            {columns.length > 0 && (
              <thead>
                <tr>
                  {columns.map((column, idx) => {
                    if (column.kind === "actions") {
                      return <th key={`h-${idx}`} style={{ width: "1%", whiteSpace: "nowrap" }} />;
                    }
                    const sortable = isSortable(column);
                    const isCurrent = sortable && effectiveSort?.key === indicatorKey(column);
                    return (
                      <th
                        key={`h-${idx}`}
                        className={`align-middle ${sortable ? "cursor-pointer hover:opacity-60" : ""}`}
                        onClick={sortable ? () => handleHeaderClick(column) : undefined}
                      >
                        {column.header}
                        {sortable && (
                          <Icon
                            className={isCurrent ? "" : "opacity-0"}
                            path={effectiveSort?.order === "asc" ? mdiChevronUp : mdiChevronDown}
                            viewBox="0 0 18 18"
                          />
                        )}
                      </th>
                    );
                  })}
                </tr>
              </thead>
            )}
            <tbody>
              {loading ? (
                Array.from({ length: Math.min(effectivePerPage, 5) }).map((_, i) => (
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
                      const content = column.render(row);
                      if (column.maxWidth) {
                        const text = extractText(content);
                        return (
                          <td key={j} className="text-nowrap">
                            <div
                              style={{
                                maxWidth: column.maxWidth,
                                whiteSpace: "nowrap",
                                overflow: "hidden",
                                textOverflow: "ellipsis",
                              }}
                              title={text || undefined}
                            >
                              {content}
                            </div>
                          </td>
                        );
                      }
                      return (
                        <td key={j} className="text-nowrap">
                          {content}
                        </td>
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
            currentPage={effectivePage}
            perPage={effectivePerPage}
            pagesList={pagesList}
            filteredData={filteredData}
            backendTotalRows={pagination?.totalRows}
            setCurrentPage={setCurrentPage}
            setPerPage={setPerPage}
          />
        </div>
        <div /> {/* Empty element just to even the last element gap */}
      </div>
    </Card>
  );
}

