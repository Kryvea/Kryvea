import { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import { getData, postData } from "../api/api";
import Card from "../components/Composition/Card";
import Divider from "../components/Composition/Divider";
import Grid from "../components/Composition/Grid";
import PageHeader from "../components/Composition/PageHeader";
import Button from "../components/Form/Button";
import Buttons from "../components/Form/Buttons";
import Input from "../components/Form/Input";
import SelectWrapper from "../components/Form/SelectWrapper";
import { Customer } from "../types/common.types";
import { getPageTitle } from "../utils/helpers";

export default function AddUser() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState("");
  const [selectedCustomers, setSelectedCustomers] = useState<string[]>([]);
  const [customers, setCustomers] = useState<Customer[]>([]);

  const navigate = useNavigate();

  useEffect(() => {
    document.title = getPageTitle("User");

    getData<Customer[]>("/api/customers", setCustomers);
  }, []);

  // Prepare options for the customers select dropdown
  const customerOptions = customers.map(customer => ({
    label: customer.name,
    value: customer.id,
  }));

  // Handle changes in the customers multi-select
  const handleSelectChange = (selectedOptions: { value: string; label: string }[] | null) => {
    setSelectedCustomers(selectedOptions ? selectedOptions.map(option => option.value) : []);
  };

  const handleSubmit = () => {
    const payload = {
      username: username.trim(),
      password,
      role,
      customers: selectedCustomers,
    };
    postData("/api/admin/users", payload, () => navigate("/users"));
  };

  return (
    <div>
      <PageHeader title="New user" />
      <Card>
        <Grid className="gap-4">
          <Input
            type="text"
            label="Username"
            placeholder="username"
            id="username"
            value={username}
            onChange={e => setUsername(e.target.value)}
          />
          <Input
            type="password"
            label="Password"
            placeholder="password"
            id="password"
            value={password}
            onChange={e => setPassword(e.target.value)}
          />
          <SelectWrapper
            label="Role"
            id="role-selection"
            options={[
              { value: "admin", label: "Admin" },
              { value: "user", label: "User" },
            ]}
            closeMenuOnSelect
            onChange={option => setRole(option.value)}
            value={role ? { value: role, label: role } : null}
          />
          <SelectWrapper
            label="Customers"
            options={customerOptions}
            isMulti
            disabled={role === "admin"}
            value={customerOptions.filter(option => selectedCustomers.includes(option.value))}
            onChange={handleSelectChange}
            closeMenuOnSelect={false}
            id="customer-selection"
          />
          <Divider />
          <Buttons>
            <Button text="Submit" onClick={handleSubmit} />
            <Button variant="outline-only" text="Cancel" onClick={() => navigate("/users")} />
          </Buttons>
        </Grid>
      </Card>
    </div>
  );
}
