import { mdiTrashCan } from "@mdi/js";
import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router";
import { toast } from "react-toastify";
import { deleteData, getData, patchData, postData } from "../api/api";
import Card from "../components/Composition/Card";
import Divider from "../components/Composition/Divider";
import Flex from "../components/Composition/Flex";
import Grid from "../components/Composition/Grid";
import Modal from "../components/Composition/Modal";
import PageHeader from "../components/Composition/PageHeader";
import Subtitle from "../components/Composition/Subtitle";
import Button from "../components/Form/Button";
import Buttons from "../components/Form/Buttons";
import Input from "../components/Form/Input";
import Label from "../components/Form/Label";
import SelectWrapper from "../components/Form/SelectWrapper";
import { SelectOption } from "../components/Form/SelectWrapper.types";
import Textarea from "../components/Form/Textarea";
import { Category, Settings } from "../types/common.types";
import { languageMapping } from "../utils/constants";
import { getPageTitle } from "../utils/helpers";

export const sourceCategoryOptions: SelectOption[] = [
  { value: "owasp_web", label: "OWASP Top 10 Web" },
  { value: "owasp_mobile", label: "OWASP Top 10 Mobile" },
  { value: "owasp_api", label: "OWASP Top 10 API" },
  { value: "owasp_llm", label: "OWASP Top 10 for LLM" },
  { value: "att&ck", label: "ATT&CK" },
  { value: "burp", label: "Burp" },
  { value: "cwe", label: "CWE" },
  { value: "nessus", label: "Nessus" },
];

const languageOptions: SelectOption[] = Object.entries(languageMapping).map(([value, label]) => ({ value, label }));

