import CentralSomaHome from "@/components/dashboard/CentralSomaHome";

export default function DashboardPage({
    searchParams,
}: {
    searchParams?: Promise<{ team_id?: string | string[] }>;
}) {
    const requestedTeamIdPromise = searchParams;
    return (
        <div className="h-full overflow-hidden bg-cortex-bg px-3 py-3 lg:px-4">
            <div className="mx-auto h-full max-w-[1400px] min-h-0">
                <CentralSomaHome requestedTeamIdPromise={requestedTeamIdPromise} />
            </div>
        </div>
    );
}
