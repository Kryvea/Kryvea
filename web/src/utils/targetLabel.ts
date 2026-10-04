export function getTargetLabel(target): string {
  const fqdn = target.fqdn?.trim() || "";
  const ipParts = [target.ipv4, target.ipv6].filter(Boolean) as string[];
  const ip = ipParts.join(" - ");
  const tagValue = target.tag?.trim();
  const tag = tagValue ? ` (${tagValue})` : "";
  const portPart = target.port ? `:${target.port}` : "";
  const protocolPart = target.protocol ? `/${target.protocol}` : "";
  const suffix = portPart + protocolPart;

  let label = "";

  if (fqdn && ip) {
    label = `${fqdn} - ${ip}${suffix}`;
  } else if (fqdn) {
    label = `${fqdn}${suffix}`;
  } else if (ip) {
    label = `${ip}${suffix}`;
  }

  return label + tag;
}
