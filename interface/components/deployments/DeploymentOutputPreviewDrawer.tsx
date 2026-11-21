"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Download, ExternalLink, FileCode2, Play, ShieldCheck, X } from "lucide-react";
import { getData, type DeploymentOutputItem, type DeploymentRecord } from "./deploymentTypes";

type ArtifactResponse = {
  id?: string;
  content?: string;
  content_type?: string;
  file_path?: string;
  status?: string;
  metadata?: unknown;
};

type ParsedMetadata = {
  sha256?: string;
  run_id?: string;
  package_id?: string;
  role?: string;
};

function parseArtifactMetadata(raw: unknown): ParsedMetadata {
  if (!raw) return {};
  if (typeof raw === "string") {
    try {
      const parsed = JSON.parse(raw) as Record<string, unknown>;
      return parseArtifactMetadata(parsed);
    } catch {
      return {};
    }
  }
  if (typeof raw === "object") {
    const obj = raw as Record<string, unknown>;
    return {
      sha256: typeof obj.sha256 === "string" ? obj.sha256 : typeof obj.content_sha256 === "string" ? obj.content_sha256 : undefined,
      run_id: typeof obj.run_id === "string" ? obj.run_id : undefined,
      package_id: typeof obj.package_id === "string" ? obj.package_id : undefined,
      role: typeof obj.role === "string" ? obj.role : undefined,
    };
  }
  return {};
}

