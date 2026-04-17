import { createContext, Dispatch, SetStateAction, useCallback, useLayoutEffect, useState } from "react";
import { BrowserRouter, Navigate, Route, Routes } from "react-router";
import { ToastContainer } from "react-toastify";
import Layout from "./components/Layout/Layout";
import RouteWatcher from "./components/Layout/RouteWatcher";
import {
  AddCustomer,
  AddUser,
  Assessments,
  AssessmentUpsert,
  AssessmentVulnerabilities,
  Categories,
  CategoryUpsert,
  CustomerDetail,
  Customers,
  Dashboard,
  Login,
  Logs,
  PocsUpsert,
  Profile,
  Settings,
  Targets,
  Templates,
  Users,
  VulnerabilityDetail,
  VulnerabilitySearch,
  VulnerabilityUpsert,
} from "./pages";
import { Assessment as AssessmentObj, Category, Customer, ThemeMode, Vulnerability } from "./types/common.types";
import { getLocalStorageCtxState, GlobalContextKeys, setLocalStorageCtxState } from "./utils/contextPersistence";
import { getBrowser, SidebarItemLabel } from "./utils/helpers";

export type GlobalContextType = {
  useThemeMode: [ThemeMode, Dispatch<SetStateAction<ThemeMode>>];
  useBrowser: [string, Dispatch<SetStateAction<string>>];
  useCtxUsername: [string, Dispatch<SetStateAction<string>>];
  useFullscreen: [boolean, Dispatch<SetStateAction<boolean>>];
  useCtxAssessment: [Partial<AssessmentObj>, Dispatch<SetStateAction<Partial<AssessmentObj>>>];
  useCtxCustomer: [Customer, Dispatch<SetStateAction<Customer>>];
  useCtxVulnerability: [Partial<Vulnerability>, Dispatch<SetStateAction<Partial<Vulnerability>>>];
  useCtxCategory: [Category, Dispatch<SetStateAction<Category>>];
  useCtxLastPage: [string, Dispatch<SetStateAction<string>>];
  useCtxSelectedSidebarItemLabel: [SidebarItemLabel, Dispatch<SetStateAction<SidebarItemLabel>>];
  useCtxCodeHighlightColor: [string, Dispatch<SetStateAction<string>>];
  useCtxLinewrap: [boolean, Dispatch<SetStateAction<boolean>>];
  useCtxMinimap: [boolean, Dispatch<SetStateAction<boolean>>];
};

export const GlobalContext = createContext<GlobalContextType>(null);

