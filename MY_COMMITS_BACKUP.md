# zch-api 本地修改备份

本文档记录当前 fork 相对上游基点仍然保留的本地改动。

上游基点：

```text
172114422 fix(auth): keep login state on rate-limited or failing token refresh
```

更新时间：2026-07-23。已逐个查看当前 fork 独有提交，并按“当前工作区最终仍保留的差异”重新整理。Classic 前端及其 Electron 远程后端壳已迁移到独立仓库 `mumingluan/new-api-desktop`；本仓库的 `web/` 只保留新版前端，`electron/` 与上游保持一致。Redis Sentinel 支持和 Sentinel 故障转移重试逻辑已经从当前工作区移除，不再作为保留改动记录。

## 当前保留的改动

### 1. 按 API 密钥维度的请求限流

在原有按用户/分组的模型请求限流之外，新增按 API key/token 维度的限流。

保留行为：

- 分钟级密钥限流：
  - 总请求数限制，包含失败请求；
  - 成功请求数限制，只统计成功请求；
  - 支持全局默认值和分组覆盖。
- 每日密钥限流：
  - 每日总请求数限制，包含失败请求；
  - 每日成功请求数限制，只统计成功请求；
  - 支持全局默认值和分组覆盖；
  - 以北京时间自然日为周期，UTC+8，每日 00:00 重置。
- Redis 和内存模式都支持。
- Redis 每日成功请求数使用预占配额方式，请求失败时回滚。
- per-user 限流继续使用用户分组；per-token 限流优先使用 token 分组，缺失时回退到用户分组。
- Suno、Midjourney、通用视频、Kling、Jimeng 任务路由补上 `ModelRequestRateLimit()`，任务类接口也会进入限流链路。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `middleware/model-rate-limit.go` | 密钥分钟/每日限流、Redis/内存计数、每日配额回滚、成功请求记录 |
| `common/rate-limit.go` | 新增 `InMemoryRateLimiter.Check`，用于只检查不记录 |
| `common/limiter/lua/rate_limit.lua` | 恢复 Redis token bucket key 过期时间设置 |
| `constant/context_key.go` | 新增每日成功配额预占相关 context key |
| `setting/rate_limit.go` | 新增密钥分钟/每日限流设置和分组 JSON 辅助函数 |
| `model/option.go` | 将密钥限流配置接入现有 OptionMap |
| `router/relay-router.go` | Suno、Midjourney 任务路由加入限流中间件 |
| `router/video-router.go` | video、Kling、Jimeng 任务路由加入限流中间件 |

前端配置：

- 新版前端新增 `TokenRateLimitSection` 和 `TokenDailyRateLimitSection`。
- en、zh、fr、ja、ru、vi 翻译文件补齐相关文案。

补充说明：

- `TokenRateLimitCount` 和 `TokenDailyRateLimitCount` 统计全部请求。
- `TokenRateLimitSuccessCount` 和 `TokenDailyRateLimitSuccessCount` 只统计成功请求。
- 数值为 `0` 表示不限制。

### 2. 上游空响应、拒绝原因和阻断响应处理

增强 relay 对上游异常/空响应的判断，避免把不可用响应当成成功请求继续处理。

保留行为：

- OpenAI 兼容流式响应如果没有任何 stream data、没有累计文本、也没有 usage-only final response，则返回 `ErrorCodeEmptyResponse`，触发重试/错误处理。
- OpenAI 兼容非流式 chat/completions 响应如果 `choices` 为空，则返回 `ErrorCodeEmptyResponse`。
- 如果 chat/completions 路由收到上游 Responses API payload，会尝试转换成 chat completions，而不是直接判定为空响应。
- Gemini prompt 被阻断时记录 `ContextKeyAdminRejectReason`，并使用 `ErrOptionWithSkipRetry()` 避免无意义重试。
- Gemini 无 candidates 时记录 `gemini_empty_candidates`，并返回 `ErrorCodeEmptyResponse`。
- Claude 流式/非流式处理会检查空响应，并从 stop reason 记录拒绝原因。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `relay/channel/openai/relay-openai.go` | OpenAI 空响应检测、Responses fallback、工具调用 token 计数 |
| `relay/channel/gemini/relay-gemini.go` | Gemini 空 candidates、block reason、流式空响应检测 |
| `relay/channel/gemini/relay-gemini-native.go` | Gemini native 空响应/block 处理 |
| `relay/channel/claude/relay-claude.go` | Claude 空响应检测和拒绝原因记录 |

