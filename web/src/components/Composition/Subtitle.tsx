export default function Subtitle({ disabled = undefined, className = "", text }) {
  return (
    <span data-disabled={disabled} className={`place-self-start text-xs font-light ${className}`}>
      {text === "" ? <>&nbsp;</> : text}
    </span>
  );
}
