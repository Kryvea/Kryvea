import { mdiClose, mdiPlus } from "@mdi/js";
import { Fragment, useState } from "react";
import { useParams } from "react-router";
import { toast } from "react-toastify";
import { patchData } from "../../api/api";
import { Target } from "../../types/common.types";
import Flex from "../Composition/Flex";
import Grid from "../Composition/Grid";
import Modal from "../Composition/Modal";
import Button from "../Form/Button";
import Input from "../Form/Input";
import Label from "../Form/Label";

interface UpsertTargetModalProps {
  setShowModal: (v: boolean) => void;
  onSaved?: (insertedIds?: string[]) => void;
  assessmentId?: string;
  existing?: Target[];
}

type PortRow = { id?: string; port: number; protocol: string };

export default function UpsertTargetModal({
  setShowModal,
  onSaved,
  assessmentId,
  existing,
}: UpsertTargetModalProps) {
  const { customerId } = useParams<{ customerId: string }>();
  const isEdit = !!existing && existing.length > 0;
  const seed = isEdit ? existing![0] : null;

  const [ipv4, setIpv4] = useState(seed?.ipv4 ?? "");
  const [ipv6, setIpv6] = useState(seed?.ipv6 ?? "");
  const [fqdn, setFqdn] = useState(seed?.fqdn ?? "");
  const [tag, setTag] = useState(seed?.tag ?? "");
  const [ports, setPorts] = useState<PortRow[]>(
    isEdit
      ? existing!
          .slice()
          .sort((a, b) => (a.port || 0) - (b.port || 0))
          .map(r => ({ id: r.id, port: r.port || 0, protocol: r.protocol }))
      : [{ port: 0, protocol: "" }]
  );
  const [removedPortIds, setRemovedPortIds] = useState<string[]>([]);

  const updatePort = (idx: number, patch: Partial<PortRow>) => {
    setPorts(prev => prev.map((p, i) => (i === idx ? { ...p, ...patch } : p)));
  };

  const removePort = (idx: number) => {
    setPorts(prev => {
      const target = prev[idx];
      if (target.id) setRemovedPortIds(ids => [...ids, target.id!]);
      return prev.filter((_, i) => i !== idx);
    });
  };

  const addPort = () => {
    setPorts(prev => [...prev, { port: 0, protocol: "" }]);
  };

  const handleConfirm = () => {
    const seen = new Set<string>();
    for (const p of ports) {
      const k = `${p.port}|${p.protocol}`;
      if (seen.has(k)) {
        toast.error("Duplicate port/protocol combination");
        return;
      }
      seen.add(k);
    }

    const payload = {
      customer_id: customerId,
      assessment_id: assessmentId,
      upsert: ports.map(p => ({
        id: p.id,
        ipv4: ipv4.trim(),
        ipv6: ipv6.trim(),
        fqdn: fqdn.trim(),
        tag: tag.trim(),
        port: p.port,
        protocol: p.protocol.trim(),
      })),
      delete: removedPortIds,
    };

    patchData<{ message: string; inserted_ids: string[] }>("/api/targets/bulk", payload, data => {
      toast.success(data.message);
      setShowModal(false);
      onSaved?.(data.inserted_ids);
    });
  };

  return (
    <Modal
      title={isEdit ? "Edit Target" : "New Target"}
      confirmButtonLabel="Save"
      onConfirm={handleConfirm}
      onCancel={() => setShowModal(false)}
    >
      <Grid className="grid-cols-1 gap-4">
        <Input
          type="text"
          id="ipv4"
          label="IPv4"
          placeholder="IPv4 address"
          value={ipv4}
          onChange={e => setIpv4(e.target.value)}
        />
        <Input
          type="text"
          id="ipv6"
          label="IPv6"
          placeholder="IPv6 address"
          value={ipv6}
          onChange={e => setIpv6(e.target.value)}
        />
        <Grid className="!items-center grid-cols-[1fr_1fr_auto]">
          <Label text="Port" />
          <Label text="Protocol" />
          <span />
          {ports.map((p, idx) => (
            <Fragment key={p.id ?? `new-${idx}`}>
              <Input
                type="number"
                id={`port-${idx}`}
                placeholder="host-level"
                min={0}
                max={65535}
                value={p.port || ""}
                onChange={(v: number) => updatePort(idx, { port: v || 0 })}
              />
              <Input
                type="text"
                id={`protocol-${idx}`}
                placeholder="tcp, udp, http, https"
                value={p.protocol}
                onChange={e => updatePort(idx, { protocol: e.target.value })}
              />
              <Button
                variant="danger"
                icon={mdiClose}
                small
                title="Remove port"
                onClick={() => removePort(idx)}
                disabled={ports.length === 1}
              />
            </Fragment>
          ))}
        </Grid>
        <Flex justify="start">
          <Button variant="tertiary" icon={mdiPlus} text="Add port" small onClick={addPort} />
        </Flex>
        <Input
          type="text"
          id="fqdn"
          label="FQDN | Target name"
          placeholder="Fully Qualified Domain Name or target name"
          value={fqdn}
          onChange={e => setFqdn(e.target.value)}
        />
        <Input
          type="text"
          id="tag"
          label="Tag"
          placeholder="This value is used to differentiate between duplicate entries"
          value={tag}
          onChange={e => setTag(e.target.value)}
        />
      </Grid>
    </Modal>
  );
}