export default function CategoryUpsert() {
  const [defaultLanguageFromSettings, setDefaultLanguageFromSettings] = useState(null);
  const [identifier, setIdentifier] = useState("");
  const [name, setName] = useState("");
  const [subcategory, setSubcategory] = useState("");
  const [source, setSource] = useState<Category["source"]>();
  const [category, setCategory] = useState<Category>({} as Category);
  const [references, setReferences] = useState<string[]>([]);

  const [isModalTrashActive, setIsModalTrashActive] = useState(false);
  const [selectedLanguagesOptions, setSelectedLanguagesOptions] = useState<SelectOption[]>([]);

  const navigate = useNavigate();
  const { categoryId } = useParams<{ categoryId?: string }>();

  useEffect(() => {
    document.title = getPageTitle(categoryId ? "Edit Category" : "New Category");

    getData<Settings>("/api/admin/settings", ({ default_category_language }) => {
      setDefaultLanguageFromSettings(default_category_language);

      const languageOption = languageOptions.find(option => option.value === default_category_language);
      if (languageOption == undefined) {
        console.warn("Default language option not found in language options:", { defaultLanguageFromSettings });
        return;
      }

      setSelectedLanguagesOptions([languageOption]);
    });

    if (categoryId == undefined) {
      return;
    }

    getData<Category>(`/api/categories/${categoryId}`, category => {
      setIdentifier(category.identifier);
      setName(category.name);
      setSubcategory(category.subcategory);
      setSource(category.source);
      setReferences(category.references || []);
      setCategory(category);

      setSelectedLanguagesOptions(
        category.languages_order.map(lang => languageOptions.find(option => option.value === lang))
      );
    });
  }, []);

  const confirmDeleteCategory = () => {
    if (!categoryId) {
      return;
    }

    deleteData<{ message: string }>(`/api/admin/categories/${categoryId}`, () => {
      toast.success(`Category deleted successfully`);
      setIsModalTrashActive(false);
      navigate("/categories");
    });
  };

  const handleTrashConfirm = () => {
    confirmDeleteCategory();
  };

  const removeLanguage = (value: string) => {
    setSelectedLanguagesOptions(prev => prev.filter(option => option.value !== value));
  };

  const onChangeLanguageField = (lang: string, field: "description" | "remediation", value: string) => {
    const genericX = `generic_${field}`;
    setCategory(prev => ({ ...prev, [genericX]: { ...prev[genericX], [lang]: value } }));
  };

  // Submit handler
  const handleSubmit = () => {
    if (!identifier.trim() || !name.trim()) {
      toast.error("Identifier and Name are required");
      return;
    }

    if (selectedLanguagesOptions.length < 1) {
      toast.error("At least one language is required");
      return;
    }

    const generic_description = {};
    selectedLanguagesOptions.forEach(lang => {
      generic_description[lang.value] = category?.generic_description?.[lang.value] || "";
    });

    const generic_remediation = {};
    selectedLanguagesOptions.forEach(lang => {
      generic_remediation[lang.value] = category?.generic_remediation?.[lang.value] || "";
    });

    const payload: Omit<Category, "id" | "updated_at"> = {
      identifier: identifier.trim(),
      name: name.trim(),
      subcategory: subcategory.trim(),
      source: source,
      generic_description,
      generic_remediation,
      languages_order: selectedLanguagesOptions.map(lang => lang.value),
      references: references,
    };

    if (categoryId) {
      patchData<Category>(`/api/admin/categories/${categoryId}`, payload, () => {
        toast.success("Category updated successfully");
        navigate("/categories");
      });

      return;
    }

    postData<Category>("/api/admin/categories", payload, () => {
      toast.success("Category created successfully");
      navigate("/categories");
    });
  };

  const getGenericDescription = (lang: string) => {
    return category?.generic_description?.[lang] || "";
  };

  const getGenericRemediation = (lang: string) => {
    return category?.generic_remediation?.[lang] || "";
  };

  return (
    <div>
      {/* Delete single category modal */}
      {isModalTrashActive && (
        <Modal
          title="Please confirm: action irreversible"
          confirmButtonLabel="Confirm"
          onConfirm={handleTrashConfirm}
          onCancel={() => setIsModalTrashActive(false)}
        >
          <p>Are you sure you want to delete this category?</p>
        </Modal>
      )}

      <PageHeader title={categoryId ? "Edit Category" : "New Category"}>
        <Buttons className="justify-end">
          {categoryId && (
            <Button
              icon={mdiTrashCan}
              text="Remove category"
              small
              onClick={() => setIsModalTrashActive(true)}
              variant="danger"
            />
          )}
        </Buttons>
      </PageHeader>
      <Card>
        <Grid className="items-center">
          <Grid>
            <SelectWrapper
              isMulti
              label="Languages"
              options={languageOptions}
              onChange={setSelectedLanguagesOptions}
              value={selectedLanguagesOptions}
            />
            <Subtitle
              text={
                <>
                  Default: <b>{languageMapping[defaultLanguageFromSettings]}</b>
                </>
              }
            />
          </Grid>
        </Grid>
        <Divider />
        <Grid className="grid-cols-[0.6fr_1fr_1fr_1fr] gap-4">
          <Input
            type="text"
            label="Identifier"
            id="identifier"
            placeholder="A01:2021"
            value={identifier}
            onChange={e => setIdentifier(e.target.value)}
          />
          <Input
            type="text"
            label="Name"
            id="name"
            placeholder="Category name"
            value={name}
            onChange={e => setName(e.target.value)}
          />
          <Input
            type="text"
            label="Subcategory"
            id="subcategory"
            placeholder="Subcategory name"
            value={subcategory}
            onChange={e => setSubcategory(e.target.value)}
          />
          <SelectWrapper
            label="Source"
            id="source"
            options={sourceCategoryOptions}
            value={sourceCategoryOptions.find(option => option.value === source) || undefined}
            onChange={option => setSource(option.value)}
          />
        </Grid>
        <Divider />
        <Textarea
          label="References"
          id="references"
          placeholder="Enter one reference per line"
          value={references.join("\n")}
          onChange={e => setReferences(e.target.value.split("\n"))}
        />
        <Divider />

        {selectedLanguagesOptions.map(language => (
          <div key={language.value}>
            <Flex items="center" justify="between">
              <Label text={language.label} />
              <Button
                variant="danger"
                icon={mdiTrashCan}
                small
                title="Delete language category"
                onClick={() => removeLanguage(language.value)}
              />
            </Flex>
            <Grid className="grid-cols-2 !items-start gap-4">
              <Textarea
                label="Generic description"
                id={`gen_desc_${language.value}`}
                placeholder="Description here"
                value={getGenericDescription(language.value)}
                onChange={e => onChangeLanguageField(language.value, "description", e.target.value)}
                rows={10}
              />
              <Textarea
                label="Generic remediation"
                id={`gen_rem_${language.value}`}
                placeholder="Remediation here"
                value={getGenericRemediation(language.value)}
                onChange={e => onChangeLanguageField(language.value, "remediation", e.target.value)}
                rows={10}
              />
            </Grid>
            <Divider />
          </div>
        ))}

        <Grid>
          <Buttons>
            <Button text="Submit" onClick={handleSubmit} />
            <Button variant="outline-only" text="Cancel" onClick={() => navigate("/categories")} />
          </Buttons>
        </Grid>
      </Card>
    </div>
  );
}
