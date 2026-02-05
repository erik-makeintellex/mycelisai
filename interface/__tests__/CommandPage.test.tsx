import { render, screen, act } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import CommandPage from '../app/page'

// Mock useSWR to avoid network calls and act warnings
import useSWR from 'swr'
import { vi } from 'vitest'

vi.mock('swr', () => ({
    default: vi.fn()
}))

describe('Command Page (Genesis Terminal)', () => {
    it('renders the Pathfinder Deck navigation', async () => {
        (useSWR as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
            data: [],
            error: undefined
        })

        await act(async () => {
            render(<CommandPage />)
        })

        expect(screen.getByText('Manual Entry')).toBeDefined()
        expect(screen.getByText('Template Library')).toBeDefined()
        expect(screen.getByText('Node Explorer')).toBeDefined()
    })

    it('renders the Resource Grid empty state', async () => {
        (useSWR as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
            data: [],
            error: undefined
        })

        await act(async () => {
            render(<CommandPage />)
        })

        expect(screen.getByText(/DETECTED SIGNALS/i)).toBeDefined()
        expect(screen.getByText('NO HARDWARE DETECTED')).toBeDefined()
        expect(screen.getByText('SCANNING...')).toBeDefined()
    })

    it('switches to Manual Entry mode', async () => {
        (useSWR as unknown as ReturnType<typeof vi.fn>).mockReturnValue({ data: [] })
        render(<CommandPage />)

        const manualTab = screen.getByText('Manual Entry')
        await act(async () => {
            manualTab.click()
        })

        expect(screen.getByText('MANUAL INGESTION PROTOCOL')).toBeDefined()
        expect(screen.getByText('INITIATE FORM')).toBeDefined()
    })

    it('switches to Template Library mode', async () => {
        (useSWR as unknown as ReturnType<typeof vi.fn>).mockReturnValue({ data: [] })
        render(<CommandPage />)

        const templateTab = screen.getByText('Template Library')
        await act(async () => {
            templateTab.click()
        })

        expect(screen.getByText('TEMPLATE LIBRARY UNLINKED')).toBeDefined()
        expect(screen.getByText('CONNECT REGISTRY')).toBeDefined()
    })
})
