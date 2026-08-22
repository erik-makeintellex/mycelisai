import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import MissionControlMarkdown from "@/components/dashboard/MissionControlMarkdown";

describe("MissionControlMarkdown", () => {
    it("renders internal Mycelis links without external-link treatment", () => {
        render(<MissionControlMarkdown content={"Open [System Status](/system)."} />);

        const link = screen.getByRole("link", { name: "System Status" });
        expect(link.getAttribute("href")).toBe("/system");
        expect(link.getAttribute("target")).toBeNull();
        expect(link.getAttribute("rel")).toBeNull();
    });

    it("keeps external links protected", () => {
        render(<MissionControlMarkdown content={"Read [the docs](https://example.test/docs)."} />);

        const link = screen.getByRole("link", { name: /the docs/i });
        expect(link.getAttribute("href")).toBe("https://example.test/docs");
        expect(link.getAttribute("target")).toBe("_blank");
        expect(link.getAttribute("rel")).toBe("noopener noreferrer");
    });
});
