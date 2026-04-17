import { useContext, useEffect, useRef } from "react";
import { Outlet, useLocation, useNavigate, useParams } from "react-router";
import { toast } from "react-toastify";
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
      getData<Customer>(`/api/customers/${customerId}`, setCtxCustomer, () =>
        toast.error("Could not get customer by id: " + customerId)
      );
    }
    if (assessmentId != undefined && ctxAssessment?.id !== assessmentId) {
      getData<Assessment>(`/api/assessments/${assessmentId}`, setCtxAssessment, () =>
        toast.error("Could not get assessment by id: " + assessmentId)
      );
    }
    if (vulnerabilityId != undefined && ctxVulnerability?.id !== vulnerabilityId) {
      getData<Vulnerability>(`/api/vulnerabilities/${vulnerabilityId}`, setCtxVulnerability, () =>
        toast.error("Could not get vulnerability by id: " + vulnerabilityId)
      );
    }
    if (categoryId != undefined && ctxCategory?.id !== categoryId) {
      getData<Category>(`/api/categories/${categoryId}`, setCtxCategory, () =>
        toast.error("Could not get category by id: " + categoryId)
      );
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
