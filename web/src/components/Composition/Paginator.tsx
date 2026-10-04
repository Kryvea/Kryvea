import Button from "../Form/Button";
import Buttons from "../Form/Buttons";
import Input from "../Form/Input";
import Flex from "./Flex";

interface PaginatorProps {
  currentPage: number;
  numPages: number;
  perPage: number;
  totalRows: number;
  setCurrentPage: (page: number) => void;
  setPerPage: (perPage: number) => void;
}

const ELLIPSIS = "...";

const range = (from: number, to: number) => Array.from({ length: Math.max(0, to - from + 1) }, (_, i) => from + i);

// First 3, the 3 around the current page, and last 3; an ellipsis goes only where the
// shown pages actually skip a number, so it never sits between adjacent pages.
function pageItems(currentPage: number, numPages: number): (number | typeof ELLIPSIS)[] {
  if (numPages < 10) {
    return range(1, numPages);
  }
  const wanted = new Set([
    ...range(1, 3),
    ...range(currentPage - 1, currentPage + 1),
    ...range(numPages - 2, numPages),
  ]);
  const pages = [...wanted].filter(page => page >= 1 && page <= numPages).sort((a, b) => a - b);
  const items: (number | typeof ELLIPSIS)[] = [];
  pages.forEach((page, i) => {
    if (i > 0 && page - pages[i - 1] > 1) {
      items.push(ELLIPSIS);
    }
    items.push(page);
  });
  return items;
}

export default function Paginator({
  currentPage,
  numPages,
  perPage,
  totalRows,
  setCurrentPage,
  setPerPage,
}: PaginatorProps) {
  const startIndex = totalRows === 0 ? 0 : (currentPage - 1) * perPage + 1;
  const endIndex = Math.min(currentPage * perPage, totalRows);

  return (
    <div className="hide-scrollbar flex flex-col items-center justify-between gap-4 px-3 md:flex-row md:py-0">
      <Buttons className="flex-nowrap !overflow-x-scroll">
        {pageItems(currentPage, numPages).map((item, i) =>
          item === ELLIPSIS ? (
            <Button key={`e-${i}`} variant="secondary" small text={ELLIPSIS} disabled onClick={() => {}} />
          ) : (
            <Button
              key={item}
              small
              className="aspect-square !max-h-9 !min-w-9 justify-center !p-0 text-center"
              variant={currentPage === item ? "tertiary" : "secondary"}
              text={String(item)}
              disabled={item === currentPage}
              onClick={() => setCurrentPage(item)}
            />
          )
        )}
      </Buttons>

      <Flex className="sticky right-0 gap-2" items="center">
        <small className="mr-2 text-[color:var(--text-secondary)]">
          {totalRows > 0 ? `Showing ${startIndex}-${endIndex} of ${totalRows} entries` : "No entries found"}
        </small>
        <Input
          type="number"
          className="mr-2 max-h-8 w-[50px] rounded-md text-center"
          value={perPage}
          min={1}
          onChange={setPerPage}
        />
        <small className="text-nowrap">per page</small>
      </Flex>
    </div>
  );
}
