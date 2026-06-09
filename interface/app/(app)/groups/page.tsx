import GroupManagementPanel from "@/components/teams/GroupManagementPanel";
import AdvancedModeRoute from "@/components/shared/AdvancedModeRoute";

export default async function GroupsPage({
    searchParams,
}: {
    searchParams?: Promise<{ group_id?: string }>;
}) {
    const resolvedSearchParams = (await searchParams) ?? {};
    const initialSelectedGroupId = typeof resolvedSearchParams.group_id === "string" ? resolvedSearchParams.group_id : null;

    return (
        <AdvancedModeRoute
            title="Groups are an Advanced coordination view"
            summary="Use Soma and Work to Review for day-to-day outcomes. Open Groups when an admin needs to inspect standing teams, broadcasts, and cross-team coordination."
        >
            <div className="h-full overflow-auto bg-cortex-bg px-6 py-8">
                <div className="mx-auto max-w-7xl">
                    <GroupManagementPanel initialSelectedGroupId={initialSelectedGroupId} />
                </div>
            </div>
        </AdvancedModeRoute>
    );
}
