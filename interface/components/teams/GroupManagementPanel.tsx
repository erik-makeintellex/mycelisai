"use client";

import React, { useEffect, useMemo, useState } from "react";
import { Megaphone, RefreshCw, Users } from "lucide-react";
import { extractApiData } from "@/lib/apiContracts";

type Group = {
    group_id: string;
    name: string;
    goal_statement: string;
    work_mode: string;
    status: string;
    team_ids: string[];
    member_user_ids: string[];
    approval_policy_ref?: string;
};

type GroupBusSnapshot = {
    status?: string;
    published_count?: number;
    last_group_id?: string;
    last_published_at?: string;
    last_error?: string;
};

type ApprovalToken = {
    token?: string;
    intent_proof_id?: string;
};

type ApprovalResponse = {
    requires_approval?: boolean;
    confirm_token?: ApprovalToken | string;
    intent_proof?: { id?: string };
};

const WORK_MODES = [
    "read_only",
    "propose_only",
    "execute_with_approval",
    "execute_bounded",
] as const;

function parseConfirmToken(confirmToken: ApprovalResponse["confirm_token"]): string {
    if (!confirmToken) return "";
    if (typeof confirmToken === "string") return confirmToken;
    return typeof confirmToken.token === "string" ? confirmToken.token : "";
}

