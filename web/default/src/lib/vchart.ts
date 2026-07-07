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
import { VChart, registerBrowserEnv } from '@visactor/vchart'

// VChart 的浏览器环境注册模块是带副作用的，但 @visactor/vchart 的 package.json
// 只把少量 `vchart-all`/`vchart-simple` 入口标记为 sideEffects，浏览器环境注册并不在其中。
// 在生产构建（rspack tree-shaking）下这段副作用会被摇掉，导致运行时没有任何 env 被激活，
// `application.global.envContribution` 为 undefined，渲染首个图表时抛出
// `Cannot read properties of undefined (reading 'createCanvas')`。
//
// 这里在任何图表挂载前，显式且不可被摇树的方式注册浏览器环境。useRegisters 幂等，
// 底层 vrender 容器是单例，react-vchart 与 vchart 共享同一容器，只需注册一次。
VChart.useRegisters([registerBrowserEnv])

export const VCHART_OPTION = {
  // 与老前端保持一致（浏览器环境渲染优化）
  mode: 'desktop-browser',
} as const
