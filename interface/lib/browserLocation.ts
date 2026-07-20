import { useSyncExternalStore } from "react";

const locationEvents = ["popstate", "hashchange"] as const;

function subscribeLocation(onChange: () => void) {
  locationEvents.forEach((event) => window.addEventListener(event, onChange));
  return () => locationEvents.forEach((event) => window.removeEventListener(event, onChange));
}

function readSearch() {
  return window.location.search;
}

function readServerSearch() {
  return "";
}

function subscribeHydration() {
  return () => undefined;
}

export function useBrowserSearch() {
  return useSyncExternalStore(subscribeLocation, readSearch, readServerSearch);
}

export function useClientReady() {
  return useSyncExternalStore(subscribeHydration, () => true, () => false);
}
