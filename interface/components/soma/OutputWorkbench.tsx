"use client";
import { Check, ExternalLink, MessageSquareReply, Quote } from "lucide-react";
import { useState } from "react";
import { sortTeamOutputRefsNewestFirst } from "@/components/teams/teamWorkProjection";
import { OutcomeHealthBadge } from "@/components/shared/OutcomeHealthBadge";
import type { ChatArtifactRef, ExecutionSummaryData, ExecutionSummaryItem, OutputProofEnvelope, TeamOutputRef, WorkOutputContractData } from "@/store/useCortexStore";
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
import { OUTPUT_PACKAGE_OPEN_LABEL, projectPackageOpenPath } from "@/lib/outputPackageModel";
import { requestSomaOutputContinuation } from "./outputContinuation";
import { deliverablePresentation } from "@/lib/deliverablePresentation";

export type OutputWorkbenchItem = {
  text: string;
  url: string | null;
  storagePath?: string;
  proof?: OutputProofEnvelope;
  proofArtifactId?: string;
  kind?: string;
  type?: string;
  contentType?: string;
  entrypoint?: string;
};
export function projectPackageOutputs(outputs: ExecutionSummaryData["outputs"]) {
  return asItems(outputs).filter((item): item is ExecutionSummaryItem => (
    typeof item !== "string" && item.kind === "project_package" && isUserDeliverableSummaryItem(item)
  ));
}
export function outputWorkbenchItems(summary?: ExecutionSummaryData, artifacts?: ChatArtifactRef[]) {
  const directOutputs = asItems(summary?.outputs)
    .filter((item) => typeof item === "string" || (item.kind !== "project_package" && isUserDeliverableSummaryItem(item)))
    .map((item) => {
      const storagePath = typeof item !== "string" ? itemWorkspacePath(item) : null;
      return {
        text: itemText(item),
        url: itemUrl(item),
        ...(storagePath ? { storagePath } : {}),
        ...(typeof item !== "string" && item.proof ? { proof: item.proof } : {}),
        ...(typeof item !== "string" && item.proof_artifact_id ? { proofArtifactId: item.proof_artifact_id } : {}),
        ...(typeof item !== "string" && item.kind ? { kind: item.kind } : {}),
        ...(typeof item !== "string" && item.type ? { type: item.type } : {}),
        ...(typeof item !== "string" && item.content_type ? { contentType: item.content_type } : {}),
        ...(typeof item !== "string" && item.entrypoint ? { entrypoint: item.entrypoint } : {}),
      };
    })
    .filter((item): item is OutputWorkbenchItem => Boolean(item.text));
  const artifactOutputs = artifactOutputItems(artifacts);

  return [
    ...directOutputs,
    ...artifactOutputs.filter((artifact) => !directOutputs.some((output) => output.text === artifact.text)),
  ];
}
function isUserDeliverableSummaryItem(item: ExecutionSummaryItem) {
  const outputClass = item.output_class?.trim().toLowerCase();
  if (outputClass) return ["user_deliverable", "deliverable", "output", "final"].includes(outputClass);
  const reference = [item.path, item.folder, item.entrypoint, item.url, item.href]
    .filter(Boolean)
    .join("/")
    .replace(/\\/g, "/")
    .toLowerCase();
  return !reference.includes("/planning/")
    && !reference.endsWith("/team_evocation.md")
    && !reference.endsWith("/research_council_handoff.md");
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
      ...(output.kind ? { kind: output.kind } : {}),
      ...(output.entrypoint ? { entrypoint: output.entrypoint } : {}),
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
  outputContract,
}: {
  outputs: OutputWorkbenchItem[];
  projectPackages?: ExecutionSummaryItem[];
  emptyMessage?: string;
  projectOpenLabel?: string;
  outputContract?: WorkOutputContractData;
}) {
  const [copiedOutputKey, setCopiedOutputKey] = useState<string | null>(null);
  const packages = projectPackages ?? [];
  const primaryPackage = packages[0] ?? null;
  const secondaryPackages = packages.slice(1);
  const packageReferences = new Set(packages.flatMap(packageReferencesFor));
  const visibleOutputs = outputs.filter((output) => {
    const reference = normalizeOutputReference(outputWorkspacePath(output));
    return !reference || !packageReferences.has(reference);
  });
  const hasOutputs = visibleOutputs.length > 0 || packages.length > 0;
  const primaryOutputIndex = preferredOutputIndex(visibleOutputs);
  const normalizedPrimaryIndex = primaryOutputIndex >= 0 ? primaryOutputIndex : visibleOutputs.length > 0 ? 0 : -1;
  const primaryOutput = normalizedPrimaryIndex >= 0 ? visibleOutputs[normalizedPrimaryIndex] : null;
  const secondaryOutputs = visibleOutputs.filter((_, index) => index !== normalizedPrimaryIndex);
  const highlightedOutput = primaryPackage ? null : primaryOutput;
  const supplementalOutputs = primaryPackage ? visibleOutputs : secondaryOutputs;
  const hasAdditionalOutputs = secondaryPackages.length > 0 || supplementalOutputs.length > 0;

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
      {primaryPackage ? (
        <OutputWorkbenchProjectPackage
          project={primaryPackage}
          index={0}
          projectOpenLabel={projectOpenLabel}
          outputContract={outputContract}
          isPrimary
        />
      ) : null}
      {highlightedOutput ? (
        <article className="rounded-lg border border-cortex-primary/50 bg-cortex-primary/15 px-3 py-2 shadow-[0_0_0_1px_rgba(115,92,255,0.12)]" aria-label="Latest output">
          <div className="space-y-3">
            <div className="min-w-0">
              <div className="text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-primary">Latest output</div>
              <div className="mt-1 flex min-w-0 flex-wrap items-center gap-2">
                <span className="truncate text-sm font-semibold text-cortex-text-main">{highlightedOutput.text}</span>
                <OutcomeHealthBadge health="completed" />
              </div>
              <p className="mt-1 text-xs leading-5 text-cortex-text-muted">Open the completed output to review the result.</p>
            </div>
            <div className="flex w-full min-w-0 flex-wrap items-center gap-2" data-testid="latest-output-actions">
              <OutputAccessActions
                label={highlightedOutput.text}
                url={highlightedOutput.url}
                storagePath={highlightedOutput.storagePath}
                openLabel={deliverablePresentation({
                  outputContract,
                  kind: highlightedOutput.kind,
                  type: highlightedOutput.type,
                  contentType: highlightedOutput.contentType,
                  title: highlightedOutput.text,
                  entrypoint: highlightedOutput.entrypoint,
                  path: highlightedOutput.storagePath ?? highlightedOutput.url,
                }).actionLabel}
                folderLabel="Open folder"
                primary
                showFolder={false}
                openInCanvas
                proofArtifactId={highlightedOutput.proofArtifactId}
              />
            </div>
          </div>
          <details className="mt-3 border-t border-cortex-border/70 pt-2">
            <summary className="cursor-pointer text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
              Details and proof
            </summary>
            <div className="mt-3 min-w-0 space-y-3">
              <OutputPathHint storagePath={highlightedOutput.storagePath} url={highlightedOutput.url} />
              <div className="flex w-full min-w-0 flex-wrap items-center gap-1.5">
                <OutputAccessActions
                  label={highlightedOutput.text}
                  url={highlightedOutput.url}
                  storagePath={highlightedOutput.storagePath}
                  folderLabel="Open folder"
                  showOpen={false}
                />
                <button
                  type="button"
                  onClick={() => requestSomaOutputContinuation({
                    title: highlightedOutput.text,
                    reference: outputContinuationReference(highlightedOutput),
                    proof: highlightedOutput.proofArtifactId,
                  })}
                  className="inline-flex h-7 items-center gap-1.5 rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-2.5 text-[11px] font-semibold text-cortex-primary transition-colors hover:border-cortex-primary/60 hover:bg-cortex-primary/15"
                  title={`Reply to ${highlightedOutput.text} in Soma`}
                  aria-label={`Reply to ${highlightedOutput.text} in Soma`}
                >
                  <MessageSquareReply className="h-3 w-3" />
                  Reply
                </button>
                <button
                  type="button"
                  onClick={() => void copyOutputQuote(highlightedOutput, `primary-${highlightedOutput.text}-${highlightedOutput.url ?? "text"}`)}
                  className="inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-lg border border-cortex-border/70 text-cortex-text-muted transition-colors hover:border-cortex-info/40 hover:bg-cortex-info/10 hover:text-cortex-info"
                  title={copiedOutputKey === `primary-${highlightedOutput.text}-${highlightedOutput.url ?? "text"}` ? "Copied output quote" : "Copy output quote"}
                  aria-label={copiedOutputKey === `primary-${highlightedOutput.text}-${highlightedOutput.url ?? "text"}` ? "Copied output quote" : `Copy output quote for ${highlightedOutput.text}`}
                >
                  {copiedOutputKey === `primary-${highlightedOutput.text}-${highlightedOutput.url ?? "text"}` ? <Check className="h-3.5 w-3.5" /> : <Quote className="h-3.5 w-3.5" />}
                </button>
              </div>
              <OutputProofDetails proof={highlightedOutput.proof} proofArtifactId={highlightedOutput.proofArtifactId} />
            </div>
          </details>
        </article>
      ) : null}
      {hasAdditionalOutputs ? (
        <details className="border-t border-cortex-border/70 px-1 pt-2">
          <summary className="cursor-pointer text-[10px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">
            More outputs and verification
          </summary>
          <div className="mt-2 grid min-w-0 gap-2">
            {secondaryPackages.map((project, index) => (
              <OutputWorkbenchProjectPackage
                key={`${itemText(project) ?? "Project package"}-${index + 1}`}
                project={project}
                index={index + 1}
                projectOpenLabel={projectOpenLabel}
              />
            ))}
            {supplementalOutputs.map((output, index) => {
              const key = `${output.text}-${output.url ?? "text"}-${index}`;
              const copied = copiedOutputKey === key;
              return (
                <div key={key} className="min-w-0 rounded-lg border border-cortex-border/60 bg-cortex-bg/70 p-2">
                  <div className="min-w-0">
                    {output.url ? (
                      <a href={output.url} target="_blank" rel="noopener noreferrer" className="inline-flex min-w-0 items-center gap-1 text-cortex-primary hover:underline">
                        <span className="truncate">{output.text}</span>
                        <ExternalLink className="h-3 w-3 shrink-0" />
                      </a>
                    ) : (
                      <span className="min-w-0 truncate text-sm text-cortex-text-main">{output.text}</span>
                    )}
                    <OutputPathHint storagePath={output.storagePath} url={output.url} />
                  </div>
                  <div className="mt-2 flex w-full min-w-0 flex-wrap items-center gap-1">
                    <OutputAccessActions label={output.text} url={output.url} storagePath={output.storagePath} openLabel="Open file" folderLabel="Open folder" openInCanvas proofArtifactId={output.proofArtifactId} />
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
                  </div>
                </div>
              );
            })}
          </div>
        </details>
      ) : null}
      <ExecutionSummaryMediaPreview outputs={visibleOutputs} />
    </div>
  );
}

function packageReferencesFor(project: ExecutionSummaryItem) {
  const folder = normalizeOutputReference(project.folder);
  const openPath = normalizeOutputReference(projectPackageOpenPath({
    folder: project.folder,
    entrypoint: project.entrypoint,
    filePath: project.path,
  }));
  return [folder, openPath].filter((reference): reference is string => Boolean(reference));
}

function normalizeOutputReference(reference?: string | null) {
  return reference?.trim().replace(/\\/g, "/").replace(/^workspace\//, "").replace(/\/+$/, "") || null;
}
