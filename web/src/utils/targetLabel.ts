export function getTargetLabel(target): string {
  const fqdn = target.fqdn?.trim() || "";
  const ipParts = [target.ipv4, target.ipv6].filter(Boolean) as string[];
  const ip = ipParts.join(" - ");
  const tagValue = target.tag?.trim();
  const tag = tagValue ? ` (${tagValue})` : "";

  let label = "";

  if (fqdn && ip) {
    label = `${fqdn} - ${ip}`;
  } else if (fqdn) {
    label = fqdn;
  } else if (ip) {
    label = ip;
  }

  return label + tag;
}