export default function App() {
  const useThemeMode = useState<ThemeMode>(() => getLocalStorageCtxState("useThemeMode") ?? "os");
  const [themeMode] = useThemeMode;
  const useBrowser = useState<string>(getBrowser);
  const useCtxUsername = useState<string>(() => getLocalStorageCtxState("useCtxUsername") ?? "");
  const useFullscreen = useState(() => getLocalStorageCtxState("useFullscreen") ?? false);
  const useCtxCustomer = useState<Customer>(() => getLocalStorageCtxState("useCtxCustomer"));
  const useCtxAssessment = useState<Partial<AssessmentObj>>(() => getLocalStorageCtxState("useCtxAssessment"));
  const useCtxVulnerability = useState<Partial<Vulnerability>>(() => getLocalStorageCtxState("useCtxVulnerability"));
  const useCtxCategory = useState<Category>(() => getLocalStorageCtxState("useCtxCategory"));
  const useCtxLastPage = useState<string>(() => getLocalStorageCtxState("useCtxLastPage") ?? "/dashboard");
  const useCtxSelectedSidebarItemLabel = useState<SidebarItemLabel>(
    () => getLocalStorageCtxState("useCtxSelectedSidebarItemLabel") ?? "Dashboard"
  );
  const useCtxCodeHighlightColor = useState<string>(
    () => getLocalStorageCtxState("useCtxCodeHighlightColor") ?? "#0D542B"
  );
  const useCtxLinewrap = useState<boolean>(() => getLocalStorageCtxState("useCtxLinewrap") ?? false);
  const useCtxMinimap = useState<boolean>(() => getLocalStorageCtxState("useCtxMinimap") ?? false);

  useLayoutEffect(() => {
    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const effectiveDark = themeMode === "os" ? mediaQuery.matches : themeMode === "dark";
    document.documentElement.classList[effectiveDark ? "add" : "remove"]("dark");

    if (themeMode !== "os") {
      return;
    }

    const handler = (event: MediaQueryListEvent) => {
      document.documentElement.classList[event.matches ? "add" : "remove"]("dark");
    };

    mediaQuery.addEventListener("change", handler);
    return () => mediaQuery.removeEventListener("change", handler);
  }, [themeMode]);

  const bindToLocalStorage = useCallback(function <T>(
    [state, setState]: [T, Dispatch<SetStateAction<T>>],
    key: GlobalContextKeys
  ): [T, (newState: T | ((prevState: T) => T)) => any] {
    return [
      state,
      (value: T | ((prevState: T) => T)) => {
        if (typeof value === "function") {
          setState(prev => {
            const newValue = (value as CallableFunction)(prev);
            setLocalStorageCtxState(key, newValue);
            return newValue;
          });
          return;
        }
        setState(value);
        setLocalStorageCtxState(key, value);
      },
    ];
  }, []);

  return (
    <GlobalContext.Provider
      value={{
        useThemeMode: bindToLocalStorage(useThemeMode, "useThemeMode"),
        useBrowser,
        useCtxUsername: bindToLocalStorage(useCtxUsername, "useCtxUsername"),
        useFullscreen: bindToLocalStorage(useFullscreen, "useFullscreen"),
        useCtxCustomer: bindToLocalStorage(useCtxCustomer, "useCtxCustomer"),
        useCtxAssessment: bindToLocalStorage(useCtxAssessment, "useCtxAssessment"),
        useCtxVulnerability: bindToLocalStorage(useCtxVulnerability, "useCtxVulnerability"),
        useCtxCategory: bindToLocalStorage(useCtxCategory, "useCtxCategory"),
        useCtxLastPage: bindToLocalStorage(useCtxLastPage, "useCtxLastPage"),
        useCtxSelectedSidebarItemLabel: bindToLocalStorage(
          useCtxSelectedSidebarItemLabel,
          "useCtxSelectedSidebarItemLabel"
        ),
        useCtxCodeHighlightColor: bindToLocalStorage(useCtxCodeHighlightColor, "useCtxCodeHighlightColor"),
        useCtxLinewrap: bindToLocalStorage(useCtxLinewrap, "useCtxLinewrap"),
        useCtxMinimap: bindToLocalStorage(useCtxMinimap, "useCtxMinimap"),
      }}
    >
      <ToastContainer
        position="bottom-center"
        autoClose={3 * 1000}
        closeOnClick
        pauseOnHover
        toastClassName="kryvea-toast"
      />
      <BrowserRouter>
        <Routes>
          <Route element={<RouteWatcher />}>
            <Route element={<Layout />}>
              {/* Dashboard and Profile */}
              <Route path="/" element={<Navigate to={"/dashboard"} replace />} />
              <Route path="/dashboard" element={<Dashboard />} />
              <Route path="/profile" element={<Profile />} />

              {/* Users */}
              <Route path="/users" element={<Users />} />
              <Route path="/users/new" element={<AddUser />} />

              {/* Customers */}
              <Route path="/customers" element={<Customers />} />
              <Route path="/customers/new" element={<AddCustomer />} />
              <Route path="/customers/:customerId" element={<CustomerDetail />} />
              <Route path="/customers/:customerId/targets" element={<Targets />} />
              <Route path="/customers/:customerId/assessments" element={<Assessments />} />
              <Route path="/customers/:customerId/assessments/new" element={<AssessmentUpsert />} />
              <Route path="/customers/:customerId/assessments/:assessmentId" element={<AssessmentUpsert />} />

              {/* Assessments */}
              <Route
                path="/customers/:customerId/assessments/:assessmentId/vulnerabilities"
                element={<AssessmentVulnerabilities />}
              />
              <Route
                path="/customers/:customerId/assessments/:assessmentId/vulnerabilities/new"
                element={<VulnerabilityUpsert />}
              />
              <Route
                path="/customers/:customerId/assessments/:assessmentId/vulnerabilities/:vulnerabilityId"
                element={<VulnerabilityDetail />}
              />
              <Route
                path="/customers/:customerId/assessments/:assessmentId/vulnerabilities/:vulnerabilityId/edit"
                element={<VulnerabilityUpsert />}
              />

              {/* Vulnerabilities */}
              <Route path="/vulnerability_search" element={<VulnerabilitySearch />} />
              <Route
                path="/customers/:customerId/assessments/:assessmentId/vulnerabilities/:vulnerabilityId/pocs"
                element={<PocsUpsert />}
              />

              {/* Categories */}
              <Route path="/categories" element={<Categories />} />
              <Route path="/categories/new" element={<CategoryUpsert />} />
              <Route path="/categories/:categoryId" element={<CategoryUpsert />} />

              {/* Other */}
              <Route path="/logs" element={<Logs />} />
              <Route path="/templates" element={<Templates />} />
              <Route path="/settings" element={<Settings />} />
            </Route>
            <Route path="/login" element={<Login />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </GlobalContext.Provider>
  );
}
