import { describe, it, expect } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import ServiceAddressBookPanel from "../ServiceAddressBookPanel";
import { mockFetch } from "@/__tests__/setup";

describe("ServiceAddressBookPanel", () => {
    it("renders Kubernetes Cluster Mode badge and service address entries when /api/v1/services/addresses returns a valid ServiceAddressBook", async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({
                ok: true,
                data: {
                    cluster_mode: true,
                    entries: [
                        {
                            name: "nats",
                            address: "nats://mycelis-core-nats:4222",
                            protocol: "nats",
                            source: "helm_values",
                        },
                        {
                            name: "postgres",
                            address: "mycelis-core-postgresql:5432",
                            protocol: "tcp",
                            source: "helm_values",
                        },
                        {
                            name: "text_ai",
                            address: "http://host.rancher-desktop.internal:11434/v1",
                            protocol: "http",
                            source: "runtime_env",
                        },
                    ],
                    warnings: [],
                },
            }),
        });

        render(<ServiceAddressBookPanel />);

        await waitFor(() => {
            expect(screen.getByTestId("service-address-book-panel")).toBeDefined();
        });

        expect(screen.getByTestId("service-address-cluster-badge").textContent).toContain(
            "Kubernetes Cluster Mode",
        );
        expect(screen.getByTestId("address-book-entry-nats").textContent).toContain(
            "nats://mycelis-core-nats:4222",
        );
        expect(screen.getByTestId("address-book-entry-postgres").textContent).toContain(
            "mycelis-core-postgresql:5432",
        );
        expect(screen.getByTestId("address-book-entry-text_ai").textContent).toContain(
            "http://host.rancher-desktop.internal:11434/v1",
        );
        expect(screen.queryByTestId("service-address-warnings")).toBeNull();
    });

    it("renders service-address-warnings when warnings array is non-empty", async () => {
        mockFetch.mockResolvedValue({
            ok: true,
            json: async () => ({
                ok: true,
                data: {
                    cluster_mode: true,
                    entries: [
                        {
                            name: "text_ai",
                            address: "http://localhost:11434/v1",
                            protocol: "http",
                            source: "default_fallback",
                            loopback_warning: "Loopback address in cluster runtime",
                        },
                    ],
                    warnings: [
                        "text_ai resolves to loopback http://localhost:11434/v1 inside a Kubernetes pod; use host.rancher-desktop.internal or a cluster service.",
                    ],
                },
            }),
        });

        render(<ServiceAddressBookPanel />);

        await waitFor(() => {
            expect(screen.getByTestId("service-address-warnings")).toBeDefined();
        });

        expect(screen.getByTestId("service-address-warnings").textContent).toContain(
            "text_ai resolves to loopback http://localhost:11434/v1 inside a Kubernetes pod",
        );
    });
});
