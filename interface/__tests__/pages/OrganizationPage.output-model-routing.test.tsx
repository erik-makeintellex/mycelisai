import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { mockFetch } from "../setup";
import OrganizationPage from "@/app/(app)/organizations/[id]/page";
import {
    jsonResponse,
    resetOrganizationPageStoreState,
    setupOrganizationFetch,
} from "./support/OrganizationPage.testSupport";

vi.mock("next/navigation", () => ({
    useRouter: () => ({ push: vi.fn(), replace: vi.fn(), back: vi.fn(), prefetch: vi.fn() }),
    usePathname: () => "/organizations/org-123",
}));

describe("OrganizationPage output model routing slice", () => {
    beforeEach(() => {
        mockFetch.mockReset();
        resetOrganizationPageStoreState();
    });

    afterEach(() => {
        vi.restoreAllMocks();
        vi.useRealTimers();
    });

    it("lets the operator configure output model routing and shows detected role models", async () => {
        setupOrganizationFetch();

        await act(async () => {
            render(<OrganizationPage params={Promise.resolve({ id: "org-123" })} />);
        });

        expect(await screen.findByRole("heading", { name: "AI Engine Settings" })).toBeDefined();
        fireEvent.click(screen.getAllByRole("button", { name: "Review AI Engine Settings" })[0]);

        expect(await screen.findByText("Output model routing")).toBeDefined();
        expect(screen.getByText("Popular self-hosted starting points")).toBeDefined();
        expect(screen.getByText("Qwen3 8B")).toBeDefined();
        expect(screen.getByText("Llama 3.1 8B")).toBeDefined();
        expect(screen.getByText("Soma model review guardrail")).toBeDefined();
        expect(screen.getByText(/Ask the owner\/admin before Soma reviews potential model behavior/i)).toBeDefined();
        expect(screen.getByText("Behavior review candidates")).toBeDefined();
        expect(screen.getByText("Code generation: Qwen2.5 Coder 7B")).toBeDefined();
        expect(screen.getByText(/prioritize implementation accuracy/i)).toBeDefined();

        fireEvent.click(screen.getByRole("button", { name: /Use detected models by output type/i }));
        fireEvent.change(screen.getByLabelText("Organization default model"), { target: { value: "qwen3:8b" } });

        const selects = screen.getAllByRole("combobox");
        fireEvent.change(selects[2], { target: { value: "llama3.1:8b" } });
        fireEvent.change(selects[3], { target: { value: "qwen2.5-coder:7b" } });
        fireEvent.click(screen.getByRole("button", { name: "Use output model routing" }));

        expect(await screen.findByText("AI Organization Home")).toBeDefined();
        fireEvent.click(screen.getAllByRole("button", { name: "Open Departments" })[0]);
        expect(await screen.findByText("Detected for this role: Llama 3.1 8B")).toBeDefined();
        expect(screen.getByText("Detected for this role: Qwen2.5 Coder 7B")).toBeDefined();
    }, 15000);
});
