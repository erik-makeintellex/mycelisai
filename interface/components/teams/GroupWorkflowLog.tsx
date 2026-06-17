"use client";

import { useEffect, useMemo, useState } from "react";
import {
  CheckCircle2,
  CircleDot,
  Clock3,
  MessageSquare,
  Radio,
  ShieldAlert,
} from "lucide-react";
import type { Artifact } from "@/store/cortexStoreTypesPlanning";
import type { TeamWorkItem } from "@/store/useCortexStore";
import {
  mapDurableTeamWorkItem,
  parseTeamWorkAPIItems,
} from "./teamWorkProjection";
import type {
  Group,
  GroupBroadcastResult,
  GroupLifecycleItem,
  Monitor,
} from "./groupWorkspaceTypes";
import { relativeTime } from "./groupWorkspaceTypes";

type WorkflowEvent = {
  id: string;
  phase: string;
  title: string;
  detail: string;
  status: "ready" | "running" | "needs_review" | "complete" | "system";
  timestamp?: string | null;
  meta?: string;
};

type TeamWorkLoadState = {
  items: TeamWorkItem[];
  loading: boolean;
  failedTeamIds: string[];
};

type WorkflowLogLoadState = {
  events: WorkflowEvent[];
  loading: boolean;
  unavailable: boolean;
};

export function GroupWorkflowLog({
  selectedGroup,
  lifecycleItem,
  outputs,
  monitor,
  lastBroadcastResult,
  onOpenOutputs,
  onOpenMessage,
}: {
  selectedGroup: Group | null;
  lifecycleItem?: GroupLifecycleItem;
  outputs: Artifact[];
  monitor: Monitor | null;
  lastBroadcastResult: GroupBroadcastResult | null;
  onOpenOutputs: () => void;
  onOpenMessage: () => void;
}) {
  const workflowLog = useGroupWorkflowLog(selectedGroup);
  const teamWork = useGroupTeamWork(selectedGroup, workflowLog.unavailable);
  const events = useMemo(
    () => {
      if (workflowLog.events.length > 0) return workflowLog.events;
      return (
      buildWorkflowEvents({
        selectedGroup,
        lifecycleItem,
        outputs,
        monitor,
        lastBroadcastResult,
        teamWork,
      })
      );
    },
    [
      selectedGroup,
      lifecycleItem,
      outputs,
      monitor,
      lastBroadcastResult,
      teamWork,
      workflowLog.events,
    ],
  );

  if (!selectedGroup) {
    return (
      <section className="rounded-2xl border border-cortex-border bg-cortex-surface p-4">
        <h2 className="text-sm font-semibold uppercase tracking-[0.16em] text-cortex-text-main">
          Workflow log
        </h2>
        <p className="mt-3 text-sm leading-6 text-cortex-text-muted">
          Select a group to see one readable stream of setup, team activity,
          retained outputs, proof, and recovery signals.
        </p>
      </section>
    );
  }

  return (
    <section className="rounded-2xl border border-cortex-border bg-cortex-surface p-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-primary">
            <Radio className="h-3.5 w-3.5" />
            Workflow log
          </div>
          <h2 className="mt-3 text-base font-semibold text-cortex-text-main">
            {selectedGroup.name}
          </h2>
          <p className="mt-1 max-w-3xl text-sm leading-6 text-cortex-text-muted">
            Follow the group as one execution lane. Detailed source files and
            deeper diagnostics stay behind Outputs, Message, and Settings.
          </p>
          {workflowLog.loading || teamWork.loading ? (
            <p className="mt-2 text-xs font-mono text-cortex-text-muted">
              Checking durable workflow history...
            </p>
          ) : null}
          {workflowLog.unavailable ? (
            <p className="mt-2 text-xs text-cortex-text-muted">
              Using compatibility view until the canonical group timeline is
              available.
            </p>
          ) : null}
          {teamWork.failedTeamIds.length > 0 ? (
            <p className="mt-2 text-xs text-cortex-warning">
              Some team work could not be loaded. Visible entries may be
              incomplete.
            </p>
          ) : null}
        </div>
        <div className="flex flex-wrap gap-2">
          <button
            type="button"
            onClick={onOpenMessage}
            className="rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-main hover:border-cortex-primary/25"
          >
            Message group
          </button>
          <button
            type="button"
            onClick={onOpenOutputs}
            className="rounded-xl border border-cortex-primary/30 px-3 py-2 text-sm font-semibold text-cortex-primary"
          >
            Review outputs
          </button>
        </div>
      </div>

      <div
        className="mt-4 max-h-[min(58dvh,620px)] overflow-y-auto rounded-2xl border border-cortex-border bg-cortex-bg"
        data-testid="groups-workflow-log"
      >
        <ol className="divide-y divide-cortex-border/70">
          {events.map((event) => (
            <li key={event.id} className="grid gap-3 p-4 sm:grid-cols-[120px_minmax(0,1fr)]">
              <div className="flex items-start gap-2 text-xs font-mono uppercase tracking-[0.14em] text-cortex-text-muted">
                <WorkflowStatusIcon status={event.status} />
                <span>{event.phase}</span>
              </div>
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                  <h3 className="text-sm font-semibold text-cortex-text-main">
                    {event.title}
                  </h3>
                  <span className={statusClassName(event.status)}>
                    {statusLabel(event.status)}
                  </span>
                </div>
                <p className="mt-1 text-sm leading-6 text-cortex-text-muted">
                  {event.detail}
                </p>
                <div className="mt-2 flex flex-wrap gap-2 text-[11px] font-mono text-cortex-text-muted">
                  {event.timestamp ? (
                    <span>{relativeTime(event.timestamp)}</span>
                  ) : null}
                  {event.meta ? <span>{event.meta}</span> : null}
                </div>
              </div>
            </li>
          ))}
        </ol>
      </div>
    </section>
  );
}

