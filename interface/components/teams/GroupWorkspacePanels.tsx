import Link from "next/link";
import { Download, MessageSquare, RefreshCw, SlidersHorizontal, Users } from "lucide-react";
import type { Artifact } from "@/store/cortexStoreTypesPlanning";
import { CreateGroupPane } from "./CreateGroupPane";
import {
    compactButtonClassName,
    groupKindLabel,
    inputClassName,
    linkClassName,
    relativeTime,
    type ApprovalPrompt,
    type Group,
    type GroupBucket,
    type GroupDraft,
    type Monitor,
    type OutputSummary,
} from "./groupWorkspaceTypes";

type WorkspaceProps = {
    buckets: GroupBucket[];
    monitor: Monitor | null;
    selectedGroup: Group | null;
    selectedGroupId: string | null;
    outputs: Artifact[];
    outputSummary: OutputSummary;
    draft: GroupDraft;
    notice: string | null;
    error: string | null;
    approvalPrompt: ApprovalPrompt | null;
    refreshing: boolean;
    saving: boolean;
    broadcasting: boolean;
    archiving: boolean;
    broadcastMessage: string;
    onRefresh: () => void;
    onSelectGroup: (groupId: string) => void;
    onDraftChange: (patch: Partial<GroupDraft>) => void;
    onCreateGroup: () => void;
    onBroadcastMessageChange: (message: string) => void;
    onBroadcast: () => void;
    onArchive: () => void;
};

export function GroupWorkspacePanels({
    buckets,
    monitor,
    selectedGroup,
    selectedGroupId,
    outputs,
    outputSummary,
    draft,
    notice,
    error,
    approvalPrompt,
    refreshing,
    saving,
    broadcasting,
    archiving,
    broadcastMessage,
    onRefresh,
    onSelectGroup,
    onDraftChange,
    onCreateGroup,
    onBroadcastMessageChange,
    onBroadcast,
    onArchive,
}: WorkspaceProps) {
    return (
        <section className="flex h-[calc(100vh-2rem)] min-h-[640px] flex-col gap-4 overflow-hidden" data-testid="groups-workspace">
            <GroupsHeader monitor={monitor} refreshing={refreshing} onRefresh={onRefresh} />
            <div className="grid min-h-0 flex-1 gap-4 xl:grid-cols-[300px_minmax(0,1fr)]">
                <GroupRail buckets={buckets} selectedGroupId={selectedGroupId} onSelectGroup={onSelectGroup} />
                <div className="grid min-h-0 min-w-0 gap-4 2xl:grid-cols-[minmax(0,1fr)_360px]">
                    <GroupDetailPane
                        selectedGroup={selectedGroup}
                        outputs={outputs}
                        outputSummary={outputSummary}
                        broadcastMessage={broadcastMessage}
                        broadcasting={broadcasting}
                        archiving={archiving}
                        onBroadcastMessageChange={onBroadcastMessageChange}
                        onBroadcast={onBroadcast}
                        onArchive={onArchive}
                    />
                    <div className="grid min-h-0 gap-4 overflow-y-auto pr-1 lg:grid-cols-2 2xl:grid-cols-1">
                        <GroupConfigPane selectedGroup={selectedGroup} />
                        <CreateGroupPane
                            draft={draft}
                            notice={notice}
                            error={error}
                            approvalPrompt={approvalPrompt}
                            saving={saving}
                            onDraftChange={onDraftChange}
                            onCreateGroup={onCreateGroup}
                        />
                    </div>
                </div>
            </div>
        </section>
    );
}

