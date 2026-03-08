"use client";

import React, { useEffect, useMemo, useState } from "react";
import { Shield, User, UserPlus, Trash2, RefreshCw, Users } from "lucide-react";
import GroupManagementPanel from "@/components/teams/GroupManagementPanel";

type UserRole = "owner" | "operator" | "viewer";

type ManagedUser = {
    id: string;
    name: string;
    email: string;
    role: UserRole;
    remoteAllowed: boolean;
    status: "active" | "disabled";
};

const STARTER_USERS: ManagedUser[] = [
    { id: "owner", name: "Admin", email: "admin@local", role: "owner", remoteAllowed: true, status: "active" },
    { id: "operator", name: "Operator", email: "operator@local", role: "operator", remoteAllowed: false, status: "active" },
    { id: "viewer", name: "Viewer", email: "viewer@local", role: "viewer", remoteAllowed: false, status: "active" },
];

function toRole(value: string): UserRole {
    if (value === "owner" || value === "operator" || value === "viewer") return value;
    return "viewer";
}

function extractData<T>(payload: unknown): T {
    if (payload && typeof payload === "object" && "data" in payload) {
        return (payload as { data: T }).data;
    }
    return payload as T;
}

export default function UsersPage() {
    const [users, setUsers] = useState<ManagedUser[]>(STARTER_USERS);
    const [name, setName] = useState("");
    const [email, setEmail] = useState("");
    const [role, setRole] = useState<UserRole>("viewer");
    const [remoteAllowed, setRemoteAllowed] = useState(false);
    const [syncing, setSyncing] = useState(false);
    const [notice, setNotice] = useState<string | null>(null);

    const activeCount = useMemo(() => users.filter((u) => u.status === "active").length, [users]);

    const refreshCurrentUser = async () => {
        setSyncing(true);
        try {
            const res = await fetch("/api/v1/user/me", { cache: "no-store" });
            if (!res.ok) return;
            const payload = await res.json();
            const me = extractData<{ id?: string; email?: string; role?: string; name?: string }>(payload);
            const meID = me?.id || "me";
            const meName = me?.name || me?.email || "Current User";
            const meEmail = me?.email || "me@local";
            const meRole = toRole(me?.role || "owner");
            setUsers((prev) => {
                const existing = prev.find((u) => u.id === meID);
                if (existing) {
                    return prev.map((u) =>
                        u.id === meID ? { ...u, name: meName, email: meEmail, role: meRole, status: "active" } : u,
                    );
                }
                return [{ id: meID, name: meName, email: meEmail, role: meRole, remoteAllowed: true, status: "active" }, ...prev];
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

    return (
        <div className="space-y-6">
            <section className="rounded-xl border border-cortex-border bg-cortex-surface/60 p-4 space-y-4" data-testid="users-management-panel">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                        <User className="w-4 h-4 text-cortex-primary" />
                        <h3 className="text-sm font-semibold text-cortex-text-main uppercase tracking-wider">User Management</h3>
                        <span className="text-[10px] font-mono text-cortex-text-muted">{activeCount}/{users.length} active</span>
                    </div>
                    <button
                        onClick={refreshCurrentUser}
                        disabled={syncing}
                        className="p-1.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors disabled:opacity-50"
                        title="Refresh current user"
                    >
                        <RefreshCw className={`w-3.5 h-3.5 ${syncing ? "animate-spin" : ""}`} />
                    </button>
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
                    <span>Directory persistence/auth integration is pending. This surface is ready for backend user CRUD hookup.</span>
                </div>

                {notice && <p className="text-[11px] font-mono text-cortex-primary">{notice}</p>}
            </section>

            <section data-testid="users-groups-section">
                <div className="flex items-center gap-2 mb-2 px-1">
                    <Users className="w-4 h-4 text-cortex-primary" />
                    <h3 className="text-sm font-semibold text-cortex-text-main uppercase tracking-wider">User Group Management</h3>
                </div>
                <GroupManagementPanel />
            </section>
        </div>
    );
}
