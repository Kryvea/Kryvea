import React, { useEffect, useRef, useState } from "react";
import Grid from "../Composition/Grid";
import UploadFile from "./UploadFile";

const IMAGE_TYPES = ["image/png", "image/jpeg"];

export const ACCEPTED_IMAGE_TYPES = IMAGE_TYPES.join(", ");

export const isAcceptedImageFile = (file: File) => IMAGE_TYPES.includes(file.type);

export const imageFileFromItems = (items?: DataTransferItemList | null): File | undefined => {
  for (const item of items ?? []) {
    if (item.kind === "file") {
      const file = item.getAsFile();
      if (file && isAcceptedImageFile(file)) {
        return file;
      }
    }
  }
  return undefined;
};

type UseImageUploadOptions = {
  initialFilename?: string;
  onSelect?: (file: File, objectUrl: string) => void;
  onClear?: () => void;
};

export function useImageUpload({ initialFilename = "", onSelect, onClear }: UseImageUploadOptions = {}) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [filename, setFilename] = useState<string>(initialFilename);
  const [imageUrl, setImageUrl] = useState<string>();

  // Latest-callback refs: async consumers (blob fetches) must not act on props
  // captured at first render, or a reordered/deleted item receives the update.
  const onSelectRef = useRef(onSelect);
  const onClearRef = useRef(onClear);
  onSelectRef.current = onSelect;
  onClearRef.current = onClear;

  // The hook owns the current preview URL; ownedUrlRef tracks what to revoke.
  const ownedUrlRef = useRef<string>(undefined);

  useEffect(() => {
    return () => {
      if (ownedUrlRef.current) {
        URL.revokeObjectURL(ownedUrlRef.current);
      }
    };
  }, []);

  const setPreviewUrl = (url?: string) => {
    if (ownedUrlRef.current) {
      URL.revokeObjectURL(ownedUrlRef.current);
    }
    ownedUrlRef.current = url;
    setImageUrl(url);
  };

  const selectImageFile = (file: File) => {
    const objectUrl = URL.createObjectURL(file);
    setFilename(file.name);
    setPreviewUrl(objectUrl);
    onSelectRef.current?.(file, objectUrl);
  };

  const onInputChange = ({ target }: React.ChangeEvent<HTMLInputElement>) => {
    const file = target.files?.[0];
    if (!file || !isAcceptedImageFile(file)) {
      target.value = "";
      return;
    }

    selectImageFile(file);
  };

  const reset = (revokeUrl: boolean) => {
    if (inputRef.current) {
      inputRef.current.value = "";
    }
    if (!revokeUrl) {
      ownedUrlRef.current = undefined;
    }
    setFilename("");
    setPreviewUrl(undefined);
    onClearRef.current?.();
  };

  const clearImage = (e?: { preventDefault?: () => void }) => {
    e?.preventDefault?.();
    reset(true);
  };

  // Like clearImage, but transfers ownership of the object URL to the caller
  // (who inserted it somewhere that outlives this hook) instead of revoking it.
  const releaseImage = () => reset(false);

  return { inputRef, filename, imageUrl, selectImageFile, onInputChange, clearImage, releaseImage };
}

type ImageUploadFieldProps = {
  inputId: string;
  upload: ReturnType<typeof useImageUpload>;
};

export function ImageUploadField({ inputId, upload }: ImageUploadFieldProps) {
  const { inputRef, filename, imageUrl, onInputChange, clearImage } = upload;

  return (
    <Grid>
      <UploadFile
        label="Choose Image"
        inputId={inputId}
        filename={filename}
        inputRef={inputRef}
        name="imagePoc"
        accept={ACCEPTED_IMAGE_TYPES}
        onChange={onInputChange}
        onButtonClick={clearImage}
      />
      {imageUrl && <img src={imageUrl} alt="Selected image preview" className="max-h-[550px] w-fit object-contain" />}
    </Grid>
  );
}
