import { render, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import AuthProvidersPage from '@/components/settings/auth-providers/AuthProvidersPage';

describe('AuthProvidersPage', () => {
    it('renders the planned enterprise auth provider concepts', () => {
        render(<AuthProvidersPage />);

        expect(screen.getByRole('heading', { name: 'Auth Providers' })).toBeDefined();

        for (const provider of [
            'Local',
            'OIDC / OAuth',
            'SAML',
            'Entra ID',
            'Google Workspace',
            'GitHub',
            'SCIM',
        ]) {
            expect(screen.getByRole('heading', { name: provider })).toBeDefined();
        }

        expect(screen.getByText('Future provisioning')).toBeDefined();
        expect(screen.getAllByText('Future integration').length).toBeGreaterThan(0);
    });

    it('emphasizes secret manager references instead of raw secret fields', () => {
        render(<AuthProvidersPage />);

        expect(screen.getByText(/secret manager references, not inline values/i)).toBeDefined();
        expect(screen.getByText(/should never be entered as raw values/i)).toBeDefined();

        for (const secretRef of [
            'MYCELIS_AUTH_OIDC_CLIENT_SECRET_REF',
            'MYCELIS_AUTH_SAML_CERT_REF',
            'MYCELIS_AUTH_SAML_PRIVATE_KEY_REF',
            'MYCELIS_AUTH_ENTRA_CLIENT_SECRET_REF',
            'MYCELIS_AUTH_GOOGLE_CLIENT_SECRET_REF',
            'MYCELIS_AUTH_GITHUB_CLIENT_SECRET_REF',
            'MYCELIS_AUTH_SCIM_BEARER_TOKEN_REF',
        ]) {
            expect(screen.getByText(secretRef)).toBeDefined();
        }

        expect(screen.queryByLabelText(/client secret/i)).toBeNull();
        expect(screen.queryByLabelText(/private key/i)).toBeNull();
        expect(screen.queryByDisplayValue(/secret/i)).toBeNull();
    });

    it('keeps the local provider scaffold free of external secret requirements', () => {
        render(<AuthProvidersPage />);

        const localCard = screen.getByRole('heading', { name: 'Local' }).closest('article');
        expect(localCard).not.toBeNull();
        expect(within(localCard as HTMLElement).getByText('No external secret reference required.')).toBeDefined();
        expect(within(localCard as HTMLElement).getByText('Available scaffold')).toBeDefined();
    });
});
