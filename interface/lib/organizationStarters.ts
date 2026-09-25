import type { OrganizationTemplateSummary } from "@/lib/organizations";

// Types for GET /api/v1/templates?view=organization-starters (RC-004A/C).
// Field names mirror core/internal/server/organization_starter_catalog.go
// and core/internal/bootstrap/template_starter.go.

export type OrganizationStarterSource = "bundle" | "starter";

export type OrganizationStarterPersona =
    | "studio"
    | "saas_product_team"
    | "operations"
    | "regulated_enterprise";

export interface OrganizationStarterLabels {
    label: string;
    persona?: OrganizationStarterPersona;
    order: number;
    sample_name?: string;
    sample_purpose?: string;
    soma_intro?: string;
}

export interface OrganizationDeliveryOutputs {
    retain: boolean;
    proof_required: boolean;
}

export interface OrganizationDeliveryDefaults {
    work_mode: string;
    approval_policy_ref: string;
    approval_posture: string;
    audit: string;
    outputs: OrganizationDeliveryOutputs;
    max_team_members: number;
}

export interface OrganizationStarterDeliverySummary {
    work_mode_label: string;
    approval_posture: string;
    output_proof: string;
    team_shape: string;
}

// OrganizationStarterTemplate is one item of
// GET /api/v1/templates?view=organization-starters. It carries every
// OrganizationTemplateSummary field plus the RC-004A starter projection.
export interface OrganizationStarterTemplate extends OrganizationTemplateSummary {
    source: OrganizationStarterSource;
    available: boolean;
    unavailable_reason?: string;
    starter?: OrganizationStarterLabels;
    delivery_defaults?: OrganizationDeliveryDefaults;
    delivery_summary?: OrganizationStarterDeliverySummary;
}
