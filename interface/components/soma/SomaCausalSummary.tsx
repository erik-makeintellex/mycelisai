import { CheckCircle2, FileText, Gauge, RotateCcw, Route, ShieldCheck, Sparkles } from "lucide-react";
import type React from "react";
import type {
  ChatArtifactRef,
  ChatMessage,
  ExecutionSummaryCapabilityUse,
  ExecutionSummaryData,
  ExecutionSummaryItem,
  ExecutionSummaryLink,
} from "@/store/useCortexStore";

type SummaryValue = string | ExecutionSummaryItem;

function lastMessage(messages: ChatMessage[], role: ChatMessage["role"]) {
  return [...messages].reverse().find((message) => message.role === role);
}

function lastSomaMessage(messages: ChatMessage[]) {
  return [...messages]
    .reverse()
    .find((message) => message.role !== "user" && message.role !== "system");
}

function compactText(value?: string | null) {
  const trimmed = value?.trim();
  return trimmed || null;
}

function itemText(item: SummaryValue): string | null {
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

function intentText(intent: ExecutionSummaryData["intent"]) {
  if (!intent) return null;
  if (typeof intent === "string") return compactText(intent);
  return firstText([
    compactText(intent.resolved),
    compactText(intent.original),
  ]);
}

function understandingText(understanding: ExecutionSummaryData["understanding"]) {
  if (!understanding) return null;
  if (typeof understanding === "string") return compactText(understanding);
  const assumptions = understanding.assumptions?.filter(Boolean).join("; ");
  const summary = compactText(understanding.summary);
  return [summary, assumptions ? `Assumptions: ${assumptions}` : null].filter(Boolean).join(" ") || null;
}

function asItems(value: ExecutionSummaryData["outputs"]): SummaryValue[] {
  if (!value) return [];
  if (typeof value === "string") return [value];
  return value;
}

function linkLabel(link: string | ExecutionSummaryLink): string | null {
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

function proofLinks(proof: ExecutionSummaryData["proof"]): Array<string | ExecutionSummaryLink> {
  if (!proof) return [];
  return Array.isArray(proof) ? proof : [proof];
}

function capabilityText(capabilityUse: ExecutionSummaryData["capability_use"]) {
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

function auditText(value: ExecutionSummaryData["audit_recovery"]) {
  if (!value) return null;
  if (typeof value === "string") return compactText(value);
  const status = compactText(value.status) ?? compactText(value.approval_status);
  const recovery = compactText(value.recovery_state);
  const summary = compactText(value.summary) ?? compactText(value.value) ?? compactText(value.label);
  const blocker = compactText(value.blocker);
  return [status, recovery, summary, blocker].filter(Boolean).join(": ") || null;
}

function nextStepText(value: ExecutionSummaryData["next_step"]) {
  if (!value) return null;
  if (typeof value === "string") return compactText(value);
  return compactText(value.label)
    ?? compactText(value.title)
    ?? compactText(value.action)
    ?? compactText(value.href)
    ?? compactText(value.url)
    ?? null;
}

function artifactLabels(artifacts?: ChatArtifactRef[]) {
  return artifacts?.map((artifact) => artifact.title || artifact.type || artifact.id || "artifact").filter(Boolean) ?? [];
}

export function SomaCausalSummary({
  messages,
  fallbackAction = "Ready for your first Soma request",
  teams = ["Soma"],
  outputs = ["Conversation guidance"],
  updated = ["Soma thread"],
}: {
  messages: ChatMessage[];
  fallbackAction?: string;
  teams?: string[];
  outputs?: string[];
  updated?: string[];
}) {
  const latestUser = lastMessage(messages, "user");
  const latestSoma = lastSomaMessage(messages);
  const summary = latestSoma?.execution_summary;
  const action = latestUser?.content?.replace(/^\[BROADCAST\]\s*/i, "").trim();
  const executionShape = compactText(summary?.execution?.shape) ?? compactText(summary?.execution_shape);
  const executionStatus = compactText(summary?.execution?.status) ?? compactText(summary?.execution_status);
  const executionSummary = compactText(summary?.execution?.summary) ?? compactText(summary?.execution_summary);
  const execution = [executionStatus, executionShape, executionSummary].filter(Boolean).join(": ");
  const outputLabels = asItems(summary?.outputs).map(itemText).filter(Boolean) as string[];
  const artifactOutputLabels = artifactLabels(latestSoma?.artifacts);
  const producedLabels = [
    ...outputLabels,
    ...artifactOutputLabels.filter((label) => !outputLabels.includes(label)),
  ];
  const produced = producedLabels.length ? producedLabels : outputs;
  const proofLabels = proofLinks(summary?.proof).map(linkLabel).filter(Boolean) as string[];
  const proof = [
    latestSoma?.run_id ? `Run ${latestSoma.run_id}` : null,
    ...proofLabels,
    auditText(summary?.audit_recovery),
  ].filter(Boolean).join(" | ");
  const next = nextStepText(summary?.next_step)
    ?? (latestSoma ? "Review Soma's response or ask for the next action." : "Tell Soma what you want to accomplish.");

  const facts = [
    {
      label: "Intent",
      value: intentText(summary?.intent) ?? action ?? fallbackAction,
      icon: <Route className="h-3.5 w-3.5" />,
    },
    {
      label: "Understood",
      value: understandingText(summary?.understanding) ?? (latestSoma ? latestSoma.content : "Waiting for Soma to resolve the request."),
      icon: <Sparkles className="h-3.5 w-3.5" />,
    },
    {
      label: "Execution",
      value: execution || "No directed execution proof has been returned yet.",
      icon: <Gauge className="h-3.5 w-3.5" />,
    },
    {
      label: "Used",
      value: capabilityText(summary?.capability_use) ?? teams.join(", "),
      icon: <CheckCircle2 className="h-3.5 w-3.5" />,
    },
    {
      label: "Outputs",
      value: produced.join(", "),
      icon: <FileText className="h-3.5 w-3.5" />,
    },
    {
      label: "Proof",
      value: proof || updated.join(", "),
      icon: proof ? <ShieldCheck className="h-3.5 w-3.5" /> : <RotateCcw className="h-3.5 w-3.5" />,
    },
    {
      label: "Next",
      value: next,
      icon: <Route className="h-3.5 w-3.5" />,
    },
  ];

  return (
    <section className="rounded-2xl border border-cortex-primary/25 bg-cortex-primary/10 p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <Sparkles className="h-4 w-4 text-cortex-primary" />
          <p className="text-sm font-semibold text-cortex-text-main">Soma just did this</p>
        </div>
        <p className="text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
          Causal package
        </p>
      </div>
      <div className="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        {facts.map((fact) => (
          <Fact key={fact.label} label={fact.label} value={fact.value} icon={fact.icon} />
        ))}
      </div>
    </section>
  );
}

function Fact({ label, value, icon }: { label: string; value: string; icon: React.ReactNode }) {
  return (
    <div className="min-w-0 rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2">
      <div className="flex items-center gap-1.5 text-cortex-primary">
        {icon}
        <p className="text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
          {label}
        </p>
      </div>
      <p className="mt-2 text-sm leading-5 text-cortex-text-main">{value}</p>
    </div>
  );
}
