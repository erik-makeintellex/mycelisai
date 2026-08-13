"use client";

import { ArrowLeft, FileWarning } from "lucide-react";
import OutputAccessActions from "./OutputAccessActions";
import { OutputProofDetails } from "./OutputWorkbenchProofDetails";
import { normalizeWorkspacePath, outputCanvasSourceHref, somaReturnHref } from "@/lib/outputPackageModel";

export interface OutputCanvasProps {
  label?: string | null;
  source?: string | null;
  storagePath?: string | null;
  returnTo?: string | null;
  proofArtifactId?: string | null;
}

export function OutputCanvas({ label, source, storagePath, returnTo, proofArtifactId }: OutputCanvasProps) {
  const title = label?.trim() || "Retained output";
  const path = normalizeWorkspacePath(storagePath);
  const outputHref = outputCanvasSourceHref(source, path);
  const somaHref = somaReturnHref(returnTo);

  return (
    <section className="flex h-full min-h-0 flex-col bg-cortex-bg" aria-label={`${title} output canvas`}>
      <header className="grid shrink-0 grid-cols-[auto_minmax(0,1fr)] items-center gap-3 border-b border-cortex-border bg-cortex-surface px-3 py-3 sm:grid-cols-[auto_minmax(0,1fr)_auto] sm:px-5">
        <a
          href={somaHref}
          className="inline-flex h-9 shrink-0 items-center gap-2 rounded-lg border border-cortex-primary/45 bg-cortex-primary/10 px-3 text-sm font-semibold text-cortex-primary transition-colors hover:bg-cortex-primary/15"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to Soma
        </a>
        <div className="min-w-0 flex-1">
          <h1 className="truncate text-base font-semibold text-cortex-text-main sm:text-lg">{title}</h1>
          <p className="text-xs text-cortex-text-muted">Retained output</p>
        </div>
        {outputHref ? (
          <div className="hidden sm:block">
            <OutputAccessActions
              label={title}
              url={outputHref}
              storagePath={path}
              openLabel="Open separately"
              showFolder={false}
            />
          </div>
        ) : null}
      </header>

      <div className="min-h-0 flex-1 p-2 sm:p-3">
        {outputHref ? (
          <iframe
            src={outputHref}
            title={title}
            className="h-full min-h-[20rem] w-full rounded-lg border border-cortex-border bg-white"
            sandbox="allow-downloads allow-forms allow-modals allow-pointer-lock allow-popups allow-scripts"
            allow="autoplay; fullscreen"
          />
        ) : (
          <div className="flex h-full min-h-[20rem] items-center justify-center rounded-lg border border-cortex-border bg-cortex-surface p-6 text-center">
            <div className="max-w-md">
              <FileWarning className="mx-auto h-8 w-8 text-cortex-warning" />
              <h2 className="mt-3 text-base font-semibold text-cortex-text-main">This output cannot be displayed</h2>
              <p className="mt-1 text-sm leading-6 text-cortex-text-muted">Return to Soma and ask it to retain the file before opening it here.</p>
            </div>
          </div>
        )}
      </div>

      {path || proofArtifactId ? (
        <details className="shrink-0 border-t border-cortex-border bg-cortex-surface px-3 py-2 sm:px-5">
          <summary className="cursor-pointer text-xs font-semibold text-cortex-text-muted">File details</summary>
          <div className="mt-2 flex min-w-0 flex-wrap items-center gap-2">
            <code className="min-w-0 flex-1 truncate text-xs text-cortex-text-muted">{path}</code>
            <OutputAccessActions label={title} url={null} storagePath={path} />
            <OutputProofDetails proofArtifactId={proofArtifactId ?? undefined} />
          </div>
        </details>
      ) : null}
    </section>
  );
}
