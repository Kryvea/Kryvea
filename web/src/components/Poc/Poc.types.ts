import { MonacoTextSelection } from "./MonacoCodeEditor.types";
import { POC_TYPE_IMAGE, POC_TYPE_REQUEST_RESPONSE, POC_TYPE_RICH_TEXT, POC_TYPE_TEXT } from "./Poc.consts";

export type PocType =
  | typeof POC_TYPE_TEXT
  | typeof POC_TYPE_IMAGE
  | typeof POC_TYPE_REQUEST_RESPONSE
  | typeof POC_TYPE_RICH_TEXT;

type PocBaseDoc = {
  key: string;
  description: string;
  index: number;
};

export interface PocTextDoc extends PocBaseDoc {
  type: typeof POC_TYPE_TEXT;
  uri: string;
  text_language: string;
  text_data: string;
  text_highlights?: MonacoTextSelection[];
  starting_line_number?: number;
}

export interface PocImageDoc extends PocBaseDoc {
  type: typeof POC_TYPE_IMAGE;
  image_reference: string;
  image_caption: string;
  image_id?: string;
  image_url?: string;
  image_filename?: string;
  /** consumed by FormData */
  image_file?: File;
}

export interface PocRequestResponseDoc extends PocBaseDoc {
  type: typeof POC_TYPE_REQUEST_RESPONSE;
  uri: string;
  request: string;
  request_highlights?: MonacoTextSelection[];
  response: string;
  response_highlights?: MonacoTextSelection[];
}

export interface PocRichTextDoc extends PocBaseDoc {
  type: typeof POC_TYPE_RICH_TEXT;
  rich_text_data: string;
}

export type PocDoc = PocTextDoc | PocImageDoc | PocRequestResponseDoc | PocRichTextDoc;