function useGroupWorkflowLog(selectedGroup: Group | null): WorkflowLogLoadState {
  const [state, setState] = useState<WorkflowLogLoadState>({
    events: [],
    loading: false,
    unavailable: false,
  });

  useEffect(() => {
    let cancelled = false;

    const load = async () => {
      if (!selectedGroup) {
        setState({ events: [], loading: false, unavailable: false });
        return;
      }

      setState((current) => ({ ...current, loading: true }));
      try {
        const response = await fetch(
          `/api/v1/groups/${encodeURIComponent(selectedGroup.group_id)}/workflow-log?limit=50&include_outputs=true&include_audit=true`,
          { cache: "no-store" },
        );
        if (!response.ok) {
          if (!cancelled) {
            setState({ events: [], loading: false, unavailable: true });
          }
          return;
        }
        const payload = await response.json();
        const events = parseWorkflowLogEvents(payload);
        if (!cancelled) {
          setState({ events, loading: false, unavailable: false });
        }
      } catch {
        if (!cancelled) {
          setState({ events: [], loading: false, unavailable: true });
        }
      }
    };

    void load();
    return () => {
      cancelled = true;
    };
  }, [selectedGroup?.group_id]);

  return state;
}

function useGroupTeamWork(
  selectedGroup: Group | null,
  enabled: boolean,
): TeamWorkLoadState {
  const [state, setState] = useState<TeamWorkLoadState>({
    items: [],
    loading: false,
    failedTeamIds: [],
  });
  const teamKey = selectedGroup?.team_ids.join("|") ?? "";

  useEffect(() => {
    let cancelled = false;

    const load = async () => {
      if (!enabled || !selectedGroup || selectedGroup.team_ids.length === 0) {
        setState({ items: [], loading: false, failedTeamIds: [] });
        return;
      }

      setState((current) => ({ ...current, loading: true }));
      const fetchedItems: TeamWorkItem[] = [];
      const failedTeamIds: string[] = [];

      await Promise.all(
        selectedGroup.team_ids.map(async (teamID) => {
          try {
            const response = await fetch(
              `/api/v1/teams/${encodeURIComponent(teamID)}/work?limit=8&include_archived=false`,
              { cache: "no-store" },
            );
            if (!response.ok) {
              failedTeamIds.push(teamID);
              return;
            }
            const payload = await response.json();
            const items = parseTeamWorkAPIItems(payload)
              .map((item) => mapDurableTeamWorkItem(item))
              .filter((item): item is TeamWorkItem => Boolean(item));
            fetchedItems.push(...items);
          } catch {
            failedTeamIds.push(teamID);
          }
        }),
      );

      if (cancelled) return;
      setState({
        items: sortTeamWorkItems(fetchedItems),
        loading: false,
        failedTeamIds,
      });
    };

    void load();
    return () => {
      cancelled = true;
    };
  }, [enabled, selectedGroup?.group_id, teamKey]);

  return state;
}

