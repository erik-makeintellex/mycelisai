import { expect, type Page } from '@playwright/test';

type SeedResponse = {
    mission_id?: string;
};

type MissionDetail = {
    teams?: Array<{
        id?: string;
        name?: string;
    }>;
};

async function parseJSONIfPossible<T>(response: { text(): Promise<string> }) {
    const raw = await response.text();
    try {
        return { raw, body: JSON.parse(raw) as T };
    } catch {
        return { raw, body: null as T | null };
    }
}

export async function createLiveMissionTeam(page: Page) {
    const seedResponse = await page.request.post('/api/v1/intent/seed/symbiotic');
    const parsedSeed = await parseJSONIfPossible<SeedResponse>(seedResponse);
    expect(seedResponse.ok(), parsedSeed.body ? JSON.stringify(parsedSeed.body) : parsedSeed.raw).toBeTruthy();
    expect(parsedSeed.body?.mission_id).toBeTruthy();

    const missionResponse = await page.request.get(`/api/v1/missions/${parsedSeed.body!.mission_id}`);
    const parsedMission = await parseJSONIfPossible<MissionDetail>(missionResponse);
    expect(missionResponse.ok(), parsedMission.body ? JSON.stringify(parsedMission.body) : parsedMission.raw).toBeTruthy();
    const team = parsedMission.body?.teams?.find((candidate) => candidate.id);
    expect(team?.id).toBeTruthy();
    return {
        missionID: parsedSeed.body!.mission_id!,
        teamID: team!.id!,
        teamName: team?.name ?? 'live mission team',
    };
}
