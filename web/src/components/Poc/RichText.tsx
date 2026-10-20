import "../../css/styles/richtext.scss";

import {
  mdiCodeBraces,
  mdiCodeTags,
  mdiFormatAlignCenter,
  mdiFormatAlignJustify,
  mdiFormatAlignLeft,
  mdiFormatAlignRight,
  mdiFormatBold,
  mdiFormatColorHighlight,
  mdiFormatColorText,
  mdiFormatHeader1,
  mdiFormatHeader2,
  mdiFormatHeader3,
  mdiFormatHeader4,
  mdiFormatHeader5,
  mdiFormatHeader6,
  mdiFormatItalic,
  mdiFormatListBulleted,
  mdiFormatListNumbered,
  mdiFormatParagraph,
  mdiFormatQuoteClose,
  mdiFormatStrikethrough,
  mdiFormatUnderline,
  mdiImage,
  mdiLink,
  mdiLinkOff,
  mdiRedo,
  mdiTable,
  mdiTableColumnPlusBefore,
  mdiTableColumnRemove,
  mdiTablePlus,
  mdiTableRemove,
  mdiTableRowPlusAfter,
  mdiTableRowRemove,
  mdiUndo,
} from "@mdi/js";
import Color from "@tiptap/extension-color";
import { Highlight } from "@tiptap/extension-highlight";
import Image from "@tiptap/extension-image";
import Placeholder from "@tiptap/extension-placeholder";
import { Table } from "@tiptap/extension-table";
import TableCell from "@tiptap/extension-table-cell";
import { TableHeader } from "@tiptap/extension-table-header";
import TableRow from "@tiptap/extension-table-row";
import TextAlign from "@tiptap/extension-text-align";
import { TextStyle } from "@tiptap/extension-text-style";
import { Editor, EditorContent, useEditor, useEditorState } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import { useId, useRef, useState } from "react";
import ImageResize from "tiptap-extension-resize-image";
import Card from "../Composition/Card";
import Modal from "../Composition/Modal";
import Button from "../Form/Button";
import ColorPicker from "../Form/ColorPicker";
import Input from "../Form/Input";
import { ImageUploadField, useImageUpload } from "../Form/ImageUpload";

const extensions = [
  StarterKit.configure({ heading: { levels: [1, 2, 3, 4, 5, 6] } }),
  Placeholder.configure({ placeholder: "Start typing..." }),
  TextStyle,
  Color,
  Highlight.configure({ multicolor: true }),
  Image,
  ImageResize,
  Table.configure({ resizable: true }),
  TableRow,
  TableHeader,
  TableCell,
  TextAlign.configure({ types: ["heading", "paragraph"] }),
];

const headingLevels = [1, 2, 3, 4, 5, 6] as const;

const headingIcons = {
  1: mdiFormatHeader1,
  2: mdiFormatHeader2,
  3: mdiFormatHeader3,
  4: mdiFormatHeader4,
  5: mdiFormatHeader5,
  6: mdiFormatHeader6,
};

const alignments = [
  { value: "left", icon: mdiFormatAlignLeft },
  { value: "center", icon: mdiFormatAlignCenter },
  { value: "right", icon: mdiFormatAlignRight },
  { value: "justify", icon: mdiFormatAlignJustify },
] as const;

