import { mdiImage } from "@mdi/js";
import React, { memo, useEffect } from "react";
import { getBlob } from "../../api/api";
import { uuidZero } from "../../types/common.types";
import { ImageUploadField, imageFileFromItems, isAcceptedImageFile, useImageUpload } from "../Form/ImageUpload";
import Input from "../Form/Input";
import Textarea from "../Form/Textarea";
import { PocImageDoc } from "./Poc.types";
import PocTemplate from "./PocTemplate";

export type PocImageProps = {
  pocDoc: PocImageDoc;
  currentIndex: number;
  pocListLength: number;
  onPositionChange: (currentIndex: number) => (newIndex: number) => void;
  onTextChange: <T>(currentIndex: number, key: keyof Omit<T, "key">) => (e: React.ChangeEvent) => void;
  onImageChange: (currentIndex: number, image: File | undefined) => void;
  onRemovePoc: (currentIndex: number) => () => void;
  selectedPoc: number;
  setSelectedPoc: (index: number) => void;
};

const blobToFile = (blob: Blob, filename: string): File => new File([blob], filename, { type: blob.type });

export default memo(function PocImage({
  pocDoc,
  currentIndex,
  pocListLength,
  onPositionChange,
  onTextChange,
  onImageChange,
  onRemovePoc,
  selectedPoc,
  setSelectedPoc,
}: PocImageProps) {
  const upload = useImageUpload({
    initialFilename: pocDoc?.image_filename,
    onSelect: file => onImageChange(currentIndex, file),
    onClear: () => onImageChange(currentIndex, undefined),
  });

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    const files = e.dataTransfer.files;
    if (files.length > 0 && isAcceptedImageFile(files[0])) {
      upload.selectImageFile(files[0]);
    }
  };

  useEffect(() => {
    if (pocDoc.image_id == undefined || pocDoc.image_id === uuidZero) {
      return;
    }
    // The poc may be deleted while its blob is still downloading.
    let cancelled = false;
    getBlob(`/api/files/images/${pocDoc.image_id}`, data => {
      if (!cancelled) {
        upload.selectImageFile(blobToFile(data, pocDoc.image_filename));
      }
    });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (selectedPoc !== currentIndex) {
      return;
    }
    const handlePaste = (e: ClipboardEvent) => {
      const file = imageFileFromItems(e.clipboardData?.items);
      if (file) {
        upload.selectImageFile(file);
      }
    };

    document.addEventListener("paste", handlePaste);
    return () => {
      document.removeEventListener("paste", handlePaste);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedPoc, currentIndex]);

  const descriptionTextareaId = `poc-description-${currentIndex}-${pocDoc.key}`;
  const imageInputId = `poc-image-${currentIndex}-${pocDoc.key}`;
  const captionTextareaId = `poc-caption-${currentIndex}-${pocDoc.key}`;

  return (
    <PocTemplate
      {...{
        pocDoc,
        currentIndex,
        pocListLength,
        handleDrop,
        icon: mdiImage,
        onPositionChange,
        onRemovePoc,
        selectedPoc,
        setSelectedPoc,
        title: "Image",
      }}
    >
      <Textarea
        label="Description"
        value={pocDoc.description}
        id={descriptionTextareaId}
        onChange={onTextChange<PocImageDoc>(currentIndex, "description")}
      />

      <ImageUploadField inputId={imageInputId} upload={upload} />

      <Input
        type="text"
        label="Caption"
        value={pocDoc.image_caption}
        id={captionTextareaId}
        onChange={onTextChange<PocImageDoc>(currentIndex, "image_caption")}
      />
    </PocTemplate>
  );
});
