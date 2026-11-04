import { test, expect } from '@playwright/test';

test.describe('MVP Decomposed Settings & Modal Surfaces', () => {
    test.skip(({ browserName }) => browserName === 'webkit', 'Chromium/Firefox target coverage for settings surfaces.');

    test.beforeEach(async ({ page }) => {
        await page.addInitScript(() => {
            window.localStorage.setItem('mycelis-advanced-mode', 'true');
        });
        await page.route('**/api/v1/user/settings', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    ok: true,
                    data: { assistant_name: 'Soma', theme: 'aero-light' },
                }),
            });
        });
    });

    test('AI Engines (BrainsPage) renders provider table, probes engine, and opens/closes Add Provider modal', async ({ page }) => {
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
                    },
                }),
            });
        });
        await page.route('**/api/v1/cognitive/config', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({ ok: true, data: { default_provider: 'ollama-local' } }),
            });
        });
        await page.route('**/api/v1/cognitive/profiles', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({ ok: true, data: { profiles: { coder: 'ollama-local' } } }),
            });
        });
        await page.route('**/api/v1/brains/*/probe', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({ ok: true, data: { alive: true } }),
            });
        });
        await page.route('**/api/v1/brains', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    ok: true,
                    data: [
                        {
                            id: 'ollama-local',
                            type: 'ollama',
                            endpoint: 'http://localhost:11434/v1',
                            model_id: 'qwen2.5-coder:7b',
                            location: 'local',
                            data_boundary: 'local_only',
                            usage_policy: 'local_first',
                            token_budget_profile: 'standard',
                            max_output_tokens: 1024,
                            enabled: true,
                            status: 'online',
                            roles_allowed: ['all'],
                        },
                        {
                            id: 'openai-remote',
                            type: 'openai',
                            endpoint: 'https://api.openai.com/v1',
                            model_id: 'gpt-4o',
                            location: 'remote',
                            data_boundary: 'leaves_org',
                            usage_policy: 'require_approval',
                            token_budget_profile: 'extended',
                            max_output_tokens: 2048,
                            enabled: false,
                            status: 'disabled',
                            roles_allowed: ['architect'],
                        },
                    ],
                }),
            });
        });

        await page.goto('/settings?tab=engines');
        await page.waitForLoadState('domcontentloaded');

        const enginesTab = page.getByRole('tab', { name: 'AI Engines' });
        await expect(enginesTab).toBeVisible();
        await enginesTab.click();

        await expect(page.getByText('Provider Management')).toBeVisible();
        await expect(page.getByText('ollama-local')).toBeVisible();
        await expect(page.getByText('qwen2.5-coder:7b')).toBeVisible();
        await expect(page.getByText('LOCAL ONLY')).toBeVisible();
        await expect(page.getByText('openai-remote')).toBeVisible();
        await expect(page.getByText('gpt-4o')).toBeVisible();
        await expect(page.getByText('LEAVES ORG')).toBeVisible();

        await page.locator('button[title="Test connection"]').first().click();
        await expect(page.getByText('alive')).toBeVisible();

        await page.getByRole('button', { name: 'Add Provider' }).click();
        await expect(page.getByRole('heading', { name: 'Add Provider' })).toBeVisible();

        const providerIdInput = page.getByPlaceholder('e.g. my-ollama');
        const providerTypeInput = page.getByPlaceholder('openai_compatible');
        const endpointInput = page.getByPlaceholder('http://localhost:11434/v1');

        await providerIdInput.fill('vllm-local');
        await providerTypeInput.fill('vllm');
        await endpointInput.fill('http://127.0.0.1:8000/v1');

        await expect(providerIdInput).toHaveValue('vllm-local');
        await expect(providerTypeInput).toHaveValue('vllm');
        await expect(endpointInput).toHaveValue('http://127.0.0.1:8000/v1');

        await page.getByRole('button', { name: 'Cancel' }).click();
        await expect(providerIdInput).toHaveCount(0);
    });

    test('People & Access (UsersPage) renders access posture, break-glass status, and enterprise directory controls', async ({ page }) => {
        await page.route('**/api/v1/user/me', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    ok: true,
                    data: {
                        id: 'me-1',
                        email: 'owner@local',
                        role: 'owner',
                        effective_role: 'owner',
                        name: 'Enterprise Owner',
                        principal_type: 'local_admin',
                        auth_source: 'local_api_key',
                        break_glass: false,
                        settings: {
                            access_management_tier: 'enterprise',
                            product_edition: 'self_hosted_enterprise',
                            identity_mode: 'hybrid',
                            shared_agent_specificity_owner: 'root_admin',
                        },
                    },
                }),
            });
        });
        await page.route('**/api/v1/groups', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({ ok: true, data: [] }),
            });
        });

        await page.goto('/settings?tab=users');
        await page.waitForLoadState('domcontentloaded');

        await expect(page.getByTestId('deployment-access-model')).toBeVisible();
        await expect(page.getByTestId('organization-access-layer')).toBeVisible();
        await expect(page.getByText('Enterprise layer', { exact: true })).toBeVisible();
        await expect(page.getByText(/Current principal contract: local_admin via local_api_key, effective role owner\./i)).toBeVisible();
        await expect(page.getByText(/Hybrid mode keeps enterprise sign-in for everyday use/i)).toBeVisible();

        const directoryPanel = page.getByTestId('users-management-panel');
        await expect(directoryPanel).toBeVisible();

        await directoryPanel.getByLabel('Name').fill('Ada Lovelace');
        await directoryPanel.getByLabel('Email').fill('ada@mycelis.local');
        await page.getByTestId('users-add-button').click();

        await expect(directoryPanel.getByText('Ada Lovelace')).toBeVisible();
        await expect(directoryPanel.getByText('ada@mycelis.local')).toBeVisible();
        await expect(directoryPanel.getByText(/User added to local UI state/i)).toBeVisible();
    });

    test('Mission Profiles (MissionProfilesPage) renders profile cards and opens/closes MissionProfileEditor', async ({ page }) => {
        await page.route('**/api/v1/user/me', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    ok: true,
                    data: { id: 'me-1', email: 'me@local', role: 'owner', effective_role: 'owner', name: 'Current User' },
                }),
            });
        });
        await page.route('**/api/v1/brains', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    ok: true,
                    data: [
                        {
                            id: 'ollama-local',
                            type: 'ollama',
                            model_id: 'qwen2.5-coder:7b',
                            location: 'local',
                            enabled: true,
                            status: 'online',
                        },
                    ],
                }),
            });
        });
        await page.route('**/api/v1/mission-profiles', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    ok: true,
                    data: [
                        {
                            id: 'prof-1',
                            name: 'Local Code Audit',
                            description: 'Deterministic local code review profile',
                            role_providers: { coder: 'ollama-local', architect: 'ollama-local' },
                            subscriptions: [{ topic: 'swarm.team.audit.signal.status' }],
                            context_strategy: 'fresh',
                            auto_start: true,
                            is_active: true,
                        },
                    ],
                }),
            });
        });

        await page.goto('/settings?tab=profiles');
        await page.waitForLoadState('domcontentloaded');

        await expect(page.getByText('Local Code Audit')).toBeVisible();
        await expect(page.getByText('Deterministic local code review profile')).toBeVisible();
        await expect(page.getByText('ACTIVE', { exact: true })).toBeVisible();
        await expect(page.getByText('AUTO', { exact: true })).toBeVisible();
        await expect(page.getByText('1 NATS watch')).toBeVisible();

        await page.getByRole('button', { name: 'New Profile' }).click();
        await expect(page.getByRole('heading', { name: 'New Profile' })).toBeVisible();

        const profileNameInput = page.getByPlaceholder('Research Mode');
        await expect(profileNameInput).toBeVisible();
        await profileNameInput.fill('Deep Architecture Recon');
        await expect(profileNameInput).toHaveValue('Deep Architecture Recon');
        await expect(page.getByRole('button', { name: 'Save Profile' })).toBeEnabled();

        await page.getByRole('button', { name: 'Cancel' }).click();
        await expect(profileNameInput).toHaveCount(0);
    });
});
