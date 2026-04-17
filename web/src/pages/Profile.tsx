import { useContext, useEffect, useState } from "react";
import { toast } from "react-toastify";
import { getData, patchData } from "../api/api";
import { GlobalContext } from "../App";
import Card from "../components/Composition/Card";
import CardTitle from "../components/Composition/CardTitle";
import Divider from "../components/Composition/Divider";
import Grid from "../components/Composition/Grid";
import PageHeader from "../components/Composition/PageHeader";
import Button from "../components/Form/Button";
import Input from "../components/Form/Input";
import { User } from "../types/common.types";
import { getPageTitle } from "../utils/helpers";

export default function Profile() {
  const {
    useCtxUsername: [, setCtxUsername],
  } = useContext(GlobalContext);

  const [user, setUser] = useState<User | null>(null);
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");

  useEffect(() => {
    document.title = getPageTitle("Profile");
    getData<User>("/api/users/me", setUser);
  }, []);

  const handleUsernameSubmit = () => {
    if (!user) {
      toast.error("Username is required");
      return;
    }

    patchData<{ message: string }>("/api/users/me", { username: user.username.trim() }, () => {
      toast.success("Username updated successfully");
      setCtxUsername(user.username);
    });
  };

  const handlePasswordSubmit = () => {
    if (!currentPassword || !newPassword || !confirmPassword) {
      toast.error("Please fill in all required fields");
      return;
    }
    if (newPassword !== confirmPassword) {
      toast.error("New password and confirmation do not match");
      return;
    }

    const payload = {
      current_password: currentPassword,
      new_password: newPassword,
    };

    patchData<{ message: string }>("/api/users/me", payload, () => {
      toast.success("Password updated successfully");
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    });
  };

  return (
    <div>
      <PageHeader title="Profile" />
      <Grid className="grid-cols-2 !items-start gap-4">
        <Card>
          <CardTitle title={"Change username"} />
          <form
            onSubmit={e => {
              e.preventDefault();
              handleUsernameSubmit();
            }}
          >
            <Grid>
              <Input
                type="text"
                id="username"
                label="Username"
                value={user?.username}
                helperSubtitle="Required"
                onChange={e => setUser({ ...user, username: e.target.value })}
              />
              <Input
                disabled
                type="datetime-local"
                id="disable_date"
                label="Account will be disabled on"
                value={user?.disabled_at.substring(0, 16)}
                helperSubtitle=""
              />
              <Input disabled type="text" id="role" label="Role" value={user?.role} helperSubtitle="" />
              <Divider />
            </Grid>
            <Button text="Update" formSubmit />
          </form>
        </Card>
        <Card>
          <CardTitle title={"Change password"} />
          <form
            onSubmit={e => {
              e.preventDefault();
              handlePasswordSubmit();
            }}
          >
            <Grid>
              <Input
                type="password"
                id="current_password"
                label="Current password"
                helperSubtitle="Required"
                value={currentPassword}
                onChange={e => setCurrentPassword(e.target.value)}
              />
              <Input
                type="password"
                id="new_password"
                label="New password"
                helperSubtitle="Required"
                value={newPassword}
                onChange={e => setNewPassword(e.target.value)}
              />
              <Input
                type="password"
                id="confirm_password"
                label="Confirm password"
                helperSubtitle="Required"
                value={confirmPassword}
                onChange={e => setConfirmPassword(e.target.value)}
              />
              <Divider />
            </Grid>
            <Button text="Update" formSubmit />
          </form>
        </Card>
      </Grid>
    </div>
  );
}
