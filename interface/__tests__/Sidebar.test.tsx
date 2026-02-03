import { render, screen } from '@testing-library/react'
import { Sidebar } from '../components/layout/Sidebar'
import { describe, it, expect, vi } from 'vitest'

// Mock next/navigation
vi.mock('next/navigation', () => ({
    usePathname: () => '/',
}))

// Mock next/link (render as 'a' tag)
vi.mock('next/link', () => ({
    default: ({ children, href, className }: any) => <a href={href} className={className}>{children}</a>
}))

describe('Sidebar Integration', () => {
    it('renders navigation links', () => {
        render(<Sidebar />)

        // Command (Root)
        const commandLink = screen.getByRole('link', { name: /Command/i })
        expect(commandLink).toBeDefined()
        expect(commandLink.getAttribute('href')).toBe('/')

        // Orchestration
        const orchLink = screen.getByRole('link', { name: /Orchestration/i })
        expect(orchLink).toBeDefined()
        expect(orchLink.getAttribute('href')).toBe('/orchestration')

        // Approvals
        const approvalsLink = screen.getByRole('link', { name: /Approvals/i })
        expect(approvalsLink).toBeDefined()
        expect(approvalsLink.getAttribute('href')).toBe('/approvals')
    })

    it('displays the approvals badge', () => {
        render(<Sidebar />)
        // Check for badge "2" (hardcoded in Sidebar.tsx now)
        expect(screen.getByText('2')).toBeDefined()
    })
})
