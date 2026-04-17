import { mdiAccountEdit, mdiDownload, mdiTabSearch, mdiTarget, mdiTrashCan } from "@mdi/js";
import { memo, useContext, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router";
import { toast } from "react-toastify";
import { deleteData, getData, patchData, postData, putData } from "../api/api";
import { getKryveaShadow } from "../api/cookie";
import { GlobalContext } from "../App";
import Card from "../components/Composition/Card";
import CardTitle from "../components/Composition/CardTitle";
import Divider from "../components/Composition/Divider";
import Flex from "../components/Composition/Flex";
import Grid from "../components/Composition/Grid";
import PageHeader from "../components/Composition/PageHeader";
import Table from "../components/Composition/Table";
import Button from "../components/Form/Button";
import Buttons from "../components/Form/Buttons";
import Input from "../components/Form/Input";
import SelectWrapper from "../components/Form/SelectWrapper";
import { SelectOption } from "../components/Form/SelectWrapper.types";
import UploadFile from "../components/Form/UploadFile";
import { Customer, Template, uuidZero } from "../types/common.types";
import { languageMapping, USER_ROLE_ADMIN } from "../utils/constants";
import { getPageTitle, sortBy } from "../utils/helpers";

export default function CustomerDetail() {
  const {
    useCtxCustomer: [ctxCustomer, setCtxCustomer],
  } = useContext(GlobalContext);
  const { customerId } = useParams();
  const navigate = useNavigate();

  const [fileObj, setFileObj] = useState<File | null>(null);
  const [customerTemplates, setCustomerTemplates] = useState<Template[]>([]);
  const [loadingCustomerTemplates, setLoadingCustomerTemplates] = useState(true);
  const [selectedTemplateLanguage, setSelectedTemplateLanguage] = useState<SelectOption | null>(null);

  const isAdmin = getKryveaShadow() === USER_ROLE_ADMIN;

  const [logoId, setLogoId] = useState<string>("");
  const [logoFile, setLogoFile] = useState<File | null>(null);

  const [formCustomer, setFormCustomer] = useState({
    name: ctxCustomer?.name,
    language: ctxCustomer?.language,
  });

  const [newTemplateData, setNewTemplateData] = useState({
    name: "",
    identifier: "",
    file: null as File | null,
  });

  const languageOptions = Object.entries(languageMapping).map(([code, label]) => ({
    value: code,
    label,
  }));

  const selectedCustomerLanguage = languageOptions.find(opt => opt.value === formCustomer.language);

  useEffect(() => {
    setFormCustomer({ name: ctxCustomer?.name, language: ctxCustomer?.language });
  }, [ctxCustomer]);

  useEffect(() => {
    document.title = getPageTitle("Customer detail");
    if (customerId) {
      fetchCustomer();
    }
  }, [customerId]);

  function fetchCustomer() {
    setLoadingCustomerTemplates(true);
    getData<Customer>(
      `/api/customers/${customerId}`,
      data => {
        setFormCustomer(data);
        setCtxCustomer(data);
        setCustomerTemplates(data.templates);
        setLogoId(data.logo_id);
      },
      undefined,
      () => setLoadingCustomerTemplates(false)
    );
  }

  const handleFormCustomerChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormCustomer(prev => ({ ...prev, [name]: value }));
  };

  const handleFormTemplateChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setNewTemplateData(prev => ({ ...prev, [name]: value }));
  };

  const handleSubmit = () => {
    if (!formCustomer.name.trim()) {
      toast.error("Company name is required");
      return;
    }

    const payload = {
      name: formCustomer.name.trim(),
      language: formCustomer.language,
    };

    patchData(`/api/admin/customers/${ctxCustomer?.id}`, payload, () => {
      toast.success("Customer updated successfully");

      fetchCustomer();
    });
  };

  const changeFile = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setFileObj(file);
      setNewTemplateData(prev => ({ ...prev, file }));
    }
  };

  const clearFile = () => {
    setFileObj(null);
    setNewTemplateData(prev => ({ ...prev, file: null }));
  };

  const handleUploadTemplate = () => {
    if (!newTemplateData.name.trim()) {
      toast.error("Template name is required");
      return;
    }
    if (!fileObj) {
      toast.error("Template file is required");
      return;
    }
    if (!selectedTemplateLanguage) {
      toast.error("Please select a language for the template");
      return;
    }

    const dataTemplate = {
      name: newTemplateData.name,
      language: selectedTemplateLanguage.value,
      identifier: newTemplateData.identifier,
    };

    const formData = new FormData();
    formData.append("template", fileObj, fileObj.name);
    formData.append("data", JSON.stringify(dataTemplate));

    postData(`/api/customers/${ctxCustomer?.id}/templates/upload`, formData, () => {
      toast.success("Template uploaded successfully");
      setFileObj(null);
      setNewTemplateData({ name: "", identifier: "", file: null });
      setSelectedTemplateLanguage(null);
      fetchCustomer();
    });
  };

  const deleteTemplate = (templateId: string) => {
    deleteData(`/api/templates/${templateId}`, () => {
      toast.success("Template deleted successfully");
      setCustomerTemplates(prev => prev.filter(t => t.id !== templateId));
    });
  };

  const downloadTemplate = (template: Template) => {
    const a = document.createElement("a");
    a.href = `/api/files/templates/${template.file_id}`;
    a.download = template.name || template.filename;
    a.click();
  };

  const handleLogoChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    if (!["image/png", "image/jpeg"].includes(file.type)) {
      e.target.value = "";
      return;
    }

    const formData = new FormData();
    formData.append("file", file, file.name);

    const toastId = toast.loading("Uploading logo...");
    putData(
      `/api/admin/customers/${ctxCustomer?.id}/logo`,
      formData,
      () => {
        getData<Customer>(`/api/customers/${customerId}`, data => setLogoId(data.logo_id));
        toast.update(toastId, { render: "Logo uploaded", type: "success", isLoading: false, autoClose: 3000 });
      },
      err =>
        toast.update(toastId, { render: err.response?.data?.error, type: "error", isLoading: false, autoClose: 3000 })
    );
  };

  const CustomerLogo = memo(
    ({ logoId, isAdmin, formCustomer }: any) => {
      return (
        <img
          className="h-full w-full object-contain"
          src={`/api/files/customers/${logoId}`}
          alt={`${formCustomer.name}'s logo`}
          title={isAdmin ? "Change logo" : ""}
        />
      );
    },
    (prevProps, nextProps) => {
      // Return true to skip re-render
      return prevProps.logoId === nextProps.logoId;
    }
  );

  return (
    <div>
      <PageHeader icon={mdiAccountEdit} title={`Customer: ${ctxCustomer?.name}`}>
        <Buttons className="justify-end">
          <Button
            small
            variant="tertiary"
            text="Assessments"
            icon={mdiTabSearch}
            onClick={() => navigate(`/customers/${ctxCustomer?.id}/assessments`)}
          />
          <Button
            small
            variant="tertiary"
            text="Targets"
            icon={mdiTarget}
            onClick={() => navigate(`/customers/${ctxCustomer?.id}/targets`)}
          />
        </Buttons>
      </PageHeader>

      <Grid className="grid-cols-2 !items-start">
        <Card>
          <CardTitle title="Customer details" />
          <Grid className="gap-4">
            <Flex className="justify-center">
              <label
                data-disabled={!isAdmin}
                className={`aspect-video max-h-52 overflow-hidden rounded-xl shadow-lg shadow-[color:--bg-primary] transition ${isAdmin ? "cursor-pointer hover:scale-95 hover:shadow-[color:--bg-secondary] active:scale-90" : "cursor-not-allowed"} `}
                htmlFor={isAdmin ? "change-logo" : undefined}
              >
                {logoId == uuidZero || logoId === "" ? (
                  <div className="h-52 w-52 content-center rounded-xl border border-[color:--border-primary-highlight] bg-gradient-to-b from-[color:--bg-tertiary] to-[color:--bg-secondary] text-center text-[color:--text-secondary]">
                    {isAdmin ? "Add logo" : "No logo available"}
                  </div>
                ) : (
                  <CustomerLogo {...{ isAdmin, logoId, formCustomer }} />
                )}
              </label>

              {isAdmin && (
                <input
                  id="change-logo"
                  type="file"
                  accept="image/png, image/jpeg"
                  style={{ display: "none" }}
                  onChange={handleLogoChange}
                />
              )}
            </Flex>
            <Input
              type="text"
              label="Customer name"
              helperSubtitle="Required"
              placeholder="Customer name"
              id="customer-name"
              name="name"
              disabled={!isAdmin}
              value={formCustomer.name}
              onChange={handleFormCustomerChange}
            />
            <SelectWrapper
              label="Default language"
              id="language"
              options={languageOptions}
              value={selectedCustomerLanguage}
              disabled={!isAdmin}
              onChange={option => setFormCustomer(prev => ({ ...prev, language: option.value }))}
            />
          </Grid>
          <Divider />
          <Buttons>
            <Button
              text="Submit"
              onClick={handleSubmit}
              disabled={!isAdmin}
              title={!isAdmin ? "Only administrators can perform this action" : ""}
            />
          </Buttons>
        </Card>
        <Card>
          <CardTitle title="Custom report templates" />
          <Grid className="gap-4">
            <Grid className="grid-cols-2">
              <Input
                type="text"
                label="Template Name"
                placeholder="Insert name for the template"
                id="template-name"
                name="name"
                value={newTemplateData.name || ""}
                onChange={handleFormTemplateChange}
              />
              <Input
                type="text"
                label="Template Identifier"
                placeholder="e.g., Template for assessments"
                id="identifier"
                name="identifier"
                value={newTemplateData.identifier}
                onChange={handleFormTemplateChange}
              />
              <UploadFile
                label="Choose template file"
                inputId="file"
                filename={fileObj?.name}
                name="templateFile"
                accept=".docx,.xlsx"
                onChange={changeFile}
                onButtonClick={clearFile}
              />
              <SelectWrapper
                label="Language"
                options={languageOptions}
                value={selectedTemplateLanguage}
                onChange={setSelectedTemplateLanguage}
              />
            </Grid>
            <Buttons>
              <Button text="Upload" onClick={handleUploadTemplate} />
            </Buttons>
            <Divider />
            <Table
              loading={loadingCustomerTemplates}
              data={customerTemplates.sort(sortBy("name", { caseInsensitive: true })).map(template => ({
                Name: template.name,
                Language: languageMapping[template.language] || template.language,
                Filename: template.filename,
                "Template Type": template.template_type,
                "Template Identifier": template.identifier,
                buttons: (
                  <Buttons noWrap>
                    <Button
                      icon={mdiDownload}
                      title="Download template"
                      onClick={() => downloadTemplate(template)}
                      variant="secondary"
                    />
                    <Button
                      icon={mdiTrashCan}
                      title="Delete template"
                      onClick={() => deleteTemplate(template.id)}
                      variant="danger"
                    />
                  </Buttons>
                ),
              }))}
              perPageCustom={5}
            />
          </Grid>
        </Card>
      </Grid>
    </div>
  );
}
