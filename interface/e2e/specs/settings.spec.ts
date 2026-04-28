import { test, expect } from '@playwright/test';

test.describe('Settings Page (/settings)', () => {
    test.skip(({ browserName }) => browserName === 'webkit', 'WebKit currently crashes on this settings surface; keep Chromium/Firefox coverage stable for now.');

    test.beforeEach(async ({ page }) => {
        const settingsState = {
            assistant_name: 'Soma',
            theme: 'aero-light',
        };
        await page.route('**/api/v1/user/me', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    ok: true,
                    data: {
                        id: 'me-1',
                        email: 'me@local',
                        role: 'admin',
                        effective_role: 'owner',
                        name: 'Current User',
                        principal_type: 'local_admin',
                        auth_source: 'local_api_key',
                        break_glass: false,
                        settings: {
                            access_management_tier: 'release',
                            product_edition: 'self_hosted_release',
                            identity_mode: 'local_only',
                            shared_agent_specificity_owner: 'root_admin',
                        },
                    },
                }),
            });
        });
        await page.route('**/api/v1/user/settings', async (route) => {
            if (route.request().method() === 'GET') {
                await route.fulfill({
                    status: 200,
                    contentType: 'application/json',
                    body: JSON.stringify({ ok: true, data: settingsState }),
                });
                return;
            }

            const body = route.request().postDataJSON() as Record<string, unknown>;
            if (typeof body.assistant_name === 'string') {
                settingsState.assistant_name = body.assistant_name;
            }
            if (body.theme === 'aero-light' || body.theme === 'midnight-cortex' || body.theme === 'system') {
                settingsState.theme = body.theme;
            }
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({ ok: true, data: settingsState }),
            });
        });
        await page.goto('/settings');
        await page.waitForLoadState('domcontentloaded');
    });

    test('page loads without errors', async ({ page }) => {
        const errorOverlay = page.locator('nextjs-portal');
        await expect(errorOverlay).not.toBeVisible();
    });

    test('settings page leads with a guided setup path', async ({ page }) => {
        await expect(page.getByText('Guided setup path')).toBeVisible();
        await expect(page.getByText('Start with the controls most operators actually need.')).toBeVisible();
        await expect(page.getByRole('button', { name: 'Open Profile' })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Open Mission Profiles' })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Open People & Access' })).toBeVisible();
        await expect(page.getByText('Advanced setup stays intentional')).toBeVisible();
    });

    test('guided workflow controls can switch visible sections', async ({ page }) => {
        await page.getByRole('button', { name: 'Open Mission Profiles' }).click();
        await expect(page.getByRole('tab', { name: 'Mission Profiles' })).toHaveAttribute('aria-current', 'page');
        await expect(page.getByText(/Mission Profiles/i).first()).toBeVisible();
    });

    test('Auth Providers scaffold is reachable from advanced settings', async ({ page }) => {
        await page.evaluate(() => window.localStorage.setItem('mycelis-advanced-mode', 'true'));
        await page.reload({ waitUntil: 'domcontentloaded' });
        await expect(page.getByRole('button', { name: 'Open Auth Providers' })).toBeVisible();
        await page.getByRole('button', { name: 'Open Auth Providers' }).click();
        await expect(page.getByRole('tab', { name: 'Auth Providers' })).toHaveAttribute('aria-current', 'page');
        await expect(page.getByRole('heading', { name: 'Auth Providers' })).toBeVisible();
        await expect(page.getByText('OIDC / OAuth')).toBeVisible();
        await expect(page.getByRole('heading', { name: 'SAML' })).toBeVisible();
        await expect(page.getByRole('heading', { name: 'SCIM' })).toBeVisible();
        await expect(page.getByText(/No provider changes are submitted from this scaffold/i)).toBeVisible();
    });

    test('People & Access keeps enterprise user management out of the base release path', async ({ page }) => {
        await page.getByRole('button', { name: 'Open People & Access' }).click();
        await expect(page.getByRole('tab', { name: 'People & Access' })).toHaveAttribute('aria-current', 'page');
        await expect(page.getByText('Base release layer')).toBeVisible();
        await expect(page.getByRole('heading', { name: 'Organization Access' })).toBeVisible();
        await expect(page.getByText('Enterprise User Directory')).toBeVisible();
        await expect(page.getByText(/intentionally kept out of the raw release workflow/i)).toBeVisible();
        await expect(page.getByText(/Current principal contract: local_admin via local_api_key, effective role owner\./i)).toBeVisible();
        await expect(page.getByRole('button', { name: /^Self-hosted release\b/i })).toBeDisabled();
        await expect(page.getByRole('button', { name: /^Self-hosted enterprise\b/i })).toBeDisabled();
        await expect(page.getByRole('button', { name: /^Hosted control plane\b/i })).toBeDisabled();
        await expect(page.getByLabel('Identity Mode')).toBeDisabled();
        await expect(page.getByLabel('Shared Agent Specificity Owner')).toBeDisabled();
        await expect(page.getByTestId('save-access-model')).toBeDisabled();
        await expect(page.getByTestId('users-groups-section')).toBeVisible();
        await expect(page.getByTestId('users-add-button')).toHaveCount(0);
    });

    test('theme and assistant identity save and persist after reload', async ({ page }) => {
        const assistantInput = page.getByLabel('Assistant Name');
        const themeSelect = page.getByLabel('Theme');
        const saveButtons = page.getByRole('button', { name: 'Save' });
        const themeSaveButton = saveButtons.nth(1);
        await expect(assistantInput).toBeVisible();
        await expect(themeSelect).toBeVisible();
        const originalAssistantName = (await assistantInput.inputValue()).trim();
        const originalTheme = await themeSelect.inputValue();
        const updatedAssistantName = originalAssistantName === 'Atlas' ? 'Soma Prime' : 'Atlas';

        await assistantInput.fill(updatedAssistantName);
        await saveButtons.nth(0).click();
        await expect(page.getByText('Saved').first()).toBeVisible();

        const themeCandidates = ['aero-light', 'midnight-cortex', 'system'] as const;
        let updatedTheme = '';
        for (const candidate of themeCandidates) {
            await themeSelect.selectOption(candidate);
            if (await themeSaveButton.isEnabled()) {
                updatedTheme = candidate;
                break;
            }
        }
        expect(updatedTheme).not.toBe('');

        await themeSaveButton.click();
        await expect(page.locator('html')).toHaveAttribute('data-theme', updatedTheme);

        await page.reload({ waitUntil: 'domcontentloaded' });
        await expect(page.getByLabel('Assistant Name')).toHaveValue(updatedAssistantName);
        await expect(page.getByLabel('Theme')).toHaveValue(updatedTheme);
        await expect(page.locator('html')).toHaveAttribute('data-theme', updatedTheme);

        await page.getByLabel('Assistant Name').fill(originalAssistantName);
        await saveButtons.nth(0).click();
        await page.getByLabel('Theme').selectOption(originalTheme);
        if (await themeSaveButton.isEnabled()) {
            await themeSaveButton.click();
        }
    });

    test('no bg-white leak on settings page', async ({ page }) => {
        const body = await page.content();
        expect(body).not.toContain('bg-white');
    });
});
