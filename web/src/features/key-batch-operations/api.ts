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
import { api } from '@/lib/http-client'

import type {
  ApiResponse,
  KeyBatchOperationPayload,
  KeyBatchPreview,
  KeyStatisticsQuery,
  KeyStatisticsResult,
} from './types'

function requireSuccess<T>(response: ApiResponse<T>): T {
  if (!response.success) throw new Error(response.message || 'Request failed')
  return response.data
}

export async function previewKeyBatch(payload: KeyBatchOperationPayload) {
  const response = await api.post<ApiResponse<KeyBatchPreview>>(
    '/api/key-batch/preview',
    payload
  )
  return requireSuccess(response.data)
}

export async function executeKeyBatch(payload: KeyBatchOperationPayload) {
  const response = await api.post<ApiResponse<{ affected: number }>>(
    '/api/key-batch/execute',
    payload
  )
  return requireSuccess(response.data)
}

export async function getKeyStatistics(query: KeyStatisticsQuery) {
  const response = await api.get<ApiResponse<KeyStatisticsResult>>(
    '/api/key-batch/statistics',
    { params: query }
  )
  return requireSuccess(response.data)
}

export async function exportKeyStatistics(query: KeyStatisticsQuery) {
  const response = await api.get('/api/key-batch/statistics/export', {
    params: query,
    responseType: 'blob',
    disableDuplicate: true,
  })
  const url = URL.createObjectURL(response.data as Blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = 'key-statistics.csv'
  anchor.click()
  URL.revokeObjectURL(url)
}