function MenuBar({ editor }: { editor: Editor | null }) {
  const [showModal, setShowModal] = useState<false | "link" | "image">(false);
  const [inputValue, setInputValue] = useState("");
  const imageInputId = useId();

  const imageUpload = useImageUpload({
    onSelect: (_file, objectUrl) => setInputValue(objectUrl),
    onClear: () => setInputValue(""),
  });

  // Recomputed per transaction, but only triggers a re-render when the selected snapshot changes.
  const editorState = useEditorState({
    editor,
    selector: ({ editor }) => {
      if (!editor) {
        return null;
      }

      const paraAlign = editor.getAttributes("paragraph")?.textAlign;
      const headingAlign = editor.getAttributes("heading")?.textAlign;

      return {
        isBold: editor.isActive("bold"),
        canBold: editor.can().chain().focus().toggleBold().run(),
        isItalic: editor.isActive("italic"),
        canItalic: editor.can().chain().focus().toggleItalic().run(),
        isUnderline: editor.isActive("underline"),
        canUnderline: editor.can().chain().focus().toggleUnderline().run(),
        isStrike: editor.isActive("strike"),
        canStrike: editor.can().chain().focus().toggleStrike().run(),
        isCode: editor.isActive("code"),
        canCode: editor.can().chain().focus().toggleCode().run(),
        isCodeBlock: editor.isActive("codeBlock"),
        canCodeBlock: editor.can().chain().focus().toggleCodeBlock().run(),
        headings: headingLevels.map(level => ({
          isActive: editor.isActive("heading", { level }),
          canToggle: editor.can().chain().focus().toggleHeading({ level }).run(),
        })),
        isParagraph: editor.isActive("paragraph"),
        textAlign: paraAlign ?? headingAlign ?? "left",
        canSetTextAlign: editor.can().chain().focus().setTextAlign("left").run(),
        isBulletList: editor.isActive("bulletList"),
        canBulletList: editor.can().chain().focus().toggleBulletList().run(),
        isOrderedList: editor.isActive("orderedList"),
        canOrderedList: editor.can().chain().focus().toggleOrderedList().run(),
        isBlockquote: editor.isActive("blockquote"),
        canBlockquote: editor.can().chain().focus().toggleBlockquote().run(),
        isLink: editor.isActive("link"),
        canSetImage: editor.can().chain().focus().setImage({ src: "https://" }).run(),
        textColor: editor.getAttributes("textStyle")?.color || "#000000",
        highlightColor: editor.getAttributes("highlight")?.color || "#FFFF00",
        canUndo: editor.can().undo(),
        canRedo: editor.can().redo(),
      };
    },
  });

  if (!editor || !editorState) return null;

  const handleModalConfirm = () => {
    if (showModal === "link") {
      editor.chain().focus().setLink({ href: inputValue }).run();
      imageUpload.clearImage();
    } else if (showModal === "image") {
      editor.chain().focus().setImage({ src: inputValue }).run();
      // The editor now owns the inserted object URL: release, don't revoke.
      imageUpload.releaseImage();
    }
    setShowModal(false);
  };

  return (
    <>
      {showModal && (
        <Modal
          title={showModal === "link" ? "Insert Link" : "Insert Image"}
          onConfirm={handleModalConfirm}
          onCancel={() => {
            imageUpload.clearImage();
            setShowModal(false);
          }}
        >
          {showModal === "link" && (
            <Input type="text" label="URL" value={inputValue} onChange={e => setInputValue(e.target.value)} autoFocus />
          )}

          {showModal === "image" && <ImageUploadField inputId={imageInputId} upload={imageUpload} />}
        </Modal>
      )}

      <div className="RichText-buttons">
        {/** Formatting */}
        <Button
          onClick={() => editor.chain().focus().toggleBold().run()}
          disabled={!editorState.canBold}
          className={editorState.isBold ? "" : "secondary"}
          icon={mdiFormatBold}
          title="Bold"
        />
        <Button
          onClick={() => editor.chain().focus().toggleItalic().run()}
          disabled={!editorState.canItalic}
          className={editorState.isItalic ? "" : "secondary"}
          icon={mdiFormatItalic}
          title="Italic"
        />
        <Button
          onClick={() => editor.chain().focus().toggleUnderline().run()}
          disabled={!editorState.canUnderline}
          className={editorState.isUnderline ? "" : "secondary"}
          icon={mdiFormatUnderline}
          title="Underline"
        />
        <Button
          onClick={() => editor.chain().focus().toggleStrike().run()}
          disabled={!editorState.canStrike}
          className={editorState.isStrike ? "" : "secondary"}
          icon={mdiFormatStrikethrough}
          title="Strikethrough"
        />
        <Button
          onClick={() => editor.chain().focus().toggleCode().run()}
          disabled={!editorState.canCode}
          className={editorState.isCode ? "" : "secondary"}
          icon={mdiCodeBraces}
          title="Code"
        />
        <Button
          onClick={() => editor.chain().focus().toggleCodeBlock().run()}
          disabled={!editorState.canCodeBlock}
          className={editorState.isCodeBlock ? "" : "secondary"}
          icon={mdiCodeTags}
          title="Code Block"
        />

        {headingLevels.map((level, i) => (
          <Button
            key={level}
            onClick={() => editor.chain().focus().toggleHeading({ level }).run()}
            disabled={!editorState.headings[i].canToggle}
            className={editorState.headings[i].isActive ? "" : "secondary"}
            icon={headingIcons[level]}
            title={`Heading ${level}`}
          />
        ))}

        <Button
          onClick={() => editor.chain().focus().setParagraph().run()}
          className={editorState.isParagraph ? "" : "secondary"}
          icon={mdiFormatParagraph}
          title="Paragraph"
        />

        {/** Alignment */}
        {alignments.map(({ value, icon }) => (
          <Button
            key={value}
            onClick={() => editor.chain().focus().setTextAlign(value).run()}
            disabled={!editorState.canSetTextAlign}
            className={editorState.textAlign === value ? "" : "secondary"}
            icon={icon}
            title={`Align ${value.charAt(0).toUpperCase() + value.slice(1)}`}
          />
        ))}

        {/** Lists */}
        <Button
          onClick={() => editor.chain().focus().toggleBulletList().run()}
          disabled={!editorState.canBulletList}
          className={editorState.isBulletList ? "" : "secondary"}
          icon={mdiFormatListBulleted}
          title="Bullet List"
        />
        <Button
          onClick={() => editor.chain().focus().toggleOrderedList().run()}
          disabled={!editorState.canOrderedList}
          className={editorState.isOrderedList ? "" : "secondary"}
          icon={mdiFormatListNumbered}
          title="Ordered List"
        />

        <Button
          onClick={() => editor.chain().focus().toggleBlockquote().run()}
          disabled={!editorState.canBlockquote}
          className={editorState.isBlockquote ? "" : "secondary"}
          icon={mdiFormatQuoteClose}
          title="Blockquote"
        />

        {!editorState.isLink ? (
          <Button
            icon={mdiLink}
            title="Add Link"
            onClick={() => {
              setInputValue("https://");
              setShowModal("link");
            }}
          />
        ) : (
          <Button icon={mdiLinkOff} title="Remove Link" onClick={() => editor.chain().focus().unsetLink().run()} />
        )}

        <Button
          disabled={!editorState.canSetImage}
          icon={mdiImage}
          title="Insert Image"
          onClick={() => setShowModal("image")}
        />

        {/** Table */}
        <Button
          icon={mdiTable}
          title="Insert Table"
          onClick={() => editor.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()}
        />
        <Button
          icon={mdiTablePlus}
          title="Add Column After"
          onClick={() => editor.chain().focus().addColumnAfter().run()}
        />
        <Button
          icon={mdiTableColumnPlusBefore}
          title="Add Column Before"
          onClick={() => editor.chain().focus().addColumnBefore().run()}
        />
        <Button
          icon={mdiTableColumnRemove}
          title="Delete Column"
          onClick={() => editor.chain().focus().deleteColumn().run()}
        />
        <Button
          icon={mdiTableRowPlusAfter}
          title="Add Row After"
          onClick={() => editor.chain().focus().addRowAfter().run()}
        />
        <Button icon={mdiTableRowRemove} title="Delete Row" onClick={() => editor.chain().focus().deleteRow().run()} />
        <Button icon={mdiTableRemove} title="Delete Table" onClick={() => editor.chain().focus().deleteTable().run()} />

        <ColorPicker
          icon={mdiFormatColorText}
          title="Text Color"
          value={editorState.textColor}
          onChange={color => editor.chain().focus().setColor(color).run()}
        />

        <ColorPicker
          icon={mdiFormatColorHighlight}
          title="Highlight"
          value={editorState.highlightColor}
          onChange={color => editor.chain().focus().setHighlight({ color }).run()}
        />

        <Button
          onClick={() => editor.chain().focus().undo().run()}
          disabled={!editorState.canUndo}
          icon={mdiUndo}
          title="Undo"
        />
        <Button
          onClick={() => editor.chain().focus().redo().run()}
          disabled={!editorState.canRedo}
          icon={mdiRedo}
          title="Redo"
        />
      </div>
    </>
  );
}

type RichTextEditorProps = {
  /** Initial HTML content; the editor owns the document after mount. */
  content?: string;
  /** Called with the editor HTML on every document update. */
  onChange?: (html: string) => void;
};

export default function RichTextEditor({ content = "", onChange }: RichTextEditorProps) {
  // Keep the latest callback available to onUpdate without recreating the editor.
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;

  const editor = useEditor({
    extensions,
    content,
    onUpdate: ({ editor }) => onChangeRef.current?.(editor.getHTML()),
  });

  return (
    <Card className="RichText">
      <MenuBar editor={editor} />
      <EditorContent className="RichText-editor-content" editor={editor} />
    </Card>
  );
}
