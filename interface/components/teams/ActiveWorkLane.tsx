"use client";

import Link from "next/link";
import { Archive, Eye, Pause, Play, Radio, RefreshCw, Search, Send } from "lucide-react";
import {
  type TeamInteraction,
  type TeamWorkItem,
  type TeamWorkItemState,
} from "@/store/useCortexStore";
import { ActiveWorkAdvancedProjection } from "./ActiveWorkAdvancedProjection";
import { ActiveWorkEvidence } from "./ActiveWorkEvidence";
import {
  compactActions,
  compactDescription,
  compactNextAction,
  compactTitle,
  isStaleFailedPlanItem,
} from "./activeWorkCompact";
import { ReviewDecisionGuide } from "./ReviewDecisionGuide";
import { TeamAskForm } from "./TeamAskForm";
import { WorkTruthSummary } from "./WorkTruthSummary";

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

const stateLabels: Record<TeamWorkItemState, string> = { new: "Ready to brief", briefed: "Ready to start", queued: "Queued", running: "In progress", reviewing: "In review", paused: "Paused", output_ready: "Output ready", degraded: "Degraded", needs_operator: "Needs operator", archived: "Archived" };

const actionIcons = { inspect: Eye, steer: Send, start_work: Play, pause: Pause, resume: Play, recover: RefreshCw, archive: Archive };

export function ActiveWorkLane({
  title = "Active work lane",
  items,
  emptyMessage = "No active work found yet.",
  statusLabel,
  degradedMessage,
  frame = true,
  maxVisibleItems,
  totalItemCount,
  moreItemsHref = "/teams",
  onAction,
  onTeamAsk,
}: {
  title?: string;
  items: TeamWorkItem[];
  emptyMessage?: string;
  statusLabel?: string;
  degradedMessage?: string | null;
  frame?: boolean;
  maxVisibleItems?: number;
  totalItemCount?: number;
  moreItemsHref?: string;
  onAction?: (item: TeamWorkItem, action: TeamInteraction) => void;
  onTeamAsk?: (item: TeamWorkItem, message: string) => Promise<void> | void;
}) {
  const visibleItems =
    typeof maxVisibleItems === "number" && maxVisibleItems > 0
      ? items.slice(0, maxVisibleItems)
      : items;
  const shownCount = visibleItems.length;
  const count = totalItemCount ?? items.length;
  const hiddenCount = Math.max(count - shownCount, 0);
  const compact = !frame;
  const className = frame
    ? "rounded-2xl border border-cortex-border bg-cortex-surface p-4"
    : "min-w-0";
  return (
    <section className={className} data-testid="active-work-lane">
      {compact ? null : <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <Radio className="h-4 w-4 text-cortex-primary" />
          <h2 className="text-sm font-semibold uppercase tracking-[0.16em] text-cortex-text-main">
            {title}
          </h2>
        </div>
        <span className="font-mono text-[11px] text-cortex-text-muted">
          {count} item{count === 1 ? "" : "s"}
        </span>
      </div>}
      {!compact && (statusLabel || degradedMessage) ? (
        <div className="mt-3 rounded-lg border border-cortex-border bg-cortex-bg px-3 py-2 text-xs leading-5 text-cortex-text-muted">
          {statusLabel ? <span className="font-semibold text-cortex-text-main">{statusLabel}</span> : null}
          {degradedMessage ? <span>{statusLabel ? " " : ""}{degradedMessage}</span> : null}
        </div>
      ) : null}
      <div className={compact ? "space-y-2" : "mt-3 space-y-2"}>
        {items.length === 0 ? (
          <p className="rounded-xl border border-cortex-border bg-cortex-bg p-3 text-sm text-cortex-text-muted">
            {emptyMessage}
          </p>
        ) : (
          visibleItems.map((item) => (
            <WorkItemRow key={item.id} item={item} compact={compact} onAction={onAction} onTeamAsk={onTeamAsk} />
          ))
        )}
        {hiddenCount > 0 ? (
          <Link
            href={moreItemsHref}
            className="flex items-center justify-between rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-muted hover:border-cortex-primary/35 hover:text-cortex-text-main"
          >
            <span>{hiddenCount} more work item{hiddenCount === 1 ? "" : "s"} in Teams</span>
            <span className="font-mono text-[10px] uppercase tracking-[0.14em]">
              Open backlog
            </span>
          </Link>
        ) : null}
      </div>
    </section>
  );
}

