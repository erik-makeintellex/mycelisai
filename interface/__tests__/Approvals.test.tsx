import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import ApprovalsPage from '../app/approvals/page'

describe('Approvals Page (Governance)', () => {
    it('renders pending decisions', () => {
        render(<ApprovalsPage />)
        expect(screen.getByText('Approvals')).toBeDefined()
        expect(screen.getByText('File Write Access')).toBeDefined()
        expect(screen.getByText(/@coder/i)).toBeDefined()
    })

    it('removes a decision card on approval', () => {
        render(<ApprovalsPage />)
        const approveButton = screen.getAllByText(/Approve Request/i)[0]

        fireEvent.click(approveButton)

        // Should be removed (or we check if only 1 remains)
        // Original count 2, new count 1
        expect(screen.getAllByText(/Risk/i)).toHaveLength(1)
    })

    it('shows empty state when all approved', () => {
        render(<ApprovalsPage />)
        const approveButtons = screen.getAllByText(/Approve Request/i)

        approveButtons.forEach(btn => fireEvent.click(btn))

        expect(screen.getByText('All Clear')).toBeDefined()
    })
})
