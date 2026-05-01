"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import {
  Activity,
  Database,
  Radio,
  RefreshCw,
  Route,
  Server,
  Zap,
} from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";
import {
  BusActivityPanel,
  RunList,
  SummaryCard,
  summarizeSignals,
} from "@/components/activity/ActivityPanels";
import { RunEventInspector } from "@/components/activity/RunEventInspector";

export default function ActivityPage() {
  const [selectedRunId, setSelectedRunId] = useState<string | null>(null);
  const recentRuns = useCortexStore((s) => s.recentRuns);
  const isFetchingRuns = useCortexStore((s) => s.isFetchingRuns);
  const fetchRecentRuns = useCortexStore((s) => s.fetchRecentRuns);
  const runTimeline = useCortexStore((s) => s.runTimeline);
  const isFetchingTimeline = useCortexStore((s) => s.isFetchingTimeline);
  const fetchRunTimeline = useCortexStore((s) => s.fetchRunTimeline);
  const streamLogs = useCortexStore((s) => s.streamLogs);
  const streamState = useCortexStore((s) => s.streamConnectionState);
  const initializeStream = useCortexStore((s) => s.initializeStream);
  const services = useCortexStore((s) => s.servicesStatus);
  const fetchServicesStatus = useCortexStore((s) => s.fetchServicesStatus);

  useEffect(() => {
    void fetchRecentRuns();
    void fetchServicesStatus();
    initializeStream();
    const interval = setInterval(() => {
      void fetchRecentRuns();
      void fetchServicesStatus();
    }, 10000);
    return () => clearInterval(interval);
  }, [fetchRecentRuns, fetchServicesStatus, initializeStream]);

  useEffect(() => {
    if (recentRuns.length === 0) {
      setSelectedRunId(null);
      return;
    }
    if (!selectedRunId || !recentRuns.some((run) => run.id === selectedRunId)) {
      setSelectedRunId(recentRuns[0].id);
    }
  }, [recentRuns, selectedRunId]);

  useEffect(() => {
    if (selectedRunId) {
      void fetchRunTimeline(selectedRunId);
    }
  }, [fetchRunTimeline, selectedRunId]);

  const activeRuns = useMemo(
    () => recentRuns.filter((run) => run.status === "running"),
    [recentRuns],
  );
  const signalSummary = useMemo(
    () => summarizeSignals(streamLogs),
    [streamLogs],
  );
  const selectedRun = useMemo(
    () => recentRuns.find((run) => run.id === selectedRunId) ?? null,
    [recentRuns, selectedRunId],
  );
  const nats = services.find((svc) => svc.name === "nats");
  const groupsBus = services.find((svc) => svc.name === "groups_bus");

  return (
    <div className="h-full overflow-y-auto bg-cortex-bg px-6 py-6 text-cortex-text-main">
      <div className="mx-auto flex max-w-6xl flex-col gap-5">
        <header className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.24em] text-cortex-primary">
              <Activity className="h-3.5 w-3.5" />
              Admin Activity
            </div>
            <h1 className="mt-3 text-3xl font-semibold tracking-tight">
              Workflow and bus review
            </h1>
            <p className="mt-2 max-w-3xl text-sm leading-7 text-cortex-text-muted">
              Review active work, recent run outcomes, and the live event stream
              without dropping into raw system diagnostics.
            </p>
          </div>
          <button
            type="button"
            onClick={() => {
              void fetchRecentRuns();
              void fetchServicesStatus();
              initializeStream(true);
            }}
            className="inline-flex items-center gap-2 rounded-2xl border border-cortex-border bg-cortex-surface px-3 py-2 text-sm font-medium hover:border-cortex-primary/30"
          >
            <RefreshCw className="h-4 w-4" />
            Refresh
          </button>
        </header>

        <section className="grid gap-3 md:grid-cols-4">
          <SummaryCard
            icon={Route}
            label="Active workflows"
            value={String(activeRuns.length)}
            href="/runs?status=running"
          />
          <SummaryCard
            icon={Zap}
            label="Recent runs"
            value={String(recentRuns.length)}
            href="/runs"
          />
          <SummaryCard
            icon={Radio}
            label="Live signals"
            value={String(streamLogs.length)}
            href="/activity#message-bus"
          />
          <SummaryCard
            icon={Server}
            label="Bus state"
            value={nats?.status ?? streamState}
            href="/system?tab=nats"
          />
        </section>

        <section className="grid gap-5 xl:grid-cols-[0.85fr_1.15fr]">
          <RunList
            runs={recentRuns}
            selectedRunId={selectedRunId}
            isFetching={isFetchingRuns}
            onSelect={setSelectedRunId}
          />

          <div className="flex flex-col gap-5">
            <RunEventInspector
              run={selectedRun}
              events={runTimeline ?? []}
              isFetching={isFetchingTimeline}
            />
            <BusActivityPanel
              signals={streamLogs}
              summary={signalSummary}
              natsStatus={nats?.status}
              groupsStatus={groupsBus?.status}
            />
          </div>
        </section>

        <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-5">
          <div className="flex items-center gap-2 text-sm text-cortex-text-muted">
            <Database className="h-4 w-4 text-cortex-primary" />
            Use this page for admin review. Use{" "}
            <Link
              href="/system?tab=nats"
              className="text-cortex-primary hover:underline"
            >
              System
            </Link>{" "}
            only when you need lower-level recovery or service commands.
          </div>
        </div>
      </div>
    </div>
  );
}
