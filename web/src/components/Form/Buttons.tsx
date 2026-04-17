import Grid from "../Composition/Grid";
import Label from "./Label";

type Props = {
  type?: string;
  noWrap?: boolean;
  className?: string;
  label?: string;
  containerClassname?: string;
  children;
};

export default function Buttons({
  type = "justify-start",
  noWrap = false,
  children,
  className,
  label,
  containerClassname,
}: Props) {
  return (
    <Grid className={containerClassname}>
      {label && <Label text={label} />}
      <div
        className={`flex items-center gap-2 text-nowrap ${type} ${noWrap ? "flex-nowrap" : "flex-wrap"} ${className}`}
      >
        {children}
      </div>
    </Grid>
  );
}
