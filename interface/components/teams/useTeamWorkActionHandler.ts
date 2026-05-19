"use client";

import { useCallback, useState } from "react";
import type { TeamInteraction, TeamWorkItem } from "@/store/useCortexStore";
import { postTeamWorkAction } from "./teamWorkActions";

export function useTeamWorkActionHandler(
  selectTeam: (teamId: string | null) => void,
) {
  const [activeWorkRefreshVersion, setActiveWorkRefreshVersion] = useState(0);
  const [activeWorkActionError, setActiveWorkActionError] = useState<string | null>(null);

  const handleActiveWorkAction = useCallback(
    async (item: TeamWorkItem, action: TeamInteraction) => {
      if (action.action === "inspect") {
        selectTeam(item.teamIds[0] ?? item.id);
        return;
      }
      if (action.action === "steer") {
        return;
      }
      setActiveWorkActionError(null);
      try {
        await postTeamWorkAction(item, action.action);
        setActiveWorkRefreshVersion((version) => version + 1);
      } catch (error) {
        setActiveWorkActionError(
          error instanceof Error ? error.message : "Team work action failed.",
        );
      }
    },
    [selectTeam],
  );

  return {
    activeWorkRefreshVersion,
    activeWorkActionError,
    handleActiveWorkAction,
  };
}
