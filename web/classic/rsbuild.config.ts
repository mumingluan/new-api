import path from 'path'
import { createRequire } from 'module'
import { fileURLToPath } from 'url'
import { defineConfig, loadEnv } from '@rsbuild/core'
import { pluginReact } from '@rsbuild/plugin-react'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const require = createRequire(import.meta.url)
const semiUiDir = path.resolve(
  path.dirname(require.resolve('@douyinfe/semi-ui')),
  '../..',
)
const semiDateFnsDir = path.resolve(
  semiUiDir,
  'node_modules/date-fns',
)

// VChart 依赖的底层 vrender-*/vutils 在 classic 依赖树里存在多份物理副本
// （react-vchart 与 vchart 各自嵌套一份 0.17.17）。vrender-core 内部通过
// application.global 维护渲染环境单例，多份副本 = 多个互不相通的单例：
// registerBrowserEnv 注册到其中一份，而 <VChart> 渲染时读取的是另一份，
// application.global.envContribution 仍为 undefined，首个图表 createCanvas 崩溃。
//
// 这里把这些底层包强制指向 vchart 自带的那份 0.17.17（完整集合，含
// vrender-components），保证 vchart 与 react-vchart 共用同一个 vrender-core 单例。
// 注意不能指向 workspace 顶层 hoist 的 1.1.4：那是给新前端 vchart 2.x 用的，
// 与 classic 的 vchart 1.8.11 大版本不兼容。
const vchartDir = path.dirname(require.resolve('@visactor/vchart/package.json'))
const vrenderDedupeAlias = Object.fromEntries(
  [
    '@visactor/vrender-core',
    '@visactor/vrender-kits',
    '@visactor/vrender-components',
    '@visactor/vutils',
  ].map((pkg) => [
    pkg,
    path.join(vchartDir, 'node_modules', pkg),
  ]),
)

export default defineConfig(({ envMode }) => {
  const env = loadEnv({ mode: envMode, prefixes: ['VITE_'] })
  const clientServerUrl =
    process.env.VITE_REACT_APP_SERVER_URL ||
    env.rawPublicVars.VITE_REACT_APP_SERVER_URL ||
    ''
  const proxyServerUrl =
    clientServerUrl ||
    'http://localhost:3000'
  const isProd = envMode === 'production'
  const devProxy = Object.fromEntries(
    (['/api', '/mj', '/pg'] as const).map((key) => [
      key,
      { target: proxyServerUrl, changeOrigin: true },
    ]),
  ) as Record<string, { target: string; changeOrigin: boolean }>

  return {
    plugins: [pluginReact()],
    source: {
      entry: {
        index: './src/index.jsx',
      },
      define: {
        'import.meta.env.VITE_REACT_APP_SERVER_URL': JSON.stringify(
          clientServerUrl,
        ),
      },
    },
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
        '@douyinfe/semi-ui/dist/css/semi.css': path.resolve(
          semiUiDir,
          'dist/css/semi.css',
        ),
        'date-fns': semiDateFnsDir,
        // Force a single physical copy of vrender-*/vutils so VChart's render
        // environment singleton is shared (see vrenderDedupeAlias above).
        ...vrenderDedupeAlias,
      },
    },
    html: {
      template: './index.html',
    },
    server: {
      host: '0.0.0.0',
      strictPort: false,
      proxy: devProxy,
    },
    output: {
      minify: isProd,
      target: 'web',
      distPath: {
        root: 'dist',
      },
    },
    performance: {
      removeConsole: isProd ? ['log'] : false,
      buildCache: {
        cacheDigest: [process.env.VITE_REACT_APP_VERSION],
      },
    },
    tools: {
      rspack: {
        module: {
          rules: [
            {
              test: /src[\\/].*\.js$/,
              type: 'javascript/auto',
              use: [
                {
                  loader: 'builtin:swc-loader',
                  options: {
                    jsc: {
                      parser: {
                        syntax: 'ecmascript',
                        jsx: true,
                      },
                      transform: {
                        react: {
                          runtime: 'automatic',
                          development: !isProd,
                          refresh: !isProd,
                        },
                      },
                    },
                  },
                },
              ],
            },
          ],
        },
      },
    },
  }
})
