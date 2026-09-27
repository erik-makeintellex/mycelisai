import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import MissionControlMessageBubble from "@/components/dashboard/MissionControlMessageBubble";
import type { ChatMessage } from "@/store/useCortexStore";

const reply: ChatMessage = {
    role: "council",
    content: "Blueberry-lavender scones, 3 for $10 this weekend.\nSee you at the counter!",
    context_sources: [
        { artifact_id: "a-1", title: "Juniper & Rye Bakery", knowledge_class: "company_knowledge", retrieval_mode: "keyword", used: true },
        { artifact_id: "a-2", title: "Opening Hours", knowledge_class: "company_knowledge", retrieval_mode: "keyword", used: false },
    ],
};

describe("MissionControlMessageBubble context sources", () => {
    it("shows a Sources line that separates used from consulted sources", () => {
        render(<MissionControlMessageBubble msg={reply} />);
        const line = screen.getByLabelText("Context sources");
        expect(line.textContent).toContain("Sources:");
        expect(line.textContent).toContain("Juniper & Rye Bakery (Used)");
        expect(line.textContent).toContain("Opening Hours (Consulted)");
    });

    it("shows no Sources line when no context was injected", () => {
        render(<MissionControlMessageBubble msg={{ role: "council", content: "Hello." }} />);
        expect(screen.queryByLabelText("Context sources")).toBeNull();
    });
});
