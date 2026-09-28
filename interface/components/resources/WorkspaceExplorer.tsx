"use client";

import React, { useCallback, useEffect, useMemo, useState } from "react";
import { useCortexStore, type MCPServerWithTools } from "@/store/useCortexStore";
import type { Artifact } from "@/store/cortexStoreTypesPlanning";
import { extractApiError, formatMCPToolResult, type ResourceCallRequest } from "@/lib/apiContracts";
import { workspaceBrowserPath } from "@/lib/outputPackageModel";
import { requestSomaOutputContinuation } from "@/components/soma/outputContinuation";
import { useIsAdmin } from "@/lib/useIsAdmin";
import InlineBlockerNotice from "@/components/shared/InlineBlockerNotice";
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

// MCPS D9: direct create/write calls are approver-only (D4); Core is the
// authority, this is just a UI hint (`role === "admin"`) to avoid a doomed
// request. Any of these codes coming back from the API still renders honest
// blocker copy rather than a raw error string.
type ToolCallError = Error & { blockerCode?: string; httpStatus?: number };
const MCP_BLOCKER_CODES = new Set(["admin_required", "mcp_call_forbidden", "mcp_tool_not_found", "service_unavailable"]);

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
    const [blocker, setBlocker] = useState<{ code: string; httpStatus: number } | null>(null);
    const { isAdmin } = useIsAdmin();
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
    const {
        outputGroups,
        selectedOutputGroupID,
        setSelectedOutputGroupID,
        includeTeamSourceFiles,
        setIncludeTeamSourceFiles,
        outputGroupStatus,
    } = useWorkspaceOutputGroups(canBrowse);

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
                const code = (payload as { data?: { code?: unknown } })?.data?.code;
                const toolError: ToolCallError = new Error(err || "tool call failed");
                toolError.httpStatus = res.status;
                toolError.blockerCode = typeof code === "string" ? code : undefined;
                throw toolError;
            }
            return formatMCPToolResult(payload);
        },
        [filesystemServer]
    );

    // Surfaces a real backend blocker as honest copy instead of a raw status
    // string; anything else keeps the existing plain-text status behavior.
    const reportToolFailure = useCallback((err: unknown, fallback: string) => {
        const code = err instanceof Error ? (err as ToolCallError).blockerCode : undefined;
        if (code && MCP_BLOCKER_CODES.has(code)) {
            setBlocker({ code, httpStatus: (err as ToolCallError).httpStatus ?? 0 });
            return;
        }
        setStatus(err instanceof Error ? err.message : fallback);
    }, []);

    const changeActivePane = (pane: WorkspacePane) => {
        if (pane === "create" && !isAdmin) {
            setBlocker({ code: "admin_required", httpStatus: 403 });
            return;
        }
        setBlocker(null);
        setActivePane(pane);
    };

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
            reportToolFailure(err, "List failed");
            setEntries([]);
        } finally {
            setBusy(false);
        }
    }, [canBrowse, currentPath, callTool, reportToolFailure]);

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
            reportToolFailure(err, "Read failed");
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

    const askSomaWithSelectedFile = () => {
        if (!selectedFile) return;
        const title = selectedFile.split("/").filter(Boolean).at(-1) ?? selectedFile;
        requestSomaOutputContinuation(
            {
                title,
                reference: selectedFile,
                sourceLabel: "selected file",
            },
            { persist: true, openSoma: true },
        );
    };

    const createDirectory = async () => {
        if (!newDir.trim()) return;
        if (!isAdmin) {
            setBlocker({ code: "admin_required", httpStatus: 403 });
            return;
        }
        setBusy(true);
        try {
            const path = joinPath(currentPath, newDir.trim());
            await callTool("create_directory", { path });
            setNewDir("");
            setBlocker(null);
            setStatus(`Created directory ${path}`);
            refreshList();
        } catch (err) {
            reportToolFailure(err, "Create directory failed");
        } finally {
            setBusy(false);
        }
    };

    const createFile = async () => {
        if (!newFile.trim()) return;
        if (!isAdmin) {
            setBlocker({ code: "admin_required", httpStatus: 403 });
            return;
        }
        setBusy(true);
        try {
            const path = joinPath(currentPath, newFile.trim());
            await callTool("write_file", { path, content: newFileContent });
            setBlocker(null);
            setStatus(`Created file ${path}`);
            setNewFile("");
            setNewFileContent("");
            refreshList();
        } catch (err) {
            reportToolFailure(err, "Create file failed");
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
        <div className="flex h-full min-h-0 flex-col gap-3 p-4 sm:p-5">
            <div className="flex flex-shrink-0 flex-col gap-3">
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

                {blocker ? (
                    <InlineBlockerNotice
                        code={blocker.code}
                        httpStatus={blocker.httpStatus}
                        viewerIsAdmin={isAdmin}
                        reason="mcp_call"
                        onDismiss={() => setBlocker(null)}
                    />
                ) : null}
            </div>

            <div className="min-h-[24rem] flex-1">
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
                    onActivePaneChange={changeActivePane}
                    onCreateDirectory={createDirectory}
                    onCreateFile={createFile}
                    onCurrentPathChange={setCurrentPath}
                    onNewDirChange={setNewDir}
                    onNewFileChange={setNewFile}
                    onNewFileContentChange={setNewFileContent}
                    onAskSomaWithSelectedFile={askSomaWithSelectedFile}
                    onOpenFile={openFile}
                    onPreviewChange={setPreview}
                    onRefresh={refreshList}
                />
            </div>
        </div>
    );
}
