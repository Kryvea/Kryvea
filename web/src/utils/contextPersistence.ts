import { GlobalContextType } from "../App";
import { Keys } from "../types/utils.types";

export type GlobalContextKeys = Keys<GlobalContextType>;
type GlobalContextStatesTypes = GlobalContextType[GlobalContextKeys][0];

const REACT_GLOBAL_CONTEXT = "GlobalContextValues";

export function getLocalStorageCtxStates() {
  const globalContextValues = localStorage.getItem(REACT_GLOBAL_CONTEXT) ?? "{}";
  return JSON.parse(globalContextValues);
}

export function getLocalStorageCtxState(key: GlobalContextKeys) {
  const state = (JSON.parse(localStorage.getItem(REACT_GLOBAL_CONTEXT)) ?? {})[key];
  return state;
}

export function setLocalStorageCtxState(key: GlobalContextKeys, val) {
  const globalContextValues = getLocalStorageCtxStates();
  globalContextValues[key] = val;
  localStorage.setItem(REACT_GLOBAL_CONTEXT, JSON.stringify(globalContextValues));
}

export function clearLocalStorageCtxState(key: GlobalContextKeys) {
  setLocalStorageCtxState(key, undefined);
}
