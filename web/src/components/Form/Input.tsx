import { mdiEye, mdiEyeOff } from "@mdi/js";
import { HTMLInputTypeAttribute, useEffect, useState } from "react";
import Grid from "../Composition/Grid";
import Subtitle from "../Composition/Subtitle";
import Button from "./Button";
import Label from "./Label";

interface BaseInputProps {
  id?: string;
  className?: string;
  label?: string;
  helperSubtitle?: string;
  placeholder?: string;
  value?: string | number;
  autoFocus?: boolean;
  disabled?: boolean;
  name?: string;
  onEnter?: (e: React.KeyboardEvent<HTMLInputElement>) => void;
}

interface InputProps extends BaseInputProps {
  type: HTMLInputTypeAttribute;
  step?: undefined;
  min?: undefined;
  max?: undefined;
  accept?: undefined;
  onChange?: any;
}

interface FileInputProps extends BaseInputProps {
  type: "file";
  step?: undefined;
  min?: undefined;
  max?: undefined;
  accept?: string;
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

interface NumberInputProps extends BaseInputProps {
  type: "number";
  step?: number;
  min?: number;
  max?: number;
  accept?: string;
  onChange?: (value: number) => void;
}

export default function Input({
  className,
  disabled,
  type,
  step,
  id,
  label,
  helperSubtitle,
  placeholder,
  value,
  min,
  max,
  accept,
  autoFocus,
  name,
  onChange,
  onEnter,
}: InputProps | FileInputProps | NumberInputProps) {
  const [numberPreview, setNumberPreview] = useState(value);
  const [showPassword, setShowPassword] = useState(false);

  useEffect(() => {
    setNumberPreview(value);
  }, [value]);

  return (
    <Grid>
      {label && <Label text={label} htmlFor={id} />}
      <div className="grid">
        {type === "number" ? (
          <input
            disabled={disabled}
            className={className}
            type={type}
            step={step}
            id={id}
            placeholder={placeholder}
            value={numberPreview}
            accept={accept}
            autoFocus={autoFocus}
            name={name}
            onChange={e => {
              const val = e.target.value;
              setNumberPreview(val);
            }}
            onKeyDown={e => {
              switch (e.key) {
                case "Escape":
                  setNumberPreview(value);
                  break;
                case "Enter":
                  e.currentTarget.blur();
                  onEnter?.(e);
                  break;
              }
            }}
            onBlur={e => {
              let value = e.currentTarget.value;
              let num = parseInt(value);
              if (step != undefined) {
                num = parseFloat(value);
              }

              if (+value === 0) {
                num = 0;
              }
              if (value !== "0") {
                value = value.replace(/^0/, "");
              }

              if (min !== undefined && num < min) {
                num = min;
              }
              if (max !== undefined && num > max) {
                num = max;
              }

              setNumberPreview(num.toString());
              onChange(num);
            }}
          />
        ) : (
          <>
            <div className="relative w-full">
              <input
                disabled={disabled}
                className={`${type === "password" ? "pr-10" : ""} w-full ${className}`}
                type={type === "password" ? (showPassword ? "text" : "password") : type}
                id={id}
                placeholder={placeholder}
                value={value}
                accept={accept}
                autoFocus={autoFocus}
                name={name}
                onChange={onChange}
                onKeyDown={e => {
                  switch (e.key) {
                    case "Enter":
                      onEnter?.(e);
                      break;
                  }
                }}
              />
              {type === "password" && (
                <Button
                  disabled={disabled}
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-2 top-1/2 -translate-y-1/2 transform cursor-pointer p-1"
                  icon={showPassword ? mdiEye : mdiEyeOff}
                  variant="transparent"
                  title={showPassword ? "Hide password" : "Show password"}
                  tabIndex={-1}
                />
              )}
            </div>
            {(helperSubtitle || helperSubtitle === "") && <Subtitle disabled={disabled} text={helperSubtitle} />}
          </>
        )}
      </div>
    </Grid>
  );
}
