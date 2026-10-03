import { useState } from "react";
import { useParams } from "react-router";
import { toast } from "react-toastify";
import { postData } from "../../api/api";
import Grid from "../Composition/Grid";
import Modal from "../Composition/Modal";
import Input from "../Form/Input";

interface AddTargetModalProps {
  setShowModal;
  assessmentId?: string;
  onTargetCreated?;
}

export default function AddTargetModal({ setShowModal, assessmentId, onTargetCreated }: AddTargetModalProps) {
  const { customerId } = useParams<{ customerId: string }>();

  const [ipv4, setIpv4] = useState("");
  const [ipv6, setIpv6] = useState("");
  const [fqdn, setFqdn] = useState("");
  const [tag, setTag] = useState("");

  const handleModalConfirm = () => {
    const payload = {
      ipv4: ipv4.trim(),
      ipv6: ipv6.trim(),
      fqdn: fqdn.trim(),
      tag: tag.trim(),
      customer_id: customerId,
      assessment_id: assessmentId,
    };

    postData<{ message: string; target_id: string }>("/api/targets", payload, data => {
      toast.success(data.message);
      setShowModal(false);
      setIpv4("");
      setIpv6("");
      setFqdn("");
      setTag("");

      if (onTargetCreated) {
        onTargetCreated(data.target_id);
      }
    });
  };

  return (
    <Modal
      title="New Target"
      confirmButtonLabel="Save"
      onConfirm={handleModalConfirm}
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
