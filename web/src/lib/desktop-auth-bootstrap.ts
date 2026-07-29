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
import { applyAuthBundle } from '@/lib/auth-session'
import type { AuthUser } from '@/stores/auth-store'

type DesktopBackend = {
  id: string
  baseUrl: string
  authMode: 'interactive' | 'accessToken'
}

export type NewApiDesktopBootstrap = {
  desktop: boolean
  backend: DesktopBackend | null
  flavor: string
  language: string
  user?: AuthUser | null
}

declare global {
  interface Window {
    __NEW_API_DESKTOP__?: NewApiDesktopBootstrap
  }
}

const DESKTOP_SESSION_EXPIRES_AT = 4_102_444_800
const DESKTOP_PROXY_ACCESS_TOKEN = 'desktop-proxy'

function isDesktopAuthUser(value: unknown): value is AuthUser {
  if (!value || typeof value !== 'object') return false
  const user = value as Partial<AuthUser>
  return (
    Number.isInteger(user.id) &&
    Number(user.id) > 0 &&
    typeof user.username === 'string' &&
    typeof user.role === 'number'
  )
}

export function initializeDesktopAuthentication(
  bootstrap = typeof window === 'undefined'
    ? undefined
    : window.__NEW_API_DESKTOP__
): boolean {
  if (
    !bootstrap?.desktop ||
    bootstrap.backend?.authMode !== 'accessToken' ||
    !bootstrap.backend.id ||
    !isDesktopAuthUser(bootstrap.user)
  ) {
    return false
  }

  const now = Math.floor(Date.now() / 1000)
  applyAuthBundle(
    {
      access_token: DESKTOP_PROXY_ACCESS_TOKEN,
      token_type: 'Bearer',
      access_expires_at: DESKTOP_SESSION_EXPIRES_AT,
      user: bootstrap.user,
      session: {
        sid: `desktop:${bootstrap.backend.id}`,
        current: true,
        login_method: 'desktop_access_token',
        ip: '127.0.0.1',
        user_agent: 'New-API-Desktop',
        created_at: now,
        last_active_at: now,
        expires_at: DESKTOP_SESSION_EXPIRES_AT,
      },
    },
    false
  )
  return true
}
