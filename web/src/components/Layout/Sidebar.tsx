import { mdiMenuOpen } from "@mdi/js";
import { Fragment, useContext, useMemo, useState } from "react";
import { Link, useNavigate } from "react-router";
import { GlobalContext } from "../../App";
import { getSidebarItems } from "../../utils/helpers";
import Flex from "../Composition/Flex";
import Icon from "../Composition/Icon";
import Button from "../Form/Button";
// @ts-ignore
import logo from "../../assets/logo.svg";

export default function Sidebar() {
  const [isCollapsed, setIsCollapsed] = useState(false);
  const [dropdownMenus, setDropdownMenus] = useState({
    Administration: false,
    Customers: false,
  });
  const {
    useCtxCustomer: [ctxCustomer],
    useCtxSelectedSidebarItemLabel: [ctxSelectedSidebarItem, setCtxSelectedSidebarItemLabel],
  } = useContext(GlobalContext);
  const navigate = useNavigate();

  const defaultMenu = useMemo(() => getSidebarItems(ctxCustomer, navigate), [ctxCustomer]);

  const iconSize = isCollapsed ? 22 : 18;

  return (
    <aside className={`layout-sidebar ${isCollapsed ? "w-min" : "min-w-[280px] max-w-[280px]"}`}>
      <Flex className="h-full w-full" col>
        {/* Header */}
        <header className={`flex items-center p-4 ${isCollapsed ? "justify-center" : "justify-between"} `}>
          {!isCollapsed && (
            <Link to="/dashboard" className="text-xl font-black">
              <Flex>
                <img className="w-7" src={logo} alt="" />
                ryvea
              </Flex>
            </Link>
          )}
          <Button
            onClick={() => setIsCollapsed(!isCollapsed)}
            className={`p-2 !text-[color:--link] ${isCollapsed ? "rotate-180" : ""}`}
            variant="transparent"
            icon={mdiMenuOpen}
            iconSize={22}
          />
        </header>

        {/* Content */}
        <Flex col className={`flex-1 gap-2 overflow-y-auto p-4`}>
          {defaultMenu.filter(Boolean).map(item =>
            item.menu == undefined ? (
              <Link
                className={`sidebar-item ${ctxSelectedSidebarItem === item.label ? "sidebar-item-active" : ""} ${isCollapsed ? "aspect-square h-12 justify-center" : "!pl-2"}`}
                to={item.href}
                onClick={() => setCtxSelectedSidebarItemLabel(item.label)}
                title={item.label}
                key={`sidebar-${item.label}`}
              >
                <Icon path={item.icon} size={iconSize} />
                <span className={isCollapsed ? "hidden" : ""}>{item.label}</span>
              </Link>
            ) : (
              <Fragment key={`sidebar-${item.label}`}>
                <a
                  className={`sidebar-item flex-col ${ctxSelectedSidebarItem === item.label ? "sidebar-item-active" : ""} ${isCollapsed ? "aspect-square justify-center" : "!pl-2"}`}
                  onClick={e => {
                    setDropdownMenus(prev => ({ ...prev, [item.label]: !prev[item.label] }));
                    if (item.href) {
                      navigate(item.href);
                      setCtxSelectedSidebarItemLabel("Assessments");
                    }
                  }}
                  title={`${item.label} menu`}
                >
                  <Flex className={`cursor-pointer gap-4 break-all ${isCollapsed ? "justify-center" : ""}`}>
                    <Icon path={item.icon} size={iconSize} />
                    <span className={isCollapsed ? "hidden" : ""}>{item.label}</span>
                  </Flex>
                </a>
                {(item.label === ctxCustomer?.name || dropdownMenus[item.label]) && (
                  <>
                    {item.menu.map(subItem => (
                      <Link
                        className={`sidebar-item ${ctxSelectedSidebarItem === subItem.label ? "sidebar-item-active" : ""} ${isCollapsed ? "ml-0 aspect-square justify-center" : "ml-4 !pl-2"}`}
                        to={subItem.href}
                        onClick={() => setCtxSelectedSidebarItemLabel(subItem.label)}
                        title={subItem.label}
                        key={`sidebar-${subItem.label}`}
                      >
                        <Icon path={subItem.icon} size={iconSize} />
                        <span className={isCollapsed ? "hidden" : ""}>{subItem.label}</span>
                      </Link>
                    ))}
                  </>
                )}
              </Fragment>
            )
          )}
        </Flex>
      </Flex>
    </aside>
  );
}
