import { vi } from "vitest";

export const mockTeams = [
  {
    id: "team-alpha",
    name: "Alpha Squad",
    role: "action",
    type: "standing" as const,
    mission_id: null,
    mission_intent: null,
    inputs: ["nats.input.alpha"],
    deliveries: ["nats.output.alpha"],
    agents: [
      {
        id: "agent-1",
        role: "cognitive",
        status: 1,
        last_heartbeat: new Date().toISOString(),
        tools: [],
        model: "qwen",
      },
      {
        id: "agent-2",
        role: "sensory",
        status: 0,
        last_heartbeat: new Date().toISOString(),
        tools: [],
        model: "qwen",
      },
    ],
  },
  {
    id: "team-bravo",
    name: "Bravo Ops",
    role: "expression",
    type: "mission" as const,
    mission_id: "mission-001",
    mission_intent: "Deploy sentinel network",
    inputs: [],
    deliveries: [],
    agents: [
      {
        id: "agent-3",
        role: "actuation",
        status: 2,
        last_heartbeat: new Date().toISOString(),
        tools: ["exec"],
        model: "llama",
      },
    ],
  },
];

export const mockTemplates = [
  {
    id: "template-marketing-writer",
    name: "Marketing Writer",
    role: "cognitive",
    system_prompt: "Write and refine launch copy.",
    model: "qwen3:8b",
    tools: ["recall"],
    inputs: ["briefs"],
    outputs: ["campaign copy", "launch messaging"],
    verification_strategy: "semantic",
    verification_rubric: ["clear", "on-brand"],
    validation_command: "",
    created_at: new Date("2026-04-07T10:00:00Z").toISOString(),
    updated_at: new Date("2026-04-07T12:00:00Z").toISOString(),
  },
  {
    id: "template-researcher",
    name: "Audience Researcher",
    role: "cognitive",
    system_prompt: "Research campaigns and audience insight.",
    model: "llama3.1:8b",
    tools: ["fetch"],
    inputs: ["requests"],
    outputs: ["research briefs"],
    verification_strategy: "semantic",
    verification_rubric: ["grounded"],
    validation_command: "",
    created_at: new Date("2026-04-07T09:00:00Z").toISOString(),
    updated_at: new Date("2026-04-07T11:00:00Z").toISOString(),
  },
];

export function mockTeamWorkFetch(mockFetch: ReturnType<typeof vi.fn>) {
  mockFetch.mockImplementation(async (url: string) => {
    if (url.includes("/api/v1/teams/team-alpha/work")) {
      return {
        ok: true,
        json: async () => ({
          data: [
            {
              work_item_id: "work-alpha",
              team_id: "team-alpha",
              run_id: "run-alpha",
              objective: "Draft launch package",
              execution_shape: "deliverable",
              state: "output_ready",
              output_refs: [
                {
                  output_id: "out-alpha",
                  team_id: "team-alpha",
                  work_item_id: "work-alpha",
                  kind: "file",
                  label: "Launch brief",
                  storage_ref: "generated/alpha/brief.md",
                  proof_ref: "proof-alpha",
                },
              ],
              proof_refs: ["proof-alpha"],
              audit_refs: ["audit-alpha"],
              updated_at: "2026-05-19T12:00:00Z",
            },
          ],
        }),
      };
    }
    if (url.includes("/api/v1/teams/team-bravo/work")) {
      return {
        ok: true,
        json: async () => ({
          data: [
            {
              work_item_id: "work-bravo",
              team_id: "team-bravo",
              objective: "Review deployment proof",
              execution_shape: "delegated_work",
              state: "running",
              updated_at: "2026-05-19T12:01:00Z",
            },
          ],
        }),
      };
    }
    return { ok: true, json: async () => ({ data: [] }) };
  });
}
