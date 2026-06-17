import { describe, expect, it } from "vitest";
import {
  projectCapabilityRecoveryItems,
  projectMissionEventRecoveryItems,
  projectRecoveryQueue,
  projectRunRecoveryItems,
  projectSearchRecoveryItems,
  projectTeamWorkRecoveryItems,
} from "@/components/recovery/recoveryQueueProjection";
import type {
  CapabilityManifest,
  MissionEvent,
  MissionRun,
  SearchCapabilityStatus,
  TeamWorkItem,
} from "@/store/useCortexStore";

describe("recoveryQueueProjection", () => {
  it("normalizes degraded and operator-needed team work", () => {
    const workItems: TeamWorkItem[] = [
      {
        id: "work-1",
        title: "Finalize package",
        state: "needs_operator",
        ownerLabel: "Alpha lead",
        scopeLabel: "Durable team work",
        updatedAt: "2026-06-17T14:00:00Z",
        teamIds: ["team-alpha"],
        interactions: [{ action: "inspect", label: "Open run", href: "/runs/run-1" }],
        runId: "run-1",
        needsOperator: true,
        nextAction: "Pick recovery path",
        proofRefs: ["proof-1"],
        auditRefs: ["audit-1"],
      },
      {
        id: "work-2",
        title: "Healthy work",
        state: "running",
        ownerLabel: "Beta lead",
        scopeLabel: "Durable team work",
        teamIds: ["team-beta"],
        interactions: [],
      },
    ];

    expect(projectTeamWorkRecoveryItems(workItems)).toEqual([
      expect.objectContaining({
        id: "team-work:work-1",
        source: "team_work",
        severity: "needs_operator",
        title: "Finalize package",
        runId: "run-1",
        teamIds: ["team-alpha"],
        actionLabel: "Pick recovery path",
        actionHref: "/runs/run-1",
        evidenceRefs: ["proof-1", "audit-1"],
      }),
    ]);
  });

  it("projects failed runs and failed/degraded mission events", () => {
    const runs: MissionRun[] = [
      {
        id: "run-failed",
        mission_id: "mission-1",
        tenant_id: "tenant-1",
        status: "failed",
        run_depth: 0,
        started_at: "2026-06-17T13:00:00Z",
        completed_at: "2026-06-17T13:03:00Z",
        metadata: { summary: "Renderer was unavailable.", audit_event_id: "audit-run" },
      },
      {
        id: "run-ok",
        mission_id: "mission-1",
        tenant_id: "tenant-1",
        status: "completed",
        run_depth: 0,
        started_at: "2026-06-17T12:00:00Z",
      },
    ];
    const events: MissionEvent[] = [
      {
        id: "event-1",
        run_id: "run-2",
        tenant_id: "tenant-1",
        event_type: "tool.degraded",
        severity: "warning",
        source_agent: "agent-1",
        source_team: "team-alpha",
        payload: { message: "Search fell back to local sources.", proof_id: "proof-event" },
        audit_event_id: "audit-event",
        emitted_at: "2026-06-17T13:05:00Z",
      },
    ];

    expect(projectRunRecoveryItems(runs)).toEqual([
      expect.objectContaining({
        id: "run:run-failed",
        source: "run",
        severity: "failed",
        detail: "Renderer was unavailable.",
        actionHref: "/runs/run-failed",
        evidenceRefs: ["audit-run"],
      }),
    ]);
    expect(projectMissionEventRecoveryItems(events)).toEqual([
      expect.objectContaining({
        id: "mission-event:event-1",
        source: "mission_event",
        severity: "degraded",
        detail: "Search fell back to local sources.",
        runId: "run-2",
        teamIds: ["team-alpha"],
        agentId: "agent-1",
        evidenceRefs: ["audit-event", "proof-event"],
      }),
    ]);
  });

  it("projects search and capability blockers without backend wiring", () => {
    const search: SearchCapabilityStatus = {
      provider: "brave",
      enabled: true,
      configured: false,
      supports_local_sources: true,
      supports_public_web: true,
      soma_tool_name: "web_search",
      direct_soma_interaction: true,
      requires_hosted_api_token: true,
      max_results: 8,
      blocker: {
        code: "missing_secret",
        message: "BRAVE_API_KEY is not configured.",
        next_action: "Add BRAVE_API_KEY to the local secret store.",
      },
    };
    const capabilities: CapabilityManifest[] = [
      {
        id: "search.web",
        name: "Web search",
        source: "builtin",
        category: "research",
        risk: "medium",
        approval: "optional",
        availability_status: "unavailable",
        fallback_behavior: "Configure search before retrying web research.",
        secret_refs: ["BRAVE_API_KEY"],
      },
      {
        id: "files.read",
        name: "Read files",
        source: "builtin",
        category: "filesystem",
        risk: "low",
        approval: "policy_resolved",
        availability_status: "available",
      },
    ];

    expect(projectSearchRecoveryItems(search)).toEqual([
      expect.objectContaining({
        id: "search:capability",
        source: "search",
        severity: "blocked",
        detail: "BRAVE_API_KEY is not configured.",
        recoveryOptions: ["Add BRAVE_API_KEY to the local secret store."],
      }),
    ]);
    expect(projectCapabilityRecoveryItems(capabilities)).toEqual([
      expect.objectContaining({
        id: "capability:search.web",
        source: "capability",
        severity: "blocked",
        evidenceRefs: ["search.web", "BRAVE_API_KEY"],
      }),
    ]);
  });

  it("sorts by operator priority, then recency, and applies the queue limit", () => {
    const queue = projectRecoveryQueue({
      teamWorkItems: [{
        id: "work-old",
        title: "Operator decision",
        state: "needs_operator",
        ownerLabel: "Alpha lead",
        scopeLabel: "Durable team work",
        updatedAt: "2026-06-17T12:00:00Z",
        teamIds: ["team-alpha"],
        interactions: [],
      }],
      recentRuns: [{
        id: "run-new",
        mission_id: "mission-1",
        tenant_id: "tenant-1",
        status: "failed",
        run_depth: 0,
        started_at: "2026-06-17T15:00:00Z",
      }],
      searchCapabilityError: "Search status endpoint is unavailable.",
      limit: 2,
    });

    expect(queue.map((item) => item.id)).toEqual([
      "team-work:work-old",
      "run:run-new",
    ]);
  });

  it("keeps MCP registry errors as bounded recovery blockers", () => {
    expect(projectRecoveryQueue({
      mcpServersError: "MCP registry unreachable (HTTP 500)",
      mcpToolSetsError: "MCP access layers unreachable (HTTP 503)",
    })).toEqual([
      expect.objectContaining({
        id: "registry:mcp-servers",
        source: "registry",
        severity: "blocked",
        detail: "MCP registry unreachable (HTTP 500)",
      }),
      expect.objectContaining({
        id: "registry:mcp-toolsets",
        source: "registry",
        severity: "blocked",
        detail: "MCP access layers unreachable (HTTP 503)",
      }),
    ]);
  });
});
