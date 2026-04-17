import React, { useEffect, useRef, useState } from "react";
import Grid from "../Composition/Grid";
import UploadFile from "./UploadFile";

type UploadImageProps = {
  label?: string;
  onChange: (file: File | null) => void;
  previewHeight?: number;
  name?: string;
};

export default function UploadImage({
  label = "Choose Image",
  onChange,
  previewHeight = 200,
  name = "image",
}: UploadImageProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [filename, setFilename] = useState<string>("");
  const [previewUrl, setPreviewUrl] = useState<string>();

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    if (!["image/png", "image/jpeg"].includes(file.type)) {
      e.target.value = "";
      return;
    }

    setFilename(file.name);
    setPreviewUrl(URL.createObjectURL(file));
    onChange(file);
  };

  const clearImage = (e?: React.MouseEvent) => {
    e?.preventDefault();
    if (inputRef.current) inputRef.current.value = "";
    setFilename("");
    setPreviewUrl(undefined);
    onChange(null);
  };

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    const file = e.dataTransfer.files[0];
    if (file && ["image/png", "image/jpeg"].includes(file.type)) {
      setFilename(file.name);
      setPreviewUrl(URL.createObjectURL(file));
      onChange(file);
    }
  };

  const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
  };

  useEffect(() => {
    const handlePaste = (e: ClipboardEvent) => {
      const items = e.clipboardData?.items;
      if (!items) return;

      for (const item of items) {
        if (item.kind === "file") {
          const file = item.getAsFile();
          if (file && ["image/png", "image/jpeg"].includes(file.type)) {
            setFilename(file.name);
            setPreviewUrl(URL.createObjectURL(file));
            onChange(file);
          }
        }
      }
    };

    document.addEventListener("paste", handlePaste);
    return () => document.removeEventListener("paste", handlePaste);
  }, [onChange]);

  useEffect(() => {
    return () => {
      if (previewUrl) URL.revokeObjectURL(previewUrl);
    };
  }, [previewUrl]);

  return (
    <div onDrop={handleDrop} onDragOver={handleDragOver}>
      <Grid>
        <UploadFile
          label={label}
          inputId="image-upload-input"
          filename={filename}
          inputRef={inputRef}
          name={name}
          accept={"image/png, image/jpeg"}
          onChange={handleFileChange}
          onButtonClick={clearImage}
        />
        {previewUrl && (
          <img
            src={previewUrl}
            alt="Selected image preview"
            className="justify-self-center"
            style={{ maxHeight: `${previewHeight}px` }}
          />
        )}
      </Grid>
    </div>
  );
}
