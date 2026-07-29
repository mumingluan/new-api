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
export type ActivationCode = {
  id: number
  user_id: number
  code: string
  days: number
  channel: string
  expired_time: number
  status: number
  created_time: number
  used_time: number
}

export type ActivationLog = {
  id: number
  activation_code_id: number
  user_id: number
  activation_code: string
  action: 'create' | 'renew'
  days: number
  identifier: string
  token_id: number
  api_key: string
  client_ip: string
  used_time: number
}

export type PageData<T> = {
  page: number
  page_size: number
  total: number
  items: T[]
}

export type ApiResponse<T> = {
  success: boolean
  message: string
  data: T
}

export type ActivationFilters = {
  search: string
  channel: string
  status: string
  days: string
  createdFrom: string
  createdTo: string
}
