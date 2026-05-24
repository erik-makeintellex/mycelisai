"use client";

import type React from "react";
import { Boxes, FileText, Radio, ShieldCheck, Sparkles } from "lucide-react";

function SlotPanel({
  icon,
  label,
  description,
  children,
  className = "",
}: {
  icon: React.ReactNode;
  label: string;
  description?: string;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <section className={`min-w-0 rounded-2xl border border-cortex-border bg-cortex-bg p-3 ${className}`}>
      <div className="mb-2 flex items-center gap-2 text-[10px] font-mono font-bold uppercase tracking-[0.16em] text-cortex-text-muted">
        <span className="text-cortex-primary">{icon}</span>
        {label}
      </div>
      {description ? (
        <p className="mb-2 text-[11px] leading-5 text-cortex-text-muted">{description}</p>
      ) : null}
      {children}
    </section>
  );
}

export function SomaWorkspaceFrame({
  expression,
  activeWork,
  output,
  trust,
  context,
}: {
  expression: React.ReactNode;
  activeWork?: React.ReactNode;
  output?: React.ReactNode;
  trust?: React.ReactNode;
  context?: React.ReactNode;
}) {
  return (
    <div className="grid gap-4 xl:grid-cols-[minmax(0,1.45fr)_minmax(340px,0.85fr)]" data-testid="soma-workspace-frame">
      <div className="min-w-0 space-y-4">
        <SlotPanel
          icon={<Sparkles className="h-3.5 w-3.5" />}
          label="Expression"
          description="Intent, output shape, constraints, and proof before topology."
          className="p-2"
        >
          {expression}
        </SlotPanel>
        {activeWork ? (
          <SlotPanel
            icon={<Radio className="h-3.5 w-3.5" />}
            label="Active work"
            description="Current work, operator attention, and queued follow-up."
          >
            {activeWork}
          </SlotPanel>
        ) : null}
      </div>
      <div className="min-w-0 space-y-4">
        {trust ? (
          <SlotPanel
            icon={<ShieldCheck className="h-3.5 w-3.5" />}
            label="Trust"
            description="Trusted evidence, failure state, and safe next action."
          >
            {trust}
          </SlotPanel>
        ) : null}
        {output ? (
          <SlotPanel
            icon={<FileText className="h-3.5 w-3.5" />}
            label="Output"
            description="Retained files, packages, and reviewable results."
          >
            {output}
          </SlotPanel>
        ) : null}
        {context ? (
          <SlotPanel
            icon={<Boxes className="h-3.5 w-3.5" />}
            label="Context"
            description="Setup evidence, tools, memory, and activity links."
          >
            {context}
          </SlotPanel>
        ) : null}
      </div>
    </div>
  );
}
