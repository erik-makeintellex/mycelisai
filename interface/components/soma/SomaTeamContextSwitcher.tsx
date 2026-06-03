"use client";

import Link from "next/link";
import { Users } from "lucide-react";
import type { TeamDetailEntry, TeamWorkItem } from "@/store/useCortexStore";

export function SomaTeamContextSwitcher({
  teams,
  workItems,
  focusedTeamId,
  onRootSelect,
  onTeamSelect,
}: {
  teams: TeamDetailEntry[];
  workItems: TeamWorkItem[];
  focusedTeamId?: string | null;
  onRootSelect: () => void;
  onTeamSelect: (teamId: string) => void;
}) {
  const teamSummaries = buildTeamContextSummaries(teams, workItems, focusedTeamId).slice(0, 6);
  if (teamSummaries.length === 0) return null;

  return (
    <div
      className="mb-3 rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2"
      data-testid="soma-team-context-switcher"
    >
      <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
        <p className="inline-flex items-center gap-2 font-mono text-[10px] font-bold uppercase tracking-[0.16em] text-cortex-text-muted">
          <Users className="h-3.5 w-3.5 text-cortex-primary" />
          Work contexts
        </p>
        <Link href="/teams" className="text-xs font-semibold text-cortex-primary hover:underline">
          Manage teams
        </Link>
      </div>
      <div className="flex gap-2 overflow-x-auto pb-1">
        <button
          type="button"
          onClick={onRootSelect}
          aria-pressed={!focusedTeamId}
          className={`shrink-0 rounded-lg border px-3 py-2 text-left transition-colors ${
            !focusedTeamId
              ? "border-cortex-primary/40 bg-cortex-primary/15"
              : "border-cortex-border bg-cortex-surface hover:border-cortex-primary/30"
          }`}
        >
          <span className="block font-mono text-[10px] uppercase tracking-[0.14em] text-cortex-primary">
            Soma root
          </span>
          <span className="mt-1 block text-xs text-cortex-text-muted">
            General chat
          </span>
        </button>
        {teamSummaries.map((team) => {
          const selected = focusedTeamId === team.id;
          return (
            <button
              key={team.id}
              type="button"
              onClick={() => onTeamSelect(team.id)}
              aria-pressed={selected}
              className={`min-w-[180px] shrink-0 rounded-lg border px-3 py-2 text-left transition-colors ${
                selected
                  ? "border-cortex-primary/40 bg-cortex-primary/15"
                  : "border-cortex-border bg-cortex-surface hover:border-cortex-primary/30"
              }`}
            >
              <span className="block truncate text-sm font-semibold text-cortex-text-main">
                {team.name}
              </span>
              <span className="mt-1 block font-mono text-[10px] uppercase tracking-[0.12em] text-cortex-text-muted">
                {team.stateLabel}
                {team.outputCount > 0 ? ` | ${team.outputCount} output${team.outputCount === 1 ? "" : "s"}` : ""}
              </span>
            </button>
          );
        })}
      </div>
    </div>
  );
}

function buildTeamContextSummaries(teams: TeamDetailEntry[], workItems: TeamWorkItem[], focusedTeamId?: string | null) {
  return teams.map((team) => {
    const teamWork = workItems.filter((item) => item.teamIds.includes(team.id));
    const outputCount = teamWork.reduce((total, item) => total + (item.outputCount ?? item.outputRefs?.length ?? 0), 0);
    const highestPriority = teamWork.reduce(
      (current, item) => Math.min(current, contextItemPriority(item)),
      Number.POSITIVE_INFINITY,
    );
    const stateLabel = teamWork.length === 0
      ? `${team.type === "standing" ? "Standing" : "Mission"} team`
      : `${teamWork.length} work item${teamWork.length === 1 ? "" : "s"}`;
    return {
      id: team.id,
      name: team.name || team.id,
      outputCount,
      stateLabel,
      priority: Number.isFinite(highestPriority) ? highestPriority : 50,
      visible: focusedTeamId === team.id || teamWork.length > 0 || outputCount > 0,
    };
  }).filter((team) => team.visible).sort((left, right) => {
    if (left.priority !== right.priority) return left.priority - right.priority;
    if (right.outputCount !== left.outputCount) return right.outputCount - left.outputCount;
    return left.name.localeCompare(right.name);
  });
}

function contextItemPriority(item: TeamWorkItem) {
  if (item.needsOperator) return -1;
  if (item.state === "degraded") return 1;
  if (item.state === "running") return 2;
  if (item.state === "reviewing") return 3;
  if (item.state === "queued") return 4;
  if (item.state === "output_ready") return 5;
  return 10;
}
