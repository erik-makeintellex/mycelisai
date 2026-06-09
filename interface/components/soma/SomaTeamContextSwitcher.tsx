"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import { Check, ChevronDown, FolderKanban, MessageSquareText, Users } from "lucide-react";
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
  const [isOpen, setIsOpen] = useState(false);
  const pickerRef = useRef<HTMLDivElement | null>(null);
  const teamSummaries = useMemo(
    () => buildTeamContextSummaries(teams, workItems, focusedTeamId),
    [focusedTeamId, teams, workItems],
  );
  const selectedTeam = focusedTeamId
    ? teamSummaries.find((team) => team.id === focusedTeamId) ?? fallbackTeamSummary(focusedTeamId)
    : null;
  const selectedContextName = selectedTeam?.name ?? "Soma root";
  const selectedContextDescription = selectedTeam
    ? "Team chat, work, outputs, and proof"
    : "Cross-team continuity and all work";

  useEffect(() => {
    if (!isOpen) return;
    const handlePointerDown = (event: PointerEvent) => {
      if (!pickerRef.current?.contains(event.target as Node)) setIsOpen(false);
    };
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") setIsOpen(false);
    };
    document.addEventListener("pointerdown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [isOpen]);

  if (teamSummaries.length === 0 && !focusedTeamId) return null;

  const selectRoot = () => {
    onRootSelect();
    setIsOpen(false);
  };

  const selectTeam = (teamId: string) => {
    onTeamSelect(teamId);
    setIsOpen(false);
  };

  return (
    <div
      className="mb-2 rounded-xl border border-cortex-border bg-cortex-bg/80 px-2.5 py-2"
      data-testid="soma-team-context-switcher"
    >
      <div className="grid gap-2 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-center">
        <div ref={pickerRef} className="relative min-w-0">
        <button
          type="button"
          aria-haspopup="listbox"
          aria-expanded={isOpen}
          aria-controls="soma-work-context-listbox"
          onClick={() => setIsOpen((open) => !open)}
          className="flex w-full min-w-0 items-center justify-between gap-2 rounded-lg border border-cortex-border bg-cortex-surface px-2.5 py-2 text-left transition-colors hover:border-cortex-primary/40"
        >
          <span className="flex min-w-0 items-center gap-2">
            <span className="inline-flex shrink-0 items-center gap-1 rounded-md border border-cortex-primary/25 bg-cortex-primary/10 px-1.5 py-1 font-mono text-[9px] font-bold uppercase tracking-[0.12em] text-cortex-primary">
              <Users className="h-3 w-3" />
              Working in
            </span>
            {selectedTeam ? (
              <FolderKanban className="h-3.5 w-3.5 shrink-0 text-cortex-primary" />
            ) : (
              <MessageSquareText className="h-3.5 w-3.5 shrink-0 text-cortex-primary" />
            )}
            <span className="min-w-0">
              <span className="block truncate text-sm font-semibold text-cortex-text-main lg:inline">
                {selectedContextName}
              </span>
              <span className="block truncate text-[11px] text-cortex-text-muted lg:ml-2 lg:inline">
                {selectedContextDescription}
              </span>
            </span>
          </span>
          <span className="flex shrink-0 items-center gap-1.5">
            {selectedTeam ? <ContextStats team={selectedTeam} /> : <ContextBadge label="all work" />}
            <ChevronDown className={`h-4 w-4 text-cortex-text-muted transition-transform ${isOpen ? "rotate-180" : ""}`} />
          </span>
        </button>

        {isOpen ? (
          <div
            id="soma-work-context-listbox"
            role="listbox"
            aria-label="Choose current workflow"
            className="absolute left-0 right-0 top-full z-40 mt-2 overflow-hidden rounded-xl border border-cortex-border bg-cortex-surface shadow-2xl shadow-black/30"
          >
            <button
              type="button"
              role="option"
              aria-selected={!focusedTeamId}
              onClick={selectRoot}
              className="flex w-full items-center justify-between gap-3 border-b border-cortex-border/70 px-3 py-2 text-left hover:bg-cortex-primary/10"
            >
              <span className="flex min-w-0 items-center gap-2">
                <MessageSquareText className="h-3.5 w-3.5 shrink-0 text-cortex-primary" />
                <span className="min-w-0">
                  <span className="block truncate text-sm font-semibold text-cortex-text-main">Soma root</span>
                  <span className="block truncate text-[11px] text-cortex-text-muted">Cross-team continuity and all work.</span>
                </span>
              </span>
              <span className="flex shrink-0 items-center gap-2">
                <ContextBadge label="all work" />
                {!focusedTeamId ? <Check className="h-3.5 w-3.5 text-cortex-primary" /> : null}
              </span>
            </button>
            <div className="max-h-72 overflow-y-auto py-1" data-testid="soma-work-context-list">
              {teamSummaries.map((team) => {
                const selected = focusedTeamId === team.id;
                return (
                  <button
                    key={team.id}
                    type="button"
                    role="option"
                    aria-selected={selected}
                    onClick={() => selectTeam(team.id)}
                    className="flex w-full min-w-0 items-center justify-between gap-3 px-3 py-2 text-left hover:bg-cortex-primary/10"
                  >
                    <span className="flex min-w-0 items-center gap-2">
                      <FolderKanban className="h-3.5 w-3.5 shrink-0 text-cortex-primary" />
                      <span className="min-w-0">
                        <span className="block truncate text-sm font-semibold text-cortex-text-main">
                          {team.name}
                        </span>
                        <span className="block truncate text-[11px] text-cortex-text-muted">
                          Team workflow scope
                        </span>
                      </span>
                    </span>
                    <span className="flex shrink-0 items-center gap-2">
                      <ContextStats team={team} />
                      {selected ? <Check className="h-3.5 w-3.5 text-cortex-primary" /> : null}
                    </span>
                  </button>
                );
              })}
            </div>
          </div>
        ) : null}
        </div>
        <Link href="/teams" className="justify-self-start rounded-lg border border-cortex-border px-2.5 py-2 text-xs font-semibold text-cortex-primary hover:border-cortex-primary/40 hover:bg-cortex-primary/10 lg:justify-self-end">
          Manage teams
        </Link>
      </div>
    </div>
  );
}

function ContextStats({ team }: { team: ReturnType<typeof buildTeamContextSummaries>[number] }) {
  return (
    <span className="flex items-center gap-1 overflow-hidden">
      <ContextBadge label={`${team.workCount} work`} />
      {team.outputCount > 0 ? <ContextBadge label={`${team.outputCount} output${team.outputCount === 1 ? "" : "s"}`} /> : null}
      {team.needsAttention ? <ContextBadge label="needs review" tone="attention" /> : null}
    </span>
  );
}

function ContextBadge({ label, tone = "neutral" }: { label: string; tone?: "neutral" | "attention" }) {
  return (
    <span className={`shrink-0 rounded border px-1.5 py-0.5 font-mono text-[9px] uppercase tracking-[0.08em] ${
      tone === "attention"
        ? "border-amber-400/30 bg-amber-400/10 text-amber-300"
        : "border-cortex-border bg-cortex-bg/70 text-cortex-text-muted"
    }`}>
      {label}
    </span>
  );
}

function fallbackTeamSummary(teamId: string) {
  return {
    id: teamId,
    name: "Focused team",
    outputCount: 0,
    workCount: 0,
    needsAttention: false,
    priority: 50,
    visible: true,
  };
}

function buildTeamContextSummaries(teams: TeamDetailEntry[], workItems: TeamWorkItem[], focusedTeamId?: string | null) {
  return teams.map((team) => {
    const teamWork = workItems.filter((item) => item.teamIds.includes(team.id));
    const outputCount = teamWork.reduce((total, item) => total + (item.outputCount ?? item.outputRefs?.length ?? 0), 0);
    const workCount = teamWork.length;
    const needsAttention = teamWork.some((item) => item.needsOperator || ["degraded", "running", "reviewing", "queued"].includes(item.state));
    const highestPriority = teamWork.reduce(
      (current, item) => Math.min(current, contextItemPriority(item)),
      Number.POSITIVE_INFINITY,
    );
    return {
      id: team.id,
      name: team.name || team.id,
      outputCount,
      workCount,
      needsAttention,
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
