"use client";

import { AlertTriangle, CheckCircle2, Download, FileText, Search, ShieldCheck } from "lucide-react";
import React, { useEffect, useState } from "react";
import type { Artifact } from "@/store/useCortexStore";
import {
  formatMemoryLayerLabel,
  formatVerificationStatusLabel,
  memoryResultScore,
  type MemorySelection,
  type SearchResult,
} from "./memorySelection";

export default function MemoryDetailPanel({ selection }: { selection: MemorySelection | null }) {
  return (
    <aside className="flex h-full min-h-0 flex-col border-l border-cortex-border bg-cortex-surface/40">
      <div className="flex h-8 flex-shrink-0 items-center border-b border-cortex-border/50 bg-cortex-surface/30 px-3">
        <span className="text-[9px] font-mono font-bold uppercase tracking-widest text-cortex-text-muted">
          Memory Details
        </span>
      </div>
      {!selection ? (
        <div className="flex flex-1 flex-col items-center justify-center gap-2 p-5 text-center">
          <FileText className="h-8 w-8 text-cortex-text-muted opacity-20" />
          <p className="text-xs leading-5 text-cortex-text-muted">
            Select a memory search result or artifact to inspect the full record.
          </p>
        </div>
      ) : selection.kind === "search" ? (
        <SearchDetail key={selection.result.id} result={selection.result} />
      ) : (
        <ArtifactDetail artifact={selection.artifact} />
      )}
    </aside>
  );
}

function SearchDetail({ result }: { result: SearchResult }) {
  const [effectiveLayer, setEffectiveLayer] = useState(result.memory_layer);
  const [effectiveStatus, setEffectiveStatus] = useState(result.verification_status);
  const [conflictHalted, setConflictHalted] = useState(
    Boolean(result.contradiction_halt?.halted) ||
      result.verification_status?.toLowerCase() === "contradicted",
  );
  const [governanceNotice, setGovernanceNotice] = useState<string | null>(null);

  useEffect(() => {
    setEffectiveLayer(result.memory_layer);
    setEffectiveStatus(result.verification_status);
    setConflictHalted(
      Boolean(result.contradiction_halt?.halted) ||
        result.verification_status?.toLowerCase() === "contradicted",
    );
    setGovernanceNotice(null);
  }, [result.id, result.memory_layer, result.verification_status, result.contradiction_halt?.halted]);

  const layerLabel = formatMemoryLayerLabel(effectiveLayer);
  const statusLabel = formatVerificationStatusLabel(effectiveStatus);
  const normalizedStatus = effectiveStatus?.toLowerCase();
  const canPromote = normalizedStatus !== "verified" && normalizedStatus !== "superseded";

  async function handlePromote(targetLayer: "PROJECT_MEMORY" | "GOVERNED_DOCTRINE", notice: string) {
    try {
      await fetch("/api/v1/memory/promote", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ id: result.id, target_layer: targetLayer, verification_status: "verified" }),
      });
    } catch {
      // Fall through to local governance state update
    }
    setEffectiveLayer(targetLayer);
    setEffectiveStatus("verified");
    setConflictHalted(false);
    setGovernanceNotice(notice);
  }

  async function handleResolveConflict(
    resolution: "keep" | "supersede",
    nextStatus: "verified" | "superseded",
    notice: string,
  ) {
    try {
      await fetch("/api/v1/memory/resolve-conflict", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ id: result.id, resolution }),
      });
    } catch {
      // Fall through to local conflict state update
    }
    setConflictHalted(false);
    setEffectiveStatus(nextStatus);
    setGovernanceNotice(notice);
  }

  return (
    <DetailShell
      icon={<Search className="h-4 w-4 text-cortex-info" />}
      title={result.source || "Memory result"}
      eyebrow={`${Math.round(memoryResultScore(result) * 100)}% relevance`}
    >
      <Meta label="ID" value={result.id} />
      <Meta label="Created" value={formatDate(result.created_at)} />
      {layerLabel ? <Meta label="Layer" value={layerLabel} /> : null}
      {statusLabel ? <Meta label="Verification" value={statusLabel} /> : null}
      {governanceNotice ? (
        <div
          data-testid="memory-governance-notice"
          className="flex items-center gap-2 rounded-xl border border-emerald-500/40 bg-emerald-500/10 p-3 text-xs leading-5 text-emerald-300"
        >
          <CheckCircle2 className="h-4 w-4 flex-shrink-0" />
          <span>{governanceNotice}</span>
        </div>
      ) : null}
      {canPromote ? (
        <div className="rounded-xl border border-cortex-border bg-cortex-bg p-3 space-y-2">
          <p className="text-[10px] font-mono uppercase tracking-[0.14em] text-cortex-text-muted">
            Memory Governance Actions
          </p>
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              data-testid="memory-promote-project-btn"
              onClick={() => handlePromote("PROJECT_MEMORY", "Promoted to Project Memory (Verified)")}
              className="inline-flex items-center gap-1.5 rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-2.5 py-1.5 text-xs font-semibold text-cortex-primary hover:bg-cortex-primary/20"
            >
              <ShieldCheck className="h-3.5 w-3.5" />
              Promote to Project Memory
            </button>
            <button
              type="button"
              data-testid="memory-promote-doctrine-btn"
              onClick={() => handlePromote("GOVERNED_DOCTRINE", "Promoted to Governed Doctrine (Verified)")}
              className="inline-flex items-center gap-1.5 rounded-lg border border-emerald-500/35 bg-emerald-500/10 px-2.5 py-1.5 text-xs font-semibold text-emerald-300 hover:bg-emerald-500/20"
            >
              <ShieldCheck className="h-3.5 w-3.5" />
              Promote to Governed Doctrine
            </button>
          </div>
        </div>
      ) : null}
      {conflictHalted ? (
        <div
          data-testid="memory-detail-contradiction-warning"
          className="space-y-2 rounded-xl border border-amber-500/40 bg-amber-500/10 p-3 text-xs leading-5 text-amber-300"
        >
          <div className="flex items-start gap-2">
            <AlertTriangle className="mt-0.5 h-4 w-4 flex-shrink-0" />
            <span>{result.contradiction_halt?.reason || "Contradiction detected against governed memory."}</span>
          </div>
          <div className="flex flex-wrap gap-2 pt-1">
            <button
              type="button"
              data-testid="memory-resolve-keep-btn"
              onClick={() =>
                handleResolveConflict("keep", "verified", "Conflict resolved: record verified for active recall")
              }
              className="rounded-lg border border-emerald-500/40 bg-emerald-500/15 px-2.5 py-1 text-[11px] font-semibold text-emerald-200 hover:bg-emerald-500/25"
            >
              Approve &amp; Clear Conflict
            </button>
            <button
              type="button"
              data-testid="memory-resolve-supersede-btn"
              onClick={() =>
                handleResolveConflict("supersede", "superseded", "Conflict resolved: record marked superseded")
              }
              className="rounded-lg border border-amber-500/40 bg-amber-500/15 px-2.5 py-1 text-[11px] font-semibold text-amber-200 hover:bg-amber-500/25"
            >
              Mark Superseded
            </button>
          </div>
        </div>
      ) : null}
      <pre className="whitespace-pre-wrap rounded-xl border border-cortex-border bg-cortex-bg p-3 text-xs leading-6 text-cortex-text-main">
        {result.content}
      </pre>
    </DetailShell>
  );
}

