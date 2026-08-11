# qiqi-api 本地变更备份

本文档记录当前 fork 相对上游基点仍然保留的本地改动。

上游基点：

```text
172114422 fix(auth): keep login state on rate-limited or failing token refresh
```

更新时间：2026-08-11。已逐个查看当前 fork 独有提交及当前工作区未提交改动，并按“当前工作区最终仍保留的差异”重新整理。本文档只记录本仓库内保留的源码差异，不记录外部部署过程或其他项目的构建流程。Redis Sentinel 支持和 Sentinel 故障转移重试逻辑已经从当前工作区移除，不再作为保留改动记录。

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

### 10. 测试稳定性清理

保留行为：

- Channel affinity usage cache 测试改用 atomic counter 生成唯一 key，替代 `time.Now().UnixNano()`，降低并发/快速执行时的 key 碰撞风险。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `service/channel_affinity_usage_cache_test.go` | atomic 唯一测试 key |

### 11. 上游响应语义校验与安全重试

修复 OpenAI Chat/Completions/Responses、Claude 和 Gemini 上游返回 HTTP 200、JSON/SSE 外壳合法但没有任何可消费语义输出时，被误当成成功响应并计费的问题，同时覆盖纯工具调用和流式边缘场景。

保留行为：

- 统一按“语义输出”而不是响应体非空判断成功：文本、reasoning、refusal/content filter、音频/图片/代码结果和结构完整的工具调用均可构成有效输出；只有 role、usage、ping、start/stop、空 candidate/choice/content 等协议外壳不算输出。
- OpenAI Chat/Completions 同时校验非流式和流式响应；工具调用必须有函数名，聚合后的 arguments 必须是完整 JSON，并正确统计并行工具调用。
- OpenAI Responses API 校验 `completed`/`incomplete`/`failed` 状态、输出项及 function/custom tool call；流式工具调用按 item id 与 output index 关联，拒绝缺名或参数截断。
- Claude 不再把任意合法 SSE 事件或 `stop_reason=tool_use` 的空 content 当作成功；工具流必须有 `tool_use` 块和完整 JSON 输入。缺少终止事件但已经产生语义输出的异常流按第 20 节处理。
- Gemini 拒绝空 candidate、空 parts 和只有 usage 的流片段；安全过滤仍按明确拒绝处理，纯 function call 保持有效。缺少终止原因但已经产生语义输出的异常流按第 20 节处理。
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

### 12. 音频格式魔数识别与 AMR 时长解析

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

### 13. 用户与 API Key 额度原子扣减

修复普通 Chat Completions、旧版按次接口和异步任务等计费路径在并发请求下先检查余额、再无条件扣减，可能导致 API Key 或用户余额变成负数的问题。

保留行为：

- 用户钱包扣减使用带 `quota >= 扣费额` 条件的单条数据库更新；并发请求只有余额充足的请求能够成功扣款。
- 有限额度 API Key 使用带 `remain_quota >= 扣费额` 条件的单条数据库更新，扣减与使用量累计在同一条更新中完成。
- 无限额度 API Key 只累计 `used_quota`，不再扣减 `remain_quota`；退款时同样保持剩余额度不变。
- 额度增减不再进入延迟批量更新队列，数据库余额立即生效；Redis 中的 API Key 缓存在变更成功后失效，避免继续读取旧余额。
- 高于信任额度阈值的请求不再走免预扣旁路，固定高价模型会先预扣完整费用。
- 钱包计费预检强制读取数据库实时余额，原子扣减失败统一返回额度不足。
- 使用余额购买订阅时增加原子余额条件，避免并发购买超扣。
- 管理员覆盖用户额度时拒绝负数，并在成功覆盖后清理用户缓存。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `model/user.go` | 用户额度即时更新和原子非负扣减 |
| `model/token.go` | 有限/无限 API Key 的原子扣减、退款与缓存失效 |
| `model/utils.go` | 用户和 API Key 额度不足哨兵错误 |
| `model/subscription.go` | 余额购买订阅的原子扣减保护 |
| `service/billing_session.go` | 高价请求预扣保护、实时钱包余额读取和错误映射 |
| `controller/user.go` | 拒绝管理员将用户额度覆盖为负数 |

测试：

