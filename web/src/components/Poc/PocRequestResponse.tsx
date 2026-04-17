import { mdiCableData } from "@mdi/js";
import React from "react";
import Grid from "../Composition/Grid";
import Input from "../Form/Input";
import Label from "../Form/Label";
import Textarea from "../Form/Textarea";
import { MonacoTextSelection } from "./MonacoCodeEditor.types";
import { PocDoc, PocRequestResponseDoc } from "./Poc.types";
import PocCodeEditor from "./PocCodeEditor";
import PocTemplate from "./PocTemplate";

type PocRequestResponseProps = {
  pocDoc: PocRequestResponseDoc;
  currentIndex;
  pocList: PocDoc[];
  selectedPoc: number;
  setSelectedPoc: (index: number) => void;
  onPositionChange: (currentIndex: number) => (e: React.ChangeEvent<HTMLInputElement>) => void;
  onTextChange: <T>(currentIndex, key: keyof Omit<T, "key">) => (e: React.ChangeEvent) => void;
  onRemovePoc: (currentIndex: number) => void;
  onSetCodeSelection: <T>(
    currentIndex: number,
    property: keyof Omit<T, "key">,
    textSelection: MonacoTextSelection[]
  ) => void;
};

export default function PocRequestResponse({
  pocDoc,
  currentIndex,
  pocList,
  selectedPoc,
  setSelectedPoc,
  onPositionChange,
  onTextChange,
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
        pocList,
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
        <Grid>
          <Label text="Request" />
          <PocCodeEditor
            pocDoc={pocDoc}
            disableViewHighlights={(pocDoc?.request_highlights ?? []).length <= 0}
            currentIndex={currentIndex}
            highlightsProperty="request_highlights"
            code={pocDoc.request}
            selectedLanguage="http"
            ideStartingLineNumber={1}
            textHighlights={pocDoc.request_highlights}
            lineWrapId="request"
            onChange={code =>
              onTextChange<PocRequestResponseDoc>(currentIndex, "request")({ target: { value: code } } as any)
            }
            onSetCodeSelection={onSetCodeSelection}
          />
        </Grid>

        <Grid>
          <Label text="Response" />
          <PocCodeEditor
            pocDoc={pocDoc}
            disableViewHighlights={(pocDoc?.response_highlights ?? []).length <= 0}
            currentIndex={currentIndex}
            highlightsProperty="response_highlights"
            code={pocDoc.response}
            selectedLanguage="http"
            ideStartingLineNumber={1}
            textHighlights={pocDoc.response_highlights}
            lineWrapId="response"
            onChange={code =>
              onTextChange<PocRequestResponseDoc>(currentIndex, "response")({ target: { value: code } } as any)
            }
            onSetCodeSelection={onSetCodeSelection}
          />
        </Grid>
      </Grid>
    </PocTemplate>
  );
}
