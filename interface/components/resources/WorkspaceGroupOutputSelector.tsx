import React, { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import type { Artifact } from "@/store/cortexStoreTypesPlanning";
import {
    parentWorkspacePath,
    projectPackageOpenPath,
    projectPackageRevealPath,
} from "@/lib/outputPackageModel";

export type OutputGroup = {
    group_id: string;
    name: string;
    workspace_folder?: string;
    outputs: Artifact[];
};

type ContributorLevel = "all" | "lead" | "coder" | "review" | "media" | "other";

const CONTRIBUTOR_LEVELS: Array<{ id: ContributorLevel; label: string }> = [
    { id: "all", label: "All" },
    { id: "lead", label: "Team lead" },
    { id: "coder", label: "Coders" },
    { id: "review", label: "Review" },
    { id: "media", label: "Media" },
    { id: "other", label: "Other" },
];

export default function WorkspaceGroupOutputSelector({
    groups,
    selectedGroupID,
    includeTeamSourceFiles,
    status,
    onSelectGroup,
    onToggleTeamSourceFiles,
    onOpenArtifact,
}: {
    groups: OutputGroup[];
    selectedGroupID: string;
    includeTeamSourceFiles: boolean;
    status: string;
    onSelectGroup: (groupID: string) => void;
    onToggleTeamSourceFiles: (checked: boolean) => void;
    onOpenArtifact: (artifact: Artifact) => void;
}) {
    const [selectedLevel, setSelectedLevel] = useState<ContributorLevel>("all");
    const selectedGroup = groups.find((group) => group.group_id === selectedGroupID) ?? null;
    const selectedGroupHref = selectedGroup
        ? `/groups?group_id=${encodeURIComponent(selectedGroup.group_id)}`
        : "";
    const levelCounts = useMemo(
        () => countContributorLevels(selectedGroup?.outputs ?? []),
        [selectedGroup?.outputs],
    );
    const visibleOutputs = useMemo(
        () =>
            (selectedGroup?.outputs ?? []).filter(
                (artifact) =>
                    selectedLevel === "all" ||
                    artifactContributorLevel(artifact) === selectedLevel,
            ),
        [selectedGroup?.outputs, selectedLevel],
    );

    useEffect(() => {
        setSelectedLevel("all");
    }, [selectedGroupID]);

    return (
        <section
            className="rounded-lg border border-cortex-border bg-cortex-surface p-3"
            data-testid="workspace-group-output-selector"
        >
            <div className="flex flex-wrap items-center gap-2">
                <div className="min-w-[10rem]">
                    <p className="text-xs font-semibold uppercase tracking-[0.18em] text-cortex-primary">
                        Group outputs
                    </p>
                    <span
                        className="mt-1 inline-flex rounded border border-cortex-border bg-cortex-bg px-2 py-1 text-[10px] font-mono uppercase tracking-normal text-cortex-text-muted"
                        title={status}
                    >
                        {groups.length} with retained output
                    </span>
                </div>

                <div className="min-w-[12rem] flex-1">
                    <label
                        className="text-[11px] font-mono uppercase tracking-[0.16em] text-cortex-text-muted"
                        htmlFor="workspace-output-group"
                    >
                        Select group
                    </label>
                    <select
                        id="workspace-output-group"
                        value={selectedGroupID}
                        onChange={(event) => onSelectGroup(event.target.value)}
                        disabled={groups.length === 0}
                        className="mt-1 w-full rounded border border-cortex-border bg-cortex-bg px-3 py-1.5 text-sm text-cortex-text-main outline-none disabled:opacity-60"
                    >
                        {groups.length === 0 ? (
                            <option value="">No group outputs</option>
                        ) : (
                            groups.map((group) => (
                                <option key={group.group_id} value={group.group_id}>
                                    {group.name} ({group.outputs.length})
                                </option>
                            ))
                        )}
                    </select>
                </div>

                <label className="inline-flex h-9 items-center gap-2 rounded border border-cortex-border bg-cortex-bg px-3 text-xs text-cortex-text-main">
                    <input
                        type="checkbox"
                        checked={includeTeamSourceFiles}
                        disabled={!selectedGroup?.workspace_folder}
                        onChange={(event) => onToggleTeamSourceFiles(event.target.checked)}
                        className="h-4 w-4 accent-cortex-primary"
                    />
                    Include team source files
                </label>

                {selectedGroup ? (
                    <div className="flex flex-wrap gap-2">
                        <Link
                            href={`${selectedGroupHref}&panel=outputs`}
                            className="inline-flex h-9 items-center rounded border border-cortex-primary/35 px-3 text-xs font-semibold text-cortex-primary hover:bg-cortex-primary/10"
                        >
                            Open group outputs
                        </Link>
                        <Link
                            href={`${selectedGroupHref}&panel=workflow`}
                            className="inline-flex h-9 items-center rounded border border-cortex-border bg-cortex-bg px-3 text-xs text-cortex-text-main hover:border-cortex-primary/30"
                        >
                            Workflow log
                        </Link>
                        <Link
                            href={`${selectedGroupHref}&panel=message`}
                            className="inline-flex h-9 items-center rounded border border-cortex-border bg-cortex-bg px-3 text-xs text-cortex-text-main hover:border-cortex-primary/30"
                        >
                            Message group
                        </Link>
                    </div>
                ) : null}
            </div>

            {selectedGroup ? (
                <div className="mt-2 space-y-2">
                    <div
                        className="grid grid-cols-3 gap-2"
                        role="tablist"
                        aria-label="Output contributor level"
                    >
                        {CONTRIBUTOR_LEVELS.map((level) => {
                            const count = level.id === "all"
                                ? selectedGroup.outputs.length
                                : levelCounts[level.id] ?? 0;
                            const selected = selectedLevel === level.id;
                            return (
                                <button
                                    key={level.id}
                                    type="button"
                                    role="tab"
                                    aria-label={`${level.label} ${count}`}
                                    aria-selected={selected}
                                    onClick={() => setSelectedLevel(level.id)}
                                    className={`min-w-0 rounded border px-3 py-1.5 text-left text-xs transition-colors ${
                                        selected
                                            ? "border-cortex-primary/50 bg-cortex-primary/10 text-cortex-text-main"
                                            : "border-cortex-border bg-cortex-bg text-cortex-text-muted hover:bg-cortex-bg/80"
                                    }`}
                                >
                                    <span className="block truncate font-semibold">{level.label}</span>
                                    <span className="font-mono text-[10px]">{count}</span>
                                </button>
                            );
                        })}
                    </div>

                    <div
                        className="max-h-40 overflow-y-auto rounded border border-cortex-border bg-cortex-bg"
                        role="table"
                        aria-label="Retained output artifacts"
                    >
                        {visibleOutputs.length === 0 ? (
                            <p className="p-3 text-xs text-cortex-text-muted">
                                No retained outputs at this contributor level.
                            </p>
                        ) : (
                            visibleOutputs.map((artifact) => (
                                <button
                                    key={artifact.id}
                                    type="button"
                                    onClick={() => onOpenArtifact(artifact)}
                                    className="block w-full border-b border-cortex-border/50 px-3 py-2 text-left transition-colors last:border-b-0 hover:bg-cortex-primary/10"
                                >
                                    <span className="block truncate text-sm font-semibold text-cortex-text-main">
                                        {artifact.title}
                                    </span>
                                    <span className="mt-1 block truncate text-[11px] font-mono uppercase tracking-normal text-cortex-text-muted">
                                        {artifactContributorLabel(artifact)} | {artifact.agent_id}
                                    </span>
                                    <span className="mt-1 block truncate text-[11px] font-mono text-cortex-primary">
                                        {artifactFilePath(artifact) || artifactBrowsePath(artifact) || "retained artifact"}
                                    </span>
                                </button>
                            ))
                        )}
                    </div>
                </div>
            ) : null}
        </section>
    );
}

export function artifactFilePath(artifact: Artifact | undefined) {
    if (!artifact) return null;
    if (artifact.artifact_type === "project_package") {
        return projectPackageOpenPath({
            folder: artifactMetadataString(artifact, "folder"),
            entrypoint: artifactMetadataString(artifact, "entrypoint"),
            filePath: artifact.file_path,
        });
    }
    return (
        artifact.file_path?.trim()
        || artifactMetadataString(artifact, "file_path", "path", "saved_path", "storage_ref", "entrypoint")
        || null
    );
}

export function artifactBrowsePath(artifact: Artifact | undefined) {
    if (!artifact) return null;
    if (artifact.artifact_type === "project_package") {
        return projectPackageRevealPath({
            folder: artifactMetadataString(artifact, "folder"),
            entrypoint: artifactMetadataString(artifact, "entrypoint"),
            filePath: artifact.file_path,
        });
    }
    const filePath = artifactFilePath(artifact);
    const folder = artifactMetadataString(artifact, "folder");
    return parentWorkspacePath(filePath) ?? (folder || filePath);
}

function artifactContributorLabel(artifact: Artifact) {
    return CONTRIBUTOR_LEVELS.find(
        (level) => level.id === artifactContributorLevel(artifact),
    )?.label ?? "Other";
}

function artifactContributorLevel(artifact: Artifact): Exclude<ContributorLevel, "all"> {
    const source = [
        artifactMetadataString(
            artifact,
            "contributor_level",
            "agent_level",
            "role",
            "agent_role",
            "team_role",
            "member_role",
        ),
        artifact.agent_id,
        artifact.title,
        artifact.artifact_type,
    ].join(" ").toLowerCase();

    if (/(lead|coordinator|manager|architect|planner|director)/.test(source)) return "lead";
    if (/(review|qa|test|validate|verification|proof)/.test(source)) return "review";
    if (/(media|image|audio|art|artist|design|sprite|asset)/.test(source)) return "media";
    if (/(code|coder|engineer|developer|frontend|backend|gameplay|systems?)/.test(source)) return "coder";
    return "other";
}

function countContributorLevels(outputs: Artifact[]) {
    return outputs.reduce<Record<Exclude<ContributorLevel, "all">, number>>(
        (counts, artifact) => {
            counts[artifactContributorLevel(artifact)] += 1;
            return counts;
        },
        { lead: 0, coder: 0, review: 0, media: 0, other: 0 },
    );
}

function artifactMetadataString(artifact: Artifact | undefined, ...keys: string[]) {
    if (!artifact?.metadata) return "";
    for (const key of keys) {
        const value = artifact.metadata[key];
        if (typeof value === "string" && value.trim()) return value.trim();
    }
    return "";
}
