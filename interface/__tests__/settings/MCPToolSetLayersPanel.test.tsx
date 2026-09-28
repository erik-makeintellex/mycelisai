import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MCPToolSetLayersPanel } from "@/components/settings/MCPToolSetLayersPanel";
import type { MCPToolSet } from "@/store/useCortexStore";
import type { ComponentProps } from "react";

const toolSets: MCPToolSet[] = [
    {
        id: "set-all",
        name: "workspace",
        description: "Shared workspace file tools.",
        tool_refs: ["mcp:filesystem/*"],
        scope_kind: "all",
    },
    {
        id: "set-group",
        name: "research",
        tool_refs: ["mcp:fetch/fetch"],
        scope_kind: "group",
        scope_ref: "alpha-lane",
    },
    {
        id: "set-host",
        name: "deploy",
        tool_refs: ["mcp:ssh/*"],
        scope_kind: "host",
        scope_ref: "edge-node-1",
    },
];

function renderPanel(overrides: Partial<ComponentProps<typeof MCPToolSetLayersPanel>> = {}) {
    const onRefresh = vi.fn();
    const onCreate = vi.fn().mockResolvedValue({ ok: true });
    render(
        <MCPToolSetLayersPanel
            toolSets={toolSets}
            isLoading={false}
            error={null}
            onRefresh={onRefresh}
            onCreate={onCreate}
            isAdmin
            {...overrides}
        />,
    );
    return { onCreate, onRefresh };
}

describe("MCPToolSetLayersPanel", () => {
    it("shows shared, group, and host permission groups", () => {
        renderPanel();

        expect(screen.getByText("Capability permissions")).toBeDefined();
        expect(screen.getByText("Choose where Soma can use connected tools")).toBeDefined();
        expect(screen.getByText("workspace")).toBeDefined();
        expect(screen.getByText("Group: alpha-lane")).toBeDefined();
        expect(screen.getByText("Host: edge-node-1")).toBeDefined();
        expect(screen.getAllByText(/workspace default/i).length).toBeGreaterThan(0);
        expect(screen.getByText(/only inside the selected group/i)).toBeDefined();
        expect(screen.getByText("mcp:filesystem/*")).toBeDefined();
    });

    it("requires a target before saving group layers", async () => {
        const { onCreate } = renderPanel();

        fireEvent.change(screen.getByLabelText(/Name/i), { target: { value: "lane-tools" } });
        fireEvent.click(screen.getByRole("button", { name: "Group" }));
        fireEvent.change(screen.getByLabelText(/Capability refs/i), { target: { value: "mcp:filesystem/*" } });
        fireEvent.click(screen.getByRole("button", { name: "Save permissions" }));

        expect(await screen.findByText("Group permissions need a target.")).toBeDefined();
        expect(onCreate).not.toHaveBeenCalled();
    });

    it("saves a host-targeted layer with parsed refs", async () => {
        const { onCreate } = renderPanel();

        fireEvent.change(screen.getByLabelText(/Name/i), { target: { value: "deploy" } });
        fireEvent.click(screen.getByRole("button", { name: "Host" }));
        fireEvent.change(screen.getByLabelText(/Target Host/i), { target: { value: "edge-node-2" } });
        fireEvent.change(screen.getByLabelText(/Capability refs/i), {
            target: { value: "mcp:ssh/*\ntoolset:workspace" },
        });
        fireEvent.change(screen.getByLabelText(/Description/i), { target: { value: "Deploy tools" } });
        fireEvent.click(screen.getByRole("button", { name: "Save permissions" }));

        await waitFor(() => expect(onCreate).toHaveBeenCalledTimes(1));
        expect(onCreate).toHaveBeenCalledWith({
            name: "deploy",
            description: "Deploy tools",
            tool_refs: ["mcp:ssh/*", "toolset:workspace"],
            scope_kind: "host",
            scope_ref: "edge-node-2",
        });
        expect(await screen.findByText("Permission group saved and refreshed.")).toBeDefined();
    });

    it("lets users choose common capabilities without typing raw refs first", async () => {
        const { onCreate } = renderPanel();

        fireEvent.change(screen.getByLabelText(/Name/i), { target: { value: "research lane" } });
        fireEvent.click(screen.getByRole("button", { name: /Web research/i }));
        fireEvent.change(screen.getByLabelText(/Target Group/i), { target: { value: "market-research" } });
        fireEvent.click(screen.getByRole("button", { name: "Save permissions" }));

        await waitFor(() => expect(onCreate).toHaveBeenCalledTimes(1));
        expect(onCreate).toHaveBeenCalledWith({
            name: "research lane",
            description: undefined,
            tool_refs: ["mcp:fetch/fetch", "tool:web_search"],
            scope_kind: "group",
            scope_ref: "market-research",
        });
    });

    // MCPA D10: new-layer is an admin-only control. isAdmin is a UI hint
    // only (Core still enforces mcp_config:write).
    it("hides the add-permission-group form for a non-admin viewer", () => {
        renderPanel({ isAdmin: false });

        expect(screen.queryByRole("button", { name: "Save permissions" })).toBeNull();
        expect(screen.getByText(/Only an admin can add or change permission groups/i)).toBeDefined();
        // Reads stay visible for everyone (D3).
        expect(screen.getByText("workspace")).toBeDefined();
    });

    // MCPA D10: honest failure -- a rejected save renders blocker copy
    // instead of the generic "could not be saved" text.
    it("renders blocker copy on a rejected save", async () => {
        const onCreate = vi.fn().mockResolvedValue({ ok: false, code: "admin_required", httpStatus: 403 });
        renderPanel({ onCreate });

        fireEvent.change(screen.getByLabelText(/Name/i), { target: { value: "lane-tools" } });
        fireEvent.change(screen.getByLabelText(/Capability refs/i), { target: { value: "mcp:filesystem/*" } });
        fireEvent.click(screen.getByRole("button", { name: "Save permissions" }));

        expect(await screen.findByRole("alert")).toBeDefined();
        expect(screen.queryByText(/"code":"admin_required"/)).toBeNull();
    });
});
