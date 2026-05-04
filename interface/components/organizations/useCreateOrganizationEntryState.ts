"use client";

import { useRouter } from "next/navigation";
import { useEffect, useMemo, useRef, useState, useTransition } from "react";
import { extractApiData, extractApiError } from "@/lib/apiContracts";
import { readLastOrganization, rememberLastOrganization, subscribeLastOrganizationChange } from "@/lib/lastOrganization";
import type {
    OrganizationCreateRequest,
    OrganizationStartMode,
    OrganizationSummary,
    OrganizationTemplateSummary,
} from "@/lib/organizations";

type ResourceState = "loading" | "ready" | "error";

async function readJson(response: Response) {
    try {
        return await response.json();
    } catch {
        return null;
    }
}

function isDiagnosticOrganization(organization: OrganizationSummary) {
    const name = organization.name.trim();
    const purpose = organization.purpose.trim();
    return /^qa scenario\b/i.test(name)
        || /^qa\b/i.test(name)
        || /^testing setup\b/i.test(name)
        || /live governance verification/i.test(purpose)
        || /ui verification/i.test(purpose)
        || /qa_browser_/i.test(purpose);
}

export function useCreateOrganizationEntryState() {
    const router = useRouter();
    const [templates, setTemplates] = useState<OrganizationTemplateSummary[]>([]);
    const [organizations, setOrganizations] = useState<OrganizationSummary[]>([]);
    const [lastOrganization, setLastOrganization] = useState<{ id: string; name: string } | null>(null);
    const [templatesState, setTemplatesState] = useState<ResourceState>("loading");
    const [organizationsState, setOrganizationsState] = useState<ResourceState>("loading");
    const [templatesError, setTemplatesError] = useState<string | null>(null);
    const [organizationsError, setOrganizationsError] = useState<string | null>(null);
    const [templatesReloadToken, setTemplatesReloadToken] = useState(0);
    const [organizationsReloadToken, setOrganizationsReloadToken] = useState(0);
    const [selectedMode, setSelectedMode] = useState<OrganizationStartMode | null>(null);
    const [selectedTemplateId, setSelectedTemplateId] = useState<string>("");
    const [name, setName] = useState("");
    const [purpose, setPurpose] = useState("");
    const [submitError, setSubmitError] = useState<string | null>(null);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [openingOrganizationId, setOpeningOrganizationId] = useState<string | null>(null);
    const [showDiagnosticOrganizations, setShowDiagnosticOrganizations] = useState(false);
    const [, startTransition] = useTransition();
    const createSectionRef = useRef<HTMLElement | null>(null);
    const nameInputRef = useRef<HTMLInputElement | null>(null);

    const focusCreateFlow = (mode: OrganizationStartMode) => {
        setSelectedMode(mode);
        requestAnimationFrame(() => {
            createSectionRef.current?.scrollIntoView?.({ behavior: "smooth", block: "start" });
            nameInputRef.current?.focus();
        });
    };

    const rememberOrganization = (organization: Pick<OrganizationSummary, "id" | "name">) => {
        rememberLastOrganization({ id: organization.id, name: organization.name });
    };

    const openOrganization = (organization: Pick<OrganizationSummary, "id" | "name">) => {
        if (openingOrganizationId === organization.id || isSubmitting) {
            return;
        }
        setOpeningOrganizationId(organization.id);
        rememberOrganization(organization);
        startTransition(() => router.push(`/organizations/${organization.id}`));
    };

    useEffect(() => {
        setLastOrganization(readLastOrganization());
        return subscribeLastOrganizationChange((organization) => setLastOrganization(organization));
    }, []);

    useEffect(() => {
        let cancelled = false;
        const loadTemplates = async () => {
            setTemplatesState("loading");
            setTemplatesError(null);
            try {
                const response = await fetch("/api/v1/templates?view=organization-starters", { cache: "no-store" });
                const payload = await readJson(response);
                if (!response.ok) {
                    throw new Error(extractApiError(payload) || "Starter templates are unavailable right now.");
                }
                if (cancelled) {
                    return;
                }
                const nextTemplates = extractApiData<OrganizationTemplateSummary[]>(payload) || [];
                setTemplates(nextTemplates);
                setSelectedTemplateId((current) => current && nextTemplates.some((template) => template.id === current) ? current : nextTemplates[0]?.id ?? "");
                setTemplatesState("ready");
            } catch (err) {
                if (!cancelled) {
                    setTemplatesError(err instanceof Error ? err.message : "Starter templates are unavailable right now.");
                    setTemplatesState("error");
                }
            }
        };
        void loadTemplates();
        return () => {
            cancelled = true;
        };
    }, [templatesReloadToken]);

    useEffect(() => {
        let cancelled = false;
        const loadOrganizations = async () => {
            setOrganizationsState("loading");
            setOrganizationsError(null);
            try {
                const response = await fetch("/api/v1/organizations?view=summary", { cache: "no-store" });
                const payload = await readJson(response);
                if (!response.ok) {
                    throw new Error(extractApiError(payload) || "Recent AI Organizations are unavailable right now.");
                }
                if (!cancelled) {
                    setOrganizations(extractApiData<OrganizationSummary[]>(payload) || []);
                    setOrganizationsState("ready");
                }
            } catch (err) {
                if (!cancelled) {
                    setOrganizationsError(err instanceof Error ? err.message : "Recent AI Organizations are unavailable right now.");
                    setOrganizationsState("error");
                }
            }
        };
        void loadOrganizations();
        return () => {
            cancelled = true;
        };
    }, [organizationsReloadToken]);

    const selectedTemplate = templates.find((template) => template.id === selectedTemplateId) ?? null;
    const isTemplateMode = selectedMode === "template";
    const canSubmit = name.trim().length > 1 && purpose.trim().length > 8 && selectedMode !== null && (!isTemplateMode || (templatesState === "ready" && !!selectedTemplate));
    const regularOrganizations = useMemo(() => organizations.filter((organization) => !isDiagnosticOrganization(organization)), [organizations]);
    const diagnosticOrganizations = useMemo(() => organizations.filter((organization) => isDiagnosticOrganization(organization)), [organizations]);
    const visibleDiagnosticOrganizations = useMemo(() => diagnosticOrganizations.slice(0, 2), [diagnosticOrganizations]);
    const hiddenDiagnosticCount = Math.max(0, diagnosticOrganizations.length - visibleDiagnosticOrganizations.length);
    const testingSetupCountLabel = visibleDiagnosticOrganizations.length === 1 ? "testing setup" : "testing setups";
    const visibleOrganizations = showDiagnosticOrganizations ? [...regularOrganizations, ...visibleDiagnosticOrganizations] : regularOrganizations;
    const creationReadinessMessage =
        selectedMode === null ? "Choose Start from template or Start empty first."
            : name.trim().length <= 1 ? "Add a clear AI Organization name."
            : purpose.trim().length <= 8 ? "Add a short purpose so Soma knows what this organization is for."
            : isTemplateMode && templatesState === "loading" ? "Wait for starter templates to finish loading."
            : isTemplateMode && templatesState === "error" ? "Starter templates are unavailable right now. Retry or start empty."
            : isTemplateMode && !selectedTemplate ? "Choose a starter template to continue."
            : null;

    const handleCreate = async () => {
        if (!canSubmit || isSubmitting || selectedMode === null) {
            return;
        }
        const payload: OrganizationCreateRequest = {
            name: name.trim(),
            purpose: purpose.trim(),
            start_mode: selectedMode,
            template_id: isTemplateMode ? selectedTemplate?.id : undefined,
        };
        setIsSubmitting(true);
        setSubmitError(null);
        try {
            const response = await fetch("/api/v1/organizations", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload),
            });
            const responsePayload = await readJson(response);
            if (!response.ok) {
                throw new Error(extractApiError(responsePayload) || "Unable to create AI Organization.");
            }
            const created = extractApiData<OrganizationSummary>(responsePayload);
            if (!created?.id) {
                throw new Error("AI Organization created, but the workspace route is unavailable right now.");
            }
            rememberOrganization(created);
            startTransition(() => router.push(`/organizations/${created.id}`));
        } catch (err) {
            setSubmitError(err instanceof Error ? err.message : "Unable to create AI Organization.");
        } finally {
            setIsSubmitting(false);
        }
    };

    return {
        templates, organizations, lastOrganization, templatesState, organizationsState, templatesError, organizationsError,
        selectedMode, selectedTemplateId, selectedTemplate, name, purpose, submitError, isSubmitting,
        openingOrganizationId, showDiagnosticOrganizations, diagnosticOrganizations, visibleDiagnosticOrganizations,
        hiddenDiagnosticCount, testingSetupCountLabel, visibleOrganizations, isTemplateMode, canSubmit,
        creationReadinessMessage, createSectionRef, nameInputRef, setSelectedTemplateId, setSelectedMode,
        setName, setPurpose, setShowDiagnosticOrganizations, focusCreateFlow, rememberOrganization,
        openOrganization, handleCreate, reloadTemplates: () => setTemplatesReloadToken((value) => value + 1),
        reloadOrganizations: () => setOrganizationsReloadToken((value) => value + 1),
        showRecentOrganizations: organizationsState !== "ready" || organizations.length > 0,
    };
}
