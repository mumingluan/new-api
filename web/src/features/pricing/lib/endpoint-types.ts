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
import type { PricingModel } from '../types'

type EndpointTypeModel = Pick<PricingModel, 'supported_endpoint_types'>

export type AvailableEndpointType = {
  value: string
  label: string
  count: number
}

export function getAvailableEndpointTypes(
  models: EndpointTypeModel[],
  preferredOrder: readonly string[],
  labels: Readonly<Record<string, string>>
): AvailableEndpointType[] {
  const counts = new Map<string, number>()

  for (const model of models) {
    const modelEndpointTypes = new Set(model.supported_endpoint_types)
    for (const endpointType of modelEndpointTypes) {
      if (!endpointType) continue
      counts.set(endpointType, (counts.get(endpointType) ?? 0) + 1)
    }
  }

  const preferredEndpointTypes = preferredOrder.filter((endpointType) =>
    counts.has(endpointType)
  )
  const preferredEndpointTypeSet = new Set(preferredEndpointTypes)
  const remainingEndpointTypes = [...counts.keys()]
    .filter((endpointType) => !preferredEndpointTypeSet.has(endpointType))
    .sort()

  return [...preferredEndpointTypes, ...remainingEndpointTypes].map(
    (value) => ({
      value,
      label: labels[value] ?? value,
      count: counts.get(value) ?? 0,
    })
  )
}
