import { mdiDownload, mdiPencil, mdiPlus, mdiTrashCan, mdiUpload } from "@mdi/js";
import { useContext, useEffect, useMemo, useState } from "react";
import { Link, useLocation, useNavigate, useParams } from "react-router";
import { toast } from "react-toastify";
import { deleteData, getData, postData } from "../api/api";
import { GlobalContext } from "../App";
import Flex from "../components/Composition/Flex";
import Grid from "../components/Composition/Grid";
import Modal from "../components/Composition/Modal";
import PageHeader from "../components/Composition/PageHeader";
import Table from "../components/Composition/Table";
import Button from "../components/Form/Button";
import Buttons from "../components/Form/Buttons";
import SelectWrapper from "../components/Form/SelectWrapper";
import UploadFile from "../components/Form/UploadFile";
import AddTargetModal from "../components/Modals/AddTargetModal";
import ExportReportModal from "../components/Modals/ExportReportModal";
import { useDebounce } from "../hooks/hooks";
import { Category, Template, Vulnerability } from "../types/common.types";
import { formatDate } from "../utils/dates";
import { formatVulnerabilityTitle, getPageTitle } from "../utils/helpers";
import { getTargetLabel } from "../utils/targetLabel";

const DEFAULT_QUERY = "";
const DEFAULT_PAGE = 1;
const DEFAULT_LIMIT = 25;

