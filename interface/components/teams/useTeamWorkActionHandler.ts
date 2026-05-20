"use client";

import { useCallback, useState } from "react";
import type { TeamInteraction, TeamWorkItem } from "@/store/useCortexStore";
import { postTeamWorkAction, postTeamWorkAsk } from "./teamWorkActions";

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
      setActiveWorkActionError(null);
      try {
        await postTeamWorkAction(item, action.action, actionSummary(item, action));
        setActiveWorkRefreshVersion((version) => version + 1);
      } catch (error) {
        setActiveWorkActionError(
          error instanceof Error ? error.message : "Team work action failed.",
        );
      }
    },
    [selectTeam],
  );

  const handleTeamAsk = useCallback(async (item: TeamWorkItem, message: string) => {
    setActiveWorkActionError(null);
    try {
      await postTeamWorkAsk(item, message);
      setActiveWorkRefreshVersion((version) => version + 1);
    } catch (error) {
      const message = error instanceof Error ? error.message : "Team ask failed.";
      setActiveWorkActionError(message);
      throw new Error(message);
    }
  }, []);

  return {
    activeWorkRefreshVersion,
    activeWorkActionError,
    handleActiveWorkAction,
    handleTeamAsk,
  };
}

function actionSummary(item: TeamWorkItem, action: TeamInteraction) {
  if (action.action === "steer") {
    return `Operator requested steering for "${item.title}". Continue from the current objective and ask Soma for clarified guidance if needed.`;
  }
  if (action.action === "recover") {
    return `Operator requested recovery for "${item.title}" using retained context, outputs, proof, and audit refs.`;
  }
  return undefined;
}
