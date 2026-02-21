"use client";

import React from "react";
import { Shield, User } from "lucide-react";

const STUB_USERS = [
    { name: "Admin", role: "owner", remoteAllowed: true },
    { name: "Operator", role: "operator", remoteAllowed: false },
    { name: "Viewer", role: "viewer", remoteAllowed: false },
];

export default function UsersPage() {
    return (
        <div className="space-y-4">
            <h3 className="text-sm font-semibold text-cortex-text-muted uppercase tracking-wider">User Management</h3>

            <div className="rounded-lg border border-cortex-border overflow-hidden">
                <table className="w-full text-xs font-mono">
                    <thead>
                        <tr className="bg-cortex-surface/50 text-cortex-text-muted">
                            <th className="text-left px-4 py-2">User</th>
                            <th className="text-left px-4 py-2">Role</th>
                            <th className="text-left px-4 py-2">Remote Providers</th>
                            <th className="text-left px-4 py-2">Status</th>
                        </tr>
                    </thead>
                    <tbody>
                        {STUB_USERS.map((u) => (
                            <tr key={u.name} className="border-t border-cortex-border hover:bg-cortex-surface/30 transition-colors">
                                <td className="px-4 py-2.5">
                                    <div className="flex items-center gap-2">
                                        <div className="w-6 h-6 rounded-full bg-cortex-primary/10 flex items-center justify-center text-cortex-primary text-[10px] font-bold border border-cortex-primary/30">
                                            {u.name[0]}
                                        </div>
                                        <span className="text-cortex-text-main">{u.name}</span>
                                    </div>
                                </td>
                                <td className="px-4 py-2.5">
                                    <span className={`px-2 py-0.5 rounded border text-[10px] ${
                                        u.role === "owner"
                                            ? "text-cortex-primary border-cortex-primary/30 bg-cortex-primary/5"
                                            : u.role === "operator"
                                                ? "text-cortex-success border-cortex-success/30 bg-cortex-success/5"
                                                : "text-cortex-text-muted border-cortex-border bg-cortex-bg"
                                    }`}>
                                        {u.role.toUpperCase()}
                                    </span>
                                </td>
                                <td className="px-4 py-2.5">
                                    {u.remoteAllowed ? (
                                        <span className="text-cortex-success text-[10px]">Allowed</span>
                                    ) : (
                                        <span className="text-cortex-text-muted text-[10px]">Denied</span>
                                    )}
                                </td>
                                <td className="px-4 py-2.5">
                                    <div className="flex items-center gap-1.5">
                                        <div className="w-1.5 h-1.5 rounded-full bg-cortex-success" />
                                        <span className="text-cortex-text-muted">Active</span>
                                    </div>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            <div className="p-3 rounded border border-cortex-border bg-cortex-surface/30 text-xs text-cortex-text-muted flex items-center gap-2">
                <Shield className="w-4 h-4" />
                <span>Authentication is not yet implemented. User roles shown above are placeholders.</span>
            </div>
        </div>
    );
}