function GroupsHeader({ monitor, refreshing, onRefresh }: { monitor: Monitor | null; refreshing: boolean; onRefresh: () => void }) {
    return (
        <div className="rounded-2xl border border-cortex-border bg-cortex-surface p-4">
            <div className="flex flex-wrap items-start justify-between gap-3">
                <div className="min-w-0">
                    <div className="inline-flex items-center gap-2 rounded-full border border-cortex-primary/25 bg-cortex-primary/10 px-3 py-1 text-[11px] font-mono uppercase tracking-[0.18em] text-cortex-primary">
                        <Users className="h-3.5 w-3.5" />Groups Workspace
                    </div>
                    <h1 className="mt-2 text-xl font-semibold text-cortex-text-main">Create, review, and coordinate focused groups.</h1>
                    <p className="mt-1 max-w-3xl text-sm leading-6 text-cortex-text-muted">Use groups as compact collaboration lanes while Soma stays the root admin chat.</p>
                </div>
                <button type="button" onClick={onRefresh} className={compactButtonClassName}>
                    <RefreshCw className={`mr-2 h-4 w-4 ${refreshing ? "animate-spin" : ""}`} />Refresh
                </button>
            </div>
            {monitor ? (
                <div className="mt-3 rounded-xl border border-cortex-border bg-cortex-bg px-3 py-2 text-xs text-cortex-text-muted">
                    Bus {monitor.status || "offline"} | {monitor.published_count ?? 0} published | {monitor.last_group_id ? `last ${monitor.last_group_id}` : "no recent group activity"}{monitor.last_error ? ` | ${monitor.last_error}` : ""}
                </div>
            ) : null}
        </div>
    );
}

function GroupRail({ buckets, selectedGroupId, onSelectGroup }: { buckets: GroupBucket[]; selectedGroupId: string | null; onSelectGroup: (groupId: string) => void }) {
    return (
        <aside className="flex min-h-0 flex-col rounded-2xl border border-cortex-border bg-cortex-surface p-3">
            <div className="flex items-center justify-between border-b border-cortex-border px-1 pb-3">
                <h2 className="text-sm font-semibold uppercase tracking-[0.16em] text-cortex-text-main">Groups</h2>
                <span className="text-[11px] font-mono text-cortex-text-muted">{buckets.reduce((count, bucket) => count + bucket.groups.length, 0)}</span>
            </div>
            <div className="mt-3 min-h-0 flex-1 space-y-4 overflow-y-auto pr-1" data-testid="groups-list">
                {buckets.map((bucket) => (
                    <div key={bucket.id}>
                        <div className="mb-2 flex items-center justify-between gap-2 px-1">
                            <h3 className="text-[11px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted">{bucket.title}</h3>
                            <span className="text-[11px] font-mono text-cortex-text-muted">{bucket.groups.length}</span>
                        </div>
                        <div className="space-y-1">
                            {bucket.groups.length === 0 ? (
                                <p className="px-2 py-1 text-xs text-cortex-text-muted">Nothing here yet.</p>
                            ) : (
                                bucket.groups.map((group) => (
                                    <button
                                        key={group.group_id}
                                        type="button"
                                        data-testid={`groups-list-item-${group.group_id}`}
                                        onClick={() => onSelectGroup(group.group_id)}
                                        className={`w-full rounded-xl px-3 py-2 text-left transition ${selectedGroupId === group.group_id ? "bg-cortex-primary/10 text-cortex-text-main ring-1 ring-cortex-primary/30" : "text-cortex-text-muted hover:bg-cortex-bg hover:text-cortex-text-main"}`}
                                    >
                                        <span className="block truncate text-sm font-semibold">{group.name}</span>
                                        <span className="mt-0.5 flex items-center gap-2 text-[11px] font-mono uppercase tracking-[0.12em]">
                                            {group.status === "archived" ? "Archived" : group.work_mode}
                                            <span className="h-1 w-1 rounded-full bg-current opacity-50" />
                                            {group.team_ids.length} team{group.team_ids.length === 1 ? "" : "s"}
                                        </span>
                                    </button>
                                ))
                            )}
                        </div>
                    </div>
                ))}
            </div>
        </aside>
    );
}

