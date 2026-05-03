import { mdiDownload, mdiPencil, mdiPlus, mdiShapePlus, mdiTrashCan, mdiUpload } from "@mdi/js";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import { toast } from "react-toastify";
import { deleteData, getData, postData } from "../api/api";
import Grid from "../components/Composition/Grid";
import Modal from "../components/Composition/Modal";
import PageHeader from "../components/Composition/PageHeader";
import Table, { Column } from "../components/Composition/Table";
import Button from "../components/Form/Button";
import Buttons from "../components/Form/Buttons";
import Checkbox from "../components/Form/Checkbox";
import UploadFile from "../components/Form/UploadFile";
import { Category } from "../types/common.types";
import { formatDate } from "../utils/dates";
import { getPageTitle } from "../utils/helpers";
import { useTableUrlState } from "../utils/useTableUrlState";
import { sourceCategoryOptions } from "./CategoryUpsert";

const sourceCategoryMap = Object.fromEntries(sourceCategoryOptions.map(({ value, label }) => [value, label]));

export default function Categories() {
  const navigate = useNavigate();

  const [categories, setCategories] = useState<Category[]>([]);
  const [loadingCategories, setLoadingCategories] = useState(true);

  const t = useTableUrlState({ defaultLimit: 50, defaultSort: { key: "Identifier", order: "asc" } });
  const [isModalManageActive, setIsModalManageActive] = useState(false);
  const [isModalTrashActive, setIsModalTrashActive] = useState(false);
  const [categoryToDelete, setCategoryToDelete] = useState<Category | null>(null);

  const [fileObj, setFileObj] = useState<File | null>(null);
  const [overrideExisting, setOverrideExisting] = useState(false);

  useEffect(() => {
    document.title = getPageTitle("Categories");
    fetchCategories();
  }, []);

  function fetchCategories() {
    setLoadingCategories(true);
    getData<Category[]>("/api/categories", setCategories, undefined, () => setLoadingCategories(false));
  }

  const changeFile = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files;
    if (!files || !files[0]) return;
    setFileObj(files[0]);
  };

  const clearFile = () => {
    setFileObj(null);
  };

  // Delete single category
  const confirmDeleteCategory = () => {
    if (!categoryToDelete) return;

    deleteData<{ message: string }>(`/api/admin/categories/${categoryToDelete.id}`, () => {
      toast.success(`Category "${categoryToDelete.name}" deleted successfully`);
      setIsModalTrashActive(false);
      setCategoryToDelete(null);
      setCategories(prev => prev.filter(c => c.id !== categoryToDelete.id));
    });
  };

  // Export categories file
  const handleExport = () => {
    const url = "/api/admin/categories/export";
    const link = document.createElement("a");
    link.href = url;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  // Upload categories file
  const handleUploadCategories = () => {
    if (!fileObj) {
      toast.error("Please select a file to upload");
      return;
    }

    const formData = new FormData();
    formData.append("categories", fileObj);
    formData.append("override", overrideExisting ? "true" : "false");

    postData<{ message: string }>("/api/admin/categories/upload", formData, () => {
      toast.success("Categories uploaded successfully");
      setIsModalManageActive(false);
      setFileObj(null);
      fetchCategories();
    });
  };

  const handleModalManageConfirm = () => {
    handleUploadCategories();
  };

  const handleModalTrashConfirm = () => {
    confirmDeleteCategory();
  };

  const categoryColumns: Column<Category>[] = [
    {
      header: "Identifier",
      sortValue: category => category.identifier + category.name + category.subcategory,
      render: category => category.identifier,
    },
    { header: "Name", maxWidth: "20rem", render: category => category.name },
    { header: "Subcategory", render: category => category.subcategory },
    { header: "Source", render: category => sourceCategoryMap[category.source] },
    {
      header: "Languages",
      render: category =>
        Object.values(category.languages_order || {})
          .sort()
          .join(" | ")
          .toUpperCase(),
    },
    {
      header: "Last update",
      sortValue: category => new Date(category.updated_at),
      render: category => formatDate(category.updated_at),
    },
    {
      kind: "actions",
      render: category => (
        <Buttons noWrap>
          <Button
            variant="tertiary"
            icon={mdiPencil}
            small
            title="Edit category"
            onClick={() => navigate(`${category.id}`)}
          />
          <Button
            variant="danger"
            icon={mdiTrashCan}
            title="Delete category"
            small
            onClick={() => {
              setCategoryToDelete(category);
              setIsModalTrashActive(true);
            }}
          />
        </Buttons>
      ),
    },
  ];

  return (
    <div>
      {/* Delete single category modal */}
      {isModalTrashActive && (
        <Modal
          title="Please confirm: action irreversible"
          confirmButtonLabel="Confirm"
          onConfirm={handleModalTrashConfirm}
          onCancel={() => {
            setIsModalTrashActive(false);
            setCategoryToDelete(null);
          }}
        >
          <p>
            Are you sure you want to delete category <strong>{categoryToDelete?.name}</strong>?
          </p>
        </Modal>
      )}

      {/* Upload categories */}
      {isModalManageActive && (
        <Modal
          title="Upload categories"
          confirmButtonLabel="Upload"
          onConfirm={handleModalManageConfirm}
          onCancel={() => setIsModalManageActive(false)}
        >
          <Grid className="gap-4">
            <UploadFile
              inputId="imported_categories"
              filename={fileObj?.name}
              name="categories"
              accept="application/json"
              onChange={changeFile}
              onButtonClick={clearFile}
            />
            <Checkbox
              id="override_categories"
              label="Override existing categories"
              checked={overrideExisting}
              onChange={e => setOverrideExisting(e.target.checked)}
            />
          </Grid>
        </Modal>
      )}

      <PageHeader icon={mdiShapePlus} title="Categories">
        <Buttons className="justify-end">
          <Button icon={mdiPlus} text="New category" small onClick={() => navigate("new")} />
          <Button
            variant="tertiary"
            icon={mdiUpload}
            text="Upload categories"
            small
            onClick={() => setIsModalManageActive(true)}
          />
          <Button variant="tertiary" icon={mdiDownload} text="Export categories" small onClick={handleExport} />
        </Buttons>
      </PageHeader>

      <div>
        <Table
          loading={loadingCategories}
          columns={categoryColumns}
          data={categories}
          search={t.search}
          sort={t.sort}
          onSortChange={t.onSortChange}
          pagination={t.pagination}
        />
      </div>
    </div>
  );
}
