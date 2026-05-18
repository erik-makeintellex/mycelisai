"use client";

import Link from "next/link";
import { useState } from "react";
import {
  Archive,
  Eye,
  Pause,
  Play,
  Radio,
  Search,
  Send,
} from "lucide-react";
import {
  type TeamInteraction,
  type TeamWorkItem,
  type TeamWorkItemState,
} from "@/store/useCortexStore";

const stateStyles: Record<TeamWorkItemState, string> = {
  new: "border-cortex-primary/25 bg-cortex-primary/10 text-cortex-primary",
  briefed: "border-cortex-primary/25 bg-cortex-primary/10 text-cortex-primary",
  queued: "border-cortex-border bg-cortex-bg text-cortex-text-muted",
  running: "border-cortex-success/25 bg-cortex-success/10 text-cortex-success",
  reviewing: "border-cortex-info/25 bg-cortex-info/10 text-cortex-info",
  paused: "border-cortex-border bg-cortex-bg text-cortex-text-muted",
  output_ready:
    "border-cortex-primary/30 bg-cortex-primary/10 text-cortex-primary",
  degraded: "border-amber-400/30 bg-amber-400/10 text-amber-300",
  needs_operator: "border-amber-400/30 bg-amber-400/10 text-amber-300",
  archived: "border-cortex-border bg-cortex-bg text-cortex-text-muted",
};

const actionIcons = {
  inspect: Eye,
  steer: Send,
  start_work: Play,
  pause: Pause,
  resume: Play,
  archive: Archive,
};

export function ActiveWorkLane({
  title = "Active work lane",
  items,
  emptyMessage = "No active work found yet.",
  statusLabel,
  degradedMessage,
  frame = true,
  onAction,
}: {
  title?: string;
  items: TeamWorkItem[];
  emptyMessage?: string;
  statusLabel?: string;
  degradedMessage?: string | null;
  frame?: boolean;
  onAction?: (item: TeamWorkItem, action: TeamInteraction) => void;
}) {
  const className = frame
    ? "rounded-2xl border border-cortex-border bg-cortex-surface p-4"
    : "min-w-0";
  return (
    <section
      className={className}
      data-testid="active-work-lane"
    >
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <Radio className="h-4 w-4 text-cortex-primary" />
          <h2 className="text-sm font-semibold uppercase tracking-[0.16em] text-cortex-text-main">
            {title}
          </h2>
        </div>
        <span className="font-mono text-[11px] text-cortex-text-muted">
          {items.length} item{items.length === 1 ? "" : "s"}
        </span>
      </div>
      {statusLabel || degradedMessage ? (
        <div className="mt-3 rounded-lg border border-cortex-border bg-cortex-bg px-3 py-2 text-xs leading-5 text-cortex-text-muted">
          {statusLabel ? <span className="font-semibold text-cortex-text-main">{statusLabel}</span> : null}
          {degradedMessage ? <span>{statusLabel ? " " : ""}{degradedMessage}</span> : null}
        </div>
      ) : null}
      <div className="mt-3 space-y-2">
        {items.length === 0 ? (
          <p className="rounded-xl border border-cortex-border bg-cortex-bg p-3 text-sm text-cortex-text-muted">
            {emptyMessage}
          </p>
        ) : (
          items.map((item) => (
            <WorkItemRow key={item.id} item={item} onAction={onAction} />
          ))
        )}
      </div>
    </section>
  );
}

function WorkItemRow({
  item,
  onAction,
}: {
  item: TeamWorkItem;
  onAction?: (item: TeamWorkItem, action: TeamInteraction) => void;
}) {
  return (
    <article className="rounded-xl border border-cortex-border bg-cortex-bg p-3">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <span
              className={`rounded-full border px-2 py-0.5 font-mono text-[10px] uppercase tracking-[0.14em] ${stateStyles[item.state]}`}
            >
              {item.state.replace("_", " ")}
            </span>
            {item.sourceLabel ? (
              <span className="rounded-full border border-cortex-border bg-cortex-surface px-2 py-0.5 font-mono text-[10px] uppercase tracking-[0.14em] text-cortex-text-muted">
                {item.sourceLabel}
              </span>
            ) : null}
            <span className="font-mono text-[11px] text-cortex-text-muted">
              {item.scopeLabel}
            </span>
          </div>
          <h3 className="mt-2 truncate text-sm font-semibold text-cortex-text-main">
            {item.title}
          </h3>
          {item.description ? (
            <p className="mt-1 line-clamp-2 text-sm leading-5 text-cortex-text-muted">
              {item.description}
            </p>
          ) : null}
          <p className="mt-2 font-mono text-[11px] text-cortex-text-muted">
            {item.ownerLabel}
            {typeof item.outputCount === "number"
              ? ` | ${item.outputCount} output${item.outputCount === 1 ? "" : "s"}`
              : ""}
          </p>
          {item.nextAction ? (
            <p className="mt-2 text-xs leading-5 text-cortex-text-main">
              Next: {item.nextAction}
            </p>
          ) : null}
          {item.fallbackReason ? (
            <p className="mt-2 text-xs leading-5 text-amber-300">
              {item.fallbackReason}
            </p>
          ) : null}
        </div>
        <div className="grid grid-cols-3 gap-1.5 sm:flex sm:flex-wrap sm:justify-end">
          {item.interactions.map((action) => (
            <ActionControl
              key={action.action}
              item={item}
              action={action}
              onAction={onAction}
            />
          ))}
        </div>
      </div>
      {item.advanced ? <AdvancedProjection item={item} /> : null}
    </article>
  );
}

