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
    type OutputGroup,
} from "./WorkspaceGroupOutputSelector";
import WorkspaceMCPRecoveryCard from "./WorkspaceMCPRecoveryCard";
import {
    WorkspaceBrowsePane,
    WorkspaceCreatePane,
    WorkspacePaneTabs,
    WorkspacePreviewPane,
} from "./WorkspaceExplorerPanes";
import { joinPath, normalizePath, parseListOutput } from "./WorkspaceExplorerUtils";

const WORKSPACE_ROOT_PATH = "workspace";
export type WorkspaceEntry = { name: string; path: string; type: "file" | "dir" };
export type WorkspacePane = "browse" | "preview" | "create";

type GroupRecord = {
    group_id: string;
    name: string;
    workspace_folder?: string;
};

function initialWorkspacePath(path?: string | null) {
    const normalized = normalizePath(path?.trim() || WORKSPACE_ROOT_PATH);
    if (normalized === ".") return WORKSPACE_ROOT_PATH;
    if (/^(groups|generated|outputs|reports|logs|saved-media)(\/|$)/i.test(normalized)) {
        return `${WORKSPACE_ROOT_PATH}/${normalized}`;
    }
    return normalized;
}

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
    const [outputGroups, setOutputGroups] = useState<OutputGroup[]>([]);
    const [selectedOutputGroupID, setSelectedOutputGroupID] = useState("");
    const [includeTeamSourceFiles, setIncludeTeamSourceFiles] = useState(false);
    const [outputGroupStatus, setOutputGroupStatus] = useState("Loading group outputs...");

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

    useEffect(() => {
        let cancelled = false;

        const loadOutputGroups = async () => {
            setOutputGroupStatus("Loading group outputs...");
            try {
                const groupsRes = await fetch("/api/v1/groups", { cache: "no-store" });
                if (!groupsRes.ok) {
                    throw new Error("Could not load groups");
                }
                const groups = await responseData<GroupRecord[]>(groupsRes);
                const loaded: Array<OutputGroup | null> = await Promise.all(
                    groups.map(async (group) => {
                        try {
                            const outputsRes = await fetch(
                                `/api/v1/groups/${encodeURIComponent(group.group_id)}/outputs?limit=20`,
                                { cache: "no-store" },
                            );
                            if (!outputsRes.ok) return null;
                            const outputs = await responseData<Artifact[]>(outputsRes);
                            if (!Array.isArray(outputs) || outputs.length === 0) return null;
                            return {
                                group_id: group.group_id,
                                name: group.name || group.group_id,
                                workspace_folder: group.workspace_folder,
                                outputs,
                            } satisfies OutputGroup;
                        } catch {
                            return null;
                        }
                    }),
                );

                if (cancelled) return;
                const withOutputs = loaded.filter((group): group is OutputGroup => Boolean(group));
                setOutputGroups(withOutputs);
                setSelectedOutputGroupID((current) => {
                    if (current && withOutputs.some((group) => group.group_id === current)) return current;
                    return withOutputs[0]?.group_id ?? "";
                });
                setOutputGroupStatus(
                    withOutputs.length > 0
                        ? `Loaded ${withOutputs.length} group${withOutputs.length === 1 ? "" : "s"} with outputs`
                        : "No groups with retained user outputs yet",
                );
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
            <WorkspaceMCPRecoveryCard
                title="Filesystem MCP not installed"
                detail="Output Files needs the filesystem capability before Mycelis can browse generated files here. Install it from Capabilities, or view storage roots to confirm where generated content is mounted."
                onOpenToolsTab={onOpenToolsTab}
                onRefresh={fetchMCPServers}
            />
        );
    }

    if (filesystemServer.status !== "connected") {
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
                onRefresh={fetchMCPServers}
            />
        );
    }

    return (
        <div className="flex h-full min-h-0 flex-col gap-3 p-4 sm:p-6">
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

            <section className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-lg border border-cortex-border bg-cortex-surface">
                <div className="border-b border-cortex-border bg-cortex-bg/70 px-3 py-3">
                    <div className="mb-3 flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
                        <div className="min-w-0">
                            <p className="text-xs font-semibold text-cortex-text-main">Workspace output access</p>
                            <p className="mt-1 text-[11px] leading-5 text-cortex-text-muted">
                                Start with retained files, then open a preview or create a small handoff artifact when needed.
                            </p>
                        </div>
                        <div className="flex min-w-0 items-center gap-2">
                            <span className="text-[11px] font-mono text-cortex-text-muted">Path</span>
                            <code className="max-w-full truncate rounded border border-cortex-border bg-cortex-surface px-2 py-1 text-xs font-mono text-cortex-primary md:max-w-[22rem]">
                                {currentPath}
                            </code>
                        </div>
                    </div>

                    <WorkspacePaneTabs activePane={activePane} onSelect={setActivePane} />
                </div>

                <div className="min-h-0 flex-1 overflow-y-auto p-3">
                    {activePane === "browse" && (
                        <WorkspaceBrowsePane
                            currentPath={currentPath}
                            entries={entries}
                            busy={busy}
                            isFetchingMCPServers={isFetchingMCPServers}
                            onUp={() =>
                                setCurrentPath(
                                    currentPath === WORKSPACE_ROOT_PATH
                                        ? WORKSPACE_ROOT_PATH
                                        : normalizePath(`${currentPath}/..`),
                                )
                            }
                            onRefresh={refreshList}
                            onOpenFolder={setCurrentPath}
                            onOpenFile={openFile}
                        />
                    )}

                    {activePane === "preview" && (
                        <WorkspacePreviewPane
                            selectedFile={selectedFile}
                            preview={preview}
                            onPreviewChange={setPreview}
                        />
                    )}

                    {activePane === "create" && (
                        <WorkspaceCreatePane
                            newDir={newDir}
                            newFile={newFile}
                            newFileContent={newFileContent}
                            onNewDirChange={setNewDir}
                            onNewFileChange={setNewFile}
                            onNewFileContentChange={setNewFileContent}
                            onCreateDirectory={createDirectory}
                            onCreateFile={createFile}
                        />
                    )}
                </div>

                <div className="border-t border-cortex-border bg-cortex-bg px-3 py-2 text-[11px] font-mono text-cortex-text-muted">
                    <span className="block truncate">{status}</span>
                </div>
            </section>
        </div>
    );
}

async function responseData<T>(res: Response): Promise<T> {
    const payload = await res.json();
    return (
        payload && typeof payload === "object" && "data" in payload
            ? (payload as { data: T }).data
            : payload
    ) as T;
}
