import { Sparkles } from "lucide-react";
import type { OrganizationDeliveryDefaults } from "@/lib/organizations";

// OrganizationStarterIntro (RC-004C-2) shows Soma's opening line and a
// one-line delivery posture summary for organizations created from a
// company starter. It renders nothing when neither field is present, so
// organizations created empty or before delivery_defaults/soma_intro
// existed look unchanged.

const APPROVAL_POSTURE_LABEL: Record<string, string> = {
    approve_external_sends: "nothing leaves the organization without your approval",
    approve_changes: "deploys and writes need your approval",
    approve_high_risk: "high-risk actions need your approval",
    approve_all_mutations: "every mutation needs your approval",
};

function approvalPostureLabel(postureId: string): string {
    return APPROVAL_POSTURE_LABEL[postureId] ?? "approvals follow the organization's delivery policy";
}

function outputsLabel(defaults: OrganizationDeliveryDefaults): string {
    const { retain, proof_required } = defaults.outputs;
    if (retain && proof_required) {
        return "outputs are kept with proof";
    }
    if (retain) {
        return "outputs are kept";
    }
    if (proof_required) {
        return "outputs require proof";
    }
    return "outputs are not retained";
}

function postureSummary(defaults: OrganizationDeliveryDefaults): string {
    const teamSize = `teams top out at ${defaults.max_team_members} member${defaults.max_team_members === 1 ? "" : "s"}`;
    return `${approvalPostureLabel(defaults.approval_posture)}, ${outputsLabel(defaults)}, and ${teamSize}.`;
}

export function OrganizationStarterIntro({
    somaIntro,
    deliveryDefaults,
}: {
    somaIntro?: string;
    deliveryDefaults?: OrganizationDeliveryDefaults;
}) {
    if (!somaIntro && !deliveryDefaults) {
        return null;
    }

    return (
        <section className="rounded-2xl border border-cortex-border bg-cortex-surface px-5 py-4">
            <div className="flex items-start gap-3">
                <Sparkles className="mt-0.5 h-4 w-4 shrink-0 text-cortex-primary" />
                <div className="space-y-2 text-sm leading-6 text-cortex-text-muted">
                    {somaIntro ? <p className="text-cortex-text-main">&ldquo;{somaIntro}&rdquo;</p> : null}
                    {deliveryDefaults ? <p>{postureSummary(deliveryDefaults)}</p> : null}
                    <p>
                        Departments are planned — Soma will propose launching teams here for you to approve.
                    </p>
                </div>
            </div>
        </section>
    );
}
