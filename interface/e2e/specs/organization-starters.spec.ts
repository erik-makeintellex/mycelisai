import { expect, test, type Page } from '@playwright/test';
import {
    clickStartMode,
    mockOrganizationEntryApis,
    openCreatedOrganization,
    openOrganizationSetup,
} from '../support/organization-entry';
import type { OrganizationStarterTemplate } from '@/lib/organizations';

// RC-004C / RC-004C-2: data-driven starter cards from
// GET /api/v1/templates?view=organization-starters, and the resulting
// Soma intro + delivery posture line on the created organization's home.

test.skip(({ browserName }) => browserName !== 'chromium', 'Organization starter coverage is stabilized in Chromium for the MVP audit.');

/**
 * Opens organization setup and switches to template mode, waiting for the
 * starter templates request to land first so the mode switch is not racing
 * the initial async load (see clickStartMode in support/organization-entry).
 */
async function openTemplateStarters(page: Page) {
    const templatesLoaded = page.waitForResponse(
        (response) => response.url().includes('/api/v1/templates') && response.request().method() === 'GET',
    );
    await openOrganizationSetup(page);
    await templatesLoaded;
    await clickStartMode(page, /Start from template/i);
}

function starter(overrides: Partial<OrganizationStarterTemplate>): OrganizationStarterTemplate {
    return {
        id: overrides.id || 'starter-x',
        name: overrides.name || 'Starter',
        description: overrides.description || 'A starter organization.',
        organization_type: 'AI Organization starter',
        team_lead_label: 'Team Lead',
        advisor_count: 1,
        department_count: 1,
        specialist_count: 2,
        ai_engine_settings_summary: 'Local engines',
        response_contract_summary: 'Balanced',
        memory_personality_summary: 'Starter defaults',
        source: 'starter',
        available: true,
        ...overrides,
    } as OrganizationStarterTemplate;
}

const studio = starter({
    id: 'starter-client-delivery-studio',
    name: 'Client delivery studio',
    starter: {
        label: 'Client delivery studio',
        persona: 'studio',
        order: 10,
        sample_name: 'Brightline Studio',
        sample_purpose: 'Draft client work and hold it for approval.',
        soma_intro: "I'll keep client work drafted and reviewable until you approve it.",
    },
    delivery_defaults: {
        work_mode: 'propose_only',
        approval_policy_ref: 'delivery-approve-external',
        approval_posture: 'approve_external_sends',
        audit: 'standard',
        outputs: { retain: true, proof_required: true },
        max_team_members: 3,
    },
    delivery_summary: {
        work_mode_label: 'Propose only',
        approval_posture: 'Nothing leaves the organization until you approve it.',
        output_proof: 'Outputs are retained with proof.',
        team_shape: 'Up to 3 members per team, plus Soma.',
    },
});

const product = starter({
    id: 'starter-product-delivery-team',
    name: 'Product delivery team',
    starter: {
        label: 'Product delivery team',
        persona: 'saas_product_team',
        order: 20,
        sample_name: 'Northstar Labs',
        sample_purpose: 'Plan priorities, propose changes, and ship them with proof.',
    },
    delivery_defaults: {
        work_mode: 'execute_with_approval',
        approval_policy_ref: 'delivery-approve-changes',
        approval_posture: 'approve_changes',
        audit: 'standard',
        outputs: { retain: true, proof_required: true },
        max_team_members: 3,
    },
    delivery_summary: {
        work_mode_label: 'Execute with approval',
        approval_posture: 'Deploy and write actions wait for your approval.',
        output_proof: 'Outputs are retained with proof.',
        team_shape: 'Up to 3 members per team, plus Soma.',
    },
});

const operations = starter({
    id: 'starter-operations-desk',
    name: 'Operations desk',
    starter: {
        label: 'Operations desk',
        persona: 'operations',
        order: 30,
        sample_name: 'Skylight Works',
        sample_purpose: 'Watch what you connect and summarize daily.',
    },
    delivery_defaults: {
        work_mode: 'execute_bounded',
        approval_policy_ref: 'delivery-approve-state-change',
        approval_posture: 'approve_high_risk',
        audit: 'standard',
        outputs: { retain: true, proof_required: true },
        max_team_members: 3,
    },
    delivery_summary: {
        work_mode_label: 'Bounded execution',
        approval_posture: 'Routine checks run within bounds; high-risk actions wait for your approval.',
        output_proof: 'Outputs are retained with proof.',
        team_shape: 'Up to 3 members per team, plus Soma.',
    },
});

const regulated = starter({
    id: 'starter-governed-enterprise-program',
    name: 'Governed enterprise program',
    starter: {
        label: 'Governed enterprise program',
        persona: 'regulated_enterprise',
        order: 40,
        sample_name: 'Meridian Compliance',
        sample_purpose: 'Every mutation is a proposal with proof.',
    },
    available: false,
    unavailable_reason: 'approval policy "delivery-approve-all" is not a group in policy.yaml',
    delivery_summary: undefined,
});