function ActionControl({
  item,
  action,
  onAction,
}: {
  item: TeamWorkItem;
  action: TeamInteraction;
  onAction?: (item: TeamWorkItem, action: TeamInteraction) => void;
}) {
  const Icon = actionIcons[action.action] ?? Search;
  const hasExecutableTarget = Boolean(action.href) || Boolean(onAction);
  const isDisabled = Boolean(action.disabled) || !hasExecutableTarget;
  const className =
    "inline-flex h-8 min-w-8 items-center justify-center gap-1 rounded-lg border border-cortex-border px-2 text-[11px] font-semibold text-cortex-text-main hover:border-cortex-primary/30 disabled:cursor-not-allowed disabled:opacity-50";
  const title =
    action.disabled && action.disabledReason
      ? `${action.label}: ${action.disabledReason}`
      : !hasExecutableTarget
      ? `${action.label}: action API is not connected yet.`
      : action.label;
  if (action.href && !isDisabled) {
    return (
      <Link href={action.href} className={className} title={title}>
        <Icon className="h-3.5 w-3.5" />
        <span className="hidden sm:inline">{action.label}</span>
      </Link>
    );
  }
  return (
    <button
      type="button"
      className={className}
      disabled={isDisabled}
      title={title}
      onClick={() => onAction?.(item, action)}
    >
      <Icon className="h-3.5 w-3.5" />
      <span className="hidden sm:inline">{action.label}</span>
    </button>
  );
}

function AdvancedProjection({ item }: { item: TeamWorkItem }) {
  const [isOpen, setIsOpen] = useState(false);
  const advanced = item.advanced;
  if (!advanced) return null;
  const rows = [
    ["Inputs", advanced.inputs],
    ["Deliveries", advanced.deliveries],
    ["Models", advanced.modelIds],
    ["Tools", advanced.toolIds],
    ["Capabilities", advanced.capabilityIds],
    ["Expected outputs", advanced.expectedOutputs],
    ["Expected proof", advanced.expectedProof],
    ["Execution shape", advanced.executionShape],
    ["Policy", advanced.policyRef ? [advanced.policyRef] : []],
  ].filter(([, values]) => Array.isArray(values) && values.length > 0) as Array<
    [string, string[]]
  >;

  if (rows.length === 0 && !advanced.promptCount) return null;

  return (
    <details
      className="mt-3 rounded-lg border border-cortex-border bg-cortex-surface px-3 py-2"
      onToggle={(event) => setIsOpen(event.currentTarget.open)}
    >
      <summary className="cursor-pointer font-mono text-[11px] uppercase tracking-[0.14em] text-cortex-text-muted">
        Advanced inspect
      </summary>
      {isOpen ? <div className="mt-3 space-y-2">
        {rows.map(([label, values]) => (
          <div key={label}>
            <p className="font-mono text-[10px] uppercase tracking-[0.14em] text-cortex-primary">
              {label}
            </p>
            <div className="mt-1 flex flex-wrap gap-1">
              {values.map((value) => (
                <span
                  key={`${label}-${value}`}
                  className="max-w-full break-all rounded border border-cortex-border bg-cortex-bg px-2 py-1 font-mono text-[10px] text-cortex-text-muted"
                >
                  {value}
                </span>
              ))}
            </div>
          </div>
        ))}
        {advanced.promptCount ? (
          <p className="font-mono text-[10px] text-cortex-text-muted">
            {advanced.promptCount} prompt
            {advanced.promptCount === 1 ? "" : "s"} available in agent inspect.
          </p>
        ) : null}
      </div> : null}
    </details>
  );
}
