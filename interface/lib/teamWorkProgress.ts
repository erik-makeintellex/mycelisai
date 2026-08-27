import type {
  ChatMessage,
  SomaThreadEvent,
  TeamWorkItemState,
} from "@/store/useCortexStore";

export type TeamWorkProgressState =
  | "planning"
  | "queued"
  | "building"
  | "validating"
  | "ready"
  | "recovery"
  | "archived";

export type TeamWorkProgressTone = "info" | "success" | "warning";

export interface TeamWorkProgress {
  state: TeamWorkProgressState;
  label: string;
  detail: string;
  tone: TeamWorkProgressTone;
}

const PROGRESS_COPY: Record<TeamWorkProgressState, TeamWorkProgress> = {
  planning: {
    state: "planning",
    label: "Planning",
    detail: "Review the proposed outcome before work starts.",
    tone: "info",
  },
  queued: {
    state: "queued",
    label: "Queued",
    detail: "The approved work is waiting for the team to accept it.",
    tone: "info",
  },
  building: {
    state: "building",
    label: "Building",
    detail: "The team is working on the requested outcome.",
    tone: "info",
  },
  validating: {
    state: "validating",
    label: "Validating",
    detail: "The deliverable is being checked and is not ready yet.",
    tone: "info",
  },
  ready: {
    state: "ready",
    label: "Ready",
    detail: "The validated deliverable is ready to open.",
    tone: "success",
  },
  recovery: {
    state: "recovery",
    label: "Recovery",
    detail: "The work needs a response or a safe recovery step.",
    tone: "warning",
  },
  archived: {
    state: "archived",
    label: "Archived",
    detail: "This work is retained outside active progress.",
    tone: "info",
  },
};

export function progressForDurableState(state: TeamWorkItemState): TeamWorkProgress {
  if (state === "new" || state === "briefed") return PROGRESS_COPY.planning;
  if (state === "queued") return PROGRESS_COPY.queued;
  if (state === "running" || state === "paused") return PROGRESS_COPY.building;
  if (state === "reviewing") return PROGRESS_COPY.validating;
  if (state === "output_ready") return PROGRESS_COPY.ready;
  if (state === "degraded" || state === "needs_operator") return PROGRESS_COPY.recovery;
  return PROGRESS_COPY.archived;
}

export function progressForThreadEvent(event: SomaThreadEvent): TeamWorkProgress {
  const status = event.status?.trim().toLowerCase();
  if (event.kind === "attention_required") return PROGRESS_COPY.recovery;
  if (event.kind === "result_ready") return PROGRESS_COPY.ready;
  if (status === "reviewing" || status === "validating") return PROGRESS_COPY.validating;
  if (status === "queued" || status === "confirming" || status === "pending") return PROGRESS_COPY.queued;
  return PROGRESS_COPY.building;
}

export function progressForChatMessage(message: ChatMessage): TeamWorkProgress | null {
  const event = message.thread_events?.at(-1) ?? message.thread_event;
  if (event) return progressForThreadEvent(event);
  if (message.proposal) return PROGRESS_COPY.planning;
  return null;
}
