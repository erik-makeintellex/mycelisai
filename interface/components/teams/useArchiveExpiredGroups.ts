import { useState } from "react";
import {
  errorMessage,
  type GroupLifecycleReport,
} from "./groupWorkspaceTypes";

export function useArchiveExpiredGroups({
  loadGroups,
  setLifecycleReport,
  setNotice,
  setError,
}: {
  loadGroups: () => Promise<void>;
  setLifecycleReport: (report: GroupLifecycleReport) => void;
  setNotice: (message: string | null) => void;
  setError: (message: string | null) => void;
}) {
  const [archivingExpired, setArchivingExpired] = useState(false);

  const archiveExpiredGroups = async () => {
    setArchivingExpired(true);
    setNotice(null);
    setError(null);
    try {
      const res = await fetch("/api/v1/groups/lifecycle/archive-expired", {
        method: "POST",
      });
      const payload = (await res.json()).data as {
        archived_count?: number;
        report?: GroupLifecycleReport;
      };
      if (!res.ok) throw new Error("Could not archive expired groups.");
      const archivedCount = payload.archived_count ?? 0;
      if (payload.report) setLifecycleReport(payload.report);
      setNotice(
        archivedCount > 0
          ? `${archivedCount} expired temporary group${archivedCount === 1 ? "" : "s"} archived. Retained outputs remain reviewable.`
          : "No expired temporary groups needed cleanup.",
      );
      await loadGroups();
    } catch (archiveError) {
      setError(errorMessage(archiveError, "Could not archive expired groups."));
    } finally {
      setArchivingExpired(false);
    }
  };

  return { archivingExpired, archiveExpiredGroups };
}
