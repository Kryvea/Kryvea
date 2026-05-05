import { Target } from "../types/common.types";

export type TargetGroup = {
  id: string;
  rows: Target[];
};

export const groupTargets = (targets: Target[]): TargetGroup[] => {
  const map = new Map<string, Target[]>();
  for (const tgt of targets) {
    const k = `${tgt.ipv4}|${tgt.ipv6}|${tgt.fqdn}|${tgt.tag}`;
    if (!map.has(k)) map.set(k, []);
    map.get(k)!.push(tgt);
  }
  return [...map.entries()].map(([id, rows]) => ({ id, rows }));
};
