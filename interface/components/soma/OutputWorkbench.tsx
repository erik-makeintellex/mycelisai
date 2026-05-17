"use client";

import { Check, ExternalLink, Quote, ShieldCheck } from "lucide-react";
import { useState } from "react";
import type { ChatArtifactRef, ExecutionSummaryData, ExecutionSummaryItem } from "@/store/useCortexStore";
import ExecutionSummaryMediaPreview from "./ExecutionSummaryMediaPreview";
import {
  artifactOutputItems,
  asItems,
  itemText,
  itemUrl,
} from "./ExecutionSummaryCardModel";
import OutputAccessActions from "./OutputAccessActions";

export type OutputWorkbenchItem = {
  text: string;
  url: string | null;
};

export function projectPackageOutputs(outputs: ExecutionSummaryData["outputs"]) {
  return asItems(outputs).filter((item): item is ExecutionSummaryItem => (
    typeof item !== "string" && item.kind === "project_package"
  ));
}

export function outputWorkbenchItems(summary?: ExecutionSummaryData, artifacts?: ChatArtifactRef[]) {
  const directOutputs = asItems(summary?.outputs)
    .filter((item) => typeof item === "string" || item.kind !== "project_package")
    .map((item) => ({ text: itemText(item), url: itemUrl(item) }))
    .filter((item): item is OutputWorkbenchItem => Boolean(item.text));
  const artifactOutputs = artifactOutputItems(artifacts);

  return [
    ...directOutputs,
    ...artifactOutputs.filter((artifact) => !directOutputs.some((output) => output.text === artifact.text)),
  ];
}

function quotedOutputText(output: OutputWorkbenchItem) {
  return output.url ? `> ${output.text}\n${output.url}` : `> ${output.text}`;
}

export function OutputWorkbench({
  outputs,
  projectPackages,
  emptyMessage = "Soma outputs will appear here when a run, package, or retained artifact is available.",
  projectOpenLabel = "Open output",
}: {
  outputs: OutputWorkbenchItem[];
  projectPackages?: ExecutionSummaryItem[];
  emptyMessage?: string;
  projectOpenLabel?: string;
}) {
  const [copiedOutputKey, setCopiedOutputKey] = useState<string | null>(null);
  const packages = projectPackages ?? [];
  const hasOutputs = outputs.length > 0 || packages.length > 0;

  const copyOutputQuote = async (output: OutputWorkbenchItem, key: string) => {
    await navigator.clipboard.writeText(quotedOutputText(output));
    setCopiedOutputKey(key);
    window.setTimeout(() => setCopiedOutputKey((current) => current === key ? null : current), 1200);
  };

  if (!hasOutputs) {
    return (
      <div className="rounded-lg border border-cortex-border bg-cortex-bg p-3 text-sm leading-6 text-cortex-text-muted">
        {emptyMessage}
      </div>
    );
  }

  return (
    <div className="space-y-3" data-testid="output-workbench">
      {packages.length > 0 ? (
        <div className="space-y-2">
          {packages.map((project, index) => {
            const title = itemText(project) ?? "Project package";
            const href = itemUrl(project);
            const folder = project.folder ?? project.entrypoint ?? null;
            const files = project.files ?? [];

            return (
              <article key={`${title}-${index}`} className="rounded-lg border border-cortex-border/70 bg-cortex-bg px-3 py-2">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div className="min-w-0">
                    <div className="truncate text-sm font-semibold text-cortex-text-main">{title}</div>
                    {project.summary ? <div className="text-xs leading-5 text-cortex-text-muted">{project.summary}</div> : null}
                  </div>
                  <OutputAccessActions label={title} url={href} storagePath={folder} openLabel={projectOpenLabel} />
                </div>
                {(project.entrypoint || folder) ? (
                  <div className="mt-2 flex flex-wrap gap-1.5 text-[10px] font-mono text-cortex-text-muted">
                    {project.entrypoint ? <span>entry: {project.entrypoint}</span> : null}
                    {folder ? <span>folder: {folder}</span> : null}
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
              </article>
            );
          })}
        </div>
      ) : null}
      {outputs.length > 0 ? (
        <div className="flex flex-wrap gap-x-3 gap-y-2">
          {outputs.map((output, index) => {
            const key = `${output.text}-${output.url ?? "text"}-${index}`;
            const copied = copiedOutputKey === key;
            return (
              <span key={key} className="inline-flex max-w-full items-center gap-1">
                {output.url ? (
                  <a href={output.url} target="_blank" rel="noopener noreferrer" className="inline-flex min-w-0 items-center gap-1 text-cortex-primary hover:underline">
                    <span className="truncate">{output.text}</span>
                    <ExternalLink className="h-3 w-3 shrink-0" />
                  </a>
                ) : (
                  <span className="min-w-0 truncate text-sm text-cortex-text-main">{output.text}</span>
                )}
                <OutputAccessActions label={output.text} url={output.url} />
                <button
                  type="button"
                  onClick={() => void copyOutputQuote(output, key)}
                  className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-lg border border-cortex-border/70 text-cortex-text-muted transition-colors hover:border-cortex-info/40 hover:bg-cortex-info/10 hover:text-cortex-info"
                  title={copied ? "Copied output quote" : "Copy output quote"}
                  aria-label={copied ? "Copied output quote" : `Copy output quote for ${output.text}`}
                >
                  {copied ? <Check className="h-3.5 w-3.5" /> : <Quote className="h-3.5 w-3.5" />}
                </button>
              </span>
            );
          })}
        </div>
      ) : null}
      <ExecutionSummaryMediaPreview outputs={outputs} />
    </div>
  );
}