function parseWorkflowLogEvents(payload: unknown): WorkflowEvent[] {
  const data =
    payload && typeof payload === "object" && "data" in payload
      ? (payload as { data?: unknown }).data
      : payload;
  const rawEvents =
    data && typeof data === "object" && "timeline" in data
      ? (data as { timeline?: unknown }).timeline
      : data && typeof data === "object" && "events" in data
        ? (data as { events?: unknown }).events
      : [];
  if (!Array.isArray(rawEvents)) return [];
  return rawEvents
    .map((event, index) => normalizeWorkflowEvent(event, index))
    .filter((event): event is WorkflowEvent => Boolean(event));
}

function normalizeWorkflowEvent(event: unknown, index: number): WorkflowEvent | null {
  if (!event || typeof event !== "object") return null;
  const record = event as Record<string, unknown>;
  const title = stringValue(record.title);
  const detail =
    stringValue(record.detail) ||
    stringValue(record.summary) ||
    refsSummary(record);
  if (!title || !detail) return null;
  return {
    id: stringValue(record.id) || `workflow-event-${index}`,
    phase: stringValue(record.phase) || phaseFromKind(record.kind),
    title,
    detail,
    status: normalizeWorkflowStatus(record.status || record.state),
    timestamp: stringValue(record.timestamp) || null,
    meta: stringValue(record.meta) || timelineMeta(record),
  };
}

function normalizeWorkflowStatus(status: unknown): WorkflowEvent["status"] {
  if (
    status === "ready" ||
    status === "running" ||
    status === "needs_review" ||
    status === "complete" ||
    status === "system"
  ) {
    return status;
  }
  if (status === "output_ready" || status === "archived" || status === "approved")
    return "complete";
  if (
    status === "degraded" ||
    status === "needs_operator" ||
    status === "reviewing" ||
    status === "review_work" ||
    status === "review_standing" ||
    status === "archive_expired"
  ) {
    return "needs_review";
  }
  if (status === "queued" || status === "running") return "running";
  if (status === "new" || status === "briefed" || status === "keep_active")
    return "ready";
  return "system";
}

