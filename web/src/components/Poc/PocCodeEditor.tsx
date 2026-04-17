import { mdiBroom, mdiClipboardText, mdiEraser, mdiMarker, mdiPalette } from "@mdi/js";
import type * as monaco from "monaco-editor";
import { useCallback, useContext, useState } from "react";
import { GlobalContext } from "../../App";
import {
  emptyCurry as curryEmptyFunc,
  onelineJsonBody as oneLineJsonBody,
  prettifyJsonBody,
} from "../../utils/helpers";
import DescribedCode from "../Composition/DescribedCode";
import Grid from "../Composition/Grid";
import Modal from "../Composition/Modal";
import Button from "../Form/Button";
import Buttons from "../Form/Buttons";
import Checkbox from "../Form/Checkbox";
import ColorPicker from "../Form/ColorPicker";
import { SelectOption } from "../Form/SelectWrapper.types";
import MonacoCodeEditor from "./MonacoCodeEditor";
import { MonacoTextSelection } from "./MonacoCodeEditor.types";
import { PocDoc } from "./Poc.types";

type PocCodeEditorProps = {
  label?: string;
  pocDoc: PocDoc;
  currentIndex;
  highlightsProperty;
  code: string;
  disableViewHighlights: boolean;
  selectedLanguage: string;
  ideStartingLineNumber?: number;
  textHighlights?: MonacoTextSelection[];
  onChange?: (value: string) => void;
  onSetCodeSelection?: (currentIndex: number, property: string, textSelection: MonacoTextSelection[]) => void;
  onLanguageOptionsInit?: (options: SelectOption[]) => void;
  options?: monaco.editor.IStandaloneEditorConstructionOptions;
  lineWrapId?: string;
};

type Position = { line: number; col: number };

function getOffset(fullText: string, line: number, col: number): number {
  const lines = fullText.split("\n");
  return lines.slice(0, line - 1).reduce((acc, l) => acc + l.length + 1, 0) + (col - 1);
}

function offsetToPos(fullText: string, offset: number): Position {
  const lines = fullText.split("\n");
  let acc = 0;
  for (let i = 0; i < lines.length; i++) {
    const len = lines[i].length + 1;
    if (acc + len > offset) return { line: i + 1, col: offset - acc + 1 };
    acc += len;
  }
  return { line: lines.length, col: lines[lines.length - 1].length + 1 };
}

