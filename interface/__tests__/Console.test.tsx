import { render, screen, fireEvent } from '@testing-library/react'
import { Console } from '../components/operator/Console'
import { describe, it, expect, vi } from 'vitest'

// Mock useChat
vi.mock('@ai-sdk/react', () => ({
    useChat: () => ({
        messages: [
            { id: '1', role: 'user', content: 'Status report' },
            { id: '2', role: 'assistant', content: 'Systems nominal.' }
        ],
        input: '',
        handleInputChange: vi.fn(),
        handleSubmit: vi.fn(),
        isLoading: false
    })
}))

// Mock scrollIntoView
window.HTMLElement.prototype.scrollIntoView = vi.fn()


describe('Operator Console', () => {
    it('renders in minimized state initially', () => {
        render(<Console />)
        expect(screen.getByText('Operator Console')).toBeDefined()
        // Content should not be visible
        expect(screen.queryByText('Status report')).toBeNull()
    })

    it('expands when clicked', () => {
        render(<Console />)
        const header = screen.getByText('Operator Console')
        fireEvent.click(header)

        // Content should now be visible
        expect(screen.getByText('Status report')).toBeDefined()
        expect(screen.getByText('Systems nominal.')).toBeDefined()
    })
})
