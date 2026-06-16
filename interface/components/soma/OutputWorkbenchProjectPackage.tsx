"use client";

import { FolderOpen, ShieldCheck } from "lucide-react";
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

export function OutputWorkbenchProjectPackage({
  project,
  index,
  projectOpenLabel,
}: {
  project: ExecutionSummaryItem;
  index: number;
  projectOpenLabel: string;
}) {
  const title = itemText(project) ?? "Project package";
  const openPath = projectPackageOpenPath({ folder: project.folder, entrypoint: project.entrypoint, filePath: project.path });
  const href = itemUrl(project) ?? workspaceFileHref(openPath);
  const folder = project.folder ?? null;
  const revealPath = projectPackageRevealPath({ folder: project.folder, entrypoint: project.entrypoint, filePath: project.path });
  const resourcesHref = projectPackageResourcesHref({ folder: project.folder, entrypoint: project.entrypoint, filePath: project.path });
  const files = project.files ?? [];

  return (
    <article key={`${title}-${index}`} className="rounded-lg border border-cortex-border/70 bg-cortex-bg px-3 py-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="min-w-0">
          <div className="truncate text-sm font-semibold text-cortex-text-main">{title}</div>
          {project.summary ? <div className="text-xs leading-5 text-cortex-text-muted">{project.summary}</div> : null}
        </div>
        <span className="inline-flex shrink-0 flex-wrap items-center gap-1">
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
          <OutputAccessActions label={title} url={href} storagePath={revealPath} openLabel={projectOpenLabel} folderLabel={OUTPUT_PACKAGE_FOLDER_LABEL} />
        </span>
      </div>
      {(project.entrypoint || folder) ? (
        <div className="mt-2 flex flex-wrap gap-1.5 text-[10px] text-cortex-text-muted">
          {folder ? <PackagePath label="Workspace folder" value={folder} /> : null}
          {project.entrypoint ? <PackagePath label="Open file" value={project.entrypoint} /> : null}
        </div>
      ) : null}
      {files.length > 0 ? (
        <div className="mt-2 flex flex-wrap gap-1">
          {files.map((file) => (
            <span key={file} className="rounded border border-cortex-border/60 px-1.5 py-0.5 text-[10px] font-mono text-cortex-text-muted">
              {file}
            </span>
          ))}
        </div>
      ) : null}
      {project.validation ? (
        <div className="mt-2 inline-flex items-start gap-1 text-xs leading-5 text-cortex-success">
          <ShieldCheck className="mt-0.5 h-3.5 w-3.5 shrink-0" />
          <span>{project.validation}</span>
        </div>
      ) : null}
      <OutputProofDetails proof={project.proof} proofArtifactId={project.proof_artifact_id} />
    </article>
  );
}

function PackagePath({ label, value }: { label: string; value: string }) {
  return (
    <span className="inline-flex min-w-0 items-center gap-1 rounded border border-cortex-border/60 bg-cortex-bg/70 px-1.5 py-0.5">
      <span>{label}</span>
      <code className="max-w-64 truncate font-mono text-cortex-text-main">{value}</code>
    </span>
  );
}