测试：

| 文件 | 说明 |
| ---- | ---- |
| `relay/channel/openai/relay_openai_responses_fallback_test.go` | Responses payload fallback 到 chat completions |
| `relay/channel/claude/relay_claude_tool_use_empty_content_test.go` | Claude tool_use 且文本为空的响应场景 |

### 3. Claude 和 Ollama 支持 Gemini 格式请求

为 Claude 和 Ollama adaptor 补上 `ConvertGeminiRequest`。

实现方式：

- 先通过 `service.GeminiToOpenAIRequest` 将 Gemini 请求转成 OpenAI 请求；
- 再复用各自已有的 `ConvertOpenAIRequest` 转换链路。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `relay/channel/claude/adaptor.go` | Gemini -> OpenAI -> Claude |
| `relay/channel/ollama/adaptor.go` | Gemini -> OpenAI -> Ollama |

### 4. Claude 响应可转换回 Gemini 格式

当入口 relay format 是 Gemini 时，Claude 的流式和非流式响应可以转换回 Gemini 响应格式。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `relay/channel/claude/relay-claude.go` | Claude 流式 chunk、final usage、非流式响应转 Gemini |

### 5. 工具调用 token 计数和流式 final chunk 保留

保留行为：

- OpenAI 非流式响应在上游 usage 缺少 completion tokens 时，将 tool call 的函数名和参数计入估算文本。
- OpenAI 流式 final chunk 判断中，将 tool calls 视为有效输出，避免 tool-call-only final chunk 被丢弃。
- Gemini 流式 usage 估算会把 function call 名称和序列化参数加入 response text。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `relay/channel/openai/relay-openai.go` | 工具调用 token 计数 |
| `relay/channel/openai/helper.go` | 保留包含 tool calls 的最后流式响应 |
| `relay/channel/gemini/relay-gemini.go` | Gemini function call usage 估算 |

测试：

| 文件 | 说明 |
| ---- | ---- |
| `relay/channel/openai/helper_last_response_test.go` | tool-call-only last stream response 不应被丢弃 |

### 6. Embeddings 响应走 usage-aware 处理

让 embeddings 路由在适用场景下走带 usage 提取的 OpenAI 处理器，保证用量和计费能拿到响应 usage。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `relay/channel/openai/adaptor.go` | embeddings 使用 `OpenaiHandlerWithUsage` |
| `relay/channel/jina/adaptor.go` | Jina embeddings 使用 `OpenaiHandlerWithUsage` |

### 7. Gemini 视频代理和任务查询改进

保留行为：

- Gemini 视频代理支持直接返回 `bytesBase64Encoded` 内联视频数据。
- 原有 remote URL 提取仍然保留。
- Gemini task fetch 改为 POST `models/{model}:fetchPredictOperation`，body 中传 `operationName`。
- 视频代理解析 JSON 使用 `common.Unmarshal`，符合项目 JSON wrapper 规则。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `controller/video_proxy.go` | 根据结果直接写视频 bytes 或继续代理 remote URL |
| `controller/video_proxy_gemini.go` | 提取 Gemini remote URL 或 base64 视频 payload |
| `relay/channel/task/gemini/adaptor.go` | 使用新的 `fetchPredictOperation` POST 查询方式 |

### 8. API key 和兑换码额度默认值/快捷金额

保留行为：

- 新版前端新建 API key 默认 1 美元额度，并默认不是无限额度。
- 新版前端的 API key 和兑换码表单加入常用美元金额快捷按钮。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `web/src/features/keys/lib/api-key-form.ts` | API key 默认有限额度 |
| `web/src/features/keys/components/api-keys-mutate-drawer.tsx` | API key 金额预设按钮 |
| `web/src/features/redemption-codes/components/redemptions-mutate-drawer.tsx` | 兑换码金额预设按钮 |

### 9. 控制台登录仅面向管理员的前端提示

保留行为：

