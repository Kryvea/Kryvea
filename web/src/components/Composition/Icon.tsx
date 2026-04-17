import { ReactNode } from "react";

type Props = {
  path: string;
  size?: string | number | null;
  className?: string;
  children?: ReactNode;
  viewBox?: string;
};

export default function Icon({ path, size = 16, className = "", children, viewBox = "0 0 24 24" }: Props) {
  return (
    <span className={`inline-flex items-center justify-center ${className}`}>
      <svg viewBox={viewBox} width={size} height={size} className="inline-block" aria-hidden="true" focusable="false">
        <path fill="currentColor" d={path} />
      </svg>
      {children}
    </span>
  );
}
