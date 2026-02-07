import { render, screen } from '@testing-library/react'
import { CommandRail } from '../components/shell/CommandRail'
import { describe, it, expect } from 'vitest'

describe('Command Rail Integration', () => {
    it('renders navigation buttons', () => {
        render(<CommandRail />)

        expect(screen.getByText('Mission Control')).toBeDefined()
        expect(screen.getByText('System Status')).toBeDefined()
        expect(screen.getByText('Governance')).toBeDefined()
        expect(screen.getByText('Settings')).toBeDefined()
    })

    it('renders Identity', () => {
        render(<CommandRail />)
        expect(screen.getByText('Mycelis')).toBeDefined()
    })
})
