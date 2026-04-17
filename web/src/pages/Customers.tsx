import { mdiListBox, mdiPlus, mdiTrashCan } from "@mdi/js";
import { useContext, useEffect, useState } from "react";
import { useNavigate } from "react-router";
import { toast } from "react-toastify";
import { deleteData, getData } from "../api/api";
import { getKryveaShadow } from "../api/cookie";
import { GlobalContext } from "../App";
import Flex from "../components/Composition/Flex";
import Modal from "../components/Composition/Modal";
import PageHeader from "../components/Composition/PageHeader";
import Table from "../components/Composition/Table";
import Button from "../components/Form/Button";
import Buttons from "../components/Form/Buttons";
import { Customer } from "../types/common.types";
import { languageMapping, USER_ROLE_ADMIN } from "../utils/constants";
import { getPageTitle, sortBy } from "../utils/helpers";

export default function Customers() {
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [loadingCustomers, setLoadingCustomers] = useState(true);
  const [isModalTrashActive, setIsModalTrashActive] = useState(false);
  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null);

  const isAdmin = getKryveaShadow() === USER_ROLE_ADMIN;

  const {
    useCtxCustomer: [, setCtxCustomer],
    useCtxSelectedSidebarItemLabel: [, setCtxSelectedSidebarItemLabel],
  } = useContext(GlobalContext);

  const navigate = useNavigate();

  useEffect(() => {
    document.title = getPageTitle("Customers");
    fetchCustomers();
  }, []);

  function fetchCustomers() {
    setLoadingCustomers(true);
    getData<Customer[]>("/api/customers", setCustomers, undefined, () => setLoadingCustomers(false));
  }

  const openDeleteModal = (customer: Customer) => {
    setSelectedCustomer(customer);
    setIsModalTrashActive(true);
  };

  const handleDeleteConfirm = () => {
    if (!selectedCustomer) return;

    deleteData<{ message: string }>(`/api/admin/customers/${selectedCustomer.id}`, () => {
      toast.success("Customer deleted successfully");
      setIsModalTrashActive(false);
      setCustomers(prev => prev.filter(c => c.id !== selectedCustomer.id));
      setCtxCustomer(undefined);
    });
  };

  const handleModalClose = () => {
    setIsModalTrashActive(false);
    setSelectedCustomer(null);
  };

  return (
    <div>
      {/* Delete Confirmation Modal */}
      {isModalTrashActive && (
        <Modal
          title="Please confirm: action irreversible"
          confirmButtonLabel="Confirm"
          onConfirm={handleDeleteConfirm}
          onCancel={handleModalClose}
        >
          <Flex col className="gap-4">
            <p>
              You are about to permanently delete the customer <strong>{selectedCustomer?.name}</strong>.
            </p>
            <p className="text-[color:--error]">
              <strong>Warning:</strong> This action <em>cannot be undone</em> and will remove{" "}
              <u>all associated assessments, targets, and vulnerabilities</u> for this customer.
            </p>
          </Flex>
        </Modal>
      )}
      <PageHeader icon={mdiListBox} title="Customers">
        {isAdmin && <Button icon={mdiPlus} text="New customer" small onClick={() => navigate("/customers/new")} />}
      </PageHeader>

      <Table
        loading={loadingCustomers}
        data={customers.sort(sortBy("name", { caseInsensitive: true })).map(customer => ({
          Name: (
            <a
              className="cursor-pointer"
              onClick={() => {
                setCtxCustomer(customer);
                setCtxSelectedSidebarItemLabel("Assessments");
                navigate(`${customer.id}/assessments`);
              }}
            >
              {customer.name}
            </a>
          ),
          "Default language": languageMapping[customer.language] || customer.language,
          buttons: (
            <Buttons noWrap>
              <Button
                title={!isAdmin ? "Only administrators can perform this action" : "Delete customer"}
                disabled={!isAdmin}
                small
                variant="danger"
                onClick={() => openDeleteModal(customer)}
                icon={mdiTrashCan}
              />
            </Buttons>
          ),
        }))}
        perPageCustom={100}
      />
    </div>
  );
}
