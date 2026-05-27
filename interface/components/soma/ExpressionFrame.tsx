"use client";

import type { ReactNode } from "react";
import {
  AlertTriangle,
  CheckCircle2,
  ChevronDown,
  CircleHelp,
  FileText,
  Gauge,
  LockKeyhole,
  Radio,
  RotateCcw,
  Route,
  ShieldCheck,
  Sparkles,
} from "lucide-react";

export type ExpressionFrameKind =
  | "direct_answer"
  | "proposal"
  | "active_work"
  | "output_ready"
  | "proof"
  | "degraded"
  | "blocked"
  | "recovery";

export type ExpressionFrameTone = "neutral" | "info" | "success" | "warning" | "danger";

export type ExpressionFrameAction = {
  label: string;
  href?: string;
  onClick?: () => void;
  disabled?: boolean;
};

export type ExpressionFrameReference = {
  label: string;
  detail?: string;
  href?: string;
};

export type ExpressionFrameInspectItem = {
  label: string;
  value: ReactNode;
};

export type ExpressionFrameProps = {
  kind: ExpressionFrameKind;
  title?: string;
  intent: ReactNode;
  state?: ReactNode;
  nextAction?: ReactNode;
  risk?: ReactNode;
  approval?: ReactNode;
  outputs?: ExpressionFrameReference[];
  proof?: ExpressionFrameReference[];
  recovery?: ReactNode;
  inspect?: ExpressionFrameInspectItem[];
  primaryAction?: ExpressionFrameAction;
  secondaryAction?: ExpressionFrameAction;
  className?: string;
};

const KIND_COPY: Record<ExpressionFrameKind, { label: string; tone: ExpressionFrameTone; icon: ReactNode }> = {
  direct_answer: { label: "Direct answer", tone: "info", icon: <Sparkles className="h-4 w-4" /> },
  proposal: { label: "Proposal", tone: "warning", icon: <CircleHelp className="h-4 w-4" /> },
  active_work: { label: "Active work", tone: "info", icon: <Radio className="h-4 w-4" /> },
  output_ready: { label: "Output ready", tone: "success", icon: <FileText className="h-4 w-4" /> },
  proof: { label: "Proof", tone: "success", icon: <ShieldCheck className="h-4 w-4" /> },
  degraded: { label: "Degraded", tone: "danger", icon: <AlertTriangle className="h-4 w-4" /> },
  blocked: { label: "Blocked", tone: "danger", icon: <LockKeyhole className="h-4 w-4" /> },
  recovery: { label: "Recovery", tone: "warning", icon: <RotateCcw className="h-4 w-4" /> },
};

function toneClass(tone: ExpressionFrameTone) {
  if (tone === "success") return "border-cortex-success/25 bg-cortex-success/10 text-cortex-success";
  if (tone === "warning") return "border-amber-400/25 bg-amber-400/10 text-amber-300";
  if (tone === "danger") return "border-red-400/30 bg-red-400/10 text-red-300";
  if (tone === "info") return "border-cortex-info/25 bg-cortex-info/10 text-cortex-info";
  return "border-cortex-border bg-cortex-surface/70 text-cortex-text-muted";
}

function FrameFact({
  icon,
  label,
  children,
}: {
  icon: ReactNode;
  label: string;
  children: ReactNode;
}) {
  return (
    <div className="grid grid-cols-[16px_82px_minmax(0,1fr)] items-start gap-2">
      <div className="mt-0.5 text-cortex-info">{icon}</div>
      <div className="text-[9px] font-mono font-bold uppercase tracking-widest text-cortex-text-muted">
        {label}
      </div>
      <div className="min-w-0 text-[11px] leading-5 text-cortex-text-main">{children}</div>
    </div>
  );
}

function ReferenceList({ items }: { items: ExpressionFrameReference[] }) {
  return (
    <div className="flex flex-wrap gap-1.5">
      {items.map((item) => {
        const content = (
          <span
            className="inline-flex max-w-full items-center gap-1 rounded border border-cortex-info/20 bg-cortex-info/10 px-1.5 py-0.5 text-[9px] font-mono text-cortex-info"
            title={item.detail ?? item.label}
          >
            <span className="truncate">{item.label}</span>
          </span>
        );
        return item.href ? (
          <a key={`${item.label}-${item.href}`} href={item.href} className="min-w-0 hover:underline">
            {content}
          </a>
        ) : (
          <span key={item.label} className="min-w-0">
            {content}
          </span>
        );
      })}
    </div>
  );
}

