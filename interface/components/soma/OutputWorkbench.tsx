"use client";

import { Check, ExternalLink, Quote, ShieldCheck } from "lucide-react";
import { useState } from "react";
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

export type OutputWorkbenchItem = {
  text: string;
  url: string | null;
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
    .map((item) => ({
      text: itemText(item),
      url: itemUrl(item),
      ...(typeof item !== "string" && item.proof ? { proof: item.proof } : {}),
      ...(typeof item !== "string" && item.proof_artifact_id ? { proofArtifactId: item.proof_artifact_id } : {}),
    }))
    .filter((item): item is OutputWorkbenchItem => Boolean(item.text));
  const artifactOutputs = artifactOutputItems(artifacts);

  return [
    ...directOutputs,
    ...artifactOutputs.filter((artifact) => !directOutputs.some((output) => output.text === artifact.text)),
  ];
}

export function teamOutputWorkbenchItems(outputRefs: TeamOutputRef[]): OutputWorkbenchItem[] {
  return outputRefs
    .filter((output) => output.kind !== "project_package" && !output.entrypoint)
    .map((output) => ({
      text: output.label?.trim() || "Team output",
      url: outputUrl(output.storage_ref),
      ...(output.proof ? { proof: output.proof } : {}),
      ...(output.proof_id ? { proofArtifactId: output.proof_id } : {}),
    }))
    .filter((item): item is OutputWorkbenchItem => Boolean(item.text));
}

export function teamOutputProjectPackages(outputRefs: TeamOutputRef[]): ExecutionSummaryItem[] {
  return outputRefs
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

function outputUrl(storageRef?: string | null): string | null {
  return normalizeWorkspaceOutputUrl(storageRef);
}

function quotedOutputText(output: OutputWorkbenchItem) {
  return output.url ? `> ${output.text}\n${output.url}` : `> ${output.text}`;
}

function shortHash(value?: string) {
  return value && value.length > 12 ? value.slice(0, 12) : value;
}

function proofStatusLabel(value?: string) {
  if (!value) return null;
  return value.replace(/[_-]+/g, " ");
}

function OutputProofBadges({ proof, proofArtifactId }: { proof?: OutputProofEnvelope; proofArtifactId?: string }) {
  const pathStatus = proofStatusLabel(proof?.path_boundary_status);
  const readbackStatus = proofStatusLabel(proof?.readback_status);
  const checksum = shortHash(proof?.checksum);
  if (!proof && !proofArtifactId) return null;
  return (
    <span className="inline-flex max-w-full flex-wrap items-center gap-1 text-[10px] leading-4 text-cortex-text-muted">
      {pathStatus ? <span className="rounded border border-cortex-success/40 px-1 text-cortex-success">path {pathStatus}</span> : null}
      {readbackStatus ? <span className="rounded border border-cortex-success/40 px-1 text-cortex-success">readback {readbackStatus}</span> : null}
      {checksum ? <span className="rounded border border-cortex-border/70 px-1 font-mono">sha256 {checksum}</span> : null}
      {proofArtifactId && !checksum ? <span className="rounded border border-cortex-border/70 px-1 font-mono">proof {shortHash(proofArtifactId)}</span> : null}
    </span>
  );
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
  const primaryOutputIndex = outputs.findIndex((output) => Boolean(output.url));
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
                  <OutputAccessActions label={title} url={href} storagePath={folder} openLabel={projectOpenLabel} folderLabel="Open folder" />
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
                <div className="mt-2">
                  <OutputProofBadges proof={project.proof} proofArtifactId={project.proof_artifact_id} />
                </div>
              </article>
            );
          })}
        </div>
      ) : null}
      {primaryOutput ? (
        <article className="rounded-lg border border-cortex-primary/30 bg-cortex-primary/10 px-3 py-2">
          <div className="flex flex-wrap items-start justify-between gap-2">
            <div className="min-w-0">
              <div className="text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-primary">Latest output</div>
              <div className="mt-1 truncate text-sm font-semibold text-cortex-text-main">{primaryOutput.text}</div>
            </div>
            <div className="flex shrink-0 flex-wrap items-center gap-1">
              <OutputAccessActions label={primaryOutput.text} url={primaryOutput.url} openLabel="Open file" folderLabel="Open folder" />
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
          <div className="mt-2">
            <OutputProofBadges proof={primaryOutput.proof} proofArtifactId={primaryOutput.proofArtifactId} />
          </div>
        </article>
      ) : null}
      {secondaryOutputs.length > 0 ? (
        <details className="rounded-lg border border-cortex-border/70 bg-cortex-bg/50 px-3 py-2">
          <summary className="cursor-pointer text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
            Output details and proof
          </summary>
          <div className="mt-2 flex flex-wrap gap-x-3 gap-y-2">
          {secondaryOutputs.map((output, index) => {
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
                <OutputAccessActions label={output.text} url={output.url} folderLabel="Open folder" />
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
