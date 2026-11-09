import type { Artifact, ChatMessage } from "@/store/useCortexStore";
import { councilLabel } from "@/lib/labels";

export interface DeliverableItem {
    id: string;
    title: string;
    contentType: string;
    filePath?: string;
    content?: string;
    producer: string;
    createdAt?: string;
    saved: boolean;
}

export function isPlayableHtmlDeliverable(item: DeliverableItem): boolean {
    return item.contentType.toLowerCase().includes("html") || Boolean(item.filePath?.toLowerCase().endsWith(".html"));
}

function downloadBlob(body: string, contentType: string, filename: string) {
    if (typeof document === "undefined") return;
    if (typeof URL !== "undefined" && typeof URL.createObjectURL === "function") {
        const blob = new Blob([body], { type: contentType });
        const url = URL.createObjectURL(blob);
        const link = document.createElement("a");
        link.href = url;
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        if (typeof URL.revokeObjectURL === "function") {
            URL.revokeObjectURL(url);
        }
    }
}

export function triggerDownload(item: DeliverableItem) {
    const body = item.content ?? `Saved workspace file: ${item.filePath ?? item.title}\n`;
    const filename = item.filePath?.split("/").pop() || item.title || "deliverable.txt";
    downloadBlob(body, item.contentType || "text/plain;charset=utf-8", filename);
}

export function triggerBundleExport(items: DeliverableItem[]): string {
    const sections = items.map((item, idx) => {
        const header = `## ${idx + 1}. ${item.title}\n- **Producer:** ${item.producer}\n- **Content-Type:** ${item.contentType}\n- **Path:** ${item.filePath || "workspace-inline"}`;
        const body = item.content || `Saved workspace file: ${item.filePath ?? item.title}`;
        return `${header}\n\n\`\`\`\n${body}\n\`\`\``;
    });
    const bundleContent = `# Workspace Deliverables Bundle (${items.length} items)\n\n${sections.join("\n\n---\n\n")}\n`;
    downloadBlob(bundleContent, "text/markdown;charset=utf-8", "workspace-deliverables-bundle.md");
    return bundleContent;
}

export function buildDeliverableItems(
    artifacts: Artifact[],
    missionChat: ChatMessage[],
    assistantName: string,
): DeliverableItem[] {
    const list: DeliverableItem[] = artifacts.map((art: Artifact) => ({
        id: art.id,
        title: art.title || art.file_path || "Generated Deliverable",
        contentType: art.content_type || "text/plain",
        filePath: art.file_path,
        content: art.content,
        producer: councilLabel(art.agent_id || "admin", assistantName).name,
        createdAt: art.created_at,
        saved: Boolean(art.file_path) || art.status === "approved",
    }));

    const seenIds = new Set(list.map((i) => i.id));
    for (const msg of missionChat) {
        if (!Array.isArray(msg.artifacts)) continue;
        for (const chatArt of msg.artifacts) {
            const artId = chatArt.id || chatArt.saved_path || chatArt.title || "workspace-artifact";
            if (seenIds.has(artId)) continue;
            seenIds.add(artId);
            list.push({
                id: artId,
                title: chatArt.title || chatArt.saved_path || "Workspace Artifact",
                contentType: chatArt.content_type || "text/plain",
                filePath: chatArt.saved_path,
                content: chatArt.content,
                producer: assistantName,
                saved: Boolean(chatArt.saved_path),
            });
        }
    }
    return list;
}
