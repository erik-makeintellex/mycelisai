"use client";

import React, { useCallback, useEffect, useMemo, useState } from "react";
import { useCortexStore, type MCPServerWithTools } from "@/store/useCortexStore";
import type { Artifact } from "@/store/cortexStoreTypesPlanning";
import { extractApiError, formatMCPToolResult, type ResourceCallRequest } from "@/lib/apiContracts";
import { workspaceBrowserPath } from "@/lib/outputPackageModel";
import WorkspaceFolderAccessCard from "./WorkspaceFolderAccessCard";
import WorkspaceGroupOutputSelector, {
    artifactBrowsePath,
    artifactFilePath,
} from "./WorkspaceGroupOutputSelector";
import WorkspaceExplorerMainPane from "./WorkspaceExplorerMainPane";
import {
    initialWorkspacePath,
    useWorkspaceOutputGroups,
    WorkspaceFilesystemUnavailable,
} from "./WorkspaceExplorerSupport";
import { joinPath, parseListOutput } from "./WorkspaceExplorerUtils";

export type WorkspaceEntry = { name: string; path: string; type: "file" | "dir" };
export type WorkspacePane = "browse" | "preview" | "create";

export default function WorkspaceExplorer({ initialPath, onOpenToolsTab }: { initialPath?: string | null; onOpenToolsTab: () => void }) {
    const mcpServers = useCortexStore((s) => s.mcpServers);
    const isFetchingMCPServers = useCortexStore((s) => s.isFetchingMCPServers);
    const fetchMCPServers = useCortexStore((s) => s.fetchMCPServers);

    const [currentPath, setCurrentPath] = useState(() => initialWorkspacePath(initialPath));
    const [entries, setEntries] = useState<WorkspaceEntry[]>([]);
    const [selectedFile, setSelectedFile] = useState<string | null>(null);
    const [preview, setPreview] = useState("");
    const [status, setStatus] = useState<string>("Ready");
    const [busy, setBusy] = useState(false);
    const [newDir, setNewDir] = useState("");
    const [newFile, setNewFile] = useState("");
    const [newFileContent, setNewFileContent] = useState("");
    const [activePane, setActivePane] = useState<WorkspacePane>("browse");
    const {
        outputGroups,
        selectedOutputGroupID,
        setSelectedOutputGroupID,
        includeTeamSourceFiles,
        setIncludeTeamSourceFiles,
        outputGroupStatus,
    } = useWorkspaceOutputGroups();

    useEffect(() => {
        fetchMCPServers();
    }, [fetchMCPServers]);

    useEffect(() => {
        setCurrentPath(initialWorkspacePath(initialPath));
    }, [initialPath]);

    const filesystemServer = useMemo<MCPServerWithTools | null>(() => {
        const match = mcpServers.find((s) => s.name === "filesystem");
        return match ?? null;
    }, [mcpServers]);

    const canBrowse = filesystemServer?.status === "connected";

    const callTool = useCallback(
        async (toolName: string, args: Record<string, unknown>): Promise<string> => {
            if (!filesystemServer) {
                throw new Error("filesystem MCP server not installed");
            }
            const body: ResourceCallRequest<Record<string, unknown>> = { arguments: args };
            const res = await fetch(`/api/v1/mcp/servers/${filesystemServer.id}/tools/${toolName}/call`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(body),
            });
            const payload = await res.json().catch(async () => ({ error: await res.text() }));
            if (!res.ok) {
                const err = extractApiError(payload) ?? (typeof (payload as { error?: unknown }).error === "string" ? (payload as { error?: string }).error : undefined);
                throw new Error(err || "tool call failed");
            }
            return formatMCPToolResult(payload);
        },
        [filesystemServer]
    );

    const refreshList = useCallback(async () => {
        if (!canBrowse) return;
        setBusy(true);
        setStatus(`Listing ${currentPath}...`);
        try {
            const text = await callTool("list_directory", { path: currentPath });
            const parsed = parseListOutput(text, currentPath);
            parsed.sort((a, b) => {
                if (a.type !== b.type) return a.type === "dir" ? -1 : 1;
                return a.name.localeCompare(b.name);
            });
            setEntries(parsed);
            setStatus(`Loaded ${parsed.length} entries`);
        } catch (err) {
            setStatus(err instanceof Error ? err.message : "List failed");
            setEntries([]);
        } finally {
            setBusy(false);
        }
    }, [canBrowse, currentPath, callTool]);

    useEffect(() => {
        if (canBrowse) {
            refreshList();
        }
    }, [canBrowse, currentPath, refreshList]);

    const openFile = async (path: string) => {
        setBusy(true);
        setStatus(`Reading ${path}...`);
        try {
            const text = await callTool("read_text_file", { path });
            setSelectedFile(path);
            setPreview(text);
            setActivePane("preview");
            setStatus(`Opened ${path}`);
        } catch (err) {
            setStatus(err instanceof Error ? err.message : "Read failed");
        } finally {
            setBusy(false);
        }
    };

    const openArtifact = async (artifact: Artifact) => {
        const filePath = artifactFilePath(artifact);
        const browsePath = artifactBrowsePath(artifact);
        const workspacePath = workspaceBrowserPath(browsePath ?? filePath);
        if (workspacePath) {
            setCurrentPath(workspacePath);
        }
        if (filePath && artifact.artifact_type !== "project_package") {
            const readablePath = workspaceBrowserPath(filePath) ?? filePath;
            await openFile(readablePath);
        } else {
            setActivePane("browse");
            setStatus(workspacePath ? `Opened output folder ${workspacePath}` : "Output has no workspace path");
        }
    };

    const openSelectedGroupSourceFolder = (groupID: string) => {
        const group = outputGroups.find((candidate) => candidate.group_id === groupID);
        if (!group?.workspace_folder) {
            setStatus("Selected group has no source folder");
            return;
        }
        const workspacePath = workspaceBrowserPath(group.workspace_folder);
        if (!workspacePath) {
            setStatus("Selected group source folder is not workspace-readable");
            return;
        }
        setCurrentPath(workspacePath);
        setActivePane("browse");
        setStatus(`Showing team source files for ${group.name}`);
    };

    const selectOutputGroup = (groupID: string) => {
        setSelectedOutputGroupID(groupID);
        setIncludeTeamSourceFiles(false);
        const group = outputGroups.find((candidate) => candidate.group_id === groupID);
        const firstOutput = group?.outputs[0];
        const browsePath = workspaceBrowserPath(artifactBrowsePath(firstOutput) ?? artifactFilePath(firstOutput));
        if (browsePath) {
            setCurrentPath(browsePath);
            setActivePane("browse");
            setStatus(`Selected outputs for ${group?.name ?? groupID}`);
        }
    };

    const toggleTeamSourceFiles = (checked: boolean) => {
        setIncludeTeamSourceFiles(checked);
        if (checked) {
            openSelectedGroupSourceFolder(selectedOutputGroupID);
            return;
        }
        const group = outputGroups.find((candidate) => candidate.group_id === selectedOutputGroupID);
        const firstOutput = group?.outputs[0];
        const browsePath = workspaceBrowserPath(artifactBrowsePath(firstOutput) ?? artifactFilePath(firstOutput));
        if (browsePath) {
            setCurrentPath(browsePath);
            setActivePane("browse");
            setStatus(`Showing retained outputs for ${group?.name ?? selectedOutputGroupID}`);
        }
    };

    const createDirectory = async () => {
        if (!newDir.trim()) return;
        setBusy(true);
        try {
            const path = joinPath(currentPath, newDir.trim());
            await callTool("create_directory", { path });
            setNewDir("");
            setStatus(`Created directory ${path}`);
            refreshList();
        } catch (err) {
            setStatus(err instanceof Error ? err.message : "Create directory failed");
        } finally {
            setBusy(false);
        }
    };

    const createFile = async () => {
        if (!newFile.trim()) return;
        setBusy(true);
        try {
            const path = joinPath(currentPath, newFile.trim());
            await callTool("write_file", { path, content: newFileContent });
            setStatus(`Created file ${path}`);
            setNewFile("");
            setNewFileContent("");
            refreshList();
        } catch (err) {
            setStatus(err instanceof Error ? err.message : "Create file failed");
        } finally {
            setBusy(false);
        }
    };

    if (!filesystemServer) {
        return (
            <WorkspaceFilesystemUnavailable
                filesystemServer={filesystemServer}
                onOpenToolsTab={onOpenToolsTab}
                onRefresh={fetchMCPServers}
            />
        );
    }

    if (filesystemServer.status !== "connected") {
        return (
            <WorkspaceFilesystemUnavailable
                filesystemServer={filesystemServer}
                onOpenToolsTab={onOpenToolsTab}
                onRefresh={fetchMCPServers}
            />
        );
    }

    return (
        <div className="grid h-full min-h-0 gap-3 p-4 sm:p-6 lg:grid-cols-[minmax(13rem,17rem)_minmax(19rem,1fr)]">
            <div className="flex min-h-0 flex-col gap-3 overflow-y-auto pr-1">
                <WorkspaceGroupOutputSelector
                    groups={outputGroups}
                    selectedGroupID={selectedOutputGroupID}
                    includeTeamSourceFiles={includeTeamSourceFiles}
                    status={outputGroupStatus}
                    onSelectGroup={selectOutputGroup}
                    onToggleTeamSourceFiles={toggleTeamSourceFiles}
                    onOpenArtifact={openArtifact}
                />

                <WorkspaceFolderAccessCard currentPath={currentPath} onStatus={setStatus} />
            </div>

            <WorkspaceExplorerMainPane
                activePane={activePane}
                busy={busy}
                currentPath={currentPath}
                entries={entries}
                isFetchingMCPServers={isFetchingMCPServers}
                newDir={newDir}
                newFile={newFile}
                newFileContent={newFileContent}
                preview={preview}
                selectedFile={selectedFile}
                status={status}
                onActivePaneChange={setActivePane}
                onCreateDirectory={createDirectory}
                onCreateFile={createFile}
                onCurrentPathChange={setCurrentPath}
                onNewDirChange={setNewDir}
                onNewFileChange={setNewFile}
                onNewFileContentChange={setNewFileContent}
                onOpenFile={openFile}
                onPreviewChange={setPreview}
                onRefresh={refreshList}
            />
        </div>
    );
}
