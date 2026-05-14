"use client";

import { Check, CheckCircle2, FileText, Gauge, Quote, RotateCcw, Route, ShieldCheck, Sparkles } from "lucide-react";
import type React from "react";
import { useState } from "react";
import type { ChatArtifactRef, ChatMessage, ExecutionSummaryCapabilityUse, ExecutionSummaryData, ExecutionSummaryItem, ExecutionSummaryLink } from "@/store/useCortexStore";

type SummaryValue = string | ExecutionSummaryItem;
type FactModel = { label: string; value: string; icon: React.ReactNode; quoteValue?: string };

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
  const [copiedFact, setCopiedFact] = useState<string | null>(null);
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
      quoteValue: produced.join(", "),
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

  const primaryFacts = facts.slice(0, 4);
  const secondaryFacts = facts.slice(4);
  const copyFactQuote = async (fact: FactModel) => {
    if (!fact.quoteValue) return;
    await navigator.clipboard.writeText(`> ${fact.quoteValue}`);
    setCopiedFact(fact.label);
    window.setTimeout(() => setCopiedFact((current) => current === fact.label ? null : current), 1200);
  };

  return (
    <section className="rounded-lg border border-cortex-primary/25 bg-cortex-primary/10 p-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <Sparkles className="h-4 w-4 text-cortex-primary" />
          <p className="text-sm font-semibold text-cortex-text-main">Soma just did this</p>
        </div>
        <p className="text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">Trust package</p>
      </div>
      <div className="mt-3 grid gap-2 md:grid-cols-2 xl:grid-cols-4">
        {primaryFacts.map((fact) => (
          <Fact key={fact.label} fact={fact} copied={copiedFact === fact.label} onQuote={() => void copyFactQuote(fact)} />
        ))}
      </div>
      {secondaryFacts.length ? (
        <div className="mt-2 grid gap-2 md:grid-cols-3">
          {secondaryFacts.map((fact) => (
            <Fact key={fact.label} fact={fact} copied={copiedFact === fact.label} compact onQuote={() => void copyFactQuote(fact)} />
          ))}
        </div>
      ) : null}
    </section>
  );
}

function QuoteFactButton({ fact, copied, onQuote }: { fact: FactModel; copied: boolean; onQuote: () => void }) {
  if (!fact.quoteValue) return null;
  return (
    <button
      type="button"
      onClick={onQuote}
      className="inline-flex h-5 w-5 shrink-0 items-center justify-center rounded border border-cortex-border/70 text-cortex-text-muted transition-colors hover:border-cortex-info/40 hover:bg-cortex-info/10 hover:text-cortex-info"
      title={copied ? "Copied output quote" : "Copy output quote"}
      aria-label={copied ? "Copied output quote" : `Copy output quote for ${fact.value}`}
    >
      {copied ? <Check className="h-3 w-3" /> : <Quote className="h-3 w-3" />}
    </button>
  );
}

function Fact({
  fact,
  copied,
  onQuote,
  compact = false,
}: {
  fact: FactModel;
  copied: boolean;
  onQuote: () => void;
  compact?: boolean;
}) {
  const shellClass = compact
    ? "min-w-0 rounded-lg border border-cortex-border bg-cortex-bg/80 px-3 py-2"
    : "min-h-[82px] min-w-0 overflow-hidden rounded-lg border border-cortex-border bg-cortex-bg px-3 py-2";
  const textClass = compact
    ? "mt-1 overflow-hidden text-xs leading-4 text-cortex-text-main [display:-webkit-box] [-webkit-box-orient:vertical] [-webkit-line-clamp:1]"
    : "mt-2 overflow-hidden text-sm leading-5 text-cortex-text-main [display:-webkit-box] [-webkit-box-orient:vertical] [-webkit-line-clamp:2]";

  return (
    <div className={shellClass}>
      <div className="flex items-center justify-between gap-2 text-cortex-primary">
        <div className="flex min-w-0 items-center gap-1.5">
          {fact.icon}
          <p className="text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">{fact.label}</p>
        </div>
        <QuoteFactButton fact={fact} copied={copied} onQuote={onQuote} />
      </div>
      <p
        className={textClass}
        title={fact.value}
      >
        {fact.value}
      </p>
    </div>
  );
}