- Public header 不再给未登录访客展示登录/注册按钮。
- 登录页增加提示：控制台登录仅供管理员使用，普通用户请访问首页查看教程。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `web/src/components/layout/components/public-header.tsx` | 隐藏未登录访客登录入口 |
| `web/src/features/auth/sign-in/index.tsx` | 登录页增加管理员提示 |
| `web/src/i18n/locales/*.json` | 补充提示文案翻译 |

### 10. Classic 前端和桌面壳拆分

仓库边界：

- Classic 前端完整源码迁移到独立仓库 `mumingluan/new-api-desktop` 的 `web/classic/`。
- 独立桌面仓库负责构建 Classic 前端，并保留切换新版/Classic 前端、连接远程后端的桌面逻辑。
- 本仓库删除 `web/classic/`，`web/` 只保留上游新版前端。
- 本仓库 `electron/`、Electron 构建工作流以及 `web/package.json`、`web/bun.lock` 均还原为上游版本。
- 桌面应用及产物名统一为 `New-API-Desktop`。

### 11. 测试稳定性清理

保留行为：

- Channel affinity usage cache 测试改用 atomic counter 生成唯一 key，替代 `time.Now().UnixNano()`，降低并发/快速执行时的 key 碰撞风险。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `service/channel_affinity_usage_cache_test.go` | atomic 唯一测试 key |

### 12. VChart 图表崩溃与深色模式修复

修复新版前端仪表盘图表在生产构建下运行时崩溃（`TypeError: Cannot read properties of undefined (reading 'createCanvas')`）的问题。

保留行为：

- VChart 的浏览器环境注册是带副作用的，但 `@visactor/vchart` 的 `package.json` `sideEffects` 未列出它，生产构建 tree-shaking 会摇掉，运行时没有 env 被激活，`application.global.envContribution` 为 undefined，首个图表 `createCanvas` 崩溃。
- 改为显式且不可被摇树的 `VChart.useRegisters([registerBrowserEnv])`，在任何图表挂载前注册。
- Classic 的 VisActor 去重和深色主题修复随 Classic 源码迁移到 `mumingluan/new-api-desktop`，不再属于本仓库差异。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `web/src/lib/vchart.ts` | 新版前端显式注册浏览器环境 `VChart.useRegisters([registerBrowserEnv])` |
| `web/src/main.tsx` | 入口 side-effect 导入 `@/lib/vchart`，防止被 tree-shaking 丢弃 |

对应提交：

```text
724ecece2 fix(web): register VChart browser env to prevent createCanvas crash
```

### 13. 上游响应语义校验与安全重试

修复 OpenAI Chat/Completions/Responses、Claude 和 Gemini 上游返回 HTTP 200、JSON/SSE 外壳合法但没有任何可消费语义输出时，被误当成成功响应并计费的问题，同时覆盖纯工具调用和流式边缘场景。

保留行为：

- 统一按“语义输出”而不是响应体非空判断成功：文本、reasoning、refusal/content filter、音频/图片/代码结果和结构完整的工具调用均可构成有效输出；只有 role、usage、ping、start/stop、空 candidate/choice/content 等协议外壳不算输出。
- OpenAI Chat/Completions 同时校验非流式和流式响应；工具调用必须有函数名，聚合后的 arguments 必须是完整 JSON，并正确统计并行工具调用。
- OpenAI Responses API 校验 `completed`/`incomplete`/`failed` 状态、输出项及 function/custom tool call；流式工具调用按 item id 与 output index 关联，拒绝缺名或参数截断。
- Claude 不再把任意合法 SSE 事件或 `stop_reason=tool_use` 的空 content 当作成功；工具流必须有 `tool_use` 块、完整 JSON 输入以及终止事件。
- Gemini 拒绝空 candidate、空 parts、只有 usage 的流片段和无终止原因的截断流；安全过滤仍按明确拒绝处理，纯 function call 保持有效。
- 在首个语义输出前缓存流式协议外壳，使真正的空回仍可在未写入客户端时触发渠道间重试；一旦响应已经写出则禁止切换渠道，避免把两个上游流拼接到同一客户端响应，并发送对应协议的终止错误事件。
- 每次渠道重试前重置 response count、stream status、首包时间、thinking/Claude 转换状态和 Responses 内置工具计数，防止上一渠道状态污染下一渠道。
- 删除不再使用的 40k stars light 海报 PNG/SVG。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `relay/responsevalidator/validator.go` | OpenAI、Responses、Claude、Gemini 的统一语义与终止状态校验 |
| `relay/channel/openai/relay-openai.go` | Chat/Completions 非流式及 SSE 校验、首个语义输出前缓存 |
| `relay/channel/openai/relay_responses.go` | Responses API 非流式及事件流校验 |
| `relay/channel/claude/relay-claude.go` | Claude 消息及事件流校验 |
| `relay/channel/gemini/relay-gemini.go` | Gemini 转 OpenAI 响应的空回、工具调用及终止校验 |
| `relay/channel/gemini/relay-gemini-native.go` | Gemini 原生响应的对应校验 |
| `controller/relay.go` | 已写响应禁止跨渠道重试，流内返回协议错误 |
| `relay/common/relay_info.go` | 渠道重试前清理响应状态 |
| `relay/helper/common.go` | OpenAI、Responses、Claude SSE 终止错误输出 |
| `relay/responsevalidator/validator_test.go` | 空外壳、纯工具调用、参数截断、过滤和终止状态测试 |