function WorkItemRow({
  item,
  compact = false,
  onAction,
  onTeamAsk,
}: {
  item: TeamWorkItem;
  compact?: boolean;
  onAction?: (item: TeamWorkItem, action: TeamInteraction) => void;
  onTeamAsk?: (item: TeamWorkItem, message: string) => Promise<void> | void;
}) {
  const visibleActions = compact || isStaleFailedPlanItem(item)
    ? compactActions(item)
    : item.interactions;
  const usePlainReviewCopy = compact || isStaleFailedPlanItem(item);
  const title = usePlainReviewCopy ? compactTitle(item) : item.title;
  const description = usePlainReviewCopy ? compactDescription(item) : item.description;
  const nextAction = usePlainReviewCopy ? compactNextAction(item) : item.nextAction;
  return (
    <article className="rounded-xl border border-cortex-border bg-cortex-bg p-3">
      <div className={compact ? "space-y-3" : "flex flex-wrap items-start justify-between gap-3"}>
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <span
              className={`rounded-full border px-2 py-0.5 font-mono text-[10px] uppercase tracking-[0.14em] ${stateStyles[item.state]}`}
            >
              {stateLabels[item.state]}
            </span>
            {!compact && item.sourceLabel ? (
              <span className="rounded-full border border-cortex-border bg-cortex-surface px-2 py-0.5 font-mono text-[10px] uppercase tracking-[0.14em] text-cortex-text-muted">
                {item.sourceLabel}
              </span>
            ) : null}
            {!compact ? <span className="font-mono text-[11px] text-cortex-text-muted">
              {item.scopeLabel}
            </span> : null}
          </div>
          <h3 className={compact ? "mt-2 text-base font-semibold leading-6 text-cortex-text-main" : "mt-2 truncate text-sm font-semibold text-cortex-text-main"}>
            {title}
          </h3>
          {description ? (
            <p className={compact ? "mt-1 text-sm leading-6 text-cortex-text-muted" : "mt-1 line-clamp-2 text-sm leading-5 text-cortex-text-muted"}>
              {description}
            </p>
          ) : null}
          {!compact ? <p className="mt-2 font-mono text-[11px] text-cortex-text-muted">
            {item.ownerLabel}
            {typeof item.outputCount === "number"
              ? ` | ${item.outputCount} output${item.outputCount === 1 ? "" : "s"}`
              : ""}
          </p> : null}
          {nextAction ? (
            <p className={compact ? "mt-3 rounded-lg border border-cortex-primary/20 bg-cortex-primary/10 px-3 py-2 text-sm leading-6 text-cortex-text-main" : "mt-2 text-xs leading-5 text-cortex-text-main"}>
              <span className="font-semibold">Next:</span> {nextAction}
            </p>
          ) : null}
          {!compact && item.fallbackReason ? (
            <p className="mt-2 text-xs leading-5 text-amber-300">
              {item.fallbackReason}
            </p>
          ) : null}
          <WorkTruthSummary item={item} compact={compact} />
          {compact ? null : <ReviewDecisionGuide item={item} />}
          <ActiveWorkEvidence item={item} />
          {compact || isStaleFailedPlanItem(item) ? null : (
            <TeamAskForm item={item} onTeamAsk={onTeamAsk} />
          )}
        </div>
        <div className={compact ? "flex flex-wrap gap-2" : "grid grid-cols-3 gap-1.5 sm:flex sm:flex-wrap sm:justify-end"}>
          {visibleActions.map((action) => (
            <ActionControl
              key={action.action}
              item={item}
              action={action}
              onAction={onAction}
              compact={compact}
            />
          ))}
        </div>
      </div>
      {item.advanced ? <ActiveWorkAdvancedProjection item={item} /> : null}
    </article>
  );
}

function ActionControl({
  item,
  action,
  compact = false,
  onAction,
}: {
  item: TeamWorkItem;
  action: TeamInteraction;
  compact?: boolean;
  onAction?: (item: TeamWorkItem, action: TeamInteraction) => void;
}) {
  const Icon = actionIcons[action.action] ?? Search;
  const hasExecutableTarget = Boolean(action.href) || Boolean(onAction);
  const isDisabled = Boolean(action.disabled) || !hasExecutableTarget;
  const isPrimary = compact && (
    action.action === "recover" ||
    action.action === "archive" ||
    action.action === "inspect"
  );
  const className = isPrimary
    ? "inline-flex h-9 min-w-9 items-center justify-center gap-1.5 rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-3 text-xs font-semibold text-cortex-primary hover:border-cortex-primary/60 disabled:cursor-not-allowed disabled:opacity-50"
    : "inline-flex h-8 min-w-8 items-center justify-center gap-1 rounded-lg border border-cortex-border px-2 text-[11px] font-semibold text-cortex-text-main hover:border-cortex-primary/30 disabled:cursor-not-allowed disabled:opacity-50";
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
        <span className={compact ? "" : "hidden sm:inline"}>{action.label}</span>
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
      <span className={compact ? "" : "hidden sm:inline"}>{action.label}</span>
    </button>
  );
}
