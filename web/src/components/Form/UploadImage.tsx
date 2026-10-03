import React, { useEffect } from "react";
import Grid from "../Composition/Grid";
import { ACCEPTED_IMAGE_TYPES, imageFileFromItems, isAcceptedImageFile, useImageUpload } from "./ImageUpload";
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
  const upload = useImageUpload({
    onSelect: file => onChange(file),
    onClear: () => onChange(null),
  });

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    const file = e.dataTransfer.files[0];
    if (file && isAcceptedImageFile(file)) {
      upload.selectImageFile(file);
    }
  };

  const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
  };

  useEffect(() => {
    const handlePaste = (e: ClipboardEvent) => {
      const file = imageFileFromItems(e.clipboardData?.items);
      if (file) {
        upload.selectImageFile(file);
      }
    };

    document.addEventListener("paste", handlePaste);
    return () => document.removeEventListener("paste", handlePaste);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div onDrop={handleDrop} onDragOver={handleDragOver}>
      <Grid>
        <UploadFile
          label={label}
          inputId="image-upload-input"
          filename={upload.filename}
          inputRef={upload.inputRef}
          name={name}
          accept={ACCEPTED_IMAGE_TYPES}
          onChange={upload.onInputChange}
          onButtonClick={upload.clearImage}
        />
        {upload.imageUrl && (
          <img
            src={upload.imageUrl}
            alt="Selected image preview"
            className="justify-self-center"
            style={{ maxHeight: `${previewHeight}px` }}
          />
        )}
      </Grid>
    </div>
  );
}
