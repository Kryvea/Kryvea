type LineAndCol = { line: number; col: number };
export type MonacoTextSelection = {
  start: LineAndCol;
  end: LineAndCol;
  selectionPreview: string;
  color?: string;
};
