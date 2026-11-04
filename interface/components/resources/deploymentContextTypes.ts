export type DeploymentContextEntry = {
    artifact_id: string;
    knowledge_class: string;
    title: string;
    source_label: string;
    source_kind: string;
    visibility: string;
    sensitivity_class: string;
    trust_class: string;
    chunk_count: number;
    vector_count: number;
    content_preview: string;
    content_length: number;
    content_domain?: string;
    target_goal_sets?: string[];
    created_at: string;
};

export type LoadResponse = {
    artifact_id: string;
    knowledge_class: string;
    title: string;
    chunk_count: number;
    vector_count: number;
};

export type DeploymentContextFormState = {
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

export const INITIAL_DEPLOYMENT_CONTEXT_FORM: DeploymentContextFormState = {
    knowledge_class: "user_private_context",
    title: "",
    source_label: "",
    content: "",
    content_type: "text/markdown",
    source_kind: "user_record",
    visibility: "private",
    sensitivity_class: "restricted",
    trust_class: "user_provided",
    content_domain: "private_records",
    target_goal_sets: "",
    tags: "user-private-context",
};

export const KNOWLEDGE_CLASS_OPTIONS = [
    { value: "user_private_context", label: "Private user content" },
    { value: "customer_context", label: "Customer context" },
    { value: "company_knowledge", label: "Approved company knowledge" },
    { value: "soma_operating_context", label: "Admin-shaped Soma context" },
    { value: "reflection_synthesis", label: "Reflection / synthesis memory" },
];

export const SOURCE_KIND_OPTIONS = [
    { value: "user_document", label: "User document" },
    { value: "user_note", label: "User note" },
    { value: "user_record", label: "User/private record" },
    { value: "diary_entry", label: "Diary entry" },
    { value: "finance_record", label: "Finance record" },
    { value: "lesson", label: "Distilled lesson" },
    { value: "inferred_pattern", label: "Inferred pattern" },
    { value: "contradiction", label: "Contradiction" },
    { value: "trajectory_shift", label: "Trajectory shift" },
    { value: "meta_observation", label: "Meta-observation" },
    { value: "synthesis_note", label: "Synthesis note" },
    { value: "workspace_file", label: "Workspace file" },
    { value: "web_research", label: "Web research" },
];

export const CONTENT_DOMAIN_OPTIONS = [
    { value: "private_records", label: "Private records" },
    { value: "diary", label: "Diary / journal" },
    { value: "finance", label: "Finance" },
    { value: "health", label: "Health" },
    { value: "legal", label: "Legal" },
    { value: "creative", label: "Creative" },
    { value: "operations", label: "Operations" },
    { value: "reflection", label: "Reflection / synthesis" },
];

export const VISIBILITY_OPTIONS = [
    { value: "global", label: "Global" },
    { value: "team", label: "Team" },
    { value: "private", label: "Private" },
];

export const SENSITIVITY_OPTIONS = [
    { value: "role_scoped", label: "Role scoped" },
    { value: "team_scoped", label: "Team scoped" },
    { value: "restricted", label: "Restricted" },
];

export const TRUST_OPTIONS = [
    { value: "user_provided", label: "User provided" },
    { value: "validated_external", label: "Validated external" },
    { value: "bounded_external", label: "Bounded external" },
    { value: "trusted_internal", label: "Trusted internal" },
];

export const INPUT_CLASS = "w-full rounded-lg border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-main placeholder:text-cortex-text-muted/60 focus:outline-none focus:border-cortex-primary";

export function readFileAsText(file: File): Promise<string> {
    if (typeof file.text === "function") {
        return file.text();
    }
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => resolve(typeof reader.result === "string" ? reader.result : "");
        reader.onerror = () => reject(reader.error ?? new Error("Unable to read uploaded file."));
        reader.readAsText(file);
    });
}

export function knowledgeClassLabel(value: string) {
    switch (value) {
        case "company_knowledge":
            return "approved company knowledge";
        case "soma_operating_context":
            return "admin-shaped Soma context";
        case "user_private_context":
            return "private user content";
        case "reflection_synthesis":
            return "reflection / synthesis memory";
        default:
            return "customer context";
    }
}

export function isReflectionSourceKind(value: string) {
    return ["lesson", "inferred_pattern", "contradiction", "trajectory_shift", "meta_observation", "synthesis_note"].includes(value);
}

export function applyKnowledgeClassPreset(
    current: DeploymentContextFormState,
    knowledgeClass: string,
): DeploymentContextFormState {
    if (knowledgeClass === "user_private_context") {
        return {
            ...current,
            knowledge_class: knowledgeClass,
            source_kind: current.source_kind === "user_document" || isReflectionSourceKind(current.source_kind) ? "user_record" : current.source_kind,
            visibility: "private",
            sensitivity_class: "restricted",
            content_domain: current.content_domain || "private_records",
            tags: current.tags.trim() ? current.tags : "user-private-context",
        };
    }
    if (knowledgeClass === "reflection_synthesis") {
        return {
            ...current,
            knowledge_class: knowledgeClass,
            source_kind: current.source_kind === "user_document" || current.source_kind === "user_record" ? "synthesis_note" : current.source_kind,
            visibility: "private",
            sensitivity_class: "restricted",
            trust_class: "trusted_internal",
            content_domain: "reflection",
            tags: current.tags.trim() ? current.tags : "reflection-synthesis-memory",
        };
    }
    if (knowledgeClass === "soma_operating_context") {
        return {
            ...current,
            knowledge_class: knowledgeClass,
            source_kind: current.source_kind === "user_record" || isReflectionSourceKind(current.source_kind) ? "user_note" : current.source_kind,
            visibility: "global",
            sensitivity_class: "restricted",
            trust_class: "trusted_internal",
            tags: current.tags.trim() ? current.tags : "soma-operating-context",
        };
    }
    return {
        ...current,
        knowledge_class: knowledgeClass,
        visibility: current.visibility === "private" ? "global" : current.visibility,
        sensitivity_class: current.sensitivity_class === "restricted" ? "role_scoped" : current.sensitivity_class,
        source_kind: current.source_kind === "user_record" || isReflectionSourceKind(current.source_kind) ? "user_document" : current.source_kind,
    };
}
