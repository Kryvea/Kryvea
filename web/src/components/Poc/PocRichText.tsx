import { mdiPencil } from "@mdi/js";
import React, { memo } from "react";
import Label from "../Form/Label";
import Textarea from "../Form/Textarea";
import { PocRichTextDoc } from "./Poc.types";
import PocTemplate from "./PocTemplate";
import RichText from "./RichText";

type PocRichTextProps = {
  pocDoc: PocRichTextDoc;
  currentIndex: number;
  pocListLength: number;
  onPositionChange: (currentIndex: number) => (newIndex: number) => void;
  onTextChange: <T>(currentIndex: number, key: keyof Omit<T, "key">) => (e: React.ChangeEvent) => void;
  onValueChange: <T>(currentIndex: number, key: keyof Omit<T, "key">, value: string) => void;
  onRemovePoc: (currentIndex: number) => () => void;
  selectedPoc: number;
  setSelectedPoc: (index: number) => void;
};

export default memo(function PocRichText({
  pocDoc,
  currentIndex,
  pocListLength,
  onPositionChange,
  onTextChange,
  onValueChange,
  onRemovePoc,
  selectedPoc,
  setSelectedPoc,
}: PocRichTextProps) {
  const descriptionTextareaId = `poc-description-${currentIndex}-${pocDoc.key}`;
  const textInputId = `poc-richtext-${currentIndex}-${pocDoc.key}`;

  return (
    <PocTemplate
      {...{
        pocDoc,
        currentIndex,
        pocListLength,
        icon: mdiPencil,
        onPositionChange,
        onRemovePoc,
        selectedPoc,
        setSelectedPoc,
        title: "Rich Text",
      }}
    >
      <div className="poc-richtext col-span-8 grid">
        <Label htmlFor={descriptionTextareaId} text="Description" />
        <Textarea
          value={pocDoc.description}
          id={descriptionTextareaId}
          onChange={onTextChange<PocRichTextDoc>(currentIndex, "description")}
        />
      </div>

      <div className="col-span-8 mt-4 grid w-full max-w-full">
        <Label htmlFor={textInputId} text="Rich Text Content" />
        <RichText
          content={pocDoc.rich_text_data}
          onChange={html => onValueChange<PocRichTextDoc>(currentIndex, "rich_text_data", html)}
        />
      </div>
    </PocTemplate>
  );
});
