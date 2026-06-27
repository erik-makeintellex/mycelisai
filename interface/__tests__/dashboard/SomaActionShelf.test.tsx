import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { SomaActionShelf } from "@/components/soma/SomaActionShelf";

describe("SomaActionShelf", () => {
  beforeEach(() => {
    window.localStorage.clear();
    vi.restoreAllMocks();
  });

  it("saves a reusable quick action without sending a chat prompt immediately", async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(apiResponse([]))
      .mockResolvedValueOnce(apiResponse({
        id: "template-client-risk",
        name: "Client risk brief",
        template_body: "Quick action: Client risk brief. Outcome: Create a retained brief.",
        governance_tags: ["quick_action", "button_studio"],
        output_contract: {
          output_format: "Markdown",
          approval_behavior: "Ask before running",
        },
      }));
    vi.stubGlobal("fetch", fetchMock);
    const onRunAction = vi.fn();
    render(<SomaActionShelf onRunAction={onRunAction} />);

    fireEvent.click(screen.getByRole("button", { name: /Create new quick action/i }));
    fireEvent.change(screen.getByLabelText("Button label"), { target: { value: "Client risk brief" } });
    fireEvent.change(screen.getByLabelText("Outcome"), { target: { value: "Create a retained brief with risks and next steps" } });
    fireEvent.change(screen.getByLabelText("Output format"), { target: { value: "Markdown" } });
    fireEvent.click(screen.getByRole("button", { name: /Save action/i }));

    expect(onRunAction).not.toHaveBeenCalled();
    await waitFor(() => expect(screen.getByRole("button", { name: /Client risk brief/i })).toBeDefined());
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/conversation-templates", expect.objectContaining({
      method: "POST",
    }));
    expect(fetchMock.mock.calls[1]?.[1]?.body).toContain("quick_action");
  });

  it("runs saved backend actions as natural Soma requests", async () => {
    vi.stubGlobal("fetch", vi.fn()
      .mockResolvedValueOnce(apiResponse([
        {
          id: "template-client-risk",
          name: "Client risk brief",
          template_body: "Quick action: Client risk brief. Outcome: Create a retained brief.",
          governance_tags: ["quick_action", "button_studio"],
        },
      ]))
      .mockResolvedValueOnce(apiResponse({
        rendered_prompt: "Rendered client risk brief prompt.",
      })));
    const onRunAction = vi.fn();

    render(<SomaActionShelf onRunAction={onRunAction} />);
    await waitFor(() => expect(screen.getByRole("button", { name: /Client risk brief/i })).toBeDefined());
    fireEvent.click(screen.getByRole("button", { name: /Client risk brief/i }));

    await waitFor(() => expect(onRunAction).toHaveBeenCalledWith("Rendered client risk brief prompt."));
  });

  it("keeps local fallback actions usable when backend is unavailable", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("offline")));
    window.localStorage.setItem("mycelis-soma-saved-actions", JSON.stringify([
      {
        label: "Client risk brief",
        prompt: "Quick action: Client risk brief. Outcome: Create a retained brief.",
        userSaved: true,
      },
    ]));
    const onRunAction = vi.fn();

    render(<SomaActionShelf onRunAction={onRunAction} />);
    fireEvent.click(await screen.findByRole("button", { name: /Client risk brief/i }));

    expect(onRunAction).toHaveBeenCalledWith(expect.stringContaining("Quick action: Client risk brief"));
  });
});

function apiResponse(data: unknown) {
  return {
    ok: true,
    json: async () => ({ ok: true, data }),
  } as Response;
}
