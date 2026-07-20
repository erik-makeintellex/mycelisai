"use client";

import { AlertCircle, AlertTriangle, Archive, CheckCircle2, Radio, ShieldCheck } from "lucide-react";
import {
  outcomeHealthClassName,
  outcomeHealthLabel,
  type OutcomeHealthState,
} from "@/lib/outcomeHealth";

export function OutcomeHealthBadge({
  health,
  className = "",
}: {
  health: OutcomeHealthState;
  className?: string;
}) {
  const label = outcomeHealthLabel(health);
  return (
    <span
      aria-label={`Outcome health: ${label}`}
      className={`inline-flex h-6 shrink-0 items-center gap-1.5 rounded-full border px-2 text-[10px] font-semibold ${outcomeHealthClassName(health)} ${className}`}
    >
      <OutcomeHealthIcon health={health} />
      {label}
    </span>
  );
}

function OutcomeHealthIcon({ health }: { health: OutcomeHealthState }) {
  const props = { className: "h-3 w-3", "aria-hidden": true } as const;
  if (health === "blocked") return <AlertTriangle {...props} />;
  if (health === "degraded" || health === "waiting") return <AlertCircle {...props} />;
  if (health === "running") return <Radio {...props} />;
  if (health === "completed") return <CheckCircle2 {...props} />;
  if (health === "archived") return <Archive {...props} />;
  return <ShieldCheck {...props} />;
}
