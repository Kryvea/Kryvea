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
import Placeholder from "@tiptap/extension-placeholder";
import { Table } from "@tiptap/extension-table";
import TableCell from "@tiptap/extension-table-cell";
import { TableHeader } from "@tiptap/extension-table-header";
import TableRow from "@tiptap/extension-table-row";
import TextAlign from "@tiptap/extension-text-align";
import { TextStyle } from "@tiptap/extension-text-style";
import { Editor, EditorContent, useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import { useEffect, useId, useRef, useState } from "react";
import ImageResize from "tiptap-extension-resize-image";
import Card from "../Composition/Card";
import Grid from "../Composition/Grid";
import Modal from "../Composition/Modal";
import Button from "../Form/Button";
import ColorPicker from "../Form/ColorPicker";
import Input from "../Form/Input";
import UploadFile from "../Form/UploadFile";

const extensions = [
  StarterKit.configure({ heading: { levels: [1, 2, 3, 4, 5, 6] } }),
  Placeholder.configure({ placeholder: "Start typing..." }),
  TextStyle,
  Color,
  Highlight.configure({ multicolor: true }),
  ImageResize,
  Table.configure({ resizable: true }),
  TableRow,
  TableHeader,
  TableCell,
  TextAlign.configure({ types: ["heading", "paragraph"] }),
];

const headingIcons = {
  1: mdiFormatHeader1,
  2: mdiFormatHeader2,
  3: mdiFormatHeader3,
  4: mdiFormatHeader4,
  5: mdiFormatHeader5,
  6: mdiFormatHeader6,
};

function MenuBar({ editor }: { editor: Editor | null }) {
  const [, setState] = useState(0);
  const [showModal, setShowModal] = useState<false | "link" | "image">(false);
  const [inputValue, setInputValue] = useState("");

  const imageInputRef = useRef<HTMLInputElement>(null);
  const [filename, setFilename] = useState<string>("");
  const [imageUrl, setImageUrl] = useState<string>("");
  const imageInputId = useId();

  useEffect(() => {
    if (!editor) return;

    const callback = () => setState(v => v + 1);
    editor.on("transaction", callback);

    return () => {
      editor.off("transaction", callback);
    };
  }, [editor]);

  if (!editor) return null;

  const is = (name: string, attrs?: any) => editor.isActive(name, attrs);
  const can = (command: () => boolean) => command();
  const isAlign = (value: "left" | "center" | "right" | "justify") => {
    const paraAlign = editor.getAttributes("paragraph")?.textAlign;
    const headingAlign = editor.getAttributes("heading")?.textAlign;
    return (paraAlign ?? headingAlign ?? "left") === value;
  };

  const handleModalConfirm = () => {
    if (showModal === "link") {
      editor.chain().focus().setLink({ href: inputValue }).run();
    } else if (showModal === "image") {
      editor.chain().focus().setImage({ src: inputValue }).run();
    }
    setShowModal(false);
    setInputValue("");
    clearImage();
  };

  const onImageChangeWrapper = ({ target: { files } }) => {
    if (!files || files.length === 0) return;

    const file = files[0];
    if (file.type !== "image/png" && file.type !== "image/jpeg") return;

    const objectUrl = URL.createObjectURL(file);
    setFilename(file.name);
    setImageUrl(objectUrl);
    setInputValue(objectUrl);
  };

  const clearImage = () => {
    imageInputRef.current.value = "";
    setFilename("");
    setImageUrl("");
    setInputValue("");
  };

  useEffect(() => {
    return () => {
      if (imageUrl) {
        URL.revokeObjectURL(imageUrl);
      }
    };
  }, []);

  return (
    <>
      {showModal && (
        <Modal
          title={showModal === "link" ? "Insert Link" : "Insert Image"}
          onConfirm={handleModalConfirm}
          onCancel={() => setShowModal(false)}
        >
          {showModal === "link" && (
            <Input type="text" label="URL" value={inputValue} onChange={e => setInputValue(e.target.value)} autoFocus />
          )}

          {showModal === "image" && (
            <Grid>
              <UploadFile
                label="Choose Image"
                inputId={imageInputId}
                filename={filename}
                inputRef={imageInputRef}
                name="imagePoc"
                accept="image/png, image/jpeg"
                onChange={onImageChangeWrapper}
                onButtonClick={clearImage}
              />
              {imageUrl && (
                <img src={imageUrl} alt="Selected image preview" className="max-h-[550px] w-fit object-contain" />
              )}
            </Grid>
          )}
        </Modal>
      )}

      <div className="RichText-buttons">
        {/** Formatting */}
        <Button
          onClick={() => editor.chain().focus().toggleBold().run()}
          disabled={!can(() => editor.can().chain().focus().toggleBold().run())}
          className={is("bold") ? "" : "secondary"}
          icon={mdiFormatBold}
          title="Bold"
        />
        <Button
          onClick={() => editor.chain().focus().toggleItalic().run()}
          disabled={!can(() => editor.can().chain().focus().toggleItalic().run())}
          className={is("italic") ? "" : "secondary"}
          icon={mdiFormatItalic}
          title="Italic"
        />
        <Button
          onClick={() => editor.chain().focus().toggleUnderline().run()}
          disabled={!can(() => editor.can().chain().focus().toggleUnderline().run())}
          className={is("underline") ? "" : "secondary"}
          icon={mdiFormatUnderline}
          title="Underline"
        />
        <Button
          onClick={() => editor.chain().focus().toggleStrike().run()}
          disabled={!can(() => editor.can().chain().focus().toggleStrike().run())}
          className={is("strike") ? "" : "secondary"}
          icon={mdiFormatStrikethrough}
          title="Strikethrough"
        />
        <Button
          onClick={() => editor.chain().focus().toggleCode().run()}
          disabled={!can(() => editor.can().chain().focus().toggleCode().run())}
          className={is("code") ? "" : "secondary"}
          icon={mdiCodeBraces}
          title="Code"
        />
        <Button
          onClick={() => editor.chain().focus().toggleCodeBlock().run()}
          disabled={!can(() => editor.can().chain().focus().toggleCodeBlock().run())}
          className={is("codeBlock") ? "" : "secondary"}
          icon={mdiCodeTags}
          title="Code Block"
        />

        {([1, 2, 3, 4, 5, 6] as const).map(level => (
          <Button
            key={level}
            onClick={() => editor.chain().focus().toggleHeading({ level }).run()}
            disabled={!can(() => editor.can().chain().focus().toggleHeading({ level }).run())}
            className={is("heading", { level }) ? "" : "secondary"}
            icon={headingIcons[level]}
            title={`Heading ${level}`}
          />
        ))}

        <Button
          onClick={() => editor.chain().focus().setParagraph().run()}
          className={is("paragraph") ? "" : "secondary"}
          icon={mdiFormatParagraph}
          title="Paragraph"
        />

        {/** Alignment */}
        {["left", "center", "right", "justify"].map(value => (
          <Button
            key={value}
            onClick={() => editor.chain().focus().setTextAlign(value).run()}
            disabled={!can(() => editor.can().chain().focus().setTextAlign(value).run())}
            className={isAlign(value as any) ? "" : "secondary"}
            icon={
              {
                left: mdiFormatAlignLeft,
                center: mdiFormatAlignCenter,
                right: mdiFormatAlignRight,
                justify: mdiFormatAlignJustify,
              }[value]
            }
            title={`Align ${value.charAt(0).toUpperCase() + value.slice(1)}`}
          />
        ))}

        {/** Lists */}
        <Button
          onClick={() => editor.chain().focus().toggleBulletList().run()}
          disabled={!can(() => editor.can().chain().focus().toggleBulletList().run())}
          className={is("bulletList") ? "" : "secondary"}
          icon={mdiFormatListBulleted}
          title="Bullet List"
        />
        <Button
          onClick={() => editor.chain().focus().toggleOrderedList().run()}
          disabled={!can(() => editor.can().chain().focus().toggleOrderedList().run())}
          className={is("orderedList") ? "" : "secondary"}
          icon={mdiFormatListNumbered}
          title="Ordered List"
        />

        <Button
          onClick={() => editor.chain().focus().toggleBlockquote().run()}
          disabled={!can(() => editor.can().chain().focus().toggleBlockquote().run())}
          className={is("blockquote") ? "" : "secondary"}
          icon={mdiFormatQuoteClose}
          title="Blockquote"
        />

        {!is("link") ? (
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
          disabled={!can(() => editor.can().chain().focus().setImage({ src: "https://" }).run())}
          icon={mdiImage}
          title="Insert Image"
          onClick={() => {
            setShowModal("image");
            imageInputRef.current?.click();
          }}
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
          value={editor.getAttributes("textStyle")?.color || "#000000"}
          onChange={color => editor.chain().focus().setColor(color).run()}
        />

        <ColorPicker
          icon={mdiFormatColorHighlight}
          title="Highlight"
          value={editor.getAttributes("highlight")?.color || "#FFFF00"}
          onChange={color => editor.chain().focus().setHighlight({ color }).run()}
        />

        <Button
          onClick={() => editor.chain().focus().undo().run()}
          disabled={!can(() => editor.can().undo())}
          icon={mdiUndo}
          title="Undo"
        />
        <Button
          onClick={() => editor.chain().focus().redo().run()}
          disabled={!can(() => editor.can().redo())}
          icon={mdiRedo}
          title="Redo"
        />
      </div>
    </>
  );
}

export default function RichTextEditor() {
  const editor = useEditor({ extensions });

  return (
    <Card className="RichText">
      <MenuBar editor={editor} />
      <EditorContent className="RichText-editor-content" editor={editor} />
    </Card>
  );
}
