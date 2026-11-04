"use client";

import React from "react";
import { Shield, RefreshCw } from "lucide-react";
import {
    AccessManagementTier,
    EDITION_COPY,
    IDENTITY_COPY,
    IdentityMode,
    ManagedUser,
    PrincipalType,
    ProductEdition,
    SharedAgentSpecificityOwner,
    UserRole,
    toIdentityMode,
    toSharedAgentSpecificityOwner,
} from "./usersTypes";

interface UsersAccessPostureSectionProps {
    productEdition: ProductEdition;
    setProductEdition: (edition: ProductEdition) => void;
    identityMode: IdentityMode;
    setIdentityMode: (mode: IdentityMode) => void;
    sharedAgentSpecificityOwner: SharedAgentSpecificityOwner;
    setSharedAgentSpecificityOwner: (owner: SharedAgentSpecificityOwner) => void;
    principalType: PrincipalType;
    authSource: string;
    effectiveRole: UserRole;
    breakGlass: boolean;
    canManageAccessModel: boolean;
    savingAccessModel: boolean;
    saveAccessModel: () => void;
    accessManagementTier: AccessManagementTier;
    syncing: boolean;
    refreshCurrentUser: () => void;
    currentUser: ManagedUser | null;
    currentUserRole: UserRole;
    showsEnterpriseDirectory: boolean;
}

function AccessRoleCard({ title, summary, active }: { title: string; summary: string; active: boolean }) {
    return (
        <div className={`rounded-2xl border px-4 py-4 ${active ? "border-cortex-primary/40 bg-cortex-primary/10" : "border-cortex-border bg-cortex-bg"}`}>
            <p className="text-sm font-semibold text-cortex-text-main">{title}</p>
            <p className="mt-2 text-sm leading-6 text-cortex-text-muted">{summary}</p>
        </div>
    );
}

export default function UsersAccessPostureSection({
    productEdition,
    setProductEdition,
    identityMode,
    setIdentityMode,
    sharedAgentSpecificityOwner,
    setSharedAgentSpecificityOwner,
    principalType,
    authSource,
    effectiveRole,
    breakGlass,
    canManageAccessModel,
    savingAccessModel,
    saveAccessModel,
    accessManagementTier,
    syncing,
    refreshCurrentUser,
    currentUser,
    currentUserRole,
    showsEnterpriseDirectory,
}: UsersAccessPostureSectionProps) {
    return (
        <>
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
        </>
    );
}
