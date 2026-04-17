import Editor, { Monaco, OnMount } from "@monaco-editor/react";
import type * as monaco from "monaco-editor";
import { useEffect, useRef, useState } from "react";
import Grid from "../Composition/Grid";
import Label from "../Form/Label";
import { MonacoTextSelection } from "./MonacoCodeEditor.types";

interface MonacoCodeEditorProps {
  language?: string;
  label?: string;
  value?: string;
  theme?: string;
  ideStartingLineNumber?: number;
  height?: string;
  stopLineNumberAt?: number;
  textHighlights?: MonacoTextSelection[];
  removeDisappearedHighlights?: (indexes: number[]) => void;
  options?: monaco.editor.IStandaloneEditorConstructionOptions;
  onChange?: (value: string) => void;
  onLanguageOptionsInit?;
  onTextSelection?;
  color?: string;
}

export default function MonacoCodeEditor({
  language,
  label = "",
  value,
  theme = "vs-dark",
  ideStartingLineNumber = 1,
  height = "100%",
  stopLineNumberAt,
  textHighlights = [],
  removeDisappearedHighlights = () => {},
  options,
  onChange = () => {},
  onLanguageOptionsInit = () => {},
  onTextSelection = (x: MonacoTextSelection) => {},
}: MonacoCodeEditorProps) {
  const [editor, setEditor] = useState<monaco.editor.IStandaloneCodeEditor>();
  const decorationsRef = useRef<monaco.editor.IEditorDecorationsCollection | null>(null);
  const monacoRef = useRef<typeof monaco | null>(null);

  function getContrastTextColor(hexColor: string): string {
    const hex = hexColor.replace("#", "");
    const r = parseInt(hex.slice(0, 2), 16) / 255;
    const g = parseInt(hex.slice(2, 4), 16) / 255;
    const b = parseInt(hex.slice(4, 6), 16) / 255;

    const [R, G, B] = [r, g, b].map(c => (c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4)));
    const luminance = 0.2126 * R + 0.7152 * G + 0.0722 * B;

    return luminance > 0.179 ? "#000000" : "#ffffff";
  }

  function getOrCreateHighlightClass(hexValue: string): string {
    const hexValueClean = hexValue.replace("#", "");
    const className = `monaco-editor-highlight-${hexValueClean}`;

    if (!document.querySelector(`.${className}`)) {
      const textColor = getContrastTextColor(hexValue);

      const style = document.createElement("style");
      style.innerHTML = `
      .${className} {
        background-color: ${hexValue};
        color: ${textColor};
      }
    `;
      document.head.appendChild(style);
    }

    return className;
  }

  const highlightCode = () => {
    if (!editor) {
      return;
    }
    const model = editor.getModel();
    if (!model) {
      return;
    }

    const decorations: monaco.editor.IModelDeltaDecoration[] = textHighlights
      .filter(({ start, end, selectionPreview }) => {
        const range = new monacoRef.current!.Range(start.line, start.col, end.line, end.col);
        const actualText = model.getValueInRange(range);
        return actualText === selectionPreview;
      })
      .map(({ start, end, color }) => ({
        range: new monacoRef.current!.Range(start.line, start.col, end.line, end.col),
        options: {
          inlineClassName: getOrCreateHighlightClass(color),
          stickiness: monacoRef.current?.editor.TrackedRangeStickiness.NeverGrowsWhenTypingAtEdges,
        },
      }));

    if (!decorationsRef.current) {
      decorationsRef.current = editor.createDecorationsCollection(decorations);
    } else {
      decorationsRef.current.set(decorations);
    }
  };

  const checkDisappearedHighlights = () => {
    const model = editor.getModel();
    if (!model) {
      return;
    }
    const disappearedHighlightsIndexes = textHighlights.flatMap(({ start, end, selectionPreview }, i) => {
      const range = new monacoRef.current!.Range(start.line, start.col, end.line, end.col);
      const actualText = model.getValueInRange(range);
      const textHighlightMatches = actualText === selectionPreview;

      if (!textHighlightMatches) {
        return i;
      }

      return [];
    });

    if (disappearedHighlightsIndexes.length > 0) {
      removeDisappearedHighlights(disappearedHighlightsIndexes);
    }
  };

  useEffect(() => {
    highlightCode();
  }, [textHighlights, editor]);

  useEffect(() => {
    if (!editor) {
      return;
    }

    const defuseDisapparedHighlights = setTimeout(checkDisappearedHighlights, 500);
    return () => {
      clearTimeout(defuseDisapparedHighlights);
    };
  }, [value]);

  const handleBeforeMount = (monaco: Monaco) => {
    monacoRef.current = monaco;
    if (!monaco.languages.getLanguages().some(lang => lang.id === "http")) {
      monaco.languages.register({ id: "http", aliases: ["HTTP"] });
    }

    monaco.languages.setMonarchTokensProvider("http", {
      tokenizer: {
        root: [
          [/^HTTP\/\d\.\d \d{3} .*/, "keyword"],

          // Request methods
          [/^(GET|POST|PUT|DELETE|PATCH|OPTIONS|HEAD)\b/, "keyword"],

          // URLs
          [/https?:\/\/\S+/, "string"],

          // Headers
          [/^[\w-]+(?=:)/, "type.identifier"],

          // JSON key:value pairs inside payloads
          [/"(\w+)"\s*:/, "attribute.name"],

          // JSON strings
          [/"([^"\\]|\\.)*"/, "string"],

          // JSON numbers
          [/\b-?\d+(\.\d+)?([eE][+-]?\d+)?\b/, "number"],

          // JSON booleans and null
          [/\b(true|false|null)\b/, "keyword"],

          [
            /(<\?)(xml)(\s+version\s*=\s*)("\d\.\d")(\s*\?>)/,
            ["metatag", "metatag", "attribute.name", "string", "metatag"],
          ],

          // XML tags (start or end)
          [/<\/?[\w-]+(\s+[\w-]+(\s*=\s*(?:"[^"]*"|'[^']*'|[^\s"'=<>`]+))?)*\s*\/?>/, "tag"],

          // XML attribute names and values inside tags
          [/(\w+)(=)(".*?"|'.*?')/, ["attribute.name", "delimiter", "string"]],

          // URL encoded key=value pairs (application/x-www-form-urlencoded)
          [/\b[\w%.-]+=[\w%.-]*\b/, "string"],

          // Multipart boundaries (e.g. --boundary12345)
          [/^--[\w-]+$/, "delimiter"],

          // HTTP status line (response)
          [/^(HTTP\/\d\.\d)\s+(\d{3})\s+([^\r\n]+)/, ["keyword", "number", "string"]],

          // HTTP request line (already have methods, but matching full line)
          [/^(GET|POST|PUT|DELETE|PATCH|OPTIONS|HEAD)\s+\S+\s+HTTP\/\d\.\d/, "keyword"],

          // Strings
          [/"[^"]*"/, "string"],

          // Numbers
          [/\b\d+(\.\d+)?\b/, "number"],

          // Booleans and null
          [/\b(true|false|null)\b/, "keyword"],

          // Comments
          [/^#.*$/, "comment"],

          // Brackets
          [/[{}[\]]/, "delimiter.bracket"],
        ],
      },
    });

    monaco.languages.setLanguageConfiguration("http", {
      brackets: [
        ["{", "}"],
        ["[", "]"],
      ],
      autoClosingPairs: [
        { open: "{", close: "}" },
        { open: "[", close: "]" },
        { open: '"', close: '"' },
      ],
    });

    const sortAlphaNum = (a, b) =>
      (a.aliases?.[0] || a.id).localeCompare(b.aliases?.[0] || b.id, "en", { numeric: true });
    const languages = monaco.languages.getLanguages();
    onLanguageOptionsInit(
      languages.sort(sortAlphaNum).map(lang => ({
        label: lang.aliases?.[0] || lang.id,
        value: lang.id,
      }))
    );
  };

  const handleEditorMount: OnMount = editor => {
    setEditor(editor);

    editor.onDidChangeCursorSelection(e => {
      let { selection, secondarySelections } = e;
      const allCursorSelections = [selection, ...secondarySelections];

      const toMonacoTextSelection = (sel: monaco.Selection): MonacoTextSelection[] =>
        sel.startLineNumber === sel.endLineNumber && sel.startColumn === sel.endColumn
          ? []
          : [
              {
                start: { line: sel.startLineNumber, col: sel.startColumn },
                end: { line: sel.endLineNumber, col: sel.endColumn },
                selectionPreview: `${editor.getModel()?.getValueInRange(sel)}`,
              },
            ];

      const allSelections = allCursorSelections.flatMap(toMonacoTextSelection);

      onTextSelection(allSelections);
    });
  };

  return (
    <Grid>
      {label && <Label text={label} />}
      <div className="h-full min-h-[400px] w-full min-w-0 resize-y overflow-auto border border-[color:--border-primary]">
        <Editor
          height={height}
          language={language}
          value={value}
          theme={theme}
          onChange={val => onChange(val || "")}
          beforeMount={handleBeforeMount}
          onMount={handleEditorMount}
          options={{
            lineNumbers: i => (i >= stopLineNumberAt ? "" : `${i - 1 + ideStartingLineNumber}`),
            lineNumbersMinChars: 2,
            glyphMargin: false,
            scrollBeyondLastLine: false,
            selectOnLineNumbers: true,
            roundedSelection: true,
            readOnly: false,
            cursorStyle: "line",
            automaticLayout: true,
            wordWrap: "on",
            formatOnType: true,
            formatOnPaste: true,
            scrollbar: { alwaysConsumeMouseWheel: false },
            minimap: { enabled: true, renderCharacters: true },
            tabSize: 2,
            "semanticHighlighting.enabled": false,
            ...options,
          }}
        />
      </div>
    </Grid>
  );
}
