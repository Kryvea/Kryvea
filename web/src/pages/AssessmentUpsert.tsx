import { mdiPlus } from "@mdi/js";
import { useContext, useEffect, useMemo, useReducer, useState } from "react";
import { useNavigate, useParams } from "react-router";
import { toast } from "react-toastify";
import { getData, patchData, postData } from "../api/api";
import { GlobalContext } from "../App";
import Card from "../components/Composition/Card";
import Divider from "../components/Composition/Divider";
import Flex from "../components/Composition/Flex";
import Grid from "../components/Composition/Grid";
import PageHeader from "../components/Composition/PageHeader";
import Button from "../components/Form/Button";
import Buttons from "../components/Form/Buttons";
import Checkbox from "../components/Form/Checkbox";
import DateCalendar from "../components/Form/DateCalendar";
import Input from "../components/Form/Input";
import Label from "../components/Form/Label";
import SelectWrapper from "../components/Form/SelectWrapper";
import { SelectOption } from "../components/Form/SelectWrapper.types";
import AddTargetModal from "../components/Modals/AddTargetModal";
import { Assessment, Target } from "../types/common.types";
import { Keys } from "../types/utils.types";
import { languageMapping } from "../utils/constants";
import { getPageTitle } from "../utils/helpers";
import { getTargetLabel } from "../utils/targetLabel";

type AssessmentPayload = Omit<
  Assessment,
  "id" | "created_at" | "updated_at" | "vulnerability_count" | "customer" | "is_owned"
>;

const ASSESSMENT_TYPE: SelectOption[] = [
  {
    value: { short: "VAPT", full: "Vulnerability Assessment Penetration Test" },
    label: "Vulnerability Assessment Penetration Test",
  },
  { value: { short: "NPT", full: "Network Penetration Test" }, label: "Network Penetration Test" },
  { value: { short: "WAPT", full: "Web Application Penetration Test" }, label: "Web Application Penetration Test" },
  {
    value: { short: "MAPT", full: "Mobile Application Penetration Test" },
    label: "Mobile Application Penetration Test",
  },
  { value: { short: "API PT", full: "API Penetration Test" }, label: "API Penetration Test" },
  { value: { short: "Wi-Fi PT", full: "Wi-Fi Penetration Test" }, label: "Wi-Fi Penetration Test" },
  { value: { short: "RT", full: "Red Team Assessment" }, label: "Red Team Assessment" },
  { value: { short: "IOT PT", full: "IoT Device Penetration Test" }, label: "IoT Device Penetration Test" },
  {
    value: { short: "SAST", full: "Static Application Security Testing" },
    label: "Static Application Security Testing",
  },
  {
    value: { short: "DAST", full: "Dynamic Application Security Testing" },
    label: "Dynamic Application Security Testing",
  },
];

const CVSS_VERSIONS: SelectOption[] = [
  { value: "2.0", label: "2.0" },
  { value: "3.1", label: "3.1" },
  { value: "4.0", label: "4.0" },
];

const ENVIRONMENT: SelectOption[] = [
  { value: "Testing", label: "Testing" },
  { value: "Pre-Production", label: "Pre-Production" },
  { value: "Production", label: "Production" },
];

const TESTING_TYPE: SelectOption[] = [
  { value: "White Box", label: "White Box" },
  { value: "Gray Box", label: "Gray Box" },
  { value: "Black Box", label: "Black Box" },
];

const OSSTMM_VECTOR: SelectOption[] = [
  { value: "Inside to Inside", label: "Inside to Inside" },
  { value: "Inside to Outside", label: "Inside to Outside" },
  { value: "Outside to Outside", label: "Outside to Outside" },
  { value: "Outside to Inside", label: "Outside to Inside" },
];

const initialSelectedOptionsState: {
  type: SelectOption;
  language: SelectOption;
  environment?: SelectOption;
  testing_type?: SelectOption;
  osstmm_vector?: SelectOption;
} = {
  type: undefined,
  language: undefined,
  environment: undefined,
  testing_type: undefined,
  osstmm_vector: undefined,
};

function reducer(
  state: typeof initialSelectedOptionsState,
  {
    action,
    field,
    value,
  }:
    | { action: "field"; field: Keys<typeof initialSelectedOptionsState>; value: SelectOption }
    | { action: "all"; field: ""; value: AssessmentPayload }
) {
  switch (action) {
    case "field":
      return { ...state, [field]: value };
    case "all":
      return {
        type: { label: value.type.full, value: value.type },
        language: { label: languageMapping[value.language], value: value.language },
        environment: { label: value.environment, value: value.environment },
        testing_type: { label: value.testing_type, value: value.testing_type },
        osstmm_vector: { label: value.osstmm_vector, value: value.osstmm_vector },
      };
  }
}

