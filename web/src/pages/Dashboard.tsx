import { mdiDotsCircle, mdiHistory } from "@mdi/js";
import { useContext, useEffect, useState } from "react";
import { Link } from "react-router";
import { getData } from "../api/api";
import { GlobalContext } from "../App";
import PageHeader from "../components/Composition/PageHeader";
import Table, { Column } from "../components/Composition/Table";
import { Assessment } from "../types/common.types";
import { formatDate } from "../utils/dates";
import { getPageTitle } from "../utils/helpers";

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

  const assessmentColumns: Column<Assessment>[] = [
    {
      header: "Customer",
      render: assessment => (
        <Link
          to={`/customers/${assessment.customer.id}/assessments`}
          onClick={() => setCtxCustomer(assessment.customer)}
        >
          {assessment.customer.name}
        </Link>
      ),
    },
    {
      header: "Assessment Name",
      render: assessment => (
        <Link to={`/customers/${assessment.customer.id}/assessments/${assessment.id}/vulnerabilities`}>
          {assessment.name}
        </Link>
      ),
    },
    { header: "Assessment Type", render: assessment => assessment.type.short },
    { header: "Vulnerability Count", render: assessment => assessment.vulnerability_count },
    {
      header: "Start",
      sortValue: assessment => new Date(assessment.start_date_time),
      render: assessment => formatDate(assessment.start_date_time),
    },
    {
      header: "End",
      sortValue: assessment => new Date(assessment.end_date_time),
      render: assessment => formatDate(assessment.end_date_time),
    },
    { header: "Status", render: assessment => assessment.status },
  ];

  return (
    <div className="flex flex-col gap-4">
      <div>
        <PageHeader icon={mdiDotsCircle} title="Ongoing Assessments" />
        <Table
          loading={loadingAssessments}
          columns={assessmentColumns}
          data={assessments.filter(a => a.status !== "Completed")}
          perPage={10}
          defaultSort={{ key: "End", order: "desc" }}
        />
      </div>
      <div>
        <PageHeader icon={mdiHistory} title="Completed Assessments" />
        <Table
          loading={loadingAssessments}
          columns={assessmentColumns}
          data={assessments.filter(a => a.status === "Completed")}
          perPage={10}
          defaultSort={{ key: "End", order: "desc" }}
        />
      </div>
    </div>
  );
}
