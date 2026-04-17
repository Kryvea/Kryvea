import { useEffect, useRef, useState } from "react";
import { HexColorPicker } from "react-colorful";
import Button from "./Button";
import Input from "./Input";

export default function ColorPicker({
  value,
  icon,
  onChange,
  title = "Pick a color",
}: {
  value: string;
  icon?: string;
  onChange: (color: string) => void;
  title?: string;
}) {
  const [open, setOpen] = useState(false);
  const pickerRef = useRef<HTMLDivElement>(null);
  const [hexInput, setHexInput] = useState(value);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (pickerRef.current && !pickerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  const handleChange = (hex: string) => {
    setHexInput(hex);
    onChange(hex);
  };

  const handleHexChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const hex = e.target.value.startsWith("#") ? e.target.value : `#${e.target.value}`;
    setHexInput(hex);
    if (/^#([0-9A-Fa-f]{6}|[0-9A-Fa-f]{8})$/.test(hex)) {
      onChange(hex);
    }
  };

  return (
    <div className="relative inline-block" ref={pickerRef}>
      <Button variant="tertiary" icon={icon} title={title} onClick={() => setOpen(o => !o)} />

      {open && (
        <div className="absolute z-20 flex flex-col gap-2 rounded p-2">
          <HexColorPicker color={hexInput} onChange={handleChange} />
          <Input
            type="text"
            value={hexInput}
            className="px-2 py-1 text-center text-sm"
            onChange={handleHexChange}
            placeholder="#rrggbb"
          />
        </div>
      )}
    </div>
  );
}
