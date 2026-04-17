type FlexProps = {
  className?: string;
  col?: boolean;
  justify?: "normal" | "start" | "end" | "center" | "between" | "around" | "evenly" | "stretch";
  items?: "start" | "end" | "center" | "baseline" | "stretch";
  children;
};

export default function Flex({ className, col, justify = "normal", items = "start", children }: FlexProps) {
  return (
    <div className={`flex ${col ? "flex-col" : ""} justify-${justify} items-${items} ${className}`}>{children}</div>
  );
}
