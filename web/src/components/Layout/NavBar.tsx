import {
  mdiAccount,
  mdiChevronDoubleRight,
  mdiFullscreen,
  mdiFullscreenExit,
  mdiLogout,
  mdiMonitor,
  mdiWeatherNight,
  mdiWhiteBalanceSunny,
} from "@mdi/js";
import { useContext, useMemo } from "react";
import { useNavigate } from "react-router";
import { GlobalContext } from "../../App";
import { postData } from "../../api/api";
import Icon from "../Composition/Icon";
import Button from "../Form/Button";
import Buttons from "../Form/Buttons";
import Breadcrumb from "./Breadcrumb";

export default function NavBar() {
  const {
    useCtxUsername: [ctxUsername, setCtxUsername],
    useThemeMode: [themeMode, setThemeMode],
    useFullscreen: [fullscreen, setFullScreen],
  } = useContext(GlobalContext);

  const navigate = useNavigate();

  const handleLogout = () => {
    postData("/api/logout", undefined, () => {
      navigate("/login", { replace: false });
      setCtxUsername(undefined);
    });
  };

  const titleTheme = useMemo(() => {
    switch (themeMode) {
      case "os":
        return "OS Theme";
      case "light":
        return "Light Theme";
      case "dark":
        return "Dark Theme";
    }
  }, [themeMode]);

  return (
    <nav className="navbar">
      <Breadcrumb
        homeElement={"Home"}
        separator={<Icon viewBox="4 4 20 7.5" path={mdiChevronDoubleRight} />}
        capitalizeLinks
      />
      <Buttons noWrap>
        <Button
          onClick={() => navigate("/profile")}
          icon={mdiAccount}
          text={ctxUsername}
          className="bg-transparent p-2 text-[color:--link]"
        />
        <Button
          onClick={() => setThemeMode(prev => (prev === "light" ? "dark" : prev === "dark" ? "os" : "light"))}
          className="relative bg-transparent text-[color:--link]"
          title={titleTheme}
        >
          <Icon
            path={mdiWhiteBalanceSunny}
            className={`absolute left-0 top-0 opacity-0 ${themeMode === "light" ? "rotateFadeIn" : ""}`}
          />
          <Icon
            path={mdiWeatherNight}
            className={`absolute left-0 top-0 opacity-0 ${themeMode === "dark" ? "rotateFadeIn" : ""}`}
          />
          <Icon
            path={mdiMonitor}
            className={`absolute left-0 top-0 opacity-0 ${themeMode === "os" ? "rotateFadeIn" : ""}`}
          />
        </Button>
        <Button
          onClick={() => setFullScreen(prev => !prev)}
          icon={fullscreen ? mdiFullscreenExit : mdiFullscreen}
          className="bg-transparent !pl-3 !pr-0 text-[color:--link]"
          title={`${fullscreen ? "Exit fullscreen" : "Fullscreen"}`}
        />
        <Button
          onClick={handleLogout}
          icon={mdiLogout}
          text="Logout"
          className="bg-transparent p-2 text-[color:--link]"
        />
      </Buttons>
    </nav>
  );
}
