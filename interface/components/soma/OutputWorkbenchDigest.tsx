"use client";

import type { ExecutionSummaryItem } from "@/store/useCortexStore";
import { itemText, itemUrl } from "./ExecutionSummaryCardModel";
import OutputAccessActions, { workspacePathFromOutputUrl } from "./OutputAccessActions";
import type { OutputWorkbenchItem } from "./OutputWorkbench";
import { projectPackageOpenPath, projectPackageRevealPath, workspaceFileHref } from "@/lib/outputPackageModel";

export type OutputWorkbenchDigest = {
  text: string;
  url: string | null;
  storagePath?: string | null;
  count: number;
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
      count: outputs.length + packages.length,
    };
  }

  if (!primaryPackage) return null;

  return {
    text: itemText(primaryPackage) ?? "Project package",
    url: itemUrl(primaryPackage) ?? workspaceFileHref(projectPackageOpenPath({
      folder: primaryPackage.folder,
      entrypoint: primaryPackage.entrypoint,
      filePath: primaryPackage.path,
    })),
    storagePath: projectPackageRevealPath({
      folder: primaryPackage.folder,
      entrypoint: primaryPackage.entrypoint,
      filePath: primaryPackage.path,
    }),
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

  return (
    <aside
      className="min-w-0"
      data-testid="soma-workbench-output-digest"
      aria-label="Latest output"
    >
      <div className="flex min-w-0 items-center justify-between gap-2">
        <div className="min-w-0 text-xs">
          <span className="font-semibold text-cortex-text-main">Latest: </span>
          <span className="font-semibold text-cortex-text-main">{digest.text}</span>
          {workspacePath && workspacePath !== digest.text ? (
            <span className="sr-only"> Workspace path: {workspacePath}</span>
          ) : null}
        </div>
        <OutputAccessActions
          label={digest.text}
          url={digest.url}
          storagePath={digest.storagePath}
          openLabel="Open file"
          folderLabel="Open folder"
        />
      </div>
    </aside>
  );
}