对应提交：

```text
55ac05637 fix(relay): reject semantically empty upstream responses
```

### 14. 音频格式魔数识别与 AMR 时长解析

修复客户端上传内容与文件扩展名不一致时，token 预估阶段按错误格式解析音频并返回 `count_token_failed` 的问题。

保留行为：

- 读取音频魔数识别 WAV、FLAC、Ogg/Opus、MP4/M4A、AIFF、WebM、AAC、MP3、AMR-NB 和 AMR-WB，扩展名只作为无法识别时的回退。
- AMR-NB/AMR-WB 按帧头和每帧 20ms 计算时长，拒绝非法或截断帧。
- 转发 multipart 请求时，依据实际音频格式规范化发往上游的文件扩展名和 `Content-Type`。
- 原始客户端文件名和表单字段只用于客户端侧语义，不因上游规范化而改变。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `common/audio.go` | 音频魔数识别、AMR 时长解析及统一时长入口 |
| `common/audio_test.go` | 格式识别、AMR-NB/AMR-WB 和错误输入回归测试 |
| `relay/channel/openai/adaptor.go` | 上游 multipart 文件名和 MIME 类型规范化 |
| `relay/channel/openai/audio_request_test.go` | 错扩展名、AMR 和客户端表单保持测试 |

## 已移除或不再保留的改动

### Redis Sentinel 支持

已从当前工作区移除：

- `common/redis.go` 中基于 `REDIS_SENTINEL_ADDRS` 和 `REDIS_SENTINEL_MASTER_NAME` 的 Sentinel 初始化。
- `redis.NewFailoverClient` / `redis.FailoverOptions` 使用。
- 仅服务于 Sentinel 配置的 `parseCommaSeparated`、`splitString`、`trimSpace`、`GetEnvOrDefaultInt` 辅助函数。
- `middleware/rate-limit.go` 中 Sentinel failover 重试和 fail-open 逻辑。

移除后，当前工作区中不再有 Redis Sentinel 专属符号：

```text
REDIS_SENTINEL_*
NewFailoverClient
FailoverOptions
MasterName
Sentinel failover
Redis Sentinel
```

普通业务代码中的英文 `sentinel value` 或前端占位常量不属于 Redis Sentinel。

### 模型 reasoning effort 后缀透传

不再保留。此前本地曾经让 `-low`、`-medium`、`-high`、`-minimal`、`-max`、`-xhigh` 等模型名后缀原样透传；该策略已撤回，恢复上游自动解析行为。

恢复上游行为的范围：

- Claude adaptive thinking 转换。
- Gemini thinking level 转换。
- OpenAI Chat / Responses 的 reasoning effort 解析。
- Gemini / Vertex upstream model name 的后缀剥离。

### 过期 token 延期后自动启用

不再保留。此前本地曾经在 token 的 `ExpiredTime` 被延长时自动重新启用过期 token；该逻辑已移除，对应测试也已删除。该能力更适合放在外部发卡/授权流程里，而不是 API gateway 内部。

## 当前工作区状态说明

截至 2026-07-23，Classic 与桌面壳已拆分到独立仓库，本仓库 Electron 已回归上游，音频魔数/AMR 修复及相关测试已纳入提交。
