"use client";

import { ExternalLink, MessageSquareReply } from "lucide-react";
import { OutcomeHealthBadge } from "@/components/shared/OutcomeHealthBadge";
import { resourcesWorkspaceHref } from "@/lib/outputPackageModel";
import type { OutputWorkbenchDigest } from "./OutputWorkbenchDigest";
import OutputAccessActions, { workspacePathFromOutputUrl } from "./OutputAccessActions";
import { requestSomaOutputContinuation } from "./outputContinuation";

export function SomaOutcomeVaultDeliverableCard({ output }: { output: OutputWorkbenchDigest }) {
  const target = deliverableTarget(output);
  const path = deliverablePath(output);
  const continuationReference = deliverableContinuationReference(output);

  return (
    <div className="rounded-xl border border-cortex-border bg-cortex-bg px-4 py-3" aria-label="Recent deliverable">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            {target ? (
              <a
                href={target.href}
                data-target-reference={target.reference}
                target={target.external ? "_blank" : undefined}
                rel={target.external ? "noopener noreferrer" : undefined}
                className="inline-flex max-w-full items-center gap-1.5 font-semibold text-cortex-primary hover:underline focus:outline-none focus:ring-2 focus:ring-cortex-primary/40"
                aria-label={`Open latest deliverable ${output.text}`}
              >
                <span className="truncate">{output.text}</span>
                <ExternalLink className="h-3.5 w-3.5 shrink-0" />
              </a>
            ) : (
              <div className="font-semibold text-cortex-text-main">{output.text}</div>
            )}
            <OutcomeHealthBadge health={output.health ?? "completed"} />
          </div>
          {path ? (
            <details className="mt-1 text-xs text-cortex-text-muted">
              <summary className="inline-flex cursor-pointer list-none font-semibold text-cortex-primary hover:underline">
                File details
              </summary>
              <code className="mt-1 block max-w-64 truncate font-mono">{path}</code>
            </details>
          ) : (
            <div className="mt-1 text-sm text-cortex-text-muted">Saved output ready to revisit.</div>
          )}
        </div>
        <OutputAccessActions
          label={output.text}
          url={output.url}
          storagePath={output.storagePath}
          openLabel="Open"
          folderLabel="Show folder"
        />
        {continuationReference ? (
          <button
            type="button"
            onClick={() => requestSomaOutputContinuation({
              title: output.text,
              reference: continuationReference,
              proof: output.proofArtifactId,
            })}
            className="inline-flex h-8 items-center gap-1.5 rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-2.5 text-xs font-semibold text-cortex-primary transition-colors hover:border-cortex-primary/60 hover:bg-cortex-primary/15 focus:outline-none focus:ring-2 focus:ring-cortex-primary/40"
            title={`Reply to ${output.text} in Soma`}
            aria-label={`Reply to ${output.text} in Soma`}
          >
            <MessageSquareReply className="h-3.5 w-3.5" />
            Reply
          </button>
        ) : null}
      </div>
      {output.count > 1 ? (
        <div className="mt-2 text-xs text-cortex-text-muted">
          {output.count} saved item{output.count === 1 ? "" : "s"} in this outcome.
        </div>
      ) : null}
    </div>
  );
}

function deliverablePath(output: OutputWorkbenchDigest) {
  return output.storagePath?.trim() || workspacePathFromOutputUrl(output.url);
}

function deliverableContinuationReference(output: OutputWorkbenchDigest) {
  return output.replyReference?.trim() || deliverablePath(output) || output.url?.trim() || null;
}

function deliverableTarget(output: OutputWorkbenchDigest) {
  const url = output.url?.trim();
  const path = deliverablePath(output);
  if (url) {
    return {
      href: url,
      reference: path || url,
      external: /^(https?:)?\/\//i.test(url),
    };
  }
  const resourcesHref = resourcesWorkspaceHref(path);
  return resourcesHref ? {
    href: resourcesHref,
    reference: path ?? resourcesHref,
    external: false,
  } : null;
}
