import type {
    OrganizationAutomationItem,
    OrganizationHomePayload,
    OrganizationLearningInsightItem,
    OrganizationLoopActivityItem,
} from "@/lib/organizations";

export function formatConfiguredCount(count: number, label: string) {
    if (count === 0) {
        return "Not configured yet";
    }
    return `${count} ${label}${count === 1 ? "" : "s"}`;
}

export function advisorSummary(count: number, teamLeadName: string) {
    if (count === 0) {
        return `Soma is handling planning and review directly for now. Advisor support will appear here when ${teamLeadName} needs a second set of eyes.`;
    }
    if (count === 1) {
        return `1 Advisor is ready to help Soma and ${teamLeadName} with review, priorities, and decision support.`;
    }
    return `${count} Advisors are ready to help Soma and ${teamLeadName} review decisions and keep the organization aligned.`;
}

export function advisorSupportItems(count: number) {
    if (count === 0) {
        return [
            "Review support appears here",
            "Advisors help with decisions and checks",
            "Try reviewing your organization setup",
        ];
    }
    return ["Planning review", "Decision support", "Priority checks"].slice(0, Math.max(2, Math.min(count + 1, 3)));
}

export function departmentSummary(count: number, specialistCount: number, teamLeadName: string) {
    if (count === 0) {
        return `Soma can still shape the first working lane through ${teamLeadName} before Departments are configured. Departments will appear here once the organization has a clear first focus.`;
    }
    return `${count} Departments and ${formatConfiguredCount(specialistCount, "Specialist").toLowerCase()} are visible here so Soma and ${teamLeadName} can work with a clear delivery structure.`;
}

export function departmentSupportItems(organization: OrganizationHomePayload) {
    const visibleRoles = (organization.departments ?? [])
        .flatMap((department) =>
            (department.agent_type_profiles ?? []).map((profile) => ({
                label: profile.name,
                detail: profile.helps_with,
            })),
        )
        .slice(0, 3);

    if (visibleRoles.length > 0) {
        return visibleRoles;
    }

    return [
        {
            label: organization.start_mode === "template" && organization.template_name
                ? `Started from ${organization.template_name}`
                : "Started from Empty",
            detail: "This organization already has a visible starting structure for Soma to work through.",
        },
        {
            label: formatConfiguredCount(organization.specialist_count, "Specialist"),
            detail: "Specialist roles will appear here with short purpose labels as the structure becomes visible.",
        },
        {
            label: organization.department_count > 0 ? "Open the current team structure" : "Try reviewing your organization setup",
            detail: "Use the Department view to inspect the current lanes and role bindings in more detail.",
        },
    ];
}

export function formatAutomationCount(count: number, loading: boolean, error: string | null) {
    if (error) {
        return "Unavailable right now";
    }
    if (loading && count === 0) {
        return "Checking now";
    }
    if (count === 0) {
        return "Not configured yet";
    }
    return `${count} ${count === 1 ? "Automation" : "Automations"}`;
}

export function automationStatusLabel(loading: boolean, error: string | null) {
    if (error) {
        return "Read only";
    }
    if (loading) {
        return "Refreshing";
    }
    return "Read only";
}

export function automationSummary(count: number, teamLeadName: string) {
    const mentalModel = "This system runs ongoing reviews and checks to help your organization improve over time.";
    if (count === 0) {
        return `${mentalModel} Soma will show those ongoing reviews and checks here as this AI Organization becomes more active.`;
    }
    if (count === 1) {
        return `${mentalModel} 1 Automation is visible here so Soma can explain what ongoing review is supporting this AI Organization through ${teamLeadName}.`;
    }
    return `${mentalModel} ${count} Automations are visible here so Soma can explain what ongoing reviews and checks are supporting this AI Organization through ${teamLeadName}.`;
}

export function automationSupportItems(items: OrganizationAutomationItem[], loading: boolean, error: string | null) {
    if (error) {
        return ["Automations unavailable", "Workspace still ready", "Read only"];
    }
    if (loading && items.length === 0) {
        return ["Checking Reviews", "Checking Watchers", "Read only"];
    }
    if (items.length === 0) {
        return ["Reviews appear here", "Watchers appear here", "Read only"];
    }

    return items.slice(0, 3).map((item) => {
        if (item.trigger_type === "scheduled") {
            return `${item.name} • Scheduled`;
        }
        return `${item.name} • Event-driven`;
    });
}

export function aiEngineSummary(summary: string) {
    const normalized = summary.trim();
    if (!normalized || normalized === "Set up later in Advanced mode") {
        return "The current AI Engine Settings keep the organization on a simple starter profile until deeper tuning is needed.";
    }
    return `The current AI Engine Settings profile is ${normalized.toLowerCase()} and shapes how the organization responds, plans, and carries work forward.`;
}

export function aiEngineSupportItems(summary: string) {
    if (summary.trim() === "Set up later in Advanced mode") {
        return ["Response style", "Planning depth", "Inspect only for now"];
    }
    return [summary, "Response style", "Planning depth"];
}

export function responseContractSummary(summary: string) {
    const normalized = summary.trim();
    if (!normalized) {
        return "The current Response Style keeps the organization on a clear, steady default for safe day-to-day guidance.";
    }
    return `The current Response Style is ${normalized.toLowerCase()}, which shapes how Soma presents tone, structure, and detail.`;
}

export function responseContractSupportItems(summary: string) {
    if (!summary.trim()) {
        return ["Tone", "Structure", "Verbosity"];
    }
    return [summary, "Tone", "Structure", "Verbosity"];
}

export function learningContextSummary(summary: string) {
    const normalized = summary.trim();
    if (!normalized || normalized === "Set up later in Advanced mode") {
        return "Memory & Continuity stay on a simple starter posture so Soma can keep working continuity without turning every conversation into durable memory.";
    }
    return `Memory & Continuity are currently ${normalized.toLowerCase()}, which shapes what Soma retains for later recall and what stays as temporary working context.`;
}

export function learningContextSupportItems(summary: string) {
    if (summary.trim() === "Set up later in Advanced mode") {
        return ["Durable memory recall", "Temporary planning continuity", "Inspect only for now"];
    }
    return [summary, "Durable memory recall", "Temporary planning continuity"];
}

export function parseKnownTimestamp(value?: string) {
    if (!value) {
        return 0;
    }
    const parsed = Date.parse(value);
    return Number.isNaN(parsed) ? 0 : parsed;
}

export function panelsUpdatedSince(
    since: number,
    recentActivity: OrganizationLoopActivityItem[],
    automations: OrganizationAutomationItem[],
    learningInsights: OrganizationLearningInsightItem[],
) {
    if (!since) {
        return [];
    }

    const panels: string[] = [];
    if (recentActivity.some((item) => parseKnownTimestamp(item.last_run_at) >= since)) {
        panels.push("Recent Activity");
    }
    if (automations.some((item) => (item.recent_outcomes ?? []).some((outcome) => parseKnownTimestamp(outcome.occurred_at) >= since))) {
        panels.push("Automations");
    }
    if (learningInsights.some((item) => parseKnownTimestamp(item.observed_at) >= since)) {
        panels.push("Memory & Continuity");
    }
    return panels.length > 0 ? panels : ["Quick Checks"];
}

export function toTitleCase(value: string) {
    return value
        .split(/[\s_-]+/)
        .filter(Boolean)
        .map((segment) => segment.charAt(0).toUpperCase() + segment.slice(1))
        .join(" ");
}
