"use client";

import React, { useEffect, useMemo, useState } from "react";
import {
    AccessManagementTier,
    IdentityMode,
    ManagedUser,
    PrincipalType,
    ProductEdition,
    STARTER_DIRECTORY_USERS,
    SharedAgentSpecificityOwner,
    UserRole,
    extractData,
    toAccessManagementTier,
    toIdentityMode,
    toPrincipalType,
    toProductEdition,
    toRole,
    toSharedAgentSpecificityOwner,
} from "./usersTypes";
import UsersAccessPostureSection from "./UsersAccessPostureSection";
import UsersDirectorySection from "./UsersDirectorySection";

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
            <UsersAccessPostureSection
                productEdition={productEdition}
                setProductEdition={setProductEdition}
                identityMode={identityMode}
                setIdentityMode={setIdentityMode}
                sharedAgentSpecificityOwner={sharedAgentSpecificityOwner}
                setSharedAgentSpecificityOwner={setSharedAgentSpecificityOwner}
                principalType={principalType}
                authSource={authSource}
                effectiveRole={effectiveRole}
                breakGlass={breakGlass}
                canManageAccessModel={canManageAccessModel}
                savingAccessModel={savingAccessModel}
                saveAccessModel={saveAccessModel}
                accessManagementTier={accessManagementTier}
                syncing={syncing}
                refreshCurrentUser={refreshCurrentUser}
                currentUser={currentUser}
                currentUserRole={currentUserRole}
                showsEnterpriseDirectory={showsEnterpriseDirectory}
            />
            <UsersDirectorySection
                canManageEnterpriseDirectory={canManageEnterpriseDirectory}
                showsEnterpriseDirectory={showsEnterpriseDirectory}
                activeCount={activeCount}
                users={users}
                name={name}
                setName={setName}
                email={email}
                setEmail={setEmail}
                role={role}
                setRole={setRole}
                remoteAllowed={remoteAllowed}
                setRemoteAllowed={setRemoteAllowed}
                createUser={createUser}
                updateUser={updateUser}
                removeUser={removeUser}
                notice={notice}
            />
        </div>
    );
}
