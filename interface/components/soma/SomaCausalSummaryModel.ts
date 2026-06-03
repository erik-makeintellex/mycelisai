import type {
  ChatArtifactRef,
  ChatMessage,
  ExecutionSummaryCapabilityUse,
  ExecutionSummaryData,
  ExecutionSummaryItem,
  ExecutionSummaryLink,
} from "@/store/useCortexStore";

type SummaryValue = string | ExecutionSummaryItem;

export function lastMessage(messages: ChatMessage[], role: ChatMessage["role"]) {
  return [...messages].reverse().find((message) => message.role === role);
}

export function lastSomaMessage(messages: ChatMessage[]) {
  return [...messages]
    .reverse()
    .find((message) => message.role !== "user" && message.role !== "system");
}

export function compactText(value?: string | null) {
  const trimmed = value?.trim();
  return trimmed || null;
}

export function itemText(item: SummaryValue): string | null {
  if (typeof item === "string") return compactText(item);
  return compactText(item.label)
    ?? compactText(item.title)
    ?? compactText(item.name)
    ?? compactText(item.summary)
    ?? compactText(item.value)
    ?? compactText(item.id)
    ?? null;
}

function firstText(values: Array<string | null | undefined>) {
  return values.find((value): value is string => Boolean(compactText(value))) ?? null;
}

export function intentText(intent: ExecutionSummaryData["intent"]) {
  if (!intent) return null;
  if (typeof intent === "string") return compactText(intent);
  return firstText([
    compactText(intent.resolved),
    compactText(intent.original),
  ]);
}

export function understandingText(understanding: ExecutionSummaryData["understanding"]) {
  if (!understanding) return null;
  if (typeof understanding === "string") return compactText(understanding);
  const assumptions = understanding.assumptions?.filter(Boolean).join("; ");
  const summary = compactText(understanding.summary);
  return [summary, assumptions ? `Assumptions: ${assumptions}` : null].filter(Boolean).join(" ") || null;
}

export function asItems(value: ExecutionSummaryData["outputs"]): SummaryValue[] {
  if (!value) return [];
  if (typeof value === "string") return [value];
  return value;
}

export function linkLabel(link: string | ExecutionSummaryLink): string | null {
  if (typeof link === "string") return compactText(link);
  return compactText(link.label)
    ?? compactText(link.title)
    ?? (compactText(link.run_id) ? `Run ${link.run_id}` : null)
    ?? (compactText(link.audit_event_id) ? `Audit ${link.audit_event_id}` : null)
    ?? (compactText(link.intent_proof_id) ? `Proof ${link.intent_proof_id}` : null)
    ?? compactText(link.id)
    ?? compactText(link.path)
    ?? compactText(link.url)
    ?? compactText(link.href)
    ?? null;
}

export function proofLinks(proof: ExecutionSummaryData["proof"]): Array<string | ExecutionSummaryLink> {
  if (!proof) return [];
  return Array.isArray(proof) ? proof : [proof];
}

export function capabilityText(capabilityUse: ExecutionSummaryData["capability_use"]) {
  if (!capabilityUse) return null;
  if (Array.isArray(capabilityUse)) {
    const values = capabilityUse.map(itemText).filter(Boolean);
    return values.length ? values.join(", ") : null;
  }

  const source = capabilityUse as ExecutionSummaryCapabilityUse;
  const parts = [
    ["Teams", source.teams],
    ["Agents", source.agents],
    ["Capabilities", source.capabilities],
    ["Tools", source.tools],
    ["Used", source.used],
  ].flatMap(([label, values]) => {
    const text = (values as Array<string | ExecutionSummaryItem> | undefined)
      ?.map(itemText)
      .filter(Boolean)
      .join(", ");
    return text ? [`${label}: ${text}`] : [];
  });

  return parts.length ? parts.join(" | ") : null;
}

export function auditText(value: ExecutionSummaryData["audit_recovery"]) {
  if (!value) return null;
  if (typeof value === "string") return compactText(value);
  const status = compactText(value.status) ?? compactText(value.approval_status);
  const recovery = compactText(value.recovery_state);
  const summary = compactText(value.summary) ?? compactText(value.value) ?? compactText(value.label);
  const blocker = compactText(value.blocker);
  return [status, recovery, summary, blocker].filter(Boolean).join(": ") || null;
}

export function nextStepText(value: ExecutionSummaryData["next_step"]) {
  if (!value) return null;
  if (typeof value === "string") return compactText(value);
  return compactText(value.label)
    ?? compactText(value.title)
    ?? compactText(value.action)
    ?? compactText(value.href)
    ?? compactText(value.url)
    ?? null;
}

export function artifactLabels(artifacts?: ChatArtifactRef[]) {
  return artifacts?.map((artifact) => artifact.title || artifact.type || artifact.id || "artifact").filter(Boolean) ?? [];
}

export function executionShapeText(value?: string | null) {
  const shape = compactText(value);
  if (!shape) return null;
  const labels: Record<string, string> = {
    directed_execution: "completed work",
    team_execution: "team work",
    native_team: "team work",
    tool_assisted_work: "tool-assisted work",
    proposal: "proposal",
    guided_proposal: "proposal",
  };
  return labels[shape] ?? shape.replace(/[_-]+/g, " ");
}
