import { ReactNode } from "react";
import Grid from "./Grid";

type Props = {
  className?: string;
  children: ReactNode;
  footer?: ReactNode;
  noHighlight?: boolean;
};

export default function Card({ className, children, footer, noHighlight }: Props) {
  return (
    <Grid className={`cardbox ${noHighlight ? `no-highlight` : ""} ${className}`}>
      {children}
      {footer && <div className="mt-4 justify-self-end">{footer}</div>}
    </Grid>
  );
}
