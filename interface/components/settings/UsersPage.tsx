"use client";

import React, { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { Shield, User, UserPlus, Trash2, RefreshCw, Users } from "lucide-react";

type UserRole = "owner" | "operator" | "viewer";
type AccessManagementTier = "release" | "enterprise";
type ProductEdition = "self_hosted_release" | "self_hosted_enterprise" | "hosted_control_plane";
type IdentityMode = "local_only" | "hybrid" | "federated";
type SharedAgentSpecificityOwner = "root_admin" | "delegated_owner";
type PrincipalType = "local_admin" | "break_glass_admin" | "federated_user" | "service_principal" | "user";

type ManagedUser = {
    id: string;
    name: string;
    email: string;
    role: UserRole;
    remoteAllowed: boolean;
    status: "active" | "disabled";
};

const STARTER_DIRECTORY_USERS: ManagedUser[] = [
    { id: "owner", name: "Admin", email: "admin@local", role: "owner", remoteAllowed: true, status: "active" },
    { id: "operator", name: "Operator", email: "operator@local", role: "operator", remoteAllowed: false, status: "active" },
    { id: "viewer", name: "Viewer", email: "viewer@local", role: "viewer", remoteAllowed: false, status: "active" },
];

const EDITION_COPY: Record<ProductEdition, { label: string; summary: string }> = {
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

const IDENTITY_COPY: Record<IdentityMode, string> = {
    local_only: "Local credentials for self-hosted environments without external identity.",
    hybrid: "Enterprise sign-in with local break-glass admins preserved for self-hosted recovery.",
    federated: "Standard SAML/OIDC-style federation with enterprise-managed sign-in.",
};

function toRole(value: string): UserRole {
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

function toAccessManagementTier(value: unknown): AccessManagementTier {
    return String(value).trim().toLowerCase() === "enterprise" ? "enterprise" : "release";
}

function toProductEdition(value: unknown): ProductEdition {
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

function toIdentityMode(value: unknown): IdentityMode {
    switch (String(value).trim().toLowerCase()) {
        case "hybrid":
            return "hybrid";
        case "federated":
            return "federated";
        default:
            return "local_only";
    }
}

function toSharedAgentSpecificityOwner(value: unknown): SharedAgentSpecificityOwner {
    return String(value).trim().toLowerCase() === "delegated_owner" ? "delegated_owner" : "root_admin";
}

function toPrincipalType(value: unknown): PrincipalType {
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

function extractData<T>(payload: unknown): T {
    if (payload && typeof payload === "object" && "data" in payload) {
        return (payload as { data: T }).data;
    }
    return payload as T;
}

export default function UsersPage() {
    const [users, setUsers] = useState<ManagedUser[]>(STARTER_DIRECTORY_USERS);
    const [name, setName] = useState("");
    const [email, setEmail] = useState("");
    const [role, setRole] = useState<UserRole>("viewer");
    const [remoteAllowed, setRemoteAllowed] = useState(false);
    const [syncing, setSyncing] = useState(false);
    const [notice, setNotice] = useState<string | null>(null);
    const [currentUser, setCurrentUser] = useState<ManagedUser | null>(null);
    const [accessManagementTier, setAccessManagementTier] = useState<AccessManagementTier>("release");
    const [productEdition, setProductEdition] = useState<ProductEdition>("self_hosted_release");
    const [identityMode, setIdentityMode] = useState<IdentityMode>("local_only");
    const [sharedAgentSpecificityOwner, setSharedAgentSpecificityOwner] = useState<SharedAgentSpecificityOwner>("root_admin");
    const [principalType, setPrincipalType] = useState<PrincipalType>("local_admin");
    const [authSource, setAuthSource] = useState("local_api_key");
    const [effectiveRole, setEffectiveRole] = useState<UserRole>("owner");
    const [breakGlass, setBreakGlass] = useState(false);
    const [savingAccessModel, setSavingAccessModel] = useState(false);

    const activeCount = useMemo(() => users.filter((u) => u.status === "active").length, [users]);
    const currentUserRole = currentUser?.role ?? "owner";
    const showsEnterpriseDirectory = accessManagementTier === "enterprise";
    const canManageEnterpriseDirectory = showsEnterpriseDirectory && currentUserRole === "owner";
    const canManageAccessModel = false;

    const refreshCurrentUser = async () => {
        setSyncing(true);
        try {
            const res = await fetch("/api/v1/user/me", { cache: "no-store" });
            if (!res.ok) return;
            const payload = await res.json();
            const me = extractData<{
                id?: string;
                email?: string;
                role?: string;
                name?: string;
                effective_role?: string;
                principal_type?: string;
                auth_source?: string;
                break_glass?: boolean;
                settings?: {
                    access_management_tier?: string;
                    product_edition?: string;
                    identity_mode?: string;
                    shared_agent_specificity_owner?: string;
                };
            }>(payload);
            const meID = me?.id || "me";
            const meName = me?.name || me?.email || "Current User";
            const meEmail = me?.email || "me@local";
            const meRole = toRole(me?.role || "owner");
            const meRemoteAllowed = meRole === "owner";
            setEffectiveRole(toRole(me?.effective_role || me?.role || "owner"));
            setPrincipalType(toPrincipalType(me?.principal_type));
            setAuthSource(String(me?.auth_source || "local_api_key").trim().toLowerCase() || "local_api_key");
            setBreakGlass(Boolean(me?.break_glass));
            const settings = me?.settings || {};
            setAccessManagementTier(toAccessManagementTier(settings.access_management_tier));
            setProductEdition(toProductEdition(settings.product_edition));
            setIdentityMode(toIdentityMode(settings.identity_mode));
            setSharedAgentSpecificityOwner(toSharedAgentSpecificityOwner(settings.shared_agent_specificity_owner));
            setCurrentUser({ id: meID, name: meName, email: meEmail, role: meRole, remoteAllowed: meRemoteAllowed, status: "active" });
            setUsers((prev) => {
                const existing = prev.find((u) => u.id === meID);
                if (existing) {
                    return prev.map((u) =>
                        u.id === meID ? { ...u, name: meName, email: meEmail, role: meRole, remoteAllowed: meRemoteAllowed, status: "active" } : u,
                    );
                }
                return [{ id: meID, name: meName, email: meEmail, role: meRole, remoteAllowed: meRemoteAllowed, status: "active" }, ...prev];
            });
        } catch {
            // best-effort sync; keep local UI usable when endpoint is unavailable
        } finally {
            setSyncing(false);
        }
    };

    useEffect(() => {
        refreshCurrentUser();
    }, []);

    const createUser = () => {
        if (!name.trim() || !email.trim()) return;
        const id = `local-${Date.now()}`;
        setUsers((prev) => [
            ...prev,
            {
                id,
                name: name.trim(),
                email: email.trim(),
                role,
                remoteAllowed,
                status: "active",
            },
        ]);
        setName("");
        setEmail("");
        setRole("viewer");
        setRemoteAllowed(false);
        setNotice("User added to local UI state. Connect auth backend to persist user directory.");
    };

    const updateUser = (id: string, patch: Partial<ManagedUser>) => {
        setUsers((prev) => prev.map((u) => (u.id === id ? { ...u, ...patch } : u)));
    };

    const removeUser = (id: string) => {
        setUsers((prev) => prev.filter((u) => u.id !== id));
    };

    const saveAccessModel = async () => {
        if (!canManageAccessModel) return;
        const nextTier: AccessManagementTier = productEdition === "self_hosted_release" ? "release" : "enterprise";
        setSavingAccessModel(true);
        setNotice(null);
        try {
            const res = await fetch("/api/v1/user/settings", {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    access_management_tier: nextTier,
                    product_edition: productEdition,
                    identity_mode: identityMode,
                    shared_agent_specificity_owner: sharedAgentSpecificityOwner,
                }),
            });
            if (!res.ok) {
                throw new Error("save failed");
            }
            setAccessManagementTier(nextTier);
            setNotice("Deployment access model saved for this environment owner view.");
        } catch {
            setNotice("Could not save the deployment access model. Keep using the current review posture until the backend is reachable.");
        } finally {
            setSavingAccessModel(false);
        }
    };

    return (
        <div className="space-y-6">
            <section className="rounded-xl border border-cortex-border bg-cortex-surface/60 p-4 space-y-4" data-testid="deployment-access-model">
                <div className="space-y-1">
                    <div className="flex items-center gap-2">
                        <Shield className="w-4 h-4 text-cortex-primary" />
                        <h3 className="text-sm font-semibold text-cortex-text-main uppercase tracking-wider">Deployment Access Model</h3>
                    </div>
                    <p className="text-sm leading-6 text-cortex-text-muted">
                        This deployment-owned posture is reported from runtime configuration. Base self-hosting, self-hosted enterprise, and hosted control-plane layers stay visible here without making edition or identity mode user-editable.
                    </p>
                </div>

                <div className="grid gap-3 lg:grid-cols-3">
                    {(Object.keys(EDITION_COPY) as ProductEdition[]).map((edition) => (
                        <button
                            key={edition}
                            type="button"
                            onClick={() => canManageAccessModel && setProductEdition(edition)}
                            disabled={!canManageAccessModel}
                            className={`rounded-2xl border px-4 py-4 text-left transition-colors ${
                                productEdition === edition ? "border-cortex-primary/40 bg-cortex-primary/10" : "border-cortex-border bg-cortex-bg"
                            } ${!canManageAccessModel ? "cursor-default" : "hover:border-cortex-primary/30"}`}
                        >
                            <p className="text-sm font-semibold text-cortex-text-main">{EDITION_COPY[edition].label}</p>
                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{EDITION_COPY[edition].summary}</p>
                        </button>
                    ))}
                </div>

                <div className="grid gap-4 lg:grid-cols-2">
                    <label className="flex flex-col gap-1 text-sm">
                        <span className="text-cortex-text-main font-semibold">Identity Mode</span>
                        <select
                            aria-label="Identity Mode"
                            value={identityMode}
                            onChange={(e) => setIdentityMode(toIdentityMode(e.target.value))}
                            disabled={!canManageAccessModel}
                            className="px-3 py-2 rounded bg-cortex-bg border border-cortex-border text-cortex-text-main"
                        >
                            <option value="local_only">Local only</option>
                            <option value="hybrid">Hybrid</option>
                            <option value="federated">Federated</option>
                        </select>
                        <span className="text-xs leading-5 text-cortex-text-muted">{IDENTITY_COPY[identityMode]}</span>
                    </label>
                    <label className="flex flex-col gap-1 text-sm">
                        <span className="text-cortex-text-main font-semibold">Shared Agent Specificity Owner</span>
                        <select
                            aria-label="Shared Agent Specificity Owner"
                            value={sharedAgentSpecificityOwner}
                            onChange={(e) => setSharedAgentSpecificityOwner(toSharedAgentSpecificityOwner(e.target.value))}
                            disabled={!canManageAccessModel}
                            className="px-3 py-2 rounded bg-cortex-bg border border-cortex-border text-cortex-text-main"
                        >
                            <option value="root_admin">Root admin only</option>
                            <option value="delegated_owner">Delegated owner</option>
                        </select>
                        <span className="text-xs leading-5 text-cortex-text-muted">
                            Shared Soma and specialist output specificity stays organization-owned. Ordinary user chats can request temporary formatting, but they do not redefine shared output posture.
                        </span>
                    </label>
                </div>

                <div className="rounded-lg border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted space-y-2">
                    <p className="font-medium text-cortex-text-main">Shared Soma ownership</p>
                    <p className="leading-6">
                        <span className="font-medium text-cortex-text-main">Soma is one organization-owned persona.</span> Root-admin or delegated-owner interaction can shape durable organization-level Soma context, while ordinary user conversations remain private or audience-scoped unless explicitly promoted.
                    </p>
                    <p className="leading-6">
                        Current principal contract: <span className="font-mono text-cortex-text-main">{principalType}</span> via <span className="font-mono text-cortex-text-main">{authSource}</span>, effective role <span className="font-mono text-cortex-text-main">{effectiveRole}</span>{breakGlass ? " with break-glass recovery active." : "."}
                    </p>
                    <p className="leading-6">
                        {identityMode === "hybrid"
                            ? "Hybrid mode keeps enterprise sign-in for everyday use and preserves local break-glass admins for self-hosted recovery."
                            : identityMode === "federated"
                              ? "Federated mode represents the enterprise SAML/OIDC path. Break-glass local admins should still exist in self-hosted environments even when not exposed in the default workflow."
                              : "Local-only mode keeps the environment self-contained for smaller self-hosted deployments and early-stage review environments."}
                    </p>
                </div>

                <div className="flex items-center justify-between gap-3">
                    <p className="text-xs leading-5 text-cortex-text-muted">
                        Current access layer resolves to <span className="font-mono text-cortex-text-main">{productEdition === "self_hosted_release" ? "release" : "enterprise"}</span> based on the selected product edition.
                    </p>
                    <button
                        type="button"
                        onClick={saveAccessModel}
                        disabled={!canManageAccessModel || savingAccessModel}
                        className="px-3 py-2 rounded border border-cortex-primary/30 text-cortex-primary text-xs font-mono hover:bg-cortex-primary/10 disabled:opacity-50"
                        data-testid="save-access-model"
                    >
                        {savingAccessModel ? "Saving..." : "Deployment-owned"}
                    </button>
                </div>
            </section>

            <section className="rounded-xl border border-cortex-border bg-cortex-surface/60 p-4 space-y-4" data-testid="organization-access-layer">
                <div className="flex items-center justify-between gap-3">
                    <div className="space-y-1">
                        <div className="flex items-center gap-2">
                            <Shield className="w-4 h-4 text-cortex-primary" />
                            <h3 className="text-sm font-semibold text-cortex-text-main uppercase tracking-wider">Organization Access</h3>
                        </div>
                        <p className="text-sm leading-6 text-cortex-text-muted">
                            The base release keeps People &amp; Access centered on visible organization roles and collaboration groups. Enterprise can add a separate user-directory layer without forcing it into every deployment.
                        </p>
                    </div>
                    <div className="flex items-center gap-2">
                        <span className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1 text-[10px] font-mono uppercase tracking-[0.2em] text-cortex-text-muted">
                            {accessManagementTier === "enterprise" ? "Enterprise layer" : "Base release layer"}
                        </span>
                        <button
                            onClick={refreshCurrentUser}
                            disabled={syncing}
                            className="p-1.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors disabled:opacity-50"
                            title="Refresh current user"
                        >
                            <RefreshCw className={`w-3.5 h-3.5 ${syncing ? "animate-spin" : ""}`} />
                        </button>
                    </div>
                </div>

                <div className="grid gap-3 lg:grid-cols-3">
                    <AccessRoleCard
                        title="Owner"
                        summary="Can shape organization access, remote posture, and enterprise user management when that layer is enabled."
                        active={currentUserRole === "owner"}
                    />
                    <AccessRoleCard
                        title="Operator"
                        summary="Works inside the organization and collaboration groups without taking over the user directory."
                        active={currentUserRole === "operator"}
                    />
                    <AccessRoleCard
                        title="Viewer"
                        summary="Keeps context readable for review and visibility without mutating organization access or user directory state."
                        active={currentUserRole === "viewer"}
                    />
                </div>

                <div className="rounded-lg border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">
                    <p className="font-medium text-cortex-text-main">Current access posture</p>
                    <p className="mt-1 leading-6">
                        {currentUser
                            ? `${currentUser.name} is currently mapped as ${currentUser.role}. Collaboration groups stay in the organization layer for every deployment.`
                            : "Refresh current user to confirm the active organization role and access layer."}
                    </p>
                    <p className="mt-2 leading-6">
                        {showsEnterpriseDirectory
                            ? "Enterprise access management is enabled for this workspace, so the user directory can stay separate from the base release path."
                            : "Enterprise user-directory management stays intentionally outside the raw release path so the default workflow does not turn into a full IAM console."}
                    </p>
                </div>
            </section>

            {canManageEnterpriseDirectory ? (
                <section className="rounded-xl border border-cortex-border bg-cortex-surface/60 p-4 space-y-4" data-testid="users-management-panel">
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                            <User className="w-4 h-4 text-cortex-primary" />
                            <h3 className="text-sm font-semibold text-cortex-text-main uppercase tracking-wider">Enterprise User Directory</h3>
                            <span className="text-[10px] font-mono text-cortex-text-muted">{activeCount}/{users.length} active</span>
                        </div>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-5 gap-2 text-xs">
                        <label className="flex flex-col gap-1 md:col-span-1">
                            <span className="text-cortex-text-muted font-mono">Name</span>
                            <input value={name} onChange={(e) => setName(e.target.value)} className="px-2 py-1.5 rounded bg-cortex-bg border border-cortex-border text-cortex-text-main" />
                        </label>
                        <label className="flex flex-col gap-1 md:col-span-2">
                            <span className="text-cortex-text-muted font-mono">Email</span>
                            <input value={email} onChange={(e) => setEmail(e.target.value)} className="px-2 py-1.5 rounded bg-cortex-bg border border-cortex-border text-cortex-text-main" />
                        </label>
                        <label className="flex flex-col gap-1 md:col-span-1">
                            <span className="text-cortex-text-muted font-mono">Role</span>
                            <select value={role} onChange={(e) => setRole(toRole(e.target.value))} className="px-2 py-1.5 rounded bg-cortex-bg border border-cortex-border text-cortex-text-main">
                                <option value="viewer">viewer</option>
                                <option value="operator">operator</option>
                                <option value="owner">owner</option>
                            </select>
                        </label>
                        <div className="flex items-end gap-2 md:col-span-1">
                            <label className="flex items-center gap-1 text-[11px] font-mono text-cortex-text-muted">
                                <input
                                    type="checkbox"
                                    checked={remoteAllowed}
                                    onChange={(e) => setRemoteAllowed(e.target.checked)}
                                />
                                Remote
                            </label>
                            <button
                                onClick={createUser}
                                disabled={!name.trim() || !email.trim()}
                                className="ml-auto px-2.5 py-1.5 rounded border border-cortex-primary/30 text-cortex-primary text-[11px] font-mono hover:bg-cortex-primary/10 disabled:opacity-50 flex items-center gap-1"
                                data-testid="users-add-button"
                            >
                                <UserPlus className="w-3 h-3" />
                                Add
                            </button>
                        </div>
                    </div>

                    <div className="rounded-lg border border-cortex-border overflow-hidden">
                        <table className="w-full text-xs font-mono">
                            <thead>
                                <tr className="bg-cortex-surface/50 text-cortex-text-muted">
                                    <th className="text-left px-3 py-2">User</th>
                                    <th className="text-left px-3 py-2">Role</th>
                                    <th className="text-left px-3 py-2">Remote</th>
                                    <th className="text-left px-3 py-2">Status</th>
                                    <th className="text-right px-3 py-2">Actions</th>
                                </tr>
                            </thead>
                            <tbody>
                                {users.map((u) => (
                                    <tr key={u.id} className="border-t border-cortex-border hover:bg-cortex-surface/30 transition-colors">
                                        <td className="px-3 py-2.5">
                                            <div className="flex items-center gap-2">
                                                <div className="w-6 h-6 rounded-full bg-cortex-primary/10 flex items-center justify-center text-cortex-primary text-[10px] font-bold border border-cortex-primary/30">
                                                    {u.name[0]}
                                                </div>
                                                <div className="min-w-0">
                                                    <p className="text-cortex-text-main truncate">{u.name}</p>
                                                    <p className="text-[10px] text-cortex-text-muted truncate">{u.email}</p>
                                                </div>
                                            </div>
                                        </td>
                                        <td className="px-3 py-2.5">
                                            <select
                                                value={u.role}
                                                onChange={(e) => updateUser(u.id, { role: toRole(e.target.value) })}
                                                className="px-1.5 py-1 rounded bg-cortex-bg border border-cortex-border text-cortex-text-main text-[11px]"
                                            >
                                                <option value="viewer">viewer</option>
                                                <option value="operator">operator</option>
                                                <option value="owner">owner</option>
                                            </select>
                                        </td>
                                        <td className="px-3 py-2.5">
                                            <label className="flex items-center gap-1 text-[11px] text-cortex-text-muted">
                                                <input
                                                    type="checkbox"
                                                    checked={u.remoteAllowed}
                                                    onChange={(e) => updateUser(u.id, { remoteAllowed: e.target.checked })}
                                                />
                                                {u.remoteAllowed ? "Allowed" : "Denied"}
                                            </label>
                                        </td>
                                        <td className="px-3 py-2.5">
                                            <button
                                                onClick={() => updateUser(u.id, { status: u.status === "active" ? "disabled" : "active" })}
                                                className={`px-2 py-0.5 rounded border text-[10px] ${
                                                    u.status === "active"
                                                        ? "text-cortex-success border-cortex-success/30 bg-cortex-success/5"
                                                        : "text-cortex-text-muted border-cortex-border bg-cortex-bg"
                                                }`}
                                            >
                                                {u.status.toUpperCase()}
                                            </button>
                                        </td>
                                        <td className="px-3 py-2.5 text-right">
                                            <button
                                                onClick={() => removeUser(u.id)}
                                                className="inline-flex items-center gap-1 px-2 py-0.5 rounded border border-cortex-danger/30 text-cortex-danger text-[10px] hover:bg-cortex-danger/10"
                                                title="Remove user from local list"
                                            >
                                                <Trash2 className="w-3 h-3" />
                                                Remove
                                            </button>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>

                    <div className="p-3 rounded border border-cortex-border bg-cortex-surface/30 text-xs text-cortex-text-muted flex items-center gap-2">
                        <Shield className="w-4 h-4" />
                        <span>Directory persistence/auth integration is pending. This enterprise layer stays separate from the base release path.</span>
                    </div>

                    {notice && <p className="text-[11px] font-mono text-cortex-primary">{notice}</p>}
                </section>
            ) : showsEnterpriseDirectory ? (
                <section className="rounded-xl border border-cortex-border bg-cortex-surface/60 p-4 space-y-3" data-testid="enterprise-directory-readonly">
                    <div className="flex items-center gap-2">
                        <User className="w-4 h-4 text-cortex-primary" />
                        <h3 className="text-sm font-semibold text-cortex-text-main uppercase tracking-wider">Enterprise User Directory</h3>
                    </div>
                    <p className="text-sm leading-6 text-cortex-text-muted">
                        This workspace has enterprise access management enabled, but only organization owners can change the user directory. Your current role can still work through groups and visible organization access without taking over account administration.
                    </p>
                </section>
            ) : (
                <section className="rounded-xl border border-cortex-border bg-cortex-surface/60 p-4 space-y-3" data-testid="enterprise-directory-locked">
                    <div className="flex items-center gap-2">
                        <User className="w-4 h-4 text-cortex-primary" />
                        <h3 className="text-sm font-semibold text-cortex-text-main uppercase tracking-wider">Enterprise User Directory</h3>
                    </div>
                    <p className="text-sm leading-6 text-cortex-text-muted">
                        User-directory management is intentionally kept out of the raw release workflow. The enterprise layer can add user administration without turning the default People &amp; Access surface into a full control plane.
                    </p>
                </section>
            )}

            <section data-testid="users-groups-section">
                <div className="flex items-center gap-2 mb-2 px-1">
                    <Users className="w-4 h-4 text-cortex-primary" />
                    <h3 className="text-sm font-semibold text-cortex-text-main uppercase tracking-wider">Advanced Group Operations</h3>
                </div>
                <div className="rounded-xl border border-cortex-border bg-cortex-surface/60 p-4 space-y-3">
                    <p className="text-sm leading-6 text-cortex-text-muted">
                        Collaboration groups live in Advanced mode so this page can stay centered on identity, access posture, and environment ownership.
                    </p>
                    <Link
                        href="/groups"
                        className="inline-flex items-center rounded-2xl border border-cortex-primary/30 px-4 py-2 text-sm font-semibold text-cortex-primary hover:bg-cortex-primary/10"
                    >
                        Open group operations
                    </Link>
                </div>
            </section>
        </div>
    );
}

function AccessRoleCard({ title, summary, active }: { title: string; summary: string; active: boolean }) {
    return (
        <div className={`rounded-2xl border px-4 py-4 ${active ? "border-cortex-primary/40 bg-cortex-primary/10" : "border-cortex-border bg-cortex-bg"}`}>
            <p className="text-sm font-semibold text-cortex-text-main">{title}</p>
            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{summary}</p>
        </div>
    );
}
