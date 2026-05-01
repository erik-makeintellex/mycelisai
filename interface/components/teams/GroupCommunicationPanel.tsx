import { MessageSquare, RefreshCw } from "lucide-react";
import {
  compactButtonClassName,
  inputClassName,
  linkClassName,
  type Group,
  type Monitor,
} from "./groupWorkspaceTypes";

export function GroupCommunicationPanel({
  monitor,
  selectedGroup,
  broadcastMessage,
  broadcasting,
  onBroadcastMessageChange,
  onBroadcast,
  onRefresh,
}: {
  monitor: Monitor | null;
  selectedGroup: Group | null;
  broadcastMessage: string;
  broadcasting: boolean;
  onBroadcastMessageChange: (message: string) => void;
  onBroadcast: () => void;
  onRefresh: () => void;
}) {
  const archived = selectedGroup?.status === "archived";
  return (
    <section className="rounded-2xl border border-cortex-border bg-cortex-surface p-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <div className="flex items-center gap-2">
            <MessageSquare className="h-4 w-4 text-cortex-primary" />
            <h2 className="text-sm font-semibold uppercase tracking-[0.16em] text-cortex-text-main">
              Coordination activity
            </h2>
          </div>
          <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
            Use this after selecting and reviewing the group. Bus health and
            broadcasts support the lane; they are not the primary setup surface.
          </p>
        </div>
        <button
          type="button"
          onClick={onRefresh}
          className={compactButtonClassName}
        >
          <RefreshCw className="mr-2 h-4 w-4" />
          Refresh
        </button>
      </div>
      <div className="mt-4 grid gap-4 lg:grid-cols-[minmax(0,1fr)_320px]">
        <div className="rounded-xl border border-cortex-border bg-cortex-bg p-4">
          <h3 className="text-sm font-semibold text-cortex-text-main">
            {archived ? "Group is archived" : "Broadcast a focused ask"}
          </h3>
          {!selectedGroup ? (
            <p className="mt-3 text-sm leading-6 text-cortex-text-muted">
              Select a group before sending coordination messages.
            </p>
          ) : archived ? (
            <p
              className="mt-3 text-sm leading-6 text-cortex-text-muted"
              data-testid="groups-archived-readonly-note"
            >
              This temporary group is archived. Keep it available for retained
              output review, but send new coordination through an active group
              or Soma.
            </p>
          ) : (
            <>
              <textarea
                aria-label="Broadcast message"
                value={broadcastMessage}
                onChange={(event) =>
                  onBroadcastMessageChange(event.target.value)
                }
                rows={4}
                className={`${inputClassName} mt-3 resize-y`}
              />
              <button
                type="button"
                onClick={onBroadcast}
                disabled={broadcasting || !broadcastMessage.trim()}
                className="mt-3 rounded-xl border border-cortex-primary/30 px-4 py-2 text-sm font-semibold text-cortex-primary disabled:opacity-60"
              >
                {broadcasting ? "Broadcasting..." : "Broadcast to group"}
              </button>
            </>
          )}
        </div>
        <div className="rounded-xl border border-cortex-border bg-cortex-bg p-4">
          <h3 className="text-sm font-semibold text-cortex-text-main">
            Message bus
          </h3>
          <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
            {monitor
              ? `Bus ${monitor.status || "offline"} with ${monitor.published_count ?? 0} published messages.`
              : "Bus monitor has not reported yet."}
          </p>
          {monitor?.last_group_id ? (
            <p className="mt-2 text-xs font-mono text-cortex-text-muted">
              Last group: {monitor.last_group_id}
            </p>
          ) : null}
          {monitor?.last_error ? (
            <p className="mt-2 text-xs text-cortex-danger">
              {monitor.last_error}
            </p>
          ) : null}
          <a
            href="/system?tab=nats"
            className={`${linkClassName} mt-3 inline-flex`}
          >
            Open bus diagnostics
          </a>
        </div>
      </div>
    </section>
  );
}
