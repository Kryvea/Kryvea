import { mdiImage } from "@mdi/js";
import React, { useCallback, useEffect, useRef, useState } from "react";
import { getBlob } from "../../api/api";
import { uuidZero } from "../../types/common.types";
import Grid from "../Composition/Grid";
import Input from "../Form/Input";
import Textarea from "../Form/Textarea";
import UploadFile from "../Form/UploadFile";
import { PocDoc, PocImageDoc } from "./Poc.types";
import PocTemplate from "./PocTemplate";

export type PocImageProps = {
  pocDoc: PocImageDoc;
  currentIndex;
  pocList: PocDoc[];
  onPositionChange: (currentIndex: number) => (e: React.ChangeEvent<HTMLInputElement>) => void;
  onTextChange: <T>(currentIndex, key: keyof Omit<T, "key">) => (e: React.ChangeEvent) => void;
  onImageChange: (currentIndex, image: File) => void;
  onRemovePoc: (currentIndex: number) => void;
  selectedPoc: number;
  setSelectedPoc: (index: number) => void;
};

export default function PocImage({
  pocDoc,
  currentIndex,
  pocList,
  onPositionChange,
  onTextChange,
  onImageChange,
  onRemovePoc,
  selectedPoc,
  setSelectedPoc,
}: PocImageProps) {
  const [imageUrl, setImageUrl] = useState<string>();
  const [filename, setFilename] = useState<string>(pocDoc?.image_filename);
  const imageInput = useRef<HTMLInputElement>(null);

  const handleDrop = pocTemplateRef => (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();

    document.dispatchEvent(new MouseEvent("dragend", { bubbles: true }));
    pocTemplateRef.current?.classList.remove("dragged-over");

    const files = e.dataTransfer.files;
    if (files.length > 0) {
      const file = files[0];
      if (file.type === "image/png" || file.type === "image/jpeg") {
        const reader = new FileReader();
        reader.readAsDataURL(file);
      }

      setFilename(file.name);
      setImageUrl(URL.createObjectURL(file));
      onImageChange(currentIndex, file);
    }
  };

  const blobToFile = useCallback((blob: Blob, filename: string): File => {
    return new File([blob], filename, { type: blob.type });
  }, []);

  useEffect(() => {
    if (pocDoc.image_id == undefined || pocDoc.image_id === uuidZero) {
      return;
    }
    getBlob(`/api/files/images/${pocDoc.image_id}`, data => {
      onImageChangeWrapper({ target: { files: [blobToFile(data, pocDoc.image_filename)] } });
    });
  }, []);

  useEffect(() => {
    if (selectedPoc !== currentIndex) {
      return;
    }
    const handlePaste = (e: ClipboardEvent) => {
      const items = e.clipboardData?.items;
      if (!items) {
        return;
      }

      for (const item of items) {
        if (item.kind === "file") {
          const file = item.getAsFile();

          if (!file || (file.type !== "image/png" && file.type !== "image/jpeg")) {
            continue;
          }

          setFilename(file.name);
          setImageUrl(URL.createObjectURL(file));
          onImageChange(currentIndex, file);
        }
      }
    };

    document.addEventListener("paste", handlePaste);
    return () => {
      document.removeEventListener("paste", handlePaste);
    };
  }, [selectedPoc, currentIndex]);

  useEffect(() => {
    return () => {
      if (!imageUrl) {
        return;
      }

      URL.revokeObjectURL(imageUrl);
    };
  }, [imageUrl]);

  const onImageChangeWrapper = ({ target: { files } }) => {
    if (!files || !files[0]) {
      return;
    }

    const image_file: File = files[0];

    // const checkFilenameDuplicate = (pocs: PocDoc[]) =>
    //   pocs.some((poc, i) => {
    //     if (poc.type !== POC_TYPE_IMAGE || image_file.name !== poc?.image_file?.name) {
    //       return false;
    //     }

    //     toast.error(`Image with name ${image_file.name} already exists in the list at index ${i + 1}.`);
    //     return true;
    //   });

    // if (checkFilenameDuplicate(pocList)) {
    //   imageInput.current.value = ""; // clean implicit default input change behaviour
    //   return;
    // }

    setFilename(image_file.name);
    setImageUrl(URL.createObjectURL(image_file));
    onImageChange(currentIndex, image_file);
  };

  const clearImage = e => {
    e.preventDefault();
    imageInput.current.value = "";
    setFilename("");
    setImageUrl(undefined);
    onImageChange(currentIndex, undefined);
  };

  const descriptionTextareaId = `poc-description-${currentIndex}-${pocDoc.key}`;
  const imageInputId = `poc-image-${currentIndex}-${pocDoc.key}`;
  const captionTextareaId = `poc-caption-${currentIndex}-${pocDoc.key}`;

  return (
    <PocTemplate
      {...{
        pocDoc,
        currentIndex,
        pocList,
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

      <Grid>
        <UploadFile
          label="Choose Image"
          inputId={imageInputId}
          filename={filename}
          inputRef={imageInput}
          name={"imagePoc"}
          accept={"image/png, image/jpeg"}
          onChange={onImageChangeWrapper}
          onButtonClick={clearImage}
        />
        {imageUrl && <img src={imageUrl} alt="Selected image preview" className="max-h-[550px] w-fit object-contain" />}
      </Grid>

      <Input
        type="text"
        label="Caption"
        value={pocDoc.image_caption}
        id={captionTextareaId}
        onChange={onTextChange<PocImageDoc>(currentIndex, "image_caption")}
      />
    </PocTemplate>
  );
}