| 文件 | 说明 |
| ---- | ---- |
| `model/quota_atomic_test.go` | 用户/API Key 并发扣减、无限额度扣费与退款 |
| `service/billing_session_trust_test.go` | 高于信任阈值的费用必须预扣 |
| `controller/user_manage_test.go` | 管理员负数额度覆盖必须拒绝 |

验证：

- Go 后端各包测试通过。
- Windows amd64 和 Linux amd64 编译通过。

### 14. Gemini 转 Claude 工具调用兼容

修复 OpenAI Chat Completions 经 Gemini 兼容接口转发到 Claude 模型时，工具 schema 校验失败、历史工具调用 ID 错位，以及流式工具参数被拆分后丢失的问题。

保留行为：

- 原生 Gemini 请求发送前递归清理函数声明中的非 Gemini Schema 字段；字符串 `const` 转换为单元素 `enum`，不支持的排他边界直接移除，避免改变边界语义。
- Claude 映射模型使用保守的工具输入 schema，并保证无参数工具也带有根级 `"type": "object"`。
- Claude 映射链路中的历史工具调用与结果转为文本上下文，规避中转服务重新编号 `tool_use_id` 后调用与结果无法配对。
- 兼容中转服务将一次 Gemini `functionCall` 拆成“函数名空参数”和“空函数名完整参数”两个 SSE 片段的行为；首段先缓存，收到参数后以同一调用 ID 输出完整 OpenAI 工具调用。
- 通用 Gemini 转换继续使用原有 Schema 类型格式，Claude 专用路径才输出 JSON Schema Draft 2020-12 所需的小写类型，避免影响其他 Gemini 渠道。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `relay/channel/gemini/adaptor.go` | 原生及批量 Gemini 工具 schema 清理入口 |
| `relay/channel/gemini/relay-gemini.go` | 分段流式工具调用缓存、合并和稳定 ID |
| `service/relayconvert/internal/oai_chat/to_gemini_chat_req.go` | Claude schema 兼容及历史工具消息处理 |
| `service/relayconvert/internal/gemini_chat/to_oai_chat_resp.go` | Gemini 工具调用 ID 向 OpenAI 响应透传 |
| `service/relayconvert/internal/shared/gemini/schema.go` | Gemini/Claude 工具 schema 清理规则 |
| `dto/gemini.go` | Gemini function call ID 字段 |

验证：

- Gemini relay、请求/响应转换、共享 schema 和响应语义校验测试通过。
- 相关包 `go vet` 通过。
- Windows amd64 和 Linux amd64 编译通过。
- OpenAI 流式工具调用端到端验证可得到完整参数而非空对象。

### 15. Xuancat 原版风格主页和密钥工具

在核心前端中集成 Xuancat 密钥工具，同时继续使用原版主页和原版关于页。页面使用
核心前端现有的 Base UI/shadcn 组件、主题变量和响应式断点，覆盖手机、平板和桌面布局。

保留行为：

- 通过编译时变量 `VITE_LANDING_PAGE_VARIANT` 切换页面：
  - `xuancat`：使用之前的“全标准化、中立的 LLM 接入”Hero 文案，隐藏常用应用支持和 Hero API 调用演示，并增加 Xuancat 密钥工具；
  - `default` 或未设置：使用完全原版主页。关于页始终使用原版实现。
- Xuancat 模式主页 Hero 顶部标识使用当前实例的 `system_name`，不再固定显示
  “玄喵大模型 HUB”；普通模式的原版标识保持不变。
- Xuancat 模式主页 Hero 操作区只保留“文档”和“模型列表”：文档按钮使用蓝色主按钮并排在最前，模型列表继续进入 `/pricing`；未登录时不再显示“开始使用”，登录后也不再显示“前往仪表板”。普通模式主页按钮保持原有行为。
- 内置版不提供服务器选择，激活、续期、激活码查询、令牌查询均使用当前域名下
  的 `/api/activation/*`、`/v1/dashboard/billing/*` 和 `/api/log/token`。
- 令牌查询展示令牌名称、总额度、剩余额度、已用额度、过期时间和近期调用；
  即使令牌尚无调用日志，也可从订阅接口读取令牌名称。近期调用同时包含消费日志和错误日志，展示成功/错误状态与完整错误详情，便于用户直接定位失败请求。
