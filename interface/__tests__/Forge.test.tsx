import { render, screen, fireEvent } from '@testing-library/react'
import { Forge } from '../components/loom/Forge'
import { describe, it, expect, vi } from 'vitest'

describe('Forge Wizard', () => {
    it('renders nothing when closed', () => {
        const { container } = render(<Forge isOpen={false} onClose={vi.fn()} />)
        expect(container.firstChild).toBeNull()
    })

    it('renders Prime Directive input when open', () => {
        render(<Forge isOpen={true} onClose={vi.fn()} />)
        expect(screen.getByText('The Forge')).toBeDefined()
        expect(screen.getByPlaceholderText(/Use the Twitter API/i)).toBeDefined()
    })

    it('advances to Blueprint step on click', () => {
        render(<Forge isOpen={true} onClose={vi.fn()} />)

        const nextBtn = screen.getByText('Generate Blueprint')
        fireEvent.click(nextBtn)

        expect(screen.getByText('Architecting Solution...')).toBeDefined()
    })
})