const createdStudioOrganization = {
    id: 'org-studio-1',
    name: 'Brightline Studio',
    purpose: 'Draft client work and hold it for approval.',
    start_mode: 'template',
    template_id: studio.id,
    team_lead_label: 'Team Lead',
    advisor_count: 1,
    department_count: 1,
    specialist_count: 2,
    ai_engine_settings_summary: 'Local engines',
    response_contract_summary: 'Balanced',
    memory_personality_summary: 'Starter defaults',
    status: 'ready',
    soma_intro: studio.starter!.soma_intro,
    delivery_defaults: studio.delivery_defaults,
};

test.describe('AI Organization starter cards (/dashboard create flow)', () => {
    test('starter cards render in order', async ({ page }) => {
        await mockOrganizationEntryApis(page, { templates: [studio, product, operations, regulated] });

        await page.goto('/dashboard', { waitUntil: 'domcontentloaded' });
        await openTemplateStarters(page);

        const cards = page.locator('[data-testid^="starter-card-"]');
        await expect(cards).toHaveCount(4);
        await expect(cards.nth(0)).toHaveAttribute('data-testid', 'starter-card-starter-client-delivery-studio');
        await expect(cards.nth(1)).toHaveAttribute('data-testid', 'starter-card-starter-product-delivery-team');
        await expect(cards.nth(2)).toHaveAttribute('data-testid', 'starter-card-starter-operations-desk');
        await expect(cards.nth(3)).toHaveAttribute('data-testid', 'starter-card-starter-governed-enterprise-program');
    });

    test('selecting a starter prefills name and purpose and sends template_id on create', async ({ page }) => {
        let capturedRequestBody: Record<string, unknown> | null = null;
        await mockOrganizationEntryApis(page, {
            templates: [studio, product, operations, regulated],
            homeResponsesById: { [createdStudioOrganization.id]: createdStudioOrganization },
            createHandler: (requestBody) => {
                capturedRequestBody = requestBody;
                return { status: 201, body: { ok: true, data: createdStudioOrganization } };
            },
        });

        await page.goto('/dashboard', { waitUntil: 'domcontentloaded' });
        await openTemplateStarters(page);

        await page.getByTestId(`starter-card-${studio.id}`).click();
        await expect(page.getByLabel('AI Organization name')).toHaveValue(studio.starter!.sample_name!);
        await expect(page.getByLabel('Purpose')).toHaveValue(studio.starter!.sample_purpose!);

        await Promise.all([
            page.waitForResponse(
                (response) => response.url().endsWith('/api/v1/organizations') && response.request().method() === 'POST',
            ),
            page.getByRole('button', { name: 'Create AI Organization', exact: true }).last().click(),
        ]);

        await expect.poll(() => capturedRequestBody).toEqual(
            expect.objectContaining({ template_id: studio.id }),
        );
    });

    test('an unavailable starter is disabled and shows its reason', async ({ page }) => {
        await mockOrganizationEntryApis(page, { templates: [studio, product, operations, regulated] });

        await page.goto('/dashboard', { waitUntil: 'domcontentloaded' });
        await openTemplateStarters(page);

        const regulatedCard = page.getByTestId(`starter-card-${regulated.id}`);
        await expect(regulatedCard).toBeDisabled();
        await expect(regulatedCard).toContainText(regulated.unavailable_reason!);

        await regulatedCard.click({ force: true });
        await expect(page.getByLabel('AI Organization name')).not.toHaveValue(regulated.starter!.sample_name!);
    });

    test('starter API failure falls back to Start empty', async ({ page }) => {
        await mockOrganizationEntryApis(page, { templateStatus: 500, templateError: 'Starter templates are unavailable right now.' });

        await page.goto('/dashboard', { waitUntil: 'domcontentloaded' });
        await openTemplateStarters(page);

        await expect(page.getByText('Starter templates are unavailable', { exact: true })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Start empty instead' })).toBeVisible();
        await page.getByRole('button', { name: 'Start empty instead' }).click();
        await expect(page.getByLabel('AI Organization name')).toBeVisible();
    });

    test('after creation, the organization home shows the Soma intro and the posture line', async ({ page }) => {
        await mockOrganizationEntryApis(page, {
            templates: [studio, product, operations, regulated],
            homeResponsesById: { [createdStudioOrganization.id]: createdStudioOrganization },
            createHandler: () => ({ status: 201, body: { ok: true, data: createdStudioOrganization } }),
        });

        await page.goto('/dashboard', { waitUntil: 'domcontentloaded' });
        await openTemplateStarters(page);
        await page.getByTestId(`starter-card-${studio.id}`).click();

        await Promise.all([
            page.waitForResponse(
                (response) => response.url().endsWith('/api/v1/organizations') && response.request().method() === 'POST',
            ),
            page.getByRole('button', { name: 'Create AI Organization', exact: true }).last().click(),
        ]);
        await openCreatedOrganization(page, createdStudioOrganization.id);

        await expect(page.getByText(studio.starter!.soma_intro!)).toBeVisible();
        await expect(page.getByText(/nothing leaves the organization without your approval/i)).toBeVisible();
        await expect(page.getByText(/outputs are kept with proof/i)).toBeVisible();
        await expect(page.getByText(/teams top out at 3 members/i)).toBeVisible();
    });
});
