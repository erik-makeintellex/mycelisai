import { render, screen, act, fireEvent } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import CommandPage from '../app/page'

describe('Command Page (Genesis Terminal)', () => {
    it('renders the Pathfinder State (Role Selection)', async () => {
        await act(async () => {
            render(<CommandPage />)
        })

        // Header
        expect(screen.getByText('System Initialization')).toBeDefined()
        expect(screen.getByText(/The Cortex is untethered/i)).toBeDefined()

        // Cards
        expect(screen.getByText('Architect')).toBeDefined()
        expect(screen.getByText('Commander')).toBeDefined()
        expect(screen.getByText('Explorer')).toBeDefined()
    })

    it('transitions to Negotiation State on role selection', async () => {
        await act(async () => {
            render(<CommandPage />)
        })

        const architectCard = screen.getByText('Architect')

        await act(async () => {
            fireEvent.click(architectCard)
        })

        // Negotiation View Check
        expect(screen.getByText('Consular Link')).toBeDefined()
        expect(screen.getByText('Online')).toBeDefined()
        expect(screen.getByPlaceholderText('Type your directive...')).toBeDefined()
    })
})
