"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { Archive, Eye, Pause, Play, RefreshCw, Send, ShieldCheck } from "lucide-react";
import type { TeamInteraction, TeamOutputRef, TeamWorkItem } from "@/store/useCortexStore";
import { outcomeHealthFromRunStatus } from "@/lib/outcomeHealth";
import { OutcomeHealthBadge } from "@/components/shared/OutcomeHealthBadge";
import { ActiveWorkEvidence } from "./ActiveWorkEvidence";
import {
  compactDescription,
  compactNextAction,
  compactTitle,
  isStaleFailedPlanItem,
  needsExternalMutationVerification,
} from "./activeWorkCompact";
import { ExternalOutcomeVerificationForm } from "./ExternalOutcomeVerificationForm";
import type { ExternalOutcomeVerification } from "./teamWorkActions";
import { ReviewDecisionGuide } from "./ReviewDecisionGuide";
import { ReviewQueueSummary } from "./ReviewQueueSummary";
import { TeamAskForm } from "./TeamAskForm";
import { WorkTruthSummary } from "./WorkTruthSummary";

const actionIcons = { inspect: Eye, steer: Send, verify_external_outcome: ShieldCheck, start_work: Play, pause: Pause, resume: Play, recover: RefreshCw, archive: Archive };

export function WorkReviewInbox({
  title = "Work to review",
  items,
  emptyMessage,
  statusLabel,
  degradedMessage,
  onAction,
  onVerifyExternalOutcome,
  onTeamAsk,
}: {
  title?: string;
  items: TeamWorkItem[];
  emptyMessage: string;
  statusLabel?: string;
  degradedMessage?: string | null;
  onAction?: (item: TeamWorkItem, action: TeamInteraction) => void;
  onVerifyExternalOutcome?: (
    item: TeamWorkItem,
    verification: ExternalOutcomeVerification,
  ) => Promise<void> | void;
  onTeamAsk?: (item: TeamWorkItem, message: string) => Promise<void> | void;
}) {
  const [selectedId, setSelectedId] = useState(items[0]?.id ?? null);
  const selectedItem = useMemo(() => items.find((item) => item.id === selectedId) ?? items[0] ?? null, [items, selectedId]);

  if (items.length === 0) {
    return (
      <section className="rounded-2xl border border-cortex-border bg-cortex-surface p-4">
        <p className="rounded-xl border border-cortex-border bg-cortex-bg p-3 text-sm text-cortex-text-muted">
          {emptyMessage}
        </p>
      </section>
    );
  }

  return (
    <section className="rounded-2xl border border-cortex-border bg-cortex-surface p-4" data-testid="work-review-inbox">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="font-mono text-sm font-bold uppercase tracking-[0.16em] text-cortex-text-main">
            {title}
          </h2>
          <p className="mt-1 max-w-3xl text-sm leading-6 text-cortex-text-muted">
            Select one item, read the trusted state, then take the one next action that moves it forward.
          </p>
        </div>
        <span className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1 font-mono text-[11px] text-cortex-text-muted">
          {items.length} item{items.length === 1 ? "" : "s"}
        </span>
      </div>

      <ReviewQueueSummary items={items} />

      {statusLabel || degradedMessage ? (
        <div className="mt-3 rounded-lg border border-cortex-border bg-cortex-bg px-3 py-2 text-xs leading-5 text-cortex-text-muted">
          {statusLabel ? <span className="font-semibold text-cortex-text-main">{statusLabel}</span> : null}
          {degradedMessage ? <span>{statusLabel ? " " : ""}{degradedMessage}</span> : null}
        </div>
      ) : null}

      <div
        className="mt-4 grid gap-3 lg:h-[min(32rem,calc(100vh-15rem))] lg:min-h-[22rem] lg:grid-cols-[minmax(18rem,0.8fr)_minmax(0,1.2fr)]"
        data-testid="work-review-monitor"
      >
        <div className="min-h-0 rounded-xl border border-cortex-border bg-cortex-bg p-2 lg:h-full">
          <div className="max-h-[24rem] space-y-2 overflow-y-auto pr-1 lg:max-h-full" role="list" aria-label="Review work items">
            {items.map((item) => (
              <ReviewListRow
                key={item.id}
                item={item}
                selected={selectedItem?.id === item.id}
                onSelect={() => setSelectedId(item.id)}
                onAction={onAction}
              />
            ))}
          </div>
        </div>

        {selectedItem ? (
          <ReviewDetailPane
            item={selectedItem}
            onAction={onAction}
            onVerifyExternalOutcome={onVerifyExternalOutcome}
            onTeamAsk={onTeamAsk}
          />
        ) : null}
      </div>
    </section>
  );
}

