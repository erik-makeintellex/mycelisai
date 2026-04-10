"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { Download, MessageSquare, RefreshCw, Users } from "lucide-react";
import type { Artifact } from "@/store/cortexStoreTypesPlanning";

type WorkMode = "read_only" | "propose_only" | "execute_with_approval" | "execute_bounded";
type Group = {
    group_id: string;
    name: string;
    goal_statement: string;
    work_mode: WorkMode;
    member_user_ids: string[];
    team_ids: string[];
    coordinator_profile: string;
    approval_policy_ref: string;
    status: "active" | "paused" | "archived";
    expiry?: string | null;
    created_by: string;
    created_at: string;
};
type Monitor = { status?: string; published_count?: number; last_group_id?: string; last_message?: string; last_published_at?: string; last_error?: string };
type ApprovalPrompt = { confirm_token?: { token?: string } };

const inputClassName = "w-full rounded-2xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-main outline-none placeholder:text-cortex-text-muted";

const splitList = (value: string) => value.split(",").map((item) => item.trim()).filter(Boolean);
const getData = async <T,>(res: Response): Promise<T> => {
    const payload = await res.json();
    return (payload && typeof payload === "object" && "data" in payload ? payload.data : payload) as T;
};
const relativeTime = (value?: string | null) => {
    if (!value) return "not yet";
    const diff = Date.now() - new Date(value).getTime();
    if (diff < 60_000) return `${Math.max(1, Math.floor(diff / 1000))}s ago`;
    if (diff < 3_600_000) return `${Math.floor(diff / 60_000)}m ago`;
    if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)}h ago`;
    return `${Math.floor(diff / 86_400_000)}d ago`;
};

export default function GroupManagementPanel({ initialSelectedGroupId = null }: { initialSelectedGroupId?: string | null }) {
    const [groups, setGroups] = useState<Group[]>([]);
    const [monitor, setMonitor] = useState<Monitor | null>(null);
    const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null);
    const [outputs, setOutputs] = useState<Artifact[]>([]);
    const [notice, setNotice] = useState<string | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [approvalPrompt, setApprovalPrompt] = useState<ApprovalPrompt | null>(null);
    const [refreshing, setRefreshing] = useState(false);
    const [saving, setSaving] = useState(false);
    const [broadcasting, setBroadcasting] = useState(false);
    const [archiving, setArchiving] = useState(false);
    const [name, setName] = useState("");
    const [goalStatement, setGoalStatement] = useState("");
    const [workMode, setWorkMode] = useState<WorkMode>("propose_only");
    const [expiry, setExpiry] = useState("");
    const [teamIDs, setTeamIDs] = useState("");
    const [memberIDs, setMemberIDs] = useState("");
    const [coordinatorProfile, setCoordinatorProfile] = useState("");
    const [approvalPolicyRef, setApprovalPolicyRef] = useState("");
    const [allowedCapabilities, setAllowedCapabilities] = useState("");
    const [broadcastMessage, setBroadcastMessage] = useState("");

    const selectedGroup = useMemo(() => groups.find((group) => group.group_id === selectedGroupId) ?? null, [groups, selectedGroupId]);
    const standingGroups = useMemo(() => groups.filter((group) => group.status !== "archived" && !group.expiry), [groups]);
    const temporaryGroups = useMemo(() => groups.filter((group) => group.status !== "archived" && !!group.expiry), [groups]);
    const archivedTemporaryGroups = useMemo(() => groups.filter((group) => group.status === "archived" && !!group.expiry), [groups]);
    const selectedGroupIsArchived = selectedGroup?.status === "archived";
    const outputSummary = useMemo(() => {
        const artifactCount = outputs.length;
        const uniqueAgents = new Set(outputs.map((artifact) => artifact.agent_id).filter(Boolean));
        return {
            artifactCount,
            agentCount: uniqueAgents.size,
        };
    }, [outputs]);

    const loadGroups = async () => {
        setRefreshing(true);
        setError(null);
        try {
            const [groupsRes, monitorRes] = await Promise.all([fetch("/api/v1/groups", { cache: "no-store" }), fetch("/api/v1/groups/monitor", { cache: "no-store" })]);
            if (!groupsRes.ok) throw new Error("Could not load groups.");
            const nextGroups = await getData<Group[]>(groupsRes);
            setGroups(nextGroups);
            setSelectedGroupId((current) => {
                if (current && nextGroups.some((group) => group.group_id === current)) {
                    return current;
                }
                if (initialSelectedGroupId && nextGroups.some((group) => group.group_id === initialSelectedGroupId)) {
                    return initialSelectedGroupId;
                }
                return nextGroups[0]?.group_id ?? null;
            });
            if (monitorRes.ok) setMonitor(await getData<Monitor>(monitorRes));
        } catch (loadError) {
            setError(loadError instanceof Error ? loadError.message : "Could not load groups.");
        } finally {
            setRefreshing(false);
        }
    };

    useEffect(() => { void loadGroups(); }, []);

    useEffect(() => {
        if (!selectedGroup) {
            setOutputs([]);
            return;
        }
        let cancelled = false;
        const loadOutputs = async () => {
            const res = await fetch(`/api/v1/groups/${encodeURIComponent(selectedGroup.group_id)}/outputs?limit=8`, { cache: "no-store" });
            if (cancelled) return;
            if (!res.ok) {
                setOutputs([]);
                return;
            }
            const items = await getData<Artifact[]>(res);
            if (cancelled) return;
            setOutputs(Array.isArray(items) ? items : []);
        };
        void loadOutputs();
        return () => { cancelled = true; };
    }, [selectedGroup]);

    const createGroup = async () => {
        setSaving(true);
        setNotice(null);
        setError(null);
        try {
            const res = await fetch("/api/v1/groups", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({
                name: name.trim(), goal_statement: goalStatement.trim(), work_mode: workMode,
                team_ids: splitList(teamIDs), member_user_ids: splitList(memberIDs), coordinator_profile: coordinatorProfile.trim(),
                approval_policy_ref: approvalPolicyRef.trim(), allowed_capabilities: splitList(allowedCapabilities),
                expiry: expiry ? new Date(expiry).toISOString() : null, confirm_token: approvalPrompt?.confirm_token?.token ?? "",
            }) });
            const payload = await getData<ApprovalPrompt | Group>(res);
            if (res.status === 202) {
                setApprovalPrompt(payload as ApprovalPrompt);
                setNotice("Approval required before the group can be created.");
                return;
            }
            if (!res.ok) throw new Error("Could not create the group.");
            setApprovalPrompt(null);
            setNotice("Group created successfully.");
            setName(""); setGoalStatement(""); setTeamIDs(""); setMemberIDs(""); setCoordinatorProfile(""); setApprovalPolicyRef(""); setAllowedCapabilities(""); setExpiry("");
            await loadGroups();
        } catch (createError) {
            setError(createError instanceof Error ? createError.message : "Could not create the group.");
        } finally {
            setSaving(false);
        }
    };

    const broadcastToGroup = async () => {
        if (!selectedGroup || !broadcastMessage.trim()) return;
        setBroadcasting(true);
        setNotice(null);
        setError(null);
        try {
            const res = await fetch(`/api/v1/groups/${encodeURIComponent(selectedGroup.group_id)}/broadcast`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ message: broadcastMessage.trim() }) });
            if (!res.ok) throw new Error("Could not broadcast to the selected group.");
            setNotice("Broadcast queued for the selected group.");
            setBroadcastMessage("");
            await loadGroups();
        } catch (broadcastError) {
            setError(broadcastError instanceof Error ? broadcastError.message : "Could not broadcast to the selected group.");
        } finally {
            setBroadcasting(false);
        }
    };

    const archiveSelectedGroup = async () => {
        if (!selectedGroup || !selectedGroup.expiry || selectedGroupIsArchived) return;
        setArchiving(true);
        setNotice(null);
        setError(null);
        try {
            const res = await fetch(`/api/v1/groups/${encodeURIComponent(selectedGroup.group_id)}/status`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ status: "archived" }),
            });
            if (!res.ok) throw new Error("Could not archive the selected temporary group.");
            setNotice("Temporary group archived. Retained outputs are still available for review.");
            await loadGroups();
        } catch (archiveError) {
            setError(archiveError instanceof Error ? archiveError.message : "Could not archive the selected temporary group.");
        } finally {
            setArchiving(false);
        }
    };

    return (
        <section className="space-y-5" data-testid="groups-workspace">
            <div className="rounded-3xl border border-cortex-border bg-cortex-surface p-5">
                <div className="flex items-start justify-between gap-4">
                    <div>
                        <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.22em] text-cortex-primary"><Users className="h-3.5 w-3.5" />Groups Workspace</div>
                        <h1 className="mt-3 text-2xl font-semibold text-cortex-text-main">Create, review, and coordinate focused groups.</h1>
                        <p className="mt-2 max-w-3xl text-sm leading-7 text-cortex-text-muted">Keep Soma as the root admin chat. Use groups for focused temporary or standing collaborations, lead workspaces, and output review.</p>
                    </div>
                    <button type="button" onClick={() => void loadGroups()} className="inline-flex items-center gap-2 rounded-2xl border border-cortex-border bg-cortex-bg px-3 py-2 text-sm text-cortex-text-main hover:border-cortex-primary/25"><RefreshCw className={`h-4 w-4 ${refreshing ? "animate-spin" : ""}`} />Refresh</button>
                </div>
                {monitor ? <div className="mt-4 rounded-2xl border border-cortex-border bg-cortex-bg px-4 py-3 text-sm text-cortex-text-muted">Bus {monitor.status || "offline"} • {monitor.published_count ?? 0} published • {monitor.last_group_id ? `last ${monitor.last_group_id}` : "no recent group activity"}{monitor.last_error ? ` • ${monitor.last_error}` : ""}</div> : null}
            </div>

            <div className="grid gap-5 xl:grid-cols-[0.72fr_1.28fr]">
                <div className="space-y-5">
                    <section className="rounded-3xl border border-cortex-border bg-cortex-surface p-5">
                        <h2 className="text-sm font-semibold uppercase tracking-[0.2em] text-cortex-text-main">Add group</h2>
                        <div className="mt-4 grid gap-3">
                            <Field label="Name"><input aria-label="Name" value={name} onChange={(e) => setName(e.target.value)} className={inputClassName} /></Field>
                            <Field label="Goal Statement"><textarea aria-label="Goal Statement" rows={4} value={goalStatement} onChange={(e) => setGoalStatement(e.target.value)} className={`${inputClassName} resize-y`} /></Field>
                            <div className="grid gap-3 md:grid-cols-2">
                                <Field label="Work Mode"><select aria-label="Work Mode" value={workMode} onChange={(e) => setWorkMode(e.target.value as WorkMode)} className={inputClassName}><option value="read_only">read_only</option><option value="propose_only">propose_only</option><option value="execute_with_approval">execute_with_approval</option><option value="execute_bounded">execute_bounded</option></select></Field>
                                <Field label="Expiry"><input aria-label="Expiry" type="datetime-local" value={expiry} onChange={(e) => setExpiry(e.target.value)} className={inputClassName} /></Field>
                            </div>
                            <Field label="Team IDs"><input aria-label="Team IDs" value={teamIDs} onChange={(e) => setTeamIDs(e.target.value)} className={inputClassName} /></Field>
                            <Field label="Member IDs"><input aria-label="Member IDs" value={memberIDs} onChange={(e) => setMemberIDs(e.target.value)} className={inputClassName} /></Field>
                            <Field label="Coordinator Profile"><input aria-label="Coordinator Profile" value={coordinatorProfile} onChange={(e) => setCoordinatorProfile(e.target.value)} className={inputClassName} /></Field>
                            <Field label="Allowed Capabilities"><input aria-label="Allowed Capabilities" value={allowedCapabilities} onChange={(e) => setAllowedCapabilities(e.target.value)} className={inputClassName} /></Field>
                            <Field label="Approval Policy Ref"><input aria-label="Approval Policy Ref" value={approvalPolicyRef} onChange={(e) => setApprovalPolicyRef(e.target.value)} className={inputClassName} /></Field>
                        </div>
                        {approvalPrompt ? <div className="mt-4 rounded-2xl border border-cortex-primary/25 bg-cortex-primary/10 p-4" data-testid="groups-approval-card"><p className="text-sm font-semibold text-cortex-text-main">Approval required before creation</p><input readOnly data-testid="groups-confirm-token-input" value={approvalPrompt.confirm_token?.token ?? ""} className={`${inputClassName} mt-3 font-mono`} /></div> : null}
                        {notice ? <p className="mt-4 text-sm text-cortex-primary" data-testid="groups-notice">{notice}</p> : null}
                        {error ? <p className="mt-4 text-sm text-cortex-danger" data-testid="groups-error">{error}</p> : null}
                        <button type="button" onClick={() => void createGroup()} disabled={saving} data-testid="groups-create-button" className="mt-4 rounded-2xl bg-cortex-primary px-4 py-2 text-sm font-semibold text-cortex-bg disabled:opacity-60">{saving ? "Saving..." : approvalPrompt ? "Confirm and create group" : "Create group"}</button>
                    </section>

                    <GroupList title="Standing groups" groups={standingGroups} selectedGroupId={selectedGroupId} onSelect={setSelectedGroupId} />
                    <GroupList title="Temporary groups" groups={temporaryGroups} selectedGroupId={selectedGroupId} onSelect={setSelectedGroupId} />
                    <GroupList title="Archived temporary groups" groups={archivedTemporaryGroups} selectedGroupId={selectedGroupId} onSelect={setSelectedGroupId} />
                </div>

                <section className="rounded-3xl border border-cortex-border bg-cortex-surface p-5">
                    {selectedGroup ? (
                        <div className="space-y-5">
                            <div className="border-b border-cortex-border pb-4">
                                <div className="flex flex-wrap gap-2">
                                    <span className="rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-primary">{selectedGroup.status === "archived" && selectedGroup.expiry ? "Archived temporary group" : selectedGroup.expiry ? "Temporary group" : "Standing group"}</span>
                                    <span className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1 text-[11px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">{selectedGroup.work_mode}</span>
                                    {selectedGroup.expiry && !selectedGroupIsArchived ? (
                                        <button
                                            type="button"
                                            onClick={() => void archiveSelectedGroup()}
                                            disabled={archiving}
                                            className="rounded-full border border-cortex-border bg-cortex-bg px-3 py-1 text-[11px] font-mono uppercase tracking-[0.16em] text-cortex-text-main disabled:opacity-60"
                                        >
                                            {archiving ? "Archiving..." : "Archive temporary group"}
                                        </button>
                                    ) : null}
                                </div>
                                <h2 className="mt-3 text-2xl font-semibold text-cortex-text-main">{selectedGroup.name}</h2>
                                <p className="mt-2 text-sm leading-7 text-cortex-text-muted">{selectedGroup.goal_statement}</p>
                            </div>
                            <div className="grid gap-4 lg:grid-cols-[0.7fr_1.3fr]">
                                <div className="space-y-4">
                                    <Info label="Focused team lead" value={selectedGroup.coordinator_profile || `${selectedGroup.name} lead`} detail="Enter the narrower group lane through this lead counterpart." />
                                    <Info label="Created by" value={selectedGroup.created_by} detail={`Created ${relativeTime(selectedGroup.created_at)}`} />
                                    <div className="rounded-2xl border border-cortex-border bg-cortex-bg p-4">
                                        <p className="text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-primary">Jump into work</p>
                                        <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                                            Open Soma for root coordination or drop into one of the connected lead workspaces for this group.
                                        </p>
                                        <div className="mt-3 flex flex-col gap-2">
                                            <Link href="/dashboard" className={linkClassName}>Open Soma admin home</Link>
                                            {selectedGroup.team_ids.map((teamId) => <Link key={teamId} href={`/dashboard?team_id=${encodeURIComponent(teamId)}`} className={linkClassName}>Open {teamId} lead</Link>)}
                                        </div>
                                    </div>
                                </div>
                                <div className="space-y-4">
                                    <div className="rounded-2xl border border-cortex-border bg-cortex-bg p-4">
                                        <div className="flex items-center gap-2"><MessageSquare className="h-4 w-4 text-cortex-primary" /><p className="text-sm font-semibold text-cortex-text-main">{selectedGroupIsArchived ? "Group is archived" : "Broadcast a focused ask"}</p></div>
                                        {selectedGroupIsArchived ? (
                                            <p className="mt-3 text-sm leading-6 text-cortex-text-muted" data-testid="groups-archived-readonly-note">
                                                This temporary group is archived. Keep it available for retained output review, but send any new coordination through an active group or directly through Soma.
                                            </p>
                                        ) : (
                                            <>
                                                <textarea value={broadcastMessage} onChange={(e) => setBroadcastMessage(e.target.value)} rows={4} className={`${inputClassName} mt-3 resize-y`} />
                                                <button type="button" onClick={() => void broadcastToGroup()} disabled={broadcasting || !broadcastMessage.trim()} className="mt-3 rounded-2xl border border-cortex-primary/30 px-4 py-2 text-sm font-semibold text-cortex-primary disabled:opacity-60">{broadcasting ? "Broadcasting..." : "Broadcast to group"}</button>
                                            </>
                                        )}
                                    </div>
                                    <div className="rounded-2xl border border-cortex-border bg-cortex-bg p-4">
                                        <p className="text-sm font-semibold text-cortex-text-main">{selectedGroupIsArchived ? "Retained outputs" : "Recent outputs"}</p>
                                        <div className="mt-2 flex flex-wrap gap-2" data-testid="groups-output-summary">
                                            <span className="rounded-full border border-cortex-border bg-cortex-surface px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
                                                {outputSummary.artifactCount} output{outputSummary.artifactCount === 1 ? "" : "s"}
                                            </span>
                                            <span className="rounded-full border border-cortex-border bg-cortex-surface px-3 py-1 text-[11px] font-mono text-cortex-text-muted">
                                                {outputSummary.agentCount} contributing lead{outputSummary.agentCount === 1 ? "" : "s"}
                                            </span>
                                        </div>
                                        {selectedGroupIsArchived ? (
                                            <p className="mt-2 text-sm leading-6 text-cortex-text-muted" data-testid="groups-retained-outputs-note">
                                                Review the outputs this archived temporary group already produced. Downloads remain available so the work can still be inspected after the collaboration window closes.
                                            </p>
                                        ) : null}
                                        <div className="mt-3 space-y-3">
                                            {outputs.length === 0 ? <p className="text-sm text-cortex-text-muted">No recent team outputs found for this group yet.</p> : outputs.slice(0, 8).map((artifact) => <ArtifactRow key={artifact.id} artifact={artifact} />)}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    ) : <p className="text-sm text-cortex-text-muted">Create a group or select one to review its lead lane and outputs.</p>}
                </section>
            </div>
        </section>
    );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
    return <label className="flex flex-col gap-1 text-sm"><span className="font-semibold text-cortex-text-main">{label}</span>{children}</label>;
}

function GroupList({ title, groups, selectedGroupId, onSelect }: { title: string; groups: Group[]; selectedGroupId: string | null; onSelect: (groupId: string) => void }) {
    return (
        <section className="rounded-3xl border border-cortex-border bg-cortex-surface p-5">
            <div className="flex items-center justify-between gap-3"><h2 className="text-sm font-semibold uppercase tracking-[0.2em] text-cortex-text-main">{title}</h2><span className="text-[11px] font-mono text-cortex-text-muted">{groups.length}</span></div>
            <div className="mt-4 space-y-3">{groups.length === 0 ? <p className="text-sm text-cortex-text-muted">Nothing here yet.</p> : groups.map((group) => <button key={group.group_id} type="button" onClick={() => onSelect(group.group_id)} className={`w-full rounded-2xl border px-4 py-3 text-left ${selectedGroupId === group.group_id ? "border-cortex-primary/30 bg-cortex-primary/10" : "border-cortex-border bg-cortex-bg hover:border-cortex-primary/25"}`}><div className="flex flex-wrap items-center gap-2"><p className="text-sm font-semibold text-cortex-text-main">{group.name}</p>{group.status === "archived" ? <span className="rounded-full border border-cortex-border bg-cortex-surface px-2 py-0.5 text-[10px] font-mono uppercase tracking-[0.14em] text-cortex-text-muted">Archived</span> : null}</div><p className="mt-1 text-sm leading-6 text-cortex-text-muted">{group.goal_statement}</p></button>)}</div>
        </section>
    );
}

function Info({ label, value, detail }: { label: string; value: string; detail: string }) {
    return <div className="rounded-2xl border border-cortex-border bg-cortex-bg p-4"><p className="text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-primary">{label}</p><p className="mt-2 text-sm font-semibold text-cortex-text-main">{value}</p><p className="mt-1 text-sm leading-6 text-cortex-text-muted">{detail}</p></div>;
}

function ArtifactRow({ artifact }: { artifact: Artifact }) {
    const readable = artifact.artifact_type === "code" || artifact.artifact_type === "document" || artifact.artifact_type === "data" || artifact.artifact_type === "chart";
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3">
            <div className="flex items-start justify-between gap-3">
                <div><p className="text-sm font-semibold text-cortex-text-main">{artifact.title}</p><p className="mt-1 text-[11px] font-mono uppercase tracking-[0.14em] text-cortex-text-muted">{artifact.artifact_type} • {artifact.agent_id} • {relativeTime(artifact.created_at)}</p></div>
                <a href={`/api/v1/artifacts/${encodeURIComponent(artifact.id)}/download`} download className="inline-flex items-center gap-1 rounded-xl border border-cortex-primary/30 px-3 py-2 text-xs font-semibold text-cortex-primary hover:bg-cortex-primary/10"><Download className="h-3.5 w-3.5" />Download</a>
            </div>
            {readable && artifact.content ? <pre className="mt-3 overflow-x-auto rounded-2xl border border-cortex-border bg-cortex-bg p-3 text-xs leading-6 text-cortex-text-muted">{artifact.content}</pre> : artifact.file_path ? <p className="mt-3 text-sm text-cortex-text-muted">Saved path: {artifact.file_path}</p> : null}
        </div>
    );
}

const linkClassName = "rounded-2xl border border-cortex-border bg-cortex-surface px-4 py-3 text-sm text-cortex-text-main hover:border-cortex-primary/25";
