import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";

import RouteTemplatePicker from "@/components/automations/RouteTemplatePicker";
import { TEAM_PROFILE_TEMPLATES } from "@/lib/workflowContracts";

describe("RouteTemplatePicker", () => {
    it("switches from basic to guided mode and applies a template", () => {
        const onRoutesChange = vi.fn();
        const profile = TEAM_PROFILE_TEMPLATES[0];

        render(<RouteTemplatePicker profile={profile} onRoutesChange={onRoutesChange} />);

        fireEvent.click(screen.getByRole("button", { name: /Guided/i }));
        fireEvent.click(screen.getByRole("button", { name: /Apply Template/i }));

        expect(screen.getByText(/Impact preview/i)).toBeDefined();
        expect(onRoutesChange).toHaveBeenCalled();
    });

    it("supports expert mode and rollback", () => {
        const profile = TEAM_PROFILE_TEMPLATES[1];
        render(<RouteTemplatePicker profile={profile} />);

        fireEvent.click(screen.getByRole("button", { name: /Expert/i }));
        const textarea = screen.getByPlaceholderText("One subject per line");
        fireEvent.change(textarea, { target: { value: "swarm.custom.*\nswarm.health.*" } });

        expect(screen.getByText("swarm.custom.*")).toBeDefined();
        fireEvent.click(screen.getByRole("button", { name: /^Rollback$/i }));
        expect(screen.queryByText("swarm.custom.*")).toBeNull();
    });
});
