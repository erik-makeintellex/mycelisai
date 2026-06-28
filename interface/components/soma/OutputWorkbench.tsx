"use client";

import { Check, ExternalLink, MessageSquareReply, Quote } from "lucide-react";
import { useState } from "react";
import { sortTeamOutputRefsNewestFirst } from "@/components/teams/teamWorkProjection";
import type { ChatArtifactRef, ExecutionSummaryData, ExecutionSummaryItem, OutputProofEnvelope, TeamOutputRef } from "@/store/useCortexStore";
import ExecutionSummaryMediaPreview from "./ExecutionSummaryMediaPreview";
import {
  artifactOutputItems,
  asItems,
  itemText,
  itemUrl,
  normalizeWorkspaceOutputUrl,
} from "./ExecutionSummaryCardModel";
import OutputAccessActions from "./OutputAccessActions";
import { itemWorkspacePath, outputWorkspacePath, OutputPathHint } from "./OutputWorkbenchPathHint";
import { OutputProofBadges, OutputProofDetails } from "./OutputWorkbenchProofDetails";
import { OutputWorkbenchProjectPackage } from "./OutputWorkbenchProjectPackage";
import { OUTPUT_PACKAGE_OPEN_LABEL } from "@/lib/outputPackageModel";
import { requestSomaOutputContinuation } from "./outputContinuation";

export type OutputWorkbenchItem = {
  text: string;
  url: string | null;
  storagePath?: string;
  proof?: OutputProofEnvelope;
  proofArtifactId?: string;
};

export function projectPackageOutputs(outputs: ExecutionSummaryData["outputs"]) {
  return asItems(outputs).filter((item): item is ExecutionSummaryItem => (
    typeof item !== "string" && item.kind === "project_package"
  ));
}

export function outputWorkbenchItems(summary?: ExecutionSummaryData, artifacts?: ChatArtifactRef[]) {
  const directOutputs = asItems(summary?.outputs)
    .filter((item) => typeof item === "string" || item.kind !== "project_package")
    .map((item) => {
      const storagePath = typeof item !== "string" ? itemWorkspacePath(item) : null;
      return {
        text: itemText(item),
        url: itemUrl(item),
        ...(storagePath ? { storagePath } : {}),
        ...(typeof item !== "string" && item.proof ? { proof: item.proof } : {}),
        ...(typeof item !== "string" && item.proof_artifact_id ? { proofArtifactId: item.proof_artifact_id } : {}),
      };
    })
    .filter((item): item is OutputWorkbenchItem => Boolean(item.text));
  const artifactOutputs = artifactOutputItems(artifacts);

  return [
    ...directOutputs,
    ...artifactOutputs.filter((artifact) => !directOutputs.some((output) => output.text === artifact.text)),
  ];
}

export function teamOutputWorkbenchItems(outputRefs: TeamOutputRef[]): OutputWorkbenchItem[] {
  return sortTeamOutputRefsNewestFirst(outputRefs)
    .filter((output) => output.kind !== "project_package" && !output.entrypoint)
    .map((output) => ({
      text: output.label?.trim() || "Team output",
      url: outputUrl(output.storage_ref),
      ...(output.storage_ref ? { storagePath: output.storage_ref } : {}),
      ...(output.proof ? { proof: output.proof } : {}),
      ...(output.proof_id ? { proofArtifactId: output.proof_id } : {}),
    }))
    .filter((item): item is OutputWorkbenchItem => Boolean(item.text));
}

export function teamOutputProjectPackages(outputRefs: TeamOutputRef[]): ExecutionSummaryItem[] {
  return sortTeamOutputRefsNewestFirst(outputRefs)
    .filter((output) => output.kind === "project_package" || Boolean(output.entrypoint))
    .map((output) => ({
      kind: "project_package",
      title: output.label?.trim() || "Team output package",
      summary: output.proof_ref || output.validation_ref ? "Proof and validation links are available in the retained team-work record." : undefined,
      folder: output.storage_ref || undefined,
      entrypoint: output.entrypoint || undefined,
      validation: output.validation_ref || output.proof_ref ? "Linked proof or validation record" : undefined,
      proof: output.proof,
      proof_artifact_id: output.proof_id,
    }));
}

