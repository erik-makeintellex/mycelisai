import { render, screen, fireEvent } from '@testing-library/react'
import { Console } from '../components/operator/Console'
import { describe, it, expect, vi } from 'vitest'

// Stubbed Console Test for Maintenance Mode
describe('Operator Console (Stubbed)', () => {
    it('renders in minimized state initially', () => {
        render(<Console />)
        expect(screen.getByText('Operator Console (Offline)')).toBeDefined()
    })

    it('shows maintenance message when expanded', () => {
        render(<Console />)
        const header = screen.getByText('Operator Console (Offline)')
        fireEvent.click(header)
        expect(screen.getByText('Console currently disabled for maintenance.')).toBeDefined()
    })
})
