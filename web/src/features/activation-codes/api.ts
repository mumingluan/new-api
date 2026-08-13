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
  ActivationCode,
  ActivationFilters,
  ActivationLog,
  ApiResponse,
  PageData,
} from './types'

function toTimestamp(value: string): number | undefined {
  if (!value) return undefined
  const date = new Date(value)
  return Number.isNaN(date.getTime())
    ? undefined
    : Math.floor(date.getTime() / 1000)
}

function filterParams(filters: ActivationFilters) {
  return {
    search: filters.search || undefined,
    channel: filters.channel || undefined,
    group: filters.group || undefined,
    status: filters.status || undefined,
    days: filters.days || undefined,
    created_from: toTimestamp(filters.createdFrom),
    created_to: toTimestamp(filters.createdTo),
  }
}

function requireSuccess<T>(response: ApiResponse<T>): ApiResponse<T> {
  if (!response.success) {
    throw new Error(response.message || 'Request failed')
  }
  return response
}

export async function getActivationCodes(
  page: number,
  pageSize: number,
  filters: ActivationFilters
): Promise<ApiResponse<PageData<ActivationCode>>> {
  const response = await api.get('/api/activation-code/', {
    params: { p: page, page_size: pageSize, ...filterParams(filters) },
  })
  return requireSuccess(response.data)
}

export async function getActivationLogs(
  page: number,
  pageSize: number,
  search: string,
  action: string
): Promise<ApiResponse<PageData<ActivationLog>>> {
  const response = await api.get('/api/activation-code/logs', {
    params: {
      p: page,
      page_size: pageSize,
      search: search || undefined,
      action: action || undefined,
    },
  })
  return requireSuccess(response.data)
}

export async function createActivationCodes(payload: {
  count: number
  days: number
  channel: string
  group: string
  expired_time: number
  codes: string[]
}): Promise<ApiResponse<ActivationCode[]>> {
  const response = await api.post('/api/activation-code/', payload)
  return requireSuccess(response.data)
}

export async function updateActivationCodes(payload: {
  ids?: number[]
  codes?: string[]
  days?: number
  channel?: string
  group?: string
  expired_time?: number
  status?: number
}): Promise<ApiResponse<{ updated: number }>> {
  const response = await api.put('/api/activation-code/batch', payload)
  return requireSuccess(response.data)
}

export async function deleteActivationCodes(payload: {
  ids?: number[]
  codes?: string[]
}): Promise<ApiResponse<{ deleted: number }>> {
  const response = await api.post('/api/activation-code/batch/delete', payload)
  return requireSuccess(response.data)
}

export async function exportActivationCodes(filters: ActivationFilters) {
  const response = await api.get('/api/activation-code/export', {
    params: filterParams(filters),
    responseType: 'blob',
    disableDuplicate: true,
  })
  const url = URL.createObjectURL(response.data as Blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = 'activation-codes.csv'
  anchor.click()
  URL.revokeObjectURL(url)
}
