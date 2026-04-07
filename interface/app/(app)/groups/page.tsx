import GroupManagementPanel from "@/components/teams/GroupManagementPanel";

export default function GroupsPage() {
    return (
        <div className="h-full overflow-auto bg-cortex-bg px-6 py-8">
            <div className="mx-auto max-w-7xl">
                <GroupManagementPanel />
            </div>
        </div>
    );
}
