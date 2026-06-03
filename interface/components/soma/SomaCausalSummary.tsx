"use client";

import { CheckCircle2, ChevronDown, ChevronUp, FileText, Gauge, RotateCcw, Route, ShieldCheck, Sparkles } from "lucide-react";
import { useState } from "react";
import type { ChatMessage } from "@/store/useCortexStore";
import { CompactFact, Fact, type FactModel } from "./SomaCausalSummaryFact";
import {
  artifactLabels,
  asItems,
  auditText,
  capabilityText,
  compactText,
  executionShapeText,
  intentText,
  itemText,
  lastMessage,
  lastSomaMessage,
  linkLabel,
  nextStepText,
  proofLinks,
  understandingText,
} from "./SomaCausalSummaryModel";
import { defaultResponseState, responseStateToneClass } from "./SomaCausalSummaryState";

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
  const [detailsOpen, setDetailsOpen] = useState(false);
  const latestUser = lastMessage(messages, "user");
  const latestSoma = lastSomaMessage(messages);
  const summary = latestSoma?.execution_summary;
  const responseState = defaultResponseState(latestSoma);
  const action = latestUser?.content?.replace(/^\[BROADCAST\]\s*/i, "").trim();
  const executionShape = executionShapeText(summary?.execution?.shape) ?? executionShapeText(summary?.execution_shape);
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
      label: "Request",
      value: intentText(summary?.intent) ?? action ?? fallbackAction,
      icon: <Route className="h-3.5 w-3.5" />,
    },
    {
      label: "Understood",
      value: understandingText(summary?.understanding) ?? (latestSoma ? latestSoma.content : "Waiting for Soma to resolve the request."),
      icon: <Sparkles className="h-3.5 w-3.5" />,
    },
    {
      label: "Progress",
      value: execution || "No run result has been returned yet.",
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
      label: "Evidence",
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
  const outputFact = facts.find((fact) => fact.label === "Outputs");
  const proofFact = facts.find((fact) => fact.label === "Evidence");
  const nextFact = facts.find((fact) => fact.label === "Next");
  const outcome = executionSummary ?? latestSoma?.content ?? "Soma is ready for your next request.";
  const copyFactQuote = async (fact: FactModel) => {
    if (!fact.quoteValue) return;
    await navigator.clipboard.writeText(`> ${fact.quoteValue}`);
    setCopiedFact(fact.label);
    window.setTimeout(() => setCopiedFact((current) => current === fact.label ? null : current), 1200);
  };

  if (!latestSoma) {
    return (
      <section className="rounded-lg border border-cortex-border bg-cortex-surface/70 p-3">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <Sparkles className="h-4 w-4 text-cortex-primary" />
            <p className="text-sm font-semibold text-cortex-text-main">Ready for your first request</p>
          </div>
          <span className={`rounded border px-1.5 py-0.5 text-[9px] font-mono font-bold uppercase ${responseStateToneClass(responseState.tone)}`}>
            {responseState.label}
          </span>
        </div>
        <p className="mt-2 text-sm leading-6 text-cortex-text-main">Tell Soma what you want to plan, review, create, or run.</p>
      </section>
    );
  }

  return (
    <section className="rounded-lg border border-cortex-primary/25 bg-cortex-primary/10 p-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <Sparkles className="h-4 w-4 text-cortex-primary" />
          <p className="text-sm font-semibold text-cortex-text-main">Soma just did this</p>
        </div>
        <div className="flex flex-wrap items-center gap-1.5">
          <span className={`rounded border px-1.5 py-0.5 text-[9px] font-mono font-bold uppercase ${responseStateToneClass(responseState.tone)}`}>
            {responseState.label ?? responseState.kind.replace(/_/g, " ")}
          </span>
          <p className="text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">Review details</p>
        </div>
      </div>
      <p className="mt-2 text-sm leading-6 text-cortex-text-main">{outcome}</p>
      <div className="mt-3 grid gap-2 md:grid-cols-3">
        {outputFact ? (
          <CompactFact fact={outputFact} copied={copiedFact === outputFact.label} onQuote={() => void copyFactQuote(outputFact)} />
        ) : null}
        {proofFact ? (
          <CompactFact fact={proofFact} copied={copiedFact === proofFact.label} onQuote={() => void copyFactQuote(proofFact)} />
        ) : null}
        {nextFact ? (
          <CompactFact fact={nextFact} copied={copiedFact === nextFact.label} onQuote={() => void copyFactQuote(nextFact)} />
        ) : null}
      </div>
      <button
        type="button"
        onClick={() => setDetailsOpen((open) => !open)}
        className="mt-3 inline-flex items-center gap-1.5 text-xs font-mono text-cortex-primary hover:text-cortex-primary/80 transition-colors"
        aria-expanded={detailsOpen}
      >
        {detailsOpen ? <ChevronUp className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
        {detailsOpen ? "Hide details" : "Show details"}
      </button>
      {detailsOpen ? (
        <div className="mt-3 grid gap-2 md:grid-cols-2 xl:grid-cols-4">
          {primaryFacts.map((fact) => (
            <Fact key={fact.label} fact={fact} copied={copiedFact === fact.label} onQuote={() => void copyFactQuote(fact)} />
          ))}
          {secondaryFacts.map((fact) => (
            <Fact key={fact.label} fact={fact} copied={copiedFact === fact.label} compact onQuote={() => void copyFactQuote(fact)} />
          ))}
        </div>
      ) : null}
    </section>
  );
}
