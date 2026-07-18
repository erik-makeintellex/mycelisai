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
  const Icon = outcomeHealthIcon(health);
  return (
    <span
      aria-label={`Outcome health: ${label}`}
      className={`inline-flex h-6 shrink-0 items-center gap-1.5 rounded-full border px-2 text-[10px] font-semibold ${outcomeHealthClassName(health)} ${className}`}
    >
      <Icon className="h-3 w-3" aria-hidden="true" />
      {label}
    </span>
  );
}

function outcomeHealthIcon(health: OutcomeHealthState) {
  if (health === "blocked") return AlertTriangle;
  if (health === "degraded" || health === "waiting") return AlertCircle;
  if (health === "running") return Radio;
  if (health === "completed") return CheckCircle2;
  if (health === "archived") return Archive;
  return ShieldCheck;
}
