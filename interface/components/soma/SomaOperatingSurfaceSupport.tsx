"use client";

import { Activity, CheckSquare, ListChecks, Wrench } from "lucide-react";
import { recoveryReviewQueueItems } from "@/components/recovery/recoveryQueue";
import { targetRefHref, targetRefReference } from "@/components/teams/teamWorkProjection";
import type { TeamDetailEntry, TeamOutputRef, TeamWorkItem } from "@/store/useCortexStore";
import type { OutcomeProjectSummary } from "./OutcomeProjectSummaryCard";
import type { DashboardRailAlert, DashboardRailAlertTarget } from "./SomaOutcomeVaultPanel";
import type { SomaEvidenceItem } from "./SomaEvidencePanel";

export const defaultSomaEvidence: SomaEvidenceItem[] = [
  {
    title: "Outcome approvals",
    detail: "Decide what Soma can run next.",
    href: "/approvals",
    icon: <CheckSquare className="h-4 w-4" />,
  },
  {
    title: "Outcome history",
    detail: "Review what happened and what changed.",
    href: "/activity",
    icon: <ListChecks className="h-4 w-4" />,
  },
  {
    title: "Outcome context",
    detail: "Review saved patterns that shape future work.",
    href: "/memory",
    icon: <Activity className="h-4 w-4" />,
  },
  {
    title: "What Soma can use",
    detail: "Check capability readiness and repair paths.",
    href: "/resources?tab=tools",
    icon: <Wrench className="h-4 w-4" />,
  },
];

export function prioritizeSomaHomeWorkItems(items: TeamWorkItem[]) {
  return recoveryReviewQueueItems(items);
}

export function railAlertsFromWorkItems(items: TeamWorkItem[]): DashboardRailAlert[] {
  return items
    .filter((item) => item.state !== "output_ready")
    .slice(0, 3)
    .map((item) => {
      const target = railAlertTarget(item);
      return {
        id: `work:${item.id}`,
        kind: railAlertKind(item),
        severity: railAlertSeverity(item),
        title: railAlertTitle(item),
        detail: railAlertDetail(item),
        href: target.href,
        actionLabel: target.type === "run" ? "Open receipt" : "Open item",
        targetReference: targetRefReference(item.targetRef) ?? `${target.type}:${target.id}`,
        target,
        updatedAt: item.updatedAt ?? null,
      };
    });
}

export function outcomeProjectSummaryFromWork({
  teams,
  focusedTeamId,
  workItems,
  outputRefs,
}: {
  teams: TeamDetailEntry[];
  focusedTeamId?: string | null;
  workItems: TeamWorkItem[];
  outputRefs: TeamOutputRef[];
}): OutcomeProjectSummary | null {
  const activeItems = workItems.filter((item) => item.state !== "archived");
  if (!focusedTeamId && activeItems.length === 0 && outputRefs.length === 0) return null;

  const teamIds = new Set<string>();
  if (focusedTeamId) teamIds.add(focusedTeamId);
  activeItems.forEach((item) => item.teamIds.forEach((teamId) => teamIds.add(teamId)));
  outputRefs.forEach((output) => teamIds.add(output.team_id));
  const knownTeams = teams.filter((team) => teamIds.has(team.id));
  const lead = knownTeams.find((team) => team.id === focusedTeamId) ?? knownTeams[0] ?? null;
  const recoveryCount = activeItems.filter((item) => (
    item.state === "degraded" || item.state === "needs_operator" || item.needsOperator
  )).length;
  return {
    title: lead ? `${lead.name} outcome workspace` : "Outcome project workspace",
    detail: outcomeProjectDetail(activeItems.length, outputRefs.length, recoveryCount),
    ownerLabel: "Soma",
    leadLabel: lead ? `${lead.name}${lead.role ? `, ${lead.role}` : ""}` : undefined,
    registryOwnerLabel: lead ? `${lead.name} lead` : "Soma lead",
    teamCount: Math.max(knownTeams.length, teamIds.size),
    workCount: activeItems.length,
    outputCount: outputRefs.length,
    recoveryCount,
    href: "/teams?view=work",
    hrefLabel: activeItems.length > 0 ? "Open work" : "Open team workspace",
    outputHref: outputRefs.length > 0 ? "/resources?tab=workspace" : undefined,
    outputLabel: outputRefs.length > 0 ? "Open outputs" : undefined,
  };
}

function outcomeProjectDetail(workCount: number, outputCount: number, recoveryCount: number) {
  if (recoveryCount > 0) return "Soma has team work attached to this outcome and some items need recovery.";
  if (outputCount > 0) return "Soma has retained team output for this outcome and kept the workspace open for revisit.";
  if (workCount > 0) return "Soma has assigned work toward this outcome and will keep the thread updated.";
  return "Soma is focused on this team workspace for the current outcome.";
}

function railAlertKind(item: TeamWorkItem): DashboardRailAlert["kind"] {
  if (item.runId && (item.state === "degraded" || item.state === "needs_operator")) return "run_failed";
  if (item.needsOperator || item.recoveryOptions?.length || item.state === "degraded" || item.state === "needs_operator") {
    return "recovery";
  }
  return "work_review";
}

function railAlertSeverity(item: TeamWorkItem): DashboardRailAlert["severity"] {
  if (item.state === "degraded" || item.state === "needs_operator") return "danger";
  if (item.needsOperator || item.recoveryOptions?.length) return "warning";
  return "info";
}

function railAlertTarget(item: TeamWorkItem): NonNullable<DashboardRailAlert["target"]> {
  const apiTarget = railAlertTargetFromRef(item);
  if (apiTarget) return apiTarget;
  if (item.runId) {
    return {
      type: "run",
      id: item.runId,
      href: `/runs/${encodeURIComponent(item.runId)}`,
      label: "Run receipt",
    };
  }
  return {
    type: item.needsOperator || item.recoveryOptions?.length ? "recovery" : "work",
    id: item.id,
    href: `/teams?view=work&work_item_id=${encodeURIComponent(item.id)}`,
    label: item.needsOperator || item.recoveryOptions?.length ? "Recovery item" : "Work item",
  };
}

function railAlertTargetFromRef(item: TeamWorkItem): DashboardRailAlertTarget | null {
  const targetRef = item.targetRef;
  const href = targetRefHref(targetRef);
  if (!targetRef || !href) return null;
  const type = dashboardTargetType(targetRef.type);
  if (!type) return null;
  return {
    type,
    id: targetRef.id,
    href,
    label: targetRef.label || dashboardTargetLabel(type),
  };
}

function dashboardTargetType(type: string): DashboardRailAlertTarget["type"] | null {
  if (
    type === "run" ||
    type === "work" ||
    type === "recovery" ||
    type === "capability" ||
    type === "output" ||
    type === "outcome_project"
  ) {
    return type;
  }
  return null;
}

function dashboardTargetLabel(type: DashboardRailAlertTarget["type"]) {
  if (type === "run") return "Run receipt";
  if (type === "recovery") return "Recovery item";
  if (type === "output") return "Output";
  if (type === "outcome_project") return "Outcome project";
  if (type === "capability") return "Capability";
  return "Work item";
}

function railAlertTitle(item: TeamWorkItem) {
  if (item.state === "degraded" || item.state === "needs_operator") return "Work needs attention";
  if (item.needsOperator || item.recoveryOptions?.length) return "Review needed";
  if (item.state === "running" || item.state === "queued" || item.state === "briefed") return "Work in progress";
  return "Work ready to review";
}

function railAlertDetail(item: TeamWorkItem) {
  return item.nextAction ?? item.description ?? item.title;
}
