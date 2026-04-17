import { mdiPencil } from "@mdi/js";
import { ChangeEvent, useState } from "react";
import { Keys } from "../../types/utils.types";
import Flex from "../Composition/Flex";
import Grid from "../Composition/Grid";
import Input from "../Form/Input";
import SelectWrapper from "../Form/SelectWrapper";
import { SelectOption } from "../Form/SelectWrapper.types";
import Textarea from "../Form/Textarea";
import { MonacoTextSelection } from "./MonacoCodeEditor.types";
import { PocDoc, PocTextDoc } from "./Poc.types";
import PocCodeEditor from "./PocCodeEditor";
import PocTemplate from "./PocTemplate";

type PocTextProps = {
  pocDoc: PocTextDoc;
  currentIndex;
  pocList: PocDoc[];
  selectedPoc: number;
  setSelectedPoc: (index: number) => void;
  onPositionChange: (currentIndex: number) => (e: ChangeEvent<HTMLInputElement>) => void;
  onTextChange: <T>(currentIndex, key: keyof Omit<T, "key">) => (e: ChangeEvent) => void;
  onRemovePoc: (currentIndex: number) => void;
  onSetCodeSelection: <T>(
    currentIndex: number,
    property: keyof Omit<T, "key">,
    textSelection: MonacoTextSelection[]
  ) => void;
  onStartingLineNumberChange?: <T>(currentIndex, property: keyof Omit<T, "key">) => (num: number) => void;
};

export default function PocText({
  pocDoc,
  currentIndex,
  pocList,
  selectedPoc,
  setSelectedPoc,
  onPositionChange,
  onTextChange,
  onRemovePoc,
  onSetCodeSelection,
  onStartingLineNumberChange,
}: PocTextProps) {
  const [languageOptions, setLanguageOptions] = useState<SelectOption[]>([]);
  const [selectedLanguage, setSelectedLanguage] = useState<SelectOption>({
    label: pocDoc.text_language || "Plaintext",
    value: pocDoc.text_language || "plaintext",
  });

  const descriptionTextareaId = `poc-description-${currentIndex}-${pocDoc.key}`;
  const uriInputId = `poc-uri-${currentIndex}-${pocDoc.key}`;
  const languageInputId = `poc-language-${currentIndex}-${pocDoc.key}`;
  const startingLineNumberId = `poc-starting-line-number-${currentIndex}-${pocDoc.key}`;

  // Should only be called once when the MonacoEditor initializes
  const onLanguageOptionsInit = options => {
    setLanguageOptions(options);
    setSelectedLanguage({
      label: options.find(option => option.value === pocDoc.text_language)?.label,
      value: pocDoc.text_language || "plaintext",
    });
  };

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
        title: "Text",
      }}
    >
      <Grid className="poc-text">
        <Textarea
          label="Description"
          value={pocDoc.description}
          id={descriptionTextareaId}
          onChange={onTextChange<PocTextDoc>(currentIndex, "description")}
        />
        <Input
          type="text"
          label="URI"
          id={uriInputId}
          value={pocDoc.uri}
          onChange={onTextChange<PocTextDoc>(currentIndex, "uri")}
        />
        <Grid className="gap-4">
          <Flex className="gap-4">
            <SelectWrapper
              label="Language"
              className="w-64"
              options={languageOptions.map(({ label, value }) => ({ label, value }))}
              value={{
                label: selectedLanguage.label || pocDoc.text_language,
                value: selectedLanguage.value || pocDoc.text_language,
              }}
              onChange={selected => {
                setSelectedLanguage(selected);
                onTextChange<PocTextDoc>(
                  currentIndex,
                  "text_language"
                )({
                  target: { value: selected.value },
                } as any);
              }}
              id={languageInputId}
            />
            <Input
              className="w-[55px] rounded text-center"
              label="Start"
              type="number"
              value={pocDoc?.starting_line_number}
              min={1}
              onChange={onStartingLineNumberChange<PocTextDoc>(currentIndex, "starting_line_number")}
              id={startingLineNumberId}
            />
          </Flex>

          <PocCodeEditor
            pocDoc={pocDoc}
            selectedLanguage={selectedLanguage.value}
            ideStartingLineNumber={pocDoc?.starting_line_number}
            textHighlights={pocDoc.text_highlights}
            code={pocDoc.text_data}
            disableViewHighlights={(pocDoc?.text_highlights ?? []).length <= 0}
            currentIndex={currentIndex}
            highlightsProperty={"text_highlights" as Keys<PocTextDoc>}
            lineWrapId="text"
            onSetCodeSelection={onSetCodeSelection}
            onChange={code =>
              onTextChange<PocTextDoc>(
                currentIndex,
                "text_data"
              )({
                target: { value: code },
              } as any)
            }
            onLanguageOptionsInit={onLanguageOptionsInit}
          />
        </Grid>
      </Grid>
    </PocTemplate>
  );
}