function stringValue(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function phaseFromKind(kind: unknown): string {
  switch (kind) {
    case "group_brief":
      return "Brief";
    case "lifecycle":
      return "Review";
    case "team_work":
      return "Work";
    case "team_output_ref":
    case "retained_artifact":
      return "Output";
    case "proof_cue":
      return "Proof";
    case "broadcast":
      return "Message";
    case "degraded":
      return "Recovery";
    default:
      return "Event";
  }
}

function timelineMeta(record: Record<string, unknown>): string {
  const parts = [
    stringValue(record.team_id),
    stringValue(record.run_id) ? `run ${stringValue(record.run_id)}` : "",
    countLabel(record.output_refs, "output ref"),
    countLabel(record.proof_refs, "proof ref"),
    countLabel(record.audit_refs, "audit ref"),
    stringValue(record.trust_posture),
  ].filter(Boolean);
  return parts.join(" • ");
}

function refsSummary(record: Record<string, unknown>): string {
  const parts = [
    countLabel(record.output_refs, "output ref"),
    countLabel(record.proof_refs, "proof ref"),
    countLabel(record.audit_refs, "audit ref"),
    stringValue(record.storage_ref) ? `Stored at ${stringValue(record.storage_ref)}.` : "",
  ].filter(Boolean);
  return parts.join(" ");
}

function countLabel(value: unknown, label: string): string {
  if (!Array.isArray(value) || value.length === 0) return "";
  return `${value.length} ${label}${value.length === 1 ? "" : "s"}`;
}

function buildWorkflowEvents({
  selectedGroup,
  lifecycleItem,
  outputs,
  monitor,
  lastBroadcastResult,
  teamWork,
}: {
  selectedGroup: Group | null;
  lifecycleItem?: GroupLifecycleItem;
  outputs: Artifact[];
  monitor: Monitor | null;
  lastBroadcastResult: GroupBroadcastResult | null;
  teamWork: TeamWorkLoadState;
}): WorkflowEvent[] {
  if (!selectedGroup) return [];

  const events: WorkflowEvent[] = [
    {
      id: "group-created",
      phase: "Brief",
      title: "Group lane created",
      detail: selectedGroup.goal_statement || "No goal statement was stored.",
      status: selectedGroup.status === "archived" ? "complete" : "ready",
      timestamp: selectedGroup.created_at,
      meta: `${selectedGroup.work_mode} • ${selectedGroup.team_ids.length} team${selectedGroup.team_ids.length === 1 ? "" : "s"}`,
    },
  ];

  if (selectedGroup.team_ids.length > 0) {
    events.push({
      id: "team-lanes",
      phase: "Team",
      title: "Team lane available",
      detail: `Lead work can continue through ${selectedGroup.team_ids.join(", ")}.`,
      status: "running",
      meta: "Open the lead workspace from Overview when focused work is needed.",
    });
  }

  if (lifecycleItem) {
    events.push({
      id: "lifecycle",
      phase: "Review",
      title: lifecycleTitle(lifecycleItem),
      detail: lifecycleItem.reason,
      status: lifecycleStatus(lifecycleItem),
      timestamp: lifecycleItem.latest_work_at,
      meta: `${lifecycleItem.team_work_count} work item${lifecycleItem.team_work_count === 1 ? "" : "s"} • ${lifecycleItem.output_count} output${lifecycleItem.output_count === 1 ? "" : "s"}`,
    });
  }

  teamWork.items.slice(0, 8).forEach((item) => {
    const outputCount = item.outputRefs?.length ?? item.outputCount ?? 0;
    const proofCount =
      (item.proofRefs?.length ?? 0) + (item.auditRefs?.length ?? 0);
    events.push({
      id: `work-${item.id}`,
      phase: "Work",
      title: item.title,
      detail: teamWorkDetail(item),
      status: teamWorkStatus(item),
      timestamp: item.updatedAt,
      meta: [
        item.teamIds?.[0],
        item.runId ? `run ${item.runId}` : null,
        outputCount ? `${outputCount} output${outputCount === 1 ? "" : "s"}` : null,
        proofCount ? `${proofCount} proof ref${proofCount === 1 ? "" : "s"}` : null,
      ]
        .filter(Boolean)
        .join(" • "),
    });
  });

  teamWork.failedTeamIds.forEach((teamID) => {
    events.push({
      id: `work-load-failed-${teamID}`,
      phase: "Work",
      title: "Team work could not be loaded",
      detail: `Durable work state for ${teamID} was unavailable during this refresh.`,
      status: "needs_review",
      meta: "Refresh or open the lead workspace if this team should be active.",
    });
  });

  outputs.slice(0, 6).forEach((artifact, index) => {
    events.push({
      id: `output-${artifact.id || index}`,
      phase: "Output",
      title: artifact.title || "Retained output",
      detail: outputDetail(artifact),
      status: artifact.status === "approved" ? "complete" : "needs_review",
      timestamp: artifact.created_at,
      meta: [artifact.agent_id, artifact.artifact_type].filter(Boolean).join(" • "),
    });
  });

  if (lastBroadcastResult?.execution_summary) {
    events.push({
      id: "latest-message",
      phase: "Message",
      title: "Latest group message queued",
      detail:
        lastBroadcastResult.execution_summary.execution_summary ||
        lastBroadcastResult.execution_summary.ui_response_state?.detail ||
        "The selected group received the latest operator message.",
      status: "running",
      meta:
        lastBroadcastResult.execution_summary.execution_status ||
        lastBroadcastResult.execution_summary.ui_response_state?.label,
    });
  }

  if (monitor) {
    events.push({
      id: "bus-monitor",
      phase: "System",
      title: `Message bus ${monitor.status || "unknown"}`,
      detail: monitor.last_error || "Group coordination monitor is reporting.",
      status: monitor.last_error ? "needs_review" : "system",
      timestamp: monitor.last_published_at,
      meta: `${monitor.published_count ?? 0} published message${monitor.published_count === 1 ? "" : "s"}`,
    });
  }

  return events;
}

function teamWorkDetail(item: TeamWorkItem): string {
  const outputLabel = item.outputRefs?.[0]?.label
    ? `Output: ${item.outputRefs[0].label}`
    : null;
  return [
    item.description,
    outputLabel,
    item.nextAction ? `Next: ${item.nextAction}` : null,
  ]
    .filter(Boolean)
    .join(" ");
}

function teamWorkStatus(item: TeamWorkItem): WorkflowEvent["status"] {
  if (item.state === "output_ready" || item.state === "archived")
    return "complete";
  if (item.state === "degraded" || item.state === "needs_operator")
    return "needs_review";
  if (item.state === "queued" || item.state === "running" || item.state === "reviewing")
    return "running";
  return "ready";
}

function sortTeamWorkItems(items: TeamWorkItem[]) {
  return [...items].sort((left, right) => {
    const leftTime = left.updatedAt ? Date.parse(left.updatedAt) : 0;
    const rightTime = right.updatedAt ? Date.parse(right.updatedAt) : 0;
    return rightTime - leftTime;
  });
}

function outputDetail(artifact: Artifact): string {
  const path = artifact.file_path || "";
  if (path) return `Stored output is available at ${path}.`;
  if (artifact.content) return "Retained text output is available for review.";
  return "Retained output metadata is available for review.";
}

function lifecycleTitle(item: GroupLifecycleItem): string {
  switch (item.recommendation) {
    case "archive_expired":
      return "Temporary group is ready to archive";
    case "review_work":
      return "Work needs review";
    case "review_standing":
      return "Standing group needs review";
    case "archive_completed":
      return "Completed group can be archived";
    case "retained":
      return "Retained history";
    default:
      return "Group can continue";
  }
}

function lifecycleStatus(item: GroupLifecycleItem): WorkflowEvent["status"] {
  if (item.recommendation === "review_work" || item.recommendation === "review_standing")
    return "needs_review";
  if (item.recommendation === "archive_completed" || item.recommendation === "retained")
    return "complete";
  if (item.recommendation === "archive_expired") return "needs_review";
  return "ready";
}

function WorkflowStatusIcon({ status }: { status: WorkflowEvent["status"] }) {
  if (status === "complete") return <CheckCircle2 className="mt-0.5 h-3.5 w-3.5 text-cortex-success" />;
  if (status === "needs_review") return <ShieldAlert className="mt-0.5 h-3.5 w-3.5 text-cortex-warning" />;
  if (status === "running") return <Clock3 className="mt-0.5 h-3.5 w-3.5 text-cortex-primary" />;
  if (status === "system") return <MessageSquare className="mt-0.5 h-3.5 w-3.5 text-cortex-text-muted" />;
  return <CircleDot className="mt-0.5 h-3.5 w-3.5 text-cortex-text-muted" />;
}

function statusLabel(status: WorkflowEvent["status"]) {
  if (status === "needs_review") return "needs review";
  if (status === "complete") return "complete";
  if (status === "running") return "active";
  if (status === "system") return "system";
  return "ready";
}

function statusClassName(status: WorkflowEvent["status"]) {
  const base =
    "rounded-full border px-2 py-0.5 text-[10px] font-mono uppercase tracking-[0.14em]";
  if (status === "needs_review")
    return `${base} border-cortex-warning/40 bg-cortex-warning/10 text-cortex-warning`;
  if (status === "complete")
    return `${base} border-cortex-success/35 bg-cortex-success/10 text-cortex-success`;
  if (status === "running")
    return `${base} border-cortex-primary/35 bg-cortex-primary/10 text-cortex-primary`;
  return `${base} border-cortex-border bg-cortex-surface text-cortex-text-muted`;
}