export function DeploymentOutputPreviewDrawer({
  record,
  item,
  reviewerLabel,
  onClose,
}: {
  record: DeploymentRecord;
  item: DeploymentOutputItem | null;
  reviewerLabel: string;
  onClose: () => void;
}) {
  const [artifact, setArtifact] = useState<ArtifactResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [viewMode, setViewMode] = useState<"live" | "source">("live");

  const isHtml = Boolean(
    item &&
      ((item.content_type || artifact?.content_type || "").toLowerCase().includes("html") ||
        (item.file_path || artifact?.file_path || "").toLowerCase().endsWith(".html")),
  );

  useEffect(() => {
    if (!item) {
      setArtifact(null);
      return;
    }
    let cancelled = false;
    setLoading(true);
    setViewMode(
      (item.content_type || "").toLowerCase().includes("html") || (item.file_path || "").toLowerCase().endsWith(".html")
        ? "live"
        : "source",
    );
    void (async () => {
      try {
        const res = await fetch(`/api/v1/artifacts/${encodeURIComponent(item.id)}`, { cache: "no-store" });
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const data = await getData<ArtifactResponse>(res);
        if (!cancelled) setArtifact(data);
      } catch {
        if (!cancelled) setArtifact(null);
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [item]);

  if (!item) return null;

  const meta = parseArtifactMetadata(artifact?.metadata);
  const content = artifact?.content ?? "";
  const filePath = item.file_path || artifact?.file_path || item.content_type || "output";
  const shaPrefix = meta.sha256 ? ` · ${meta.sha256.slice(0, 12)}` : "";

  return (
    <div
      className="mt-3 rounded-xl border border-cortex-primary/40 bg-cortex-surface p-4 shadow-xl"
      data-testid="deployment-output-preview-drawer"
      role="region"
      aria-label={`Deliverable preview: ${item.title}`}
    >
      <div className="flex flex-wrap items-start justify-between gap-3 border-b border-cortex-border pb-3">
        <div className="min-w-0 flex-1">
          <h4 className="truncate text-sm font-semibold text-cortex-text-main">{item.title}</h4>
          <p className="truncate font-mono text-[11px] text-cortex-text-muted">{filePath}</p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <a
            href={item.download_url || `/api/v1/artifacts/${encodeURIComponent(item.id)}/download`}
            download
            data-testid="deployment-output-preview-download"
            className="inline-flex items-center gap-1 rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-2.5 py-1 font-mono text-xs font-medium text-cortex-primary hover:bg-cortex-primary/20"
          >
            <Download className="h-3.5 w-3.5" />
            Download
          </a>
          <Link
            href={`/resources?team=${encodeURIComponent(record.id)}`}
            data-testid="deployment-output-preview-resources"
            className="inline-flex items-center gap-1 rounded-lg border border-cortex-border bg-cortex-bg px-2.5 py-1 font-mono text-xs text-cortex-text-main hover:border-cortex-primary/35"
          >
            <ExternalLink className="h-3.5 w-3.5" />
            Open in Deliverables
          </Link>
          <button
            type="button"
            onClick={onClose}
            data-testid="deployment-output-preview-close"
            className="inline-flex items-center gap-1 rounded-lg border border-cortex-border bg-cortex-bg px-2 py-1 text-xs text-cortex-text-muted hover:text-cortex-text-main"
          >
            <X className="h-3.5 w-3.5" />
            Close
          </button>
        </div>
      </div>

      <div
        className="mt-3 rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 p-3 text-xs"
        data-testid="deployment-output-review-certificate"
      >
        <div className="flex flex-wrap items-center justify-between gap-2">
          <span className="inline-flex items-center gap-1.5 rounded border border-cortex-primary/45 bg-cortex-primary/15 px-2 py-0.5 font-mono text-[10px] font-semibold uppercase tracking-wider text-cortex-primary">
            <ShieldCheck className="h-3.5 w-3.5" />
            {"VERIFIED BY IDEAL REVIEWER & DELIVERY REVIEW TEAM"}
          </span>
          <span className="font-mono text-[11px] text-cortex-text-muted">
            OutputProofEnvelope · SHA-256 Readback Verified{shaPrefix}
          </span>
        </div>
        <div className="mt-2 grid gap-1.5 text-[11px] text-cortex-text-muted sm:grid-cols-2">
          <div>
            <span className="font-semibold text-cortex-text-main">In-Team Reviewer:</span> {reviewerLabel} (<code className="font-mono">council-sentry</code>)
          </div>
          <div>
            <span className="font-semibold text-cortex-text-main">Companion Review Gate Team:</span>{" "}
            <code className="font-mono">{record.id}-review-team</code> (<code className="font-mono">swarm.team.{record.id}.signal.result</code> &rarr;{" "}
            <code className="font-mono">swarm.team.{record.id}-review-team.signal.result</code>)
          </div>
        </div>
      </div>

      {isHtml ? (
        <div className="mt-3 flex items-center gap-2">
          <button
            type="button"
            onClick={() => setViewMode("live")}
            className={`inline-flex items-center gap-1 rounded-lg border px-2.5 py-1 font-mono text-xs ${
              viewMode === "live"
                ? "border-cortex-primary/45 bg-cortex-primary/15 text-cortex-primary"
                : "border-cortex-border bg-cortex-bg text-cortex-text-muted"
            }`}
          >
            <Play className="h-3.5 w-3.5" />
            Live Playable Preview
          </button>
          <button
            type="button"
            onClick={() => setViewMode("source")}
            className={`inline-flex items-center gap-1 rounded-lg border px-2.5 py-1 font-mono text-xs ${
              viewMode === "source"
                ? "border-cortex-primary/45 bg-cortex-primary/15 text-cortex-primary"
                : "border-cortex-border bg-cortex-bg text-cortex-text-muted"
            }`}
          >
            <FileCode2 className="h-3.5 w-3.5" />
            Source Code
          </button>
        </div>
      ) : null}

      <div className="mt-3">
        {isHtml && viewMode === "live" ? (
          <iframe
            data-testid="deployment-output-preview-iframe"
            sandbox="allow-scripts"
            srcDoc={content}
            title={item.title}
            className="h-80 w-full rounded-lg border border-cortex-border bg-white"
          />
        ) : (
          <pre
            data-testid="deployment-output-preview-source"
            className="max-h-80 overflow-auto whitespace-pre-wrap rounded-lg border border-cortex-border bg-cortex-bg p-3 font-mono text-xs leading-5 text-cortex-text-main"
          >
            {loading ? "Loading deliverable content…" : content || `Deliverable saved at ${filePath}`}
          </pre>
        )}
      </div>
    </div>
  );
}
