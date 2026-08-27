import { beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import MissionControlChat from "@/components/dashboard/MissionControlChat";
import { savePendingSomaOutputContinuation } from "@/components/soma/outputContinuation";
import { useCortexStore } from "@/store/useCortexStore";
import { mockFetch } from "../setup";
import { resetMissionControlChatStore } from "./support/missionControlChatTestUtils";

const realSendMissionChat = useCortexStore.getState().sendMissionChat;

describe("MissionControlChat output-canvas continuation", () => {
  beforeEach(() => {
    localStorage.clear();
    sessionStorage.clear();
    resetMissionControlChatStore();
    useCortexStore.setState({ sendMissionChat: realSendMissionChat });
  });

  it("shows the continuation chip and sends trusted output context with the revision", async () => {
    savePendingSomaOutputContinuation({
      title: "Moonlit Keep First Playable",
      reference: "groups/game-team/generated/package/index.html",
      proof: "proof-moonlit",
      sourceLabel: "playable output",
      teamId: "game-team",
      runId: "run-moonlit",
      workItemId: "work-moonlit",
      outputId: "output-moonlit",
      contentDigest: "sha256:moonlit",
    });
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({
        ok: true,
        data: {
          meta: { source_node: "admin", timestamp: new Date().toISOString() },
          template_id: "chat-to-answer",
          mode: "answer",
          payload: { text: "I can revise that playable output.", tools_used: [] },
        },
      }),
    } as Response);

    render(<MissionControlChat simpleMode />);

    await waitFor(() => expect(screen.getByText("Continuing from")).toBeDefined());
    expect(screen.getByText("Moonlit Keep First Playable")).toBeDefined();
    const composer = screen.getByPlaceholderText(/Tell Soma/i);
    fireEvent.change(composer, { target: { value: "Make the jump less floaty." } });
    fireEvent.keyDown(composer, { key: "Enter" });

    await waitFor(() => expect(mockFetch).toHaveBeenCalled());
    const request = mockFetch.mock.calls.find((call) => String(call[0]).includes("/api/v1/chat"));
    expect(request).toBeDefined();
    const body = JSON.parse(String((request?.[1] as RequestInit).body));
    expect(body.continuation_context).toEqual({
      kind: "output",
      title: "Moonlit Keep First Playable",
      reference: "groups/game-team/generated/package/index.html",
      proof: "proof-moonlit",
      team_id: "game-team",
      run_id: "run-moonlit",
      work_item_id: "work-moonlit",
      output_id: "output-moonlit",
      content_digest: "sha256:moonlit",
    });
  });
});
