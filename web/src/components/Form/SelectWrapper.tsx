import { useLayoutEffect, useRef, useState } from "react";
import Select, { ActionMeta, InputActionMeta } from "react-select";
import makeAnimated from "react-select/animated";
import Grid from "../Composition/Grid";
import Label from "./Label";
import { SelectOption } from "./SelectWrapper.types";

interface CommonProps {
  options: SelectOption[];
  defaultValue?: SelectOption | SelectOption[];
  onInputChange?: (input: string, actionMeta?: InputActionMeta) => any;
  disabled?: boolean;
  small?: boolean;
  widthFixed?: boolean;
  closeMenuOnSelect?: boolean;
  id?: string;
  label?: string;
  className?: string;
  isClearable?: boolean;
}
interface SelectWrapperSingleProps extends CommonProps {
  isMulti?: false | undefined;
  value?: SelectOption;
  onChange: (newValue: SelectOption, actionMeta: ActionMeta<any>) => any;
}

interface SelectWrapperMultiProps extends CommonProps {
  isMulti: true;
  value?: SelectOption[];
  onChange: (newValue: SelectOption[], actionMeta: ActionMeta<any>) => any;
}
export default function SelectWrapper({
  options,
  value,
  defaultValue,
  onChange,
  onInputChange,
  isMulti,
  disabled,
  small,
  widthFixed,
  closeMenuOnSelect,
  id,
  label,
  className,
  isClearable,
}: SelectWrapperSingleProps | SelectWrapperMultiProps) {
  const [inputValue, setInputValue] = useState("");
  const [width, setWidth] = useState<number>(0);
  const measureRef = useRef<HTMLSpanElement>(null);

  useLayoutEffect(() => {
    if (measureRef.current) {
      setWidth(measureRef.current.offsetWidth + 48); // extra space for arrow + padding
    }
  }, []);

  const handleOnInputChange = (input: string, actionMeta: InputActionMeta) => {
    setInputValue(input);
    if (onInputChange) {
      onInputChange(input, actionMeta);
    }
  };

  const onChangeWrapper = (newValue: any, actionMeta: ActionMeta<any>) => {
    if (isMulti && newValue.some(option => option.value === "all")) {
      newValue = options.filter(option => option.value !== "all");
      onChange(newValue, actionMeta);
      return;
    }
    onChange(newValue, actionMeta);
  };

  if (isMulti) {
    options = [{ label: "Select all", value: "all" }, ...options];
  }
  if (isMulti && value.length === options.length - 1) {
    options = [];
  }

  const animatedComponents = makeAnimated();

  const longestLabel = options.length ? options.reduce((a, b) => (a.label.length > b.label.length ? a : b)).label : "";

  const longestLabelFixedWidth = widthFixed
    ? {
        container: base => ({
          ...base,
          width,
        }),
        control: base => ({
          ...base,
          width: "100%",
          minWidth: width,
        }),
        menu: base => ({
          ...base,
          width: "100%",
          minWidth: width,
        }),
      }
    : {};

  const controlStyleSmall = small
    ? {
        control: (base: any) => ({
          ...base,
          padding: 0,
          paddingLeft: "8px",
          paddingRight: "8px",
        }),
      }
    : {};

  const combinedStyles = {
    menuPortal: (base: any) => ({ ...base, zIndex: 10 }),
    ...longestLabelFixedWidth,
    ...controlStyleSmall,
  };

  return (
    <Grid>
      {label && <Label disabled={disabled} text={label} htmlFor={id} />}
      {widthFixed && (
        <span
          ref={measureRef}
          style={{
            position: "fixed",
            visibility: "hidden",
            whiteSpace: "nowrap",
            fontFamily: "inherit",
            fontSize: "inherit",
            fontWeight: "inherit",
          }}
        >
          {longestLabel}
        </span>
      )}
      <Select
        {...{
          options,
          value,
          inputValue,
          defaultValue,
          onChange: onChangeWrapper,
          onInputChange: handleOnInputChange,
          isMulti,
          isDisabled: disabled,
          closeMenuOnSelect,
          inputId: id,
          classNamePrefix: "select-wrapper",
          className: `select-wrapper-class ${className}`,
        }}
        isClearable={isClearable}
        unstyled
        components={animatedComponents}
        menuPortalTarget={document.body}
        styles={combinedStyles}
      />
    </Grid>
  );
}
