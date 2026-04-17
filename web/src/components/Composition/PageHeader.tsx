import { ReactNode } from "react";
import Icon from "./Icon";

type Props = {
  icon?: string;
  title: string;
  main?: boolean;
  children?: ReactNode;
};

export default function PageHeader({ icon, title, main = false, children }: Props) {
  return (
    <section className="mb-6 mt-2 flex items-end justify-between">
      <div className="flex items-center justify-start">
        {icon && main && <Icon path={icon} className="mr-3" />}
        {icon && !main && <Icon path={icon} className="mr-2" size="20" />}
        <h1 className={`${main ? "text-3xl" : "text-2xl"}`}>{title}</h1>
      </div>
      <div className="sticky right-0 bg-[color:--bg-primary]">{children}</div>
    </section>
  );
}
