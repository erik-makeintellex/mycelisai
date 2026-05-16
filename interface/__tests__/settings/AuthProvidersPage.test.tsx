import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import AuthProvidersPage from '@/components/settings/auth-providers/AuthProvidersPage';

describe('AuthProvidersPage', () => {
    it('renders a compact provider menu with focused provider details', () => {
        render(<AuthProvidersPage />);

        expect(screen.getByRole('heading', { name: 'Auth Providers' })).toBeDefined();
        expect(screen.getByRole('navigation', { name: 'Auth provider menu' })).toBeDefined();

        for (const provider of [
            'Local',
            'OIDC / OAuth',
            'SAML',
            'Entra ID',
            'Google Workspace',
            'GitHub',
            'SCIM',
        ]) {
            expect(screen.getByRole('button', { name: new RegExp(provider, 'i') })).toBeDefined();
        }

        expect(screen.getByRole('heading', { name: 'Local' })).toBeDefined();
        expect(screen.getByText('Active release path')).toBeDefined();

        fireEvent.click(screen.getByRole('button', { name: /SCIM/i }));

        expect(screen.getByRole('heading', { name: 'SCIM' })).toBeDefined();
        expect(screen.getByText('Future provisioning path')).toBeDefined();
        expect(screen.getByText('Future integration')).toBeDefined();
    });

    it('emphasizes secret manager references instead of raw secret fields', () => {
        render(<AuthProvidersPage />);

        expect(screen.getByText(/secret manager references, not inline values/i)).toBeDefined();
        expect(screen.getByText(/should never be entered as raw values/i)).toBeDefined();

        fireEvent.click(screen.getByRole('button', { name: /OIDC \/ OAuth/i }));
        expect(screen.getByText('MYCELIS_AUTH_OIDC_CLIENT_SECRET_REF')).toBeDefined();

        fireEvent.click(screen.getByRole('button', { name: /SAML/i }));
        expect(screen.getByText('MYCELIS_AUTH_SAML_CERT_REF')).toBeDefined();
        expect(screen.getByText('MYCELIS_AUTH_SAML_PRIVATE_KEY_REF')).toBeDefined();

        fireEvent.click(screen.getByRole('button', { name: /Entra ID/i }));
        expect(screen.getByText('MYCELIS_AUTH_ENTRA_CLIENT_SECRET_REF')).toBeDefined();

        fireEvent.click(screen.getByRole('button', { name: /Google Workspace/i }));
        expect(screen.getByText('MYCELIS_AUTH_GOOGLE_CLIENT_SECRET_REF')).toBeDefined();

        fireEvent.click(screen.getByRole('button', { name: /GitHub/i }));
        expect(screen.getByText('MYCELIS_AUTH_GITHUB_CLIENT_SECRET_REF')).toBeDefined();

        fireEvent.click(screen.getByRole('button', { name: /SCIM/i }));
        expect(screen.getByText('MYCELIS_AUTH_SCIM_BEARER_TOKEN_REF')).toBeDefined();

        expect(screen.queryByLabelText(/client secret/i)).toBeNull();
        expect(screen.queryByLabelText(/private key/i)).toBeNull();
        expect(screen.queryByDisplayValue(/secret/i)).toBeNull();
    });

    it('keeps the local provider contract free of external secret requirements', () => {
        render(<AuthProvidersPage />);

        expect(screen.getByRole('button', { name: /Local/i }).getAttribute('aria-current')).toBe('page');
        expect(screen.getByText('No external secret reference required.')).toBeDefined();
        expect(screen.getAllByText('Available').length).toBeGreaterThan(0);
    });
});
