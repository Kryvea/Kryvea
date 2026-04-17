export default function Grid({ className = "", children }) {
  return <div className={`grid items-end gap-2 ${className}`}>{children}</div>;
}
