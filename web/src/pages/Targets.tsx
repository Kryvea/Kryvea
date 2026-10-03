import { mdiPencil, mdiPlus, mdiTarget, mdiTrashCan } from "@mdi/js";
import { useEffect, useMemo, useState } from "react";
import { useParams } from "react-router";
import { toast } from "react-toastify";
import { getData, patchData } from "../api/api";
import Grid from "../components/Composition/Grid";
import Modal from "../components/Composition/Modal";
import PageHeader from "../components/Composition/PageHeader";
import Table, { Column } from "../components/Composition/Table";
import Button from "../components/Form/Button";
import Buttons from "../components/Form/Buttons";
import SelectWrapper from "../components/Form/SelectWrapper";
import { SelectOption } from "../components/Form/SelectWrapper.types";
import UpsertTargetModal from "../components/Modals/UpsertTargetModal";
import { Target } from "../types/common.types";
import { getPageTitle } from "../utils/helpers";
import { groupTargets, TargetGroup } from "../utils/targetGroups";
import { getTargetLabel } from "../utils/targetLabel";
import { useTableUrlState } from "../utils/useTableUrlState";

const uniqueSorted = <T,>(values: T[], sort?: (a: T, b: T) => number) => {
  const out = [...new Set(values.filter(Boolean))];
  return sort ? out.sort(sort) : out;
};

const portOption = (row: Target): SelectOption => ({
  value: row.id,
  label: row.port
    ? `${row.port}${row.protocol ? ` / ${row.protocol}` : ""}`
    : `host-level${row.protocol ? ` / ${row.protocol}` : ""}`,
});

export default function Targets() {
  const [targets, setTargets] = useState<Target[]>([]);
  const [loadingTargets, setLoadingTargets] = useState(true);

  const t = useTableUrlState({ defaultLimit: 10, defaultSort: { key: "FQDN | Target name", order: "asc" } });

  const [isModalAddActive, setIsModalAddActive] = useState(false);

  const [editingGroup, setEditingGroup] = useState<TargetGroup | null>(null);

  const [isModalTrashActive, setIsModalTrashActive] = useState(false);
  const [groupToDelete, setGroupToDelete] = useState<TargetGroup | null>(null);
  const [deleteSelection, setDeleteSelection] = useState<Set<string>>(new Set());

  const { customerId } = useParams<{ customerId: string }>();

  const fetchTargets = () => {
    setLoadingTargets(true);
    getData<Target[]>(`/api/customers/${customerId}/targets`, setTargets, undefined, () => setLoadingTargets(false));
  };

  useEffect(() => {
    document.title = getPageTitle("Targets");
    fetchTargets();
  }, [customerId]);

  const groups: TargetGroup[] = useMemo(() => groupTargets(targets), [targets]);

  const openAddModal = () => {
    setIsModalAddActive(true);
  };

  const openEditModal = (group: TargetGroup) => {
    setEditingGroup(group);
  };

  const openDeleteModal = (group: TargetGroup) => {
    setGroupToDelete(group);
    setDeleteSelection(new Set(group.rows.map(r => r.id)));
    setIsModalTrashActive(true);
  };

  const handleDeleteConfirm = () => {
    if (!groupToDelete || deleteSelection.size === 0) return;

    const payload = {
      customer_id: customerId,
      upsert: [],
      delete: [...deleteSelection],
    };

    const hostLabel = getTargetLabel({ ...groupToDelete.rows[0], port: 0, protocol: "" });

    patchData<{ message: string }>("/api/targets/bulk", payload, () => {
      toast.success(`Target "${hostLabel}" deleted`);
      setIsModalTrashActive(false);
      setGroupToDelete(null);
      setDeleteSelection(new Set());
      fetchTargets();
    });
  };

  const targetColumns: Column<TargetGroup>[] = [
    { header: "FQDN | Target name", render: g => g.rows[0].fqdn },
    { header: "IPv4", render: g => g.rows[0].ipv4 },
    { header: "IPv6", render: g => g.rows[0].ipv6 },
    { header: "Ports", render: g => uniqueSorted(g.rows.map(r => r.port), (a, b) => a - b).join(", ") },
    { header: "Protocols", render: g => uniqueSorted(g.rows.map(r => r.protocol)).join(", ") },
    { header: "Tag", render: g => g.rows[0].tag },
    {
      kind: "actions",
      render: g => (
        <Buttons noWrap>
          <Button variant="tertiary" icon={mdiPencil} title="Edit target" onClick={() => openEditModal(g)} small />
          <Button variant="danger" icon={mdiTrashCan} title="Delete target" onClick={() => openDeleteModal(g)} small />
        </Buttons>
      ),
    },
  ];

  return (
    <div>
      {isModalAddActive && (
        <UpsertTargetModal setShowModal={setIsModalAddActive} onSaved={fetchTargets} />
      )}

      {editingGroup && (
        <UpsertTargetModal
          setShowModal={open => !open && setEditingGroup(null)}
          existing={editingGroup.rows}
          onSaved={fetchTargets}
        />
      )}

      {isModalTrashActive && groupToDelete && (
        <Modal
          title="Please confirm: action irreversible"
          confirmButtonLabel="Confirm"
          onConfirm={handleDeleteConfirm}
          onCancel={() => setIsModalTrashActive(false)}
        >
          {groupToDelete.rows.length === 1 ? (
            <p>
              Are you sure to delete <strong>{getTargetLabel(groupToDelete.rows[0])}</strong> target?
            </p>
          ) : (
            <Grid className="gap-2">
              <p>Select which entries to delete:</p>
              <SelectWrapper
                isMulti
                id="delete-selection"
                options={groupToDelete.rows
                  .slice()
                  .sort((a, b) => (a.port || 0) - (b.port || 0))
                  .map(portOption)}
                value={groupToDelete.rows.filter(row => deleteSelection.has(row.id)).map(portOption)}
                onChange={(opts: SelectOption[] | null) =>
                  setDeleteSelection(new Set((opts ?? []).map(o => o.value as string)))
                }
                closeMenuOnSelect={false}
              />
            </Grid>
          )}
        </Modal>
      )}

      <PageHeader icon={mdiTarget} title="Targets">
        <Button icon={mdiPlus} text="New target" small onClick={openAddModal} />
      </PageHeader>

      <Table
        tableId="targets"
        loading={loadingTargets}
        columns={targetColumns}
        data={groups}
        search={t.search}
        sort={t.sort}
        onSortChange={t.onSortChange}
        pagination={t.pagination}
      />
    </div>
  );
}
