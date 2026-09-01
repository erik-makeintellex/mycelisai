import Link from "next/link";
import { ArrowLeft, Plus, RefreshCw, Users } from "lucide-react";
import {
  compactButtonClassName,
  type GroupLifecycleReport,
  type Monitor,
} from "./groupWorkspaceTypes";

export function GroupsHeader({
  monitor,
  lifecycleReport,
  refreshing,
  archivingExpired,
  onArchiveExpired,
  onRefresh,
}: {
  monitor: Monitor | null;
  lifecycleReport: GroupLifecycleReport | null;
  refreshing: boolean;
  archivingExpired: boolean;
  onArchiveExpired: () => void;
  onRefresh: () => void;
}) {
  const summary = lifecycleReport?.summary;
  const expiredCount = summary?.expired_active_groups ?? 0;
  const reviewCount = summary?.review_needed_groups ?? 0;
  const workCount = summary?.team_work_needing_attention ?? 0;
  return (
    <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="min-w-0">
          <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-primary">
            <Users className="h-3.5 w-3.5" />
            Work
          </div>
          <h1 className="mt-2 text-base font-semibold text-cortex-text-main sm:text-lg">
            Review active work and delivered results.
          </h1>
          <p className="mt-1 hidden max-w-3xl text-sm leading-5 text-cortex-text-muted sm:line-clamp-2">
            Start with the outcome or deliverable. Group and team details stay
            available here for advanced review.
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          {summary ? (
            <div className="inline-flex flex-wrap items-center gap-2 rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-xs text-cortex-text-muted">
              <span className="font-semibold text-cortex-text-main">
                {reviewCount} need review
              </span>
              <span className="hidden sm:inline">{expiredCount} expired</span>
              <span className="hidden sm:inline">{workCount} work items</span>
            </div>
          ) : null}
          {expiredCount > 0 ? (
            <button
              type="button"
              onClick={onArchiveExpired}
              disabled={archivingExpired}
              className={compactButtonClassName}
            >
              Archive expired
            </button>
          ) : null}
          <Link href="/dashboard" className={compactButtonClassName}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            Open Soma
          </Link>
          <Link href="/groups?panel=create" className={compactButtonClassName}>
            <Plus className="mr-2 h-4 w-4" />
            Advanced group setup
          </Link>
          <button type="button" onClick={onRefresh} className={compactButtonClassName}>
            <RefreshCw className={`mr-2 h-4 w-4 ${refreshing ? "animate-spin" : ""}`} />
            Refresh
          </button>
        </div>
      </div>
      {monitor ? (
        <p className="mt-2 text-xs text-cortex-text-muted">
          Advanced group monitor is {monitor.status || "offline"}.
        </p>
      ) : null}
    </div>
  );
}
