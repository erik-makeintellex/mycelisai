"use client";

import { useCallback, useEffect, useState } from "react";
import { Copy, Download, X } from "lucide-react";
import LoadFailureState from "@/components/shared/LoadFailureState";
import { getData, type DeploymentConfigRef } from "./deploymentTypes";
import { CloneTemplateDialog } from "./CloneTemplateDialog";
import { ApplyConfigDialog } from "./ApplyConfigDialog";

type ConfigFormat = "yaml" | "json";

type ConfigDocument = {
  kind: string;
  id: string;
  format: ConfigFormat;
  document: string;
  source: "file" | "database" | "runtime" | "operator";
  source_path?: string;
  redactions: string[];
  generated_at: string;
};

type GovernanceDialog = "clone" | "apply" | null;

type ViewState =
  | { status: "loading" }
  | { status: "restricted" }
  | { status: "not_found" }
  | { status: "error"; detail: string }
  | { status: "ready"; doc: ConfigDocument };

/** Label for the "kind" line. Planned departments resolve to an organization config. */
function kindLabel(configRef: DeploymentConfigRef): string {
  if (configRef.kind === "organization") return "Organization config";
  if (configRef.kind === "team") return "Team config";
  if (configRef.kind === "group") return "Group config";
  return "Config";
}

function sourceLabel(doc: ConfigDocument): string {
  const base = doc.source === "file" ? "File" : doc.source === "database" ? "Database" : "Runtime";
  return doc.source_path ? `${base}: ${doc.source_path}` : base;
}

async function fetchConfig(path: string, format: ConfigFormat): Promise<ViewState> {
  try {
    const res = await fetch(`${path}?format=${format}`, { cache: "no-store" });
    if (res.status === 403) return { status: "restricted" };
    if (res.status === 404) return { status: "not_found" };
    if (!res.ok) return { status: "error", detail: `Config request failed (HTTP ${res.status}).` };
    const doc = await getData<ConfigDocument>(res);
    return { status: "ready", doc };
  } catch (failure) {
    return {
      status: "error",
      detail: failure instanceof Error ? failure.message : "Could not load config.",
    };
  }
}

