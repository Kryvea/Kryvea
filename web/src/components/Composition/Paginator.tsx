import { memo, useCallback } from "react";
import { v4 } from "uuid";
import Button from "../Form/Button";
import Buttons from "../Form/Buttons";
import Input from "../Form/Input";
import Flex from "./Flex";

export default function Paginator({
  currentPage,
  perPage,
  pagesList,
  filteredData,
  backendTotalRows,
  setCurrentPage,
  setPerPage,
}) {
  const getPaginatorKey = useCallback((page: number) => `paginator-${page}-${v4()}`, []);

  const isInTheMiddle = currentPage > 2 && currentPage < pagesList.length - 1;

  const isLessThan10 = pagesList.length < 10;

  const EllipsisButton = memo(() => <Button variant="secondary" small text="..." disabled onClick={() => {}} />);

  const isOverlapping = page => page < 3 || page > pagesList.length - 2;

  const totalRows = backendTotalRows ?? filteredData?.length ?? 0;
  const startIndex = totalRows === 0 ? 0 : (currentPage - 1) * perPage + 1;
  const endIndex = Math.min(currentPage * perPage, totalRows);

  return (
    <div className="hide-scrollbar flex flex-col items-center justify-between gap-4 px-3 md:flex-row md:py-0">
      <Buttons className="flex-nowrap !overflow-x-scroll">
        {isLessThan10 ? (
          pagesList.map(page => (
            <Button
              small
              className="aspect-square !max-h-9 !min-w-9 justify-center !p-0 text-center"
              variant={currentPage === page ? "tertiary" : "secondary"}
              text={page}
              disabled={page === currentPage}
              onClick={() => setCurrentPage(page)}
              key={getPaginatorKey(page)}
            />
          ))
        ) : (
          <>
            {pagesList.slice(0, isInTheMiddle ? 2 : 3).map(page => (
              <Button
                small
                className="aspect-square !max-h-9 !min-w-9 justify-center !p-0 text-center"
                variant={currentPage === page ? "tertiary" : "secondary"}
                text={page}
                disabled={page === currentPage}
                onClick={() => setCurrentPage(page)}
                key={getPaginatorKey(page)}
              />
            ))}

            {isInTheMiddle ? (
              <>
                <EllipsisButton />
                {pagesList
                  .slice(currentPage - 3, currentPage + 2)
                  .map(page =>
                    isOverlapping(page) ? null : (
                      <Button
                        small
                        className="sticky left-0 aspect-square !max-h-9 !min-w-9 justify-center !p-0 text-center"
                        variant={currentPage === page ? "tertiary" : "secondary"}
                        text={page}
                        disabled={page === currentPage}
                        onClick={() => setCurrentPage(page)}
                        key={getPaginatorKey(page)}
                      />
                    )
                  )}
                <EllipsisButton />
              </>
            ) : (
              <EllipsisButton />
            )}

            {pagesList.splice(isInTheMiddle ? -2 : -3).map(page => (
              <Button
                small
                className="sticky right-0 aspect-square !max-h-9 !min-w-9 justify-center !p-0 text-center"
                variant={currentPage === page ? "tertiary" : "secondary"}
                text={page}
                disabled={page === currentPage}
                onClick={() => setCurrentPage(page)}
                key={getPaginatorKey(page)}
              />
            ))}
          </>
        )}
      </Buttons>

      <Flex className="sticky right-0 gap-2" items="center">
        <small className="mr-2 text-[color:var(--text-secondary)]">
          {totalRows > 0 ? `Showing ${startIndex}-${endIndex} of ${totalRows} entries` : "No entries found"}
        </small>
        <Input
          type="number"
          className="mr-2 max-h-8 w-[50px] rounded-md text-center"
          id={`paginator-per-page-input-${v4()}`}
          value={perPage}
          min={1}
          onChange={setPerPage}
        />
        <small className="text-nowrap">per page</small>
      </Flex>
    </div>
  );
}