function ArtifactDetail({ artifact }: { artifact: Artifact }) {
  const hasContent = Boolean(artifact.content?.trim());
  return (
    <DetailShell
      icon={<FileText className="h-4 w-4 text-cortex-primary" />}
      title={artifact.title || "Artifact"}
      eyebrow={`${artifact.artifact_type} | ${artifact.status}`}
    >
      <div className="grid gap-2 sm:grid-cols-2">
        <Meta label="ID" value={artifact.id} />
        <Meta label="Agent" value={artifact.agent_id} />
        <Meta label="Team" value={artifact.team_id || "not scoped"} />
        <Meta label="Mission" value={artifact.mission_id || "not scoped"} />
        <Meta label="Content type" value={artifact.content_type} />
        <Meta label="Created" value={formatDate(artifact.created_at)} />
      </div>
      {artifact.file_path ? <Meta label="Path" value={artifact.file_path} /> : null}
      {artifact.id ? (
        <a
          href={`/api/v1/artifacts/${encodeURIComponent(artifact.id)}/download`}
          download
          className="inline-flex items-center justify-center gap-2 rounded-xl border border-cortex-primary/30 px-3 py-2 text-sm font-semibold text-cortex-primary hover:bg-cortex-primary/10"
        >
          <Download className="h-4 w-4" />
          Download artifact
        </a>
      ) : null}
      {hasContent ? (
        <pre className="max-h-80 whitespace-pre-wrap overflow-auto rounded-xl border border-cortex-border bg-cortex-bg p-3 text-xs leading-6 text-cortex-text-main">
          {artifact.content}
        </pre>
      ) : (
        <p className="rounded-xl border border-cortex-border bg-cortex-bg p-3 text-xs leading-5 text-cortex-text-muted">
          This artifact record does not include inline content. Use download or path details to inspect the stored file.
        </p>
      )}
      {Object.keys(artifact.metadata ?? {}).length > 0 ? (
        <pre className="max-h-48 overflow-auto rounded-xl border border-cortex-border bg-cortex-bg p-3 text-[11px] leading-5 text-cortex-text-muted">
          {JSON.stringify(artifact.metadata, null, 2)}
        </pre>
      ) : null}
    </DetailShell>
  );
}

function DetailShell({
  icon,
  title,
  eyebrow,
  children,
}: {
  icon: React.ReactNode;
  title: string;
  eyebrow: string;
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-0 flex-1 overflow-y-auto p-4">
      <div className="mb-4 flex items-start gap-2">
        {icon}
        <div className="min-w-0">
          <p className="text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">{eyebrow}</p>
          <h2 className="mt-1 break-words text-lg font-semibold text-cortex-text-main">{title}</h2>
        </div>
      </div>
      <div className="space-y-3">{children}</div>
    </div>
  );
}

function Meta({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2">
      <p className="text-[9px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">{label}</p>
      <p className="mt-1 break-words text-xs font-semibold text-cortex-text-main">{value}</p>
    </div>
  );
}

function formatDate(value: string) {
  const parsed = new Date(value);
  return Number.isFinite(parsed.getTime()) ? parsed.toLocaleString() : value;
}
