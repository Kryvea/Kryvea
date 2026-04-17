import {
  mdiAccountEdit,
  mdiAccountMultiple,
  mdiCog,
  mdiCogs,
  mdiDomain,
  mdiFileChart,
  mdiListBox,
  mdiMagnify,
  mdiMathLog,
  mdiShapePlus,
  mdiTabSearch,
  mdiTarget,
  mdiViewDashboard,
} from "@mdi/js";
import { NavigateFunction } from "react-router";
import { getKryveaShadow } from "../api/cookie";
import { Customer, Vulnerability } from "../types/common.types";
import { Keys } from "../types/utils.types";
import { appTitle, USER_ROLE_ADMIN } from "./constants";

export const getPageTitle = (currentPageTitle: string) => `${currentPageTitle} - ${appTitle}`;

export function getBrowser() {
  if (navigator.userAgent.indexOf("Chrome") != -1) {
    return "Chrome";
  } else if (navigator.userAgent.indexOf("Opera") != -1) {
    return "Opera";
  } else if (navigator.userAgent.indexOf("MSIE") != -1) {
    return "IE";
  } else if (navigator.userAgent.indexOf("Firefox") != -1) {
    return "Firefox";
  } else {
    return "unknown";
  }
}

export type SidebarItemLabel =
  | "Dashboard"
  | "Customers"
  | "Assessments"
  | "Targets"
  | "Edit Customer"
  | "Vulnerability Search"
  | "Administration"
  | "Categories"
  | "Users"
  | "Logs"
  | "Report Templates"
  | "Settings";
export type SidebarItem = {
  icon: string;
  label: SidebarItemLabel;
  href?: string;
  onClick?: () => void;
  menu?: SidebarItem[];
};
export const getSidebarItems: (ctxCustomer: Customer, navigate: NavigateFunction) => SidebarItem[] = ctxCustomer => [
  { href: "/dashboard", icon: mdiViewDashboard, label: "Dashboard" },
  { href: "/customers", icon: mdiListBox, label: "Customers" },
  ...(ctxCustomer != null
    ? [
        {
          label: ctxCustomer.name,
          icon: mdiDomain,
          href: `/customers/${ctxCustomer.id}/assessments`,
          menu: [
            {
              href: `/customers/${ctxCustomer.id}/assessments`,
              icon: mdiTabSearch,
              label: "Assessments",
            },
            { href: `/customers/${ctxCustomer.id}/targets`, icon: mdiTarget, label: "Targets" },
            {
              href: `/customers/${ctxCustomer.id}`,
              icon: mdiAccountEdit,
              label: "Edit Customer",
            },
          ],
        } as SidebarItem & { label: string }, // ctxCustomer name is not castable as SidebarItemLabel
      ]
    : []),
  { href: "/vulnerability_search", icon: mdiMagnify, label: "Vulnerability Search" },
  getKryveaShadow() === USER_ROLE_ADMIN && {
    label: "Administration",
    icon: mdiCogs,
    menu: [
      { href: "/categories", icon: mdiShapePlus, label: "Categories" },
      { href: "/users", icon: mdiAccountMultiple, label: "Users" },
      { href: "/logs", icon: mdiMathLog, label: "Logs" },
      { href: "/templates", icon: mdiFileChart, label: "Report Templates" },
      { href: "/settings", icon: mdiCog, label: "Settings" },
    ],
  },
];

export function formatVulnerabilityTitle(vulnerability: Partial<Vulnerability>): string {
  if (!vulnerability || !vulnerability.category) return "";

  const { category, detailed_title } = vulnerability;
  const { identifier, name, subcategory } = category;

  let title = `${identifier} - ${name}`;
  if (subcategory) title += ` - ${subcategory}`;
  if (detailed_title) title += ` (${detailed_title})`;

  return title;
}

export function prettifyJsonBody(http: string): [string, number] {
  let prettyHttp = "";
  let jsonStartingLine = 0xabadcafe;

  http = http?.replaceAll("\r\n", "\n");

  const [httpHeaders, httpBody] = (http ?? "")?.split("\n\n");
  try {
    const prettyJson = JSON.stringify(JSON.parse(httpBody), null, 2);
    prettyHttp = `${httpHeaders}\n\n${prettyJson}`;
    jsonStartingLine = (`${httpHeaders}\n\n`.match(/\n/g) || []).length + 2;
  } catch (e) {
    prettyHttp = http;
  }

  return [prettyHttp, jsonStartingLine];
}

export function onelineJsonBody(http: string): [string, number] {
  let oneLineHttpBody = "";
  let jsonStartingLine = 0xabadcafe;

  http = http?.replaceAll("\r\n", "\n");

  const [httpHeaders, httpBody] = (http ?? "")?.split("\n\n");
  try {
    const jsonOneline = JSON.stringify(JSON.parse(httpBody), null, 0);
    oneLineHttpBody = `${httpHeaders}\n\n${jsonOneline}`;
    jsonStartingLine = (`${httpHeaders}\n\n`.match(/\n/g) || []).length + 2;
  } catch (e) {
    oneLineHttpBody = http;
  }

  return [oneLineHttpBody, jsonStartingLine];
}

export function emptyCurry() {
  return () => {};
}

type SortByOptions = {
  caseInsensitive?: boolean;
  reverse?: boolean;
};
export const sortBy =
  <T>(property: Keys<T>, options?: SortByOptions) =>
  (a: T, b: T): number => {
    let aValue = a[property] as any;
    let bValue = b[property] as any;

    if (options?.caseInsensitive) {
      aValue = aValue.toLowerCase();
      bValue = bValue.toLowerCase();
    }

    if (options?.reverse) {
      if (aValue > bValue) {
        return -1;
      }
      if (aValue < bValue) {
        return 1;
      }
      return 0;
    }
    if (aValue < bValue) {
      return -1;
    }
    if (aValue > bValue) {
      return 1;
    }
    return 0;
  };
