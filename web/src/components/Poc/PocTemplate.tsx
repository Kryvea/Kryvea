import { mdiDelete } from "@mdi/js";
import React, { ReactNode, useState } from "react";
import Flex from "../Composition/Flex";
import Icon from "../Composition/Icon";
import Button from "../Form/Button";
import Buttons from "../Form/Buttons";
import Input from "../Form/Input";
import { PocDoc } from "./Poc.types";

type PocTemplateProps = {
  icon: string;
  title: string;
  pocDoc: PocDoc;
  currentIndex: number;
  pocListLength: number;
  onPositionChange: (currentIndex: number) => (newIndex: number) => void;
  onRemovePoc: (currentIndex: number) => () => void;
  selectedPoc: number;
  setSelectedPoc: (index: number) => void;
  handleDrop?: (e: React.DragEvent<HTMLDivElement>) => void;
  children: ReactNode;
};

export default function PocTemplate({
  icon,
  title,
  pocDoc,
  currentIndex,
  pocListLength,
  onPositionChange,
  onRemovePoc,
  selectedPoc,
  setSelectedPoc,
  handleDrop = () => {},
  children,
}: PocTemplateProps) {
  const [isDraggedOver, setIsDraggedOver] = useState(false);

  const positionInputId = `poc-position-${currentIndex}-${pocDoc.key}`;

  return (
    <div
      className={isDraggedOver ? "poc-template dragged-over" : "poc-template"}
      data-type={pocDoc.type}
      data-focused={selectedPoc === currentIndex}
      onDragEnter={e => {
        e.preventDefault();
        setIsDraggedOver(true);
      }}
      onDragLeave={e => {
        e.preventDefault();
        setIsDraggedOver(false);
      }}
      onDragOver={e => {
        e.preventDefault();
        setIsDraggedOver(true);
      }}
      onDrop={e => {
        e.preventDefault();
        setIsDraggedOver(false);
        handleDrop(e);
      }}
      onClick={() => setSelectedPoc(currentIndex)}
      data-name="poc-template"
    >
      <div className="drop-image-over-hinter">
        <br />
        Drop image here
      </div>
      <Flex className="gap-4" items="center">
        <h1 className="flex items-center gap-2 rounded text-xl uppercase">
          <Icon path={icon} size={25} />
          {title}
        </h1>
        <Button variant="danger" small icon={mdiDelete} onClick={onRemovePoc(currentIndex)} />
      </Flex>
      <Flex col className="gap-2">
        <div className="flex gap-6">
          <Input
            label="Index"
            type="number"
            className="input h-8 w-[55px] rounded text-center"
            id={positionInputId}
            value={currentIndex + 1}
            min={1}
            max={pocListLength}
            onChange={newPosition => onPositionChange(currentIndex)(newPosition - 1)}
          />

          <Buttons>
            <Button
              variant="tertiary"
              text="Move Up"
              small
              className="h-8"
              disabled={currentIndex === 0}
              onClick={() => onPositionChange(currentIndex)(currentIndex <= 0 ? 0 : currentIndex - 1)}
            />
            <Button
              variant="tertiary"
              text="Move Down"
              small
              disabled={currentIndex === pocListLength - 1}
              onClick={() =>
                onPositionChange(currentIndex)(currentIndex >= pocListLength - 1 ? pocListLength - 1 : currentIndex + 1)
              }
            />
          </Buttons>
        </div>
        {children}
      </Flex>
    </div>
  );
}
