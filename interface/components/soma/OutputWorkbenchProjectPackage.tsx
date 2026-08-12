"use client";

import { FolderOpen, MessageSquareReply, ShieldCheck } from "lucide-react";
import { OutcomeHealthBadge } from "@/components/shared/OutcomeHealthBadge";
import type { ExecutionSummaryItem } from "@/store/useCortexStore";
import {
  OUTPUT_PACKAGE_FOLDER_LABEL,
  OUTPUT_PACKAGE_RESOURCES_LABEL,
  projectPackageOpenPath,
  projectPackageResourcesHref,
  projectPackageRevealPath,
  workspaceFileHref,
} from "@/lib/outputPackageModel";
import { itemText, itemUrl } from "./ExecutionSummaryCardModel";
import OutputAccessActions from "./OutputAccessActions";
import { OutputProofDetails } from "./OutputWorkbenchProofDetails";
import { requestSomaOutputContinuation } from "./outputContinuation";

export function OutputWorkbenchProjectPackage({
  project,
  index,
  projectOpenLabel,
  isPrimary = false,
}: {
  project: ExecutionSummaryItem;
  index: number;
  projectOpenLabel: string;
  isPrimary?: boolean;
}) {
  const title = itemText(project) ?? "Project package";
  const openPath = projectPackageOpenPath({ folder: project.folder, entrypoint: project.entrypoint, filePath: project.path });
  const href = itemUrl(project) ?? workspaceFileHref(openPath);
  const folder = project.folder ?? null;
  const revealPath = projectPackageRevealPath({ folder: project.folder, entrypoint: project.entrypoint, filePath: project.path });
  const resourcesHref = projectPackageResourcesHref({ folder: project.folder, entrypoint: project.entrypoint, filePath: project.path });
  const files = project.files ?? [];
  const primaryOpenLabel = projectOpenLabel === "Open file"
    ? project.entrypoint?.toLowerCase().endsWith(".html") ? "Open app" : "Open output"
    : projectOpenLabel;

  return (
    <article
      key={`${title}-${index}`}
      className={`rounded-lg border px-3 py-3 ${
        isPrimary
          ? "border-cortex-primary/45 bg-cortex-primary/10"
          : "border-cortex-border/70 bg-cortex-bg"
      }`}
    >
      <div className="space-y-3">
        <div className="min-w-0">
          {isPrimary ? (
            <div className="mb-1 text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-primary">
              Latest output
            </div>
          ) : null}
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            <span className="truncate text-sm font-semibold text-cortex-text-main">{title}</span>
            <OutcomeHealthBadge health="completed" />
          </div>
          {project.summary ? <div className="text-xs leading-5 text-cortex-text-muted">{project.summary}</div> : null}
        </div>
        <div className="flex w-full min-w-0 flex-wrap items-center gap-2" data-testid="project-package-actions">
          <OutputAccessActions
            label={title}
            url={href}
            storagePath={revealPath}
            openLabel={primaryOpenLabel}
            folderLabel={OUTPUT_PACKAGE_FOLDER_LABEL}
            primary={isPrimary}
            showFolder={false}
          />
        </div>
      </div>
      <details className="mt-3 border-t border-cortex-border/70 pt-2">
        <summary className="cursor-pointer text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
          Details and proof
        </summary>
        <div className="mt-3 min-w-0 space-y-3">
          <div className="flex w-full min-w-0 flex-wrap items-center gap-1.5">
            <OutputAccessActions
              label={title}
              url={href}
              storagePath={revealPath}
              folderLabel={OUTPUT_PACKAGE_FOLDER_LABEL}
              showOpen={false}
            />
            {resourcesHref ? (
              <a
                href={resourcesHref}
                className="inline-flex h-7 items-center gap-1.5 rounded-lg border border-cortex-border/80 bg-cortex-bg/70 px-2.5 text-[11px] font-semibold text-cortex-text-main transition-colors hover:border-cortex-primary/45 hover:bg-cortex-primary/10 hover:text-cortex-primary"
                title={`Browse ${title} in Resources`}
                aria-label={`Open ${title} in Resources`}
              >
                <FolderOpen className="h-3 w-3" />
                {OUTPUT_PACKAGE_RESOURCES_LABEL}
              </a>
            ) : null}
          <button
            type="button"
            onClick={() => requestSomaOutputContinuation({
              title,
              reference: revealPath || href || folder,
              proof: project.proof_artifact_id,
            })}
            className="inline-flex h-7 items-center gap-1.5 rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-2.5 text-[11px] font-semibold text-cortex-primary transition-colors hover:border-cortex-primary/60 hover:bg-cortex-primary/15"
            title={`Reply to ${title} in Soma`}
            aria-label={`Reply to ${title} in Soma`}
          >
            <MessageSquareReply className="h-3 w-3" />
            Reply
          </button>
          </div>
          {(project.entrypoint || folder) ? (
            <div className="flex flex-wrap gap-1.5 text-[10px] text-cortex-text-muted">
              {folder ? <PackagePath label="Workspace folder" value={folder} /> : null}
              {project.entrypoint ? <PackagePath label="Open file" value={project.entrypoint} /> : null}
            </div>
          ) : null}
          {files.length > 0 ? (
            <div className="flex flex-wrap gap-1">
              {files.map((file) => (
                <span key={file} className="rounded border border-cortex-border/60 px-1.5 py-0.5 text-[10px] font-mono text-cortex-text-muted">
                  {file}
                </span>
              ))}
            </div>
          ) : null}
          {project.validation ? (
            <div className="inline-flex items-start gap-1 text-xs leading-5 text-cortex-success">
              <ShieldCheck className="mt-0.5 h-3.5 w-3.5 shrink-0" />
              <span>{project.validation}</span>
            </div>
          ) : null}
          <OutputProofDetails proof={project.proof} proofArtifactId={project.proof_artifact_id} />
        </div>
      </details>
    </article>
  );
}

function PackagePath({ label, value }: { label: string; value: string }) {
  return (
    <span className="inline-flex max-w-full min-w-0 items-center gap-1 rounded border border-cortex-border/60 bg-cortex-bg/70 px-1.5 py-0.5">
      <span className="shrink-0">{label}</span>
      <code className="min-w-0 flex-1 truncate font-mono text-cortex-text-main">{value}</code>
    </span>
  );
}