export default function GroupManagementPanel() {
    const [groups, setGroups] = useState<Group[]>([]);
    const [monitor, setMonitor] = useState<GroupBusSnapshot | null>(null);
    const [loading, setLoading] = useState(false);
    const [creating, setCreating] = useState(false);
    const [broadcasting, setBroadcasting] = useState(false);
    const [selectedGroupID, setSelectedGroupID] = useState("");
    const [broadcastMessage, setBroadcastMessage] = useState("");
    const [error, setError] = useState<string | null>(null);
    const [notice, setNotice] = useState<string | null>(null);
    const [pendingConfirmToken, setPendingConfirmToken] = useState("");
    const [pendingIntentProofID, setPendingIntentProofID] = useState("");

    const [name, setName] = useState("");
    const [goal, setGoal] = useState("");
    const [workMode, setWorkMode] = useState<(typeof WORK_MODES)[number]>("propose_only");
    const [teamsCSV, setTeamsCSV] = useState("admin-core,council-core");
    const [membersCSV, setMembersCSV] = useState("");

    const sortedGroups = useMemo(
        () => [...groups].sort((a, b) => a.name.localeCompare(b.name)),
        [groups],
    );

    const fetchGroups = async () => {
        setLoading(true);
        setError(null);
        try {
            const [groupsRes, monitorRes] = await Promise.all([
                fetch("/api/v1/groups", { cache: "no-store" }),
                fetch("/api/v1/groups/monitor", { cache: "no-store" }),
            ]);
            if (!groupsRes.ok) throw new Error("failed to fetch groups");
            const groupsPayload = await groupsRes.json();
            const monitorPayload = monitorRes.ok ? await monitorRes.json() : null;
            setGroups(extractApiData<Group[] | unknown>(groupsPayload) as Group[]);
            setMonitor(monitorPayload ? (extractApiData<GroupBusSnapshot>(monitorPayload) as GroupBusSnapshot) : null);
        } catch (err) {
            setError(err instanceof Error ? err.message : "failed to load groups");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchGroups();
        const interval = setInterval(fetchGroups, 10000);
        return () => clearInterval(interval);
    }, []);

    const createGroup = async () => {
        if (!name.trim() || !goal.trim()) return;
        setCreating(true);
        setError(null);
        setNotice(null);
        try {
            const body = {
                name: name.trim(),
                goal_statement: goal.trim(),
                work_mode: workMode,
                team_ids: teamsCSV.split(",").map((v) => v.trim()).filter(Boolean),
                member_user_ids: membersCSV.split(",").map((v) => v.trim()).filter(Boolean),
                allowed_capabilities: ["runs.read", "runs.propose"],
                approval_policy_ref: "policy.default",
                ...(pendingConfirmToken ? { confirm_token: pendingConfirmToken } : {}),
            };
            const res = await fetch("/api/v1/groups", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(body),
            });
            const payload = await res.json().catch(() => ({}));
            if (!res.ok) {
                throw new Error((payload as { error?: string }).error ?? "failed to create group");
            }

            const data = extractApiData<ApprovalResponse | unknown>(payload) as ApprovalResponse;
            if (data?.requires_approval) {
                const token = parseConfirmToken(data.confirm_token);
                const proofID = data.intent_proof?.id ?? "";
                setPendingConfirmToken(token);
                setPendingIntentProofID(proofID);
                setNotice("Approval required. Confirm the operation token, then submit create again.");
                return;
            }

            setName("");
            setGoal("");
            setMembersCSV("");
            setPendingConfirmToken("");
            setPendingIntentProofID("");
            setNotice("Group created successfully.");
            await fetchGroups();
        } catch (err) {
            setError(err instanceof Error ? err.message : "failed to create group");
        } finally {
            setCreating(false);
        }
    };

    const sendBroadcast = async () => {
        if (!selectedGroupID || !broadcastMessage.trim()) return;
        setBroadcasting(true);
        setError(null);
        setNotice(null);
        try {
            const res = await fetch(`/api/v1/groups/${selectedGroupID}/broadcast`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ message: broadcastMessage.trim() }),
            });
            if (!res.ok) {
                const payload = await res.json().catch(() => ({}));
                throw new Error((payload as { error?: string }).error ?? "broadcast failed");
            }
            setBroadcastMessage("");
            setNotice("Broadcast queued.");
            await fetchGroups();
        } catch (err) {
            setError(err instanceof Error ? err.message : "broadcast failed");
        } finally {
            setBroadcasting(false);
        }
    };

    return (
        <section className="rounded-xl border border-cortex-border bg-cortex-surface/60 p-4 space-y-4">
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                    <Users className="w-4 h-4 text-cortex-primary" />
                    <h2 className="text-xs font-mono uppercase tracking-wider text-cortex-text-main">Collaboration Groups</h2>
                </div>
                <button
                    onClick={fetchGroups}
                    disabled={loading}
                    className="p-1.5 rounded hover:bg-cortex-border text-cortex-text-muted hover:text-cortex-text-main transition-colors disabled:opacity-50"
                    title="Refresh groups"
                >
                    <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin" : ""}`} />
                </button>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-3 text-xs">
                <label className="flex flex-col gap-1">
                    <span className="text-cortex-text-muted font-mono">Name</span>
                    <input value={name} onChange={(e) => setName(e.target.value)} className="px-2 py-1.5 rounded bg-cortex-bg border border-cortex-border text-cortex-text-main" />
                </label>
                <label className="flex flex-col gap-1">
                    <span className="text-cortex-text-muted font-mono">Work Mode</span>
                    <select value={workMode} onChange={(e) => setWorkMode(e.target.value as (typeof WORK_MODES)[number])} className="px-2 py-1.5 rounded bg-cortex-bg border border-cortex-border text-cortex-text-main">
                        {WORK_MODES.map((mode) => (
                            <option key={mode} value={mode}>{mode}</option>
                        ))}
                    </select>
                </label>
                <label className="flex flex-col gap-1 lg:col-span-2">
                    <span className="text-cortex-text-muted font-mono">Goal Statement</span>
                    <input value={goal} onChange={(e) => setGoal(e.target.value)} className="px-2 py-1.5 rounded bg-cortex-bg border border-cortex-border text-cortex-text-main" />
                </label>
                <label className="flex flex-col gap-1">
                    <span className="text-cortex-text-muted font-mono">Team IDs (comma-separated)</span>
                    <input value={teamsCSV} onChange={(e) => setTeamsCSV(e.target.value)} className="px-2 py-1.5 rounded bg-cortex-bg border border-cortex-border text-cortex-text-main" />
                </label>
                <label className="flex flex-col gap-1">
                    <span className="text-cortex-text-muted font-mono">Member IDs (comma-separated)</span>
                    <input value={membersCSV} onChange={(e) => setMembersCSV(e.target.value)} className="px-2 py-1.5 rounded bg-cortex-bg border border-cortex-border text-cortex-text-main" />
                </label>
            </div>

            <div className="flex items-center gap-2">
                <button
                    onClick={createGroup}
                    disabled={creating || !name.trim() || !goal.trim()}
                    className="px-3 py-1.5 rounded border border-cortex-primary/30 text-cortex-primary text-[11px] font-mono hover:bg-cortex-primary/10 disabled:opacity-50"
                    data-testid="groups-create-button"
                >
                    {creating ? "Creating..." : pendingConfirmToken ? "Create With Approval Token" : "Create Group"}
                </button>
            </div>

            {pendingConfirmToken && (
                <div className="rounded-lg border border-cortex-warning/40 bg-cortex-warning/10 p-3 space-y-2" data-testid="groups-approval-card">
                    <p className="text-[11px] font-mono text-cortex-warning">Approval required for high-impact mutation.</p>
                    <label className="flex flex-col gap-1">
                        <span className="text-[10px] font-mono text-cortex-text-muted uppercase">Confirm Token</span>
                        <input
                            value={pendingConfirmToken}
                            onChange={(e) => setPendingConfirmToken(e.target.value)}
                            className="px-2 py-1.5 rounded bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs font-mono"
                            data-testid="groups-confirm-token-input"
                        />
                    </label>
                    {pendingIntentProofID && (
                        <p className="text-[10px] font-mono text-cortex-text-muted">
                            Intent proof: {pendingIntentProofID}
                        </p>
                    )}
                </div>
            )}

            <div className="rounded-lg border border-cortex-border overflow-hidden">
                <div className="grid grid-cols-[1.4fr_1fr_0.8fr_0.8fr] gap-2 px-3 py-2 text-[10px] font-mono uppercase tracking-wider text-cortex-text-muted bg-cortex-bg/70">
                    <span>Name</span>
                    <span>Work Mode</span>
                    <span>Status</span>
                    <span>Teams</span>
                </div>
                {sortedGroups.length === 0 ? (
                    <div className="px-3 py-3 text-[11px] text-cortex-text-muted">No groups configured.</div>
                ) : (
                    sortedGroups.map((group) => (
                        <button
                            key={group.group_id}
                            onClick={() => setSelectedGroupID(group.group_id)}
                            className={`w-full grid grid-cols-[1.4fr_1fr_0.8fr_0.8fr] gap-2 px-3 py-2 text-[11px] text-left border-t border-cortex-border hover:bg-cortex-bg/40 ${
                                selectedGroupID === group.group_id ? "bg-cortex-primary/10" : ""
                            }`}
                        >
                            <span className="text-cortex-text-main truncate">{group.name}</span>
                            <span className="text-cortex-text-muted font-mono truncate">{group.work_mode}</span>
                            <span className="text-cortex-text-muted font-mono">{group.status}</span>
                            <span className="text-cortex-text-muted font-mono">{group.team_ids?.length ?? 0}</span>
                        </button>
                    ))
                )}
            </div>

            <div className="rounded-lg border border-cortex-border p-3 space-y-2">
                <div className="flex items-center gap-2 text-[11px] font-mono text-cortex-text-main">
                    <Megaphone className="w-3.5 h-3.5 text-cortex-warning" />
                    Group Broadcast
                </div>
                <div className="flex flex-col sm:flex-row gap-2">
                    <select
                        value={selectedGroupID}
                        onChange={(e) => setSelectedGroupID(e.target.value)}
                        className="px-2 py-1.5 rounded bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs"
                    >
                        <option value="">Select group…</option>
                        {sortedGroups.map((g) => (
                            <option key={g.group_id} value={g.group_id}>{g.name}</option>
                        ))}
                    </select>
                    <input
                        value={broadcastMessage}
                        onChange={(e) => setBroadcastMessage(e.target.value)}
                        placeholder="Broadcast message"
                        className="flex-1 px-2 py-1.5 rounded bg-cortex-bg border border-cortex-border text-cortex-text-main text-xs"
                    />
                    <button
                        onClick={sendBroadcast}
                        disabled={broadcasting || !selectedGroupID || !broadcastMessage.trim()}
                        className="px-3 py-1.5 rounded border border-cortex-warning/30 text-cortex-warning text-[11px] font-mono hover:bg-cortex-warning/10 disabled:opacity-50"
                    >
                        {broadcasting ? "Sending..." : "Send"}
                    </button>
                </div>
            </div>

            <div className="text-[11px] text-cortex-text-muted font-mono">
                Bus: {monitor?.status ?? "unknown"} | Published: {monitor?.published_count ?? 0}
                {monitor?.last_group_id ? ` | Last group: ${monitor.last_group_id}` : ""}
                {monitor?.last_error ? ` | Last error: ${monitor.last_error}` : ""}
            </div>

            {notice && <div className="text-[11px] text-cortex-primary font-mono" data-testid="groups-notice">{notice}</div>}
            {error && <div className="text-[11px] text-cortex-danger font-mono" data-testid="groups-error">{error}</div>}
        </section>
    );
}