export default function AssessmentVulnerabilities() {
  const navigate = useNavigate();
  const location = useLocation();
  const {
    useCtxAssessment: [ctxAssessment],
  } = useContext(GlobalContext);
  const { assessmentId } = useParams<{ assessmentId: string }>();

  const [isModalTargetActive, setIsModalTargetActive] = useState(false);
  const [isModalDownloadActive, setIsModalDownloadActive] = useState(false);
  const [isModalUploadActive, setIsModalUploadActive] = useState(false);
  const [isModalTrashActive, setIsModalTrashActive] = useState(false);

  const sourceOptions = [
    { label: "Nessus", value: "nessus" },
    { label: "Burp", value: "burp" },
  ];
  const [source, setSource] = useState<Category["source"]>();
  const [fileObj, setFileObj] = useState<File | null>(null);

  const [allTemplates, setAllTemplates] = useState<Template[]>([]);

  const [vulnerabilities, setVulnerabilities] = useState<Vulnerability[]>([]);
  const [totalVulnerabilities, setTotalVulnerabilities] = useState(0);
  const [totalPages, setTotalPages] = useState(0);
  const [loadingVulnerabilities, setLoadingVulnerabilities] = useState(true);
  const [vulnerabilityToDelete, setVulnerabilityToDelete] = useState<Vulnerability | null>(null);

  const searchAPI = useMemo(() => `/api/assessments/${assessmentId}/vulnerabilities?`, []);
  const urlSearchParams = new URLSearchParams(location.search);

  // Main search
  const [query, setQuery] = useState(urlSearchParams.get("query") ?? DEFAULT_QUERY);
  const debouncedQuery = useDebounce(query, 400);

  // Pagination
  const [page, setPage] = useState(Math.max(+urlSearchParams.get("page") || DEFAULT_PAGE, DEFAULT_PAGE));
  const [limit, setLimit] = useState(+urlSearchParams.get("limit") || DEFAULT_LIMIT);

  function fetchVulnerabilitiesPaginated(searchParams) {
    const paginationLoadingsTimeout = setTimeout(() => setLoadingVulnerabilities(true), 750);
    getData<any>(
      searchAPI + searchParams,
      data => {
        setTotalVulnerabilities(data.total_documents);
        setVulnerabilities(data.data);
        setTotalPages(data.total_pages);
      },
      undefined,
      () => {
        clearTimeout(paginationLoadingsTimeout);
        setLoadingVulnerabilities(false);
      }
    );
  }

  useEffect(() => {
    document.title = getPageTitle("Assessment Vulnerabilities");
    getData<Template[]>("/api/templates", setAllTemplates);
  }, []);

  // Sync with URL when user uses browser buttons
  useEffect(() => {
    setQuery(urlSearchParams.get("query") ?? DEFAULT_QUERY);
    setPage(Math.max(+urlSearchParams.get("page") || DEFAULT_PAGE, DEFAULT_PAGE));
    setLimit(+urlSearchParams.get("limit") || DEFAULT_LIMIT);
  }, [location.search]);

  // Fetch data
  useEffect(() => {
    const searchParams = buildSearchParams();

    if (location.search !== `?${searchParams}`) {
      navigate(`?${searchParams}`, { replace: false });
    }

    fetchVulnerabilitiesPaginated(searchParams);
  }, [debouncedQuery, limit, page]);

  // Build API params
  const buildSearchParams = () => {
    const sp = new URLSearchParams({
      query: debouncedQuery,
      page: page.toString(),
      limit: limit.toString(),
    });

    return sp.toString();
  };

  const openExportModal = () => {
    setIsModalDownloadActive(true);
  };

  const openTargetModal = () => {
    setIsModalTargetActive(true);
  };

  const openDeleteModal = (vulnerability: Vulnerability) => {
    setVulnerabilityToDelete(vulnerability);
    setIsModalTrashActive(true);
  };

  const confirmDelete = () => {
    deleteData(`/api/vulnerabilities/${vulnerabilityToDelete.id}`, () => {
      setVulnerabilities(prev => prev.filter(v => v.id !== vulnerabilityToDelete.id));
      toast.success("Vulnerability deleted successfully");
      setIsModalTrashActive(false);
      setVulnerabilityToDelete(null);
    });
  };

  const changeFile = ({ target: { files } }: React.ChangeEvent<HTMLInputElement>) => {
    if (!files || !files[0]) return;

    const file = files[0];
    setFileObj(file);

    setSource(null);
    if (file.name.endsWith(".nessus")) {
      setSource("nessus");
    }
  };

  const clearFile = () => setFileObj(null);

  const handleUploadBulk = () => {
    if (!fileObj) {
      toast.error("Please select a file to upload");
      return;
    }

    if (!source) {
      toast.error("Please select the source type");
      return;
    }

    const formData = new FormData();
    formData.append("file", fileObj);
    formData.append("import_data", `{"source": "${source}"}`);

    const toastId = toast.loading("Uploading...");
    postData<{ message: string }>(
      `/api/assessments/${assessmentId}/upload`,
      formData,
      () => {
        toast.update(toastId, {
          render: "Vulnerabilities uploaded successfully!",
          type: "success",
          isLoading: false,
          autoClose: 3000,
          closeButton: true,
        });
        setIsModalUploadActive(false);
        setFileObj(null);
        setPage(1); // trigger data refetch
      },
      err => {
        toast.update(toastId, {
          render: err.response.data.error,
          type: "error",
          isLoading: false,
          autoClose: 3000,
          closeButton: true,
        });
        setFileObj(null);
      }
    );
  };

  return (
    <div>
      {/* Add Target Modal */}
      {isModalTargetActive && <AddTargetModal setShowModal={setIsModalTargetActive} assessmentId={assessmentId} />}

      {isModalDownloadActive && (
        <ExportReportModal
          setShowModal={setIsModalDownloadActive}
          assessmentId={assessmentId}
          templates={allTemplates}
          language={ctxAssessment.language || "en"}
          cvssVersions={ctxAssessment.cvss_versions}
        />
      )}

      {/* Upload Modal */}
      {isModalUploadActive && (
        <Modal
          title="Upload file"
          confirmButtonLabel="Confirm"
          onConfirm={() => {
            handleUploadBulk();
            setIsModalUploadActive(false);
          }}
          onCancel={() => setIsModalUploadActive(false)}
        >
          <Flex col className="gap-2">
            <p>Upload vulnerability scan export files from the available sources.</p>
            <UploadFile
              label="Choose bulk file"
              inputId={"file"}
              filename={fileObj?.name}
              name={"imagePoc"}
              accept={".nessus,text/xml"}
              onChange={changeFile}
              onButtonClick={clearFile}
            />
            <SelectWrapper
              label="Source"
              className="w-1/2"
              id="source"
              options={sourceOptions}
              value={source ? { label: source.charAt(0).toUpperCase() + source.slice(1), value: source } : null}
              onChange={option => setSource(option.value)}
            />
          </Flex>
        </Modal>
      )}

      {/* Delete Confirmation Modal */}
      {isModalTrashActive && (
        <Modal
          title="Please confirm: action irreversible"
          confirmButtonLabel="Confirm"
          onConfirm={confirmDelete}
          onCancel={() => setIsModalTrashActive(false)}
        >
          <p>Are you sure to delete this vulnerability?</p>
        </Modal>
      )}

      <PageHeader title={`${ctxAssessment?.name} - Vulnerabilities`}>
        <Buttons className="justify-end">
          <Button icon={mdiPlus} text="New vulnerability" small onClick={() => navigate(`new`)} />
          <Button icon={mdiPlus} text="New Target" small onClick={openTargetModal} />
          <Button
            variant="tertiary"
            icon={mdiUpload}
            text="Upload"
            small
            onClick={() => setIsModalUploadActive(true)}
          />
          <Button variant="tertiary" icon={mdiDownload} text="Download report" small onClick={openExportModal} />
          {/* <Button icon={mdiFileEye} text="Live editor" small disabled onClick={() => navigate("/live_editor")} /> */}
        </Buttons>
      </PageHeader>

      <Grid className="gap-4">
        <Table
          loading={loadingVulnerabilities}
          backendCurrentPage={page}
          backendTotalRows={totalVulnerabilities}
          backendTotalPages={totalPages}
          backendSearch={query}
          onBackendSearch={setQuery}
          onBackendChangePage={setPage}
          onBackendChangePerPage={setLimit}
          data={vulnerabilities.map(vulnerability => {
            const cvssColumns = {};
            if (ctxAssessment?.cvss_versions["2.0"]) {
              cvssColumns["CVSSv2.0 Score"] = vulnerability.cvssv2.score;
            }
            if (ctxAssessment?.cvss_versions["3.1"]) {
              cvssColumns["CVSSv3.1 Score"] = vulnerability.cvssv31.score;
            }
            if (ctxAssessment?.cvss_versions["4.0"]) {
              cvssColumns["CVSSv4.0 Score"] = vulnerability.cvssv4.score;
            }

            return {
              Vulnerability: <Link to={vulnerability.id}>{formatVulnerabilityTitle(vulnerability)}</Link>,
              Target: getTargetLabel(vulnerability.target),
              ...cvssColumns,
              Status: vulnerability.status,
              User: vulnerability.user.username,
              "Last update": formatDate(vulnerability.updated_at),
              buttons: (
                <Buttons noWrap>
                  <Button
                    variant="tertiary"
                    icon={mdiPencil}
                    small
                    onClick={() => navigate(`${vulnerability.id}/edit`)}
                  />
                  <Button variant="danger" icon={mdiTrashCan} onClick={() => openDeleteModal(vulnerability)} small />
                </Buttons>
              ),
            };
          })}
          perPageCustom={limit}
          maxWidthColumns={{ Vulnerability: "40rem", Target: "15rem" }}
        />
      </Grid>
    </div>
  );
}
