"use client";

import type { ExecutionSummaryItem } from "@/store/useCortexStore";
import { itemText, itemUrl } from "./ExecutionSummaryCardModel";
import OutputAccessActions from "./OutputAccessActions";
import type { OutputWorkbenchItem } from "./OutputWorkbench";

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
  const primaryOutputIndex = outputs.findIndex((output) => Boolean(output.url));
  const normalizedPrimaryIndex = primaryOutputIndex >= 0 ? primaryOutputIndex : outputs.length > 0 ? 0 : -1;
  const primaryOutput = normalizedPrimaryIndex >= 0 ? outputs[normalizedPrimaryIndex] : null;

  if (primaryOutput) {
    return {
      text: primaryOutput.text,
      url: primaryOutput.url,
      count: outputs.length + packages.length,
    };
  }

  const primaryPackage = packages[0];
  if (!primaryPackage) return null;

  return {
    text: itemText(primaryPackage) ?? "Project package",
    url: itemUrl(primaryPackage),
    storagePath: primaryPackage.folder ?? primaryPackage.entrypoint ?? null,
    count: outputs.length + packages.length,
  };
}

export function OutputWorkbenchCompactDigest({ digest }: { digest: OutputWorkbenchDigest }) {
  return (
    <aside
      className="max-w-[min(82vw,360px)] rounded-xl border border-cortex-primary/30 bg-cortex-surface/95 px-3 py-2 shadow-lg shadow-black/10 backdrop-blur"
      data-testid="soma-workbench-output-digest"
      aria-label="Latest output"
    >
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="min-w-0">
          <div className="font-mono text-[10px] font-bold uppercase tracking-[0.16em] text-cortex-primary">
            Latest output
          </div>
          <div className="mt-0.5 truncate text-xs font-semibold text-cortex-text-main">
            {digest.text}
          </div>
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
