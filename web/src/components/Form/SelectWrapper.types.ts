import { ObjectKey } from "../../types/common.types";

export type SelectOption = {
  label: string;
  value: any;
  [k: ObjectKey]: any;
};