/** @warning This component could probably be better, for instance it is not fully and properly typed, probably it is best to not be used outside of the pocs */
export default function PocCodeEditor({
  label = "",
  pocDoc,
  currentIndex,
  selectedLanguage,
  highlightsProperty,
  code,
  disableViewHighlights,
  ideStartingLineNumber,
  textHighlights = [],
  onChange = () => {},
  onSetCodeSelection = () => {},
  onLanguageOptionsInit = () => {},
  options = {},
  lineWrapId = "",
}: PocCodeEditorProps) {
  const [selectedText, setSelectedText] = useState<MonacoTextSelection[]>([]);
  const [showHighligtedTextModal, setShowHighlightedTextModal] = useState(false);
  const [minimap, setMinimap] = useState(false);
  const [formattingWarning, setFormattingWarning] = useState(false);
  const [doFormat, setDoFormat] = useState(curryEmptyFunc);
  const {
    useCtxCodeHighlightColor: [ctxCodeHighlightColor, setCtxCodeHighlightColor],
    useCtxLinewrap: [ctxLineWrap, setCtxLineWrap],
  } = useContext(GlobalContext);

  function subtractSelection(
    hl: MonacoTextSelection,
    erase: MonacoTextSelection,
    fullText: string
  ): MonacoTextSelection[] {
    const hlStart = getOffset(fullText, hl.start.line, hl.start.col);
    const hlEnd = getOffset(fullText, hl.end.line, hl.end.col);
    const eraseStart = getOffset(fullText, erase.start.line, erase.start.col);
    const eraseEnd = getOffset(fullText, erase.end.line, erase.end.col);

    // No overlap → keep highlight
    if (eraseEnd <= hlStart || eraseStart >= hlEnd) return [hl];

    const result: MonacoTextSelection[] = [];

    // Left fragment
    if (eraseStart > hlStart) {
      const end = offsetToPos(fullText, eraseStart);
      result.push({
        ...hl,
        start: hl.start,
        end,
        selectionPreview: fullText.slice(hlStart, eraseStart),
      });
    }

    // Right fragment
    if (eraseEnd < hlEnd) {
      const start = offsetToPos(fullText, eraseEnd);
      result.push({
        ...hl,
        start,
        end: hl.end,
        selectionPreview: fullText.slice(eraseEnd, hlEnd),
      });
    }

    // remove ghost whitespaces highlights
    return result.filter(r => r.selectionPreview.trim() !== "");
  }

  function mergeHighlights(highlights: MonacoTextSelection[], fullText: string): MonacoTextSelection[] {
    if (highlights.length === 0) return [];

    const sorted = [...highlights].sort(
      (a, b) => getOffset(fullText, a.start.line, a.start.col) - getOffset(fullText, b.start.line, b.start.col)
    );

    const merged: MonacoTextSelection[] = [sorted[0]];

    for (let i = 1; i < sorted.length; i++) {
      const last = merged[merged.length - 1];
      const curr = sorted[i];

      // Only merge if same color
      if (last.color !== curr.color) {
        merged.push(curr);
        continue;
      }

      const lastStart = getOffset(fullText, last.start.line, last.start.col);
      const lastEnd = getOffset(fullText, last.end.line, last.end.col); // exclusive-style
      const currStart = getOffset(fullText, curr.start.line, curr.start.col);
      const currEnd = getOffset(fullText, curr.end.line, curr.end.col);

      // Merge only when overlapping or exactly adjacent (no gap)
      if (currStart <= lastEnd) {
        // new end is the furthest end
        const newEndOffset = Math.max(lastEnd, currEnd);
        const newEndPos = offsetToPos(fullText, newEndOffset);

        merged[merged.length - 1] = {
          ...last,
          end: newEndPos,
          selectionPreview: fullText.slice(lastStart, newEndOffset),
        };
      } else {
        // there's a gap: keep separate
        merged.push(curr);
      }
    }

    return merged;
  }

  const prepareFormattingWith = useCallback(
    formatFn => () => {
      const [httpWithFormattedBody] = formatFn(code);
      setDoFormat(() => () => onChange(httpWithFormattedBody));
      setFormattingWarning(true);
    },
    [code]
  );

  return (
    <Grid className="gap-4">
      {showHighligtedTextModal && (
        <Modal
          title="Code that will be highlighted"
          subtitle="Click on a selected text to remove it"
          onCancel={() => setShowHighlightedTextModal(false)}
        >
          <Grid>
            {pocDoc[highlightsProperty]?.map((highlight: MonacoTextSelection, i) => {
              const {
                start: { line, col },
                selectionPreview: text,
              } = highlight;
              const codeSelectionKey = `poc-${currentIndex}-code-selection-${i}-${pocDoc.key}`;
              return (
                <Button
                  className="border border-[color:--border-secondary] hover:bg-red-400/20"
                  variant="secondary"
                  onClick={() =>
                    onSetCodeSelection(
                      currentIndex,
                      highlightsProperty,
                      pocDoc[highlightsProperty].filter((_, j) => i !== j)
                    )
                  }
                  key={codeSelectionKey}
                >
                  <DescribedCode className="p-2" subtitle={`line ${line} col ${col}`} text={text} />
                </Button>
              );
            })}
          </Grid>
        </Modal>
      )}
      {formattingWarning && (
        <Modal
          title="JSON formatting"
          confirmButtonLabel="Confirm"
          onCancel={() => {
            setFormattingWarning(false);
            setDoFormat(curryEmptyFunc);
          }}
          onConfirm={() => {
            doFormat();
            setDoFormat(curryEmptyFunc);
            setFormattingWarning(false);
          }}
        >
          <p className="text-[color:--error]">
            <strong>Warning:</strong> This action <em>cannot be undone</em> and <u>will remove highlights</u> in the
            json.
          </p>
        </Modal>
      )}
      <Buttons containerClassname="flex-grow" className="justify-between">
        <Buttons>
          <Button
            disabled={selectedText.length < 1}
            small
            variant="warning"
            title="Add highlight"
            icon={mdiMarker}
            iconSize={24}
            customColor={ctxCodeHighlightColor}
            onClick={() => {
              const colored = selectedText.map(sel => ({
                ...sel,
                color: ctxCodeHighlightColor,
              }));

              const merged = mergeHighlights([...(pocDoc[highlightsProperty] ?? []), ...colored], code);

              onSetCodeSelection(currentIndex, highlightsProperty, merged);
            }}
          />
          <Button
            disabled={selectedText.length < 1}
            small
            variant="secondary"
            title="Erase highlight"
            icon={mdiEraser}
            iconSize={24}
            onClick={() => {
              const highlights = pocDoc[highlightsProperty] ?? [];

              let newHighlights = highlights;
              for (const erase of selectedText) {
                const updated: MonacoTextSelection[] = [];
                for (const hl of newHighlights) {
                  updated.push(...subtractSelection(hl, erase, code));
                }
                newHighlights = updated;
              }

              onSetCodeSelection(currentIndex, highlightsProperty, newHighlights);
            }}
          />
          <ColorPicker
            icon={mdiPalette}
            title="Highlight color"
            value={ctxCodeHighlightColor}
            onChange={setCtxCodeHighlightColor}
          />
          <Button
            small
            disabled={disableViewHighlights}
            variant="outline-only"
            title="Show all selections"
            icon={mdiClipboardText}
            iconSize={24}
            onClick={() => setShowHighlightedTextModal(true)}
          />
          <Checkbox
            id={`poc-${pocDoc.index}-${lineWrapId}-line-wrap`}
            label="Line wrap"
            onChange={e => setCtxLineWrap(e.target.checked)}
            checked={ctxLineWrap}
          />
          <Checkbox
            id={`poc-${pocDoc.index}-minimap`}
            label="Minimap"
            onChange={e => setMinimap(e.target.checked)}
            checked={minimap}
          />
          {selectedLanguage === "http" && (
            <>
              <Button
                small
                variant="outline-only"
                text="Prettify JSON"
                onClick={prepareFormattingWith(prettifyJsonBody)}
              />
              <Button
                small
                variant="outline-only"
                text="One-Liner JSON"
                onClick={prepareFormattingWith(oneLineJsonBody)}
              />
            </>
          )}
        </Buttons>

        <Button
          small
          variant="danger"
          title="Clear highlights"
          icon={mdiBroom}
          iconSize={24}
          onClick={() => {
            onSetCodeSelection(currentIndex, highlightsProperty, []);
          }}
        />
      </Buttons>

      <MonacoCodeEditor
        label={label}
        value={code}
        ideStartingLineNumber={ideStartingLineNumber}
        textHighlights={formattingWarning ? [] : textHighlights}
        removeDisappearedHighlights={indexes => {
          const filteredHighlights = pocDoc[highlightsProperty]?.filter((_, i) => !indexes.includes(i));
          onSetCodeSelection(currentIndex, highlightsProperty, filteredHighlights);
        }}
        onTextSelection={setSelectedText}
        language={selectedLanguage}
        onLanguageOptionsInit={onLanguageOptionsInit}
        onChange={onChange}
        options={{ ...options, wordWrap: ctxLineWrap ? "on" : "off", minimap: { enabled: minimap } }}
      />
    </Grid>
  );
}
