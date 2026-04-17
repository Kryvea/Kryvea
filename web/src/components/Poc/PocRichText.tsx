import { mdiPencil } from "@mdi/js";
import { useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import React from "react";
import Label from "../Form/Label";
import Textarea from "../Form/Textarea";
import { PocDoc, PocRichTextDoc } from "./Poc.types";
import PocTemplate from "./PocTemplate";
import RichText from "./RichText";

type PocRichTextProps = {
  pocDoc: PocRichTextDoc;
  currentIndex;
  pocList: PocDoc[];
  onPositionChange: (currentIndex: number) => (e: React.ChangeEvent<HTMLInputElement>) => void;
  onTextChange: <T>(currentIndex, key: keyof Omit<T, "key">) => (e: React.ChangeEvent) => void;
  onRemovePoc: (currentIndex: number) => void;
  selectedPoc: number;
  setSelectedPoc: (index: number) => void;
};

export default function PocRichText({
  pocDoc,
  currentIndex,
  pocList,
  onPositionChange,
  onTextChange,
  onRemovePoc,
  selectedPoc,
  setSelectedPoc,
}: PocRichTextProps) {
  const descriptionTextareaId = `poc-description-${currentIndex}-${pocDoc.key}`;
  const textInputId = `poc-richtext-${currentIndex}-${pocDoc.key}`;

  const editor = useEditor({
    extensions: [StarterKit],
    content: pocDoc.rich_text_data || "",
    onUpdate: ({ editor }) => {
      const html = editor.getHTML();
      onTextChange<PocRichTextDoc>(
        currentIndex,
        "rich_text_data"
      )({
        target: { value: html },
      } as any);
    },
  });

  return (
    <PocTemplate
      {...{
        pocDoc,
        currentIndex,
        pocList,
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
        <RichText />
      </div>
    </PocTemplate>
  );
}