function GroupDetailPane({
    selectedGroup,
    outputs,
    outputSummary,
    broadcastMessage,
    broadcasting,
    archiving,
    onBroadcastMessageChange,
    onBroadcast,
    onArchive,
}: {
    selectedGroup: Group | null;
    outputs: Artifact[];
    outputSummary: OutputSummary;
    broadcastMessage: string;
    broadcasting: boolean;
    archiving: boolean;
    onBroadcastMessageChange: (message: string) => void;
    onBroadcast: () => void;
    onArchive: () => void;
}) {
    if (!selectedGroup) {
        return <section className="min-h-0 overflow-y-auto rounded-2xl border border-cortex-border bg-cortex-surface p-4 text-sm text-cortex-text-muted">Create or select a group to review its lane, config, and outputs.</section>;
    }
    const archived = selectedGroup.status === "archived";
    return (
        <section className="min-h-0 min-w-0 overflow-y-auto rounded-2xl border border-cortex-border bg-cortex-surface p-4">
            <div className="flex flex-wrap items-start justify-between gap-3 border-b border-cortex-border pb-4">
                <div className="min-w-0">
                    <div className="flex flex-wrap gap-2">
                        <Badge>{groupKindLabel(selectedGroup)}</Badge>
                        <Badge muted>{selectedGroup.work_mode}</Badge>
                    </div>
                    <h2 className="mt-3 text-2xl font-semibold text-cortex-text-main">{selectedGroup.name}</h2>
                    <p className="mt-2 max-w-3xl text-sm leading-6 text-cortex-text-muted">{selectedGroup.goal_statement}</p>
                </div>
                {selectedGroup.expiry && !archived ? (
                    <button type="button" onClick={onArchive} disabled={archiving} className={compactButtonClassName}>
                        {archiving ? "Archiving..." : "Archive temporary group"}
                    </button>
                ) : null}
            </div>
            <div className="mt-4 grid gap-4 xl:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)]">
                <section className="rounded-xl border border-cortex-border bg-cortex-bg p-4">
                    <div className="flex items-center gap-2">
                        <MessageSquare className="h-4 w-4 text-cortex-primary" />
                        <h3 className="text-sm font-semibold text-cortex-text-main">{archived ? "Group is archived" : "Broadcast a focused ask"}</h3>
                    </div>
                    {archived ? (
                        <p className="mt-3 text-sm leading-6 text-cortex-text-muted" data-testid="groups-archived-readonly-note">This temporary group is archived. Keep it available for retained output review, but send new coordination through an active group or Soma.</p>
                    ) : (
                        <>
                            <textarea aria-label="Broadcast message" value={broadcastMessage} onChange={(event) => onBroadcastMessageChange(event.target.value)} rows={5} className={`${inputClassName} mt-3 resize-y`} />
                            <button type="button" onClick={onBroadcast} disabled={broadcasting || !broadcastMessage.trim()} className="mt-3 rounded-xl border border-cortex-primary/30 px-4 py-2 text-sm font-semibold text-cortex-primary disabled:opacity-60">{broadcasting ? "Broadcasting..." : "Broadcast to group"}</button>
                        </>
                    )}
                    <div className="mt-4 flex flex-col gap-2">
                        <Link href="/dashboard" className={linkClassName}>Open Soma admin home</Link>
                        {selectedGroup.team_ids.map((teamId) => <Link key={teamId} href={`/dashboard?team_id=${encodeURIComponent(teamId)}`} className={linkClassName}>Open {teamId} lead</Link>)}
                    </div>
                </section>
                <OutputsPanel archived={archived} outputs={outputs} outputSummary={outputSummary} />
            </div>
        </section>
    );
}

function GroupConfigPane({ selectedGroup }: { selectedGroup: Group | null }) {
    const capabilities = selectedGroup?.allowed_capabilities?.length ? selectedGroup.allowed_capabilities.join(", ") : "inherits lane policy";
    return (
        <section className="rounded-2xl border border-cortex-border bg-cortex-surface p-4">
            <div className="flex items-center gap-2">
                <SlidersHorizontal className="h-4 w-4 text-cortex-primary" />
                <h2 className="text-sm font-semibold uppercase tracking-[0.16em] text-cortex-text-main">Group Config</h2>
            </div>
            {selectedGroup ? (
                <div className="mt-4 grid gap-3 text-sm">
                    <Info label="Focused team lead" value={selectedGroup.coordinator_profile || `${selectedGroup.name} lead`} detail="Narrow group lane counterpart." />
                    <Info label="Agent backend model" value="Inherits organization AI Engine" detail="Per-group model override is not persisted by the groups API yet." />
                    <Info label="Approval policy" value={selectedGroup.approval_policy_ref || "default"} detail="Used for group creation and governed actions." />
                    <Info label="Capabilities" value={capabilities} detail="Allowed runtime capabilities for this lane." />
                    <Info label="Created by" value={selectedGroup.created_by} detail={`Created ${relativeTime(selectedGroup.created_at)}`} />
                </div>
            ) : (
                <p className="mt-3 text-sm text-cortex-text-muted">Select a group to inspect lane config.</p>
            )}
        </section>
    );
}

