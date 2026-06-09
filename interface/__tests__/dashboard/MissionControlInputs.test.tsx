import { createRef } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { MissionControlAdvancedInput } from "@/components/dashboard/MissionControlAdvancedInput";
import { SomaIntentInput } from "@/components/soma/SomaIntentInput";

describe("Soma input composers", () => {
  it("keeps multiline text in the advanced composer and only submits on bare Enter", () => {
    const onChange = vi.fn();
    const onSubmit = vi.fn();
    const inputRef = createRef<HTMLTextAreaElement>();

    render(
      <MissionControlAdvancedInput
        value={"Line one\nLine two"}
        broadcastMode={false}
        isLoading={false}
        placeholder="Tell Soma what to do"
        inputRef={inputRef}
        onChange={onChange}
        onSubmit={onSubmit}
      />,
    );

    const textarea = screen.getByPlaceholderText("Tell Soma what to do");
    expect(textarea.tagName).toBe("TEXTAREA");
    expect((textarea as HTMLTextAreaElement).value).toBe("Line one\nLine two");
    expect(textarea.className).toContain("max-h-[180px]");
    expect(textarea.className).toContain("overflow-y-auto");

    fireEvent.keyDown(textarea, { key: "Enter", shiftKey: true });
    expect(onSubmit).not.toHaveBeenCalled();

    fireEvent.keyDown(textarea, { key: "Enter" });
    expect(onSubmit).toHaveBeenCalledTimes(1);
  });

  it("uses the same multiline behavior in the simplified Soma composer", () => {
    const onChange = vi.fn();
    const onSubmit = vi.fn();
    const inputRef = createRef<HTMLTextAreaElement>();

    render(
      <SomaIntentInput
        value={"Create a team\nThen generate proof"}
        placeholder="Ask Soma"
        inputRef={inputRef}
        onChange={onChange}
        onSubmit={onSubmit}
      />,
    );

    const textarea = screen.getByPlaceholderText("Ask Soma");
    expect(textarea.tagName).toBe("TEXTAREA");
    expect((textarea as HTMLTextAreaElement).value).toBe("Create a team\nThen generate proof");
    expect(textarea.className).toContain("max-h-[180px]");
    expect(textarea.className).toContain("overflow-y-auto");

    fireEvent.keyDown(textarea, { key: "Enter", shiftKey: true });
    expect(onSubmit).not.toHaveBeenCalled();

    fireEvent.keyDown(textarea, { key: "Enter" });
    expect(onSubmit).toHaveBeenCalledTimes(1);
  });
});
