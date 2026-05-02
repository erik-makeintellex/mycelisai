import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";

import LandingPage from "@/app/(marketing)/page";

describe("Landing Page (V8.2 Soma-primary AI Organization workspace)", () => {
    it("renders the living AI Organization framing with the right CTA hierarchy", () => {
        render(<LandingPage />);

        expect(
            screen.getByRole("heading", { name: "Build AI Organizations that think, review, and evolve." }),
        ).toBeDefined();
        expect(
            screen.getByText(
                /Mycelis opens with Soma at the center of a governed workspace/i,
            ),
        ).toBeDefined();
        expect(screen.getAllByRole("link", { name: "Create AI Organization" })[0]?.getAttribute("href")).toBe("/dashboard");
        expect(screen.getAllByRole("link", { name: "Review Docs" })[0]?.getAttribute("href")).toBe("/docs");
    });

    it("renders the three truths and the post-creation workspace guidance", () => {
        render(<LandingPage />);

        expect(screen.getByRole("heading", { name: "Structure" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Control" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "Continuous Operation" })).toBeDefined();
        expect(screen.getByRole("heading", { name: "The first screen after creation is a real operating workspace." })).toBeDefined();
        expect(screen.getByText("Enter the Soma workspace")).toBeDefined();
        expect(screen.getByText("See Advisors and Departments")).toBeDefined();
        expect(screen.getByText("Watch Recent Activity")).toBeDefined();
        expect(screen.getByText("Follow guided next steps")).toBeDefined();
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
