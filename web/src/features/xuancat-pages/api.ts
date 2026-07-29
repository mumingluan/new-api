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
import type {
  ActivationQueryResponse,
  PrecheckResponse,
  RedeemResponse,
  RenewResponse,
  SubscriptionResponse,
  UsageLogResponse,
  UsageResponse,
} from './types'

async function readJson<T>(response: Response): Promise<T> {
  const data = (await response.json()) as T
  if (!response.ok) {
    const message =
      typeof data === 'object' &&
      data !== null &&
      'error' in data &&
      typeof data.error === 'string'
        ? data.error
        : response.statusText
    throw new Error(message)
  }
  return data
}

async function postGranter<T>(
  baseUrl: string,
  path: string,
  body: Record<string, string>
): Promise<T> {
  const response = await fetch(`${baseUrl}/api/activation/${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  return readJson<T>(response)
}

export function precheckActivation(
  baseUrl: string,
  body: Record<string, string>
): Promise<PrecheckResponse> {
  return postGranter(baseUrl, 'precheck', body)
}

export function redeemActivation(
  baseUrl: string,
  activationCode: string,
  qq: string
): Promise<RedeemResponse> {
  return postGranter(baseUrl, 'redeem', {
    activation_code: activationCode,
    qq,
  })
}

export function renewActivation(
  baseUrl: string,
  activationCode: string,
  apiKey: string
): Promise<RenewResponse> {
  return postGranter(baseUrl, 'renew', {
    activation_code: activationCode,
    api_key: apiKey,
  })
}

export function queryActivation(
  baseUrl: string,
  activationCode: string
): Promise<ActivationQueryResponse> {
  return postGranter(baseUrl, 'query', {
    activation_code: activationCode,
  })
}

async function getWithKey<T>(url: string, apiKey: string): Promise<T> {
  const response = await fetch(url, {
    headers: { Authorization: `Bearer ${apiKey}` },
  })
  return readJson<T>(response)
}

export function getSubscription(
  baseUrl: string,
  apiKey: string
): Promise<SubscriptionResponse> {
  return getWithKey(`${baseUrl}/v1/dashboard/billing/subscription`, apiKey)
}

export function getUsage(
  baseUrl: string,
  apiKey: string,
  startDate: string,
  endDate: string
): Promise<UsageResponse> {
  const query = new URLSearchParams({
    start_date: startDate,
    end_date: endDate,
  })
  return getWithKey(
    `${baseUrl}/v1/dashboard/billing/usage?${query.toString()}`,
    apiKey
  )
}

export function getUsageLogs(
  baseUrl: string,
  apiKey: string
): Promise<UsageLogResponse> {
  return getWithKey(`${baseUrl}/api/log/token`, apiKey)
}
