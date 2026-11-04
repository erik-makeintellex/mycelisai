"use client";

import React from "react";
import Link from "next/link";
import { Shield, User, UserPlus, Trash2, Users } from "lucide-react";
import { ManagedUser, UserRole, toRole } from "./usersTypes";

interface UsersDirectorySectionProps {
    canManageEnterpriseDirectory: boolean;
    showsEnterpriseDirectory: boolean;
    activeCount: number;
    users: ManagedUser[];
    name: string;
    setName: (name: string) => void;
    email: string;
    setEmail: (email: string) => void;
    role: UserRole;
    setRole: (role: UserRole) => void;
    remoteAllowed: boolean;
    setRemoteAllowed: (allowed: boolean) => void;
    createUser: () => void;
    updateUser: (id: string, patch: Partial<ManagedUser>) => void;
    removeUser: (id: string) => void;
    notice: string | null;
}

export default function UsersDirectorySection({
    canManageEnterpriseDirectory,
    showsEnterpriseDirectory,
    activeCount,
    users,
    name,
    setName,
    email,
    setEmail,
    role,
    setRole,
    remoteAllowed,
    setRemoteAllowed,
    createUser,
    updateUser,
    removeUser,
    notice,
}: UsersDirectorySectionProps) {
    return (
        <>
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
        </>
    );
}
