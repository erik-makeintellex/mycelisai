import { expect, test } from "@playwright/test";
import {
    answerEnvelope,
    mockOrganizationWorkspace,
    openOrganization,
    sendWorkspaceMessage,
} from "../support/soma-ui-testing";

test.skip(({ browserName }) => browserName !== "chromium", "Deep UI testing coverage is stabilized in Chromium for the MVP audit.");

test.describe("Soma media artifacts", () => {
    test("renders generated media artifacts and exposes saved download paths", async ({ page }) => {
        await page.route("**/api/v1/artifacts/media-hero-1/save", async (route) => {
            await route.fulfill({
                status: 200,
                contentType: "application/json",
                body: JSON.stringify({
                    ok: true,
                    data: {
                        id: "media-hero-1",
                        file_path: "saved-media/launch-hero.png",
                    },
                    file_path: "saved-media/launch-hero.png",
                }),
            });
        });

        await mockOrganizationWorkspace(page, () =>
            answerEnvelope("The creative team generated a launch hero image and saved-package voiceover reference.", {
                askClass: "governed_artifact",
                artifacts: [
                    {
                        id: "media-hero-1",
                        type: "image",
                        title: "Launch hero image",
                        content_type: "image/png",
                        content: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMB/axX6fQAAAAASUVORK5CYII=",
                        cached: true,
                    },
                    {
                        id: "media-voiceover-1",
                        type: "file",
                        title: "launch-voiceover.wav",
                        content_type: "audio/wav",
                        saved_path: "saved-media/launch-voiceover.wav",
                    },
                ],
            })
        );

        await openOrganization(page);
        await sendWorkspaceMessage(page, "Generate media outputs for the marketing launch.");

        await expect(page.getByText("Soma prepared 2 artifacts for review: Launch hero image and launch-voiceover.wav.")).toBeVisible({ timeout: 20_000 });
        await expect(page.getByAltText("Launch hero image")).toBeVisible();
        await expect(page.getByText(/Saved object:/i)).toBeVisible();
        const binaryLink = page.getByRole("link", { name: "saved-media/launch-voiceover.wav", exact: true });
        await expect(binaryLink).toHaveAttribute("href", "/api/v1/artifacts/media-voiceover-1/download");

        await page.getByTitle("Save image to workspace/saved-media").click();
        await expect(page.getByText(/Saved to:/i)).toBeVisible({ timeout: 20_000 });
        const savedImageLink = page.getByRole("link", { name: "saved-media/launch-hero.png" });
        await expect(savedImageLink).toHaveAttribute("href", "/api/v1/artifacts/media-hero-1/download");
    });
});