- 页面文案覆盖 en、zh、zh-TW、fr、ja、ru、vi，并将内部 `zhCN` 语言代码映射为
  合法的 Intl locale，避免日期格式化异常；密钥查询按钮在各非英语语言中使用本地化动作文案。
- 旧版 Xuancat 独立主页和关于页的 41 个无引用 i18n 键已从全部语言包移除。
- Dockerfile 同样暴露 `VITE_LANDING_PAGE_VARIANT` 构建参数；NewAPI 通过
  `go:embed web/dist` 嵌入前端，因此切换后必须先重新构建前端，再编译 Go 二进制。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `web/src/features/xuancat-pages/` | 密钥开通、续期、查询功能和主页工具面板 |
| `web/src/features/home/components/sections/hero.tsx` | 挂载 Xuancat 密钥工具，并在 Xuancat 模式显示当前实例名称 |
| `web/src/features/home/components/xuancat-hero-actions.ts` | 固定 Xuancat Hero 操作项及顺序 |
| `web/src/features/home/components/__tests__/hero-actions.test.ts` | Xuancat Hero 操作项顺序和精简范围回归测试 |
| `web/src/lib/landing-page-variant.ts` | 判断是否启用 Xuancat 专属功能 |
| `web/src/i18n/locales/*.json` | “模型列表”等 Xuancat 主页文案翻译 |
| `web/rsbuild.config.ts` | 读取并注入 `VITE_LANDING_PAGE_VARIANT` |
| `web/.env.example` | 开关默认值示例 |
| `Dockerfile` | Docker 构建参数透传 |
| `controller/billing.go` | 订阅查询响应返回当前令牌名称 |
| `controller/channel-billing.go` | 订阅响应的 `token_name` 字段 |

### 16. Granter 激活码整合到 NewAPI

将 Granter 的激活、续期、查询、激活码管理和使用记录整合进 NewAPI。普通用户可在
Xuancat 前端“常规 / 激活码管理”中维护自己名下的激活码；普通版前端不显示该入口。

主要行为：

- 新激活码统一使用 `USERID_ACTIVATIONCODE` 格式，用户只能创建自己 ID 前缀的码；
  自动生成时使用 16 位随机后缀，例如 `1_Q68AT2NDBJES11OM`。后端优先按明文 ID
  查询，同时兼容历史非标准码。
- 激活码领取、状态占用、API Key 创建和使用记录写入处于同一数据库事务，避免并发
  重复兑换；生成的 API Key 归激活码创建者所有。
- 支持高级筛选、响应式桌面表格和移动端卡片、批量复制、CSV 导出、批量创建、
  批量修改、按激活码批量删除及使用记录查询。
- 页面使用 NewAPI 的 Base UI 选择器、日期时间范围选择器和日期时间选择器；
  固定内容区支持纵向滚动，批量创建、批量管理和删除弹窗位于布局槽外并可正常打开。
- 内置 Xuancat 主页改用 `/api/activation/*`，不再依赖独立 Granter HTTP 服务。
- 旧 Granter 数据通过 `scripts/migrate-granter-activation-codes.sql` 原样迁移，
  保留激活码文本、所属用户和历史使用记录。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `model/activation_code.go` | 数据模型、所有者解析、原子激活/续期及用户级管理 |
| `controller/activation_code.go` | 公开激活接口和登录用户管理接口 |
| `router/api-router.go` | 公开限流路由及非管理员用户管理路由 |
| `web/src/features/activation-codes/` | Xuancat 专属激活码管理页 |
| `web/src/features/xuancat-pages/api.ts` | 内置主页调用 NewAPI 激活接口 |
| `scripts/migrate-granter-activation-codes.sql` | Granter 到 NewAPI 的幂等迁移脚本 |

验证：

- 激活码并发兑换回归测试、相关 Go 测试与 `go vet` 通过。
- 前端类型检查、定向 lint、主页 API 回归测试和 Xuancat 生产构建通过。
- 激活码与使用记录两个工作区固定从标签栏下方开始，避免全高 Grid 把内容纵向居中。
- Xuancat 构建把“激活码管理”纳入管理员和用户侧边栏模块设置；关闭对应开关会隐藏 `/activation-codes` 导航，普通构建不显示该专属设置项。

