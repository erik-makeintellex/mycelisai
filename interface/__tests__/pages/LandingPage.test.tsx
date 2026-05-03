import { beforeEach, describe, expect, it } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { mockFetch } from "../setup";
import LandingPage from "@/app/(marketing)/page";

describe("Landing Page homepage template", () => {
    beforeEach(() => {
        localStorage.clear();
        mockFetch.mockResolvedValue({ ok: false, json: async () => ({}) });
    });

    it("renders the default Soma orchestration product framing", () => {
        render(<LandingPage />);

        expect(screen.getByRole("heading", { name: "Operate AI Organizations through Soma" })).toBeDefined();
        expect(screen.getByText(/self-hostable operating surface for ai work/i)).toBeDefined();
        expect(screen.getByText(/Example workspace preview/i)).toBeDefined();
        expect(screen.getAllByRole("link", { name: /Start with Soma/i })[0].getAttribute("href")).toBe("/dashboard");
        expect(screen.getAllByRole("link", { name: /View Documentation/i })[0].getAttribute("href")).toBe("/docs");
    });

    it("renders custom product name, hero copy, and links from safe homepage config", async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({
                ok: true,
                data: {
                    enabled: true,
                    brand: { product_name: "Acme AI", tagline: "Internal governed orchestration" },
                    hero: {
                        headline: "Coordinate work through Soma",
                        subheadline: "Use Soma to coordinate structured teams and connected tools.",
                        primary_cta: { label: "Open Soma", href: "/dashboard" },
                        secondary_cta: { label: "Support", href: "https://support.example.com", external: true },
                    },
                    sections: [{ title: "Express intent", body: "Tell Soma what needs to happen." }],
                    links: [{ label: "Support", href: "https://support.example.com", description: "Contact the platform team.", external: true }],
                    footer_text: "Internal deployment",
                },
            }),
        });

        render(<LandingPage />);

        await waitFor(() => {
            expect(screen.getByText("Acme AI")).toBeDefined();
            expect(screen.getByRole("heading", { name: "Coordinate work through Soma" })).toBeDefined();
        });
        const supportLinks = screen.getAllByRole("link", { name: /Support/i });
        expect(supportLinks.some((link) => link.getAttribute("href") === "https://support.example.com")).toBe(true);
        expect(screen.getByText("Internal deployment")).toBeDefined();
    });

    it("keeps fake live-state and internal architecture language out of the landing page", () => {
        render(<LandingPage />);

        expect(screen.queryByText(/live user state/i)).toBeNull();
        expect(screen.queryByText(/currently running in your workspace/i)).toBeNull();
        expect(screen.queryByText(/Soma Kernel/i)).toBeNull();
        expect(screen.queryByText(/Central Council/i)).toBeNull();
        expect(screen.queryByText(/Recursive Swarm/i)).toBeNull();
        expect(screen.queryByText(/sentient/i)).toBeNull();
        expect(screen.queryByText(/chatbot/i)).toBeNull();
    });
});
