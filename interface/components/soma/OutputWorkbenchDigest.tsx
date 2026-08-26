"use client";

import { FolderOpen, MessageSquareReply, ShieldCheck } from "lucide-react";
import { OutcomeHealthBadge } from "@/components/shared/OutcomeHealthBadge";
import type { OutcomeHealthState } from "@/lib/outcomeHealth";
import type { ExecutionSummaryItem } from "@/store/useCortexStore";
import { itemText, itemUrl } from "./ExecutionSummaryCardModel";
import OutputAccessActions, { workspacePathFromOutputUrl } from "./OutputAccessActions";
import type { OutputWorkbenchItem } from "./OutputWorkbench";
import {
  OUTPUT_PACKAGE_RESOURCES_LABEL,
  projectPackageOpenPath,
  projectPackageResourcesHref,
  projectPackageRevealPath,
  workspaceFileHref,
} from "@/lib/outputPackageModel";
import { requestSomaOutputContinuation } from "./outputContinuation";

export type OutputWorkbenchDigest = {
  text: string;
  url: string | null;
  storagePath?: string | null;
  count: number;
  isProjectPackage?: boolean;
  resourcesHref?: string | null;
  validation?: string | null;
  proofArtifactId?: string | null;
  replyReference?: string | null;
  health?: OutcomeHealthState;
};

export function outputWorkbenchDigest({
  outputs,
  projectPackages,
}: {
  outputs: OutputWorkbenchItem[];
  projectPackages?: ExecutionSummaryItem[];
}): OutputWorkbenchDigest | null {
  const packages = projectPackages ?? [];
  const primaryOutputIndex = preferredOutputIndex(outputs);
  const normalizedPrimaryIndex = primaryOutputIndex >= 0 ? primaryOutputIndex : outputs.length > 0 ? 0 : -1;
  const primaryOutput = normalizedPrimaryIndex >= 0 ? outputs[normalizedPrimaryIndex] : null;
  const primaryPackage = packages[0];

  if (primaryOutput && (!isGroupFolderOutput(primaryOutput) || !primaryPackage || !itemUrl(primaryPackage))) {
    const storagePath = primaryOutput.storagePath ?? workspacePathFromOutputUrl(primaryOutput.url);
    return {
      text: primaryOutput.text,
      url: primaryOutput.url,
      ...(storagePath ? { storagePath } : {}),
      proofArtifactId: primaryOutput.proofArtifactId ?? null,
      replyReference: storagePath ?? primaryOutput.url ?? null,
      health: "completed",
      count: outputs.length + packages.length,
    };
  }

  if (!primaryPackage) return null;
  const storagePath = projectPackageRevealPath({
    folder: primaryPackage.folder,
    entrypoint: primaryPackage.entrypoint,
    filePath: primaryPackage.path,
  });

  return {
    text: itemText(primaryPackage) ?? "Project package",
    url: itemUrl(primaryPackage) ?? workspaceFileHref(projectPackageOpenPath({
      folder: primaryPackage.folder,
      entrypoint: primaryPackage.entrypoint,
      filePath: primaryPackage.path,
    })),
    storagePath,
    isProjectPackage: true,
    resourcesHref: projectPackageResourcesHref({
      folder: primaryPackage.folder,
      entrypoint: primaryPackage.entrypoint,
      filePath: primaryPackage.path,
    }),
    validation: primaryPackage.validation ?? null,
    proofArtifactId: primaryPackage.proof_artifact_id ?? null,
    replyReference: storagePath ?? itemUrl(primaryPackage) ?? null,
    health: "completed",
    count: outputs.length + packages.length,
  };
}

function preferredOutputIndex(outputs: OutputWorkbenchItem[]) {
  const fileIndex = outputs.findIndex((output) => Boolean(output.url) && isFileLikeOutput(output));
  if (fileIndex >= 0) return fileIndex;
  const nonFolderIndex = outputs.findIndex((output) => Boolean(output.url) && !isGroupFolderOutput(output));
  if (nonFolderIndex >= 0) return nonFolderIndex;
  return outputs.findIndex((output) => Boolean(output.url));
}

