"use client";

import { useState } from "react";
import type { TeamWorkItem } from "@/store/useCortexStore";

export function ActiveWorkAdvancedProjection({ item }: { item: TeamWorkItem }) {
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
      {isOpen ? (
        <div className="mt-3 space-y-2">
          {rows.map(([label, values]) => (
            <AdvancedRow key={label} label={label} values={values} />
          ))}
          {advanced.promptCount ? (
            <p className="font-mono text-[10px] text-cortex-text-muted">
              {advanced.promptCount} prompt
              {advanced.promptCount === 1 ? "" : "s"} available in agent inspect.
            </p>
          ) : null}
        </div>
      ) : null}
    </details>
  );
}

function AdvancedRow({ label, values }: { label: string; values: string[] }) {
  return (
    <div>
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
  );
}