### 17. `/457` 标记

新增无需认证的独立 `GET /457` 标记路由，固定返回 `{"457":true}`，供客户端严格按
布尔值识别当前实例是否提供 Xuancat 专属前端能力。

### 18. Xuancat 用户级密钥批量操作与日志统计

在 Xuancat 模式的“激活码管理”下新增“密钥批量操作”，并接入管理员和用户侧边栏模块设置。普通用户始终只处理自己的密钥；管理员默认同样仅处理本人密钥，只有主动开启“所有用户”后才会查询或操作全站密钥，普通用户伪造全站参数会被后端拒绝。

批量操作支持延长/扣除有效期、增加/扣除额度，可按已使用状态和最低剩余额度筛选，提供操作预览、确认弹窗和审计记录。日志统计支持日期范围、密钥/模型/用户名/渠道/用户 ID 分组、排序、用户包含/排除、模型模糊查询、最小 Token、Top N、独立用户数、全量合计和 CSV 导出。

同时保留 Xuancat 主页客户端说明文案：OpenAI 兼容客户端可直接使用该地址，也支持 Gemini 或 Claude 格式调用。

补充实现范围：

| 文件 | 说明 |
| ---- | ---- |
| `controller/key_batch.go` | 用户/管理员作用域校验、预览、执行和日志统计接口 |
| `model/key_batch.go` | 跨数据库密钥筛选、批量更新和消费日志聚合 |
| `service/key_batch.go` | 批量操作计划、额度及有效期计算、审计详情生成 |
| `controller/audit.go` | 密钥批量操作审计事件类型 |
| `router/api-router.go` | 登录用户接口和管理员全用户操作入口 |
| `web/src/features/key-batch-operations/` | 批量操作、统计筛选、统计表格和 CSV 导出页面 |
| `web/src/hooks/use-sidebar-data.ts` | Xuancat 侧边栏中的“密钥批量操作”导航 |
| `web/src/features/profile/components/sidebar-modules-card.tsx` | 用户侧边栏模块开关 |
| `web/src/features/system-settings/maintenance/` | 管理员侧边栏模块配置 |

此前的 Xuancat 主页还原、演示模块隐藏、密钥查询本地化、废弃 i18n 清理、激活码工作区顶部对齐以及激活码导航开关，分别已归入第 15、16 节，不再遗漏为未记录差异。

### 19. 移动端统计表格边界

密钥批量操作的日志统计表格在移动端会把 Grid/Flex 项目的最小内容宽度逐级传到页面容器，导致整个页面视窗被表格列宽撑大。现在为页面滚动区、Tabs 根节点、Tabs 面板、操作/统计内容区和结果卡片补齐 `min-w-0`、`max-w-full` 约束，页面最外层禁止横向溢出；表格仍保留自身的 `overflow-x-auto`，因此窄屏只在表格内部横向滚动，不再扩大页面宽度。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `web/src/features/key-batch-operations/index.tsx` | 收紧移动端容器最小宽度并隔离统计表格横向滚动 |

验证：

- `bun run typecheck` 通过。
- `bunx oxlint src/features/key-batch-operations/index.tsx` 通过。
- Xuancat 与普通前端生产构建均通过。
- Xuancat Windows amd64、Xuancat Linux amd64、普通版 Linux amd64 二进制均编译成功。

### 20. 非空异常流按完成请求计费

调整 OpenAI Chat/Completions、OpenAI Responses、Claude 和 Gemini 的流式响应终止策略：终止事件完整性与语义输出有效性分开判断，避免上游已经返回有效内容、但连接在终止帧前断开时整单退款。

保留行为：