function outputWorkspacePath(output: OutputWorkbenchItem) {
  return output.storagePath?.trim() || workspacePathFromOutputUrl(output.url) || "";
}

function isFileLikeOutput(output: OutputWorkbenchItem) {
  return /\.[a-z0-9]{1,8}$/i.test(outputWorkspacePath(output));
}

function isGroupFolderOutput(output: OutputWorkbenchItem) {
  return outputWorkspacePath(output).replace(/\\/g, "/").startsWith("groups/");
}

export function OutputWorkbenchCompactDigest({ digest }: { digest: OutputWorkbenchDigest }) {
  const workspacePath = digest.storagePath?.trim() || workspacePathFromOutputUrl(digest.url);
  const label = digest.isProjectPackage ? "App/package" : "Latest";
  const openLabel = digest.isProjectPackage ? "Open app" : "Open file";

  return (
    <aside
      className="min-w-0"
      data-testid="soma-workbench-output-digest"
      aria-label="Latest output"
    >
      <div className="flex min-w-0 flex-col items-start gap-2" data-testid="soma-output-digest-layout">
        <div className="min-w-0 max-w-full text-xs leading-5">
          <span className="font-semibold text-cortex-text-main">{label}: </span>
          <span className="font-semibold text-cortex-text-main">{digest.text}</span>
          <OutcomeHealthBadge health={digest.health ?? "completed"} className="ml-2 align-middle" />
          {workspacePath && workspacePath !== digest.text ? (
            <span className="sr-only"> Workspace path: {workspacePath}</span>
          ) : null}
          {digest.validation ? (
            <span className="mt-0.5 flex items-start gap-1 text-[11px] text-cortex-success">
              <ShieldCheck className="mt-0.5 h-3 w-3 shrink-0" />
              <span className="line-clamp-2">{digest.validation}</span>
            </span>
          ) : null}
        </div>
        <div className="w-full">
          <OutputAccessActions
            label={digest.text}
            url={digest.url}
            storagePath={digest.storagePath}
            openLabel={openLabel}
            folderLabel="Open folder"
            primary
            showFolder={false}
            openInCanvas
            proofArtifactId={digest.proofArtifactId}
          />
        </div>
        {(digest.resourcesHref || digest.replyReference || digest.url || workspacePath) ? (
          <details className="w-full border-t border-cortex-border/60 pt-2">
            <summary className="cursor-pointer text-[10px] font-semibold uppercase tracking-[0.08em] text-cortex-text-muted hover:text-cortex-text-main">
              Details and follow-up
            </summary>
            <div className="mt-2 flex w-full flex-wrap items-center gap-1.5">
              {digest.resourcesHref ? (
                <a
                  href={digest.resourcesHref}
                  className="inline-flex h-7 items-center gap-1.5 rounded-lg border border-cortex-border/80 bg-cortex-bg/70 px-2.5 text-[11px] font-semibold text-cortex-text-main transition-colors hover:border-cortex-primary/45 hover:bg-cortex-primary/10 hover:text-cortex-primary"
                  title={`Browse ${digest.text} in Resources`}
                  aria-label={`Open ${digest.text} in Resources`}
                >
                  <FolderOpen className="h-3 w-3" />
                  {OUTPUT_PACKAGE_RESOURCES_LABEL}
                </a>
              ) : null}
              {digest.replyReference || digest.url || workspacePath ? (
                <button
                  type="button"
                  onClick={() => requestSomaOutputContinuation({
                    title: digest.text,
                    reference: digest.replyReference ?? workspacePath ?? digest.url,
                    proof: digest.proofArtifactId,
                  })}
                  className="inline-flex h-7 items-center gap-1.5 rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-2.5 text-[11px] font-semibold text-cortex-primary transition-colors hover:border-cortex-primary/60 hover:bg-cortex-primary/15"
                  title={`Reply to ${digest.text} in Soma`}
                  aria-label={`Reply to ${digest.text} in Soma`}
                >
                  <MessageSquareReply className="h-3 w-3" />
                  Reply
                </button>
              ) : null}
            </div>
          </details>
        ) : null}
      </div>
    </aside>
  );
}
