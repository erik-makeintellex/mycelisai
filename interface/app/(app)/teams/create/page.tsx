import { Suspense } from "react";
import TeamCreationPage from "@/components/teams/TeamCreationPage";

export default function CreateTeamRoutePage() {
    return (
        <Suspense fallback={null}>
            <TeamCreationPage />
        </Suspense>
    );
}
