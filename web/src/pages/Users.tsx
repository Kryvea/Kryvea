import { mdiAccountEdit, mdiAccountMultiple, mdiLockReset, mdiPlus, mdiTrashCan } from "@mdi/js";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import { toast } from "react-toastify";
import { deleteData, getData, patchData, postData } from "../api/api";
import Grid from "../components/Composition/Grid";
import Modal from "../components/Composition/Modal";
import PageHeader from "../components/Composition/PageHeader";
import Subtitle from "../components/Composition/Subtitle";
import Table from "../components/Composition/Table";
import Button from "../components/Form/Button";
import Buttons from "../components/Form/Buttons";
import DateCalendar from "../components/Form/DateCalendar";
import Input from "../components/Form/Input";
import SelectWrapper from "../components/Form/SelectWrapper";
import { SelectOption } from "../components/Form/SelectWrapper.types";
import { Customer, User } from "../types/common.types";
import { getPageTitle, sortBy } from "../utils/helpers";

export default function Users() {
  const [username, setUsername] = useState("");
  const [role, setRole] = useState("");
  const [selectedCustomers, setSelectedCustomers] = useState<string[]>([]);
  const [userDisabled, setUserDisabled] = useState<string>(null);
  const [temporaryPassword, setTemporaryPassword] = useState<string>(null);

  const [users, setUsers] = useState<User[]>([]);
  const [loadingUsers, setLoadingUsers] = useState(true);
  const [customers, setCustomers] = useState<Customer[]>([]);

  const [isModalResetPswActive, setIsModalResetPswActive] = useState(false);
  const [isModalTempPswActive, setIsModalTempPswActive] = useState(false);
  const [isModalEditActive, setIsModalEditActive] = useState(false);
  const [isModalTrashActive, setIsModalTrashActive] = useState(false);

  const [activeUserId, setActiveUserId] = useState<string>(null);

  const navigate = useNavigate();

  const fetchUsers = () => {
    setLoadingUsers(true);
    getData<User[]>("/api/admin/users", setUsers, undefined, () => setLoadingUsers(false));
  };

  useEffect(() => {
    document.title = getPageTitle("Users");
    fetchUsers();
    getData<Customer[]>("/api/customers", setCustomers);
  }, []);

  // Prepare customer options for SelectWrapper
  const customerOptions: SelectOption[] = customers.map(customer => ({
    label: customer.name,
    value: customer.id,
  }));

  // Handle changes in the customers multi-select
  const handleCustomerChange = (selectedOptions: { value: string; label: string }[] | null) => {
    setSelectedCustomers(selectedOptions.map(option => option.value));
  };

  const openResetPswModal = (user: User) => {
    setActiveUserId(user.id);
    setUsername(user.username);
    setIsModalResetPswActive(true);
  };

  const openEditModal = (user: User) => {
    setActiveUserId(user.id);
    setUsername(user.username);
    setRole(user.role);
    setSelectedCustomers(user.customers?.map(c => c.id) ?? []);
    setUserDisabled(user.disabled_at);
    setIsModalEditActive(true);
  };

  const openDeleteModal = (userId: string) => {
    setActiveUserId(userId);
    setIsModalTrashActive(true);
  };

  const handleModalAction = () => {
    setIsModalResetPswActive(false);
    setIsModalEditActive(false);
    setIsModalTrashActive(false);
    setActiveUserId(null);
  };

  const handleUpdateUser = () => {
    if (!activeUserId) return;

    const payload = {
      username,
      role,
      customers: selectedCustomers,
      disabled_at: userDisabled,
    };

    patchData<User>(`/api/admin/users/${activeUserId}`, payload, updatedUser => {
      toast.success(`User ${payload.username} updated successfully`);
      setIsModalEditActive(false);
      setUsers(prev => prev.map(u => (u.id === updatedUser.id ? updatedUser : u)));
      fetchUsers();
    });
  };

  const handleResetPsw = () => {
    if (!activeUserId) return;

    postData<{ password: string }>(`/api/admin/users/${activeUserId}/reset-password`, {}, response => {
      toast.success(`User ${username} password reset successfully`);
      setIsModalResetPswActive(false);
      setTemporaryPassword(response.password);
      setIsModalTempPswActive(true);
    });
  };

  const handleDeleteUser = () => {
    if (!activeUserId) return;

    deleteData<{ message: string }>(`/api/admin/users/${activeUserId}`, () => {
      toast.success("User deleted successfully");
      setIsModalTrashActive(false);
      setUsers(prev => prev.filter(u => u.id !== activeUserId));
      fetchUsers();
    });
  };

  return (
    <div>
      {/* Edit User Modal */}
      {isModalEditActive && (
        <Modal title="Edit user" confirmButtonLabel="Confirm" onConfirm={handleUpdateUser} onCancel={handleModalAction}>
          <Grid className="gap-4">
            <Input
              type="text"
              label="Username"
              placeholder="username"
              id="username"
              value={username}
              onChange={e => setUsername(e.target.value)}
            />
            <SelectWrapper
              label="Role"
              id="role-selection"
              options={[
                { value: "admin", label: "Admin" },
                { value: "user", label: "User" },
              ]}
              closeMenuOnSelect
              onChange={option => setRole(option?.value || "")}
              value={role ? { value: role, label: role } : null}
            />
            <SelectWrapper
              label="Customers"
              options={customerOptions}
              isMulti
              disabled={role === "admin"}
              value={customerOptions.filter(option => selectedCustomers.includes(option.value))}
              onChange={handleCustomerChange}
              closeMenuOnSelect={false}
              id="customer-selection"
            />
            <DateCalendar
              idDate="disable_user_at"
              label="Disable user from"
              showTime
              value={{ start: userDisabled }}
              onChange={val => {
                if (typeof val === "string") {
                  setUserDisabled(val);
                }
              }}
            />
          </Grid>
        </Modal>
      )}

      {/* Password reset Confirmation Modal */}
      {isModalResetPswActive && (
        <Modal
          title="Please confirm: action irreversible"
          confirmButtonLabel="Confirm"
          onConfirm={handleResetPsw}
          onCancel={handleModalAction}
        >
          <p>
            Are you sure you want to reset the password for user '<strong>{username}</strong>'?
          </p>
        </Modal>
      )}

      {/* Modal to display the temporary password */}
      {isModalTempPswActive && (
        <Modal
          title="Temporary Password"
          confirmButtonLabel="Close"
          onConfirm={() => {
            setIsModalTempPswActive(false);
            setTemporaryPassword(null);
            setActiveUserId(null);
          }}
          className="gap-4"
        >
          <Grid className="text-center">
            <Grid>
              <p>
                The user <strong>{username}</strong> has a new temporary password:
              </p>
              <div className="group relative w-1/2 justify-self-center">
                <pre
                  className="clickable min-w-max cursor-pointer select-all rounded-md bg-[color:--bg-quaternary] p-3 font-mono"
                  onClick={() => {
                    if (temporaryPassword) {
                      navigator.clipboard.writeText(temporaryPassword);
                      toast.success("Temporary password copied to clipboard");
                    }
                  }}
                >
                  {temporaryPassword}
                </pre>
                <Subtitle className="select-none text-[color:--text-secondary]" text="Click to copy" />
              </div>
            </Grid>

            <p>
              Please share this temporary password securely with the user.
              <br />
              Once you close this modal, the <strong>password will no longer be accessible</strong>.
            </p>
          </Grid>
        </Modal>
      )}

      {/* Delete Confirmation Modal */}
      {isModalTrashActive && (
        <Modal
          title="Please confirm: action irreversible"
          confirmButtonLabel="Confirm"
          onConfirm={handleDeleteUser}
          onCancel={handleModalAction}
        >
          <p>Are you sure you want to delete this user?</p>
        </Modal>
      )}

      <PageHeader icon={mdiAccountMultiple} title="Users">
        <Button icon={mdiPlus} text="New user" small onClick={() => navigate("/users/new")} />
      </PageHeader>

      {/* Users Table */}
      <Table
        loading={loadingUsers}
        data={users.sort(sortBy("username", { caseInsensitive: true })).map(user => ({
          Username: user.username,
          Role: user.role,
          Customers:
            user.role === "admin"
              ? "All"
              : user.customers
                  ?.map(customer => customer.name)
                  .sort()
                  .join(" | "),
          Active: Date.parse(user.disabled_at) > Date.now() ? "Yes" : "No",
          buttons: (
            <Buttons noWrap key={user.id}>
              <Button
                variant="tertiary"
                icon={mdiAccountEdit}
                onClick={() => openEditModal(user)}
                small
                title="Edit user"
              />
              <Button
                variant="warning"
                icon={mdiLockReset}
                onClick={() => openResetPswModal(user)}
                small
                title="Reset password"
              />
              <Button
                variant="danger"
                icon={mdiTrashCan}
                onClick={() => openDeleteModal(user.id)}
                small
                title="Delete user"
              />
            </Buttons>
          ),
        }))}
        perPageCustom={50}
        maxWidthColumns={{ Customers: "20rem" }}
      />
    </div>
  );
}
