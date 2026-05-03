import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";

import LandingPage from "@/app/(marketing)/page";

describe("Landing Page (V8.2 Soma-primary AI Organization workspace)", () => {
    it("renders the living AI Organization framing with the right CTA hierarchy", () => {
        render(<LandingPage />);

        expect(
            screen.getByRole("heading", { name: "What do you want Soma to do?" }),
        ).toBeDefined();
        expect(screen.getByText(/Start with an intent/i)).toBeDefined();
        expect(screen.getAllByRole("link", { name: /Ask Soma/i })[0].getAttribute("href")).toBe("/dashboard");
        expect(screen.getByRole("link", { name: "Create AI Organization" }).getAttribute("href")).toBe("/dashboard#dashboard-organization-setup");
    });

    it("renders intent cards and support-system guidance", () => {
        render(<LandingPage />);

        expect(screen.getByRole("heading", { name: "Choose the shape of work." })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Plan" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Research" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Create" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Review" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Configure tools" })).toBeDefined();
        expect(screen.getByText("Advanced stays available")).toBeDefined();
    });

    it("keeps internal architecture terms and chatbot framing out of the landing page", () => {
        render(<LandingPage />);

        expect(screen.queryByText(/Loop Profile/i)).toBeNull();
        expect(screen.queryByText(/Soma Kernel/i)).toBeNull();
        expect(screen.queryByText(/Central Council/i)).toBeNull();
        expect(screen.queryByText(/Provider Policy/i)).toBeNull();
        expect(screen.queryByText(/Recursive Swarm/i)).toBeNull();
        expect(screen.queryByText(/Deploy Cortex/i)).toBeNull();
        expect(screen.queryByText(/Launch Console/i)).toBeNull();
        expect(screen.queryByText(/chatbot/i)).toBeNull();
    });
});
