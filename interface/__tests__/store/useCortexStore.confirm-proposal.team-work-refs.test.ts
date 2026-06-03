import { beforeEach, describe, expect, it } from "vitest";
import { useCortexStore } from "@/store/useCortexStore";
import { mockFetch } from "../setup";
import { resetCortexStore } from "./useCortexStoreTestSupport";

describe("useCortexStore confirm proposal team work refs", () => {
    beforeEach(() => {
        resetCortexStore();
    });

    it("adds a concise Active Work cue when successful confirmation returns team work refs", async () => {
        useCortexStore.setState({
            pendingProposal: {
                intent: "Launch a build crew",
                teams: 1,
                agents: 2,
                tools: ["delegate_task"],
                risk_level: "medium",
                confirm_token: "ct-work",
                intent_proof_id: "ip-work",
            },
            activeConfirmToken: "ct-work",
            missionChat: [{
                role: "council",
                content: "Proposed execution path",
                mode: "proposal",
                proposal: {
                    intent: "Launch a build crew",
                    teams: 1,
                    agents: 2,
                    tools: ["delegate_task"],
                    risk_level: "medium",
                    confirm_token: "ct-work",
                    intent_proof_id: "ip-work",
                },
                proposal_status: "active",
            }],
            missionChatError: null,
            activeMode: "proposal",
            activeRunId: null,
        });
        mockFetch.mockResolvedValueOnce({
            ok: true,
            json: async () => ({
                data: {
                    run_id: "run-work-123456",
                    team_work_refs: [{ work_item_id: "work-alpha-123456", state: "running", run_id: "run-work-123456" }],
                },
            }),
        }).mockResolvedValueOnce({ ok: true, json: async () => ([]) });

        await useCortexStore.getState().confirmProposal();

        const message = useCortexStore.getState().missionChat.at(-1);
        expect(message).toMatchObject({ role: "system", mode: "execution_result", run_id: "run-work-123456" });
        expect(message?.content).toContain("Run run-work started.");
        expect(message?.content).toContain("Work work-alp is running.");
        expect(message?.content).toContain("Review Active Work and the latest output.");
        expect(message?.content).not.toContain("work-alpha-123456");
    });
});
