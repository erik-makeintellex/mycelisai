// V7 Event Spine — TypeScript types for mission runs and events.
// Mirrors core/pkg/protocol/events.go and core/internal/runs/manager.go.
// Used by RunTimeline.tsx, EventCard.tsx, ViewChain.tsx (Team E components).

export type EventType =
  | 'mission.started'
  | 'mission.completed'
  | 'mission.failed'
  | 'mission.cancelled'
  | 'team.spawned'
  | 'team.stopped'
  | 'agent.started'
  | 'agent.stopped'
  | 'tool.invoked'
  | 'tool.completed'
  | 'tool.failed'
  | 'artifact.created'
  | 'memory.stored'
  | 'memory.recalled'
  | 'trigger.fired'
  | 'trigger.skipped'
  | 'scheduler.tick';

export type EventSeverity = 'debug' | 'info' | 'warn' | 'error';

export type RunStatus = 'pending' | 'running' | 'completed' | 'failed';

// MissionEventEnvelope is the authoritative audit record for a mission execution event.
// Returned by GET /api/v1/runs/{id}/events.
export interface MissionEventEnvelope {
  id: string;
  run_id: string;
  tenant_id: string;
  event_type: EventType;
  severity: EventSeverity;
  source_agent?: string;
  source_team?: string;
  payload?: Record<string, unknown>;
  audit_event_id?: string;
  emitted_at: string; // ISO 8601
}

// MissionRun is a single execution instance of a mission definition.
// Returned by GET /api/v1/runs/{id}/chain (chain array).
export interface MissionRun {
  id: string;
  mission_id: string;
  tenant_id: string;
  status: RunStatus;
  run_depth: number;
  parent_run_id?: string;
  started_at: string; // ISO 8601
  completed_at?: string; // ISO 8601
  metadata?: Record<string, unknown>;
}

// RunChainResponse is returned by GET /api/v1/runs/{id}/chain.
export interface RunChainResponse {
  run_id: string;
  mission_id: string;
  chain: MissionRun[];
}

// Color mapping for event types — used by RunTimeline and EventCard.
export const EVENT_TYPE_COLORS: Record<string, string> = {
  'mission.started':   '#10b981', // cortex-success green
  'mission.completed': '#10b981',
  'mission.failed':    '#ef4444', // red
  'mission.cancelled': '#f59e0b', // amber
  'team.spawned':      '#06b6d4', // cortex-primary cyan
  'team.stopped':      '#71717a', // muted
  'agent.started':     '#06b6d4',
  'agent.stopped':     '#71717a',
  'tool.invoked':      '#8b5cf6', // violet
  'tool.completed':    '#10b981',
  'tool.failed':       '#ef4444',
  'artifact.created':  '#f59e0b',
  'memory.stored':     '#06b6d4',
  'memory.recalled':   '#06b6d4',
  'trigger.fired':     '#f59e0b',
  'trigger.skipped':   '#71717a',
  'scheduler.tick':    '#71717a',
};

export const SEVERITY_COLORS: Record<EventSeverity, string> = {
  debug: '#71717a',
  info:  '#06b6d4',
  warn:  '#f59e0b',
  error: '#ef4444',
};
