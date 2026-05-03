import CentralSomaHome from "@/components/dashboard/CentralSomaHome";
import CreateOrganizationEntry from "@/components/organizations/CreateOrganizationEntry";

export default function DashboardPage({
    searchParams,
}: {
    searchParams?: Promise<{ team_id?: string | string[] }>;
}) {
    const requestedTeamIdPromise = searchParams;
    return (
        <div className="h-full overflow-auto bg-cortex-bg px-4 py-4 lg:px-5 lg:py-5">
            <div className="mx-auto max-w-[1400px] space-y-4">
                <CentralSomaHome requestedTeamIdPromise={requestedTeamIdPromise} />
                <details id="dashboard-organization-setup" className="rounded-3xl border border-cortex-border bg-cortex-surface px-5 py-4">
                    <summary className="cursor-pointer list-none text-sm font-semibold text-cortex-text-main">
                        Create or open AI Organizations
                    </summary>
                    <p className="mt-2 text-sm leading-6 text-cortex-text-muted">
                        Soma is the default place to begin. Open organization setup when you want to create a durable context
                        or resume a structured workspace behind the conversation.
                    </p>
                    <div className="mt-5">
                        <CreateOrganizationEntry />
                    </div>
                </details>
            </div>
        </div>
    );
}
