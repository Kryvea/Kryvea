import { mdiCableData } from "@mdi/js";
import React, { memo } from "react";
import Grid from "../Composition/Grid";
import Input from "../Form/Input";
import Label from "../Form/Label";
import Textarea from "../Form/Textarea";
import { MonacoTextSelection } from "./MonacoCodeEditor.types";
import { PocRequestResponseDoc } from "./Poc.types";
import PocCodeEditor from "./PocCodeEditor";
import PocTemplate from "./PocTemplate";

type PocRequestResponseProps = {
  pocDoc: PocRequestResponseDoc;
  currentIndex: number;
  pocListLength: number;
  selectedPoc: number;
  setSelectedPoc: (index: number) => void;
  onPositionChange: (currentIndex: number) => (newIndex: number) => void;
  onTextChange: <T>(currentIndex: number, key: keyof Omit<T, "key">) => (e: React.ChangeEvent) => void;
  onValueChange: <T>(currentIndex: number, key: keyof Omit<T, "key">, value: string) => void;
  onRemovePoc: (currentIndex: number) => () => void;
  onSetCodeSelection: <T>(
    currentIndex: number,
    property: keyof Omit<T, "key">,
    textSelection: MonacoTextSelection[]
  ) => void;
};

const variants = [
  { label: "Request", property: "request", highlightsProperty: "request_highlights" },
  { label: "Response", property: "response", highlightsProperty: "response_highlights" },
] as const;

export default memo(function PocRequestResponse({
  pocDoc,
  currentIndex,
  pocListLength,
  selectedPoc,
  setSelectedPoc,
  onPositionChange,
  onTextChange,
  onValueChange,
  onRemovePoc,
  onSetCodeSelection,
}: PocRequestResponseProps) {
  const descriptionTextareaId = `poc-description-${currentIndex}-${pocDoc.key}`;
  const urlInputId = `poc-url-${currentIndex}-${pocDoc.key}`;

  return (
    <PocTemplate
      {...{
        pocDoc,
        currentIndex,
        pocListLength,
        icon: mdiCableData,
        onPositionChange,
        onRemovePoc,
        selectedPoc,
        setSelectedPoc,
        title: "Request/Response",
      }}
    >
      <Textarea
        label="Description"
        value={pocDoc.description}
        id={descriptionTextareaId}
        onChange={onTextChange<PocRequestResponseDoc>(currentIndex, "description")}
      />

      <Input
        type="text"
        label="URL"
        id={urlInputId}
        value={pocDoc.uri}
        onChange={onTextChange<PocRequestResponseDoc>(currentIndex, "uri")}
      />

      <Grid className="grid-cols-1 gap-4 2xl:grid-cols-2">
        {variants.map(({ label, property, highlightsProperty }) => (
          <Grid key={property}>
            <Label text={label} />
            <PocCodeEditor
              pocDoc={pocDoc}
              disableViewHighlights={(pocDoc[highlightsProperty] ?? []).length <= 0}
              currentIndex={currentIndex}
              highlightsProperty={highlightsProperty}
              code={pocDoc[property]}
              selectedLanguage="http"
              ideStartingLineNumber={1}
              textHighlights={pocDoc[highlightsProperty]}
              lineWrapId={property}
              onChange={code => onValueChange<PocRequestResponseDoc>(currentIndex, property, code)}
              onSetCodeSelection={onSetCodeSelection}
            />
          </Grid>
        ))}
      </Grid>
    </PocTemplate>
  );
});