function OutputsPanel({ archived, outputs, outputSummary }: { archived: boolean; outputs: Artifact[]; outputSummary: OutputSummary }) {
    return (
        <section className="rounded-xl border border-cortex-border bg-cortex-bg p-4">
            <h3 className="text-sm font-semibold text-cortex-text-main">{archived ? "Retained outputs" : "Recent outputs"}</h3>
            <div className="mt-2 flex flex-wrap gap-2" data-testid="groups-output-summary">
                <Badge muted>{outputSummary.artifactCount} output{outputSummary.artifactCount === 1 ? "" : "s"}</Badge>
                <Badge muted>{outputSummary.agentCount} contributing lead{outputSummary.agentCount === 1 ? "" : "s"}</Badge>
            </div>
            {archived ? <p className="mt-2 text-sm leading-6 text-cortex-text-muted" data-testid="groups-retained-outputs-note">Review the outputs this archived temporary group already produced. Downloads remain available so the work can still be inspected after the collaboration window closes.</p> : null}
            <div className="mt-3 max-h-[420px] space-y-3 overflow-y-auto pr-1">
                {outputs.length === 0 ? <p className="text-sm text-cortex-text-muted">No recent team outputs found for this group yet.</p> : outputs.slice(0, 8).map((artifact) => <ArtifactRow key={artifact.id} artifact={artifact} />)}
            </div>
        </section>
    );
}

function Info({ label, value, detail }: { label: string; value: string; detail: string }) {
    return <div className="rounded-xl border border-cortex-border bg-cortex-bg p-3"><p className="text-[11px] font-mono uppercase tracking-[0.16em] text-cortex-primary">{label}</p><p className="mt-1 truncate text-sm font-semibold text-cortex-text-main">{value}</p><p className="mt-1 text-xs leading-5 text-cortex-text-muted">{detail}</p></div>;
}

function Badge({ children, muted = false }: { children: React.ReactNode; muted?: boolean }) {
    return <span className={`rounded-full border px-3 py-1 text-[11px] font-mono uppercase tracking-[0.14em] ${muted ? "border-cortex-border bg-cortex-bg text-cortex-text-muted" : "border-cortex-primary/25 bg-cortex-primary/10 text-cortex-primary"}`}>{children}</span>;
}

function ArtifactRow({ artifact }: { artifact: Artifact }) {
    const readable = artifact.artifact_type === "code" || artifact.artifact_type === "document" || artifact.artifact_type === "data" || artifact.artifact_type === "chart";
    return (
        <div className="rounded-xl border border-cortex-border bg-cortex-surface px-3 py-3">
            <div className="flex items-start justify-between gap-3">
                <div className="min-w-0"><p className="truncate text-sm font-semibold text-cortex-text-main">{artifact.title}</p><p className="mt-1 text-[11px] font-mono uppercase tracking-[0.12em] text-cortex-text-muted">{artifact.artifact_type} | {artifact.agent_id} | {relativeTime(artifact.created_at)}</p></div>
                <a href={`/api/v1/artifacts/${encodeURIComponent(artifact.id)}/download`} download className="inline-flex items-center gap-1 rounded-xl border border-cortex-primary/30 px-3 py-2 text-xs font-semibold text-cortex-primary hover:bg-cortex-primary/10"><Download className="h-3.5 w-3.5" />Download</a>
            </div>
            {readable && artifact.content ? <pre className="mt-3 max-h-48 overflow-auto rounded-xl border border-cortex-border bg-cortex-bg p-3 text-xs leading-6 text-cortex-text-muted">{artifact.content}</pre> : artifact.file_path ? <p className="mt-3 text-sm text-cortex-text-muted">Saved path: {artifact.file_path}</p> : null}
        </div>
    );
}
