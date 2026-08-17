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
import { describe, expect, it } from 'vitest'

import { isKeyQueryLog } from '../utils'

describe('API key query logs', () => {
  it('includes consumption and error calls while excluding unrelated account logs', () => {
    expect(isKeyQueryLog(0)).toBe(true)
    expect(isKeyQueryLog(2)).toBe(true)
    expect(isKeyQueryLog(5)).toBe(true)
    expect(isKeyQueryLog(1)).toBe(false)
    expect(isKeyQueryLog(6)).toBe(false)
  })
})
