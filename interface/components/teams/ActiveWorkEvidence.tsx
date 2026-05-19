"use client";

import Link from "next/link";
import { ExternalLink } from "lucide-react";
import type { TeamWorkItem } from "@/store/useCortexStore";

export function ActiveWorkEvidence({ item }: { item: TeamWorkItem }) {
  const outputRefs = item.outputRefs ?? [];
  const proofRefs = item.proofRefs ?? [];
  const auditRefs = item.auditRefs ?? [];
  if (!item.runId && outputRefs.length === 0 && proofRefs.length === 0 && auditRefs.length === 0) {
    return null;
  }

  return (
    <div className="mt-3 flex flex-wrap gap-1.5 text-[11px]">
      {item.runId ? (
        <Link
          href={`/runs/${encodeURIComponent(item.runId)}`}
          className="inline-flex items-center gap-1 rounded border border-cortex-info/25 bg-cortex-info/10 px-2 py-1 font-mono text-cortex-info hover:underline"
        >
          Run proof
          <ExternalLink className="h-3 w-3" />
        </Link>
      ) : null}
      {outputRefs.slice(0, 3).map((output) => (
        <EvidenceLink
          key={output.output_id}
          label={output.label || "Output"}
          href={outputURL(output.storage_ref)}
          muted={!output.storage_ref}
        />
      ))}
      {proofRefs.slice(0, 3).map((proof) => (
        <EvidenceLink
          key={`proof-${proof}`}
          label={`Proof ${compactID(proof)}`}
          href={item.runId ? `/runs/${encodeURIComponent(item.runId)}` : null}
          muted={!item.runId}
        />
      ))}
      {auditRefs.slice(0, 2).map((audit) => (
        <span
          key={`audit-${audit}`}
          className="inline-flex items-center rounded border border-cortex-border bg-cortex-surface px-2 py-1 font-mono text-cortex-text-muted"
          title={audit}
        >
          Audit {compactID(audit)}
        </span>
      ))}
    </div>
  );
}

function EvidenceLink({
  label,
  href,
  muted,
}: {
  label: string;
  href: string | null;
  muted?: boolean;
}) {
  const className =
    "inline-flex max-w-full items-center gap-1 rounded border border-cortex-border bg-cortex-surface px-2 py-1 font-mono text-cortex-text-muted";
  if (!href || muted) {
    return (
      <span className={className}>
        <span className="max-w-[12rem] truncate">{label}</span>
      </span>
    );
  }
  return (
    <a href={href} className={`${className} hover:border-cortex-primary/40 hover:text-cortex-primary`}>
      <span className="max-w-[12rem] truncate">{label}</span>
      <ExternalLink className="h-3 w-3 shrink-0" />
    </a>
  );
}

function outputURL(storageRef?: string | null): string | null {
  const value = storageRef?.trim();
  if (!value) return null;
  if (value.startsWith("http://") || value.startsWith("https://") || value.startsWith("/")) {
    return value;
  }
  if (value.includes("/") || value.includes(".")) {
    return `/api/v1/workspace/files/view?path=${encodeURIComponent(value)}`;
  }
  return null;
}

function compactID(value: string) {
  return value.length > 10 ? `${value.slice(0, 8)}...` : value;
}
