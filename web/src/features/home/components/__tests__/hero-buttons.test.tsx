/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { render, screen } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'

vi.mock('@tanstack/react-router', () => ({
  Link: (
    props: React.AnchorHTMLAttributes<HTMLAnchorElement> & { to: string }
  ) => {
    const { to, ...anchorProps } = props
    return <a {...anchorProps} href={to} />
  },
}))

const { HeroButtons } = await import('../hero-buttons')

describe('Xuancat home hero actions', () => {
  test('shows emphasized Docs and a normal Model List action when signed out', () => {
    render(
      <HeroButtons
        isAuthenticated={false}
        docsUrl='https://docs.example.com'
        isXuancat
      />
    )

    const docs = screen.getByRole('button', { name: 'Docs' })
    const models = screen.getByRole('button', { name: 'Model List' })

    expect(docs).toHaveAttribute('href', 'https://docs.example.com')
    expect(docs).toHaveClass('bg-primary')
    expect(docs.querySelector('svg')).not.toBeNull()
    expect(models).toHaveAttribute('href', '/pricing')
    expect(models).toHaveClass('bg-background')
    expect(models.querySelector('svg')).toBeNull()
    expect(screen.queryByRole('button', { name: 'Get Started' })).toBeNull()
    expect(screen.queryByRole('button', { name: 'View Pricing' })).toBeNull()
  })

  test('shows Dashboard, Docs, and Model List actions when signed in', () => {
    render(
      <HeroButtons
        isAuthenticated
        docsUrl='https://docs.example.com'
        isXuancat
      />
    )

    const dashboard = screen.getByRole('button', {
      name: 'Go to Dashboard',
    })
    const docs = screen.getByRole('button', { name: 'Docs' })
    const models = screen.getByRole('button', { name: 'Model List' })

    expect(dashboard).toHaveAttribute('href', '/dashboard')
    expect(dashboard).toHaveClass('bg-primary')
    expect(dashboard.querySelector('svg')).not.toBeNull()
    expect(docs).toHaveClass('bg-background')
    expect(docs.querySelector('svg')).not.toBeNull()
    expect(models).toHaveClass('bg-background')
    expect(models.querySelector('svg')).toBeNull()
  })
})
