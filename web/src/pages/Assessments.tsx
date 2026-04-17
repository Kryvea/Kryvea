import { mdiContentDuplicate, mdiDownload, mdiFileEdit, mdiPlus, mdiStar, mdiTabSearch, mdiTrashCan } from "@mdi/js";
import { useContext, useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router";
import { toast } from "react-toastify";
import { deleteData, getData, patchData, postData } from "../api/api";
import { GlobalContext } from "../App";
import Flex from "../components/Composition/Flex";
import Grid from "../components/Composition/Grid";
import Modal from "../components/Composition/Modal";
import PageHeader from "../components/Composition/PageHeader";
import Table from "../components/Composition/Table";
import Button from "../components/Form/Button";
import Buttons from "../components/Form/Buttons";
import Checkbox from "../components/Form/Checkbox";
import Input from "../components/Form/Input";
import SelectWrapper from "../components/Form/SelectWrapper";
import { SelectOption } from "../components/Form/SelectWrapper.types";
import ExportReportModal from "../components/Modals/ExportReportModal";
import { Assessment, Template } from "../types/common.types";
import { formatDate } from "../utils/dates";
import { getPageTitle, sortBy } from "../utils/helpers";

export default function Assessments() {
  const navigate = useNavigate();
  const { customerId } = useParams<{ customerId: string }>();

  const [isModalDownloadActive, setIsModalDownloadActive] = useState(false);
  const [isModalTrashActive, setIsModalTrashActive] = useState(false);
  const [isModalCloneActive, setIsModalCloneActive] = useState(false);

  const [assessmentToClone, setAssessmentToClone] = useState<Assessment | null>(null);
  const [assessmentToDelete, setAssessmentToDelete] = useState<Assessment | null>(null);
  const [selectedAssessmentId, setSelectedAssessmentId] = useState<string | null>(null);

  const [cloneName, setCloneName] = useState("");

  const [allTemplates, setAllTemplates] = useState<Template[]>([]);

  const [statusSelectOptions] = useState<SelectOption[]>([
    { label: "On Hold", value: "On Hold" },
    { label: "In Progress", value: "In Progress" },
    { label: "Completed", value: "Completed" },
  ]);
  const [assessments, setAssessments] = useState<Assessment[]>([]);
  const [loadingAssessments, setLoadingAssessments] = useState(true);

  const [checkIncludePoc, setCheckIncludePoc] = useState(false);

  const {
    useCtxAssessment: [, setCtxAssessment],
  } = useContext(GlobalContext);

  const fetchAssessments = () => {
    setLoadingAssessments(true);
    getData<Assessment[]>(`/api/customers/${customerId}/assessments`, setAssessments, undefined, () =>
      setLoadingAssessments(false)
    );
  };

  useEffect(() => {
    document.title = getPageTitle("Assessments");
    fetchAssessments();
    getData<Template[]>("/api/templates", setAllTemplates);
  }, []);

  const handleOwnedToggle = assessment => () => {
    patchData(
      `/api/users/me/assessments`,
      { assessment: assessment.id, is_owned: !assessment.is_owned },
      fetchAssessments
    );
  };

  const openCloneModal = (assessment: Assessment) => {
    setAssessmentToClone(assessment);
    setCloneName(`${assessment.name} (Copy)`);
    setIsModalCloneActive(true);
  };

  const confirmClone = () => {
    postData<Assessment>(
      `/api/assessments/${assessmentToClone.id}/clone`,
      { name: cloneName.trim(), include_pocs: checkIncludePoc },
      _ => {
        fetchAssessments();
        setIsModalCloneActive(false);
        setAssessmentToClone(null);
        setCloneName(null);
        toast.success("Assessment cloned successfully");
      }
    );
  };

  const openDeleteModal = (assessment: Assessment) => {
    setAssessmentToDelete(assessment);
    setIsModalTrashActive(true);
  };

  const confirmDelete = () => {
    deleteData(`/api/assessments/${assessmentToDelete.id}`, () => {
      setAssessments(prev => prev.filter(a => a.id !== assessmentToDelete.id));
      setIsModalTrashActive(false);
      setAssessmentToDelete(null);
      toast.success("Assessment deleted successfully");
    });
  };

  const handleStatusChange = (assessmentId: string, selectedOption: SelectOption) => {
    patchData<Assessment>(`/api/assessments/${assessmentId}/status`, { status: selectedOption.value }, () => {
      setAssessments(prev => prev.map(a => (a.id === assessmentId ? { ...a, status: selectedOption.value } : a)));
    });
  };

  const openExportModal = (assessmentId: string) => {
    setSelectedAssessmentId(assessmentId);
    setIsModalDownloadActive(true);
  };

  return (
    <div>
      {/* Clone Modal */}
      {isModalCloneActive && (
        <Modal
          title="Clone assessment"
          confirmButtonLabel="Confirm"
          onConfirm={confirmClone}
          onCancel={() => setIsModalCloneActive(false)}
        >
          <Grid className="gap-4">
            <Input
              type="text"
              label="Assessment name"
              placeholder="Cloned assessment name"
              id="assessment_name"
              value={cloneName}
              onChange={e => setCloneName(e.target.value)}
            />
            <Checkbox
              id="include_pocs"
              label="Include PoCs"
              checked={checkIncludePoc}
              onChange={e => {
                setCheckIncludePoc(e.target.checked);
              }}
            />
          </Grid>
        </Modal>
      )}

      {/* Download Modal */}
      {isModalDownloadActive && (
        <ExportReportModal
          setShowModal={setIsModalDownloadActive}
          assessmentId={selectedAssessmentId}
          templates={allTemplates}
          language={assessments.find(a => a.id === selectedAssessmentId).language || "en"}
          cvssVersions={assessments.find(a => a.id === selectedAssessmentId).cvss_versions}
        />
      )}

      {/* Delete Confirmation Modal */}
      {isModalTrashActive && (
        <Modal
          title="Please confirm: action irreversible"
          confirmButtonLabel="Confirm"
          onConfirm={confirmDelete}
          onCancel={() => setIsModalTrashActive(false)}
        >
          <Flex col className="gap-4">
            <p>
              You are about to permanently delete the <strong>{assessmentToDelete.name}</strong> assessment.
            </p>
            <p className="text-[color:--error]">
              <strong>Warning:</strong> This action <em>cannot be undone</em> and will remove{" "}
              <u>all associated vulnerabilities</u> for this assessment.
            </p>
          </Flex>
        </Modal>
      )}

      <PageHeader icon={mdiTabSearch} title="Assessments">
        <Button
          className="justify-end"
          icon={mdiPlus}
          text="New assessment"
          small
          onClick={() => navigate(`/customers/${customerId}/assessments/new`)}
        />
      </PageHeader>

      <Table
        loading={loadingAssessments}
        data={assessments?.sort(sortBy("end_date_time", { reverse: true })).map(assessment => ({
          Title: (
            <Link
              to={`/customers/${customerId}/assessments/${assessment.id}/vulnerabilities`}
              onClick={() => setCtxAssessment(assessment)}
              title={assessment.name}
            >
              {assessment.name}
            </Link>
          ),
          Type: assessment.type.short,
          "CVSS Versions": [
            assessment.cvss_versions["2.0"] ? "2.0" : null,
            assessment.cvss_versions["3.1"] ? "3.1" : null,
            assessment.cvss_versions["4.0"] ? "4.0" : null,
          ]
            .filter(Boolean)
            .join(" | "),
          "Vuln count": assessment.vulnerability_count,
          Start: formatDate(assessment.start_date_time),
          End: formatDate(assessment.end_date_time),
          Language: assessment.language?.toUpperCase(),
          Status: (
            <SelectWrapper
              small
              widthFixed
              options={statusSelectOptions}
              value={statusSelectOptions.find(opt => opt.value === assessment.status)}
              onChange={selectedOption => handleStatusChange(assessment.id, selectedOption)}
            />
          ),
          buttons: (
            <Buttons noWrap>
              <Button
                variant={assessment.is_owned ? "warning" : "tertiary"}
                icon={mdiStar}
                onClick={handleOwnedToggle(assessment)}
                small
                title="Take ownership"
              />
              <Button
                variant="tertiary"
                icon={mdiFileEdit}
                onClick={() => navigate(`/customers/${customerId}/assessments/${assessment.id}`)}
                small
                title="Edit assessment"
              />
              <Button
                variant="tertiary"
                icon={mdiContentDuplicate}
                onClick={() => openCloneModal(assessment)}
                small
                title="Clone assessment"
              />
              <Button
                variant="tertiary"
                icon={mdiDownload}
                onClick={() => openExportModal(assessment.id)}
                small
                title="Download assessment"
              />
              <Button
                variant="danger"
                icon={mdiTrashCan}
                onClick={() => openDeleteModal(assessment)}
                small
                title="Delete assessment"
              />
            </Buttons>
          ),
        }))}
        perPageCustom={50}
        maxWidthColumns={{ Title: "24rem" }}
      />
    </div>
  );
}