export function ConfigViewer({
  configRef,
  onClose,
}: {
  configRef: DeploymentConfigRef;
  onClose: () => void;
}) {
  const [format, setFormat] = useState<ConfigFormat>("yaml");
  const [state, setState] = useState<ViewState>({ status: "loading" });
  const [copied, setCopied] = useState(false);
  const [dialog, setDialog] = useState<GovernanceDialog>(null);

  const load = useCallback(async (nextFormat: ConfigFormat) => {
    setState({ status: "loading" });
    setState(await fetchConfig(configRef.path, nextFormat));
  }, [configRef.path]);

  useEffect(() => {
    void load(format);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [configRef.path, format]);

  const handleCopy = useCallback(() => {
    if (state.status !== "ready") return;
    void navigator.clipboard?.writeText(state.doc.document);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }, [state]);

  const handleDownload = useCallback(() => {
    if (state.status !== "ready") return;
    const extension = state.doc.format === "json" ? "json" : "yaml";
    const blob = new Blob([state.doc.document], { type: "text/plain;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `${state.doc.kind}-${state.doc.id}.${extension}`;
    link.click();
    URL.revokeObjectURL(url);
  }, [state]);

  return (
    <div
      className="absolute right-0 top-0 bottom-0 z-40 flex w-[560px] max-w-full flex-col border-l border-cortex-border bg-cortex-surface shadow-2xl"
      data-testid="config-viewer"
      role="dialog"
      aria-label={`${kindLabel(configRef)} viewer`}
    >
      <div className="flex flex-shrink-0 items-center justify-between border-b border-cortex-border px-4 py-3">
        <div className="min-w-0">
          <h2 className="truncate text-sm font-semibold text-cortex-text-main">{kindLabel(configRef)}</h2>
          <p className="truncate text-xs text-cortex-text-muted">{configRef.id}</p>
        </div>
        <button
          type="button"
          onClick={onClose}
          data-testid="config-viewer-close"
          className="rounded p-1 text-cortex-text-muted hover:bg-cortex-border hover:text-cortex-text-main"
        >
          <X className="h-4 w-4" />
        </button>
      </div>

      <div className="flex flex-shrink-0 items-center justify-between gap-2 border-b border-cortex-border px-4 py-2">
        <div className="flex overflow-hidden rounded-lg border border-cortex-border" role="group" aria-label="Format">
          {(["yaml", "json"] as const).map((opt) => (
            <button
              key={opt}
              type="button"
              onClick={() => setFormat(opt)}
              data-testid={`config-viewer-format-${opt}`}
              aria-pressed={format === opt}
              className={`px-3 py-1.5 text-xs font-mono uppercase tracking-wide ${
                format === opt
                  ? "bg-cortex-primary/15 text-cortex-primary"
                  : "text-cortex-text-muted hover:bg-cortex-bg"
              }`}
            >
              {opt}
            </button>
          ))}
        </div>
        {state.status === "ready" ? (
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={handleCopy}
              data-testid="config-viewer-copy"
              className="inline-flex items-center gap-1 rounded-lg border border-cortex-border px-2.5 py-1.5 text-xs text-cortex-text-main hover:border-cortex-primary/25"
            >
              <Copy className="h-3.5 w-3.5" />
              {copied ? "Copied" : "Copy"}
            </button>
            <button
              type="button"
              onClick={handleDownload}
              data-testid="config-viewer-download"
              className="inline-flex items-center gap-1 rounded-lg border border-cortex-border px-2.5 py-1.5 text-xs text-cortex-text-main hover:border-cortex-primary/25"
            >
              <Download className="h-3.5 w-3.5" />
              Download
            </button>
          </div>
        ) : null}
      </div>

      {state.status === "ready" ? (
        <GovernanceActions configRef={configRef} doc={state.doc} onOpen={setDialog} />
      ) : null}

      <div className="min-h-0 flex-1 overflow-y-auto p-4">
        {state.status === "loading" ? (
          <p className="text-sm text-cortex-text-muted" data-testid="config-viewer-loading">
            Loading config…
          </p>
        ) : null}

        {state.status === "restricted" ? (
          <div
            className="rounded-xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted"
            data-testid="config-viewer-restricted"
          >
            This config requires admin access.
          </div>
        ) : null}

        {state.status === "not_found" ? (
          <div
            className="rounded-xl border border-cortex-border bg-cortex-bg px-4 py-4 text-sm text-cortex-text-muted"
            data-testid="config-viewer-not-found"
          >
            Config not found.
          </div>
        ) : null}

        {state.status === "error" ? (
          <LoadFailureState
            title="Could not load config"
            detail={state.detail}
            onRetry={() => void load(format)}
            testId="config-viewer-load-failure"
          />
        ) : null}

        {state.status === "ready" ? (
          <div className="space-y-3">
            <p className="text-xs text-cortex-text-muted" data-testid="config-viewer-source">
              {sourceLabel(state.doc)}
            </p>
            <pre
              className="max-h-[55vh] overflow-auto rounded-lg border border-cortex-border bg-cortex-bg p-3 font-mono text-xs leading-5 text-cortex-text-main"
              data-testid="config-viewer-document"
            >
              {state.doc.document}
            </pre>
            <div className="rounded-lg border border-cortex-border bg-cortex-bg p-3" data-testid="config-viewer-redactions">
              <h3 className="text-xs font-semibold uppercase tracking-wide text-cortex-text-muted">
                Hidden values
              </h3>
              <p className="mt-1 text-xs text-cortex-text-muted">
                Secrets are never shown here; env-var references stay visible.
              </p>
              {state.doc.redactions.length === 0 ? (
                <p className="mt-2 text-xs text-cortex-text-muted">Nothing was hidden in this document.</p>
              ) : (
                <ul className="mt-2 space-y-1 font-mono text-xs text-cortex-text-main">
                  {state.doc.redactions.map((path) => (
                    <li key={path} data-testid="config-viewer-redaction-item">
                      {path}
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </div>
        ) : null}
      </div>

      {dialog === "clone" && state.status === "ready" ? (
        <CloneTemplateDialog
          sourceTemplateId={state.doc.id}
          sourceTemplateName={state.doc.id}
          onClose={() => setDialog(null)}
          onCloned={() => void load(format)}
        />
      ) : null}
      {dialog === "apply" && state.status === "ready" ? (
        <ApplyConfigDialog
          kind={configRef.kind === "template" ? "templates" : "groups"}
          id={state.doc.id}
          initialDocument={state.doc.document}
          onClose={() => setDialog(null)}
          onApplied={() => void load(format)}
        />
      ) : null}
    </div>
  );
}

/** "Clone as custom template" for a file/starter template; "Edit and apply" for an operator template or group. */
function GovernanceActions({
  configRef,
  doc,
  onOpen,
}: {
  configRef: DeploymentConfigRef;
  doc: ConfigDocument;
  onOpen: (dialog: GovernanceDialog) => void;
}) {
  const canClone = configRef.kind === "template" && doc.source === "file";
  const canApply =
    (configRef.kind === "template" && doc.source === "operator") || configRef.kind === "group";

  if (!canClone && !canApply) return null;

  return (
    <div className="flex flex-shrink-0 items-center gap-2 border-b border-cortex-border px-4 py-2">
      {canClone ? (
        <button
          type="button"
          onClick={() => onOpen("clone")}
          data-testid="config-viewer-clone-template"
          className="rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-2.5 py-1.5 text-xs text-cortex-primary hover:bg-cortex-primary/15"
        >
          Clone as custom template
        </button>
      ) : null}
      {canApply ? (
        <button
          type="button"
          onClick={() => onOpen("apply")}
          data-testid="config-viewer-edit-apply"
          className="rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-2.5 py-1.5 text-xs text-cortex-primary hover:bg-cortex-primary/15"
        >
          Edit and apply
        </button>
      ) : null}
    </div>
  );
}
