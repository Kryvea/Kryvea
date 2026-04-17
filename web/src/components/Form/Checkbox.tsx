import Label from "./Label";

interface CheckboxProps {
  disabled?: boolean;
  id: string;
  checked?;
  onChange?;
  label: string;
}

export default function Checkbox({ disabled, id, checked, onChange, label }: CheckboxProps) {
  return (
    <div data-disabled={disabled} className="inline-flex items-center gap-2">
      <input
        disabled={disabled}
        type="checkbox"
        id={id}
        checked={checked}
        onChange={onChange}
        className="checkbox h-5 w-5 cursor-pointer"
      />
      <Label disabled={disabled} text={label} htmlFor={id} className="cursor-pointer !whitespace-normal text-sm" />
    </div>
  );
}
