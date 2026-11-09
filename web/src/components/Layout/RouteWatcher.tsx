import { useContext, useEffect, useRef } from "react";
import { Outlet, useLocation, useNavigate, useParams } from "react-router";
import { getData, setNavigate } from "../../api/api";
import { getKryveaShadow } from "../../api/cookie";
import { GlobalContext } from "../../App";
import { Assessment, Category, Customer, Vulnerability } from "../../types/common.types";

export default function RouteWatcher() {
  const {
    useCtxCustomer: [ctxCustomer, setCtxCustomer],
    useCtxAssessment: [ctxAssessment, setCtxAssessment],
    useCtxVulnerability: [ctxVulnerability, setCtxVulnerability],
    useCtxCategory: [ctxCategory, setCtxCategory],
    useCtxLastPage: [, setCtxLastPage],
  } = useContext(GlobalContext);
  const location = useLocation();
  const previousPath = useRef(location.pathname);
  const navigate = useNavigate();
  const { customerId, assessmentId, vulnerabilityId, categoryId } = useParams();

  useEffect(() => {
    setNavigate(navigate);
  }, [navigate]);

  useEffect(() => {
    if (customerId != undefined && ctxCustomer?.id !== customerId) {
      getData<Customer>(`/api/customers/${customerId}`, setCtxCustomer);
    }
    if (assessmentId != undefined && ctxAssessment?.id !== assessmentId) {
      getData<Assessment>(`/api/assessments/${assessmentId}`, setCtxAssessment);
    }
    if (vulnerabilityId != undefined && ctxVulnerability?.id !== vulnerabilityId) {
      getData<Vulnerability>(`/api/vulnerabilities/${vulnerabilityId}`, setCtxVulnerability);
    }
    if (categoryId != undefined && ctxCategory?.id !== categoryId) {
      getData<Category>(`/api/categories/${categoryId}`, setCtxCategory);
    }
  }, [customerId, assessmentId, vulnerabilityId, categoryId]);

  useEffect(() => {
    const kryveaShadow = getKryveaShadow();
    if ((!kryveaShadow || kryveaShadow === "password_expired") && location.pathname !== "/login") {
      navigate("/login");
      return;
    }

    if (location.pathname !== "/login" && location.pathname !== "/") {
      setCtxLastPage(location.pathname);
    }

    const currentPathParent = location.pathname.split("/")[1];
    if (currentPathParent === previousPath.current.split("/")[1]) {
      return;
    }

    switch (currentPathParent) {
      case "dashboard":
      case "customers":
      case "assessments":
      case "vulnerability_search":
      case "login":
        break;
      default:
        setCtxCustomer(undefined);
        setCtxAssessment(undefined);
        setCtxVulnerability(undefined);
        setCtxCategory(undefined);
        break;
    }
    previousPath.current = location.pathname;
  }, [location]); // Runs whenever the location changes

  return <Outlet />;
}