- “空回”继续沿用统一语义校验定义：只有 role、usage、ping、start/stop、空 choice/candidate/content 等协议外壳不算输出；文本、reasoning、refusal/过滤结果、音频/图片/代码结果和结构完整的工具调用才算有效输出。
- 真正空回仍返回 `ErrorCodeEmptyResponse`，进入重试/错误处理并退还预扣额度。
- 已经产生语义输出但缺少 `finish_reason`、`[DONE]`、`message_stop`、`response.completed`、`response.incomplete` 或 Gemini terminal finish reason 时，请求按完成处理并进入正常 usage 计算和计费结算。
- 非空异常流不会生成普通请求错误日志；消费日志通过 `other.stream_status` 保留 `status=error`、`end_reason`、`end_error` 和软错误信息，因此前端仍会标记“流异常”。
- `client_gone`、timeout、scanner error、EOF 等由现有 `StreamStatus` 记录；只要已经获得有效语义输出且响应解析本身没有硬错误，就不会因缺少终止事件退款。
- 不完整 JSON 工具参数、缺失工具名、无法解析的 chunk 或明确的上游错误仍按硬错误处理，不会仅因为此前出现过文本就绕过协议校验。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `relay/responsevalidator/validator.go` | 将语义输出校验拆分为 `ValidateOutput`，终止完整性继续由完整 `Validate` 检查 |
| `relay/channel/openai/relay-openai.go` | Chat/Completions 非空截断流继续估算 usage、返回成功并记录流异常 |
| `relay/channel/openai/relay_responses.go` | Responses 非空截断流继续结算并记录缺失 terminal event |
| `relay/channel/claude/relay-claude.go` | Claude 非空截断流使用现有 fallback usage 结算并记录缺失 `message_stop` |
| `relay/channel/gemini/relay-gemini.go` | Gemini 已发送有效响应但无终止原因时继续结算并记录流异常 |
| `relay/channel/openai/chat_stream_test.go` | OpenAI 非空 EOF 正常返回 usage、空流仍拒绝的端到端回归测试 |
| `relay/responsevalidator/validator_test.go` | 语义输出有效性与终止完整性分离测试 |

验证：

- `go test ./relay/...` 通过。
- OpenAI、Responses、Claude、Gemini 和 billing service 定向测试通过。

### 21. Claude 服务端工具流参数校验

修复 Claude Web Search、Code Execution 等服务端工具调用在经过 relay 时被提前截断的问题。Anthropic 的 `server_tool_use` 与普通 `tool_use` 一样，会在 `content_block_start` 后通过 `input_json_delta` 分片发送参数；旧校验器只登记普通工具，因此会把服务端工具的第一个参数分片误判为“参数早于工具块”，随后终止 SSE。

保留行为：

- `server_tool_use` 与普通 `tool_use` 都按 content block index 登记名称和参数流。
- Web Search 和 Code Execution 的空起始分片及后续 JSON 参数分片能够完整通过校验，不再丢失工具结果、最终文本和 `message_stop`。
- 服务端工具缺少名称或累计参数不是完整 JSON 时仍然返回协议错误，不放宽原有响应完整性保护。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `relay/responsevalidator/validator.go` | 将 `server_tool_use` 纳入流式工具参数关联和完整性校验 |
| `relay/responsevalidator/validator_test.go` | Web Search、Code Execution 服务端工具参数流回归测试 |

验证：

- `go test ./relay/responsevalidator ./relay/channel/claude` 通过。
- `go test ./relay/...` 通过。

### 22. 模型广场动态端点类型筛选

修复模型广场仅渲染前端预定义端点类型，导致数据库中已经配置但前端尚未登记的端点类型无法筛选、数量显示为零的问题。

保留行为：

- 端点类型筛选项从模型接口返回的 `supported_endpoint_types` 动态生成，只显示当前模型列表实际存在的端点。
- 已预定义的标准端点继续使用 Chat、Rerank、Embeddings 等友好名称。
- 未预定义的端点直接使用数据库返回的原始名称显示，并正常参与模型数量统计和筛选。
- 同一模型重复声明相同端点时只计数一次；空端点名称会被忽略。
- Veo 继续按现有非标准调用方式处理，本次不修改其端点类型配置。

主要文件：

| 文件 | 说明 |
| ---- | ---- |
| `web/src/features/pricing/components/pricing-sidebar.tsx` | 使用模型列表动态生成端点筛选项 |
| `web/src/features/pricing/lib/endpoint-types.ts` | 动态端点去重、排序、计数和显示名称回退 |
| `web/src/features/pricing/lib/__tests__/endpoint-types.test.ts` | 未知端点原名显示、计数和空值回归测试 |
| `web/src/features/pricing/constants.ts` | 允许动态端点名称访问已有友好标签映射 |

验证：

- 端点筛选回归测试通过。
- 前端类型检查和定向 lint 通过。
- Xuancat 前端生产构建通过。
