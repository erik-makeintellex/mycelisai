export type ConversationalProposalReply = "confirm" | "cancel" | null;

const CONFIRM_REPLIES = new Set([
  "approve",
  "approved",
  "do it",
  "go ahead",
  "looks good",
  "proceed",
  "run it",
  "start",
  "start it",
  "yes approve",
  "yes go ahead",
  "yes proceed",
]);

const CANCEL_REPLIES = new Set([
  "cancel",
  "cancel it",
  "do not run",
  "dont run",
  "never mind",
  "nevermind",
  "stop",
]);

function normalizedReply(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[.!?]+$/g, "")
    .replace(/['’]/g, "")
    .replace(/\s+/g, " ");
}

export function conversationalProposalReply(value: string): ConversationalProposalReply {
  const normalized = normalizedReply(value);
  if (CONFIRM_REPLIES.has(normalized)) return "confirm";
  if (CANCEL_REPLIES.has(normalized)) return "cancel";
  return null;
}
