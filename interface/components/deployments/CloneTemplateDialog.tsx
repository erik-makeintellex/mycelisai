"use client";

import { X } from "lucide-react";
import { useCloneTemplateDialog } from "./useCloneTemplateDialog";
import { GovernedStatusBanner, PreviewSummary } from "./configGovernanceUi";

/**
 * "Clone as custom template": name it, preview (202) the clone, review the
 * diff/validation, then approve with the hash-bound confirm token.
 * See docs/API_REFERENCE.md#clone.
 */
export function CloneTemplateDialog({
  sourceTemplateId,
  sourceTemplateName,
  onClose,
  onCloned,
}: {
  sourceTemplateId: string;
  sourceTemplateName: string;
  onClose: () => void;
  onCloned?: (newTemplateId: string) => void;
}) {
  const { name, setName, phase, preview, approve, reset } = useCloneTemplateDialog(sourceTemplateId);

  const busy = phase.status === "loading";
  const previewed = phase.status === "previewed" ? phase.preview : null;
  const approveDisabled = busy || !previewed || previewed.blocked || !previewed.confirm_token;

  return (
    <div
      className="absolute inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      data-testid="clone-template-dialog"
      role="dialog"
      aria-label="Clone as custom template"
    >
      <div className="flex max-h-[85vh] w-[520px] max-w-full flex-col rounded-2xl border border-cortex-border bg-cortex-surface shadow-2xl">
        <div className="flex flex-shrink-0 items-center justify-between border-b border-cortex-border px-4 py-3">
          <h2 className="text-sm font-semibold text-cortex-text-main">Clone as custom template</h2>
          <button
            type="button"
            onClick={onClose}
            data-testid="clone-template-close"
            className="rounded p-1 text-cortex-text-muted hover:bg-cortex-border hover:text-cortex-text-main"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="min-h-0 flex-1 space-y-4 overflow-y-auto p-4">
          <p className="text-xs text-cortex-text-muted">
            Cloning <span className="font-mono text-cortex-text-main">{sourceTemplateId}</span> ({sourceTemplateName})
            into an operator template. Nothing changes until you approve.
          </p>

          <label className="block space-y-2">
            <span className="text-sm font-medium text-cortex-text-main">New template name</span>
            <input
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder="Acme Studio"
              data-testid="clone-template-name"
              className="w-full rounded-lg border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-main outline-none placeholder:text-cortex-text-muted focus:border-cortex-primary/40"
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
              data-testid="clone-template-success"
            >
              Cloned as <span className="font-mono">{phase.result.template.id}</span>. It now appears as a Custom
              starter.
            </div>
          ) : null}
        </div>

        <div className="flex flex-shrink-0 items-center justify-end gap-2 border-t border-cortex-border px-4 py-3">
          {phase.status === "success" ? (
            <button
              type="button"
              onClick={() => {
                onCloned?.(phase.result.template.id);
                onClose();
              }}
              data-testid="clone-template-done"
              className="rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-3 py-1.5 text-sm text-cortex-primary hover:bg-cortex-primary/15"
            >
              Done
            </button>
          ) : (
            <>
              <button
                type="button"
                onClick={phase.status === "conflict" ? () => void preview() : reset}
                data-testid="clone-template-reset"
                className="rounded-lg border border-cortex-border px-3 py-1.5 text-sm text-cortex-text-main hover:border-cortex-primary/25"
              >
                {phase.status === "conflict" ? "Preview again" : "Reset"}
              </button>
              <button
                type="button"
                onClick={() => void preview()}
                disabled={busy || !name.trim()}
                data-testid="clone-template-preview"
                className="rounded-lg border border-cortex-border bg-cortex-bg px-3 py-1.5 text-sm text-cortex-text-main hover:border-cortex-primary/25 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {busy ? "Working…" : "Preview"}
              </button>
              <button
                type="button"
                onClick={() => void approve()}
                disabled={approveDisabled}
                data-testid="clone-template-approve"
                className="rounded-lg border border-cortex-primary/35 bg-cortex-primary/10 px-3 py-1.5 text-sm text-cortex-primary hover:bg-cortex-primary/15 disabled:cursor-not-allowed disabled:opacity-50"
              >
                Approve and clone
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
