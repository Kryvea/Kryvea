import { mdiDotsCircle, mdiHistory } from "@mdi/js";
import { useContext, useEffect, useState } from "react";
import { Link } from "react-router";
import { getData } from "../api/api";
import { GlobalContext } from "../App";
import PageHeader from "../components/Composition/PageHeader";
import Table from "../components/Composition/Table";
import { Assessment } from "../types/common.types";
import { formatDate } from "../utils/dates";
import { getPageTitle, sortBy } from "../utils/helpers";

export default function Dashboard() {
  const {
    useCtxCustomer: [, setCtxCustomer],
  } = useContext(GlobalContext);

  const [assessments, setAssessments] = useState<Assessment[]>([]);
  const [loadingAssessments, setLoadingAssessments] = useState(true);

  useEffect(() => {
    document.title = getPageTitle("Dashboard");
    setLoadingAssessments(true);
    getData<Assessment[]>("/api/assessments/owned", setAssessments, undefined, () => setLoadingAssessments(false));
  }, []);

  return (
    <div className="flex flex-col gap-4">
      <div>
        <PageHeader icon={mdiDotsCircle} title="Ongoing Assessments" />
        <Table
          loading={loadingAssessments}
          data={assessments
            .filter(a => a.status !== "Completed")
            .sort(sortBy("end_date_time", { reverse: true }))
            .map(assessment => ({
              Customer: (
                <Link
                  to={`/customers/${assessment.customer.id}/assessments`}
                  onClick={() => {
                    setCtxCustomer(assessment.customer);
                  }}
                >
                  {assessment.customer.name}
                </Link>
              ),
              "Assessment Name": (
                <Link to={`/customers/${assessment.customer.id}/assessments/${assessment.id}/vulnerabilities`}>
                  {assessment.name}
                </Link>
              ),
              "Assessment Type": assessment.type.short,
              "Vulnerability Count": assessment.vulnerability_count,
              Start: formatDate(assessment.start_date_time),
              End: formatDate(assessment.end_date_time),
              Status: assessment.status,
            }))}
          perPageCustom={10}
        />
      </div>
      <div>
        <PageHeader icon={mdiHistory} title="Completed Assessments" />
        <Table
          loading={loadingAssessments}
          data={assessments
            .filter(a => a.status === "Completed")
            .sort(sortBy("end_date_time", { reverse: true }))
            .map(assessment => ({
              Customer: (
                <Link
                  to={`/customers/${assessment.customer.id}/assessments`}
                  onClick={() => {
                    setCtxCustomer(assessment.customer);
                  }}
                >
                  {assessment.customer.name}
                </Link>
              ),
              "Assessment Name": (
                <Link to={`/customers/${assessment.customer.id}/assessments/${assessment.id}/vulnerabilities`}>
                  {assessment.name}
                </Link>
              ),
              "Assessment Type": assessment.type.short,
              "Vulnerability Count": assessment.vulnerability_count,
              Start: formatDate(assessment.start_date_time),
              End: formatDate(assessment.end_date_time),
              Status: assessment.status,
            }))}
          perPageCustom={10}
        />
      </div>
    </div>
  );
}
