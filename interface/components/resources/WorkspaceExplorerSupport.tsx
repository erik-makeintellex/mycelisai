"use client";

import React, { useEffect, useState } from "react";
import type { MCPServerWithTools } from "@/store/useCortexStore";
import type { Artifact } from "@/store/cortexStoreTypesPlanning";
import WorkspaceMCPRecoveryCard from "./WorkspaceMCPRecoveryCard";
import {
    artifactBrowsePath,
    artifactFilePath,
    type OutputGroup,
} from "./WorkspaceGroupOutputSelector";
import { normalizePath } from "./WorkspaceExplorerUtils";

export const WORKSPACE_ROOT_PATH = "workspace";

type GroupRecord = {
    group_id: string;
    name: string;
    workspace_folder?: string;
};

export function initialWorkspacePath(path?: string | null) {
    const normalized = normalizePath(path?.trim() || WORKSPACE_ROOT_PATH);
    if (normalized === ".") return WORKSPACE_ROOT_PATH;
    if (/^(groups|generated|outputs|reports|logs|saved-media)(\/|$)/i.test(normalized)) {
        return `${WORKSPACE_ROOT_PATH}/${normalized}`;
    }
    return normalized;
}

export function useWorkspaceOutputGroups() {
    const [outputGroups, setOutputGroups] = useState<OutputGroup[]>([]);
    const [selectedOutputGroupID, setSelectedOutputGroupID] = useState("");
    const [includeTeamSourceFiles, setIncludeTeamSourceFiles] = useState(false);
    const [outputGroupStatus, setOutputGroupStatus] = useState("Loading group outputs...");

    useEffect(() => {
        let cancelled = false;

        const loadOutputGroups = async () => {
            setOutputGroupStatus("Loading group outputs...");
            try {
                const groupsRes = await fetch("/api/v1/groups", { cache: "no-store" });
                if (!groupsRes.ok) throw new Error("Could not load groups");
                const groups = await responseData<GroupRecord[]>(groupsRes);
                const loaded = await Promise.all(groups.map(loadGroupOutputs));

                if (cancelled) return;
                const withOutputs = loaded.filter((group): group is OutputGroup => Boolean(group));
                setOutputGroups(withOutputs);
                setSelectedOutputGroupID((current) => {
                    if (current && withOutputs.some((group) => group.group_id === current)) return current;
                    return withOutputs[0]?.group_id ?? "";
                });
                setOutputGroupStatus(outputGroupStatusFor(withOutputs.length));
            } catch (error) {
                if (cancelled) return;
                setOutputGroups([]);
                setSelectedOutputGroupID("");
                setOutputGroupStatus(error instanceof Error ? error.message : "Could not load group outputs");
            }
        };

        loadOutputGroups();
        return () => {
            cancelled = true;
        };
    }, []);

    return {
        outputGroups,
        selectedOutputGroupID,
        setSelectedOutputGroupID,
        includeTeamSourceFiles,
        setIncludeTeamSourceFiles,
        outputGroupStatus,
    };
}

export function WorkspaceFilesystemUnavailable({
    filesystemServer,
    onOpenToolsTab,
    onRefresh,
}: {
    filesystemServer: MCPServerWithTools | null;
    onOpenToolsTab: () => void;
    onRefresh: () => void;
}) {
    if (!filesystemServer) {
        return (
            <WorkspaceMCPRecoveryCard
                title="Filesystem MCP not installed"
                detail="Output Files needs the filesystem capability before Mycelis can browse generated files here. Install it from Capabilities, or view storage roots to confirm where generated content is mounted."
                onOpenToolsTab={onOpenToolsTab}
                onRefresh={onRefresh}
            />
        );
    }

    return (
        <WorkspaceMCPRecoveryCard
            title="Filesystem MCP not connected"
            detail={(
                <>
                    Current status: <span className="font-mono">{filesystemServer.status}</span>. Reconnect it
                    from Capabilities, then retry. View storage roots if you need to find generated output while
                    the MCP server is recovering.
                </>
            )}
            onOpenToolsTab={onOpenToolsTab}
            onRefresh={onRefresh}
        />
    );
}

function outputGroupStatusFor(count: number) {
    if (count <= 0) return "No groups with retained user outputs yet";
    return `Loaded ${count} group${count === 1 ? "" : "s"} with outputs`;
}

async function loadGroupOutputs(group: GroupRecord): Promise<OutputGroup | null> {
    try {
        const outputsRes = await fetch(`/api/v1/groups/${encodeURIComponent(group.group_id)}/outputs?limit=20`, {
            cache: "no-store",
        });
        if (!outputsRes.ok) return null;
        const outputs = await responseData<Artifact[]>(outputsRes);
        if (!Array.isArray(outputs) || outputs.length === 0) return null;
        const firstOutput = outputs[0];
        const hasReadableOutput = artifactBrowsePath(firstOutput) || artifactFilePath(firstOutput);
        if (!hasReadableOutput) return null;
        return {
            group_id: group.group_id,
            name: group.name || group.group_id,
            workspace_folder: group.workspace_folder,
            outputs,
        };
    } catch {
        return null;
    }
}

async function responseData<T>(res: Response): Promise<T> {
    const payload = await res.json();
    return (
        payload && typeof payload === "object" && "data" in payload
            ? (payload as { data: T }).data
            : payload
    ) as T;
}
