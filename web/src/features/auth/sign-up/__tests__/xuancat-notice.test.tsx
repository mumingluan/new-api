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
import { expect, test, vi } from 'vitest'

vi.mock('@tanstack/react-router', () => ({
  Link: (
    props: React.AnchorHTMLAttributes<HTMLAnchorElement> & { to: string }
  ) => {
    const { to, ...anchorProps } = props
    return <a {...anchorProps} href={to} />
  },
}))
vi.mock('@/hooks/use-status', () => ({ useStatus: () => ({ status: {} }) }))
vi.mock('@/lib/landing-page-variant', () => ({ USE_XUANCAT_PAGES: true }))
vi.mock('../../auth-layout', () => ({
  AuthLayout: (props: React.PropsWithChildren) => <main>{props.children}</main>,
}))
vi.mock('../../components/terms-footer', () => ({ TermsFooter: () => null }))
vi.mock('../components/sign-up-form', () => ({ SignUpForm: () => null }))

const { SignUp } = await import('../index')

test('shows the administrator-only console notice on Xuancat registration', () => {
  render(<SignUp />)

  expect(
    screen.getByText(/The console login is for administrators only/)
  ).toBeInTheDocument()
  expect(screen.getByRole('link', { name: 'homepage' })).toHaveAttribute(
    'href',
    '/'
  )
})
