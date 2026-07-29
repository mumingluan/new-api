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
import assert from 'node:assert/strict'
import { afterEach, describe, test } from 'node:test'

import { getSubscription, precheckActivation } from '../api'
import { formatTimestamp } from '../utils'

const originalFetch = globalThis.fetch

afterEach(() => {
  globalThis.fetch = originalFetch
})

describe('current-origin API routing', () => {
  test('sends activation prechecks to the integrated NewAPI path', async () => {
    let requestedUrl = ''
    let requestedBody = ''
    globalThis.fetch = async (input, init) => {
      requestedUrl = String(input)
      requestedBody = String(init?.body)
      return Response.json({
        valid: true,
        channel: 'test',
        days: 30,
        client_ip: '127.0.0.1',
        new_expired_time: 1_900_000_000,
      })
    }

    await precheckActivation('', {
      activation_code: 'code',
      qq: '10000',
    })

    assert.equal(requestedUrl, '/api/activation/precheck')
    assert.deepEqual(JSON.parse(requestedBody), {
      activation_code: 'code',
      qq: '10000',
    })
  })

  test('queries subscriptions on the current origin with the user API key', async () => {
    let requestedUrl = ''
    let authorization = ''
    globalThis.fetch = async (input, init) => {
      requestedUrl = String(input)
      authorization = new Headers(init?.headers).get('Authorization') ?? ''
      return Response.json({
        hard_limit_usd: 100,
        token_name: 'primary',
      })
    }

    const subscription = await getSubscription('', 'sk-test')

    assert.equal(requestedUrl, '/v1/dashboard/billing/subscription')
    assert.equal(authorization, 'Bearer sk-test')
    assert.equal(subscription.token_name, 'primary')
  })

  test('surfaces a granter error response to the form', async () => {
    globalThis.fetch = async () =>
      Response.json(
        { error: 'Activation code is already in use' },
        { status: 409 }
      )

    await assert.rejects(
      precheckActivation('', { activation_code: 'used' }),
      /Activation code is already in use/
    )
  })
})

describe('localized timestamps', () => {
  test('maps the internal Simplified Chinese language code to a valid Intl locale', () => {
    const timestamp = 1_900_000_000

    assert.equal(
      formatTimestamp(timestamp, 'zhCN'),
      new Date(timestamp * 1000).toLocaleString('zh-CN')
    )
  })
})
