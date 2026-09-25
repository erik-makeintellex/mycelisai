"use client";

import { AlertTriangle, ShieldCheck, Sparkles } from "lucide-react";
import { useCortexStore } from "@/store/useCortexStore";
import {
  CORE_RECOVERY_HINT,
  CORE_UNREACHABLE_TITLE,
  selectCoreReachability,
  type CoreReachability,
} from "@/lib/coreReachability";

const STATUS_COPY: Record<CoreReachability, string> = {
  online: "Ready",
  degraded: "Degraded — some services are offline",
  unreachable: CORE_UNREACHABLE_TITLE,
  unknown: "Checking Core connection…",
};

const POSTURE_COPY: Record<Exclude<CoreReachability, "online">, string> = {
  degraded: "Governed execution may be limited",
  unreachable: "Governed execution unavailable until Core is reachable",
  unknown: "Governance posture not yet confirmed",
};

export function SomaHeader({
  organizationName,
  status,
  activeMode,
  governancePosture = "Governed execution enabled",
}: {
  organizationName?: string | null;
  status?: string;
  activeMode?: string | null;
  governancePosture?: string;
}) {
  const reachability = useCortexStore(selectCoreReachability);
  const coreDown = reachability === "unreachable";
  const scope = coreDown
    ? "Soma cannot take requests until the Core API is reachable."
    : organizationName
      ? `${reachability === "online" ? "Ready to coordinate" : "Coordinating"} work for ${organizationName}`
      : "Create or resume an AI Organization with Soma";
  const statusLabel = status ?? STATUS_COPY[reachability];
  const posture = reachability === "online" ? governancePosture : POSTURE_COPY[reachability];

  return (
    <header className="flex flex-col gap-4 border-b border-cortex-border p-5 lg:flex-row lg:items-start lg:justify-between">
      <div>
        <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.2em] text-cortex-primary">
          <Sparkles className="h-3.5 w-3.5" />
          Soma
        </div>
        <h1 className="mt-3 text-3xl font-semibold tracking-tight text-cortex-text-main">
          What do you want Soma to do?
        </h1>
        <p className="mt-2 max-w-3xl text-sm leading-7 text-cortex-text-muted">
          {scope}
        </p>
      </div>
      <div
        data-testid="soma-header-status"
        data-reachability={reachability}
        className={`rounded-2xl border px-4 py-3 text-sm text-cortex-text-muted lg:max-w-sm ${
          coreDown ? "border-cortex-danger/40 bg-cortex-danger/10" : "border-cortex-border bg-cortex-bg"
        }`}
      >
        <p className={`flex items-center gap-2 font-medium ${coreDown ? "text-cortex-danger" : "text-cortex-text-main"}`}>
          {coreDown ? <AlertTriangle className="h-4 w-4" /> : null}
          {statusLabel}
        </p>
        {coreDown ? <p className="mt-1">{CORE_RECOVERY_HINT}</p> : null}
        {activeMode ? <p className="mt-1">Mode: {activeMode}</p> : null}
        <p className="mt-1 flex items-center gap-2">
          <ShieldCheck className={`h-4 w-4 ${reachability === "online" ? "text-cortex-primary" : "text-cortex-text-muted"}`} />
          {posture}
        </p>
      </div>
    </header>
  );
}
