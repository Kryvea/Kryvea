import { useContext, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "react-toastify";
import { getData, postData } from "../api/api";
import { getKryveaShadow } from "../api/cookie";
import { GlobalContext } from "../App";
import Card from "../components/Composition/Card";
import Flex from "../components/Composition/Flex";
import Grid from "../components/Composition/Grid";
import Subtitle from "../components/Composition/Subtitle";
import Button from "../components/Form/Button";
import Input from "../components/Form/Input";
import { User } from "../types/common.types";
import { getPageTitle } from "../utils/helpers";
// @ts-ignore
import logo from "../assets/logo_stroke.svg";

export default function Login() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  const [confirmPassword, setConfirmPassword] = useState("");

  const {
    useCtxUsername: [, setCtxUsername],
    useCtxLastPage: [ctxLastPage],
  } = useContext(GlobalContext);

  const navigate = useNavigate();

  useEffect(() => {
    const kryveaShadow = getKryveaShadow();

    if (kryveaShadow && kryveaShadow !== "password_expired") {
      navigate(ctxLastPage, { replace: true });
      return;
    }

    document.title = getPageTitle("Login");
  }, []);

  const fetchUserAndSetCtxUsername = async () => {
    await getData<User>("/api/users/me", user => setCtxUsername(user.username));
    navigate(ctxLastPage, { replace: true });
  };

  const handleSubmit = () => {
    setError("");
    postData("/api/login", { username, password }, fetchUserAndSetCtxUsername, err => {
      // Check for password expired case
      const data = err.response?.data as { error: string };
      if (data?.error === "Password expired") {
        setError("");
        // Clear password input for reset
        setPassword("");
      } else {
        setError(data?.error || "Login failed");
      }
    });
  };

  // Password reset submit handler
  const handlePasswordReset = () => {
    setError("");

    if (!password) {
      setError("New password is required");
      return;
    }

    if (password !== confirmPassword) {
      setError("Passwords do not match");
      return;
    }

    postData(
      "/api/password/reset",
      { password },
      () => {
        toast.success("Password change successful");
        fetchUserAndSetCtxUsername();
      },
      err => {
        setError(err.response?.data?.error || "Failed to reset password");
      }
    );
  };

  return (
    <Flex className="card-modal fixed min-h-screen w-screen gap-4" col justify="center" items="center">
      <img className="w-36" src={logo} alt="" />
      <Card className="glasscard">
        {getKryveaShadow() !== "password_expired" ? (
          // LOGIN FORM
          <form
            onSubmit={e => {
              e.preventDefault();
              handleSubmit();
            }}
          >
            <Grid className="gap-4 p-1">
              <Input
                type="text"
                id="username"
                label="Username"
                onChange={e => setUsername(e.target.value)}
                value={username}
                autoFocus
              />
              <Input
                type="password"
                id="password"
                label="Password"
                onChange={e => setPassword(e.target.value)}
                value={password}
              />
              <Subtitle className="text-[color:--error]" text={error} />
              <Button text="Login" className="justify-center" formSubmit />
            </Grid>
          </form>
        ) : (
          // PASSWORD RESET FORM
          <form
            onSubmit={e => {
              e.preventDefault();
              handlePasswordReset();
            }}
          >
            <Grid className="gap-4 p-1">
              <p className="text-center">
                Your password has expired.
                <br />
                Please enter a new password to reset it.
              </p>
              <Input
                type="password"
                id="password"
                label="New Password"
                onChange={e => setPassword(e.target.value)}
                value={password}
                autoFocus
              />
              <Input
                type="password"
                id="confirm_password"
                label="Confirm New Password"
                onChange={e => setConfirmPassword(e.target.value)}
                value={confirmPassword}
              />
              <Subtitle className="text-[color:--error]" text={error} />
              <Button text="Reset Password" className="justify-center" formSubmit />
            </Grid>
          </form>
        )}
      </Card>
    </Flex>
  );
}
