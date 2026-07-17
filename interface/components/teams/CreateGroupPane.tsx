import { useState, type ComponentType, type ReactNode } from "react";
import {
  FolderKanban,
  Plus,
  ShieldCheck,
  SlidersHorizontal,
  Users,
} from "lucide-react";
import {
  type ApprovalPrompt,
  type GroupDraft,
  type WorkMode,
} from "./groupWorkspaceTypes";

type CreateStep = "basics" | "policy" | "people" | "advanced";

const CREATE_STEPS: Array<{
  id: CreateStep;
  label: string;
  summary: string;
  icon: ComponentType<{ className?: string }>;
}> = [
  {
    id: "basics",
    label: "Basics",
    summary: "Name and goal",
    icon: FolderKanban,
  },
  {
    id: "policy",
    label: "Policy",
    summary: "Mode and approvals",
    icon: ShieldCheck,
  },
  { id: "people", label: "People", summary: "Teams and members", icon: Users },
  {
    id: "advanced",
    label: "Advanced",
    summary: "Workspace and coordinator",
    icon: SlidersHorizontal,
  },
];

export function CreateGroupPane({
  draft,
  approvalPrompt,
  saving,
  onDraftChange,
  onCreateGroup,
}: {
  draft: GroupDraft;
  approvalPrompt: ApprovalPrompt | null;
  saving: boolean;
  onDraftChange: (patch: Partial<GroupDraft>) => void;
  onCreateGroup: () => void;
}) {
  const [activeStep, setActiveStep] = useState<CreateStep>("basics");
  const compactInputClassName =
    "w-full rounded-lg border border-cortex-border bg-cortex-bg px-3 py-1.5 text-sm text-cortex-text-main outline-none placeholder:text-cortex-text-muted";

  return (
    <section className="min-w-0 overflow-x-hidden rounded-2xl border border-cortex-border bg-cortex-surface p-3">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex min-w-0 items-center gap-2">
          <Plus className="h-4 w-4 shrink-0 text-cortex-primary" />
          <h2 className="min-w-0 break-words text-xs font-semibold uppercase tracking-[0.12em] text-cortex-text-main sm:text-sm sm:tracking-[0.16em]">
            Define group action lane
          </h2>
        </div>
        <button
          type="button"
          onClick={onCreateGroup}
          disabled={saving}
          data-testid="groups-create-button"
          className="w-full rounded-xl bg-cortex-primary px-4 py-2 text-sm font-semibold text-cortex-bg disabled:opacity-60 sm:w-auto"
        >
          {saving
            ? "Saving..."
            : approvalPrompt
              ? "Confirm and create group"
              : "Create group"}
        </button>
      </div>
      <div
        className="mt-4 flex flex-wrap gap-2"
        role="tablist"
        aria-label="Create group sections"
      >
        {CREATE_STEPS.map((step) => {
          const Icon = step.icon;
          const selected = activeStep === step.id;
          return (
            <button
              key={step.id}
              type="button"
              role="tab"
              aria-label={step.label}
              aria-selected={selected}
              onClick={() => setActiveStep(step.id)}
              className={`min-w-[10rem] flex-1 rounded-xl border px-3 py-2 text-left transition-colors sm:flex-none ${
                selected
                  ? "border-cortex-primary/45 bg-cortex-primary/10 text-cortex-text-main"
                  : "border-cortex-border bg-cortex-bg/60 text-cortex-text-muted hover:text-cortex-text-main"
              }`}
            >
              <span className="flex items-center gap-2 text-sm font-bold">
                <Icon className="h-4 w-4 text-cortex-primary" />
                {step.label}
              </span>
              <span className="mt-1 block text-[11px] leading-4 text-cortex-text-muted">
                {step.summary}
              </span>
            </button>
          );
        })}
      </div>
      <div className="mt-3 min-w-0">
        {activeStep === "basics" ? (
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
        ) : null}
        {activeStep === "policy" ? (
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
        ) : null}
        {activeStep === "people" ? (
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
          </FormSection>
        ) : null}
        {activeStep === "advanced" ? (
          <FormSection title="Workspace and coordination">
            <Field label="Workspace Folder">
              <input
                aria-label="Workspace Folder"
                placeholder="auto: groups/team-id"
                value={draft.workspaceFolder}
                onChange={(event) =>
                  onDraftChange({ workspaceFolder: event.target.value })
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
          </FormSection>
        ) : null}
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
    </section>
  );
}

function Field({
  label,
  children,
}: {
  label: string;
  children: ReactNode;
}) {
  return (
    <label className="flex min-w-0 flex-col gap-1 text-xs">
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
  children: ReactNode;
}) {
  return (
    <div className="grid min-w-0 content-start gap-2 rounded-xl border border-cortex-border bg-cortex-bg p-2.5">
      <h3 className="font-mono text-[11px] uppercase tracking-[0.16em] text-cortex-primary">
        {title}
      </h3>
      {children}
    </div>
  );
}