const createDateWithTime = (hours: number, minutes: number = 0): string => {
  const date = new Date();
  date.setHours(hours, minutes, 0, 0);
  return date.toISOString();
};

export default function AssessmentUpsert() {
  const navigate = useNavigate();
  const { customerId, assessmentId } = useParams<{ customerId: string; assessmentId?: string }>();
  const {
    useCtxCustomer: [ctxCustomer],
  } = useContext(GlobalContext);
  const [targets, setTargets] = useState<Target[]>([]);
  const isEdit = Boolean(assessmentId);
  const [isModalTargetActive, setIsModalTargetActive] = useState(false);

  const customerDefaultLanguage = ctxCustomer?.language ?? "";
  const languageOptions = useMemo(
    () => Object.entries(languageMapping).map(([code, label]) => ({ value: code, label })),
    []
  );
  const defaultSelectedLanguage = languageOptions.find(opt => opt.value === customerDefaultLanguage) || undefined;
  const [selectedOptions, updateSelectedOptions] = useReducer(reducer, {
    ...initialSelectedOptionsState,
    language: defaultSelectedLanguage,
  });

  const [form, setForm] = useState<AssessmentPayload>({
    type: { short: "", full: "" },
    name: "",
    start_date_time: createDateWithTime(9, 0),
    end_date_time: createDateWithTime(18, 0),
    kickoff_date_time: new Date().toISOString(),
    language: defaultSelectedLanguage?.value || "",
    targets: [],
    status: "On Hold",
    cvss_versions: { "2.0": false, "3.1": false, "4.0": false },
    environment: "",
    testing_type: "",
    osstmm_vector: "",
  });

  const fetchTargets = () => {
    getData<Target[]>(`/api/customers/${customerId}/targets`, setTargets);
  };

  useEffect(() => {
    document.title = getPageTitle(isEdit ? "Edit Assessment" : "Add Assessment");
    fetchTargets();

    if (isEdit) {
      getData<Assessment>(
        `/api/assessments/${assessmentId}`,
        data => {
          setForm({
            type: data.type,
            name: data.name,
            start_date_time: data.start_date_time,
            end_date_time: data.end_date_time,
            kickoff_date_time: data.kickoff_date_time,
            language: data.language,
            targets: data.targets,
            status: data.status,
            cvss_versions: data.cvss_versions,
            environment: data.environment,
            testing_type: data.testing_type,
            osstmm_vector: data.osstmm_vector,
          });
          updateSelectedOptions({ action: "all", field: "", value: data });
        },
        err => {
          toast.error(err.response.data.error);
          navigate(`/customers/${customerId}/assessments`);
        }
      );
    }
  }, [isEdit, customerId, assessmentId, navigate]);

  const targetOptions: SelectOption[] = useMemo(
    () =>
      targets.map(target => ({
        value: target.id,
        label: getTargetLabel(target),
      })),
    [targets]
  );

  const handleChange = (field: keyof typeof form, value: any) => {
    setForm(f => {
      const newForm = { ...f, [field]: value };
      if (field === "start_date_time" || field === "end_date_time") {
        const startDate = new Date(field === "start_date_time" ? value : f.start_date_time);
        const endDate = new Date(field === "end_date_time" ? value : f.end_date_time);

        if (startDate > endDate) {
          toast.error("End date cannot be before start date");
          return f;
        }
      }

      return newForm;
    });
  };

  const handleSelectChange = (field: Keys<typeof initialSelectedOptionsState>, option: SelectOption | null) => {
    updateSelectedOptions({ action: "field", field: field, value: option });
    handleChange(field, option ? option.value : "");
  };

  const toggleCvssVersion = (version: string) => {
    setForm(prev => ({
      ...prev,
      cvss_versions: {
        ...prev.cvss_versions,
        [version]: !prev.cvss_versions[version],
      },
    }));
  };

  const handleTargetsChange = (options: SelectOption[] | null) => {
    handleChange("targets", options ? options.map(opt => targets.find(t => t.id === opt.value)!).filter(Boolean) : []);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    const payload = {
      ...form,
      targets: form.targets.map(target => target.id),
      customer_id: customerId,
    };

    const endpoint = isEdit ? `/api/assessments/${assessmentId}` : `/api/assessments`;

    const apiCall = isEdit ? patchData : postData;

    if (!form.name) {
      toast.error("Assessment name required");
      return;
    }
    if (!form.type.full || !form.type.short) {
      toast.error("Assessment type required");
      return;
    }
    if (Object.values(form.cvss_versions).every(val => val === false)) {
      toast.error("Select at least one CVSS version");
      return;
    }
    if (new Date(form.start_date_time) > new Date(form.end_date_time)) {
      toast.error("End date cannot be before start date");
      return;
    }

    apiCall(endpoint, payload, data => {
      toast.success((data as any)?.message);
      navigate(`/customers/${customerId}/assessments`);
    });
  };

  const openTargetModal = () => {
    setIsModalTargetActive(true);
  };

  const handleTargetCreated = createdTargetId => {
    getData<Target[]>(`/api/customers/${customerId}/targets`, newTargets => {
      setTargets(newTargets);
      const newTarget = newTargets.find(t => t.id === createdTargetId);
      if (newTarget) {
        setForm(prev => ({
          ...prev,
          targets: [...prev.targets, newTarget],
        }));
      }
    });
  };

  return (
    <div>
      {isModalTargetActive && (
        <AddTargetModal
          setShowModal={setIsModalTargetActive}
          assessmentId={assessmentId}
          onTargetCreated={handleTargetCreated}
        />
      )}

      <PageHeader title={isEdit ? "Edit Assessment" : "New Assessment"} />
      <Card>
        <form onSubmit={handleSubmit}>
          <Grid>
            <Grid className="grid-cols-2">
              <Input
                type="text"
                label="Name"
                id="name"
                value={form.name}
                onChange={e => handleChange("name", e.target.value)}
                placeholder="Insert a name"
              />
              <SelectWrapper
                label="Assessment Type"
                id="type"
                options={ASSESSMENT_TYPE}
                value={selectedOptions.type}
                closeMenuOnSelect
                onChange={option => handleSelectChange("type", option)}
              />
            </Grid>
            <Grid className="grid-cols-5">
              <DateCalendar
                idDate="start_date_time"
                label="Activity start"
                showTime
                value={{ start: form.start_date_time }}
                onChange={val => {
                  if (typeof val === "string") {
                    handleChange("start_date_time", val);
                  }
                }}
                placeholder="Select start"
              />
              <DateCalendar
                idDate="end_date_time"
                label="Activity end"
                showTime
                value={{ start: form.end_date_time }}
                onChange={val => {
                  if (typeof val === "string") {
                    handleChange("end_date_time", val);
                  }
                }}
                placeholder="Select end"
              />
              <DateCalendar
                idDate="kickoff_date_time"
                label="Kick-off Date"
                value={{ start: form.kickoff_date_time }}
                onChange={val => {
                  if (typeof val === "string") {
                    handleChange("kickoff_date_time", val);
                  }
                }}
              />
              <SelectWrapper
                label="Language"
                options={languageOptions}
                value={selectedOptions.language}
                onChange={option => handleSelectChange("language", option)}
              />
              <Grid className="h-full !items-start justify-center text-center">
                <Label text="CVSS Versions" />
                <Flex className="gap-4">
                  {CVSS_VERSIONS.map(({ value, label }) => (
                    <Checkbox
                      key={value}
                      id={`cvss_${value}`}
                      label={label}
                      checked={form.cvss_versions[value] || false}
                      onChange={() => toggleCvssVersion(value)}
                    />
                  ))}
                </Flex>
              </Grid>
            </Grid>
            <Grid className="grid-cols-[1fr_auto]">
              <SelectWrapper
                label="Session targets"
                id="targets"
                options={targetOptions}
                isMulti
                value={targetOptions.filter(
                  opt => Array.isArray(form.targets) && form.targets.some(t => t.id === opt.value)
                )}
                onChange={handleTargetsChange}
                closeMenuOnSelect={false}
              />
              <Button icon={mdiPlus} text="New Target" onClick={openTargetModal} />
            </Grid>
            <SelectWrapper
              label="Environment"
              id="environment"
              options={ENVIRONMENT}
              value={selectedOptions.environment}
              closeMenuOnSelect
              onChange={option => handleSelectChange("environment", option)}
              isClearable
            />
            <SelectWrapper
              label="Testing type"
              id="testing_type"
              options={TESTING_TYPE}
              value={selectedOptions.testing_type}
              closeMenuOnSelect
              onChange={option => handleSelectChange("testing_type", option)}
              isClearable
            />
            <SelectWrapper
              label="OSSTMM Vector"
              id="osstmm_vector"
              options={OSSTMM_VECTOR}
              value={selectedOptions.osstmm_vector}
              closeMenuOnSelect
              onChange={option => handleSelectChange("osstmm_vector", option)}
              isClearable
            />
            <Divider />
            <Buttons>
              <Button text="Submit" onClick={() => {}} formSubmit />
              <Button variant="outline-only" text="Cancel" onClick={() => navigate(-1)} />
            </Buttons>
          </Grid>
        </form>
      </Card>
    </div>
  );
}
