import { Plus } from "lucide-react";
import {
  type ApprovalPrompt,
  type GroupDraft,
  type WorkMode,
} from "./groupWorkspaceTypes";

export function CreateGroupPane({
  draft,
  notice,
  error,
  approvalPrompt,
  saving,
  onDraftChange,
  onCreateGroup,
}: {
  draft: GroupDraft;
  notice: string | null;
  error: string | null;
  approvalPrompt: ApprovalPrompt | null;
  saving: boolean;
  onDraftChange: (patch: Partial<GroupDraft>) => void;
  onCreateGroup: () => void;
}) {
  const compactInputClassName =
    "w-full rounded-lg border border-cortex-border bg-cortex-bg px-3 py-1.5 text-sm text-cortex-text-main outline-none placeholder:text-cortex-text-muted";

  return (
    <section className="rounded-2xl border border-cortex-border bg-cortex-surface p-3">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <Plus className="h-4 w-4 text-cortex-primary" />
          <h2 className="text-sm font-semibold uppercase tracking-[0.16em] text-cortex-text-main">
            Define group action lane
          </h2>
        </div>
        <button
          type="button"
          onClick={onCreateGroup}
          disabled={saving}
          data-testid="groups-create-button"
          className="rounded-xl bg-cortex-primary px-4 py-2 text-sm font-semibold text-cortex-bg disabled:opacity-60"
        >
          {saving
            ? "Saving..."
            : approvalPrompt
              ? "Confirm and create group"
              : "Create group"}
        </button>
      </div>
      <div className="mt-3 grid gap-2 lg:grid-cols-[minmax(0,1.05fr)_minmax(0,0.95fr)_minmax(0,1fr)]">
        <FormSection title="Identity">
          <Field label="Name">
            <input
              aria-label="Name"
              value={draft.name}
              onChange={(event) => onDraftChange({ name: event.target.value })}
              className={compactInputClassName}
            />
          </Field>
          <Field label="Goal Statement">
            <textarea
              aria-label="Goal Statement"
              rows={2}
              value={draft.goalStatement}
              onChange={(event) =>
                onDraftChange({ goalStatement: event.target.value })
              }
              className={`${compactInputClassName} resize-y`}
            />
          </Field>
        </FormSection>
        <FormSection title="Action policy">
          <Field label="Work Mode">
            <select
              aria-label="Work Mode"
              value={draft.workMode}
              onChange={(event) =>
                onDraftChange({ workMode: event.target.value as WorkMode })
              }
              className={compactInputClassName}
            >
              <option value="read_only">read_only</option>
              <option value="propose_only">propose_only</option>
              <option value="execute_with_approval">
                execute_with_approval
              </option>
              <option value="execute_bounded">execute_bounded</option>
            </select>
          </Field>
          <Field label="Approval Policy Ref">
            <input
              aria-label="Approval Policy Ref"
              value={draft.approvalPolicyRef}
              onChange={(event) =>
                onDraftChange({ approvalPolicyRef: event.target.value })
              }
              className={compactInputClassName}
            />
          </Field>
          <Field label="Allowed Capabilities">
            <input
              aria-label="Allowed Capabilities"
              value={draft.allowedCapabilities}
              onChange={(event) =>
                onDraftChange({ allowedCapabilities: event.target.value })
              }
              className={compactInputClassName}
            />
          </Field>
        </FormSection>
        <FormSection title="People and duration">
          <Field label="Team IDs">
            <input
              aria-label="Team IDs"
              value={draft.teamIDs}
              onChange={(event) =>
                onDraftChange({ teamIDs: event.target.value })
              }
              className={compactInputClassName}
            />
          </Field>
          <Field label="Coordinator Profile">
            <input
              aria-label="Coordinator Profile"
              value={draft.coordinatorProfile}
              onChange={(event) =>
                onDraftChange({ coordinatorProfile: event.target.value })
              }
              className={compactInputClassName}
            />
          </Field>
          <div className="grid gap-3 sm:grid-cols-2">
            <Field label="Member IDs">
              <input
                aria-label="Member IDs"
                value={draft.memberIDs}
                onChange={(event) =>
                  onDraftChange({ memberIDs: event.target.value })
                }
                className={compactInputClassName}
              />
            </Field>
            <Field label="Expiry">
              <input
                aria-label="Expiry"
                type="datetime-local"
                value={draft.expiry}
                onChange={(event) =>
                  onDraftChange({ expiry: event.target.value })
                }
                className={compactInputClassName}
              />
            </Field>
          </div>
        </FormSection>
      </div>
      {approvalPrompt ? (
        <div
          className="mt-4 rounded-xl border border-cortex-primary/25 bg-cortex-primary/10 p-4"
          data-testid="groups-approval-card"
        >
          <p className="text-sm font-semibold text-cortex-text-main">
            Approval required before creation
          </p>
          <input
            readOnly
            data-testid="groups-confirm-token-input"
            value={approvalPrompt.confirm_token?.token ?? ""}
            className={`${compactInputClassName} mt-3 font-mono`}
          />
        </div>
      ) : null}
      {notice ? (
        <p
          className="mt-4 text-sm text-cortex-primary"
          data-testid="groups-notice"
        >
          {notice}
        </p>
      ) : null}
      {error ? (
        <p
          className="mt-4 text-sm text-cortex-danger"
          data-testid="groups-error"
        >
          {error}
        </p>
      ) : null}
    </section>
  );
}

function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <label className="flex flex-col gap-1 text-xs">
      <span className="font-semibold text-cortex-text-main">{label}</span>
      {children}
    </label>
  );
}

function FormSection({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div className="grid content-start gap-2 rounded-xl border border-cortex-border bg-cortex-bg p-2.5">
      <h3 className="font-mono text-[11px] uppercase tracking-[0.16em] text-cortex-primary">
        {title}
      </h3>
      {children}
    </div>
  );
}
