import { useCallback, useContext, useEffect, useRef } from "react";
import { Outlet } from "react-router";
import { GlobalContext } from "../../App";
import FooterBar from "./FooterBar";
import NavBar from "./NavBar";
import Sidebar from "./Sidebar";

export default function Layout() {
  const {
    useFullscreen: [fullscreen],
    useBrowser: [browser],
  } = useContext(GlobalContext);

  const mainRef = useRef<HTMLDivElement>(null);
  const contentRef = useRef<HTMLDivElement>(null);

  const getScrollOptions = useCallback<(delta: number) => ScrollToOptions>(
    delta => (browser === "Chrome" ? { top: delta, behavior: "instant" } : { top: delta * 4, behavior: "smooth" }),
    []
  );

  useEffect(() => {
    // If scrolled outside of main, scroll main instead
    function handleWheel(e: WheelEvent) {
      const main = mainRef.current;
      if (!main) {
        return;
      }
      if (main.contains(e.target as Node)) {
        return;
      }
      if (contentRef.current && contentRef.current.contains(e.target as Node)) {
        return;
      }
      if (((e.target as any).classList as DOMTokenList).contains("select-wrapper__option")) {
        return;
      }

      const delta = e.deltaY;
      const maxScroll = main.scrollHeight - main.clientHeight;

      const atTop = main.scrollTop <= 0;
      const atBottom = main.scrollTop >= maxScroll;

      if ((delta < 0 && !atTop) || (delta > 0 && !atBottom)) {
        main.scrollBy(getScrollOptions(delta));
        e.preventDefault();
      }
    }

    document.addEventListener("wheel", handleWheel, { passive: false });
    return () => {
      document.removeEventListener("wheel", handleWheel);
    };
  }, []);

  return (
    <div className={`layout-root ${fullscreen ? "layout-fullscreen" : ""}`}>
      <Sidebar />
      <div ref={contentRef} className="layout-content">
        <NavBar />
        <main ref={mainRef} className="layout-main">
          <Outlet />
          <FooterBar />
        </main>
      </div>
    </div>
  );
}
