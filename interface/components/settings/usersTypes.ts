export type UserRole = "owner" | "operator" | "viewer";
export type AccessManagementTier = "release" | "enterprise";
export type ProductEdition = "self_hosted_release" | "self_hosted_enterprise" | "hosted_control_plane";
export type IdentityMode = "local_only" | "hybrid" | "federated";
export type SharedAgentSpecificityOwner = "root_admin" | "delegated_owner";
export type PrincipalType = "local_admin" | "break_glass_admin" | "federated_user" | "service_principal" | "user";

export type ManagedUser = {
    id: string;
    name: string;
    email: string;
    role: UserRole;
    remoteAllowed: boolean;
    status: "active" | "disabled";
};

export const STARTER_DIRECTORY_USERS: ManagedUser[] = [
    { id: "owner", name: "Admin", email: "admin@local", role: "owner", remoteAllowed: true, status: "active" },
    { id: "operator", name: "Operator", email: "operator@local", role: "operator", remoteAllowed: false, status: "active" },
    { id: "viewer", name: "Viewer", email: "viewer@local", role: "viewer", remoteAllowed: false, status: "active" },
];

export const EDITION_COPY: Record<ProductEdition, { label: string; summary: string }> = {
    self_hosted_release: {
        label: "Self-hosted release",
        summary: "Base self-hosted posture with organization access, groups, and governed Soma workflows.",
    },
    self_hosted_enterprise: {
        label: "Self-hosted enterprise",
        summary: "Adds enterprise user-directory posture and hybrid identity expectations without requiring a hosted control plane.",
    },
    hosted_control_plane: {
        label: "Hosted control plane",
        summary: "Separates advanced access management into a paid hosted layer while the runtime and Soma stay governed by the organization.",
    },
};

export const IDENTITY_COPY: Record<IdentityMode, string> = {
    local_only: "Local credentials for self-hosted environments without external identity.",
    hybrid: "Enterprise sign-in with local break-glass admins preserved for self-hosted recovery.",
    federated: "Standard SAML/OIDC-style federation with enterprise-managed sign-in.",
};

export function toRole(value: string): UserRole {
    switch (value.trim().toLowerCase()) {
        case "owner":
        case "admin":
        case "root":
            return "owner";
        case "operator":
        case "operations":
        case "user":
            return "operator";
        default:
            return "viewer";
    }
}

export function toAccessManagementTier(value: unknown): AccessManagementTier {
    return String(value).trim().toLowerCase() === "enterprise" ? "enterprise" : "release";
}

export function toProductEdition(value: unknown): ProductEdition {
    switch (String(value).trim().toLowerCase()) {
        case "self_hosted_enterprise":
        case "enterprise":
            return "self_hosted_enterprise";
        case "hosted_control_plane":
        case "hosted":
            return "hosted_control_plane";
        default:
            return "self_hosted_release";
    }
}

export function toIdentityMode(value: unknown): IdentityMode {
    switch (String(value).trim().toLowerCase()) {
        case "hybrid":
            return "hybrid";
        case "federated":
            return "federated";
        default:
            return "local_only";
    }
}

export function toSharedAgentSpecificityOwner(value: unknown): SharedAgentSpecificityOwner {
    return String(value).trim().toLowerCase() === "delegated_owner" ? "delegated_owner" : "root_admin";
}

export function toPrincipalType(value: unknown): PrincipalType {
    switch (String(value).trim().toLowerCase()) {
        case "break_glass_admin":
            return "break_glass_admin";
        case "federated_user":
            return "federated_user";
        case "service_principal":
            return "service_principal";
        case "user":
            return "user";
        default:
            return "local_admin";
    }
}

export function extractData<T>(payload: unknown): T {
    if (payload && typeof payload === "object" && "data" in payload) {
        return (payload as { data: T }).data;
    }
    return payload as T;
}
