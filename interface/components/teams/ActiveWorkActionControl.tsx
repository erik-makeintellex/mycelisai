import Link from "next/link";
import { Archive, Eye, Pause, Play, RefreshCw, Search, Send } from "lucide-react";
import type { TeamInteraction, TeamWorkItem } from "@/store/useCortexStore";

const actionIcons = {
  inspect: Eye,
  steer: Send,
  start_work: Play,
  pause: Pause,
  resume: Play,
  recover: RefreshCw,
  archive: Archive,
};

export function ActiveWorkActionControl({
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
