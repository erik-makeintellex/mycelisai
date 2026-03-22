"use client";

export interface LastOrganizationRef {
    id: string;
    name: string;
}

const LAST_ORGANIZATION_CHANGED_EVENT = "mycelis:last-organization-changed";
const LAST_ORGANIZATION_ID_KEY = "mycelis-last-organization-id";
const LAST_ORGANIZATION_NAME_KEY = "mycelis-last-organization-name";

export function readLastOrganization(): LastOrganizationRef | null {
    if (typeof window === "undefined") {
        return null;
    }
    const id = window.localStorage.getItem(LAST_ORGANIZATION_ID_KEY);
    if (!id) {
        return null;
    }
    return {
        id,
        name: window.localStorage.getItem(LAST_ORGANIZATION_NAME_KEY) || "Current Organization",
    };
}

export function rememberLastOrganization(organization: LastOrganizationRef) {
    if (typeof window === "undefined") {
        return;
    }
    window.localStorage.setItem(LAST_ORGANIZATION_ID_KEY, organization.id);
    window.localStorage.setItem(LAST_ORGANIZATION_NAME_KEY, organization.name);
    window.dispatchEvent(
        new CustomEvent<LastOrganizationRef>(LAST_ORGANIZATION_CHANGED_EVENT, {
            detail: organization,
        }),
    );
}

export function clearLastOrganization() {
    if (typeof window === "undefined") {
        return;
    }
    window.localStorage.removeItem(LAST_ORGANIZATION_ID_KEY);
    window.localStorage.removeItem(LAST_ORGANIZATION_NAME_KEY);
}

export function subscribeLastOrganizationChange(listener: (organization: LastOrganizationRef | null) => void) {
    if (typeof window === "undefined") {
        return () => {};
    }

    const handleChange = (event: Event) => {
        const detail = (event as CustomEvent<LastOrganizationRef>).detail;
        if (!detail?.id) {
            listener(readLastOrganization());
            return;
        }
        listener({ id: detail.id, name: detail.name || "Current Organization" });
    };

    window.addEventListener(LAST_ORGANIZATION_CHANGED_EVENT, handleChange as EventListener);
    return () => {
        window.removeEventListener(LAST_ORGANIZATION_CHANGED_EVENT, handleChange as EventListener);
    };
}
