"use client";

import { X } from "lucide-react";
import { useApplyConfigDialog } from "./useApplyConfigDialog";
import { GovernedStatusBanner, PreviewSummary } from "./configGovernanceUi";

/**
 * "Edit and apply": edit the YAML/JSON document, preview (202) the change,
 * review the diff/validation/write-back guard, then approve with the
 * hash-bound confirm token. See docs/API_REFERENCE.md#apply.
 */
export function ApplyConfigDialog({
  kind,
  id,
  initialDocument,
  onClose,
  onApplied,
}: {
  kind: "templates" | "groups";
  id: string;
  initialDocument: string;
  onClose: () => void;
  onApplied?: () => void;
}) {
  const { document, setDocument, phase, preview, approve, reset } = useApplyConfigDialog(kind, id, initialDocument);

  const busy = phase.status === "loading";
  const previewed = phase.status === "previewed" ? phase.preview : null;
  const approveDisabled = busy || !previewed || previewed.blocked || !previewed.confirm_token;

  return (
    <div
      className="absolute inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      data-testid="apply-config-dialog"
      role="dialog"
      aria-label="Edit and apply"
    >
      <div className="flex max-h-[85vh] w-[640px] max-w-full flex-col rounded-2xl border border-cortex-border bg-cortex-surface shadow-2xl">
        <div className="flex flex-shrink-0 items-center justify-between border-b border-cortex-border px-4 py-3">
          <h2 className="text-sm font-semibold text-cortex-text-main">Edit and apply</h2>
          <button
            type="button"
            onClick={onClose}
            data-testid="apply-config-close"
            className="rounded p-1 text-cortex-text-muted hover:bg-cortex-border hover:text-cortex-text-main"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="min-h-0 flex-1 space-y-4 overflow-y-auto p-4">
          <p className="text-xs text-cortex-text-muted">
            Editing <span className="font-mono text-cortex-text-main">{kind}/{id}</span>. Nothing changes until you
            approve; secrets can never be written back.
          </p>

          <label className="block space-y-2">
            <span className="text-sm font-medium text-cortex-text-main">Document (YAML)</span>
            <textarea
              value={document}
              onChange={(event) => setDocument(event.target.value)}
              rows={14}
              spellCheck={false}
              data-testid="apply-config-editor"
              className="w-full resize-y rounded-lg border border-cortex-border bg-cortex-bg p-3 font-mono text-xs leading-5 text-cortex-text-main outline-none focus:border-cortex-primary/40"
            />
          </label>

          {previewed ? (
            <PreviewSummary
              diff={previewed.preview.diff}
              valid={previewed.preview.valid}
              validationErrors={previewed.preview.validation_errors}
              blocked={previewed.blocked}
            />
          ) : null}

          {phase.status === "conflict" ||
          phase.status === "violations" ||
          phase.status === "denied" ||
          phase.status === "unavailable" ||
          phase.status === "failed" ? (
            <GovernedStatusBanner
              status={phase.status}
              detail={"detail" in phase ? phase.detail : undefined}
              violations={"violations" in phase ? phase.violations : undefined}
            />
          ) : null}

          {phase.status === "success" ? (
            <div
              className="rounded-lg border border-cortex-primary/30 bg-cortex-primary/10 p-3 text-sm text-cortex-text-main"
              data-testid="apply-config-success"
            >
              Applied. The stored document now matches what you approved.
            </div>
          ) : null}
        </div>

        <div className="flex flex-shrink-0 items-center justify-end gap-2 border-t border-cortex-border px-4 py-3">
          {phase.status === "success" ? (
            <button
              type="button"
              onClick={() => {
                onApplied?.();
                onClose();
              }}
              data-testid="apply-config-done"
              className="rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-3 py-1.5 text-sm text-cortex-primary hover:bg-cortex-primary/15"
            >
              Done
            </button>
          ) : (
            <>
              <button
                type="button"
                onClick={phase.status === "conflict" ? () => void preview() : reset}
                data-testid="apply-config-reset"
                className="rounded-lg border border-cortex-border px-3 py-1.5 text-sm text-cortex-text-main hover:border-cortex-primary/25"
              >
                {phase.status === "conflict" ? "Preview again" : "Reset"}
              </button>
              <button
                type="button"
                onClick={() => void preview()}
                disabled={busy || !document.trim()}
                data-testid="apply-config-preview"
                className="rounded-lg border border-cortex-border bg-cortex-bg px-3 py-1.5 text-sm text-cortex-text-main hover:border-cortex-primary/25 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {busy ? "Working…" : "Preview changes"}
              </button>
              <button
                type="button"
                onClick={() => void approve()}
                disabled={approveDisabled}
                data-testid="apply-config-approve"
                className="rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-3 py-1.5 text-sm text-cortex-primary hover:bg-cortex-primary/15 disabled:cursor-not-allowed disabled:opacity-50"
              >
                Approve and apply
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
