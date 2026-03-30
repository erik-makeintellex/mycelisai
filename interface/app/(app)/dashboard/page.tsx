import CentralSomaHome from "@/components/dashboard/CentralSomaHome";
import CreateOrganizationEntry from "@/components/organizations/CreateOrganizationEntry";

export default function DashboardPage() {
    return (
        <div className="h-full overflow-auto bg-cortex-bg px-6 py-8">
            <div className="mx-auto max-w-6xl space-y-8">
                <CentralSomaHome />
                <CreateOrganizationEntry />
            </div>
        </div>
    );
}
