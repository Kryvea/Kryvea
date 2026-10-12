import { mdiBroom, mdiClipboardText, mdiEraser, mdiMarker, mdiPalette } from "@mdi/js";
import type * as monaco from "monaco-editor";
import { useContext, useMemo, useState } from "react";
import { GlobalContext } from "../../App";
import { emptyCurry, onelineJsonBody, prettifyJsonBody } from "../../utils/helpers";
import DescribedCode from "../Composition/DescribedCode";
import Grid from "../Composition/Grid";
import Modal from "../Composition/Modal";
import Button from "../Form/Button";
import Buttons from "../Form/Buttons";
import Checkbox from "../Form/Checkbox";
import ColorPicker from "../Form/ColorPicker";
import { SelectOption } from "../Form/SelectWrapper.types";
import MonacoCodeEditor from "./MonacoCodeEditor";
import { LineAndCol, MonacoTextSelection } from "./MonacoCodeEditor.types";
import { PocDoc } from "./Poc.types";

type PocCodeEditorProps = {
  pocDoc: PocDoc;
  currentIndex: number;
  highlightsProperty: "request_highlights" | "response_highlights" | "text_highlights";
  code: string;
  disableViewHighlights: boolean;
  selectedLanguage: string;
  ideStartingLineNumber?: number;
  textHighlights?: MonacoTextSelection[];
  onChange?: (value: string) => void;
  onSetCodeSelection?: (currentIndex: number, property: string, textSelection: MonacoTextSelection[]) => void;
  onLanguageOptionsInit?: (options: SelectOption[]) => void;
  lineWrapId?: string;
};

const NO_HIGHLIGHTS: MonacoTextSelection[] = [];

type LineIndex = {
  getOffset: (line: number, col: number) => number;
  offsetToPos: (offset: number) => LineAndCol;
};

/** Precomputes a line-offset table so offset<->position conversions don't re-split the document per call. */
function buildLineIndex(fullText: string): LineIndex {
  const lineStarts: number[] = [0];
  for (let i = 0; i < fullText.length; i++) {
    if (fullText[i] === "\n") {
      lineStarts.push(i + 1);
    }
  }

  // Clamp the line: a stale highlight may point past the end of the current text.
  const getOffset = (line: number, col: number) =>
    lineStarts[Math.min(Math.max(line, 1), lineStarts.length) - 1] + (col - 1);

  const offsetToPos = (offset: number): LineAndCol => {
    // Binary search for the last line starting at or before the offset
    let lo = 0;
    let hi = lineStarts.length - 1;
    while (lo < hi) {
      const mid = (lo + hi + 1) >> 1;
      if (lineStarts[mid] <= offset) {
        lo = mid;
      } else {
        hi = mid - 1;
      }
    }
    const lineEnd = lo + 1 < lineStarts.length ? lineStarts[lo + 1] - 1 : fullText.length;
    return { line: lo + 1, col: Math.min(offset, lineEnd) - lineStarts[lo] + 1 };
  };

  return { getOffset, offsetToPos };
}

function subtractSelection(
  hl: MonacoTextSelection,
  erase: MonacoTextSelection,
  fullText: string,
  { getOffset, offsetToPos }: LineIndex
): MonacoTextSelection[] {
  const hlStart = getOffset(hl.start.line, hl.start.col);
  const hlEnd = getOffset(hl.end.line, hl.end.col);
  const eraseStart = getOffset(erase.start.line, erase.start.col);
  const eraseEnd = getOffset(erase.end.line, erase.end.col);

  // No overlap → keep highlight
  if (eraseEnd <= hlStart || eraseStart >= hlEnd) return [hl];

  const result: MonacoTextSelection[] = [];

  // Left fragment
  if (eraseStart > hlStart) {
    const end = offsetToPos(eraseStart);
    result.push({
      ...hl,
      start: hl.start,
      end,
      selectionPreview: fullText.slice(hlStart, eraseStart),
    });
  }

  // Right fragment
  if (eraseEnd < hlEnd) {
    const start = offsetToPos(eraseEnd);
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

  const { getOffset, offsetToPos } = buildLineIndex(fullText);

  const sorted = [...highlights].sort(
    (a, b) => getOffset(a.start.line, a.start.col) - getOffset(b.start.line, b.start.col)
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

    const lastStart = getOffset(last.start.line, last.start.col);
    const lastEnd = getOffset(last.end.line, last.end.col); // exclusive-style
    const currStart = getOffset(curr.start.line, curr.start.col);
    const currEnd = getOffset(curr.end.line, curr.end.col);

    // Merge only when overlapping or exactly adjacent (no gap)
    if (currStart <= lastEnd) {
      // new end is the furthest end
      const newEndOffset = Math.max(lastEnd, currEnd);
      const newEndPos = offsetToPos(newEndOffset);

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

export default function PocCodeEditor({
  pocDoc,
  currentIndex,
  selectedLanguage,
  highlightsProperty,
  code,
  disableViewHighlights,
  ideStartingLineNumber,
  textHighlights = NO_HIGHLIGHTS,
  onChange = () => {},
  onSetCodeSelection = () => {},
  onLanguageOptionsInit = () => {},
  lineWrapId = "",
}: PocCodeEditorProps) {
  const [selectedText, setSelectedText] = useState<MonacoTextSelection[]>([]);
  const [showHighlightedTextModal, setShowHighlightedTextModal] = useState(false);
  const [minimap, setMinimap] = useState(false);
  const [formattingWarning, setFormattingWarning] = useState(false);
  const [doFormat, setDoFormat] = useState(emptyCurry);
  const {
    useCtxCodeHighlightColor: [ctxCodeHighlightColor, setCtxCodeHighlightColor],
    useCtxLinewrap: [ctxLineWrap, setCtxLineWrap],
  } = useContext(GlobalContext);

  const prepareFormattingWith = (formatFn: (http: string) => [string, number]) => () => {
    const [httpWithFormattedBody] = formatFn(code);
    setDoFormat(() => () => onChange(httpWithFormattedBody));
    setFormattingWarning(true);
  };

  const editorOptions = useMemo<monaco.editor.IStandaloneEditorConstructionOptions>(
    () => ({ wordWrap: ctxLineWrap ? "on" : "off", minimap: { enabled: minimap } }),
    [ctxLineWrap, minimap]
  );

  return (
    <Grid className="gap-4">
      {showHighlightedTextModal && (
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
            setDoFormat(emptyCurry);
          }}
          onConfirm={() => {
            doFormat();
            setDoFormat(emptyCurry);
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
              const lineIndex = buildLineIndex(code);

              let newHighlights = highlights;
              for (const erase of selectedText) {
                const updated: MonacoTextSelection[] = [];
                for (const hl of newHighlights) {
                  updated.push(...subtractSelection(hl, erase, code, lineIndex));
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
                onClick={prepareFormattingWith(onelineJsonBody)}
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
        value={code}
        ideStartingLineNumber={ideStartingLineNumber}
        textHighlights={formattingWarning ? NO_HIGHLIGHTS : textHighlights}
        removeDisappearedHighlights={indexes => {
          const filteredHighlights = pocDoc[highlightsProperty]?.filter((_, i) => !indexes.includes(i));
          onSetCodeSelection(currentIndex, highlightsProperty, filteredHighlights);
        }}
        onTextSelection={setSelectedText}
        language={selectedLanguage}
        onLanguageOptionsInit={onLanguageOptionsInit}
        onChange={onChange}
        options={editorOptions}
      />
    </Grid>
  );
}
