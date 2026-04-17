import { mdiPencil, mdiPlus, mdiTarget, mdiTrashCan } from "@mdi/js";
import { useEffect, useState } from "react";
import { useParams } from "react-router";
import { toast } from "react-toastify";
import { deleteData, getData, patchData } from "../api/api";
import Grid from "../components/Composition/Grid";
import Modal from "../components/Composition/Modal";
import PageHeader from "../components/Composition/PageHeader";
import Table from "../components/Composition/Table";
import Button from "../components/Form/Button";
import Buttons from "../components/Form/Buttons";
import Input from "../components/Form/Input";
import AddTargetModal from "../components/Modals/AddTargetModal";
import { Target } from "../types/common.types";
import { getPageTitle, sortBy } from "../utils/helpers";

export default function Targets() {
  const [targets, setTargets] = useState<Target[]>([]);
  const [loadingTargets, setLoadingTargets] = useState(true);

  const [isModalAddActive, setIsModalAddActive] = useState(false);

  const [isModalEditActive, setIsModalEditActive] = useState(false);
  const [editingTarget, setEditingTarget] = useState<Target | null>(null);

  const [isModalTrashActive, setIsModalTrashActive] = useState(false);
  const [targetToDelete, setTargetToDelete] = useState<Target | null>(null);

  const [ipv4, setIpv4] = useState("");
  const [ipv6, setIpv6] = useState("");
  const [fqdn, setFqdn] = useState("");
  const [tag, setTag] = useState("");

  const { customerId } = useParams<{ customerId: string }>();

  const fetchTargets = () => {
    setLoadingTargets(true);
    getData<Target[]>(`/api/customers/${customerId}/targets`, setTargets, undefined, () => setLoadingTargets(false));
  };

  useEffect(() => {
    document.title = getPageTitle("Targets");
    fetchTargets();
  }, [customerId]);

  const openAddModal = () => {
    setIsModalAddActive(true);
  };

  const openEditModal = (target: Target) => {
    setEditingTarget(target);
    setIpv4(target.ipv4 || "");
    setIpv6(target.ipv6 || "");
    setFqdn(target.fqdn || "");
    setTag(target.tag || "");
    setIsModalEditActive(true);
  };

  const handleEditConfirm = () => {
    const payload = {
      ipv4: ipv4.trim(),
      ipv6: ipv6.trim(),
      fqdn: fqdn.trim(),
      tag: tag.trim(),
    };

    patchData<Target>(`/api/targets/${editingTarget.id}`, payload, () => {
      toast.success("Target updated successfully");
      setIsModalEditActive(false);
      setEditingTarget(null);
      fetchTargets();
    });
  };

  const openDeleteModal = (target: Target) => {
    setTargetToDelete(target);
    setIsModalTrashActive(true);
  };

  const handleDeleteConfirm = () => {
    deleteData(`/api/targets/${targetToDelete.id}`, () => {
      toast.success(
        `Target "${targetToDelete.tag || targetToDelete.fqdn || targetToDelete.ipv4 || targetToDelete.ipv6}" deleted successfully`
      );
      setIsModalTrashActive(false);
      setTargetToDelete(null);
      fetchTargets();
    });
  };

  return (
    <div>
      {/* Add Target Modal */}
      {isModalAddActive && <AddTargetModal setShowModal={setIsModalAddActive} onTargetCreated={fetchTargets} />}

      {/* Edit Target Modal */}
      {isModalEditActive && (
        <Modal
          title="Edit Target"
          confirmButtonLabel="Save"
          onConfirm={handleEditConfirm}
          onCancel={() => setIsModalEditActive(false)}
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
      )}

      {/* Delete Confirmation Modal */}
      {isModalTrashActive && (
        <Modal
          title="Please confirm: action irreversible"
          confirmButtonLabel="Confirm"
          onConfirm={handleDeleteConfirm}
          onCancel={() => setIsModalTrashActive(false)}
        >
          <p>
            Are you sure to delete{" "}
            <strong>
              {targetToDelete?.tag || targetToDelete?.fqdn || targetToDelete?.ipv4 || targetToDelete?.ipv6 || ""}
            </strong>{" "}
            target?
          </p>
        </Modal>
      )}

      <PageHeader icon={mdiTarget} title="Targets">
        <Button icon={mdiPlus} text="New target" small onClick={openAddModal} />
      </PageHeader>

      <Table
        loading={loadingTargets}
        data={targets.sort(sortBy("fqdn", { caseInsensitive: true })).map(target => ({
          "FQDN | Target name": target.fqdn,
          IPv4: target.ipv4,
          IPv6: target.ipv6,
          Tag: target.tag,
          buttons: (
            <Buttons noWrap>
              <Button
                variant="tertiary"
                icon={mdiPencil}
                title="Edit target"
                onClick={() => openEditModal(target)}
                small
              />
              <Button
                variant="danger"
                icon={mdiTrashCan}
                title="Delete target"
                onClick={() => openDeleteModal(target)}
                small
              />
            </Buttons>
          ),
        }))}
        perPageCustom={10}
      />
    </div>
  );
}
