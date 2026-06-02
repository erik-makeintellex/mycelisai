"use client";

import Link from "next/link";
import type React from "react";
import { ExternalLink, FileText, Image as ImageIcon, PackageOpen, ShieldCheck } from "lucide-react";
import type { TeamOutputRef } from "@/store/useCortexStore";
import { normalizeWorkspaceOutputUrl } from "./ExecutionSummaryCardModel";
import OutputAccessActions from "./OutputAccessActions";

export function FocusedTeamOutputDock({
  teamName,
  teamId,
  outputRefs,
}: {
  teamName?: string | null;
  teamId?: string | null;
  outputRefs: TeamOutputRef[];
}) {
  const outputs = outputRefs.map(teamOutputItem).filter((item): item is TeamOutputItem => Boolean(item));
  if (!teamId || outputs.length === 0) return null;
  const visible = outputs.slice(0, 3);
  const hiddenCount = Math.max(outputs.length - visible.length, 0);

  return (
    <section
      className="mb-3 rounded-xl border border-cortex-primary/20 bg-cortex-primary/5 p-3"
      data-testid="focused-team-output-dock"
    >
      <div className="flex flex-col gap-2 lg:flex-row lg:items-start lg:justify-between">
        <div className="min-w-0">
          <p className="font-mono text-[10px] font-bold uppercase tracking-[0.16em] text-cortex-primary">
            Team outputs
          </p>
          <p className="mt-1 text-sm leading-5 text-cortex-text-muted">
            {teamName || teamId} retained output is ready in this focused context.
          </p>
        </div>
        <Link
          href={`/teams?team_id=${encodeURIComponent(teamId)}`}
          className="inline-flex shrink-0 items-center gap-1 rounded-lg border border-cortex-border px-2 py-1 text-[10px] font-semibold uppercase tracking-[0.08em] text-cortex-text-main hover:border-cortex-primary/40"
        >
          Team page
          <ExternalLink className="h-3 w-3" />
        </Link>
      </div>
      <div className="mt-3 grid gap-2 xl:grid-cols-3">
        {visible.map((output) => (
          <article
            key={output.id}
            className="min-w-0 rounded-lg border border-cortex-border/70 bg-cortex-bg px-3 py-2"
          >
            <div className="flex items-start gap-2">
              <span className="mt-0.5 text-cortex-primary">{output.icon}</span>
              <div className="min-w-0 flex-1">
                <div className="truncate text-sm font-semibold text-cortex-text-main">{output.label}</div>
                <div className="mt-1 flex flex-wrap items-center gap-1.5 text-[10px] font-mono uppercase tracking-[0.04em] text-cortex-text-muted">
                  <span>{output.kindLabel}</span>
                  {output.proofLabel ? (
                    <span className="inline-flex items-center gap-1 text-cortex-success">
                      <ShieldCheck className="h-3 w-3" />
                      {output.proofLabel}
                    </span>
                  ) : null}
                </div>
                {output.pathLabel ? (
                  <div className="mt-1 truncate text-[10px] font-mono text-cortex-text-muted">
                    {output.pathLabel}
                  </div>
                ) : null}
              </div>
            </div>
            <div className="mt-2 flex flex-wrap gap-1.5">
              <OutputAccessActions
                label={output.label}
                url={output.url}
                storagePath={output.folderPath}
                openLabel="Open"
                folderLabel="Folder"
              />
            </div>
          </article>
        ))}
      </div>
      {hiddenCount > 0 ? (
        <div className="mt-2 text-xs text-cortex-text-muted">
          +{hiddenCount} more retained outputs are available from the Team page or work review.
        </div>
      ) : null}
    </section>
  );
}

type TeamOutputItem = {
  id: string;
  label: string;
  kindLabel: string;
  url: string | null;
  folderPath: string | null;
  pathLabel: string | null;
  proofLabel: string | null;
  icon: React.ReactNode;
};

function teamOutputItem(output: TeamOutputRef): TeamOutputItem | null {
  const label = output.label?.trim() || "Team output";
  const fullPath = outputPath(output.storage_ref, output.entrypoint);
  const url = normalizeWorkspaceOutputUrl(fullPath);
  const folderPath = output.storage_ref?.trim() || workspacePathFromURL(url);
  const kind = output.kind?.trim() || "output";
  if (!label && !url && !folderPath) return null;
  return {
    id: output.output_id || `${label}-${fullPath ?? kind}`,
    label,
    kindLabel: kind.replace(/[_-]+/g, " "),
    url,
    folderPath,
    pathLabel: fullPath || folderPath,
    proofLabel: proofLabel(output),
    icon: outputIcon(kind, fullPath),
  };
}

function outputPath(storageRef?: string, entrypoint?: string) {
  const storage = storageRef?.trim() ?? "";
  const entry = entrypoint?.trim() ?? "";
  if (!entry) return storage || null;
  if (/^(https?:)?\/\//i.test(entry) || entry.startsWith("/")) return entry;
  if (!storage) return entry;
  if (/\.[A-Za-z0-9]+$/.test(storage.split(/[\\/]/).filter(Boolean).at(-1) ?? "")) return storage;
  return `${storage.replace(/[\\/]+$/, "")}/${entry}`;
}

function workspacePathFromURL(url: string | null) {
  if (!url) return null;
  try {
    const parsed = new URL(url, "http://mycelis.local");
    if (!parsed.pathname.endsWith("/api/v1/workspace/files/view")) return null;
    return parsed.searchParams.get("path");
  } catch {
    return null;
  }
}

function proofLabel(output: TeamOutputRef) {
  const value = output.proof_id || output.proof_ref || output.validation_ref;
  if (!value) return null;
  return value.length > 10 ? value.slice(0, 8) : value;
}

function outputIcon(kind: string, path?: string | null) {
  const value = `${kind} ${path ?? ""}`.toLowerCase();
  if (/\.(png|jpe?g|webp|gif|avif)$/.test(value) || value.includes("media") || value.includes("image")) {
    return <ImageIcon className="h-4 w-4" />;
  }
  if (value.includes("package") || value.includes("project")) {
    return <PackageOpen className="h-4 w-4" />;
  }
  return <FileText className="h-4 w-4" />;
}
