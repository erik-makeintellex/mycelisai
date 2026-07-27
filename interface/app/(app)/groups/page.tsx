import GroupManagementPanel from "@/components/teams/GroupManagementPanel";
import type { GroupWorkspacePanel } from "@/components/teams/GroupWorkspaceTabs";

export default async function GroupsPage({
    searchParams,
}: {
    searchParams?: Promise<{ group_id?: string; panel?: string }>;
}) {
    const resolvedSearchParams = (await searchParams) ?? {};
    const initialSelectedGroupId = typeof resolvedSearchParams.group_id === "string" ? resolvedSearchParams.group_id : null;
    const initialPanel = parseInitialPanel(resolvedSearchParams.panel);

    return (
        <div className="h-full overflow-hidden bg-cortex-bg px-3 py-3 sm:px-4 sm:py-4 lg:px-6 lg:py-6">
            <div className="mx-auto h-full max-w-7xl">
                <GroupManagementPanel
                    initialSelectedGroupId={initialSelectedGroupId}
                    initialPanel={initialPanel}
                />
            </div>
        </div>
    );
}

function parseInitialPanel(panel: string | undefined): GroupWorkspacePanel | null {
    if (
        panel === "overview" ||
        panel === "workflow" ||
        panel === "outputs" ||
        panel === "message" ||
        panel === "settings" ||
        panel === "create"
    ) {
        return panel;
    }
    return null;
}
