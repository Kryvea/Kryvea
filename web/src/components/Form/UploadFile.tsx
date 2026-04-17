import { mdiTrashCan } from "@mdi/js";
import Grid from "../Composition/Grid";
import Button from "./Button";
import Label from "./Label";

type UploadFileProps = {
  label?: string;
  inputId: string;
  filename: string;
  inputRef?;
  name: string;
  accept: string;
  onChange;
  onButtonClick;
};

export default function UploadFile({
  label,
  inputId,
  filename,
  inputRef,
  name,
  accept,
  onChange,
  onButtonClick,
}: UploadFileProps) {
  const hasFile = Boolean(filename);

  return (
    <Grid>
      {label && <Label text={label} htmlFor={inputId} />}
      <div className="relative min-w-full">
        <label
          htmlFor={inputId}
          className="clickable flex h-10 w-full cursor-pointer items-center gap-2 overflow-hidden rounded-lg bg-[color:--bg-quaternary] p-2 pr-10"
        >
          <span className="shrink-0 text-nowrap rounded-md border border-[color:--border-primary] bg-[color:--bg-tertiary] px-[6px] py-[1px]">
            Choose File
          </span>
          <span className="truncate before:empty:font-thin before:empty:text-[color:--text-secondary] before:empty:content-['No_file_chosen']">
            {filename}
          </span>
        </label>
        <input
          ref={inputRef}
          className="hidden"
          type="file"
          name={name}
          accept={accept}
          id={inputId}
          onChange={onChange}
          onClick={e => {
            (e.target as HTMLInputElement).value = "";
          }}
        />

        {hasFile && (
          <Button
            className="absolute right-0 top-1/2 -translate-y-1/2 rounded-none rounded-r-md p-[10px]"
            variant="danger"
            onClick={onButtonClick}
            title="Clear file"
            icon={mdiTrashCan}
          />
        )}
      </div>
    </Grid>
  );
}