export function mergeOutputWorkbenchItems(...groups: OutputWorkbenchItem[][]): OutputWorkbenchItem[] {
  const seen = new Set<string>();
  return groups.flat().filter((item) => {
    const key = `${item.text}-${item.url ?? ""}`;
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}

export function actionableOutputWorkbenchItems(outputs: OutputWorkbenchItem[]): OutputWorkbenchItem[] {
  return outputs.filter((output) => Boolean(output.url || output.proof || output.proofArtifactId));
}

function preferredOutputIndex(outputs: OutputWorkbenchItem[]) {
  const fileIndex = outputs.findIndex((output) => Boolean(output.url) && isFileLikeOutput(output));
  if (fileIndex >= 0) return fileIndex;
  const nonFolderIndex = outputs.findIndex((output) => Boolean(output.url) && !isGroupFolderOutput(output));
  if (nonFolderIndex >= 0) return nonFolderIndex;
  return outputs.findIndex((output) => Boolean(output.url));
}

function isFileLikeOutput(output: OutputWorkbenchItem) { return /\.[a-z0-9]{1,8}$/i.test(outputWorkspacePath(output) ?? ""); }

function isGroupFolderOutput(output: OutputWorkbenchItem) { return (outputWorkspacePath(output) ?? "").replace(/\\/g, "/").startsWith("groups/"); }

function outputUrl(storageRef?: string | null): string | null {
  return normalizeWorkspaceOutputUrl(storageRef);
}

function quotedOutputText(output: OutputWorkbenchItem) {
  return output.url ? `> ${output.text}\n${output.url}` : `> ${output.text}`;
}

function outputContinuationReference(output: OutputWorkbenchItem) {
  return output.storagePath || output.url || output.text;
}

export function OutputWorkbench({
  outputs,
  projectPackages,
  emptyMessage = "Soma outputs will appear here when a run, package, or retained artifact is available.",
  projectOpenLabel = OUTPUT_PACKAGE_OPEN_LABEL,
}: {
  outputs: OutputWorkbenchItem[];
  projectPackages?: ExecutionSummaryItem[];
  emptyMessage?: string;
  projectOpenLabel?: string;
}) {
  const [copiedOutputKey, setCopiedOutputKey] = useState<string | null>(null);
  const packages = projectPackages ?? [];
  const hasOutputs = outputs.length > 0 || packages.length > 0;
  const primaryOutputIndex = preferredOutputIndex(outputs);
  const normalizedPrimaryIndex = primaryOutputIndex >= 0 ? primaryOutputIndex : outputs.length > 0 ? 0 : -1;
  const primaryOutput = normalizedPrimaryIndex >= 0 ? outputs[normalizedPrimaryIndex] : null;
  const secondaryOutputs = outputs.filter((_, index) => index !== normalizedPrimaryIndex);

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
          {packages.map((project, index) => (
            <OutputWorkbenchProjectPackage
              key={`${itemText(project) ?? "Project package"}-${index}`}
              project={project}
              index={index}
              projectOpenLabel={projectOpenLabel}
            />
          ))}
        </div>
      ) : null}
      {primaryOutput ? (
        <article className="rounded-lg border border-cortex-primary/50 bg-cortex-primary/15 px-3 py-2 shadow-[0_0_0_1px_rgba(115,92,255,0.12)]" aria-label="Latest output">
          <div className="flex flex-wrap items-start justify-between gap-2">
            <div className="min-w-0">
              <div className="text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-primary">Latest output</div>
              <div className="mt-1 truncate text-sm font-semibold text-cortex-text-main">{primaryOutput.text}</div>
              <p className="mt-1 text-xs leading-5 text-cortex-text-muted">Use Open file to view it, or Open folder to show it in the workspace.</p>
              <OutputPathHint storagePath={primaryOutput.storagePath} url={primaryOutput.url} />
            </div>
            <div className="flex shrink-0 flex-wrap items-center gap-1">
              <OutputAccessActions label={primaryOutput.text} url={primaryOutput.url} storagePath={primaryOutput.storagePath} openLabel="Open file" folderLabel="Open folder" />
              <button
                type="button"
                onClick={() => requestSomaOutputContinuation({
                  title: primaryOutput.text,
                  reference: outputContinuationReference(primaryOutput),
                  proof: primaryOutput.proofArtifactId,
                })}
                className="inline-flex h-7 items-center gap-1.5 rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-2.5 text-[11px] font-semibold text-cortex-primary transition-colors hover:border-cortex-primary/60 hover:bg-cortex-primary/15"
                title={`Reply to ${primaryOutput.text} in Soma`}
                aria-label={`Reply to ${primaryOutput.text} in Soma`}
              >
                <MessageSquareReply className="h-3 w-3" />
                Reply
              </button>
              <button
                type="button"
                onClick={() => void copyOutputQuote(primaryOutput, `primary-${primaryOutput.text}-${primaryOutput.url ?? "text"}`)}
                className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-lg border border-cortex-border/70 text-cortex-text-muted transition-colors hover:border-cortex-info/40 hover:bg-cortex-info/10 hover:text-cortex-info"
                title={copiedOutputKey === `primary-${primaryOutput.text}-${primaryOutput.url ?? "text"}` ? "Copied output quote" : "Copy output quote"}
                aria-label={copiedOutputKey === `primary-${primaryOutput.text}-${primaryOutput.url ?? "text"}` ? "Copied output quote" : `Copy output quote for ${primaryOutput.text}`}
              >
                {copiedOutputKey === `primary-${primaryOutput.text}-${primaryOutput.url ?? "text"}` ? <Check className="h-3.5 w-3.5" /> : <Quote className="h-3.5 w-3.5" />}
              </button>
            </div>
          </div>
          <OutputProofDetails proof={primaryOutput.proof} proofArtifactId={primaryOutput.proofArtifactId} />
        </article>
      ) : null}
      {secondaryOutputs.length > 0 ? (
        <details className="rounded-lg border border-cortex-border/70 bg-cortex-bg/50 px-3 py-2">
          <summary className="cursor-pointer text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
            More outputs and verification
          </summary>
          <div className="mt-2 flex flex-wrap gap-x-3 gap-y-2">
            {secondaryOutputs.map((output, index) => {
              const key = `${output.text}-${output.url ?? "text"}-${index}`;
              const copied = copiedOutputKey === key;
              return (
                <span key={key} className="inline-flex max-w-full items-center gap-1">
                  <span className="min-w-0">
                    {output.url ? (
                      <a href={output.url} target="_blank" rel="noopener noreferrer" className="inline-flex min-w-0 items-center gap-1 text-cortex-primary hover:underline">
                        <span className="truncate">{output.text}</span>
                        <ExternalLink className="h-3 w-3 shrink-0" />
                      </a>
                    ) : (
                      <span className="min-w-0 truncate text-sm text-cortex-text-main">{output.text}</span>
                    )}
                    <OutputPathHint storagePath={output.storagePath} url={output.url} />
                  </span>
                  <OutputAccessActions label={output.text} url={output.url} storagePath={output.storagePath} openLabel="Open file" folderLabel="Open folder" />
                  <button
                    type="button"
                    onClick={() => requestSomaOutputContinuation({
                      title: output.text,
                      reference: outputContinuationReference(output),
                      proof: output.proofArtifactId,
                    })}
                    className="inline-flex h-6 items-center gap-1 rounded-lg border border-cortex-primary/30 bg-cortex-primary/10 px-2 text-[10px] font-semibold text-cortex-primary transition-colors hover:border-cortex-primary/55 hover:bg-cortex-primary/15"
                    title={`Reply to ${output.text} in Soma`}
                    aria-label={`Reply to ${output.text} in Soma`}
                  >
                    <MessageSquareReply className="h-3 w-3" />
                    Reply
                  </button>
                  <OutputProofBadges proof={output.proof} proofArtifactId={output.proofArtifactId} />
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
        </details>
      ) : null}
      <ExecutionSummaryMediaPreview outputs={outputs} />
    </div>
  );
}
