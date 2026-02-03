import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import CommandPage from '../app/page'

describe('Command Page (Root)', () => {
    it('renders the Status Bar with precise indicators', () => {
        render(<CommandPage />)
        expect(screen.getByText('UPLINK')).toBeDefined()
        expect(screen.getByText('BRAIN')).toBeDefined()
        expect(screen.getByText('CAPACITY')).toBeDefined()
        expect(screen.getByText('4/12')).toBeDefined()
    })

    it('renders the Command Input Hero', () => {
        render(<CommandPage />)
        const input = screen.getByPlaceholderText(/Ready for instruction/i)
        expect(input).toBeDefined()
        // Check auto-focus attribute logic if possible, or just presence
    })

    it('renders the Session Feed (Active Operations)', () => {
        render(<CommandPage />)
        // Check table headers or content
        expect(screen.getByText('Agent')).toBeDefined()
        expect(screen.getByText('Intent')).toBeDefined()
        expect(screen.getByText('Output Stream')).toBeDefined()

        // Check a mock session
        // Check a mock session
        expect(screen.getByText(/@architect/i)).toBeDefined()
        expect(screen.getByText(/drafting/i)).toBeDefined()
    })
})
