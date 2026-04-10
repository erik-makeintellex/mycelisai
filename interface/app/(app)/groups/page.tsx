import GroupManagementPanel from "@/components/teams/GroupManagementPanel";

export default async function GroupsPage({
    searchParams,
}: {
    searchParams?: Promise<{ group_id?: string }>;
}) {
    const resolvedSearchParams = (await searchParams) ?? {};
    const initialSelectedGroupId = typeof resolvedSearchParams.group_id === "string" ? resolvedSearchParams.group_id : null;

    return (
        <div className="h-full overflow-auto bg-cortex-bg px-6 py-8">
            <div className="mx-auto max-w-7xl">
                <GroupManagementPanel initialSelectedGroupId={initialSelectedGroupId} />
            </div>
        </div>
    );
}