function ReviewListRow({
  item,
  selected,
  onSelect,
  onAction,
}: {
  item: TeamWorkItem;
  selected: boolean;
  onSelect: () => void;
  onAction?: (item: TeamWorkItem, action: TeamInteraction) => void;
}) {
  const primary = primaryReviewAction(item);
  return (
    <article
      role="listitem"
      data-work-item-id={item.id}
      className={`rounded-lg border p-3 transition-colors ${
        selected ? "border-cortex-primary/45 bg-cortex-primary/10" : "border-cortex-border bg-cortex-surface"
      }`}
    >
      <button type="button" onClick={onSelect} className="block w-full text-left">
        <div className="flex flex-wrap items-center gap-2">
          <OutcomeHealthBadge health={item.outcomeHealth ?? outcomeHealthFromRunStatus(item.state)} />
          <span className="font-mono text-[10px] uppercase tracking-[0.14em] text-cortex-text-muted">
            {item.scopeLabel}
          </span>
        </div>
        <h3 className="mt-2 line-clamp-2 text-sm font-semibold leading-5 text-cortex-text-main">
          {compactTitle(item)}
        </h3>
        <p className="mt-1 line-clamp-2 text-xs leading-5 text-cortex-text-muted">
          {compactDescription(item)}
        </p>
      </button>
      {primary ? (
        <div className="mt-3">
          <ActionControl item={item} action={primary} onAction={onAction} primary />
        </div>
      ) : null}
    </article>
  );
}

function ReviewDetailPane({
  item,
  onAction,
  onVerifyExternalOutcome,
  onTeamAsk,
}: {
  item: TeamWorkItem;
  onAction?: (item: TeamWorkItem, action: TeamInteraction) => void;
  onVerifyExternalOutcome?: (
    item: TeamWorkItem,
    verification: ExternalOutcomeVerification,
  ) => Promise<void> | void;
  onTeamAsk?: (item: TeamWorkItem, message: string) => Promise<void> | void;
}) {
  const needsExternalVerification = needsExternalMutationVerification(item);
  const primary = primaryReviewAction(item);
  const retainedCandidate = item.state === "degraded" ? firstOutputAction(item.outputRefs, "Open unverified output") : null;
  const secondaryActions = needsExternalVerification ? [] : [
    ...(retainedCandidate && retainedCandidate.href !== primary?.href ? [retainedCandidate] : []),
    ...item.interactions.filter((action) => !action.disabled && action.action !== primary?.action),
  ];
  return (
    <article className="min-w-0 rounded-xl border border-cortex-border bg-cortex-bg p-4 lg:h-full lg:overflow-y-auto" aria-label={`Review details for ${item.title}`}>
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="font-mono text-[10px] font-bold uppercase tracking-[0.14em] text-cortex-primary">
            Selected work
          </p>
          <h3 className="mt-2 text-lg font-semibold leading-7 text-cortex-text-main">
            {compactTitle(item)}
          </h3>
          <p className="mt-1 text-sm leading-6 text-cortex-text-muted">
            {compactDescription(item)}
          </p>
        </div>
        {needsExternalVerification ? null : <PrimaryDetailAction item={item} onAction={onAction} />}
      </div>

      {compactNextAction(item) ? (
        <div className="mt-4 rounded-lg border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-2 text-sm leading-6 text-cortex-text-main">
          <span className="font-semibold">Next:</span> {compactNextAction(item)}
        </div>
      ) : null}

      <WorkTruthSummary item={item} />
      <ReviewDecisionGuide item={item} concise />
      <ActiveWorkEvidence item={item} />

      {needsExternalVerification ? (
        <ExternalOutcomeVerificationForm item={item} onVerify={onVerifyExternalOutcome} />
      ) : null}

      {item.state === "needs_operator" ? (
        <TeamAskForm item={item} onTeamAsk={onTeamAsk} />
      ) : null}

      {secondaryActions.length > 0 ? (
        <div className="mt-4 border-t border-cortex-border pt-3">
          <p className="font-mono text-[10px] uppercase tracking-[0.14em] text-cortex-text-muted">
            Other available actions
          </p>
          <div className="mt-2 flex flex-wrap gap-2">
            {secondaryActions.map((action) => (
              <ActionControl key={action.action} item={item} action={action} onAction={onAction} />
            ))}
          </div>
        </div>
      ) : null}
    </article>
  );
}

