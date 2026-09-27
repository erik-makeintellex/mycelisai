export type SelectOption = { value: string; label: string };

export type DeploymentContextForm = {
    knowledge_class: string;
    title: string;
    source_label: string;
    content: string;
    content_type: string;
    source_kind: string;
    visibility: string;
    sensitivity_class: string;
    trust_class: string;
    content_domain: string;
    target_goal_sets: string;
    tags: string;
};

export const KNOWLEDGE_CLASS_OPTIONS: SelectOption[] = [
    { value: "company_knowledge", label: "Company knowledge" },
    { value: "customer_context", label: "Customer / project context" },
    { value: "user_private_context", label: "Restricted (only you)" },
    { value: "soma_operating_context", label: "Soma operating guidance (admin)" },
    { value: "reflection_synthesis", label: "Lessons & retrospectives" },
];

// Stored values stay stable; only labels changed. diary_* values were renamed to worklog_*.
export const SOURCE_KIND_OPTIONS: SelectOption[] = [
    { value: "user_note", label: "Note" },
    { value: "user_document", label: "Document" },
    { value: "user_record", label: "Record" },
    { value: "worklog_entry", label: "Work log entry" },
    { value: "finance_record", label: "Financial record" },
    { value: "lesson", label: "Lesson learned" },
    { value: "inferred_pattern", label: "Observed pattern" },
    { value: "contradiction", label: "Conflicting information" },
    { value: "trajectory_shift", label: "Direction change" },
    { value: "meta_observation", label: "Process observation" },
    { value: "synthesis_note", label: "Summary note" },
    { value: "workspace_file", label: "Workspace file" },
    { value: "web_research", label: "Web research" },
];

export const CONTENT_DOMAIN_OPTIONS: SelectOption[] = [
    { value: "", label: "Not set" },
    { value: "operations", label: "Operations" },
    { value: "private_records", label: "Restricted records" },
    { value: "worklog", label: "Work log" },
    { value: "finance", label: "Finance" },
    { value: "health", label: "Health & safety" },
    { value: "legal", label: "Legal" },
    { value: "creative", label: "Marketing & creative" },
    { value: "reflection", label: "Lessons & retrospectives" },
];

export const VISIBILITY_OPTIONS: SelectOption[] = [
    { value: "global", label: "Whole organization" },
    { value: "team", label: "My team" },
    { value: "private", label: "Only me" },
];

export const SENSITIVITY_OPTIONS: SelectOption[] = [
    { value: "role_scoped", label: "Role scoped" },
    { value: "team_scoped", label: "Team scoped" },
    { value: "restricted", label: "Restricted" },
];

export const TRUST_OPTIONS: SelectOption[] = [
    { value: "user_provided", label: "User provided" },
    { value: "validated_external", label: "Validated external" },
    { value: "bounded_external", label: "Bounded external" },
    { value: "trusted_internal", label: "Trusted internal" },
];

/** Simple defaults: admins save company knowledge, everyone else customer/project context. */
export function defaultForm(isAdmin: boolean): DeploymentContextForm {
    return {
        knowledge_class: isAdmin ? "company_knowledge" : "customer_context",
        title: "",
        source_label: "",
        content: "",
        content_type: "text/markdown",
        source_kind: "user_note",
        visibility: "global",
        sensitivity_class: "role_scoped",
        trust_class: isAdmin ? "trusted_internal" : "user_provided",
        content_domain: "",
        target_goal_sets: "",
        tags: "",
    };
}

export function optionLabel(options: SelectOption[], value: string) {
    return options.find((option) => option.value === value)?.label ?? value.replaceAll("_", " ");
}

const REFLECTION_KINDS = ["lesson", "inferred_pattern", "contradiction", "trajectory_shift", "meta_observation", "synthesis_note"];

/** Applies the class-specific scope the backend would otherwise enforce. */
export function withKnowledgeClass(current: DeploymentContextForm, knowledgeClass: string): DeploymentContextForm {
    const next = { ...current, knowledge_class: knowledgeClass };
    if (knowledgeClass === "user_private_context") {
        return { ...next, visibility: "private", sensitivity_class: "restricted", content_domain: current.content_domain || "private_records" };
    }
    if (knowledgeClass === "reflection_synthesis") {
        return {
            ...next, visibility: "private", sensitivity_class: "restricted", trust_class: "trusted_internal", content_domain: "reflection",
            source_kind: REFLECTION_KINDS.includes(current.source_kind) ? current.source_kind : "synthesis_note",
        };
    }
    if (knowledgeClass === "soma_operating_context") {
        return { ...next, visibility: "global", sensitivity_class: "restricted", trust_class: "trusted_internal" };
    }
    return {
        ...next,
        sensitivity_class: current.sensitivity_class === "restricted" ? "role_scoped" : current.sensitivity_class,
        source_kind: REFLECTION_KINDS.includes(current.source_kind) ? "user_note" : current.source_kind,
    };
}

export function splitList(value: string) {
    return value.split(",").map((item) => item.trim()).filter(Boolean);
}
