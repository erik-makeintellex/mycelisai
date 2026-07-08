import { useLayoutEffect } from "react";
import type React from "react";

export const SOMA_COMPOSER_MIN_HEIGHT = 40;
export const SOMA_COMPOSER_MAX_HEIGHT = 180;

export function resizeSomaComposer(input: HTMLTextAreaElement | null) {
  if (!input) return;
  input.style.height = "auto";
  const naturalHeight = Math.max(input.scrollHeight, SOMA_COMPOSER_MIN_HEIGHT);
  const nextHeight = Math.min(naturalHeight, SOMA_COMPOSER_MAX_HEIGHT);
  input.style.height = `${nextHeight}px`;
  input.style.overflowY = naturalHeight > SOMA_COMPOSER_MAX_HEIGHT ? "auto" : "hidden";
}

export function useAutoResizeSomaComposer(
  inputRef: React.RefObject<HTMLTextAreaElement | null>,
  value: string,
) {
  useLayoutEffect(() => {
    resizeSomaComposer(inputRef.current);
  }, [inputRef, value]);
}
