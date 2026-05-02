import { useEffect, useState } from "react";

import { extractApiData, extractApiError } from "@/lib/apiContracts";
import type {
    OrganizationAutomationItem,
    OrganizationLearningInsightItem,
    OrganizationLoopActivityItem,
} from "@/lib/organizations";

async function readPanelJson(response: Response) {
    try {
        return await response.json();
    } catch {
        return null;
    }
}

export function useOrganizationLivePanelData(organizationId: string) {
    const [recentActivity, setRecentActivity] = useState<OrganizationLoopActivityItem[]>([]);
    const [activityLoading, setActivityLoading] = useState(true);
    const [activityError, setActivityError] = useState<string | null>(null);
    const [activityReloadToken, setActivityReloadToken] = useState(0);
    const [automations, setAutomations] = useState<OrganizationAutomationItem[]>([]);
    const [automationsLoading, setAutomationsLoading] = useState(true);
    const [automationsError, setAutomationsError] = useState<string | null>(null);
    const [automationsReloadToken, setAutomationsReloadToken] = useState(0);
    const [learningInsights, setLearningInsights] = useState<OrganizationLearningInsightItem[]>([]);
    const [learningInsightsLoading, setLearningInsightsLoading] = useState(true);
    const [learningInsightsError, setLearningInsightsError] = useState<string | null>(null);
    const [learningInsightsReloadToken, setLearningInsightsReloadToken] = useState(0);

    useEffect(() => {
        let cancelled = false;

        const loadActivity = async (background: boolean) => {
            if (!background && !cancelled) {
                setActivityLoading(true);
            }
            try {
                const response = await fetch(`/api/v1/organizations/${organizationId}/loop-activity`, { cache: "no-store" });
                const payload = await readPanelJson(response);
                if (!response.ok) {
                    throw new Error(extractApiError(payload) || "Activity unavailable");
                }
                if (!cancelled) {
                    setRecentActivity(extractApiData<OrganizationLoopActivityItem[]>(payload) ?? []);
                    setActivityError(null);
                }
            } catch {
                if (!cancelled) {
                    setActivityError("Activity unavailable");
                }
            } finally {
                if (!cancelled) {
                    setActivityLoading(false);
                }
            }
        };

        void loadActivity(false);
        const intervalId = window.setInterval(() => void loadActivity(true), 15000);
        return () => {
            cancelled = true;
            window.clearInterval(intervalId);
        };
    }, [organizationId, activityReloadToken]);

    useEffect(() => {
        let cancelled = false;

        const loadAutomations = async (background: boolean) => {
            if (!background && !cancelled) {
                setAutomationsLoading(true);
            }
            try {
                const response = await fetch(`/api/v1/organizations/${organizationId}/automations`, { cache: "no-store" });
                const payload = await readPanelJson(response);
                if (!response.ok) {
                    throw new Error(extractApiError(payload) || "Automations unavailable");
                }
                if (!cancelled) {
                    setAutomations(extractApiData<OrganizationAutomationItem[]>(payload) ?? []);
                    setAutomationsError(null);
                }
            } catch {
                if (!cancelled) {
                    setAutomationsError("Automations unavailable");
                }
            } finally {
                if (!cancelled) {
                    setAutomationsLoading(false);
                }
            }
        };

        void loadAutomations(false);
        const intervalId = window.setInterval(() => void loadAutomations(true), 20000);
        return () => {
            cancelled = true;
            window.clearInterval(intervalId);
        };
    }, [organizationId, automationsReloadToken]);

    useEffect(() => {
        let cancelled = false;

        const loadLearningInsights = async (background: boolean) => {
            if (!background && !cancelled) {
                setLearningInsightsLoading(true);
            }
            try {
                const response = await fetch(`/api/v1/organizations/${organizationId}/learning-insights`, { cache: "no-store" });
                const payload = await readPanelJson(response);
                if (!response.ok) {
                    throw new Error(extractApiError(payload) || "Memory & Continuity updates unavailable");
                }
                if (!cancelled) {
                    setLearningInsights(extractApiData<OrganizationLearningInsightItem[]>(payload) ?? []);
                    setLearningInsightsError(null);
                }
            } catch {
                if (!cancelled) {
                    setLearningInsightsError("Memory & Continuity updates unavailable");
                }
            } finally {
                if (!cancelled) {
                    setLearningInsightsLoading(false);
                }
            }
        };

        void loadLearningInsights(false);
        const intervalId = window.setInterval(() => void loadLearningInsights(true), 25000);
        return () => {
            cancelled = true;
            window.clearInterval(intervalId);
        };
    }, [organizationId, learningInsightsReloadToken]);

    return {
        recentActivity,
        activityLoading,
        activityError,
        retryActivity: () => setActivityReloadToken((value) => value + 1),
        automations,
        automationsLoading,
        automationsError,
        retryAutomations: () => setAutomationsReloadToken((value) => value + 1),
        learningInsights,
        learningInsightsLoading,
        learningInsightsError,
        retryLearningInsights: () => setLearningInsightsReloadToken((value) => value + 1),
    };
}
