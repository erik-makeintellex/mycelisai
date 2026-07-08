import { createRef, useRef, useState } from "react";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { MissionControlAdvancedInput } from "@/components/dashboard/MissionControlAdvancedInput";
import { SomaIntentInput } from "@/components/soma/SomaIntentInput";
import { requestSomaOutputContinuation, useSomaOutputContinuation } from "@/components/soma/outputContinuation";
import { SOMA_COMPOSER_MAX_HEIGHT } from "@/components/soma/somaComposerResize";

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

  it("prefills the simplified composer when the user replies to a delivered output", () => {
    function ContinuationComposer() {
      const [value, setValue] = useState("");
      const inputRef = useRef<HTMLTextAreaElement>(null);
      useSomaOutputContinuation({ disabled: false, inputRef, setInput: setValue });
      return (
        <SomaIntentInput
          value={value}
          placeholder="Ask Soma"
          inputRef={inputRef}
          onChange={setValue}
          onSubmit={vi.fn()}
        />
      );
    }

    render(<ContinuationComposer />);

    act(() => {
      requestSomaOutputContinuation({
        title: "Launch brief",
        reference: "generated/launch/brief.md",
        proof: "proof-brief",
      });
    });

    const textarea = screen.getByPlaceholderText("Ask Soma") as HTMLTextAreaElement;
    expect(textarea.value).toContain('Use delivered output "Launch brief" as context.');
    expect(textarea.value).toContain("Reference: generated/launch/brief.md.");
    expect(textarea.value).toContain("I want an update, alternate version, or follow-up generation:");
  });

  it("expands the composer with user text until it reaches the scroll cap", () => {
    const onChange = vi.fn();
    const onSubmit = vi.fn();
    const inputRef = createRef<HTMLTextAreaElement>();

    render(
      <MissionControlAdvancedInput
        value=""
        broadcastMode={false}
        isLoading={false}
        placeholder="Tell Soma what to do"
        inputRef={inputRef}
        onChange={onChange}
        onSubmit={onSubmit}
      />,
    );

    const textarea = screen.getByPlaceholderText("Tell Soma what to do") as HTMLTextAreaElement;
    Object.defineProperty(textarea, "scrollHeight", { configurable: true, value: 132 });
    fireEvent.change(textarea, { target: { value: "Line one\nLine two\nLine three" } });

    expect(textarea.style.height).toBe("132px");
    expect(textarea.style.overflowY).toBe("hidden");

    Object.defineProperty(textarea, "scrollHeight", { configurable: true, value: 260 });
    fireEvent.change(textarea, { target: { value: "A longer ask\n".repeat(20) } });

    expect(textarea.style.height).toBe(`${SOMA_COMPOSER_MAX_HEIGHT}px`);
    expect(textarea.style.overflowY).toBe("auto");
  });
});
