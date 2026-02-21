import { describe, it, expect } from 'vitest';
import {
    toolLabel,
    councilLabel,
    councilOptionLabel,
    sourceNodeLabel,
    trustBadge,
    trustTooltip,
    deriveIntentClass,
    isGovernanceTool,
    isInternalTool,
    toolDescription,
    TOOL_LABELS,
    COUNCIL_LABELS,
    GOV_LABELS,
    WORKSPACE_LABELS,
    MODE_LABELS,
    GOV_POSTURE_LABELS,
    GOVERNANCE_TOOLS,
    TOOL_DESCRIPTIONS,
    MEMORY_LABELS,
} from '@/lib/labels';

describe('labels', () => {
    describe('toolLabel', () => {
        it('maps known internal names to human labels', () => {
            expect(toolLabel('consult_council')).toBe('Ask Specialist');
            expect(toolLabel('delegate_task')).toBe('Assign to Team');
            expect(toolLabel('search_memory')).toBe('Search Memory');
            expect(toolLabel('list_teams')).toBe('View Teams');
            expect(toolLabel('generate_blueprint')).toBe('Design Mission');
            expect(toolLabel('summarize_conversation')).toBe('Save Context');
        });

        it('returns raw name for unknown tools', () => {
            expect(toolLabel('unknown_tool')).toBe('unknown_tool');
            expect(toolLabel('custom_mcp_tool')).toBe('custom_mcp_tool');
        });

        it('covers all 20 registered internal tools', () => {
            expect(Object.keys(TOOL_LABELS)).toHaveLength(20);
        });
    });

    describe('councilLabel', () => {
        it('returns name and subtitle for known council members', () => {
            expect(councilLabel('admin')).toEqual({ name: 'Soma', subtitle: 'Executive Cortex' });
            expect(councilLabel('council-architect')).toEqual({ name: 'Architect', subtitle: 'Systems Design' });
            expect(councilLabel('council-coder')).toEqual({ name: 'Coder', subtitle: 'Implementation' });
            expect(councilLabel('council-creative')).toEqual({ name: 'Creative', subtitle: 'Design & Ideation' });
            expect(councilLabel('council-sentry')).toEqual({ name: 'Sentry', subtitle: 'Risk & Security' });
        });

        it('returns fallback for unknown IDs', () => {
            expect(councilLabel('unknown')).toEqual({ name: 'unknown', subtitle: '' });
        });

        it('covers all 5 council members', () => {
            expect(Object.keys(COUNCIL_LABELS)).toHaveLength(5);
        });
    });

    describe('councilOptionLabel', () => {
        it('formats known members as "Name — Subtitle"', () => {
            expect(councilOptionLabel('admin', 'admin')).toBe('Soma — Executive Cortex');
            expect(councilOptionLabel('council-architect', 'architect')).toBe('Architect — Systems Design');
        });

        it('capitalizes role for unknown members', () => {
            expect(councilOptionLabel('custom-id', 'analyst')).toBe('Analyst');
        });
    });

    describe('sourceNodeLabel', () => {
        it('returns human name for known source nodes', () => {
            expect(sourceNodeLabel('admin')).toBe('Soma');
            expect(sourceNodeLabel('council-architect')).toBe('Architect');
            expect(sourceNodeLabel('council-sentry')).toBe('Sentry');
        });

        it('strips "council-" prefix and capitalizes for unknown nodes', () => {
            expect(sourceNodeLabel('council-analyst')).toBe('Analyst');
        });

        it('capitalizes raw name when no prefix', () => {
            expect(sourceNodeLabel('observer')).toBe('Observer');
        });
    });

    describe('trustBadge', () => {
        it('formats as C:{score}', () => {
            expect(trustBadge(0.5)).toBe('C:0.5');
            expect(trustBadge(1.0)).toBe('C:1.0');
            expect(trustBadge(0.0)).toBe('C:0.0');
        });
    });

    describe('trustTooltip', () => {
        it('shows high confidence for scores >= 0.9', () => {
            expect(trustTooltip(0.95)).toBe('Confidence: 0.95 — High confidence');
            expect(trustTooltip(1.0)).toBe('Confidence: 1.00 — High confidence');
        });

        it('shows moderate confidence for scores 0.6-0.89', () => {
            expect(trustTooltip(0.7)).toBe('Confidence: 0.70 — Moderate confidence');
        });

        it('shows low confidence for scores < 0.6', () => {
            expect(trustTooltip(0.3)).toBe('Confidence: 0.30 — Low confidence');
            expect(trustTooltip(0.5)).toBe('Confidence: 0.50 — Low confidence');
        });
    });

    describe('GOV_LABELS', () => {
        it('contains all governance labels', () => {
            expect(GOV_LABELS.governanceReview).toBe('Review Required');
            expect(GOV_LABELS.agentOutput).toBe('Agent Output');
            expect(GOV_LABELS.proofOfWork).toBe('Verification Evidence');
            expect(GOV_LABELS.verificationMethod).toBe('Verification Method');
            expect(GOV_LABELS.rubricScore).toBe('Quality Score');
            expect(GOV_LABELS.noProof).toBe('No verification evidence provided');
            expect(GOV_LABELS.noProofSub).toBe('This agent did not submit verification evidence');
        });
    });

    describe('WORKSPACE_LABELS', () => {
        it('contains all workspace labels', () => {
            expect(WORKSPACE_LABELS.spectrum).toBe('Spectrum');
            expect(WORKSPACE_LABELS.squadRoom).toBe('Squad Room');
            expect(WORKSPACE_LABELS.internalDebate).toBe('Internal Deliberation');
            expect(WORKSPACE_LABELS.metaArchitect).toBe('Mission Architect');
            expect(WORKSPACE_LABELS.toolRegistry).toBe('Capabilities');
            expect(WORKSPACE_LABELS.internalTools).toBe('Core Capabilities');
        });
    });

    // ── CE-1: Orchestration Template labels ────────────────────────

    describe('MODE_LABELS', () => {
        it('has all four modes', () => {
            expect(Object.keys(MODE_LABELS)).toHaveLength(4);
            expect(MODE_LABELS.answer.label).toBe('ANSWER');
            expect(MODE_LABELS.proposal.label).toBe('PROPOSAL');
            expect(MODE_LABELS.broadcast.label).toBe('BROADCAST');
            expect(MODE_LABELS.execute.label).toBe('EXECUTE');
        });

        it('has colors for each mode', () => {
            expect(MODE_LABELS.answer.color).toContain('cortex-primary');
            expect(MODE_LABELS.proposal.color).toContain('amber');
        });
    });

    describe('GOV_POSTURE_LABELS', () => {
        it('has all three postures', () => {
            expect(Object.keys(GOV_POSTURE_LABELS)).toHaveLength(3);
            expect(GOV_POSTURE_LABELS.passive.label).toBe('PASSIVE');
            expect(GOV_POSTURE_LABELS.active.label).toBe('ACTIVE');
            expect(GOV_POSTURE_LABELS.strict.label).toBe('STRICT');
        });
    });

    describe('deriveIntentClass', () => {
        it('returns "Direct Answer" for empty/null tools', () => {
            expect(deriveIntentClass([])).toBe('Direct Answer');
            expect(deriveIntentClass(null as any)).toBe('Direct Answer');
            expect(deriveIntentClass(undefined as any)).toBe('Direct Answer');
        });

        it('classifies blueprint tools as "Mission Design"', () => {
            expect(deriveIntentClass(['generate_blueprint'])).toBe('Mission Design');
            expect(deriveIntentClass(['research_for_blueprint'])).toBe('Mission Design');
            expect(deriveIntentClass(['research_for_blueprint', 'generate_blueprint'])).toBe('Mission Design');
        });

        it('classifies delegate_task as "Delegation"', () => {
            expect(deriveIntentClass(['delegate_task'])).toBe('Delegation');
        });

        it('classifies consult_council as "Consultation"', () => {
            expect(deriveIntentClass(['consult_council'])).toBe('Consultation');
        });

        it('classifies memory tools as "Memory Recall"', () => {
            expect(deriveIntentClass(['search_memory'])).toBe('Memory Recall');
            expect(deriveIntentClass(['recall'])).toBe('Memory Recall');
        });

        it('classifies file tools as "File Operation"', () => {
            expect(deriveIntentClass(['read_file'])).toBe('File Operation');
            expect(deriveIntentClass(['write_file'])).toBe('File Operation');
        });

        it('returns "Direct Answer" for unclassified tools', () => {
            expect(deriveIntentClass(['list_teams'])).toBe('Direct Answer');
            expect(deriveIntentClass(['get_system_status'])).toBe('Direct Answer');
        });
    });

    describe('isGovernanceTool', () => {
        it('returns true for governance-gated tools', () => {
            expect(isGovernanceTool('delegate_task')).toBe(true);
            expect(isGovernanceTool('generate_blueprint')).toBe(true);
            expect(isGovernanceTool('write_file')).toBe(true);
            expect(isGovernanceTool('publish_signal')).toBe(true);
        });

        it('returns false for non-gated tools', () => {
            expect(isGovernanceTool('search_memory')).toBe(false);
            expect(isGovernanceTool('list_teams')).toBe(false);
            expect(isGovernanceTool('recall')).toBe(false);
        });

        it('has exactly 4 governance tools', () => {
            expect(GOVERNANCE_TOOLS.size).toBe(4);
        });
    });

    describe('isInternalTool', () => {
        it('returns true for registered internal tools', () => {
            expect(isInternalTool('consult_council')).toBe(true);
            expect(isInternalTool('delegate_task')).toBe(true);
            expect(isInternalTool('summarize_conversation')).toBe(true);
        });

        it('returns false for unknown tools', () => {
            expect(isInternalTool('custom_mcp_tool')).toBe(false);
            expect(isInternalTool('nonexistent')).toBe(false);
        });
    });

    describe('TOOL_DESCRIPTIONS', () => {
        it('has descriptions for all internal tools', () => {
            const toolKeys = Object.keys(TOOL_LABELS);
            const descKeys = Object.keys(TOOL_DESCRIPTIONS);
            // Every tool should have a description
            for (const key of toolKeys) {
                expect(TOOL_DESCRIPTIONS[key]).toBeDefined();
            }
            // At least as many descriptions as tools
            expect(descKeys.length).toBeGreaterThanOrEqual(toolKeys.length);
        });
    });

    describe('toolDescription', () => {
        it('returns description for known tools', () => {
            expect(toolDescription('delegate_task')).toContain('governance-gated');
            expect(toolDescription('search_memory')).toContain('Semantic search');
        });

        it('returns raw name for unknown tools', () => {
            expect(toolDescription('unknown_tool')).toBe('unknown_tool');
        });
    });

    describe('MEMORY_LABELS', () => {
        it('has labels for memory tools', () => {
            expect(MEMORY_LABELS.search_memory).toBe('Semantic search');
            expect(MEMORY_LABELS.recall).toBe('Context recall');
        });
    });
});
