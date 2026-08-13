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
export type GranterResponse<T> = T & {
  error?: string
}

export type PrecheckResponse = GranterResponse<{
  valid: boolean
  channel: string
  group: string
  days: number
  client_ip: string
  new_expired_time: number
  old_expired_time?: number
}>

export type RedeemResponse = GranterResponse<{
  api_key: string
  expired_time: number
  channel: string
}>

export type RenewResponse = GranterResponse<{
  new_expired_time: number
}>

export type ActivationQueryResponse = GranterResponse<{
  activation_code: string
  used_time: string
  used_ip: string
  action: 'create' | 'renew'
  api_key: string
}>

export type SubscriptionResponse = {
  hard_limit_usd: number
  access_until?: number
  token_name?: string
}

export type UsageResponse = {
  total_usage: number
}

export type UsageLog = {
  id?: number
  type?: number
  content?: string
  token_name?: string
  created_at: number
  model_name: string
  group?: string
  use_time: number | string
  is_stream: boolean
  prompt_tokens?: number
  completion_tokens?: number
  quota: number
}

export type UsageLogResponse = {
  success: boolean
  data?: UsageLog[]
}

export type TokenSummary = {
  name: string
  balance: string
  remaining: string
  used: string
  expiry: string
}
