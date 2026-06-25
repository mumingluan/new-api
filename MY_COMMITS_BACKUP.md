# zch-api 本地修改备份

本文档记录当前 fork 相对上游基点仍然保留的本地改动。

上游基点：

```text
e8c836d70 fix(web): improve form validation error focus #5163
```

更新时间：2026-06-25。已逐个查看当前 fork 独有提交，并按“当前工作区最终仍保留的差异”重新整理。Redis Sentinel 支持和 Sentinel 故障转移重试逻辑已经从当前工作区移除，不再作为保留改动记录。

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

- classic 前端请求限流页面新增密钥分钟级和每日限流配置。
- default 前端新增 `TokenRateLimitSection` 和 `TokenDailyRateLimitSection`。
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

- classic 新建 token 默认 `remain_quota = 500000`，`unlimited_quota = false`。
- default 新建 API key 默认 1 美元额度，并默认不是无限额度。
- classic/default 的 API key 和兑换码表单加入常用美元金额快捷按钮。
- classic token 分组选择恢复自动分组逻辑；当服务端返回 `default_use_auto_group` 时默认选择 `auto`。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `web/classic/src/components/table/tokens/modals/EditTokenModal.jsx` | classic token 默认值、自动分组、金额预设 |
| `web/classic/src/components/table/redemptions/modals/EditRedemptionModal.jsx` | classic 兑换码金额预设 |
| `web/default/src/features/keys/lib/api-key-form.ts` | default API key 默认有限额度 |
| `web/default/src/features/keys/components/api-keys-mutate-drawer.tsx` | default API key 金额预设按钮 |
| `web/default/src/features/redemption-codes/components/redemptions-mutate-drawer.tsx` | default 兑换码金额预设按钮 |

### 9. 控制台登录仅面向管理员的前端提示

保留行为：

- Public header 不再给未登录访客展示登录/注册按钮。
- classic header 未登录状态不再展示登录/注册区域。
- classic/default 登录页增加提示：控制台登录仅供管理员使用，普通用户请访问首页查看教程。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `web/default/src/components/layout/components/public-header.tsx` | 隐藏未登录访客登录入口 |
| `web/default/src/features/auth/sign-in/index.tsx` | default 登录页增加管理员提示 |
| `web/classic/src/components/layout/headerbar/UserArea.jsx` | classic header 隐藏访客登录/注册按钮 |
| `web/classic/src/components/auth/LoginForm.jsx` | classic 登录页增加管理员提示 |
| `web/default/src/i18n/locales/*.json` | 补充提示文案翻译 |

### 10. 前端依赖和 lockfile 更新

保留行为：

- classic/default 前端 lock/package 文件包含本地前端改动所需依赖更新。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `web/classic/package.json` | classic 依赖元数据 |
| `web/classic/bun.lock` | classic Bun lockfile |
| `web/default/bun.lock` | default Bun lockfile |
| `web/package-lock.json` | web workspace npm lockfile |

### 11. 测试稳定性清理

保留行为：

- Channel affinity usage cache 测试改用 atomic counter 生成唯一 key，替代 `time.Now().UnixNano()`，降低并发/快速执行时的 key 碰撞风险。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `service/channel_affinity_usage_cache_test.go` | atomic 唯一测试 key |

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

截至本文档更新时，当前工作区包含三类未提交改动：

- 移除 Redis Sentinel 支持：`common/redis.go`
- 移除 Sentinel failover 重试逻辑：`middleware/rate-limit.go`
- 更新本地修改备份文档：`MY_COMMITS_BACKUP.md`
