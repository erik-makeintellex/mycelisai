"use client";

import { useEffect, useMemo, useState } from "react";
import { normalizeTargetRef, targetRefHref, targetRefReference } from "@/components/teams/teamWorkProjection";
import { extractApiData } from "@/lib/apiContracts";
import { resourcesWorkspaceHref } from "@/lib/outputPackageModel";
import type { TargetRef, TeamDetailEntry, TeamOutputRef } from "@/store/useCortexStore";
import type { OutcomeProjectSummary } from "./OutcomeProjectSummary";

type OutcomeProjectStatus = "active" | "needs_attention" | "output_ready" | "archived";

type OutcomeProject = {
  project_id: string;
  outcome_id: string;
  title: string;
  purpose?: string;
  execution_mode?: string;
  workspace_folder?: string;
  status?: OutcomeProjectStatus;
  run_id?: string;
  work_item_refs?: string[];
  output_refs?: TeamOutputRef[];
  recovery_refs?: string[];
  team_registry_refs?: string[];
  target_ref?: TargetRef;
  updated_at?: string;
};

export function useOutcomeProjectSummary({
  teams,
  focusedTeamId,
  refreshKey,
}: {
  teams: TeamDetailEntry[];
  focusedTeamId?: string | null;
  refreshKey?: number;
}) {
  const [projects, setProjects] = useState<OutcomeProject[]>([]);

  useEffect(() => {
    let cancelled = false;
    async function loadOutcomeProjects() {
      if (typeof fetch !== "function") return;
      try {
        const response = await fetch("/api/v1/outcome-projects?limit=8", { cache: "no-store" });
        if (!response.ok) return;
        const payload = await response.json();
        const data = extractApiData<unknown>(payload);
        if (!cancelled) setProjects(Array.isArray(data) ? normalizeOutcomeProjects(data) : []);
      } catch {
        if (!cancelled) setProjects([]);
      }
    }
    void loadOutcomeProjects();
    return () => {
      cancelled = true;
    };
  }, [refreshKey]);

  return useMemo(() => {
    const project = selectOutcomeProject(projects, focusedTeamId);
    return project ? outcomeProjectSummaryFromAPI(project, teams) : null;
  }, [focusedTeamId, projects, teams]);
}

export function outcomeProjectSummaryFromAPI(
  project: OutcomeProject,
  teams: TeamDetailEntry[],
): OutcomeProjectSummary {
  const outputRefs = project.output_refs ?? [];
  const registryRefs = project.team_registry_refs ?? [];
  const targetHref = targetRefHref(project.target_ref);
  const targetReference = targetRefReference(project.target_ref) ?? undefined;
  const teamIds = new Set(outputRefs.map((ref) => ref.team_id).filter(Boolean));
  const knownTeams = teams.filter((team) => teamIds.has(team.id));
  const lead = knownTeams[0] ?? null;
  const recoveryCount = (project.recovery_refs ?? []).length
    || (project.status === "needs_attention" ? 1 : 0);
  return {
    title: project.title || "Outcome workspace",
    detail: durableOutcomeDetail(project, outputRefs.length, recoveryCount),
    ownerLabel: "Soma",
    leadLabel: lead ? `${lead.name}${lead.role ? `, ${lead.role}` : ""}` : undefined,
    registryOwnerLabel: lead
      ? `${lead.name} lead`
      : registryRefs.length > 0
        ? "Registered team lead"
        : undefined,
    teamCount: Math.max(teamIds.size, registryRefs.length, knownTeams.length),
    workCount: project.work_item_refs?.length ?? 0,
    outputCount: outputRefs.length,
    recoveryCount,
    href: targetHref ?? (project.run_id ? `/runs/${encodeURIComponent(project.run_id)}` : "/teams?view=work"),
    hrefLabel: project.target_ref
      ? project.target_ref.type === "run" ? "Open run receipt" : "Open work"
      : project.run_id ? "Open run receipt" : "Open work",
    targetReference,
    outputHref: resourcesWorkspaceHref(project.workspace_folder) ?? (outputRefs.length > 0 ? "/resources?tab=workspace" : undefined),
    outputLabel: project.workspace_folder || outputRefs.length > 0 ? "Open outputs" : undefined,
  };
}

export function selectOutcomeProject(projects: OutcomeProject[], focusedTeamId?: string | null) {
  const visible = projects.filter((project) => project.status !== "archived");
  if (focusedTeamId) {
    const focused = visible.find((project) => (
      project.output_refs ?? []
    ).some((ref) => ref.team_id === focusedTeamId));
    return focused ?? null;
  }
  return visible[0] ?? null;
}

function normalizeOutcomeProjects(rawProjects: unknown[]): OutcomeProject[] {
  return rawProjects
    .map((raw) => normalizeOutcomeProject(raw))
    .filter((project): project is OutcomeProject => Boolean(project?.project_id && project.title));
}

function normalizeOutcomeProject(raw: unknown): OutcomeProject | null {
  if (!raw || typeof raw !== "object") return null;
  const project = raw as OutcomeProject;
  return {
    ...project,
    target_ref: normalizeTargetRef(project.target_ref),
  };
}

function durableOutcomeDetail(project: OutcomeProject, outputCount: number, recoveryCount: number) {
  if (recoveryCount > 0) return "Soma has retained this outcome and flagged recovery work before it should be trusted.";
  if (project.status === "output_ready" || outputCount > 0) {
    return "Soma retained this outcome as a durable project with outputs ready to revisit.";
  }
  return "Soma retained this outcome as active background work.";
}
