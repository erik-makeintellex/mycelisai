"use client";

import { Download, FileText, Search } from "lucide-react";
import type React from "react";
import type { Artifact } from "@/store/useCortexStore";
import { memoryResultScore, type MemorySelection, type SearchResult } from "./memorySelection";

export default function MemoryDetailPanel({
  selection,
}: {
  selection: MemorySelection | null;
}) {
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
            Select a memory search result or artifact to inspect the full
            record.
          </p>
        </div>
      ) : selection.kind === "search" ? (
        <SearchDetail result={selection.result} />
      ) : (
        <ArtifactDetail artifact={selection.artifact} />
      )}
    </aside>
  );
}

function SearchDetail({ result }: { result: SearchResult }) {
  return (
    <DetailShell
      icon={<Search className="h-4 w-4 text-cortex-info" />}
      title={result.source || "Memory result"}
      eyebrow={`${Math.round(memoryResultScore(result) * 100)}% relevance`}
    >
      <Meta label="ID" value={result.id} />
      <Meta label="Created" value={formatDate(result.created_at)} />
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
      {artifact.file_path ? (
        <Meta label="Path" value={artifact.file_path} />
      ) : null}
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
          This artifact record does not include inline content. Use download or
          path details to inspect the stored file.
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
          <p className="text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
            {eyebrow}
          </p>
          <h2 className="mt-1 break-words text-lg font-semibold text-cortex-text-main">
            {title}
          </h2>
        </div>
      </div>
      <div className="space-y-3">{children}</div>
    </div>
  );
}

function Meta({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2">
      <p className="text-[9px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
        {label}
      </p>
      <p className="mt-1 break-words text-xs font-semibold text-cortex-text-main">
        {value}
      </p>
    </div>
  );
}

function formatDate(value: string) {
  const parsed = new Date(value);
  return Number.isFinite(parsed.getTime()) ? parsed.toLocaleString() : value;
}
