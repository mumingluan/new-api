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
import { describe, test } from 'node:test'

import { getAvailableEndpointTypes } from '../endpoint-types'

describe('getAvailableEndpointTypes', () => {
  test('includes unknown database endpoint names with their model counts', () => {
    const models = [
      { supported_endpoint_types: ['openai', 'audio-speech'] },
      { supported_endpoint_types: ['audio-speech', 'custom-endpoint'] },
      { supported_endpoint_types: ['openai', 'openai'] },
    ]

    const result = getAvailableEndpointTypes(models, ['openai', 'embeddings'], {
      openai: 'Chat',
      embeddings: 'Embeddings',
    })

    assert.deepEqual(result, [
      { value: 'openai', label: 'Chat', count: 2 },
      { value: 'audio-speech', label: 'audio-speech', count: 2 },
      { value: 'custom-endpoint', label: 'custom-endpoint', count: 1 },
    ])
  })

  test('omits empty endpoint names and endpoint types absent from all models', () => {
    const models = [
      { supported_endpoint_types: [] },
      { supported_endpoint_types: ['', 'jina-rerank'] },
      {},
    ]

    const result = getAvailableEndpointTypes(
      models,
      ['openai', 'jina-rerank'],
      {}
    )

    assert.deepEqual(result, [
      { value: 'jina-rerank', label: 'jina-rerank', count: 1 },
    ])
  })
})
