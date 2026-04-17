import { mdiDownload, mdiFileChart, mdiPlus, mdiTrashCan } from "@mdi/js";
import { useEffect, useRef, useState } from "react";
import { toast } from "react-toastify";
import { deleteData, getData, postData } from "../api/api";
import Grid from "../components/Composition/Grid";
import Modal from "../components/Composition/Modal";
import PageHeader from "../components/Composition/PageHeader";
import Table from "../components/Composition/Table";
import Button from "../components/Form/Button";
import Buttons from "../components/Form/Buttons";
import Input from "../components/Form/Input";
import SelectWrapper from "../components/Form/SelectWrapper";
import { SelectOption } from "../components/Form/SelectWrapper.types";
import UploadFile from "../components/Form/UploadFile";
import { Template } from "../types/common.types";
import { languageMapping } from "../utils/constants";
import { getPageTitle, sortBy } from "../utils/helpers";

export default function Templates() {
  const [uploadedTemplates, setUploadedTemplates] = useState<Template[]>([]);
  const [loadingAssessments, setLoadingAssessments] = useState(false);
  const [templateToDelete, setTemplateToDelete] = useState<Template | null>(null);
  const [isModalUploadActive, setIsModalUploadActive] = useState(false);
  const [isModalTrashActive, setIsModalTrashActive] = useState(false);

  const [fileObj, setFileObj] = useState<File | null>(null);
  const [nameTemplate, setNameTemplate] = useState("");
  const [templateIdentifier, setTemplateIdentifier] = useState("");
  const [selectedLanguage, setSelectedLanguage] = useState<SelectOption | null>(null);

  const templateInputRef = useRef<HTMLInputElement | null>(null);

  const languageOptions: SelectOption[] = Object.entries(languageMapping).map(([code, label]) => ({
    value: code,
    label,
  }));

  useEffect(() => {
    document.title = getPageTitle("Report Templates");
    fetchTemplates();
  }, []);

  function fetchTemplates() {
    setLoadingAssessments(true);
    getData<Template[]>("/api/templates", setUploadedTemplates, undefined, () => setLoadingAssessments(false));
  }

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) setFileObj(file);
  };

  const clearFileInput = () => {
    setFileObj(null);
    if (templateInputRef.current) templateInputRef.current.value = "";
  };

  const resetUploadForm = () => {
    setFileObj(null);
    setNameTemplate("");
    setTemplateIdentifier("");
    setSelectedLanguage(null);
    if (templateInputRef.current) templateInputRef.current.value = "";
  };

  const handleUpload = () => {
    if (!fileObj) {
      toast.error("Please select a file to upload.");
      return;
    }
    if (!nameTemplate.trim()) {
      toast.error("Please provide a name for the template.");
      return;
    }
    if (!selectedLanguage) {
      toast.error("Please select a language for the template.");
      return;
    }

    const dataTemplate = {
      name: nameTemplate,
      language: selectedLanguage.value,
      identifier: templateIdentifier,
    };

    const formData = new FormData();
    formData.append("template", fileObj, fileObj.name);
    formData.append("data", JSON.stringify(dataTemplate));

    postData("/api/admin/templates/upload", formData, () => {
      toast.success("Template uploaded successfully");
      setIsModalUploadActive(false);
      resetUploadForm();
      fetchTemplates();
    });
  };

  const downloadTemplate = (template: Template) => {
    const a = document.createElement("a");
    a.href = `/api/files/templates/${template.file_id}`;
    a.download = template.name || template.filename;
    a.click();
  };

  const handleDeleteTemplate = () => {
    if (!templateToDelete) return;

    deleteData(`/api/templates/${templateToDelete.id}`, () => {
      toast.success("Template deleted successfully");
      setUploadedTemplates(prev => prev.filter(t => t.id !== templateToDelete.id));
      setIsModalTrashActive(false);
      setTemplateToDelete(null);
    });
  };

  return (
    <div>
      {/* Upload Modal */}
      {isModalUploadActive && (
        <Modal title="Upload Template" onConfirm={handleUpload} onCancel={() => setIsModalUploadActive(false)}>
          <Grid>
            <UploadFile
              label="Choose template file"
              inputId="template-upload"
              inputRef={templateInputRef}
              name="templateFile"
              filename={fileObj?.name}
              accept=".docx,.xlsx"
              onChange={handleFileChange}
              onButtonClick={clearFileInput}
            />
            <Grid className="grid-cols-2">
              <Input
                type="text"
                label="Template Name"
                id="template_name"
                value={nameTemplate}
                onChange={e => setNameTemplate(e.target.value)}
                placeholder="Insert name for the template"
              />
              <SelectWrapper
                label="Language"
                options={languageOptions}
                value={selectedLanguage}
                onChange={setSelectedLanguage}
              />
            </Grid>
            <Input
              type="text"
              label="Template Identifier"
              placeholder="e.g., Template for assessments"
              id="identifier"
              value={templateIdentifier}
              onChange={e => setTemplateIdentifier(e.target.value)}
            />
          </Grid>
        </Modal>
      )}

      {/* Delete Modal */}
      {isModalTrashActive && (
        <Modal
          title="Please confirm: action irreversible"
          confirmButtonLabel="Confirm"
          onConfirm={handleDeleteTemplate}
          onCancel={() => setIsModalTrashActive(false)}
        >
          <p>Are you sure you want to delete this template?</p>
        </Modal>
      )}

      <PageHeader icon={mdiFileChart} title="Report Templates">
        <Button icon={mdiPlus} text="New template" small onClick={() => setIsModalUploadActive(true)} />
      </PageHeader>

      <Table
        loading={loadingAssessments}
        data={uploadedTemplates.sort(sortBy("created_at")).map(template => ({
          Name: template.name,
          Filename: template.filename,
          Customer: template.customer?.name,
          Language: languageMapping[template.language] || template.language,
          "Template Type": template.template_type,
          "Template Identifier": template.identifier,
          buttons: (
            <Buttons noWrap>
              <Button
                variant="tertiary"
                icon={mdiDownload}
                onClick={() => downloadTemplate(template)}
                small
                title="Download template"
              />
              <Button
                icon={mdiTrashCan}
                onClick={() => {
                  setTemplateToDelete(template);
                  setIsModalTrashActive(true);
                }}
                variant="danger"
                small
                title="Delete template"
              />
            </Buttons>
          ),
        }))}
        perPageCustom={10}
      />
    </div>
  );
}