function ActionButton({ action, primary = false }: { action: ExpressionFrameAction; primary?: boolean }) {
  const className = primary
    ? "inline-flex items-center justify-center rounded border border-cortex-primary/30 bg-cortex-primary/15 px-2.5 py-1.5 text-xs font-mono font-bold uppercase text-cortex-primary transition-colors hover:bg-cortex-primary/20 disabled:cursor-not-allowed disabled:opacity-50"
    : "inline-flex items-center justify-center rounded border border-cortex-border bg-cortex-surface px-2.5 py-1.5 text-xs font-mono font-bold uppercase text-cortex-text-muted transition-colors hover:text-cortex-text-main disabled:cursor-not-allowed disabled:opacity-50";

  if (action.href && !action.disabled) {
    return (
      <a href={action.href} className={className}>
        {action.label}
      </a>
    );
  }

  return (
    <button type="button" className={className} disabled={action.disabled} onClick={action.onClick}>
      {action.label}
    </button>
  );
}

export function ExpressionFrame({
  kind,
  title,
  intent,
  state,
  nextAction,
  risk,
  approval,
  outputs = [],
  proof = [],
  recovery,
  inspect = [],
  primaryAction,
  secondaryAction,
  className = "",
}: ExpressionFrameProps) {
  const metadata = KIND_COPY[kind];
  const hasInspect = inspect.length > 0;

  return (
    <section
      className={`rounded-lg border border-cortex-info/20 bg-cortex-info/5 px-3 py-2.5 shadow-sm ${className}`}
      data-expression-frame-kind={kind}
    >
      <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
        <div className="flex min-w-0 items-center gap-1.5 text-[9px] font-mono font-bold uppercase tracking-widest text-cortex-info">
          <span className="shrink-0">{metadata.icon}</span>
          <span className="truncate">{title ?? metadata.label}</span>
        </div>
        <span className={`rounded border px-1.5 py-0.5 text-[9px] font-mono font-bold uppercase ${toneClass(metadata.tone)}`}>
          {metadata.label}
        </span>
      </div>

      <div className="space-y-2">
        <FrameFact icon={<Route className="h-3.5 w-3.5" />} label="Intent">
          {intent}
        </FrameFact>
        {state ? (
          <FrameFact icon={<Gauge className="h-3.5 w-3.5" />} label="State">
            {state}
          </FrameFact>
        ) : null}
        {nextAction ? (
          <FrameFact icon={<CheckCircle2 className="h-3.5 w-3.5" />} label="Next">
            {nextAction}
          </FrameFact>
        ) : null}
        {(risk || approval) ? (
          <FrameFact icon={<ShieldCheck className="h-3.5 w-3.5" />} label="Approval">
            <div className="space-y-1">
              {risk ? <div>{risk}</div> : null}
              {approval ? <div>{approval}</div> : null}
            </div>
          </FrameFact>
        ) : null}
        {outputs.length > 0 ? (
          <FrameFact icon={<FileText className="h-3.5 w-3.5" />} label="Outputs">
            <ReferenceList items={outputs} />
          </FrameFact>
        ) : null}
        {proof.length > 0 ? (
          <FrameFact icon={<ShieldCheck className="h-3.5 w-3.5" />} label="Proof">
            <ReferenceList items={proof} />
          </FrameFact>
        ) : null}
        {recovery ? (
          <FrameFact icon={<RotateCcw className="h-3.5 w-3.5" />} label="Recovery">
            {recovery}
          </FrameFact>
        ) : null}
      </div>

      {(primaryAction || secondaryAction || hasInspect) ? (
        <div className="mt-3 flex flex-wrap items-center gap-2">
          {primaryAction ? <ActionButton action={primaryAction} primary /> : null}
          {secondaryAction ? <ActionButton action={secondaryAction} /> : null}
          {hasInspect ? (
            <details className="group basis-full">
              <summary className="inline-flex cursor-pointer list-none items-center gap-1.5 text-xs font-mono text-cortex-primary hover:text-cortex-primary/80">
                <ChevronDown className="h-3.5 w-3.5 transition-transform group-open:rotate-180" />
                Inspect
              </summary>
              <div className="mt-2 grid gap-2 rounded border border-cortex-border/60 bg-cortex-surface/60 p-2 md:grid-cols-2">
                {inspect.map((item) => (
                  <div key={item.label} className="min-w-0">
                    <div className="text-[9px] font-mono font-bold uppercase tracking-widest text-cortex-text-muted">
                      {item.label}
                    </div>
                    <div className="mt-1 text-[11px] leading-5 text-cortex-text-main">{item.value}</div>
                  </div>
                ))}
              </div>
            </details>
          ) : null}
        </div>
      ) : null}
    </section>
  );
}