function PrimaryDetailAction({ item, onAction }: { item: TeamWorkItem; onAction?: (item: TeamWorkItem, action: TeamInteraction) => void }) {
  const action = primaryReviewAction(item);
  return action ? <ActionControl item={item} action={action} onAction={onAction} primary /> : null;
}

function ActionControl({
  item,
  action,
  primary = false,
  onAction,
}: {
  item: TeamWorkItem;
  action: TeamInteraction;
  primary?: boolean;
  onAction?: (item: TeamWorkItem, action: TeamInteraction) => void;
}) {
  const Icon = actionIcons[action.action] ?? Eye;
  const hasTarget = Boolean(action.href) || Boolean(onAction);
  const isDisabled = Boolean(action.disabled) || !hasTarget;
  const className = primary
    ? "inline-flex min-h-9 items-center justify-center gap-2 rounded-lg border border-cortex-primary/40 bg-cortex-primary/15 px-3 py-2 text-sm font-semibold text-cortex-primary hover:border-cortex-primary/70 disabled:cursor-not-allowed disabled:opacity-50"
    : "inline-flex min-h-8 items-center justify-center gap-1.5 rounded-lg border border-cortex-border px-2.5 py-1.5 text-xs font-semibold text-cortex-text-main hover:border-cortex-primary/35 disabled:cursor-not-allowed disabled:opacity-50";
  if (action.href && !isDisabled) {
    return (
      <Link href={action.href} className={className} title={action.label}>
        <Icon className="h-3.5 w-3.5" />
        {action.label}
      </Link>
    );
  }
  return (
    <button
      type="button"
      className={className}
      disabled={isDisabled}
      title={action.disabledReason ?? action.label}
      onClick={() => onAction?.(item, action)}
    >
      <Icon className="h-3.5 w-3.5" />
      {action.label}
    </button>
  );
}

function primaryReviewAction(item: TeamWorkItem): TeamInteraction | null {
  if (needsExternalMutationVerification(item)) return null;
  const output = firstOutputAction(item.outputRefs);
  if (item.state === "output_ready" && output) return output;
  if (isStaleFailedPlanItem(item)) return enabledAction(item, "archive") ?? enabledAction(item, "inspect");
  if (item.state === "degraded") return enabledAction(item, "recover") ?? enabledAction(item, "inspect");
  if (item.state === "needs_operator") return enabledAction(item, "steer") ?? enabledAction(item, "recover");
  if (item.state === "queued" || item.state === "running" || item.state === "reviewing") {
    return enabledAction(item, "inspect") ?? enabledAction(item, "pause");
  }
  return enabledAction(item, "inspect") ?? enabledAction(item, "archive");
}

function enabledAction(item: TeamWorkItem, action: TeamInteraction["action"]) {
  return item.interactions.find((candidate) => candidate.action === action && !candidate.disabled) ?? null;
}

function firstOutputAction(outputs?: TeamOutputRef[], label = "Open output"): TeamInteraction | null {
  const output = outputs?.find((candidate) => outputURL(candidate));
  if (!output) return null;
  return {
    action: "inspect",
    label,
    href: outputURL(output) ?? undefined,
  };
}

function outputURL(output: TeamOutputRef): string | null {
  const storageRef = output.storage_ref?.trim() ?? "";
  const entrypoint = output.entrypoint?.trim() ?? "";
  const value = entrypoint
    ? `${storageRef.replace(/[\\/]+$/, "")}/${entrypoint}`
    : storageRef;
  if (!value) return null;
  if (/^(https?:)?\/\//i.test(value) || value.startsWith("/")) return value;
  return value.includes("/") || value.includes(".")
    ? `/api/v1/workspace/files/view?path=${encodeURIComponent(value)}`
    : null;
}
