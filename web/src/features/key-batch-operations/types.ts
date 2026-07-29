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
export type KeyBatchAction =
  | 'extend_expiry'
  | 'add_quota'
  | 'deduct_expiry'
  | 'deduct_quota'
export type KeyBatchFilter = {
  all_users: boolean
  used_only: boolean
  min_remaining_quota?: number
}
export type KeyBatchOperationPayload = {
  action: KeyBatchAction
  duration_seconds: number
  quota: number
  filter: KeyBatchFilter
}
export type KeyBatchPreview = {
  matched_tokens: number
  actionable_tokens: number
  affected_users: number
  used_tokens: number
  permanent_tokens: number
  unlimited_tokens: number
  total_remaining_quota: number
  total_used_quota: number
}
export type KeyStatisticsGroup =
  | 'token_name'
  | 'model_name'
  | 'username'
  | 'channel_name'
  | 'user_id'
export type KeyStatisticsSort =
  | 'request_count'
  | 'prompt_tokens'
  | 'completion_tokens'
  | 'quota'
export type KeyStatisticsQuery = {
  all_users: boolean
  start_timestamp: number
  end_timestamp: number
  group_by: KeyStatisticsGroup
  sort_by: KeyStatisticsSort
  sort_order: 'asc' | 'desc'
  user_id?: number
  exclude_user_id?: number
  model?: string
  min_tokens: number
  top: number
}
export type KeyStatisticsRow = {
  name: string
  request_count: number
  prompt_tokens: number
  completion_tokens: number
  quota: number
  unique_users: number
}
export type KeyStatisticsTotals = Omit<KeyStatisticsRow, 'name'>
export type KeyStatisticsResult = {
  items: KeyStatisticsRow[]
  totals: KeyStatisticsTotals
}
export type ApiResponse<T> = { success: boolean; message: string; data: T }
