import type { ConfigDiff } from "./configGovernance";

/** Common status banner for the terminal, non-preview phases both dialogs share. */
export function GovernedStatusBanner({
  status,
  detail,
  violations,
}: {
  status: "conflict" | "violations" | "denied" | "unavailable" | "failed";
  detail?: string;
  violations?: string[];
}) {
  if (status === "conflict") {
    return (
      <Banner testId="config-governance-conflict" title="The document changed since preview">
        Preview again to see the current version before approving.
      </Banner>
    );
  }
  if (status === "violations") {
    return (
      <Banner testId="config-governance-violations" title="This document cannot be written back">
        <p>Secrets must stay as env or ref references, never redaction markers or literal values.</p>
        {violations && violations.length > 0 ? (
          <ul className="mt-2 space-y-1 font-mono text-xs">
            {violations.map((path) => (
              <li key={path} data-testid="config-governance-violation-item">
                {path}
              </li>
            ))}
          </ul>
        ) : null}
      </Banner>
    );
  }
  if (status === "denied") {
    return (
      <Banner testId="config-governance-denied" title="You do not have permission for this change">
        This action requires the config:write role (and groups:write for groups).
      </Banner>
    );
  }
  if (status === "unavailable") {
    return (
      <Banner testId="config-governance-unavailable" title="Config write is unavailable">
        {detail || "Config write is unavailable right now."}
      </Banner>
    );
  }
  return (
    <Banner testId="config-governance-failed" title="This request could not be completed">
      {detail || "Something went wrong."}
    </Banner>
  );
}

function Banner({
  testId,
  title,
  children,
}: {
  testId: string;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div
      className="rounded-lg border border-cortex-danger/40 bg-cortex-danger/10 p-3 text-sm text-cortex-text-main"
      data-testid={testId}
      role="alert"
    >
      <p className="font-semibold">{title}</p>
      <div className="mt-1 text-cortex-text-muted">{children}</div>
    </div>
  );
}

/** Diff and validation summary shown once a preview is available. */
export function PreviewSummary({
  diff,
  valid,
  validationErrors,
  blocked,
}: {
  diff?: ConfigDiff;
  valid: boolean;
  validationErrors: string[];
  blocked: boolean;
}) {
  return (
    <div className="space-y-3" data-testid="config-governance-preview-summary">
      {diff ? (
        <div className="rounded-lg border border-cortex-border bg-cortex-bg p-3">
          <h4 className="text-xs font-semibold uppercase tracking-wide text-cortex-text-muted">Changes</h4>
          <DiffList label="Added" paths={diff.added} testId="config-governance-diff-added" />
          <DiffList label="Changed" paths={diff.changed} testId="config-governance-diff-changed" />
          <DiffList label="Removed" paths={diff.removed} testId="config-governance-diff-removed" />
          {diff.added.length === 0 && diff.changed.length === 0 && diff.removed.length === 0 ? (
            <p className="mt-2 text-xs text-cortex-text-muted">No path-level changes detected.</p>
          ) : null}
        </div>
      ) : null}
      <div
        className={`rounded-lg border p-3 text-xs ${
          valid ? "border-cortex-border bg-cortex-bg text-cortex-text-muted" : "border-cortex-danger/40 bg-cortex-danger/10 text-cortex-text-main"
        }`}
        data-testid="config-governance-validation"
      >
        {valid ? (
          <p>Document is valid.</p>
        ) : (
          <>
            <p className="font-semibold">Validation errors</p>
            <ul className="mt-1 space-y-1 font-mono">
              {validationErrors.map((message, index) => (
                <li key={`${index}-${message}`} data-testid="config-governance-validation-error">
                  {message}
                </li>
              ))}
            </ul>
          </>
        )}
      </div>
      {blocked ? (
        <p className="text-xs text-cortex-danger" data-testid="config-governance-blocked">
          Blocked: this document cannot be approved as previewed.
        </p>
      ) : null}
    </div>
  );
}

function DiffList({ label, paths, testId }: { label: string; paths: string[]; testId: string }) {
  if (paths.length === 0) return null;
  return (
    <div className="mt-2">
      <p className="text-[11px] font-mono uppercase tracking-[0.12em] text-cortex-text-muted">{label}</p>
      <ul className="mt-1 space-y-0.5 font-mono text-xs text-cortex-text-main" data-testid={testId}>
        {paths.map((path) => (
          <li key={path}>{path}</li>
        ))}
      </ul>
    </div>
  );
}
