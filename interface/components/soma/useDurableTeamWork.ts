"use client";

import { useEffect, useMemo, useState } from "react";
import {
  mapDurableTeamWorkItem,
  parseTeamWorkAPIItems,
  projectTeamWorkItem,
  teamOutputRefsFromItems,
} from "@/components/teams/teamWorkProjection";
import type { TeamDetailEntry, TeamOutputRef, TeamWorkItem } from "@/store/useCortexStore";

type DurableTeamWorkStatus = "idle" | "loading" | "durable" | "empty" | "degraded";

type DurableTeamWorkState = {
  status: DurableTeamWorkStatus;
  items: TeamWorkItem[];
  outputRefs: TeamOutputRef[];
  statusLabel?: string;
  degradedMessage?: string | null;
  emptyMessage: string;
};

export function useDurableTeamWork({
  teams,
  focusedTeamId,
  refreshVersion = 0,
  maxTeams = 8,
}: {
  teams: TeamDetailEntry[];
  focusedTeamId?: string | null;
  refreshVersion?: number;
  maxTeams?: number;
}): DurableTeamWorkState {
  const [isLoading, setIsLoading] = useState(false);
  const [durableItems, setDurableItems] = useState<TeamWorkItem[]>([]);
  const [failedTeamIds, setFailedTeamIds] = useState<string[]>([]);

  const selectedTeams = useMemo(() => {
    if (focusedTeamId) {
      return teams.filter((team) => team.id === focusedTeamId).slice(0, 1);
    }
    return teams.slice(0, maxTeams);
  }, [focusedTeamId, maxTeams, teams]);
  const selectedTeamKey = selectedTeams.map((team) => team.id).join("|");

  useEffect(() => {
    let cancelled = false;

    const load = async () => {
      if (selectedTeams.length === 0) {
        setDurableItems([]);
        setFailedTeamIds([]);
        setIsLoading(false);
        return;
      }

      setIsLoading(true);
      const fetchedItems: TeamWorkItem[] = [];
      const failures: string[] = [];

      await Promise.all(selectedTeams.map(async (team) => {
        try {
          const response = await fetch(`/api/v1/teams/${encodeURIComponent(team.id)}/work?limit=8`, {
            cache: "no-store",
          });
          if (!response.ok) {
            failures.push(team.id);
            return;
          }
          const payload = await response.json();
          const items = parseTeamWorkAPIItems(payload)
            .map((item) => mapDurableTeamWorkItem(item, team))
            .filter((item): item is TeamWorkItem => Boolean(item));
          fetchedItems.push(...items);
        } catch {
          failures.push(team.id);
        }
      }));

      if (cancelled) return;
      setDurableItems(sortWorkItems(fetchedItems));
      setFailedTeamIds(failures);
      setIsLoading(false);
    };

    void load();
    return () => {
      cancelled = true;
    };
  }, [refreshVersion, selectedTeamKey]);

  const fallbackItems = useMemo(() => (
    failedTeamIds.length > 0 && durableItems.length === 0
      ? selectedTeams
        .filter((team) => failedTeamIds.includes(team.id))
        .map(projectTeamWorkItem)
      : []
  ), [durableItems.length, failedTeamIds, selectedTeams]);

  const items = durableItems.length > 0 ? durableItems : fallbackItems;
  const outputRefs = teamOutputRefsFromItems(durableItems);

  if (isLoading) {
    return {
      status: "loading",
      items,
      outputRefs,
      statusLabel: "Checking durable team work.",
      emptyMessage: "Checking for active work items and retained team outputs.",
    };
  }

  if (durableItems.length > 0) {
    return {
      status: "durable",
      items: durableItems,
      outputRefs,
      statusLabel: "Durable team-work state loaded.",
      emptyMessage: "No active team work is attached to this view yet.",
      degradedMessage: failedTeamIds.length > 0
        ? "Some teams could not be checked; visible rows are durable records."
        : null,
    };
  }

  if (fallbackItems.length > 0) {
    return {
      status: "degraded",
      items: fallbackItems,
      outputRefs: [],
      statusLabel: "Degraded projection fallback.",
      degradedMessage: "The durable TeamWorkItem API was unavailable, so Soma is showing inspectable roster context only.",
      emptyMessage: "No durable team work is attached to this view yet.",
    };
  }

  return {
    status: selectedTeams.length > 0 ? "empty" : "idle",
    items: [],
    outputRefs: [],
    statusLabel: selectedTeams.length > 0 ? "No durable work items found." : undefined,
    emptyMessage: selectedTeams.length > 0
      ? "No active team work is attached to this view yet. Ask Soma to start work when you want a team to produce an output."
      : "No active team work is attached to this view yet.",
  };
}

function sortWorkItems(items: TeamWorkItem[]) {
  return [...items].sort((left, right) => {
    const leftTime = left.updatedAt ? Date.parse(left.updatedAt) : 0;
    const rightTime = right.updatedAt ? Date.parse(right.updatedAt) : 0;
    return rightTime - leftTime;
  });
}
