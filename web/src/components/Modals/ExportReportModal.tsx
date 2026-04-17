import { useContext, useEffect, useMemo, useState } from "react";
import { toast } from "react-toastify";
import { postDownloadBlob } from "../../api/api";
import { curryDownloadReport } from "../../api/curries";
import { GlobalContext } from "../../App";
import { exportTypes, Template, uuidZero } from "../../types/common.types";
import Grid from "../Composition/Grid";
import Modal from "../Composition/Modal";
import Checkbox from "../Form/Checkbox";
import DateCalendar from "../Form/DateCalendar";
import SelectWrapper from "../Form/SelectWrapper";
import { SelectOption } from "../Form/SelectWrapper.types";

interface ExportReportModalProps {
  setShowModal;
  assessmentId: string;
  templates: Template[];
  language: string;
  cvssVersions: { "2.0": boolean; "3.1": boolean; "4.0": boolean };
}

export default function ExportReportModal({
  setShowModal,
  assessmentId,
  templates,
  language,
  cvssVersions,
}: ExportReportModalProps) {
  const {
    useCtxCustomer: [ctxCustomer],
  } = useContext(GlobalContext);

  const [selectedExportTypeOption, setSelectedExportTypeOption] = useState<SelectOption>(exportTypes[0]);
  const [templateOptions, setTemplateOptions] = useState<SelectOption[]>([]);
  const [selectedExportTemplate, setSelectedExportTemplate] = useState<Template>(null);

  const [exportEncryption, setExportEncryption] = useState<SelectOption>({ value: "none", label: "None" });
  const [exportPassword, setExportPassword] = useState("");
  const [deliveryDate, setDeliveryDate] = useState(new Date().toISOString());
  const [checkIncludeInfo, setCheckIncludeInfo] = useState(false);

  // CVSS sort configuration
  const cvssOptions = useMemo(() => {
    const opts = [];
    if (cvssVersions["3.1"]) opts.push({ value: "3.1", label: "CVSS v3.1" });
    if (cvssVersions["4.0"]) opts.push({ value: "4.0", label: "CVSS v4.0" });
    return opts;
  }, [cvssVersions]);

  const [sortyByCvss, setSortyByCvss] = useState<SelectOption>(cvssOptions[0]);

  useEffect(() => {
    const filtered = templates.filter(
      t =>
        t.language === language &&
        t.template_type === selectedExportTypeOption.value &&
        (t.customer.id === uuidZero || t.customer.id === ctxCustomer.id)
    );
    setTemplateOptions(
      filtered.map(t => ({
        value: t.id,
        label: t.identifier ? `${t.name} (${t.identifier})` : t.name,
      }))
    );

    setSelectedExportTemplate(filtered.length > 0 ? filtered[0] : null);
  }, [templates, selectedExportTypeOption, language]);

  const handleConfirm = () => {
    const payload: any = {
      type: selectedExportTypeOption.value,
      delivery_date_time: deliveryDate,
      include_informational_vulnerabilities: checkIncludeInfo,
    };

    if (selectedExportTypeOption.value !== "zip-default" && selectedExportTemplate) {
      payload.template = selectedExportTemplate.id;
    }

    if (selectedExportTypeOption.value === "zip-default") {
      payload.sort_by_cvss = sortyByCvss.value;
      payload.password = exportEncryption.value === "password" ? exportPassword : undefined;
    }

    const toastDownload = toast.loading("Generating report...");
    setShowModal(false);

    postDownloadBlob(`/api/assessments/${assessmentId}/export`, payload, curryDownloadReport(toastDownload), err => {
      toast.update(toastDownload, {
        render: err.response.data.error,
        type: "error",
        isLoading: false,
        autoClose: 3000,
        closeButton: true,
      });
    });
  };

  return (
    <Modal
      title="Download report"
      confirmButtonLabel="Generate"
      onConfirm={handleConfirm}
      onCancel={() => setShowModal(false)}
    >
      <Grid className="grid-cols-2">
        <SelectWrapper
          label="Type"
          id="type"
          options={exportTypes}
          value={selectedExportTypeOption}
          onChange={setSelectedExportTypeOption}
        />
        <SelectWrapper
          label="Template Type"
          id="template"
          disabled={selectedExportTypeOption.value === "zip-default"}
          options={templateOptions}
          value={
            selectedExportTemplate
              ? {
                  value: selectedExportTemplate.id,
                  label: selectedExportTemplate.identifier
                    ? `${selectedExportTemplate.name} (${selectedExportTemplate.identifier})`
                    : selectedExportTemplate.name,
                }
              : null
          }
          onChange={option => {
            const selected = templates.find(t => t.id === option.value) || null;
            setSelectedExportTemplate(selected);
          }}
        />

        {/* <SelectWrapper
          label="Encryption"
          id="encryption"
          options={[
            { value: "none", label: "None" },
            { value: "password", label: "Password" },
          ]}
          disabled={selectedExportTypeOption.value !== "zip-default"}
          value={exportEncryption}
          onChange={setExportEncryption}
        />
        <Input
          type="password"
          id="password"
          disabled={exportEncryption.value !== "password"}
          placeholder="Insert password"
          value={exportPassword}
          onChange={e => setExportPassword(e.target.value)}
        /> */}

        <DateCalendar
          idDate="delivery_date"
          label="Report delivery date"
          value={{ start: deliveryDate }}
          onChange={val => {
            if (typeof val === "string") {
              setDeliveryDate(val);
            }
          }}
        />
        <SelectWrapper
          label="Sort by"
          id="sort_by"
          options={cvssOptions}
          disabled={selectedExportTypeOption.value !== "zip-default" || cvssOptions.length < 2}
          value={sortyByCvss}
          onChange={setSortyByCvss}
        />
        <Checkbox
          id="include_informational_vulnerabilities"
          label="Include informational vulnerabilities"
          checked={checkIncludeInfo}
          onChange={e => {
            setCheckIncludeInfo(e.target.checked);
          }}
        />
      </Grid>
    </Modal>
  );
}
