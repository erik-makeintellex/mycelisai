import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';

vi.mock('lucide-react', () => ({
    X: (props: any) => <svg data-testid="x-icon" {...props} />,
    FileJson: (props: any) => <svg data-testid="file-json-icon" {...props} />,
    Upload: (props: any) => <svg data-testid="upload-icon" {...props} />,
    Download: (props: any) => <svg data-testid="download-icon" {...props} />,
}));

import BlueprintDrawer from '@/components/workspace/BlueprintDrawer';
import { useCortexStore, type MissionBlueprint } from '@/store/useCortexStore';

const makeMockBlueprint = (id: string, teamCount: number, agentsPerTeam: number): MissionBlueprint => ({
    mission_id: id,
    intent: `Intent for ${id}`,
    teams: Array.from({ length: teamCount }, (_, tIdx) => ({
        name: `Team-${tIdx}`,
        role: 'observer',
        agents: Array.from({ length: agentsPerTeam }, (_, aIdx) => ({
            id: `agent-${tIdx}-${aIdx}`,
            role: 'coder',
        })),
    })),
});

describe('BlueprintDrawer', () => {
    beforeEach(() => {
        useCortexStore.setState({
            isBlueprintDrawerOpen: true,
            savedBlueprints: [],
            blueprint: null,
        });
    });

    it('shows team list from saved blueprints', () => {
        const bp1 = makeMockBlueprint('mission-alpha', 2, 3);
        const bp2 = makeMockBlueprint('mission-beta', 1, 2);

        useCortexStore.setState({
            savedBlueprints: [bp1, bp2],
        });

        render(<BlueprintDrawer />);

        // Both blueprint mission IDs should be visible
        expect(screen.getByText('mission-alpha')).toBeDefined();
        expect(screen.getByText('mission-beta')).toBeDefined();

        // Team counts should be displayed
        expect(screen.getByText('2 teams')).toBeDefined();
        expect(screen.getByText('1 teams')).toBeDefined();
    });

    it('displays agent count per team in blueprint cards', () => {
        const bp = makeMockBlueprint('mission-gamma', 2, 4);
        useCortexStore.setState({
            savedBlueprints: [bp],
        });

        render(<BlueprintDrawer />);

        // 2 teams * 4 agents = 8 agents total
        expect(screen.getByText('8 agents')).toBeDefined();
    });

    it('shows SAVE CURRENT button when a current blueprint exists', () => {
        const currentBp = makeMockBlueprint('mission-active', 1, 2);
        useCortexStore.setState({
            blueprint: currentBp,
            savedBlueprints: [makeMockBlueprint('mission-saved', 1, 1)],
        });

        render(<BlueprintDrawer />);

        // The "SAVE CURRENT" button should be visible when there is a current blueprint
        expect(screen.getByText('SAVE CURRENT')).toBeDefined();
    });
});
