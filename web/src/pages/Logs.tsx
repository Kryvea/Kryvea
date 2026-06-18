import { mdiMathLog } from "@mdi/js";
import { useEffect, useState } from "react";
import { getData } from "../api/api";
import Card from "../components/Composition/Card";
import Flex from "../components/Composition/Flex";
import Grid from "../components/Composition/Grid";
import PageHeader from "../components/Composition/PageHeader";
import Table, { Column } from "../components/Composition/Table";
import Checkbox from "../components/Form/Checkbox";
import { getPageTitle } from "../utils/helpers";
import { useTableUrlState } from "../utils/useTableUrlState";

type Log = {
  id: string;
  level: string;
  source: string;
  time: string;
  message: string;
  [key: string]: any;
};

const ALL_LEVELS = ["info", "debug", "error"];

export default function Logs() {
  const [logs, setLogs] = useState<Log[]>([]);
  const [loadingLogs, setLoadingLogs] = useState(false);
  const [selectedLevels, setSelectedLevels] = useState<string[]>(["error"]);

  const t = useTableUrlState({ defaultLimit: 50, defaultSort: { key: "Timestamp", order: "desc" } });

  function fetchLogs() {
    if (selectedLevels.length === 0) {
      setLogs([]);
      return;
    }

    setLoadingLogs(true);
    getData<{ logs: Log[] }>(
      `/api/admin/logs?levels=${selectedLevels.join(",")}`,
      data => {
        setLogs(data.logs.map((log, i) => ({ ...log, id: `${log.time}-${i}` })));
      },
      undefined,
      () => setLoadingLogs(false)
    );
  }

  useEffect(() => {
    document.title = getPageTitle("Logs");
    fetchLogs();
  }, [selectedLevels]);

  function toggleLevel(level: string) {
    setSelectedLevels(prev => (prev.includes(level) ? prev.filter(l => l !== level) : [...prev, level]));
  }

  const logColumns: Column<Log>[] = [
    {
      header: "Timestamp",
      sortValue: log => new Date(log.time),
      render: log => new Date(log.time).toLocaleString(),
    },
    { header: "Level", render: log => log.level },
    { header: "IP", render: log => log.ip },
    { header: "Method", render: log => log.method },
    { header: "URL", render: log => log.url },
    { header: "Status", render: log => log.status },
    { header: "Message", maxWidth: "30rem", render: log => log.message },
    { header: "Source", render: log => log.source },
  ];

  return (
    <div>
      <PageHeader icon={mdiMathLog} title="Logs" />
      <Grid>
        <Card>
          <Flex className="gap-2">
            <h1 className="font-bold">Include log level:</h1>
            {ALL_LEVELS.map(level => (
              <Checkbox
                id={`logs-levels-${level}`}
                key={level}
                label={level}
                checked={selectedLevels.includes(level)}
                onChange={() => toggleLevel(level)}
              />
            ))}
          </Flex>
        </Card>

        <Table
          tableId="logs"
          loading={loadingLogs}
          columns={logColumns}
          data={logs ?? []}
          search={t.search}
          sort={t.sort}
          onSortChange={t.onSortChange}
          pagination={t.pagination}
        />
      </Grid>
    </div>
  );
}
