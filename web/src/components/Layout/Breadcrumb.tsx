import React, { ReactNode, useContext, useEffect, useMemo, useRef } from "react";
import { Link, useResolvedPath } from "react-router";
import { GlobalContext } from "../../App";
import { scrollElementHorizontally } from "../../hooks/hooks";
import { formatVulnerabilityTitle } from "../../utils/helpers";

type TBreadCrumbProps = {
  homeElement: ReactNode;
  separator: ReactNode;
  activeClasses?: string;
  capitalizeLinks?: boolean;
};

export default function Breadcrumb({ homeElement, separator, capitalizeLinks }: TBreadCrumbProps) {
  const {
    useCtxCustomer: [ctxCustomer],
    useCtxAssessment: [ctxAssessment],
    useCtxVulnerability: [ctxVulnerability],
    useCtxCategory: [ctxCategory],
  } = useContext(GlobalContext);

  const breadcrumbUl = useRef<HTMLUListElement>(null);

  useEffect(scrollElementHorizontally(breadcrumbUl), []);

  const IdNameTuples = useMemo(
    () => [
      [ctxCustomer?.id, ctxCustomer?.name],
      [ctxAssessment?.id, ctxAssessment?.name],
      [ctxVulnerability?.id, formatVulnerabilityTitle(ctxVulnerability)],
      [ctxCategory?.id, ctxCategory?.name],
    ],
    [ctxCustomer, ctxAssessment, ctxVulnerability, ctxCategory]
  ); // will be filled as we go on building breadcrumbs with IDs

  const pathNames = useResolvedPath(undefined)
    .pathname.split("/")
    .filter(path => path);

  return (
    <ul ref={breadcrumbUl} className="flex gap-1 overflow-x-auto scroll-smooth">
      <li className={"hover:underline"}>
        <Link to={"/"}>{homeElement}</Link>
      </li>
      {pathNames.map((link, index) => {
        // Replace underscores with spaces
        const displayName = link.replace(/_/g, " ");
        // Capitalize links if required
        let itemLink = capitalizeLinks ? displayName.replace(/\b\w/g, char => char.toUpperCase()) : displayName;
        for (const [id, name] of IdNameTuples) {
          if (link === id) {
            itemLink = name;
            break;
          }
        }

        // Generate the href
        const href = `/${pathNames.slice(0, index + 1).join("/")}`;

        const isLast = pathNames.length === index + 1;

        return (
          <React.Fragment key={`breadcrumb-path-${index}`}>
            <li>
              {separator}
              {isLast ? (
                <span className="hover:no-underline">{itemLink}</span>
              ) : (
                <Link className="hover:underline" to={href}>
                  {itemLink}
                </Link>
              )}
            </li>
          </React.Fragment>
        );
      })}
    </ul>
  );
}
