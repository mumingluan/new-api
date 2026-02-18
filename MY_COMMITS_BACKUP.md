# 我的自定义提交备份

> 生成时间: 2026-03-17
> 分支: main (origin/main)
> 基于上游: upstream/main (merge-base: f77381cc)
> 共 10 个自定义提交（按时间正序排列）

---

## 提交概览

| # | 提交哈希 | 日期 | 说明 |
|---|---------|------|------|
| 1 | `c3eb6bab` | 2026-01-08 23:34:55 | feat: add multi-dimensional rate limiting and token default changes |
| 2 | `4798f8c6` | 2026-01-08 23:54:57 | fix(gemini): 优化无候选返回时的错误处理逻辑 |
| 3 | `d807d725` | 2026-01-08 23:58:39 | chore: 删除不再需要的上游新API子项目 |
| 4 | `59e2cc89` | 2026-01-28 23:41:06 | feat: 增强流式响应处理，添加拒绝原因记录和有效性检查 |
| 5 | `5fb49e4c` | 2026-01-29 00:27:32 | feat: 添加令牌速率限制设置和格式化功能 |
| 6 | `55e2cb8e` | 2026-02-01 11:12:19 | feat: 更新限流逻辑，添加检查功能并优化过期时间设置 |
| 7 | `a566bdbe` | 2026-02-18 22:18:09 | feat: 更新适配器逻辑，支持Gemini格式的请求和响应处理，并优化错误检查 |
| 8 | `6917dbcd` | 2026-02-18 21:30:16 | feat: 更新视频代理逻辑，支持直接返回内联 base64 视频数据并优化错误处理 |
| 9 | `a378dbe8` | 2026-03-05 12:07:10 | feat: 更新Gemini和OpenAI适配器逻辑，优化空响应检查和工具调用处理 |
| 10 | `87065f8f` | 2026-03-17 22:24:35 | feat: 更新EditRedemptionModal和EditTokenModal，添加新的金额选项 |

---

## 功能分类总结

### 1. 多维度限流系统（Per-Key Rate Limiting）
**涉及提交**: c3eb6bab, 5fb49e4c, 55e2cb8e, 6917dbcd

在原有的 per-user 模型请求限流基础上，新增了按 API 密钥（Token）维度的限流：
- **密钥分钟级限流**: 每个 API 密钥在 N 分钟内最多请求 X 次
- **密钥每日限流**: 每个 API 密钥每天最多请求 X 次（北京时间自然日）
- **分组配置**: 支持按分组设置不同的限流阈值
- **Redis + 内存双模式**: 同时支持 Redis 和内存存储
- **前端设置页面**: 在速率限制设置页新增密钥限流配置区域
- **Token 默认值变更**: remain_quota=500000, unlimited_quota=false
- **自动分组选择**: 恢复新建 Token 时的自动分组选择逻辑
- **限流 Bug 修复**: 修复 Redis 过期时间使用固定值的问题，改为使用 duration 参数；修复内存限流的成功请求检查逻辑（Check vs Request）；修复 Lua 脚本中过期时间被注释的问题
- **每日限流优化**: 使用北京时间自然日计算 key，Lua 原子脚本实现预占+回滚机制

**涉及文件**:
-  — 核心限流中间件
-  — 限流配置变量和函数
-  — 选项初始化和更新
-  — 内存限流器新增 Check 方法
-  — Lua 限流脚本修复
-  — 前端设置页
-  — 前端设置组件
-  — Token 默认值

### 2. 流式响应空响应检测与重试（Empty Response Detection）
**涉及提交**: 59e2cc89, a378dbe8

为 Claude、Gemini、OpenAI 三大适配器添加空响应检测，触发自动重试：
- **Claude**: 流式响应添加 hasValidResponse 标记，非流式检查 Content 和 Completion 是否为空
- **Gemini**: 非流式检查 Candidates 为空时区分 BlockReason 和空响应；流式检查 TotalTokens/CompletionTokens
- **OpenAI**: 流式检查 streamItems 为空；非流式检查 Choices 为空（仅 ChatCompletions/Completions 模式）
- **拒绝原因记录**: 为各适配器添加 reject reason 上下文记录

### 3. Gemini 格式适配器支持（Claude/Ollama → Gemini Format）
**涉及提交**: a566bdbe, a378dbe8

- **Claude 适配器**: 实现 ConvertGeminiRequest（通过 GeminiToOpenAI 转换）；流式/非流式响应支持 Gemini 格式输出
- **Ollama 适配器**: 实现 ConvertGeminiRequest（通过 GeminiToOpenAI 转换）
- **OpenAI 适配器**: Embeddings 模式使用 OpenaiHandlerWithUsage；空响应检查增加 RelayMode 判断
- **Jina 适配器**: Embeddings 模式改用 OpenaiHandlerWithUsage
- **工具调用 Token 计数**: OpenAI 非流式响应中，completionTokens 为 0 时将工具调用的 name+arguments 纳入计数
- **Gemini 流式工具调用**: 统计 FunctionCall 的 name 和 arguments 到 responseText

### 4. Gemini 视频代理优化（Inline Base64 Video）
**涉及提交**: 6917dbcd

- 重构 getGeminiVideoURL 为 getGeminiVideoResult，返回 geminiVideoResult 结构体
- 支持 Veo 返回的 bytesBase64Encoded 内联视频数据，直接写入响应
- Gemini Task Adaptor: FetchTask 改用 fetchPredictOperation POST 接口

### 5. 前端 UI 调整
**涉及提交**: 87065f8f, a378dbe8

- **金额选项扩展**: EditRedemptionModal 和 EditTokenModal 添加 15$/25$/40$/55$/65$/80$/265$/535$ 等选项
- **登录页提示**: 添加「控制台登录仅供管理员使用」提示文字
- **移除未登录按钮**: UserArea 组件未登录时返回 null（不显示登录/注册按钮）

### 6. 杂项
**涉及提交**: 4798f8c6, d807d725

- 删除 upstream-new-api 子模块引用（无实际代码变更）

---

## 各提交完整 Diff

### 提交 1: feat: add multi-dimensional rate limiting and token default changes

- **哈希**: `c3eb6bab1f47c41d60fcdfedafeac3409db6c7b2`
- **日期**: 2026-01-08 23:34:55 +0800
- **作者**: Mingluan Mu <mumingluan@qq.com>

**完整 Diff:**

```diff
diff --git a/middleware/model-rate-limit.go b/middleware/model-rate-limit.go
index 80a3995d..49d9d075 100644
--- a/middleware/model-rate-limit.go
+++ b/middleware/model-rate-limit.go
@@ -21,6 +21,14 @@ const (
 	ModelRequestRateLimitSuccessCountMark = "MRRLS"
 )
 
+// Token rate limit constants
+const (
+	TokenRateLimitCountMark             = "TRL"
+	TokenRateLimitSuccessCountMark      = "TRLS"
+	TokenDailyRateLimitCountMark        = "TDRL"
+	TokenDailyRateLimitSuccessCountMark = "TDRLS"
+)
+
 // 检查Redis中的请求限制
 func checkRedisRateLimit(ctx context.Context, rdb *redis.Client, key string, maxCount int, duration int64) (bool, error) {
 	// 如果maxCount为0，表示不限制
@@ -163,12 +171,325 @@ func memoryRateLimitHandler(duration int64, totalMaxCount, successMaxCount int)
 	}
 }
 
+// checkTokenRateLimit 检查 token 分钟级限流
+func checkTokenRateLimit(c *gin.Context) bool {
+	if !setting.TokenRateLimitEnabled {
+		return true
+	}
+
+	tokenId := common.GetContextKeyInt(c, constant.ContextKeyTokenId)
+	if tokenId == 0 {
+		// 如果没有 token ID，跳过 per-key 限流
+		return true
+	}
+
+	// 获取分组配置（使用 token group）
+	group := common.GetContextKeyString(c, constant.ContextKeyTokenGroup)
+	totalMaxCount := setting.TokenRateLimitCount
+	successMaxCount := setting.TokenRateLimitSuccessCount
+
+	// 获取分组的限流配置
+	groupTotalCount, groupSuccessCount, found := setting.GetTokenRateLimit(group)
+	if found {
+		totalMaxCount = groupTotalCount
+		successMaxCount = groupSuccessCount
+	}
+
+	// 如果两个限制都为0，表示不限制
+	if totalMaxCount == 0 && successMaxCount == 0 {
+		return true
+	}
+
+	rateLimitKey := strconv.Itoa(tokenId)
+	duration := int64(setting.TokenRateLimitDurationMinutes * 60)
+
+	if common.RedisEnabled {
+		return checkTokenRateLimitRedis(c, rateLimitKey, totalMaxCount, successMaxCount, duration)
+	} else {
+		return checkTokenRateLimitMemory(c, rateLimitKey, totalMaxCount, successMaxCount, duration)
+	}
+}
+
+// checkTokenRateLimitRedis Redis版本的分钟级限流检查
+func checkTokenRateLimitRedis(c *gin.Context, rateLimitKey string, totalMaxCount, successMaxCount int, duration int64) bool {
+	ctx := context.Background()
+	rdb := common.RDB
+
+	// 1. 检查成功请求数限制
+	if successMaxCount > 0 {
+		successKey := fmt.Sprintf("rateLimit:%s:%s", TokenRateLimitSuccessCountMark, rateLimitKey)
+		allowed, err := checkRedisRateLimit(ctx, rdb, successKey, successMaxCount, duration)
+		if err != nil {
+			fmt.Println("检查密钥成功请求数限制失败:", err.Error())
+			abortWithOpenAiMessage(c, http.StatusInternalServerError, "rate_limit_check_failed")
+			return false
+		}
+		if !allowed {
+			abortWithOpenAiMessage(c, http.StatusTooManyRequests, fmt.Sprintf("您已达到密钥请求数限制：%d分钟内最多请求%d次", setting.TokenRateLimitDurationMinutes, successMaxCount))
+			return false
+		}
+	}
+
+	// 2. 检查总请求数限制
+	if totalMaxCount > 0 {
+		totalKey := fmt.Sprintf("rateLimit:%s:%s", TokenRateLimitCountMark, rateLimitKey)
+		tb := limiter.New(ctx, rdb)
+		allowed, err := tb.Allow(
+			ctx,
+			totalKey,
+			limiter.WithCapacity(int64(totalMaxCount)*duration),
+			limiter.WithRate(int64(totalMaxCount)),
+			limiter.WithRequested(duration),
+		)
+
+		if err != nil {
+			fmt.Println("检查密钥总请求数限制失败:", err.Error())
+			abortWithOpenAiMessage(c, http.StatusInternalServerError, "rate_limit_check_failed")
+			return false
+		}
+
+		if !allowed {
+			abortWithOpenAiMessage(c, http.StatusTooManyRequests, fmt.Sprintf("您已达到密钥总请求数限制：%d分钟内最多请求%d次（包括失败请求）", setting.TokenRateLimitDurationMinutes, totalMaxCount))
+			return false
+		}
+	}
+
+	return true
+}
+
+// checkTokenRateLimitMemory 内存版本的分钟级限流检查
+func checkTokenRateLimitMemory(c *gin.Context, rateLimitKey string, totalMaxCount, successMaxCount int, duration int64) bool {
+	inMemoryRateLimiter.Init(time.Duration(setting.TokenRateLimitDurationMinutes) * time.Minute)
+
+	totalKey := TokenRateLimitCountMark + rateLimitKey
+	successKey := TokenRateLimitSuccessCountMark + rateLimitKey
+
+	// 1. 检查总请求数限制
+	if totalMaxCount > 0 && !inMemoryRateLimiter.Request(totalKey, totalMaxCount, duration) {
+		abortWithOpenAiMessage(c, http.StatusTooManyRequests, fmt.Sprintf("您已达到密钥总请求数限制：%d分钟内最多请求%d次（包括失败请求）", setting.TokenRateLimitDurationMinutes, totalMaxCount))
+		return false
+	}
+
+	// 2. 检查成功请求数限制（使用临时key检查）
+	if successMaxCount > 0 {
+		checkKey := successKey + "_check"
+		if !inMemoryRateLimiter.Request(checkKey, successMaxCount, duration) {
+			abortWithOpenAiMessage(c, http.StatusTooManyRequests, fmt.Sprintf("您已达到密钥请求数限制：%d分钟内最多请求%d次", setting.TokenRateLimitDurationMinutes, successMaxCount))
+			return false
+		}
+	}
+
+	return true
+}
+
+// recordTokenRateLimitSuccess 记录分钟级成功请求
+func recordTokenRateLimitSuccess(c *gin.Context) {
+	if !setting.TokenRateLimitEnabled {
+		return
+	}
+
+	tokenId := common.GetContextKeyInt(c, constant.ContextKeyTokenId)
+	if tokenId == 0 {
+		return
+	}
+
+	// 获取分组配置
+	group := common.GetContextKeyString(c, constant.ContextKeyTokenGroup)
+	successMaxCount := setting.TokenRateLimitSuccessCount
+
+	_, groupSuccessCount, found := setting.GetTokenRateLimit(group)
+	if found {
+		successMaxCount = groupSuccessCount
+	}
+
+	if successMaxCount == 0 {
+		return
+	}
+
+	rateLimitKey := strconv.Itoa(tokenId)
+
+	if common.RedisEnabled {
+		ctx := context.Background()
+		rdb := common.RDB
+		successKey := fmt.Sprintf("rateLimit:%s:%s", TokenRateLimitSuccessCountMark, rateLimitKey)
+		recordRedisRequest(ctx, rdb, successKey, successMaxCount)
+	} else {
+		duration := int64(setting.TokenRateLimitDurationMinutes * 60)
+		successKey := TokenRateLimitSuccessCountMark + rateLimitKey
+		inMemoryRateLimiter.Request(successKey, successMaxCount, duration)
+	}
+}
+
+// checkTokenDailyRateLimit 检查 token 每日限流
+func checkTokenDailyRateLimit(c *gin.Context) bool {
+	if !setting.TokenDailyRateLimitEnabled {
+		return true
+	}
+
+	tokenId := common.GetContextKeyInt(c, constant.ContextKeyTokenId)
+	if tokenId == 0 {
+		// 如果没有 token ID，跳过 per-key 限流
+		return true
+	}
+
+	// 获取分组配置
+	group := common.GetContextKeyString(c, constant.ContextKeyTokenGroup)
+	totalMaxCount := setting.TokenDailyRateLimitCount
+	successMaxCount := setting.TokenDailyRateLimitSuccessCount
+
+	// 获取分组的限流配置
+	groupTotalCount, groupSuccessCount, found := setting.GetTokenDailyRateLimit(group)
+	if found {
+		totalMaxCount = groupTotalCount
+		successMaxCount = groupSuccessCount
+	}
+
+	// 如果两个限制都为0，表示不限制
+	if totalMaxCount == 0 && successMaxCount == 0 {
+		return true
+	}
+
+	rateLimitKey := strconv.Itoa(tokenId)
+	duration := int64(86400) // 24小时 = 86400秒
+
+	if common.RedisEnabled {
+		return checkTokenDailyRateLimitRedis(c, rateLimitKey, totalMaxCount, successMaxCount, duration)
+	} else {
+		return checkTokenDailyRateLimitMemory(c, rateLimitKey, totalMaxCount, successMaxCount, duration)
+	}
+}
+
+// checkTokenDailyRateLimitRedis Redis版本的每日限流检查
+func checkTokenDailyRateLimitRedis(c *gin.Context, rateLimitKey string, totalMaxCount, successMaxCount int, duration int64) bool {
+	ctx := context.Background()
+	rdb := common.RDB
+
+	// 1. 检查成功请求数限制
+	if successMaxCount > 0 {
+		successKey := fmt.Sprintf("rateLimit:%s:%s", TokenDailyRateLimitSuccessCountMark, rateLimitKey)
+		allowed, err := checkRedisRateLimit(ctx, rdb, successKey, successMaxCount, duration)
+		if err != nil {
+			fmt.Println("检查每日成功请求数限制失败:", err.Error())
+			abortWithOpenAiMessage(c, http.StatusInternalServerError, "rate_limit_check_failed")
+			return false
+		}
+		if !allowed {
+			abortWithOpenAiMessage(c, http.StatusTooManyRequests, "您已达到每日请求数限制")
+			return false
+		}
+	}
+
+	// 2. 检查总请求数限制
+	if totalMaxCount > 0 {
+		totalKey := fmt.Sprintf("rateLimit:%s:%s", TokenDailyRateLimitCountMark, rateLimitKey)
+		tb := limiter.New(ctx, rdb)
+		allowed, err := tb.Allow(
+			ctx,
+			totalKey,
+			limiter.WithCapacity(int64(totalMaxCount)*duration),
+			limiter.WithRate(int64(totalMaxCount)),
+			limiter.WithRequested(duration),
+		)
+
+		if err != nil {
+			fmt.Println("检查每日总请求数限制失败:", err.Error())
+			abortWithOpenAiMessage(c, http.StatusInternalServerError, "rate_limit_check_failed")
+			return false
+		}
+
+		if !allowed {
+			abortWithOpenAiMessage(c, http.StatusTooManyRequests, "您已达到每日总请求数限制（包括失败请求）")
+			return false
+		}
+	}
+
+	return true
+}
+
+// checkTokenDailyRateLimitMemory 内存版本的每日限流检查
+func checkTokenDailyRateLimitMemory(c *gin.Context, rateLimitKey string, totalMaxCount, successMaxCount int, duration int64) bool {
+	inMemoryRateLimiter.Init(24 * time.Hour)
+
+	totalKey := TokenDailyRateLimitCountMark + rateLimitKey
+	successKey := TokenDailyRateLimitSuccessCountMark + rateLimitKey
+
+	// 1. 检查总请求数限制
+	if totalMaxCount > 0 && !inMemoryRateLimiter.Request(totalKey, totalMaxCount, duration) {
+		abortWithOpenAiMessage(c, http.StatusTooManyRequests, "您已达到每日总请求数限制（包括失败请求）")
+		return false
+	}
+
+	// 2. 检查成功请求数限制（使用临时key检查）
+	if successMaxCount > 0 {
+		checkKey := successKey + "_check"
+		if !inMemoryRateLimiter.Request(checkKey, successMaxCount, duration) {
+			abortWithOpenAiMessage(c, http.StatusTooManyRequests, "您已达到每日请求数限制")
+			return false
+		}
+	}
+
+	return true
+}
+
+// recordTokenDailySuccess 记录每日成功请求
+func recordTokenDailySuccess(c *gin.Context) {
+	if !setting.TokenDailyRateLimitEnabled {
+		return
+	}
+
+	tokenId := common.GetContextKeyInt(c, constant.ContextKeyTokenId)
+	if tokenId == 0 {
+		return
+	}
+
+	// 获取分组配置
+	group := common.GetContextKeyString(c, constant.ContextKeyTokenGroup)
+	successMaxCount := setting.TokenDailyRateLimitSuccessCount
+
+	_, groupSuccessCount, found := setting.GetTokenDailyRateLimit(group)
+	if found {
+		successMaxCount = groupSuccessCount
+	}
+
+	if successMaxCount == 0 {
+		return
+	}
+
+	rateLimitKey := strconv.Itoa(tokenId)
+
+	if common.RedisEnabled {
+		ctx := context.Background()
+		rdb := common.RDB
+		successKey := fmt.Sprintf("rateLimit:%s:%s", TokenDailyRateLimitSuccessCountMark, rateLimitKey)
+		recordRedisRequest(ctx, rdb, successKey, successMaxCount)
+	} else {
+		duration := int64(86400)
+		successKey := TokenDailyRateLimitSuccessCountMark + rateLimitKey
+		inMemoryRateLimiter.Request(successKey, successMaxCount, duration)
+	}
+}
+
 // ModelRequestRateLimit 模型请求限流中间件
 func ModelRequestRateLimit() func(c *gin.Context) {
 	return func(c *gin.Context) {
-		// 在每个请求时检查是否启用限流
+		// 1. 先检查 per-key 分钟级限流
+		if !checkTokenRateLimit(c) {
+			return
+		}
+
+		// 2. 检查 per-key 每日限流
+		if !checkTokenDailyRateLimit(c) {
+			return
+		}
+
+		// 3. 再检查原有的 per-user 限流（保持兼容性）
 		if !setting.ModelRequestRateLimitEnabled {
 			c.Next()
+			// 请求成功后记录 per-key 成功请求
+			if c.Writer.Status() < 400 {
+				recordTokenRateLimitSuccess(c)
+				recordTokenDailySuccess(c)
+			}
 			return
 		}
 
@@ -177,14 +498,11 @@ func ModelRequestRateLimit() func(c *gin.Context) {
 		totalMaxCount := setting.ModelRequestRateLimitCount
 		successMaxCount := setting.ModelRequestRateLimitSuccessCount
 
-		// 获取分组
-		group := common.GetContextKeyString(c, constant.ContextKeyTokenGroup)
-		if group == "" {
-			group = common.GetContextKeyString(c, constant.ContextKeyUserGroup)
-		}
+		// per-user 限流使用 user group（不是 token group）
+		userGroup := common.GetContextKeyString(c, constant.ContextKeyUserGroup)
 
 		//获取分组的限流配置
-		groupTotalCount, groupSuccessCount, found := setting.GetGroupRateLimit(group)
+		groupTotalCount, groupSuccessCount, found := setting.GetGroupRateLimit(userGroup)
 		if found {
 			totalMaxCount = groupTotalCount
 			successMaxCount = groupSuccessCount
@@ -196,5 +514,11 @@ func ModelRequestRateLimit() func(c *gin.Context) {
 		} else {
 			memoryRateLimitHandler(duration, totalMaxCount, successMaxCount)(c)
 		}
+
+		// 请求成功后记录 per-key 成功请求
+		if c.Writer.Status() < 400 {
+			recordTokenRateLimitSuccess(c)
+			recordTokenDailySuccess(c)
+		}
 	}
 }
diff --git a/model/option.go b/model/option.go
index 697e77df..5da23546 100644
--- a/model/option.go
+++ b/model/option.go
@@ -112,6 +112,15 @@ func InitOptionMap() {
 	common.OptionMap["ModelRequestRateLimitDurationMinutes"] = strconv.Itoa(setting.ModelRequestRateLimitDurationMinutes)
 	common.OptionMap["ModelRequestRateLimitSuccessCount"] = strconv.Itoa(setting.ModelRequestRateLimitSuccessCount)
 	common.OptionMap["ModelRequestRateLimitGroup"] = setting.ModelRequestRateLimitGroup2JSONString()
+	common.OptionMap["TokenRateLimitEnabled"] = strconv.FormatBool(setting.TokenRateLimitEnabled)
+	common.OptionMap["TokenRateLimitCount"] = strconv.Itoa(setting.TokenRateLimitCount)
+	common.OptionMap["TokenRateLimitSuccessCount"] = strconv.Itoa(setting.TokenRateLimitSuccessCount)
+	common.OptionMap["TokenRateLimitDurationMinutes"] = strconv.Itoa(setting.TokenRateLimitDurationMinutes)
+	common.OptionMap["TokenRateLimitGroup"] = setting.TokenRateLimitGroup2JSONString()
+	common.OptionMap["TokenDailyRateLimitEnabled"] = strconv.FormatBool(setting.TokenDailyRateLimitEnabled)
+	common.OptionMap["TokenDailyRateLimitCount"] = strconv.Itoa(setting.TokenDailyRateLimitCount)
+	common.OptionMap["TokenDailyRateLimitSuccessCount"] = strconv.Itoa(setting.TokenDailyRateLimitSuccessCount)
+	common.OptionMap["TokenDailyRateLimitGroup"] = setting.TokenDailyRateLimitGroup2JSONString()
 	common.OptionMap["ModelRatio"] = ratio_setting.ModelRatio2JSONString()
 	common.OptionMap["ModelPrice"] = ratio_setting.ModelPrice2JSONString()
 	common.OptionMap["CacheRatio"] = ratio_setting.CacheRatio2JSONString()
@@ -288,6 +297,10 @@ func updateOptionMap(key string, value string) (err error) {
 			setting.CheckSensitiveOnPromptEnabled = boolValue
 		case "ModelRequestRateLimitEnabled":
 			setting.ModelRequestRateLimitEnabled = boolValue
+		case "TokenRateLimitEnabled":
+			setting.TokenRateLimitEnabled = boolValue
+		case "TokenDailyRateLimitEnabled":
+			setting.TokenDailyRateLimitEnabled = boolValue
 		case "StopOnSensitiveEnabled":
 			setting.StopOnSensitiveEnabled = boolValue
 		case "SMTPSSLEnabled":
@@ -408,6 +421,20 @@ func updateOptionMap(key string, value string) (err error) {
 		setting.ModelRequestRateLimitSuccessCount, _ = strconv.Atoi(value)
 	case "ModelRequestRateLimitGroup":
 		err = setting.UpdateModelRequestRateLimitGroupByJSONString(value)
+	case "TokenRateLimitCount":
+		setting.TokenRateLimitCount, _ = strconv.Atoi(value)
+	case "TokenRateLimitSuccessCount":
+		setting.TokenRateLimitSuccessCount, _ = strconv.Atoi(value)
+	case "TokenRateLimitDurationMinutes":
+		setting.TokenRateLimitDurationMinutes, _ = strconv.Atoi(value)
+	case "TokenRateLimitGroup":
+		err = setting.UpdateTokenRateLimitGroupByJSONString(value)
+	case "TokenDailyRateLimitCount":
+		setting.TokenDailyRateLimitCount, _ = strconv.Atoi(value)
+	case "TokenDailyRateLimitSuccessCount":
+		setting.TokenDailyRateLimitSuccessCount, _ = strconv.Atoi(value)
+	case "TokenDailyRateLimitGroup":
+		err = setting.UpdateTokenDailyRateLimitGroupByJSONString(value)
 	case "RetryTimes":
 		common.RetryTimes, _ = strconv.Atoi(value)
 	case "DataExportInterval":
diff --git a/setting/rate_limit.go b/setting/rate_limit.go
index 413f3958..b2733e00 100644
--- a/setting/rate_limit.go
+++ b/setting/rate_limit.go
@@ -9,6 +9,7 @@ import (
 	"github.com/QuantumNous/new-api/common"
 )
 
+// Per-user rate limit settings (原有的按用户限流)
 var ModelRequestRateLimitEnabled = false
 var ModelRequestRateLimitDurationMinutes = 1
 var ModelRequestRateLimitCount = 0
@@ -16,6 +17,21 @@ var ModelRequestRateLimitSuccessCount = 1000
 var ModelRequestRateLimitGroup = map[string][2]int{}
 var ModelRequestRateLimitMutex sync.RWMutex
 
+// Per-key minute rate limit settings (按密钥的分钟级限流)
+var TokenRateLimitEnabled = false
+var TokenRateLimitDurationMinutes = 1
+var TokenRateLimitCount = 0
+var TokenRateLimitSuccessCount = 0
+var TokenRateLimitGroup = map[string][2]int{}
+var TokenRateLimitMutex sync.RWMutex
+
+// Per-key daily rate limit settings (按密钥的每日限流)
+var TokenDailyRateLimitEnabled = false
+var TokenDailyRateLimitCount = 0        // 每日总请求数限制（0表示不限制）
+var TokenDailyRateLimitSuccessCount = 0 // 每日成功请求数限制（0表示不限制）
+var TokenDailyRateLimitGroup = map[string][2]int{}
+var TokenDailyRateLimitMutex sync.RWMutex
+
 func ModelRequestRateLimitGroup2JSONString() string {
 	ModelRequestRateLimitMutex.RLock()
 	defer ModelRequestRateLimitMutex.RUnlock()
@@ -67,3 +83,109 @@ func CheckModelRequestRateLimitGroup(jsonStr string) error {
 
 	return nil
 }
+
+// Token minute rate limit functions
+func TokenRateLimitGroup2JSONString() string {
+	TokenRateLimitMutex.RLock()
+	defer TokenRateLimitMutex.RUnlock()
+
+	jsonBytes, err := json.Marshal(TokenRateLimitGroup)
+	if err != nil {
+		common.SysLog("error marshalling token rate limit group: " + err.Error())
+	}
+	return string(jsonBytes)
+}
+
+func UpdateTokenRateLimitGroupByJSONString(jsonStr string) error {
+	TokenRateLimitMutex.Lock()
+	defer TokenRateLimitMutex.Unlock()
+
+	TokenRateLimitGroup = make(map[string][2]int)
+	return json.Unmarshal([]byte(jsonStr), &TokenRateLimitGroup)
+}
+
+func GetTokenRateLimit(group string) (totalCount, successCount int, found bool) {
+	TokenRateLimitMutex.RLock()
+	defer TokenRateLimitMutex.RUnlock()
+
+	if TokenRateLimitGroup == nil {
+		return 0, 0, false
+	}
+
+	limits, found := TokenRateLimitGroup[group]
+	if !found {
+		return 0, 0, false
+	}
+	return limits[0], limits[1], true
+}
+
+func CheckTokenRateLimitGroup(jsonStr string) error {
+	checkTokenRateLimitGroup := make(map[string][2]int)
+	err := json.Unmarshal([]byte(jsonStr), &checkTokenRateLimitGroup)
+	if err != nil {
+		return err
+	}
+	for group, limits := range checkTokenRateLimitGroup {
+		if limits[0] < 0 || limits[1] < 0 {
+			return fmt.Errorf("group %s has negative rate limit values: [%d, %d]", group, limits[0], limits[1])
+		}
+		if limits[0] > math.MaxInt32 || limits[1] > math.MaxInt32 {
+			return fmt.Errorf("group %s [%d, %d] has max rate limits value 2147483647", group, limits[0], limits[1])
+		}
+	}
+
+	return nil
+}
+
+// Token daily rate limit functions
+func TokenDailyRateLimitGroup2JSONString() string {
+	TokenDailyRateLimitMutex.RLock()
+	defer TokenDailyRateLimitMutex.RUnlock()
+
+	jsonBytes, err := json.Marshal(TokenDailyRateLimitGroup)
+	if err != nil {
+		common.SysLog("error marshalling token daily rate limit group: " + err.Error())
+	}
+	return string(jsonBytes)
+}
+
+func UpdateTokenDailyRateLimitGroupByJSONString(jsonStr string) error {
+	TokenDailyRateLimitMutex.Lock()
+	defer TokenDailyRateLimitMutex.Unlock()
+
+	TokenDailyRateLimitGroup = make(map[string][2]int)
+	return json.Unmarshal([]byte(jsonStr), &TokenDailyRateLimitGroup)
+}
+
+func GetTokenDailyRateLimit(group string) (totalCount, successCount int, found bool) {
+	TokenDailyRateLimitMutex.RLock()
+	defer TokenDailyRateLimitMutex.RUnlock()
+
+	if TokenDailyRateLimitGroup == nil {
+		return 0, 0, false
+	}
+
+	limits, found := TokenDailyRateLimitGroup[group]
+	if !found {
+		return 0, 0, false
+	}
+	return limits[0], limits[1], true
+}
+
+func CheckTokenDailyRateLimitGroup(jsonStr string) error {
+	checkTokenDailyRateLimitGroup := make(map[string][2]int)
+	err := json.Unmarshal([]byte(jsonStr), &checkTokenDailyRateLimitGroup)
+	if err != nil {
+		return err
+	}
+	for group, limits := range checkTokenDailyRateLimitGroup {
+		if limits[0] < 0 || limits[1] < 0 {
+			return fmt.Errorf("group %s has negative rate limit values: [%d, %d]", group, limits[0], limits[1])
+		}
+		if limits[0] > math.MaxInt32 || limits[1] > math.MaxInt32 {
+			return fmt.Errorf("group %s [%d, %d] has max rate limits value 2147483647", group, limits[0], limits[1])
+		}
+	}
+
+	return nil
+}
diff --git a/web/src/components/table/tokens/modals/EditTokenModal.jsx b/web/src/components/table/tokens/modals/EditTokenModal.jsx
index fce48201..4da55785 100644
--- a/web/src/components/table/tokens/modals/EditTokenModal.jsx
+++ b/web/src/components/table/tokens/modals/EditTokenModal.jsx
@@ -66,9 +66,9 @@ const EditTokenModal = (props) => {
 
   const getInitValues = () => ({
     name: '',
-    remain_quota: 0,
+    remain_quota: 500000,
     expired_time: -1,
-    unlimited_quota: true,
+    unlimited_quota: false,
     model_limits_enabled: false,
     model_limits: [],
     allow_ips: '',
@@ -138,12 +138,14 @@ const EditTokenModal = (props) => {
       if (statusState?.status?.default_use_auto_group) {
         if (localGroupOptions.some((group) => group.value === 'auto')) {
           localGroupOptions.sort((a, b) => (a.value === 'auto' ? -1 : 1));
+        } else {
+          localGroupOptions.unshift({ label: t('自动选择'), value: 'auto' });
         }
       }
       setGroups(localGroupOptions);
-      // if (statusState?.status?.default_use_auto_group && formApiRef.current) {
-      //   formApiRef.current.setValue('group', 'auto');
-      // }
+      if (statusState?.status?.default_use_auto_group && formApiRef.current) {
+        formApiRef.current.setValue('group', 'auto');
+      }
     } else {
       showError(t(message));
     }
diff --git a/web/src/pages/Setting/RateLimit/SettingsRequestRateLimit.jsx b/web/src/pages/Setting/RateLimit/SettingsRequestRateLimit.jsx
index 12b97638..c4316d0b 100644
--- a/web/src/pages/Setting/RateLimit/SettingsRequestRateLimit.jsx
+++ b/web/src/pages/Setting/RateLimit/SettingsRequestRateLimit.jsx
@@ -39,6 +39,15 @@ export default function RequestRateLimit(props) {
     ModelRequestRateLimitSuccessCount: 1000,
     ModelRequestRateLimitDurationMinutes: 1,
     ModelRequestRateLimitGroup: '',
+    TokenRateLimitEnabled: false,
+    TokenRateLimitCount: 0,
+    TokenRateLimitSuccessCount: 0,
+    TokenRateLimitDurationMinutes: 1,
+    TokenRateLimitGroup: '',
+    TokenDailyRateLimitEnabled: false,
+    TokenDailyRateLimitCount: 0,
+    TokenDailyRateLimitSuccessCount: 0,
+    TokenDailyRateLimitGroup: '',
   });
   const refForm = useRef();
   const [inputsRow, setInputsRow] = useState(inputs);
@@ -68,11 +77,11 @@ export default function RequestRateLimit(props) {
             return showError(t('部分保存失败，请重试'));
         }
 
-        for (let i = 0; i < res.length; i++) {
-          if (!res[i].data.success) {
-            return showError(res[i].data.message);
-          }
+      for (let i = 0; i < res.length; i++) {
+        if (!res[i].data.success) {
+          return showError(res[i].data.message);
         }
+      }
 
         showSuccess(t('保存成功'));
         props.refresh();
@@ -185,6 +194,121 @@ export default function RequestRateLimit(props) {
                     '{\n  "default": [200, 100],\n  "vip": [0, 1000]\n}',
                   )}
                   field={'ModelRequestRateLimitGroup'}
+                autosize={{ minRows: 5, maxRows: 15 }}
+                trigger='blur'
+                        stopValidateWithError
+                rules={[
+                  {
+                  validator: (rule, value) => verifyJSON(value),
+                  message: t('不是合法的 JSON 字符串'),
+                  },
+                ]}
+                  extraText={
+                    <div>
+                      <p>{t('说明：')}</p>
+                      <ul>
+                        <li>{t('使用 JSON 对象格式，格式为：{"组名": [最多请求次数, 最多请求完成次数]}')}</li>
+                      <li>{t('示例：{"default": [200, 100], "vip": [0, 1000]}。')}</li>
+                      <li>{t('[最多请求次数]必须大于等于0，[最多请求完成次数]必须大于等于1。')}</li>
+                        <li>{t('[最多请求次数]和[最多请求完成次数]的最大值为2147483647。')}</li>
+                        <li>{t('分组速率配置优先级高于全局速率限制。')}</li>
+                        <li>{t('限制周期统一使用上方配置的"限制周期"值。')}</li>
+                      </ul>
+                    </div>
+                  }
+                  onChange={(value) => {
+                    setInputs({ ...inputs, ModelRequestRateLimitGroup: value });
+                  }}
+                />
+              </Col>
+            </Row>
+            <Row>
+              <Button size='default' onClick={onSubmit}>
+                {t('保存模型速率限制')}
+              </Button>
+            </Row>
+          </Form.Section>
+
+          <Form.Section text={t('密钥分钟级请求限制')}>
+            <Row gutter={16}>
+              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
+                <Form.Switch
+                  field={'TokenRateLimitEnabled'}
+                  label={t('启用密钥分钟级请求限制（按API密钥限流）')}
+                  size='default'
+                  checkedText='｜'
+                  uncheckedText='〇'
+                  onChange={(value) => {
+                    setInputs({
+                      ...inputs,
+                      TokenRateLimitEnabled: value,
+                    });
+                  }}
+                />
+              </Col>
+            </Row>
+            <Row>
+              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
+                <Form.InputNumber
+                  label={t('限制周期')}
+                  step={1}
+                  min={0}
+                  suffix={t('分钟')}
+                  extraText={t('频率限制的周期（分钟）')}
+                  field={'TokenRateLimitDurationMinutes'}
+                  onChange={(value) =>
+                    setInputs({
+                      ...inputs,
+                      TokenRateLimitDurationMinutes: String(value),
+                    })
+                  }
+                />
+              </Col>
+            </Row>
+            <Row>
+              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
+                <Form.InputNumber
+                  label={t('密钥每周期最多请求次数')}
+                  step={1}
+                  min={0}
+                  max={100000000}
+                  suffix={t('次')}
+                  extraText={t('包括失败请求的次数，0代表不限制')}
+                  field={'TokenRateLimitCount'}
+                  onChange={(value) =>
+                    setInputs({
+                      ...inputs,
+                      TokenRateLimitCount: String(value),
+                    })
+                  }
+                />
+              </Col>
+              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
+                <Form.InputNumber
+                  label={t('密钥每周期最多请求完成次数')}
+                  step={1}
+                  min={0}
+                  max={100000000}
+                  suffix={t('次')}
+                  extraText={t('只包括请求成功的次数，0代表不限制')}
+                  field={'TokenRateLimitSuccessCount'}
+                  onChange={(value) =>
+                    setInputs({
+                      ...inputs,
+                      TokenRateLimitSuccessCount: String(value),
+                    })
+                  }
+                />
+              </Col>
+            </Row>
+            <Row>
+              <Col xs={24} sm={16}>
+                <Form.TextArea
+                  label={t('分组限制')}
+                  placeholder={t(
+                    '{\n  "default": [200, 100],\n  "vip": [0, 1000]\n}',
+                  )}
+                  field={'TokenRateLimitGroup'}
                   autosize={{ minRows: 5, maxRows: 15 }}
                   trigger='blur'
                   stopValidateWithError
@@ -198,40 +322,123 @@ export default function RequestRateLimit(props) {
                     <div>
                       <p>{t('说明：')}</p>
                       <ul>
-                        <li>
-                          {t(
-                            '使用 JSON 对象格式，格式为：{"组名": [最多请求次数, 最多请求完成次数]}',
-                          )}
-                        </li>
-                        <li>
-                          {t(
-                            '示例：{"default": [200, 100], "vip": [0, 1000]}。',
-                          )}
-                        </li>
-                        <li>
-                          {t(
-                            '[最多请求次数]必须大于等于0，[最多请求完成次数]必须大于等于1。',
-                          )}
-                        </li>
-                        <li>
-                          {t(
-                            '[最多请求次数]和[最多请求完成次数]的最大值为2147483647。',
-                          )}
-                        </li>
-                        <li>{t('分组速率配置优先级高于全局速率限制。')}</li>
-                        <li>{t('限制周期统一使用上方配置的“限制周期”值。')}</li>
+                        <li>{t('使用 JSON 对象格式，格式为：{"组名": [每周期最多请求次数, 每周期最多请求完成次数]}')}</li>
+                        <li>{t('示例：{"default": [200, 100], "vip": [0, 1000]}。')}</li>
+                        <li>{t('[每周期最多请求次数]和[每周期最多请求完成次数]必须大于等于0。')}</li>
+                        <li>{t('最大值为2147483647。')}</li>
+                        <li>{t('分组配置优先级高于全局配置。')}</li>
+                        <li>{t('此限制按API密钥（Token）维度，每个密钥独立计算。')}</li>
+                        <li>{t('限制周期使用上方配置的"限制周期"值。')}</li>
                       </ul>
                     </div>
                   }
                   onChange={(value) => {
-                    setInputs({ ...inputs, ModelRequestRateLimitGroup: value });
+                    setInputs({ ...inputs, TokenRateLimitGroup: value });
                   }}
                 />
               </Col>
             </Row>
             <Row>
               <Button size='default' onClick={onSubmit}>
-                {t('保存模型速率限制')}
+                {t('保存密钥分钟级限制')}
+              </Button>
+            </Row>
+          </Form.Section>
+
+          <Form.Section text={t('密钥每日请求限制')}>
+            <Row gutter={16}>
+              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
+                <Form.Switch
+                  field={'TokenDailyRateLimitEnabled'}
+                  label={t('启用密钥每日请求限制（按API密钥限流）')}
+                  size='default'
+                  checkedText='｜'
+                  uncheckedText='〇'
+                  onChange={(value) => {
+                    setInputs({
+                      ...inputs,
+                      TokenDailyRateLimitEnabled: value,
+                    });
+                  }}
+                />
+              </Col>
+            </Row>
+            <Row>
+              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
+                <Form.InputNumber
+                  label={t('密钥每日最多请求次数')}
+                  step={1}
+                  min={0}
+                  max={100000000}
+                  suffix={t('次')}
+                  extraText={t('包括失败请求的次数，0代表不限制')}
+                  field={'TokenDailyRateLimitCount'}
+                  onChange={(value) =>
+                    setInputs({
+                      ...inputs,
+                      TokenDailyRateLimitCount: String(value),
+                    })
+                  }
+                />
+              </Col>
+              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
+                <Form.InputNumber
+                  label={t('密钥每日最多请求完成次数')}
+                  step={1}
+                  min={0}
+                  max={100000000}
+                  suffix={t('次')}
+                  extraText={t('只包括请求成功的次数，0代表不限制')}
+                  field={'TokenDailyRateLimitSuccessCount'}
+                  onChange={(value) =>
+                    setInputs({
+                      ...inputs,
+                      TokenDailyRateLimitSuccessCount: String(value),
+                    })
+                  }
+                />
+              </Col>
+            </Row>
+            <Row>
+              <Col xs={24} sm={16}>
+                <Form.TextArea
+                  label={t('分组每日限制')}
+                  placeholder={t(
+                    '{\n  "default": [1000, 500],\n  "vip": [0, 10000]\n}',
+                  )}
+                  field={'TokenDailyRateLimitGroup'}
+                  autosize={{ minRows: 5, maxRows: 15 }}
+                  trigger='blur'
+                  stopValidateWithError
+                  rules={[
+                    {
+                      validator: (rule, value) => verifyJSON(value),
+                      message: t('不是合法的 JSON 字符串'),
+                    },
+                  ]}
+                  extraText={
+                    <div>
+                      <p>{t('说明：')}</p>
+                      <ul>
+                        <li>{t('使用 JSON 对象格式，格式为：{"组名": [每日最多请求次数, 每日最多请求完成次数]}')}</li>
+                        <li>{t('示例：{"default": [1000, 500], "vip": [0, 10000]}。')}</li>
+                        <li>{t('[每日最多请求次数]和[每日最多请求完成次数]必须大于等于0。')}</li>
+                        <li>{t('最大值为2147483647。')}</li>
+                        <li>{t('分组配置优先级高于全局配置。')}</li>
+                        <li>{t('此限制按API密钥（Token）维度，每个密钥独立计算。')}</li>
+                        <li>{t('限制周期为自然日（UTC+0），每日0点重置。')}</li>
+                      </ul>
+                    </div>
+                  }
+                  onChange={(value) => {
+                    setInputs({ ...inputs, TokenDailyRateLimitGroup: value });
+                  }}
+                />
+              </Col>
+            </Row>
+            <Row>
+              <Button size='default' onClick={onSubmit}>
+                {t('保存密钥每日限制')}
               </Button>
             </Row>
           </Form.Section>
```

---

### 提交 2: feat: 增强流式响应处理，添加拒绝原因记录和有效性检查

- **哈希**: `59e2cc896ae3aeec237585b6f12d52dff8c0ec36`
- **日期**: 2026-01-28 23:41:06 +0800
- **作者**: Mingluan Mu <mumingluan@qq.com>

**完整 Diff:**

```diff
diff --git a/relay/channel/claude/relay-claude.go b/relay/channel/claude/relay-claude.go
index 069c784c..8976d063 100644
--- a/relay/channel/claude/relay-claude.go
+++ b/relay/channel/claude/relay-claude.go
@@ -698,6 +698,7 @@ func HandleStreamResponseData(c *gin.Context, info *relaycommon.RelayInfo, claud
 	if claudeError := claudeResponse.GetClaudeError(); claudeError != nil && claudeError.Type != "" {
 		return types.WithClaudeError(*claudeError, http.StatusInternalServerError)
 	}
+	// 记录拒绝原因（上游功能）
 	if claudeResponse.StopReason != "" {
 		maybeMarkClaudeRefusal(c, claudeResponse.StopReason)
 	}
@@ -768,16 +769,23 @@ func ClaudeStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.
 		ResponseText: strings.Builder{},
 		Usage:        &dto.Usage{},
 	}
-	var err *types.NewAPIError
+	var streamErr *types.NewAPIError
+	var hasValidResponse bool
 	helper.StreamScannerHandler(c, resp, info, func(data string) bool {
-		err = HandleStreamResponseData(c, info, claudeInfo, data)
-		if err != nil {
+		streamErr = HandleStreamResponseData(c, info, claudeInfo, data)
+		if streamErr != nil {
 			return false
 		}
+		hasValidResponse = true
 		return true
 	})
-	if err != nil {
-		return nil, err
+	if streamErr != nil {
+		return nil, streamErr
+	}
+
+	// 检查流式响应是否收到有效内容，如果没有则返回错误以触发重试
+	if !hasValidResponse {
+		return nil, types.NewOpenAIError(fmt.Errorf("empty response from Claude API"), types.ErrorCodeEmptyResponse, http.StatusInternalServerError)
 	}
 
 	HandleStreamFinalResponse(c, info, claudeInfo)
@@ -793,6 +801,7 @@ func HandleClaudeResponseData(c *gin.Context, info *relaycommon.RelayInfo, claud
 	if claudeError := claudeResponse.GetClaudeError(); claudeError != nil && claudeError.Type != "" {
 		return types.WithClaudeError(*claudeError, http.StatusInternalServerError)
 	}
+	// 记录拒绝原因（上游功能）
 	maybeMarkClaudeRefusal(c, claudeResponse.StopReason)
 	if claudeInfo.Usage == nil {
 		claudeInfo.Usage = &dto.Usage{}
diff --git a/relay/channel/gemini/relay-gemini-native.go b/relay/channel/gemini/relay-gemini-native.go
index 39485b16..29af6303 100644
--- a/relay/channel/gemini/relay-gemini-native.go
+++ b/relay/channel/gemini/relay-gemini-native.go
@@ -1,6 +1,7 @@
 package gemini
 
 import (
+	"errors"
 	"fmt"
 	"io"
 	"net/http"
@@ -37,8 +38,17 @@ func GeminiTextGenerationHandler(c *gin.Context, info *relaycommon.RelayInfo, re
 		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
 	}
 
-	if len(geminiResponse.Candidates) == 0 && geminiResponse.PromptFeedback != nil && geminiResponse.PromptFeedback.BlockReason != nil {
-		common.SetContextKey(c, constant.ContextKeyAdminRejectReason, fmt.Sprintf("gemini_block_reason=%s", *geminiResponse.PromptFeedback.BlockReason))
+	// 检查是否有候选返回，如果没有则返回错误以触发重试
+	if len(geminiResponse.Candidates) == 0 {
+		if geminiResponse.PromptFeedback != nil && geminiResponse.PromptFeedback.BlockReason != nil {
+			// 记录拒绝原因（上游功能）
+			common.SetContextKey(c, constant.ContextKeyAdminRejectReason, fmt.Sprintf("gemini_block_reason=%s", *geminiResponse.PromptFeedback.BlockReason))
+			return nil, types.NewOpenAIError(errors.New("request blocked by Gemini API: "+*geminiResponse.PromptFeedback.BlockReason), types.ErrorCodePromptBlocked, http.StatusBadRequest, types.ErrOptionWithSkipRetry())
+		} else {
+			// 记录空响应原因（上游功能）
+			common.SetContextKey(c, constant.ContextKeyAdminRejectReason, "gemini_empty_candidates")
+			return nil, types.NewOpenAIError(errors.New("empty response from Gemini API"), types.ErrorCodeEmptyResponse, http.StatusInternalServerError)
+		}
 	}
 
 	// 计算使用量（基于 UsageMetadata）
diff --git a/relay/channel/openai/relay-openai.go b/relay/channel/openai/relay-openai.go
index a4de1611..c07f2241 100644
--- a/relay/channel/openai/relay-openai.go
+++ b/relay/channel/openai/relay-openai.go
@@ -188,6 +188,11 @@ func OaiStreamHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Re
 
 	applyUsagePostProcessing(info, usage, common.StringToByteSlice(lastStreamData))
 
+	// 检查流式响应是否收到有效内容，如果没有则返回错误以触发重试
+	if len(streamItems) == 0 {
+		return nil, types.NewOpenAIError(fmt.Errorf("empty response from upstream API"), types.ErrorCodeEmptyResponse, http.StatusInternalServerError)
+	}
+
 	HandleFinalResponse(c, info, lastStreamData, responseId, createAt, model, systemFingerprint, usage, containStreamUsage)
 
 	return usage, nil
@@ -229,6 +234,12 @@ func OpenaiHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Respo
 		return nil, types.WithOpenAIError(*oaiError, resp.StatusCode)
 	}
 
+	// 检查是否有有效的 choices 返回，如果没有则返回错误以触发重试
+	if len(simpleResponse.Choices) == 0 {
+		return nil, types.NewOpenAIError(fmt.Errorf("empty response from upstream API"), types.ErrorCodeEmptyResponse, http.StatusInternalServerError)
+	}
+
+	// 记录 content_filter 拒绝原因（上游功能）
 	for _, choice := range simpleResponse.Choices {
 		if choice.FinishReason == constant.FinishReasonContentFilter {
 			common.SetContextKey(c, constant.ContextKeyAdminRejectReason, "openai_finish_reason=content_filter")
```

---

### 提交 3: feat: 添加令牌速率限制设置和格式化功能

- **哈希**: `5fb49e4c19ffe0800ece4fb510cd7d73f485d8cb`
- **日期**: 2026-01-29 00:27:32 +0800
- **作者**: Mingluan Mu <mumingluan@qq.com>

**完整 Diff:**

```diff
diff --git a/web/src/components/settings/RateLimitSetting.jsx b/web/src/components/settings/RateLimitSetting.jsx
index be83e027..c7b8a38d 100644
--- a/web/src/components/settings/RateLimitSetting.jsx
+++ b/web/src/components/settings/RateLimitSetting.jsx
@@ -32,6 +32,15 @@ const RateLimitSetting = () => {
     ModelRequestRateLimitSuccessCount: 1000,
     ModelRequestRateLimitDurationMinutes: 1,
     ModelRequestRateLimitGroup: '',
+    TokenRateLimitEnabled: false,
+    TokenRateLimitCount: 0,
+    TokenRateLimitSuccessCount: 0,
+    TokenRateLimitDurationMinutes: 1,
+    TokenRateLimitGroup: '',
+    TokenDailyRateLimitEnabled: false,
+    TokenDailyRateLimitCount: 0,
+    TokenDailyRateLimitSuccessCount: 0,
+    TokenDailyRateLimitGroup: '',
   });
 
   let [loading, setLoading] = useState(false);
@@ -42,8 +51,15 @@ const RateLimitSetting = () => {
     if (success) {
       let newInputs = {};
       data.forEach((item) => {
-        if (item.key === 'ModelRequestRateLimitGroup') {
-          item.value = JSON.stringify(JSON.parse(item.value), null, 2);
+        // 格式化所有 Group JSON 字段
+        if (item.key === 'ModelRequestRateLimitGroup' ||
+            item.key === 'TokenRateLimitGroup' ||
+            item.key === 'TokenDailyRateLimitGroup') {
+          try {
+            item.value = JSON.stringify(JSON.parse(item.value), null, 2);
+          } catch (e) {
+            // 如果解析失败，保持原值
+          }
         }
 
         if (item.key.endsWith('Enabled')) {
```

---

### 提交 4: feat: 更新限流逻辑，添加检查功能并优化过期时间设置

- **哈希**: `55e2cb8ec7afd31e8bbda6d828ec27f96785a260`
- **日期**: 2026-02-01 11:12:19 +0800
- **作者**: Mingluan Mu <mumingluan@qq.com>

**完整 Diff:**

```diff
diff --git a/common/limiter/lua/rate_limit.lua b/common/limiter/lua/rate_limit.lua
index c07fd3a8..e27f3e39 100644
--- a/common/limiter/lua/rate_limit.lua
+++ b/common/limiter/lua/rate_limit.lua
@@ -37,8 +37,8 @@ if tokens >= requested then
     allowed = true
 end
 
----- 更新桶状态并设置过期时间
+--- 更新桶状态并设置过期时间
 redis.call('HMSET', key, 'tokens', tokens, 'last_time', last_time)
---redis.call('EXPIRE', key, math.ceil(capacity / rate) + 60) -- 适当延长过期时间
+redis.call('EXPIRE', key, math.ceil(capacity / rate) + 60) -- 适当延长过期时间
 
 return allowed and 1 or 0
\ No newline at end of file
diff --git a/common/rate-limit.go b/common/rate-limit.go
index 301c101c..b6376570 100644
--- a/common/rate-limit.go
+++ b/common/rate-limit.go
@@ -68,3 +68,22 @@ func (l *InMemoryRateLimiter) Request(key string, maxRequestNum int, duration in
 	}
 	return true
 }
+
+// Check 只检查是否达到限制，不记录请求（用于成功请求数限制的预检查）
+func (l *InMemoryRateLimiter) Check(key string, maxRequestNum int, duration int64) bool {
+	l.mutex.Lock()
+	defer l.mutex.Unlock()
+	queue, ok := l.store[key]
+	if !ok {
+		return true // 没有记录，允许
+	}
+	now := time.Now().Unix()
+	if len(*queue) < maxRequestNum {
+		return true // 未达到限制
+	}
+	// 检查最老的记录是否已过期
+	if now-(*queue)[0] >= duration {
+		return true // 时间窗口已过，允许
+	}
+	return false // 在时间窗口内已达到限制
+}
diff --git a/middleware/model-rate-limit.go b/middleware/model-rate-limit.go
index 49d9d075..b7f49612 100644
--- a/middleware/model-rate-limit.go
+++ b/middleware/model-rate-limit.go
@@ -62,15 +62,17 @@ func checkRedisRateLimit(ctx context.Context, rdb *redis.Client, key string, max
 	// 如果在时间窗口内已达到限制，拒绝请求
 	subTime := nowTime.Sub(oldTime).Seconds()
 	if int64(subTime) < duration {
-		rdb.Expire(ctx, key, time.Duration(setting.ModelRequestRateLimitDurationMinutes)*time.Minute)
+		rdb.Expire(ctx, key, time.Duration(duration)*time.Second)
 		return false, nil
 	}
 
+	// 超过时间窗口，清空旧数据，允许新的时间窗口开始
+	rdb.Del(ctx, key)
 	return true, nil
 }
 
 // 记录Redis请求
-func recordRedisRequest(ctx context.Context, rdb *redis.Client, key string, maxCount int) {
+func recordRedisRequest(ctx context.Context, rdb *redis.Client, key string, maxCount int, duration int64) {
 	// 如果maxCount为0，不记录请求
 	if maxCount == 0 {
 		return
@@ -79,7 +81,7 @@ func recordRedisRequest(ctx context.Context, rdb *redis.Client, key string, maxC
 	now := time.Now().Format(timeFormat)
 	rdb.LPush(ctx, key, now)
 	rdb.LTrim(ctx, key, 0, int64(maxCount-1))
-	rdb.Expire(ctx, key, time.Duration(setting.ModelRequestRateLimitDurationMinutes)*time.Minute)
+	rdb.Expire(ctx, key, time.Duration(duration)*time.Second)
 }
 
 // Redis限流处理器
@@ -123,6 +125,7 @@ func redisRateLimitHandler(duration int64, totalMaxCount, successMaxCount int) g
 
 			if !allowed {
 				abortWithOpenAiMessage(c, http.StatusTooManyRequests, fmt.Sprintf("您已达到总请求数限制：%d分钟内最多请求%d次，包括失败次数，请检查您的请求是否正确", setting.ModelRequestRateLimitDurationMinutes, totalMaxCount))
+				return
 			}
 		}
 
@@ -131,7 +134,7 @@ func redisRateLimitHandler(duration int64, totalMaxCount, successMaxCount int) g
 
 		// 5. 如果请求成功，记录成功请求
 		if c.Writer.Status() < 400 {
-			recordRedisRequest(ctx, rdb, successKey, successMaxCount)
+			recordRedisRequest(ctx, rdb, successKey, successMaxCount, duration)
 		}
 	}
 }
@@ -152,10 +155,8 @@ func memoryRateLimitHandler(duration int64, totalMaxCount, successMaxCount int)
 			return
 		}
 
-		// 2. 检查成功请求数限制
-		// 使用一个临时key来检查限制，这样可以避免实际记录
-		checkKey := successKey + "_check"
-		if !inMemoryRateLimiter.Request(checkKey, successMaxCount, duration) {
+		// 2. 检查成功请求数限制（只检查不记录，使用 Check 方法）
+		if successMaxCount > 0 && !inMemoryRateLimiter.Check(successKey, successMaxCount, duration) {
 			c.Status(http.StatusTooManyRequests)
 			c.Abort()
 			return
@@ -270,10 +271,9 @@ func checkTokenRateLimitMemory(c *gin.Context, rateLimitKey string, totalMaxCoun
 		return false
 	}
 
-	// 2. 检查成功请求数限制（使用临时key检查）
+	// 2. 检查成功请求数限制（只检查不记录）
 	if successMaxCount > 0 {
-		checkKey := successKey + "_check"
-		if !inMemoryRateLimiter.Request(checkKey, successMaxCount, duration) {
+		if !inMemoryRateLimiter.Check(successKey, successMaxCount, duration) {
 			abortWithOpenAiMessage(c, http.StatusTooManyRequests, fmt.Sprintf("您已达到密钥请求数限制：%d分钟内最多请求%d次", setting.TokenRateLimitDurationMinutes, successMaxCount))
 			return false
 		}
@@ -312,7 +312,8 @@ func recordTokenRateLimitSuccess(c *gin.Context) {
 		ctx := context.Background()
 		rdb := common.RDB
 		successKey := fmt.Sprintf("rateLimit:%s:%s", TokenRateLimitSuccessCountMark, rateLimitKey)
-		recordRedisRequest(ctx, rdb, successKey, successMaxCount)
+		duration := int64(setting.TokenRateLimitDurationMinutes * 60)
+		recordRedisRequest(ctx, rdb, successKey, successMaxCount, duration)
 	} else {
 		duration := int64(setting.TokenRateLimitDurationMinutes * 60)
 		successKey := TokenRateLimitSuccessCountMark + rateLimitKey
@@ -419,10 +420,9 @@ func checkTokenDailyRateLimitMemory(c *gin.Context, rateLimitKey string, totalMa
 		return false
 	}
 
-	// 2. 检查成功请求数限制（使用临时key检查）
+	// 2. 检查成功请求数限制（只检查不记录）
 	if successMaxCount > 0 {
-		checkKey := successKey + "_check"
-		if !inMemoryRateLimiter.Request(checkKey, successMaxCount, duration) {
+		if !inMemoryRateLimiter.Check(successKey, successMaxCount, duration) {
 			abortWithOpenAiMessage(c, http.StatusTooManyRequests, "您已达到每日请求数限制")
 			return false
 		}
@@ -461,7 +461,8 @@ func recordTokenDailySuccess(c *gin.Context) {
 		ctx := context.Background()
 		rdb := common.RDB
 		successKey := fmt.Sprintf("rateLimit:%s:%s", TokenDailyRateLimitSuccessCountMark, rateLimitKey)
-		recordRedisRequest(ctx, rdb, successKey, successMaxCount)
+		duration := int64(86400)
+		recordRedisRequest(ctx, rdb, successKey, successMaxCount, duration)
 	} else {
 		duration := int64(86400)
 		successKey := TokenDailyRateLimitSuccessCountMark + rateLimitKey
diff --git a/setting/rate_limit.go b/setting/rate_limit.go
index b2733e00..7e875cb8 100644
--- a/setting/rate_limit.go
+++ b/setting/rate_limit.go
@@ -44,8 +44,8 @@ func ModelRequestRateLimitGroup2JSONString() string {
 }
 
 func UpdateModelRequestRateLimitGroupByJSONString(jsonStr string) error {
-	ModelRequestRateLimitMutex.RLock()
-	defer ModelRequestRateLimitMutex.RUnlock()
+	ModelRequestRateLimitMutex.Lock()
+	defer ModelRequestRateLimitMutex.Unlock()
 
 	ModelRequestRateLimitGroup = make(map[string][2]int)
 	return json.Unmarshal([]byte(jsonStr), &ModelRequestRateLimitGroup)
```

---

### 提交 5: feat: 更新适配器逻辑，支持Gemini格式的请求和响应处理，并优化错误检查

- **哈希**: `a566bdbe802c4d8668844e95aadf98972bb9171f`
- **日期**: 2026-02-18 22:18:09 +0800
- **作者**: Mingluan Mu <mumingluan@qq.com>

**完整 Diff:**

```diff
diff --git a/relay/channel/claude/adaptor.go b/relay/channel/claude/adaptor.go
index a713c17d..c3d409f0 100644
--- a/relay/channel/claude/adaptor.go
+++ b/relay/channel/claude/adaptor.go
@@ -9,6 +9,7 @@ import (
 	"github.com/QuantumNous/new-api/dto"
 	"github.com/QuantumNous/new-api/relay/channel"
 	relaycommon "github.com/QuantumNous/new-api/relay/common"
+	"github.com/QuantumNous/new-api/service"
 	"github.com/QuantumNous/new-api/setting/model_setting"
 	"github.com/QuantumNous/new-api/types"
 
@@ -18,9 +19,12 @@ import (
 type Adaptor struct {
 }
 
-func (a *Adaptor) ConvertGeminiRequest(*gin.Context, *relaycommon.RelayInfo, *dto.GeminiChatRequest) (any, error) {
-	//TODO implement me
-	return nil, errors.New("not implemented")
+func (a *Adaptor) ConvertGeminiRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeminiChatRequest) (any, error) {
+	openaiRequest, err := service.GeminiToOpenAIRequest(request, info)
+	if err != nil {
+		return nil, err
+	}
+	return a.ConvertOpenAIRequest(c, info, openaiRequest)
 }
 
 func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.ClaudeRequest) (any, error) {
diff --git a/relay/channel/claude/relay-claude.go b/relay/channel/claude/relay-claude.go
index 8976d063..746d6a01 100644
--- a/relay/channel/claude/relay-claude.go
+++ b/relay/channel/claude/relay-claude.go
@@ -732,6 +732,24 @@ func HandleStreamResponseData(c *gin.Context, info *relaycommon.RelayInfo, claud
 		if err != nil {
 			logger.LogError(c, "send_stream_response_failed: "+err.Error())
 		}
+	} else if info.RelayFormat == types.RelayFormatGemini {
+		response := StreamResponseClaude2OpenAI(&claudeResponse)
+
+		if !FormatClaudeResponseInfo(&claudeResponse, response, claudeInfo) {
+			return nil
+		}
+
+		geminiResponse := service.StreamResponseOpenAI2Gemini(response, info)
+		if geminiResponse == nil {
+			return nil
+		}
+		geminiResponseStr, marshalErr := common.Marshal(geminiResponse)
+		if marshalErr != nil {
+			logger.LogError(c, "marshal gemini stream response failed: "+marshalErr.Error())
+			return nil
+		}
+		c.Render(-1, common.CustomEvent{Data: "data: " + string(geminiResponseStr)})
+		_ = helper.FlushWriter(c)
 	}
 	return nil
 }
@@ -758,6 +776,18 @@ func HandleStreamFinalResponse(c *gin.Context, info *relaycommon.RelayInfo, clau
 			}
 		}
 		helper.Done(c)
+	} else if info.RelayFormat == types.RelayFormatGemini {
+		response := helper.GenerateFinalUsageResponse(claudeInfo.ResponseId, claudeInfo.Created, info.UpstreamModelName, *claudeInfo.Usage)
+		geminiResponse := service.StreamResponseOpenAI2Gemini(response, info)
+		if geminiResponse != nil {
+			geminiResponseStr, err := common.Marshal(geminiResponse)
+			if err != nil {
+				common.SysLog("marshal gemini final response failed: " + err.Error())
+				return
+			}
+			c.Render(-1, common.CustomEvent{Data: "data: " + string(geminiResponseStr)})
+			_ = helper.FlushWriter(c)
+		}
 	}
 }
 
@@ -803,6 +833,10 @@ func HandleClaudeResponseData(c *gin.Context, info *relaycommon.RelayInfo, claud
 	}
 	// 记录拒绝原因（上游功能）
 	maybeMarkClaudeRefusal(c, claudeResponse.StopReason)
+	// 检查是否有有效的内容返回，如果没有则返回错误以触发重试
+	if len(claudeResponse.Content) == 0 && claudeResponse.Completion == "" {
+		return types.NewOpenAIError(fmt.Errorf("empty response from Claude API"), types.ErrorCodeEmptyResponse, http.StatusInternalServerError)
+	}
 	if claudeInfo.Usage == nil {
 		claudeInfo.Usage = &dto.Usage{}
 	}
@@ -824,6 +858,14 @@ func HandleClaudeResponseData(c *gin.Context, info *relaycommon.RelayInfo, claud
 		if err != nil {
 			return types.NewError(err, types.ErrorCodeBadResponseBody)
 		}
+	case types.RelayFormatGemini:
+		openaiResponse := ResponseClaude2OpenAI(&claudeResponse)
+		openaiResponse.Usage = *claudeInfo.Usage
+		geminiResponse := service.ResponseOpenAI2Gemini(openaiResponse, info)
+		responseData, err = common.Marshal(geminiResponse)
+		if err != nil {
+			return types.NewError(err, types.ErrorCodeBadResponseBody)
+		}
 	case types.RelayFormatClaude:
 		responseData = data
 	}
diff --git a/relay/channel/gemini/relay-gemini-native.go b/relay/channel/gemini/relay-gemini-native.go
index 83235f33..04a8b8cd 100644
--- a/relay/channel/gemini/relay-gemini-native.go
+++ b/relay/channel/gemini/relay-gemini-native.go
@@ -125,7 +125,7 @@ func GeminiTextGenerationStreamHandler(c *gin.Context, info *relaycommon.RelayIn
 	}
 
 	// 检查是否为空响应（上游超时或返回空内容）
-	if usage.TotalTokens == 0 {
+	if usage.CompletionTokens <= 0 {
 		return nil, types.NewOpenAIError(errors.New("empty response from Gemini API"), types.ErrorCodeEmptyResponse, http.StatusInternalServerError)
 	}
 
diff --git a/relay/channel/gemini/relay-gemini.go b/relay/channel/gemini/relay-gemini.go
index e73f0280..6b5d75d5 100644
--- a/relay/channel/gemini/relay-gemini.go
+++ b/relay/channel/gemini/relay-gemini.go
@@ -1399,7 +1399,7 @@ func GeminiChatStreamHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *
 	}
 
 	// 检查是否为空响应（上游超时或返回空内容）
-	if usage.TotalTokens == 0 {
+	if usage.CompletionTokens <= 0 {
 		return nil, types.NewOpenAIError(errors.New("empty response from Gemini API"), types.ErrorCodeEmptyResponse, http.StatusInternalServerError)
 	}
 
@@ -1426,53 +1426,21 @@ func GeminiChatHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.R
 		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
 	}
 	if len(geminiResponse.Candidates) == 0 {
-		usage := dto.Usage{
-			PromptTokens: geminiResponse.UsageMetadata.PromptTokenCount,
-		}
-		usage.CompletionTokenDetails.ReasoningTokens = geminiResponse.UsageMetadata.ThoughtsTokenCount
-		usage.PromptTokensDetails.CachedTokens = geminiResponse.UsageMetadata.CachedContentTokenCount
-		for _, detail := range geminiResponse.UsageMetadata.PromptTokensDetails {
-			if detail.Modality == "AUDIO" {
-				usage.PromptTokensDetails.AudioTokens = detail.TokenCount
-			} else if detail.Modality == "TEXT" {
-				usage.PromptTokensDetails.TextTokens = detail.TokenCount
-			}
-		}
-		if usage.PromptTokens <= 0 {
-			usage.PromptTokens = info.GetEstimatePromptTokens()
-		}
-
-		var newAPIError *types.NewAPIError
 		if geminiResponse.PromptFeedback != nil && geminiResponse.PromptFeedback.BlockReason != nil {
 			common.SetContextKey(c, constant.ContextKeyAdminRejectReason, fmt.Sprintf("gemini_block_reason=%s", *geminiResponse.PromptFeedback.BlockReason))
-			newAPIError = types.NewOpenAIError(
+			return nil, types.NewOpenAIError(
 				errors.New("request blocked by Gemini API: "+*geminiResponse.PromptFeedback.BlockReason),
 				types.ErrorCodePromptBlocked,
 				http.StatusBadRequest,
+				types.ErrOptionWithSkipRetry(),
 			)
-		} else {
-			common.SetContextKey(c, constant.ContextKeyAdminRejectReason, "gemini_empty_candidates")
-			newAPIError = types.NewOpenAIError(
-				errors.New("empty response from Gemini API"),
-				types.ErrorCodeEmptyResponse,
-				http.StatusInternalServerError,
-			)
-		}
-
-		service.ResetStatusCode(newAPIError, c.GetString("status_code_mapping"))
-
-		switch info.RelayFormat {
-		case types.RelayFormatClaude:
-			c.JSON(newAPIError.StatusCode, gin.H{
-				"type":  "error",
-				"error": newAPIError.ToClaudeError(),
-			})
-		default:
-			c.JSON(newAPIError.StatusCode, gin.H{
-				"error": newAPIError.ToOpenAIError(),
-			})
 		}
-		return &usage, nil
+		common.SetContextKey(c, constant.ContextKeyAdminRejectReason, "gemini_empty_candidates")
+		return nil, types.NewOpenAIError(
+			errors.New("empty response from Gemini API"),
+			types.ErrorCodeEmptyResponse,
+			http.StatusInternalServerError,
+		)
 	}
 	fullTextResponse := responseGeminiChat2OpenAI(c, &geminiResponse)
 	fullTextResponse.Model = info.UpstreamModelName
```

---

### 提交 6: feat: 更新视频代理逻辑，支持直接返回内联 base64 视频数据并优化错误处理

- **哈希**: `6917dbcdebb920d9ed4f1f52e163fe84b7f70258`
- **日期**: 2026-02-18 21:30:16 +0800
- **作者**: Mingluan Mu <mumingluan@qq.com>

**完整 Diff:**

```diff
diff --git a/controller/video_proxy.go b/controller/video_proxy.go
index f102baae..0e2f36e6 100644
--- a/controller/video_proxy.go
+++ b/controller/video_proxy.go
@@ -118,17 +118,33 @@ func VideoProxy(c *gin.Context) {
 			return
 		}
 
-		videoURL, err = getGeminiVideoURL(channel, task, apiKey)
+		result, err := getGeminiVideoResult(channel, task, apiKey)
 		if err != nil {
-			logger.LogError(c.Request.Context(), fmt.Sprintf("Failed to resolve Gemini video URL for task %s: %s", taskID, err.Error()))
+			logger.LogError(c.Request.Context(), fmt.Sprintf("Failed to resolve Gemini video for task %s: %s", taskID, err.Error()))
 			c.JSON(http.StatusBadGateway, gin.H{
 				"error": gin.H{
-					"message": "Failed to resolve Gemini video URL",
+					"message": "Failed to resolve Gemini video",
 					"type":    "server_error",
 				},
 			})
 			return
 		}
+
+		// Inline base64 data: write directly without proxying
+		if result.Data != nil {
+			mime := result.MimeType
+			if mime == "" {
+				mime = "video/mp4"
+			}
+			c.Header("Content-Type", mime)
+			c.Header("Content-Length", fmt.Sprintf("%d", len(result.Data)))
+			c.Header("Cache-Control", "public, max-age=86400")
+			c.Writer.WriteHeader(http.StatusOK)
+			_, _ = c.Writer.Write(result.Data)
+			return
+		}
+
+		videoURL = result.URL
 		req.Header.Set("x-goog-api-key", apiKey)
 	case constant.ChannelTypeOpenAI, constant.ChannelTypeSora:
 		videoURL = fmt.Sprintf("%s/v1/videos/%s/content", baseURL, task.TaskID)
diff --git a/controller/video_proxy_gemini.go b/controller/video_proxy_gemini.go
index 053ac651..1bd37e4d 100644
--- a/controller/video_proxy_gemini.go
+++ b/controller/video_proxy_gemini.go
@@ -1,6 +1,7 @@
 package controller
 
 import (
+	"encoding/base64"
 	"encoding/json"
 	"fmt"
 	"io"
@@ -12,13 +13,21 @@ import (
 	"github.com/QuantumNous/new-api/relay"
 )
 
-func getGeminiVideoURL(channel *model.Channel, task *model.Task, apiKey string) (string, error) {
+// geminiVideoResult holds either a remote URL or inline base64 video data.
+type geminiVideoResult struct {
+	URL      string // remote video URL (preferred if available)
+	Data     []byte // decoded video bytes from bytesBase64Encoded
+	MimeType string // e.g. "video/mp4"
+}
+
+func getGeminiVideoResult(channel *model.Channel, task *model.Task, apiKey string) (*geminiVideoResult, error) {
 	if channel == nil || task == nil {
-		return "", fmt.Errorf("invalid channel or task")
+		return nil, fmt.Errorf("invalid channel or task")
 	}
 
+	// Try extracting URL from stored task data first
 	if url := extractGeminiVideoURLFromTaskData(task); url != "" {
-		return ensureAPIKey(url, apiKey), nil
+		return &geminiVideoResult{URL: ensureAPIKey(url, apiKey)}, nil
 	}
 
 	baseURL := constant.ChannelBaseURLs[channel.Type]
@@ -28,11 +37,11 @@ func getGeminiVideoURL(channel *model.Channel, task *model.Task, apiKey string)
 
 	adaptor := relay.GetTaskAdaptor(constant.TaskPlatform(strconv.Itoa(channel.Type)))
 	if adaptor == nil {
-		return "", fmt.Errorf("gemini task adaptor not found")
+		return nil, fmt.Errorf("gemini task adaptor not found")
 	}
 
 	if apiKey == "" {
-		return "", fmt.Errorf("api key not available for task")
+		return nil, fmt.Errorf("api key not available for task")
 	}
 
 	proxy := channel.GetSetting().Proxy
@@ -41,29 +50,77 @@ func getGeminiVideoURL(channel *model.Channel, task *model.Task, apiKey string)
 		"action":  task.Action,
 	}, proxy)
 	if err != nil {
-		return "", fmt.Errorf("fetch task failed: %w", err)
+		return nil, fmt.Errorf("fetch task failed: %w", err)
 	}
 	defer resp.Body.Close()
 
 	body, err := io.ReadAll(resp.Body)
 	if err != nil {
-		return "", fmt.Errorf("read task response failed: %w", err)
+		return nil, fmt.Errorf("read task response failed: %w", err)
 	}
 
+	// Try remote URL from parsed result
 	taskInfo, parseErr := adaptor.ParseTaskResult(body)
 	if parseErr == nil && taskInfo != nil && taskInfo.RemoteUrl != "" {
-		return ensureAPIKey(taskInfo.RemoteUrl, apiKey), nil
+		return &geminiVideoResult{URL: ensureAPIKey(taskInfo.RemoteUrl, apiKey)}, nil
 	}
 
+	// Try remote URL from raw payload
 	if url := extractGeminiVideoURLFromPayload(body); url != "" {
-		return ensureAPIKey(url, apiKey), nil
+		return &geminiVideoResult{URL: ensureAPIKey(url, apiKey)}, nil
+	}
+
+	// Try inline base64 video data (Veo returns bytesBase64Encoded directly)
+	if result := extractGeminiVideoBase64FromPayload(body); result != nil {
+		return result, nil
 	}
 
 	if parseErr != nil {
-		return "", fmt.Errorf("parse task result failed: %w", parseErr)
+		return nil, fmt.Errorf("parse task result failed: %w", parseErr)
+	}
+
+	return nil, fmt.Errorf("gemini video data not found")
+}
+
+// extractGeminiVideoBase64FromPayload extracts inline base64 video from fetchPredictOperation response.
+func extractGeminiVideoBase64FromPayload(body []byte) *geminiVideoResult {
+	var payload map[string]any
+	if err := json.Unmarshal(body, &payload); err != nil {
+		return nil
+	}
+
+	resp, ok := payload["response"].(map[string]any)
+	if !ok {
+		return nil
+	}
+
+	// Check response.videos[0].bytesBase64Encoded
+	if videos, ok := resp["videos"].([]any); ok && len(videos) > 0 {
+		if vm, ok := videos[0].(map[string]any); ok {
+			if b64, ok := vm["bytesBase64Encoded"].(string); ok && b64 != "" {
+				data, err := base64.StdEncoding.DecodeString(b64)
+				if err != nil {
+					return nil
+				}
+				mime := "video/mp4"
+				if m, ok := vm["mimeType"].(string); ok && m != "" {
+					mime = m
+				}
+				return &geminiVideoResult{Data: data, MimeType: mime}
+			}
+		}
+	}
+
+	// Check response.bytesBase64Encoded directly
+	if b64, ok := resp["bytesBase64Encoded"].(string); ok && b64 != "" {
+		data, err := base64.StdEncoding.DecodeString(b64)
+		if err != nil {
+			return nil
+		}
+		return &geminiVideoResult{Data: data, MimeType: "video/mp4"}
 	}
 
-	return "", fmt.Errorf("gemini video url not found")
+	return nil
 }
 
 func extractGeminiVideoURLFromTaskData(task *model.Task) string {
diff --git a/middleware/model-rate-limit.go b/middleware/model-rate-limit.go
index b7f49612..285ea5df 100644
--- a/middleware/model-rate-limit.go
+++ b/middleware/model-rate-limit.go
@@ -175,17 +175,22 @@ func memoryRateLimitHandler(duration int64, totalMaxCount, successMaxCount int)
 // checkTokenRateLimit 检查 token 分钟级限流
 func checkTokenRateLimit(c *gin.Context) bool {
 	if !setting.TokenRateLimitEnabled {
+		fmt.Println("[DEBUG] TokenRateLimitEnabled is false, skipping minute rate limit")
 		return true
 	}
 
 	tokenId := common.GetContextKeyInt(c, constant.ContextKeyTokenId)
 	if tokenId == 0 {
 		// 如果没有 token ID，跳过 per-key 限流
+		fmt.Println("[DEBUG] tokenId is 0, skipping minute rate limit")
 		return true
 	}
 
-	// 获取分组配置（使用 token group）
+	// 获取分组配置（优先使用 token group，如果为空则使用 user group）
 	group := common.GetContextKeyString(c, constant.ContextKeyTokenGroup)
+	if group == "" {
+		group = common.GetContextKeyString(c, constant.ContextKeyUserGroup)
+	}
 	totalMaxCount := setting.TokenRateLimitCount
 	successMaxCount := setting.TokenRateLimitSuccessCount
 
@@ -196,8 +201,12 @@ func checkTokenRateLimit(c *gin.Context) bool {
 		successMaxCount = groupSuccessCount
 	}
 
+	fmt.Printf("[DEBUG] Minute rate limit check: tokenId=%d, group=%s, found=%v, totalMaxCount=%d, successMaxCount=%d\n",
+		tokenId, group, found, totalMaxCount, successMaxCount)
+
 	// 如果两个限制都为0，表示不限制
 	if totalMaxCount == 0 && successMaxCount == 0 {
+		fmt.Println("[DEBUG] Both minute limits are 0, skipping minute rate limit")
 		return true
 	}
 
@@ -293,8 +302,11 @@ func recordTokenRateLimitSuccess(c *gin.Context) {
 		return
 	}
 
-	// 获取分组配置
+	// 获取分组配置（优先使用 token group，如果为空则使用 user group）
 	group := common.GetContextKeyString(c, constant.ContextKeyTokenGroup)
+	if group == "" {
+		group = common.GetContextKeyString(c, constant.ContextKeyUserGroup)
+	}
 	successMaxCount := setting.TokenRateLimitSuccessCount
 
 	_, groupSuccessCount, found := setting.GetTokenRateLimit(group)
@@ -324,17 +336,22 @@ func recordTokenRateLimitSuccess(c *gin.Context) {
 // checkTokenDailyRateLimit 检查 token 每日限流
 func checkTokenDailyRateLimit(c *gin.Context) bool {
 	if !setting.TokenDailyRateLimitEnabled {
+		fmt.Println("[DEBUG] TokenDailyRateLimitEnabled is false, skipping daily rate limit")
 		return true
 	}
 
 	tokenId := common.GetContextKeyInt(c, constant.ContextKeyTokenId)
 	if tokenId == 0 {
 		// 如果没有 token ID，跳过 per-key 限流
+		fmt.Println("[DEBUG] tokenId is 0, skipping daily rate limit")
 		return true
 	}
 
-	// 获取分组配置
+	// 获取分组配置（优先使用 token group，如果为空则使用 user group）
 	group := common.GetContextKeyString(c, constant.ContextKeyTokenGroup)
+	if group == "" {
+		group = common.GetContextKeyString(c, constant.ContextKeyUserGroup)
+	}
 	totalMaxCount := setting.TokenDailyRateLimitCount
 	successMaxCount := setting.TokenDailyRateLimitSuccessCount
 
@@ -345,13 +362,27 @@ func checkTokenDailyRateLimit(c *gin.Context) bool {
 		successMaxCount = groupSuccessCount
 	}
 
+	fmt.Printf("[DEBUG] Daily rate limit check: tokenId=%d, group=%s, found=%v, totalMaxCount=%d, successMaxCount=%d\n",
+		tokenId, group, found, totalMaxCount, successMaxCount)
+
 	// 如果两个限制都为0，表示不限制
 	if totalMaxCount == 0 && successMaxCount == 0 {
+		fmt.Println("[DEBUG] Both limits are 0, skipping daily rate limit")
 		return true
 	}
 
-	rateLimitKey := strconv.Itoa(tokenId)
-	duration := int64(86400) // 24小时 = 86400秒
+	// 使用北京时间 (UTC+8) 计算当日的 key 后缀（格式：YYYYMMDD）
+	// 这样每个自然日（北京时间）都有独立的 key，到次日 00:00 自动失效
+	beijingLoc := time.FixedZone("CST", 8*3600) // UTC+8
+	beijingNow := time.Now().In(beijingLoc)
+	dateKey := beijingNow.Format("20060102")
+	rateLimitKey := fmt.Sprintf("%d:%s", tokenId, dateKey)
+
+	// 计算到北京时间次日 00:00 的剩余秒数
+	nextDay := time.Date(beijingNow.Year(), beijingNow.Month(), beijingNow.Day()+1, 0, 0, 0, 0, beijingLoc)
+	duration := int64(nextDay.Sub(beijingNow).Seconds()) + 60 // 额外 60 秒缓冲
+
+	fmt.Printf("[DEBUG] Daily rate limit key: %s, duration until next Beijing day: %d seconds\n", rateLimitKey, duration)
 
 	if common.RedisEnabled {
 		return checkTokenDailyRateLimitRedis(c, rateLimitKey, totalMaxCount, successMaxCount, duration)
@@ -361,52 +392,72 @@ func checkTokenDailyRateLimit(c *gin.Context) bool {
 }
 
 // checkTokenDailyRateLimitRedis Redis版本的每日限流检查
+// 使用固定窗口计数器实现每日配额限制
 func checkTokenDailyRateLimitRedis(c *gin.Context, rateLimitKey string, totalMaxCount, successMaxCount int, duration int64) bool {
 	ctx := context.Background()
 	rdb := common.RDB
 
-	// 1. 检查成功请求数限制
-	if successMaxCount > 0 {
-		successKey := fmt.Sprintf("rateLimit:%s:%s", TokenDailyRateLimitSuccessCountMark, rateLimitKey)
-		allowed, err := checkRedisRateLimit(ctx, rdb, successKey, successMaxCount, duration)
+	// Lua 脚本：原子性地增加计数并检查是否超限
+	// 如果超限则回滚（减少计数）并返回 0，否则返回当前计数
+	luaScript := `
+		local count = redis.call('INCR', KEYS[1])
+		if count == 1 then
+			redis.call('EXPIRE', KEYS[1], ARGV[1])
+		end
+		if count > tonumber(ARGV[2]) then
+			redis.call('DECR', KEYS[1])
+			return 0
+		end
+		return count
+	`
+
+	// 1. 检查总请求数限制（先检查总数，因为总数包括失败请求）
+	if totalMaxCount > 0 {
+		totalKey := fmt.Sprintf("rateLimit:%s:%s", TokenDailyRateLimitCountMark, rateLimitKey)
+		count, err := rdb.Eval(ctx, luaScript, []string{totalKey}, duration, totalMaxCount).Int64()
 		if err != nil {
-			fmt.Println("检查每日成功请求数限制失败:", err.Error())
+			fmt.Println("检查每日总请求数限制失败:", err.Error())
 			abortWithOpenAiMessage(c, http.StatusInternalServerError, "rate_limit_check_failed")
 			return false
 		}
-		if !allowed {
-			abortWithOpenAiMessage(c, http.StatusTooManyRequests, "您已达到每日请求数限制")
+		if count == 0 {
+			abortWithOpenAiMessage(c, http.StatusTooManyRequests, "您已达到每日总请求数限制（包括失败请求）")
 			return false
 		}
 	}
 
-	// 2. 检查总请求数限制
-	if totalMaxCount > 0 {
-		totalKey := fmt.Sprintf("rateLimit:%s:%s", TokenDailyRateLimitCountMark, rateLimitKey)
-		tb := limiter.New(ctx, rdb)
-		allowed, err := tb.Allow(
-			ctx,
-			totalKey,
-			limiter.WithCapacity(int64(totalMaxCount)*duration),
-			limiter.WithRate(int64(totalMaxCount)),
-			limiter.WithRequested(duration),
-		)
-
+	// 2. 检查并预占成功请求数配额
+	if successMaxCount > 0 {
+		successKey := fmt.Sprintf("rateLimit:%s:%s", TokenDailyRateLimitSuccessCountMark, rateLimitKey)
+		count, err := rdb.Eval(ctx, luaScript, []string{successKey}, duration, successMaxCount).Int64()
 		if err != nil {
-			fmt.Println("检查每日总请求数限制失败:", err.Error())
+			fmt.Println("检查每日成功请求数限制失败:", err.Error())
 			abortWithOpenAiMessage(c, http.StatusInternalServerError, "rate_limit_check_failed")
 			return false
 		}
-
-		if !allowed {
-			abortWithOpenAiMessage(c, http.StatusTooManyRequests, "您已达到每日总请求数限制（包括失败请求）")
+		if count == 0 {
+			abortWithOpenAiMessage(c, http.StatusTooManyRequests, "您已达到每日请求数限制")
 			return false
 		}
+		// 标记已预占成功请求配额，请求失败时需要回滚
+		c.Set("daily_success_quota_reserved", true)
+		c.Set("daily_success_quota_key", successKey)
 	}
 
 	return true
 }
 
+// rollbackDailySuccessQuota 回滚预占的成功请求配额
+func rollbackDailySuccessQuota(c *gin.Context) {
+	if reserved, exists := c.Get("daily_success_quota_reserved"); exists && reserved.(bool) {
+		if key, exists := c.Get("daily_success_quota_key"); exists {
+			ctx := context.Background()
+			common.RDB.Decr(ctx, key.(string))
+			c.Set("daily_success_quota_reserved", false)
+		}
+	}
+}
+
 // checkTokenDailyRateLimitMemory 内存版本的每日限流检查
 func checkTokenDailyRateLimitMemory(c *gin.Context, rateLimitKey string, totalMaxCount, successMaxCount int, duration int64) bool {
 	inMemoryRateLimiter.Init(24 * time.Hour)
@@ -431,45 +482,53 @@ func checkTokenDailyRateLimitMemory(c *gin.Context, rateLimitKey string, totalMa
 	return true
 }
 
-// recordTokenDailySuccess 记录每日成功请求
+// recordTokenDailySuccess 处理每日成功请求计数
+// 由于成功请求配额在检查时已预占，这里只需要处理请求失败时的回滚
+// 请求成功时不需要额外操作（配额已预占）
 func recordTokenDailySuccess(c *gin.Context) {
-	if !setting.TokenDailyRateLimitEnabled {
-		return
-	}
-
-	tokenId := common.GetContextKeyInt(c, constant.ContextKeyTokenId)
-	if tokenId == 0 {
-		return
-	}
+	// 请求成功，配额已在检查时预占，清除标记即可
+	if common.RedisEnabled {
+		c.Set("daily_success_quota_reserved", false)
+	} else {
+		// 内存模式：请求成功时记录
+		if !setting.TokenDailyRateLimitEnabled {
+			return
+		}
 
-	// 获取分组配置
-	group := common.GetContextKeyString(c, constant.ContextKeyTokenGroup)
-	successMaxCount := setting.TokenDailyRateLimitSuccessCount
+		tokenId := common.GetContextKeyInt(c, constant.ContextKeyTokenId)
+		if tokenId == 0 {
+			return
+		}
 
-	_, groupSuccessCount, found := setting.GetTokenDailyRateLimit(group)
-	if found {
-		successMaxCount = groupSuccessCount
-	}
+		group := common.GetContextKeyString(c, constant.ContextKeyTokenGroup)
+		if group == "" {
+			group = common.GetContextKeyString(c, constant.ContextKeyUserGroup)
+		}
+		successMaxCount := setting.TokenDailyRateLimitSuccessCount
 
-	if successMaxCount == 0 {
-		return
-	}
+		_, groupSuccessCount, found := setting.GetTokenDailyRateLimit(group)
+		if found {
+			successMaxCount = groupSuccessCount
+		}
 
-	rateLimitKey := strconv.Itoa(tokenId)
+		if successMaxCount == 0 {
+			return
+		}
 
-	if common.RedisEnabled {
-		ctx := context.Background()
-		rdb := common.RDB
-		successKey := fmt.Sprintf("rateLimit:%s:%s", TokenDailyRateLimitSuccessCountMark, rateLimitKey)
-		duration := int64(86400)
-		recordRedisRequest(ctx, rdb, successKey, successMaxCount, duration)
-	} else {
+		rateLimitKey := strconv.Itoa(tokenId)
 		duration := int64(86400)
 		successKey := TokenDailyRateLimitSuccessCountMark + rateLimitKey
 		inMemoryRateLimiter.Request(successKey, successMaxCount, duration)
 	}
 }
 
+// rollbackTokenDailyQuota 请求失败时回滚预占的每日配额
+func rollbackTokenDailyQuota(c *gin.Context) {
+	if common.RedisEnabled {
+		rollbackDailySuccessQuota(c)
+	}
+}
+
 // ModelRequestRateLimit 模型请求限流中间件
 func ModelRequestRateLimit() func(c *gin.Context) {
 	return func(c *gin.Context) {
@@ -486,10 +545,12 @@ func ModelRequestRateLimit() func(c *gin.Context) {
 		// 3. 再检查原有的 per-user 限流（保持兼容性）
 		if !setting.ModelRequestRateLimitEnabled {
 			c.Next()
-			// 请求成功后记录 per-key 成功请求
+			// 请求成功后确认配额，失败则回滚
 			if c.Writer.Status() < 400 {
 				recordTokenRateLimitSuccess(c)
 				recordTokenDailySuccess(c)
+			} else {
+				rollbackTokenDailyQuota(c)
 			}
 			return
 		}
@@ -516,10 +577,12 @@ func ModelRequestRateLimit() func(c *gin.Context) {
 			memoryRateLimitHandler(duration, totalMaxCount, successMaxCount)(c)
 		}
 
-		// 请求成功后记录 per-key 成功请求
+		// 请求成功后确认配额，失败则回滚
 		if c.Writer.Status() < 400 {
 			recordTokenRateLimitSuccess(c)
 			recordTokenDailySuccess(c)
+		} else {
+			rollbackTokenDailyQuota(c)
 		}
 	}
 }
diff --git a/relay/channel/gemini/relay-gemini-native.go b/relay/channel/gemini/relay-gemini-native.go
index 29af6303..83235f33 100644
--- a/relay/channel/gemini/relay-gemini-native.go
+++ b/relay/channel/gemini/relay-gemini-native.go
@@ -110,7 +110,7 @@ func NativeGeminiEmbeddingHandler(c *gin.Context, resp *http.Response, info *rel
 func GeminiTextGenerationStreamHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.NewAPIError) {
 	helper.SetEventStreamHeaders(c)
 
-	return geminiStreamHandler(c, info, resp, func(data string, geminiResponse *dto.GeminiChatResponse) bool {
+	usage, err := geminiStreamHandler(c, info, resp, func(data string, geminiResponse *dto.GeminiChatResponse) bool {
 		err := helper.StringData(c, data)
 		if err != nil {
 			logger.LogError(c, "failed to write stream data: "+err.Error())
@@ -119,4 +119,15 @@ func GeminiTextGenerationStreamHandler(c *gin.Context, info *relaycommon.RelayIn
 		info.SendResponseCount++
 		return true
 	})
+
+	if err != nil {
+		return usage, err
+	}
+
+	// 检查是否为空响应（上游超时或返回空内容）
+	if usage.TotalTokens == 0 {
+		return nil, types.NewOpenAIError(errors.New("empty response from Gemini API"), types.ErrorCodeEmptyResponse, http.StatusInternalServerError)
+	}
+
+	return usage, nil
 }
diff --git a/relay/channel/gemini/relay-gemini.go b/relay/channel/gemini/relay-gemini.go
index b10ec06c..e73f0280 100644
--- a/relay/channel/gemini/relay-gemini.go
+++ b/relay/channel/gemini/relay-gemini.go
@@ -1319,6 +1319,11 @@ func GeminiChatStreamHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *
 	nextToolCallIndexByChoice := make(map[int]int)
 
 	usage, err := geminiStreamHandler(c, info, resp, func(data string, geminiResponse *dto.GeminiChatResponse) bool {
+		// 跳过空的 Candidates 响应
+		if len(geminiResponse.Candidates) == 0 {
+			return true
+		}
+
 		response, isStop := streamResponseGeminiChat2OpenAI(geminiResponse)
 
 		response.Id = id
@@ -1393,6 +1398,11 @@ func GeminiChatStreamHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *
 		return usage, err
 	}
 
+	// 检查是否为空响应（上游超时或返回空内容）
+	if usage.TotalTokens == 0 {
+		return nil, types.NewOpenAIError(errors.New("empty response from Gemini API"), types.ErrorCodeEmptyResponse, http.StatusInternalServerError)
+	}
+
 	response := helper.GenerateFinalUsageResponse(id, createAt, info.UpstreamModelName, *usage)
 	handleErr := handleFinalStream(c, info, response)
 	if handleErr != nil {
diff --git a/relay/channel/task/gemini/adaptor.go b/relay/channel/task/gemini/adaptor.go
index 16c6919b..8355b739 100644
--- a/relay/channel/task/gemini/adaptor.go
+++ b/relay/channel/task/gemini/adaptor.go
@@ -211,15 +211,28 @@ func (a *TaskAdaptor) FetchTask(baseUrl, key string, body map[string]any, proxy
 		return nil, fmt.Errorf("decode task_id failed: %w", err)
 	}
 
-	// For Gemini API, we use GET request to the operations endpoint
-	version := model_setting.GetGeminiVersionSetting("default")
-	url := fmt.Sprintf("%s/%s/%s", baseUrl, version, upstreamName)
+	// Extract model name from operation name to build fetchPredictOperation URL
+	modelName := extractModelFromOperationName(upstreamName)
+	if modelName == "" {
+		return nil, fmt.Errorf("cannot extract model name from operation: %s", upstreamName)
+	}
+
+	version := model_setting.GetGeminiVersionSetting(modelName)
+	url := fmt.Sprintf("%s/%s/models/%s:fetchPredictOperation", baseUrl, version, modelName)
 
-	req, err := http.NewRequest(http.MethodGet, url, nil)
+	fetchBody, err := json.Marshal(map[string]string{
+		"operationName": upstreamName,
+	})
+	if err != nil {
+		return nil, fmt.Errorf("marshal fetch body failed: %w", err)
+	}
+
+	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(fetchBody))
 	if err != nil {
 		return nil, err
 	}
 
+	req.Header.Set("Content-Type", "application/json")
 	req.Header.Set("Accept", "application/json")
 	req.Header.Set("x-goog-api-key", key)
 
diff --git a/web/src/pages/Setting/RateLimit/SettingsRequestRateLimit.jsx b/web/src/pages/Setting/RateLimit/SettingsRequestRateLimit.jsx
index c4316d0b..6cda1cf1 100644
--- a/web/src/pages/Setting/RateLimit/SettingsRequestRateLimit.jsx
+++ b/web/src/pages/Setting/RateLimit/SettingsRequestRateLimit.jsx
@@ -426,7 +426,7 @@ export default function RequestRateLimit(props) {
                         <li>{t('最大值为2147483647。')}</li>
                         <li>{t('分组配置优先级高于全局配置。')}</li>
                         <li>{t('此限制按API密钥（Token）维度，每个密钥独立计算。')}</li>
-                        <li>{t('限制周期为自然日（UTC+0），每日0点重置。')}</li>
+                        <li>{t('限制周期为自然日（北京时间 UTC+8），每日0点重置。')}</li>
                       </ul>
                     </div>
                   }
```

---

### 提交 7: feat: 更新Gemini和OpenAI适配器逻辑，优化空响应检查和工具调用处理

- **哈希**: `a378dbe8114d40d13b410a69b5fc3a708b367efa`
- **日期**: 2026-03-05 12:07:10 +0800
- **作者**: Mingluan Mu <mumingluan@qq.com>

**完整 Diff:**

```diff
diff --git a/relay/channel/gemini/relay-gemini-native.go b/relay/channel/gemini/relay-gemini-native.go
index 04a8b8cd..21050ca1 100644
--- a/relay/channel/gemini/relay-gemini-native.go
+++ b/relay/channel/gemini/relay-gemini-native.go
@@ -125,7 +125,8 @@ func GeminiTextGenerationStreamHandler(c *gin.Context, info *relaycommon.RelayIn
 	}
 
 	// 检查是否为空响应（上游超时或返回空内容）
-	if usage.CompletionTokens <= 0 {
+	// 当存在工具调用时，CompletionTokens 可能为 0，不应视为空响应
+	if usage.CompletionTokens <= 0 && info.SendResponseCount == 0 {
 		return nil, types.NewOpenAIError(errors.New("empty response from Gemini API"), types.ErrorCodeEmptyResponse, http.StatusInternalServerError)
 	}
 
diff --git a/relay/channel/gemini/relay-gemini.go b/relay/channel/gemini/relay-gemini.go
index 6b5d75d5..6b35b62d 100644
--- a/relay/channel/gemini/relay-gemini.go
+++ b/relay/channel/gemini/relay-gemini.go
@@ -1258,7 +1258,7 @@ func geminiStreamHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http
 			common.SetContextKey(c, constant.ContextKeyAdminRejectReason, fmt.Sprintf("gemini_block_reason=%s", *geminiResponse.PromptFeedback.BlockReason))
 		}
 
-		// 统计图片数量
+		// 统计图片数量和工具调用
 		for _, candidate := range geminiResponse.Candidates {
 			for _, part := range candidate.Content.Parts {
 				if part.InlineData != nil && part.InlineData.MimeType != "" {
@@ -1267,6 +1267,12 @@ func geminiStreamHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http
 				if part.Text != "" {
 					responseText.WriteString(part.Text)
 				}
+				if part.FunctionCall != nil {
+					responseText.WriteString(part.FunctionCall.FunctionName)
+					if args, err := common.Marshal(part.FunctionCall.Arguments); err == nil {
+						responseText.WriteString(string(args))
+					}
+				}
 			}
 		}
 
@@ -1399,7 +1405,8 @@ func GeminiChatStreamHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *
 	}
 
 	// 检查是否为空响应（上游超时或返回空内容）
-	if usage.CompletionTokens <= 0 {
+	// 当存在工具调用时，CompletionTokens 可能为 0，不应视为空响应
+	if usage.CompletionTokens <= 0 && info.SendResponseCount == 0 {
 		return nil, types.NewOpenAIError(errors.New("empty response from Gemini API"), types.ErrorCodeEmptyResponse, http.StatusInternalServerError)
 	}
 
diff --git a/relay/channel/jina/adaptor.go b/relay/channel/jina/adaptor.go
index 3f2d01d9..0dd28a75 100644
--- a/relay/channel/jina/adaptor.go
+++ b/relay/channel/jina/adaptor.go
@@ -85,7 +85,7 @@ func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycom
 	if info.RelayMode == constant.RelayModeRerank {
 		usage, err = common_handler.RerankHandler(c, info, resp)
 	} else if info.RelayMode == constant.RelayModeEmbeddings {
-		usage, err = openai.OpenaiHandler(c, info, resp)
+		usage, err = openai.OpenaiHandlerWithUsage(c, info, resp)
 	}
 	return
 }
diff --git a/relay/channel/ollama/adaptor.go b/relay/channel/ollama/adaptor.go
index a3013e2f..80a05184 100644
--- a/relay/channel/ollama/adaptor.go
+++ b/relay/channel/ollama/adaptor.go
@@ -11,6 +11,7 @@ import (
 	"github.com/QuantumNous/new-api/relay/channel/openai"
 	relaycommon "github.com/QuantumNous/new-api/relay/common"
 	relayconstant "github.com/QuantumNous/new-api/relay/constant"
+	"github.com/QuantumNous/new-api/service"
 	"github.com/QuantumNous/new-api/types"
 
 	"github.com/gin-gonic/gin"
@@ -19,8 +20,12 @@ import (
 type Adaptor struct {
 }
 
-func (a *Adaptor) ConvertGeminiRequest(*gin.Context, *relaycommon.RelayInfo, *dto.GeminiChatRequest) (any, error) {
-	return nil, errors.New("not implemented")
+func (a *Adaptor) ConvertGeminiRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeminiChatRequest) (any, error) {
+	openaiRequest, err := service.GeminiToOpenAIRequest(request, info)
+	if err != nil {
+		return nil, err
+	}
+	return a.ConvertOpenAIRequest(c, info, openaiRequest)
 }
 
 func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.ClaudeRequest) (any, error) {
diff --git a/relay/channel/openai/adaptor.go b/relay/channel/openai/adaptor.go
index b6954423..f21d7756 100644
--- a/relay/channel/openai/adaptor.go
+++ b/relay/channel/openai/adaptor.go
@@ -617,6 +617,8 @@ func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycom
 		err, usage = OpenaiSTTHandler(c, resp, info, a.ResponseFormat)
 	case relayconstant.RelayModeImagesGenerations, relayconstant.RelayModeImagesEdits:
 		usage, err = OpenaiHandlerWithUsage(c, info, resp)
+	case relayconstant.RelayModeEmbeddings:
+		usage, err = OpenaiHandlerWithUsage(c, info, resp)
 	case relayconstant.RelayModeRerank:
 		usage, err = common_handler.RerankHandler(c, info, resp)
 	case relayconstant.RelayModeResponses:
diff --git a/relay/channel/openai/relay-openai.go b/relay/channel/openai/relay-openai.go
index c07f2241..706b51c8 100644
--- a/relay/channel/openai/relay-openai.go
+++ b/relay/channel/openai/relay-openai.go
@@ -12,6 +12,7 @@ import (
 	"github.com/QuantumNous/new-api/logger"
 	"github.com/QuantumNous/new-api/relay/channel/openrouter"
 	relaycommon "github.com/QuantumNous/new-api/relay/common"
+	relayconstant "github.com/QuantumNous/new-api/relay/constant"
 	"github.com/QuantumNous/new-api/relay/helper"
 	"github.com/QuantumNous/new-api/service"
 
@@ -235,7 +236,8 @@ func OpenaiHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Respo
 	}
 
 	// 检查是否有有效的 choices 返回，如果没有则返回错误以触发重试
-	if len(simpleResponse.Choices) == 0 {
+	// 仅对聊天补全和补全模式检查，嵌入等模式的响应没有 choices 字段
+	if (info.RelayMode == relayconstant.RelayModeChatCompletions || info.RelayMode == relayconstant.RelayModeCompletions) && len(simpleResponse.Choices) == 0 {
 		return nil, types.NewOpenAIError(fmt.Errorf("empty response from upstream API"), types.ErrorCodeEmptyResponse, http.StatusInternalServerError)
 	}
 
@@ -257,7 +259,13 @@ func OpenaiHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Respo
 		completionTokens := simpleResponse.Usage.CompletionTokens
 		if completionTokens == 0 {
 			for _, choice := range simpleResponse.Choices {
-				ctkm := service.CountTextToken(choice.Message.StringContent()+choice.Message.ReasoningContent+choice.Message.Reasoning, info.UpstreamModelName)
+				textContent := choice.Message.StringContent() + choice.Message.ReasoningContent + choice.Message.Reasoning
+				if toolCalls := choice.Message.ParseToolCalls(); len(toolCalls) > 0 {
+					for _, tool := range toolCalls {
+						textContent += tool.Function.Name + tool.Function.Arguments
+					}
+				}
+				ctkm := service.CountTextToken(textContent, info.UpstreamModelName)
 				completionTokens += ctkm
 			}
 		}
diff --git a/web/src/components/auth/LoginForm.jsx b/web/src/components/auth/LoginForm.jsx
index 636317e4..ec939ef6 100644
--- a/web/src/components/auth/LoginForm.jsx
+++ b/web/src/components/auth/LoginForm.jsx
@@ -505,6 +505,11 @@ const LoginForm = () => {
                 {t('登 录')}
               </Title>
             </div>
+            <div className='px-6 pb-0 text-center'>
+              <Text size='small' className='text-gray-500'>
+                {t('控制台登录仅供管理员使用，用户无需登录。请前往')}<Link to='/' className='text-blue-600 hover:text-blue-800'>{t('首页')}</Link>{t('查看教程。')}
+              </Text>
+            </div>
             <div className='px-2 py-8'>
               <div className='space-y-3'>
                 {status.wechat_login && (
@@ -719,6 +724,11 @@ const LoginForm = () => {
                 {t('登 录')}
               </Title>
             </div>
+            <div className='px-6 pb-0 text-center'>
+              <Text size='small' className='text-gray-500'>
+                {t('控制台登录仅供管理员使用，用户无需登录。请前往')}<Link to='/' className='text-blue-600 hover:text-blue-800'>{t('首页')}</Link>{t('查看教程。')}
+              </Text>
+            </div>
             <div className='px-2 py-8'>
               {status.passkey_login && passkeySupported && (
                 <Button
diff --git a/web/src/components/layout/headerbar/UserArea.jsx b/web/src/components/layout/headerbar/UserArea.jsx
index 9fc011da..93aabb88 100644
--- a/web/src/components/layout/headerbar/UserArea.jsx
+++ b/web/src/components/layout/headerbar/UserArea.jsx
@@ -142,58 +142,7 @@ const UserArea = ({
       </div>
     );
   } else {
-    const showRegisterButton = !isSelfUseMode;
-
-    const commonSizingAndLayoutClass =
-      'flex items-center justify-center !py-[10px] !px-1.5';
-
-    const loginButtonSpecificStyling =
-      '!bg-semi-color-fill-0 dark:!bg-semi-color-fill-1 hover:!bg-semi-color-fill-1 dark:hover:!bg-gray-700 transition-colors';
-    let loginButtonClasses = `${commonSizingAndLayoutClass} ${loginButtonSpecificStyling}`;
-
-    let registerButtonClasses = `${commonSizingAndLayoutClass}`;
-
-    const loginButtonTextSpanClass =
-      '!text-xs !text-semi-color-text-1 dark:!text-gray-300 !p-1.5';
-    const registerButtonTextSpanClass = '!text-xs !text-white !p-1.5';
-
-    if (showRegisterButton) {
-      if (isMobile) {
-        loginButtonClasses += ' !rounded-full';
-      } else {
-        loginButtonClasses += ' !rounded-l-full !rounded-r-none';
-      }
-      registerButtonClasses += ' !rounded-r-full !rounded-l-none';
-    } else {
-      loginButtonClasses += ' !rounded-full';
-    }
-
-    return (
-      <div className='flex items-center'>
-        <Link to='/login' className='flex'>
-          <Button
-            theme='borderless'
-            type='tertiary'
-            className={loginButtonClasses}
-          >
-            <span className={loginButtonTextSpanClass}>{t('登录')}</span>
-          </Button>
-        </Link>
-        {showRegisterButton && (
-          <div className='hidden md:block'>
-            <Link to='/register' className='flex -ml-px'>
-              <Button
-                theme='solid'
-                type='primary'
-                className={registerButtonClasses}
-              >
-                <span className={registerButtonTextSpanClass}>{t('注册')}</span>
-              </Button>
-            </Link>
-          </div>
-        )}
-      </div>
-    );
+    return null;
   }
 };
 
```

---

### 提交 8: feat: 更新EditRedemptionModal和EditTokenModal，添加新的金额选项

- **哈希**: `87065f8f925e26d89883cb08f9757acc19c5db76`
- **日期**: 2026-03-17 22:24:35 +0800
- **作者**: Mingluan Mu <mumingluan@qq.com>

**完整 Diff:**

```diff
diff --git a/web/src/components/table/redemptions/modals/EditRedemptionModal.jsx b/web/src/components/table/redemptions/modals/EditRedemptionModal.jsx
index bcde7260..c107b1e0 100644
--- a/web/src/components/table/redemptions/modals/EditRedemptionModal.jsx
+++ b/web/src/components/table/redemptions/modals/EditRedemptionModal.jsx
@@ -313,6 +313,14 @@ const EditRedemptionModal = (props) => {
                           { value: 50000000, label: '100$' },
                           { value: 250000000, label: '500$' },
                           { value: 500000000, label: '1000$' },
+                          { value: 7500000, label: '15$' },
+                          { value: 12500000, label: '25$' },
+                          { value: 20000000, label: '40$' },
+                          { value: 27500000, label: '55$' },
+                          { value: 32500000, label: '65$' },
+                          { value: 40000000, label: '80$' },
+                          { value: 132500000, label: '265$' },
+                          { value: 267500000, label: '535$' },
                         ]}
                         showClear
                       />
diff --git a/web/src/components/table/tokens/modals/EditTokenModal.jsx b/web/src/components/table/tokens/modals/EditTokenModal.jsx
index 4da55785..c1810d88 100644
--- a/web/src/components/table/tokens/modals/EditTokenModal.jsx
+++ b/web/src/components/table/tokens/modals/EditTokenModal.jsx
@@ -510,6 +510,14 @@ const EditTokenModal = (props) => {
                         { value: 50000000, label: '100$' },
                         { value: 250000000, label: '500$' },
                         { value: 500000000, label: '1000$' },
+                        { value: 12500000, label: '25$' },
+                        { value: 20000000, label: '40$' },
+                        { value: 27500000, label: '55$' },
+                        { value: 32500000, label: '65$' },
+                        { value: 40000000, label: '80$' },
+                        { value: 132500000, label: '265$' },
+                        { value: 267500000, label: '535$' },
+
                       ]}
                     />
                   </Col>
```

---


## 关键文件当前完整内容

> 以下是被修改的关键文件的当前完整内容，方便直接参考重新实现。
> 仅包含与自定义修改直接相关的核心文件。

### `middleware/model-rate-limit.go`

```go
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/common/limiter"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/setting"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const (
	ModelRequestRateLimitCountMark        = "MRRL"
	ModelRequestRateLimitSuccessCountMark = "MRRLS"
)

// Token rate limit constants
const (
	TokenRateLimitCountMark             = "TRL"
	TokenRateLimitSuccessCountMark      = "TRLS"
	TokenDailyRateLimitCountMark        = "TDRL"
	TokenDailyRateLimitSuccessCountMark = "TDRLS"
)

// 检查Redis中的请求限制
func checkRedisRateLimit(ctx context.Context, rdb *redis.Client, key string, maxCount int, duration int64) (bool, error) {
	// 如果maxCount为0，表示不限制
	if maxCount == 0 {
		return true, nil
	}

	// 获取当前计数
	length, err := rdb.LLen(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// 如果未达到限制，允许请求
	if length < int64(maxCount) {
		return true, nil
	}

	// 检查时间窗口
	oldTimeStr, _ := rdb.LIndex(ctx, key, -1).Result()
	oldTime, err := time.Parse(timeFormat, oldTimeStr)
	if err != nil {
		return false, err
	}

	nowTimeStr := time.Now().Format(timeFormat)
	nowTime, err := time.Parse(timeFormat, nowTimeStr)
	if err != nil {
		return false, err
	}
	// 如果在时间窗口内已达到限制，拒绝请求
	subTime := nowTime.Sub(oldTime).Seconds()
	if int64(subTime) < duration {
		rdb.Expire(ctx, key, time.Duration(duration)*time.Second)
		return false, nil
	}

	// 超过时间窗口，清空旧数据，允许新的时间窗口开始
	rdb.Del(ctx, key)
	return true, nil
}

// 记录Redis请求
func recordRedisRequest(ctx context.Context, rdb *redis.Client, key string, maxCount int, duration int64) {
	// 如果maxCount为0，不记录请求
	if maxCount == 0 {
		return
	}

	now := time.Now().Format(timeFormat)
	rdb.LPush(ctx, key, now)
	rdb.LTrim(ctx, key, 0, int64(maxCount-1))
	rdb.Expire(ctx, key, time.Duration(duration)*time.Second)
}

// Redis限流处理器
func redisRateLimitHandler(duration int64, totalMaxCount, successMaxCount int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := strconv.Itoa(c.GetInt("id"))
		ctx := context.Background()
		rdb := common.RDB

		// 1. 检查成功请求数限制
		successKey := fmt.Sprintf("rateLimit:%s:%s", ModelRequestRateLimitSuccessCountMark, userId)
		allowed, err := checkRedisRateLimit(ctx, rdb, successKey, successMaxCount, duration)
		if err != nil {
			fmt.Println("检查成功请求数限制失败:", err.Error())
			abortWithOpenAiMessage(c, http.StatusInternalServerError, "rate_limit_check_failed")
			return
		}
		if !allowed {
			abortWithOpenAiMessage(c, http.StatusTooManyRequests, fmt.Sprintf("您已达到请求数限制：%d分钟内最多请求%d次", setting.ModelRequestRateLimitDurationMinutes, successMaxCount))
			return
		}

		//2.检查总请求数限制并记录总请求（当totalMaxCount为0时会自动跳过，使用令牌桶限流器
		if totalMaxCount > 0 {
			totalKey := fmt.Sprintf("rateLimit:%s", userId)
			// 初始化
			tb := limiter.New(ctx, rdb)
			allowed, err = tb.Allow(
				ctx,
				totalKey,
				limiter.WithCapacity(int64(totalMaxCount)*duration),
				limiter.WithRate(int64(totalMaxCount)),
				limiter.WithRequested(duration),
			)

			if err != nil {
				fmt.Println("检查总请求数限制失败:", err.Error())
				abortWithOpenAiMessage(c, http.StatusInternalServerError, "rate_limit_check_failed")
				return
			}

			if !allowed {
				abortWithOpenAiMessage(c, http.StatusTooManyRequests, fmt.Sprintf("您已达到总请求数限制：%d分钟内最多请求%d次，包括失败次数，请检查您的请求是否正确", setting.ModelRequestRateLimitDurationMinutes, totalMaxCount))
				return
			}
		}

		// 4. 处理请求
		c.Next()

		// 5. 如果请求成功，记录成功请求
		if c.Writer.Status() < 400 {
			recordRedisRequest(ctx, rdb, successKey, successMaxCount, duration)
		}
	}
}

// 内存限流处理器
func memoryRateLimitHandler(duration int64, totalMaxCount, successMaxCount int) gin.HandlerFunc {
	inMemoryRateLimiter.Init(time.Duration(setting.ModelRequestRateLimitDurationMinutes) * time.Minute)

	return func(c *gin.Context) {
		userId := strconv.Itoa(c.GetInt("id"))
		totalKey := ModelRequestRateLimitCountMark + userId
		successKey := ModelRequestRateLimitSuccessCountMark + userId

		// 1. 检查总请求数限制（当totalMaxCount为0时跳过）
		if totalMaxCount > 0 && !inMemoryRateLimiter.Request(totalKey, totalMaxCount, duration) {
			c.Status(http.StatusTooManyRequests)
			c.Abort()
			return
		}

		// 2. 检查成功请求数限制（只检查不记录，使用 Check 方法）
		if successMaxCount > 0 && !inMemoryRateLimiter.Check(successKey, successMaxCount, duration) {
			c.Status(http.StatusTooManyRequests)
			c.Abort()
			return
		}

		// 3. 处理请求
		c.Next()

		// 4. 如果请求成功，记录到实际的成功请求计数中
		if c.Writer.Status() < 400 {
			inMemoryRateLimiter.Request(successKey, successMaxCount, duration)
		}
	}
}

// checkTokenRateLimit 检查 token 分钟级限流
func checkTokenRateLimit(c *gin.Context) bool {
	if !setting.TokenRateLimitEnabled {
		fmt.Println("[DEBUG] TokenRateLimitEnabled is false, skipping minute rate limit")
		return true
	}

	tokenId := common.GetContextKeyInt(c, constant.ContextKeyTokenId)
	if tokenId == 0 {
		// 如果没有 token ID，跳过 per-key 限流
		fmt.Println("[DEBUG] tokenId is 0, skipping minute rate limit")
		return true
	}

	// 获取分组配置（优先使用 token group，如果为空则使用 user group）
	group := common.GetContextKeyString(c, constant.ContextKeyTokenGroup)
	if group == "" {
		group = common.GetContextKeyString(c, constant.ContextKeyUserGroup)
	}
	totalMaxCount := setting.TokenRateLimitCount
	successMaxCount := setting.TokenRateLimitSuccessCount

	// 获取分组的限流配置
	groupTotalCount, groupSuccessCount, found := setting.GetTokenRateLimit(group)
	if found {
		totalMaxCount = groupTotalCount
		successMaxCount = groupSuccessCount
	}

	fmt.Printf("[DEBUG] Minute rate limit check: tokenId=%d, group=%s, found=%v, totalMaxCount=%d, successMaxCount=%d\n",
		tokenId, group, found, totalMaxCount, successMaxCount)

	// 如果两个限制都为0，表示不限制
	if totalMaxCount == 0 && successMaxCount == 0 {
		fmt.Println("[DEBUG] Both minute limits are 0, skipping minute rate limit")
		return true
	}

	rateLimitKey := strconv.Itoa(tokenId)
	duration := int64(setting.TokenRateLimitDurationMinutes * 60)

	if common.RedisEnabled {
		return checkTokenRateLimitRedis(c, rateLimitKey, totalMaxCount, successMaxCount, duration)
	} else {
		return checkTokenRateLimitMemory(c, rateLimitKey, totalMaxCount, successMaxCount, duration)
	}
}

// checkTokenRateLimitRedis Redis版本的分钟级限流检查
func checkTokenRateLimitRedis(c *gin.Context, rateLimitKey string, totalMaxCount, successMaxCount int, duration int64) bool {
	ctx := context.Background()
	rdb := common.RDB

	// 1. 检查成功请求数限制
	if successMaxCount > 0 {
		successKey := fmt.Sprintf("rateLimit:%s:%s", TokenRateLimitSuccessCountMark, rateLimitKey)
		allowed, err := checkRedisRateLimit(ctx, rdb, successKey, successMaxCount, duration)
		if err != nil {
			fmt.Println("检查密钥成功请求数限制失败:", err.Error())
			abortWithOpenAiMessage(c, http.StatusInternalServerError, "rate_limit_check_failed")
			return false
		}
		if !allowed {
			abortWithOpenAiMessage(c, http.StatusTooManyRequests, fmt.Sprintf("您已达到密钥请求数限制：%d分钟内最多请求%d次", setting.TokenRateLimitDurationMinutes, successMaxCount))
			return false
		}
	}

	// 2. 检查总请求数限制
	if totalMaxCount > 0 {
		totalKey := fmt.Sprintf("rateLimit:%s:%s", TokenRateLimitCountMark, rateLimitKey)
		tb := limiter.New(ctx, rdb)
		allowed, err := tb.Allow(
			ctx,
			totalKey,
			limiter.WithCapacity(int64(totalMaxCount)*duration),
			limiter.WithRate(int64(totalMaxCount)),
			limiter.WithRequested(duration),
		)

		if err != nil {
			fmt.Println("检查密钥总请求数限制失败:", err.Error())
			abortWithOpenAiMessage(c, http.StatusInternalServerError, "rate_limit_check_failed")
			return false
		}

		if !allowed {
			abortWithOpenAiMessage(c, http.StatusTooManyRequests, fmt.Sprintf("您已达到密钥总请求数限制：%d分钟内最多请求%d次（包括失败请求）", setting.TokenRateLimitDurationMinutes, totalMaxCount))
			return false
		}
	}

	return true
}

// checkTokenRateLimitMemory 内存版本的分钟级限流检查
func checkTokenRateLimitMemory(c *gin.Context, rateLimitKey string, totalMaxCount, successMaxCount int, duration int64) bool {
	inMemoryRateLimiter.Init(time.Duration(setting.TokenRateLimitDurationMinutes) * time.Minute)

	totalKey := TokenRateLimitCountMark + rateLimitKey
	successKey := TokenRateLimitSuccessCountMark + rateLimitKey

	// 1. 检查总请求数限制
	if totalMaxCount > 0 && !inMemoryRateLimiter.Request(totalKey, totalMaxCount, duration) {
		abortWithOpenAiMessage(c, http.StatusTooManyRequests, fmt.Sprintf("您已达到密钥总请求数限制：%d分钟内最多请求%d次（包括失败请求）", setting.TokenRateLimitDurationMinutes, totalMaxCount))
		return false
	}

	// 2. 检查成功请求数限制（只检查不记录）
	if successMaxCount > 0 {
		if !inMemoryRateLimiter.Check(successKey, successMaxCount, duration) {
			abortWithOpenAiMessage(c, http.StatusTooManyRequests, fmt.Sprintf("您已达到密钥请求数限制：%d分钟内最多请求%d次", setting.TokenRateLimitDurationMinutes, successMaxCount))
			return false
		}
	}

	return true
}

// recordTokenRateLimitSuccess 记录分钟级成功请求
func recordTokenRateLimitSuccess(c *gin.Context) {
	if !setting.TokenRateLimitEnabled {
		return
	}

	tokenId := common.GetContextKeyInt(c, constant.ContextKeyTokenId)
	if tokenId == 0 {
		return
	}

	// 获取分组配置（优先使用 token group，如果为空则使用 user group）
	group := common.GetContextKeyString(c, constant.ContextKeyTokenGroup)
	if group == "" {
		group = common.GetContextKeyString(c, constant.ContextKeyUserGroup)
	}
	successMaxCount := setting.TokenRateLimitSuccessCount

	_, groupSuccessCount, found := setting.GetTokenRateLimit(group)
	if found {
		successMaxCount = groupSuccessCount
	}

	if successMaxCount == 0 {
		return
	}

	rateLimitKey := strconv.Itoa(tokenId)

	if common.RedisEnabled {
		ctx := context.Background()
		rdb := common.RDB
		successKey := fmt.Sprintf("rateLimit:%s:%s", TokenRateLimitSuccessCountMark, rateLimitKey)
		duration := int64(setting.TokenRateLimitDurationMinutes * 60)
		recordRedisRequest(ctx, rdb, successKey, successMaxCount, duration)
	} else {
		duration := int64(setting.TokenRateLimitDurationMinutes * 60)
		successKey := TokenRateLimitSuccessCountMark + rateLimitKey
		inMemoryRateLimiter.Request(successKey, successMaxCount, duration)
	}
}

// checkTokenDailyRateLimit 检查 token 每日限流
func checkTokenDailyRateLimit(c *gin.Context) bool {
	if !setting.TokenDailyRateLimitEnabled {
		fmt.Println("[DEBUG] TokenDailyRateLimitEnabled is false, skipping daily rate limit")
		return true
	}

	tokenId := common.GetContextKeyInt(c, constant.ContextKeyTokenId)
	if tokenId == 0 {
		// 如果没有 token ID，跳过 per-key 限流
		fmt.Println("[DEBUG] tokenId is 0, skipping daily rate limit")
		return true
	}

	// 获取分组配置（优先使用 token group，如果为空则使用 user group）
	group := common.GetContextKeyString(c, constant.ContextKeyTokenGroup)
	if group == "" {
		group = common.GetContextKeyString(c, constant.ContextKeyUserGroup)
	}
	totalMaxCount := setting.TokenDailyRateLimitCount
	successMaxCount := setting.TokenDailyRateLimitSuccessCount

	// 获取分组的限流配置
	groupTotalCount, groupSuccessCount, found := setting.GetTokenDailyRateLimit(group)
	if found {
		totalMaxCount = groupTotalCount
		successMaxCount = groupSuccessCount
	}

	fmt.Printf("[DEBUG] Daily rate limit check: tokenId=%d, group=%s, found=%v, totalMaxCount=%d, successMaxCount=%d\n",
		tokenId, group, found, totalMaxCount, successMaxCount)

	// 如果两个限制都为0，表示不限制
	if totalMaxCount == 0 && successMaxCount == 0 {
		fmt.Println("[DEBUG] Both limits are 0, skipping daily rate limit")
		return true
	}

	// 使用北京时间 (UTC+8) 计算当日的 key 后缀（格式：YYYYMMDD）
	// 这样每个自然日（北京时间）都有独立的 key，到次日 00:00 自动失效
	beijingLoc := time.FixedZone("CST", 8*3600) // UTC+8
	beijingNow := time.Now().In(beijingLoc)
	dateKey := beijingNow.Format("20060102")
	rateLimitKey := fmt.Sprintf("%d:%s", tokenId, dateKey)

	// 计算到北京时间次日 00:00 的剩余秒数
	nextDay := time.Date(beijingNow.Year(), beijingNow.Month(), beijingNow.Day()+1, 0, 0, 0, 0, beijingLoc)
	duration := int64(nextDay.Sub(beijingNow).Seconds()) + 60 // 额外 60 秒缓冲

	fmt.Printf("[DEBUG] Daily rate limit key: %s, duration until next Beijing day: %d seconds\n", rateLimitKey, duration)

	if common.RedisEnabled {
		return checkTokenDailyRateLimitRedis(c, rateLimitKey, totalMaxCount, successMaxCount, duration)
	} else {
		return checkTokenDailyRateLimitMemory(c, rateLimitKey, totalMaxCount, successMaxCount, duration)
	}
}

// checkTokenDailyRateLimitRedis Redis版本的每日限流检查
// 使用固定窗口计数器实现每日配额限制
func checkTokenDailyRateLimitRedis(c *gin.Context, rateLimitKey string, totalMaxCount, successMaxCount int, duration int64) bool {
	ctx := context.Background()
	rdb := common.RDB

	// Lua 脚本：原子性地增加计数并检查是否超限
	// 如果超限则回滚（减少计数）并返回 0，否则返回当前计数
	luaScript := `
		local count = redis.call('INCR', KEYS[1])
		if count == 1 then
			redis.call('EXPIRE', KEYS[1], ARGV[1])
		end
		if count > tonumber(ARGV[2]) then
			redis.call('DECR', KEYS[1])
			return 0
		end
		return count
	`

	// 1. 检查总请求数限制（先检查总数，因为总数包括失败请求）
	if totalMaxCount > 0 {
		totalKey := fmt.Sprintf("rateLimit:%s:%s", TokenDailyRateLimitCountMark, rateLimitKey)
		count, err := rdb.Eval(ctx, luaScript, []string{totalKey}, duration, totalMaxCount).Int64()
		if err != nil {
			fmt.Println("检查每日总请求数限制失败:", err.Error())
			abortWithOpenAiMessage(c, http.StatusInternalServerError, "rate_limit_check_failed")
			return false
		}
		if count == 0 {
			abortWithOpenAiMessage(c, http.StatusTooManyRequests, "您已达到每日总请求数限制（包括失败请求）")
			return false
		}
	}

	// 2. 检查并预占成功请求数配额
	if successMaxCount > 0 {
		successKey := fmt.Sprintf("rateLimit:%s:%s", TokenDailyRateLimitSuccessCountMark, rateLimitKey)
		count, err := rdb.Eval(ctx, luaScript, []string{successKey}, duration, successMaxCount).Int64()
		if err != nil {
			fmt.Println("检查每日成功请求数限制失败:", err.Error())
			abortWithOpenAiMessage(c, http.StatusInternalServerError, "rate_limit_check_failed")
			return false
		}
		if count == 0 {
			abortWithOpenAiMessage(c, http.StatusTooManyRequests, "您已达到每日请求数限制")
			return false
		}
		// 标记已预占成功请求配额，请求失败时需要回滚
		c.Set("daily_success_quota_reserved", true)
		c.Set("daily_success_quota_key", successKey)
	}

	return true
}

// rollbackDailySuccessQuota 回滚预占的成功请求配额
func rollbackDailySuccessQuota(c *gin.Context) {
	if reserved, exists := c.Get("daily_success_quota_reserved"); exists && reserved.(bool) {
		if key, exists := c.Get("daily_success_quota_key"); exists {
			ctx := context.Background()
			common.RDB.Decr(ctx, key.(string))
			c.Set("daily_success_quota_reserved", false)
		}
	}
}

// checkTokenDailyRateLimitMemory 内存版本的每日限流检查
func checkTokenDailyRateLimitMemory(c *gin.Context, rateLimitKey string, totalMaxCount, successMaxCount int, duration int64) bool {
	inMemoryRateLimiter.Init(24 * time.Hour)

	totalKey := TokenDailyRateLimitCountMark + rateLimitKey
	successKey := TokenDailyRateLimitSuccessCountMark + rateLimitKey

	// 1. 检查总请求数限制
	if totalMaxCount > 0 && !inMemoryRateLimiter.Request(totalKey, totalMaxCount, duration) {
		abortWithOpenAiMessage(c, http.StatusTooManyRequests, "您已达到每日总请求数限制（包括失败请求）")
		return false
	}

	// 2. 检查成功请求数限制（只检查不记录）
	if successMaxCount > 0 {
		if !inMemoryRateLimiter.Check(successKey, successMaxCount, duration) {
			abortWithOpenAiMessage(c, http.StatusTooManyRequests, "您已达到每日请求数限制")
			return false
		}
	}

	return true
}

// recordTokenDailySuccess 处理每日成功请求计数
// 由于成功请求配额在检查时已预占，这里只需要处理请求失败时的回滚
// 请求成功时不需要额外操作（配额已预占）
func recordTokenDailySuccess(c *gin.Context) {
	// 请求成功，配额已在检查时预占，清除标记即可
	if common.RedisEnabled {
		c.Set("daily_success_quota_reserved", false)
	} else {
		// 内存模式：请求成功时记录
		if !setting.TokenDailyRateLimitEnabled {
			return
		}

		tokenId := common.GetContextKeyInt(c, constant.ContextKeyTokenId)
		if tokenId == 0 {
			return
		}

		group := common.GetContextKeyString(c, constant.ContextKeyTokenGroup)
		if group == "" {
			group = common.GetContextKeyString(c, constant.ContextKeyUserGroup)
		}
		successMaxCount := setting.TokenDailyRateLimitSuccessCount

		_, groupSuccessCount, found := setting.GetTokenDailyRateLimit(group)
		if found {
			successMaxCount = groupSuccessCount
		}

		if successMaxCount == 0 {
			return
		}

		rateLimitKey := strconv.Itoa(tokenId)
		duration := int64(86400)
		successKey := TokenDailyRateLimitSuccessCountMark + rateLimitKey
		inMemoryRateLimiter.Request(successKey, successMaxCount, duration)
	}
}

// rollbackTokenDailyQuota 请求失败时回滚预占的每日配额
func rollbackTokenDailyQuota(c *gin.Context) {
	if common.RedisEnabled {
		rollbackDailySuccessQuota(c)
	}
}

// ModelRequestRateLimit 模型请求限流中间件
func ModelRequestRateLimit() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 1. 先检查 per-key 分钟级限流
		if !checkTokenRateLimit(c) {
			return
		}

		// 2. 检查 per-key 每日限流
		if !checkTokenDailyRateLimit(c) {
			return
		}

		// 3. 再检查原有的 per-user 限流（保持兼容性）
		if !setting.ModelRequestRateLimitEnabled {
			c.Next()
			// 请求成功后确认配额，失败则回滚
			if c.Writer.Status() < 400 {
				recordTokenRateLimitSuccess(c)
				recordTokenDailySuccess(c)
			} else {
				rollbackTokenDailyQuota(c)
			}
			return
		}

		// 计算限流参数
		duration := int64(setting.ModelRequestRateLimitDurationMinutes * 60)
		totalMaxCount := setting.ModelRequestRateLimitCount
		successMaxCount := setting.ModelRequestRateLimitSuccessCount

		// per-user 限流使用 user group（不是 token group）
		userGroup := common.GetContextKeyString(c, constant.ContextKeyUserGroup)

		//获取分组的限流配置
		groupTotalCount, groupSuccessCount, found := setting.GetGroupRateLimit(userGroup)
		if found {
			totalMaxCount = groupTotalCount
			successMaxCount = groupSuccessCount
		}

		// 根据存储类型选择并执行限流处理器
		if common.RedisEnabled {
			redisRateLimitHandler(duration, totalMaxCount, successMaxCount)(c)
		} else {
			memoryRateLimitHandler(duration, totalMaxCount, successMaxCount)(c)
		}

		// 请求成功后确认配额，失败则回滚
		if c.Writer.Status() < 400 {
			recordTokenRateLimitSuccess(c)
			recordTokenDailySuccess(c)
		} else {
			rollbackTokenDailyQuota(c)
		}
	}
}
```

### `setting/rate_limit.go`

```go
package setting

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"

	"github.com/QuantumNous/new-api/common"
)

// Per-user rate limit settings (原有的按用户限流)
var ModelRequestRateLimitEnabled = false
var ModelRequestRateLimitDurationMinutes = 1
var ModelRequestRateLimitCount = 0
var ModelRequestRateLimitSuccessCount = 1000
var ModelRequestRateLimitGroup = map[string][2]int{}
var ModelRequestRateLimitMutex sync.RWMutex

// Per-key minute rate limit settings (按密钥的分钟级限流)
var TokenRateLimitEnabled = false
var TokenRateLimitDurationMinutes = 1
var TokenRateLimitCount = 0
var TokenRateLimitSuccessCount = 0
var TokenRateLimitGroup = map[string][2]int{}
var TokenRateLimitMutex sync.RWMutex

// Per-key daily rate limit settings (按密钥的每日限流)
var TokenDailyRateLimitEnabled = false
var TokenDailyRateLimitCount = 0        // 每日总请求数限制（0表示不限制）
var TokenDailyRateLimitSuccessCount = 0 // 每日成功请求数限制（0表示不限制）
var TokenDailyRateLimitGroup = map[string][2]int{}
var TokenDailyRateLimitMutex sync.RWMutex

func ModelRequestRateLimitGroup2JSONString() string {
	ModelRequestRateLimitMutex.RLock()
	defer ModelRequestRateLimitMutex.RUnlock()

	jsonBytes, err := json.Marshal(ModelRequestRateLimitGroup)
	if err != nil {
		common.SysLog("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateModelRequestRateLimitGroupByJSONString(jsonStr string) error {
	ModelRequestRateLimitMutex.Lock()
	defer ModelRequestRateLimitMutex.Unlock()

	ModelRequestRateLimitGroup = make(map[string][2]int)
	return json.Unmarshal([]byte(jsonStr), &ModelRequestRateLimitGroup)
}

func GetGroupRateLimit(group string) (totalCount, successCount int, found bool) {
	ModelRequestRateLimitMutex.RLock()
	defer ModelRequestRateLimitMutex.RUnlock()

	if ModelRequestRateLimitGroup == nil {
		return 0, 0, false
	}

	limits, found := ModelRequestRateLimitGroup[group]
	if !found {
		return 0, 0, false
	}
	return limits[0], limits[1], true
}

func CheckModelRequestRateLimitGroup(jsonStr string) error {
	checkModelRequestRateLimitGroup := make(map[string][2]int)
	err := json.Unmarshal([]byte(jsonStr), &checkModelRequestRateLimitGroup)
	if err != nil {
		return err
	}
	for group, limits := range checkModelRequestRateLimitGroup {
		if limits[0] < 0 || limits[1] < 1 {
			return fmt.Errorf("group %s has negative rate limit values: [%d, %d]", group, limits[0], limits[1])
		}
		if limits[0] > math.MaxInt32 || limits[1] > math.MaxInt32 {
			return fmt.Errorf("group %s [%d, %d] has max rate limits value 2147483647", group, limits[0], limits[1])
		}
	}

	return nil
}

// Token minute rate limit functions
func TokenRateLimitGroup2JSONString() string {
	TokenRateLimitMutex.RLock()
	defer TokenRateLimitMutex.RUnlock()

	jsonBytes, err := json.Marshal(TokenRateLimitGroup)
	if err != nil {
		common.SysLog("error marshalling token rate limit group: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateTokenRateLimitGroupByJSONString(jsonStr string) error {
	TokenRateLimitMutex.Lock()
	defer TokenRateLimitMutex.Unlock()

	TokenRateLimitGroup = make(map[string][2]int)
	return json.Unmarshal([]byte(jsonStr), &TokenRateLimitGroup)
}

func GetTokenRateLimit(group string) (totalCount, successCount int, found bool) {
	TokenRateLimitMutex.RLock()
	defer TokenRateLimitMutex.RUnlock()

	if TokenRateLimitGroup == nil {
		return 0, 0, false
	}

	limits, found := TokenRateLimitGroup[group]
	if !found {
		return 0, 0, false
	}
	return limits[0], limits[1], true
}

func CheckTokenRateLimitGroup(jsonStr string) error {
	checkTokenRateLimitGroup := make(map[string][2]int)
	err := json.Unmarshal([]byte(jsonStr), &checkTokenRateLimitGroup)
	if err != nil {
		return err
	}
	for group, limits := range checkTokenRateLimitGroup {
		if limits[0] < 0 || limits[1] < 0 {
			return fmt.Errorf("group %s has negative rate limit values: [%d, %d]", group, limits[0], limits[1])
		}
		if limits[0] > math.MaxInt32 || limits[1] > math.MaxInt32 {
			return fmt.Errorf("group %s [%d, %d] has max rate limits value 2147483647", group, limits[0], limits[1])
		}
	}

	return nil
}

// Token daily rate limit functions
func TokenDailyRateLimitGroup2JSONString() string {
	TokenDailyRateLimitMutex.RLock()
	defer TokenDailyRateLimitMutex.RUnlock()

	jsonBytes, err := json.Marshal(TokenDailyRateLimitGroup)
	if err != nil {
		common.SysLog("error marshalling token daily rate limit group: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateTokenDailyRateLimitGroupByJSONString(jsonStr string) error {
	TokenDailyRateLimitMutex.Lock()
	defer TokenDailyRateLimitMutex.Unlock()

	TokenDailyRateLimitGroup = make(map[string][2]int)
	return json.Unmarshal([]byte(jsonStr), &TokenDailyRateLimitGroup)
}

func GetTokenDailyRateLimit(group string) (totalCount, successCount int, found bool) {
	TokenDailyRateLimitMutex.RLock()
	defer TokenDailyRateLimitMutex.RUnlock()

	if TokenDailyRateLimitGroup == nil {
		return 0, 0, false
	}

	limits, found := TokenDailyRateLimitGroup[group]
	if !found {
		return 0, 0, false
	}
	return limits[0], limits[1], true
}

func CheckTokenDailyRateLimitGroup(jsonStr string) error {
	checkTokenDailyRateLimitGroup := make(map[string][2]int)
	err := json.Unmarshal([]byte(jsonStr), &checkTokenDailyRateLimitGroup)
	if err != nil {
		return err
	}
	for group, limits := range checkTokenDailyRateLimitGroup {
		if limits[0] < 0 || limits[1] < 0 {
			return fmt.Errorf("group %s has negative rate limit values: [%d, %d]", group, limits[0], limits[1])
		}
		if limits[0] > math.MaxInt32 || limits[1] > math.MaxInt32 {
			return fmt.Errorf("group %s [%d, %d] has max rate limits value 2147483647", group, limits[0], limits[1])
		}
	}

	return nil
}
```

### `common/rate-limit.go`

```go
package common

import (
	"sync"
	"time"
)

type InMemoryRateLimiter struct {
	store              map[string]*[]int64
	mutex              sync.Mutex
	expirationDuration time.Duration
}

func (l *InMemoryRateLimiter) Init(expirationDuration time.Duration) {
	if l.store == nil {
		l.mutex.Lock()
		if l.store == nil {
			l.store = make(map[string]*[]int64)
			l.expirationDuration = expirationDuration
			if expirationDuration > 0 {
				go l.clearExpiredItems()
			}
		}
		l.mutex.Unlock()
	}
}

func (l *InMemoryRateLimiter) clearExpiredItems() {
	for {
		time.Sleep(l.expirationDuration)
		l.mutex.Lock()
		now := time.Now().Unix()
		for key := range l.store {
			queue := l.store[key]
			size := len(*queue)
			if size == 0 || now-(*queue)[size-1] > int64(l.expirationDuration.Seconds()) {
				delete(l.store, key)
			}
		}
		l.mutex.Unlock()
	}
}

// Request parameter duration's unit is seconds
func (l *InMemoryRateLimiter) Request(key string, maxRequestNum int, duration int64) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	// [old <-- new]
	queue, ok := l.store[key]
	now := time.Now().Unix()
	if ok {
		if len(*queue) < maxRequestNum {
			*queue = append(*queue, now)
			return true
		} else {
			if now-(*queue)[0] >= duration {
				*queue = (*queue)[1:]
				*queue = append(*queue, now)
				return true
			} else {
				return false
			}
		}
	} else {
		s := make([]int64, 0, maxRequestNum)
		l.store[key] = &s
		*(l.store[key]) = append(*(l.store[key]), now)
	}
	return true
}

// Check 只检查是否达到限制，不记录请求（用于成功请求数限制的预检查）
func (l *InMemoryRateLimiter) Check(key string, maxRequestNum int, duration int64) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	queue, ok := l.store[key]
	if !ok {
		return true // 没有记录，允许
	}
	now := time.Now().Unix()
	if len(*queue) < maxRequestNum {
		return true // 未达到限制
	}
	// 检查最老的记录是否已过期
	if now-(*queue)[0] >= duration {
		return true // 时间窗口已过，允许
	}
	return false // 在时间窗口内已达到限制
}
```

### `common/limiter/lua/rate_limit.lua`

```lua
-- 令牌桶限流器
-- KEYS[1]: 限流器唯一标识
-- ARGV[1]: 请求令牌数 (通常为1)
-- ARGV[2]: 令牌生成速率 (每秒)
-- ARGV[3]: 桶容量

local key = KEYS[1]
local requested = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local capacity = tonumber(ARGV[3])

-- 获取当前时间（Redis服务器时间）
local now = redis.call('TIME')
local nowInSeconds = tonumber(now[1])

-- 获取桶状态
local bucket = redis.call('HMGET', key, 'tokens', 'last_time')
local tokens = tonumber(bucket[1])
local last_time = tonumber(bucket[2])

-- 初始化桶（首次请求或过期）
if not tokens or not last_time then
    tokens = capacity
    last_time = nowInSeconds
else
    -- 计算新增令牌
    local elapsed = nowInSeconds - last_time
    local add_tokens = elapsed * rate
    tokens = math.min(capacity, tokens + add_tokens)
    last_time = nowInSeconds
end

-- 判断是否允许请求
local allowed = false
if tokens >= requested then
    tokens = tokens - requested
    allowed = true
end

--- 更新桶状态并设置过期时间
redis.call('HMSET', key, 'tokens', tokens, 'last_time', last_time)
redis.call('EXPIRE', key, math.ceil(capacity / rate) + 60) -- 适当延长过期时间

return allowed and 1 or 0
```

### `controller/video_proxy_gemini.go`

```go
package controller

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay"
)

// geminiVideoResult holds either a remote URL or inline base64 video data.
type geminiVideoResult struct {
	URL      string // remote video URL (preferred if available)
	Data     []byte // decoded video bytes from bytesBase64Encoded
	MimeType string // e.g. "video/mp4"
}

func getGeminiVideoResult(channel *model.Channel, task *model.Task, apiKey string) (*geminiVideoResult, error) {
	if channel == nil || task == nil {
		return nil, fmt.Errorf("invalid channel or task")
	}

	// Try extracting URL from stored task data first
	if url := extractGeminiVideoURLFromTaskData(task); url != "" {
		return &geminiVideoResult{URL: ensureAPIKey(url, apiKey)}, nil
	}

	baseURL := constant.ChannelBaseURLs[channel.Type]
	if channel.GetBaseURL() != "" {
		baseURL = channel.GetBaseURL()
	}

	adaptor := relay.GetTaskAdaptor(constant.TaskPlatform(strconv.Itoa(channel.Type)))
	if adaptor == nil {
		return nil, fmt.Errorf("gemini task adaptor not found")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("api key not available for task")
	}

	proxy := channel.GetSetting().Proxy
	resp, err := adaptor.FetchTask(baseURL, apiKey, map[string]any{
		"task_id": task.TaskID,
		"action":  task.Action,
	}, proxy)
	if err != nil {
		return nil, fmt.Errorf("fetch task failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read task response failed: %w", err)
	}

	// Try remote URL from parsed result
	taskInfo, parseErr := adaptor.ParseTaskResult(body)
	if parseErr == nil && taskInfo != nil && taskInfo.RemoteUrl != "" {
		return &geminiVideoResult{URL: ensureAPIKey(taskInfo.RemoteUrl, apiKey)}, nil
	}

	// Try remote URL from raw payload
	if url := extractGeminiVideoURLFromPayload(body); url != "" {
		return &geminiVideoResult{URL: ensureAPIKey(url, apiKey)}, nil
	}

	// Try inline base64 video data (Veo returns bytesBase64Encoded directly)
	if result := extractGeminiVideoBase64FromPayload(body); result != nil {
		return result, nil
	}

	if parseErr != nil {
		return nil, fmt.Errorf("parse task result failed: %w", parseErr)
	}

	return nil, fmt.Errorf("gemini video data not found")
}

// extractGeminiVideoBase64FromPayload extracts inline base64 video from fetchPredictOperation response.
func extractGeminiVideoBase64FromPayload(body []byte) *geminiVideoResult {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}

	resp, ok := payload["response"].(map[string]any)
	if !ok {
		return nil
	}

	// Check response.videos[0].bytesBase64Encoded
	if videos, ok := resp["videos"].([]any); ok && len(videos) > 0 {
		if vm, ok := videos[0].(map[string]any); ok {
			if b64, ok := vm["bytesBase64Encoded"].(string); ok && b64 != "" {
				data, err := base64.StdEncoding.DecodeString(b64)
				if err != nil {
					return nil
				}
				mime := "video/mp4"
				if m, ok := vm["mimeType"].(string); ok && m != "" {
					mime = m
				}
				return &geminiVideoResult{Data: data, MimeType: mime}
			}
		}
	}

	// Check response.bytesBase64Encoded directly
	if b64, ok := resp["bytesBase64Encoded"].(string); ok && b64 != "" {
		data, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil
		}
		return &geminiVideoResult{Data: data, MimeType: "video/mp4"}
	}

	return nil
}

func extractGeminiVideoURLFromTaskData(task *model.Task) string {
	if task == nil || len(task.Data) == 0 {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(task.Data, &payload); err != nil {
		return ""
	}
	return extractGeminiVideoURLFromMap(payload)
}

func extractGeminiVideoURLFromPayload(body []byte) string {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	return extractGeminiVideoURLFromMap(payload)
}

func extractGeminiVideoURLFromMap(payload map[string]any) string {
	if payload == nil {
		return ""
	}
	if uri, ok := payload["uri"].(string); ok && uri != "" {
		return uri
	}
	if resp, ok := payload["response"].(map[string]any); ok {
		if uri := extractGeminiVideoURLFromResponse(resp); uri != "" {
			return uri
		}
	}
	return ""
}

func extractGeminiVideoURLFromResponse(resp map[string]any) string {
	if resp == nil {
		return ""
	}
	if gvr, ok := resp["generateVideoResponse"].(map[string]any); ok {
		if uri := extractGeminiVideoURLFromGeneratedSamples(gvr); uri != "" {
			return uri
		}
	}
	if videos, ok := resp["videos"].([]any); ok {
		for _, video := range videos {
			if vm, ok := video.(map[string]any); ok {
				if uri, ok := vm["uri"].(string); ok && uri != "" {
					return uri
				}
			}
		}
	}
	if uri, ok := resp["video"].(string); ok && uri != "" {
		return uri
	}
	if uri, ok := resp["uri"].(string); ok && uri != "" {
		return uri
	}
	return ""
}

func extractGeminiVideoURLFromGeneratedSamples(gvr map[string]any) string {
	if gvr == nil {
		return ""
	}
	if samples, ok := gvr["generatedSamples"].([]any); ok {
		for _, sample := range samples {
			if sm, ok := sample.(map[string]any); ok {
				if video, ok := sm["video"].(map[string]any); ok {
					if uri, ok := video["uri"].(string); ok && uri != "" {
						return uri
					}
				}
			}
		}
	}
	return ""
}

func ensureAPIKey(uri, key string) string {
	if key == "" || uri == "" {
		return uri
	}
	if strings.Contains(uri, "key=") {
		return uri
	}
	if strings.Contains(uri, "?") {
		return fmt.Sprintf("%s&key=%s", uri, key)
	}
	return fmt.Sprintf("%s?key=%s", uri, key)
}
```

### `relay/channel/task/gemini/adaptor.go`

```go
package gemini

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay/channel"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/model_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// ============================
// Request / Response structures
// ============================

// GeminiVideoGenerationConfig represents the video generation configuration
// Based on: https://ai.google.dev/gemini-api/docs/video
type GeminiVideoGenerationConfig struct {
	AspectRatio      string  `json:"aspectRatio,omitempty"`      // "16:9" or "9:16"
	DurationSeconds  float64 `json:"durationSeconds,omitempty"`  // 4, 6, or 8 (as number)
	NegativePrompt   string  `json:"negativePrompt,omitempty"`   // unwanted elements
	PersonGeneration string  `json:"personGeneration,omitempty"` // "allow_all" for text-to-video, "allow_adult" for image-to-video
	Resolution       string  `json:"resolution,omitempty"`       // video resolution
}

// GeminiVideoRequest represents a single video generation instance
type GeminiVideoRequest struct {
	Prompt string `json:"prompt"`
}

// GeminiVideoPayload represents the complete video generation request payload
type GeminiVideoPayload struct {
	Instances  []GeminiVideoRequest        `json:"instances"`
	Parameters GeminiVideoGenerationConfig `json:"parameters,omitempty"`
}

type submitResponse struct {
	Name string `json:"name"`
}

type operationVideo struct {
	MimeType           string `json:"mimeType"`
	BytesBase64Encoded string `json:"bytesBase64Encoded"`
	Encoding           string `json:"encoding"`
}

type operationResponse struct {
	Name     string `json:"name"`
	Done     bool   `json:"done"`
	Response struct {
		Type                  string           `json:"@type"`
		RaiMediaFilteredCount int              `json:"raiMediaFilteredCount"`
		Videos                []operationVideo `json:"videos"`
		BytesBase64Encoded    string           `json:"bytesBase64Encoded"`
		Encoding              string           `json:"encoding"`
		Video                 string           `json:"video"`
		GenerateVideoResponse struct {
			GeneratedSamples []struct {
				Video struct {
					URI string `json:"uri"`
				} `json:"video"`
			} `json:"generatedSamples"`
		} `json:"generateVideoResponse"`
	} `json:"response"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ============================
// Adaptor implementation
// ============================

type TaskAdaptor struct {
	ChannelType int
	apiKey      string
	baseURL     string
}

func (a *TaskAdaptor) Init(info *relaycommon.RelayInfo) {
	a.ChannelType = info.ChannelType
	a.baseURL = info.ChannelBaseUrl
	a.apiKey = info.ApiKey
}

// ValidateRequestAndSetAction parses body, validates fields and sets default action.
func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) (taskErr *dto.TaskError) {
	// Use the standard validation method for TaskSubmitReq
	return relaycommon.ValidateBasicTaskRequest(c, info, constant.TaskActionTextGenerate)
}

// BuildRequestURL constructs the upstream URL.
func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.RelayInfo) (string, error) {
	modelName := info.OriginModelName
	version := model_setting.GetGeminiVersionSetting(modelName)

	return fmt.Sprintf(
		"%s/%s/models/%s:predictLongRunning",
		a.baseURL,
		version,
		modelName,
	), nil
}

// BuildRequestHeader sets required headers.
func (a *TaskAdaptor) BuildRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.RelayInfo) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-goog-api-key", a.apiKey)
	return nil
}

// BuildRequestBody converts request into Gemini specific format.
func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.RelayInfo) (io.Reader, error) {
	v, ok := c.Get("task_request")
	if !ok {
		return nil, fmt.Errorf("request not found in context")
	}
	req, ok := v.(relaycommon.TaskSubmitReq)
	if !ok {
		return nil, fmt.Errorf("unexpected task_request type")
	}

	// Create structured video generation request
	body := GeminiVideoPayload{
		Instances: []GeminiVideoRequest{
			{Prompt: req.Prompt},
		},
		Parameters: GeminiVideoGenerationConfig{},
	}

	metadata := req.Metadata
	medaBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, errors.Wrap(err, "metadata marshal metadata failed")
	}
	err = json.Unmarshal(medaBytes, &body.Parameters)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal metadata failed")
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

// DoRequest delegates to common helper.
func (a *TaskAdaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoTaskApiRequest(a, c, info, requestBody)
}

// DoResponse handles upstream response, returns taskID etc.
func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (taskID string, taskData []byte, taskErr *dto.TaskError) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
	}
	_ = resp.Body.Close()

	var s submitResponse
	if err := json.Unmarshal(responseBody, &s); err != nil {
		return "", nil, service.TaskErrorWrapper(err, "unmarshal_response_failed", http.StatusInternalServerError)
	}
	if strings.TrimSpace(s.Name) == "" {
		return "", nil, service.TaskErrorWrapper(fmt.Errorf("missing operation name"), "invalid_response", http.StatusInternalServerError)
	}
	taskID = encodeLocalTaskID(s.Name)
	ov := dto.NewOpenAIVideo()
	ov.ID = taskID
	ov.TaskID = taskID
	ov.CreatedAt = time.Now().Unix()
	ov.Model = info.OriginModelName
	c.JSON(http.StatusOK, ov)
	return taskID, responseBody, nil
}

func (a *TaskAdaptor) GetModelList() []string {
	return []string{"veo-3.0-generate-001", "veo-3.1-generate-preview", "veo-3.1-fast-generate-preview"}
}

func (a *TaskAdaptor) GetChannelName() string {
	return "gemini"
}

// FetchTask fetch task status
func (a *TaskAdaptor) FetchTask(baseUrl, key string, body map[string]any, proxy string) (*http.Response, error) {
	taskID, ok := body["task_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid task_id")
	}

	upstreamName, err := decodeLocalTaskID(taskID)
	if err != nil {
		return nil, fmt.Errorf("decode task_id failed: %w", err)
	}

	// Extract model name from operation name to build fetchPredictOperation URL
	modelName := extractModelFromOperationName(upstreamName)
	if modelName == "" {
		return nil, fmt.Errorf("cannot extract model name from operation: %s", upstreamName)
	}

	version := model_setting.GetGeminiVersionSetting(modelName)
	url := fmt.Sprintf("%s/%s/models/%s:fetchPredictOperation", baseUrl, version, modelName)

	fetchBody, err := json.Marshal(map[string]string{
		"operationName": upstreamName,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal fetch body failed: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(fetchBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-goog-api-key", key)

	client, err := service.GetHttpClientWithProxy(proxy)
	if err != nil {
		return nil, fmt.Errorf("new proxy http client failed: %w", err)
	}
	return client.Do(req)
}

func (a *TaskAdaptor) ParseTaskResult(respBody []byte) (*relaycommon.TaskInfo, error) {
	var op operationResponse
	if err := json.Unmarshal(respBody, &op); err != nil {
		return nil, fmt.Errorf("unmarshal operation response failed: %w", err)
	}

	ti := &relaycommon.TaskInfo{}

	if op.Error.Message != "" {
		ti.Status = model.TaskStatusFailure
		ti.Reason = op.Error.Message
		ti.Progress = "100%"
		return ti, nil
	}

	if !op.Done {
		ti.Status = model.TaskStatusInProgress
		ti.Progress = "50%"
		return ti, nil
	}

	ti.Status = model.TaskStatusSuccess
	ti.Progress = "100%"

	taskID := encodeLocalTaskID(op.Name)
	ti.TaskID = taskID
	ti.Url = fmt.Sprintf("%s/v1/videos/%s/content", system_setting.ServerAddress, taskID)

	// Extract URL from generateVideoResponse if available
	if len(op.Response.GenerateVideoResponse.GeneratedSamples) > 0 {
		if uri := op.Response.GenerateVideoResponse.GeneratedSamples[0].Video.URI; uri != "" {
			ti.RemoteUrl = uri
		}
	}

	return ti, nil
}

func (a *TaskAdaptor) ConvertToOpenAIVideo(task *model.Task) ([]byte, error) {
	upstreamName, err := decodeLocalTaskID(task.TaskID)
	if err != nil {
		upstreamName = ""
	}
	modelName := extractModelFromOperationName(upstreamName)
	if strings.TrimSpace(modelName) == "" {
		modelName = "veo-3.0-generate-001"
	}

	video := dto.NewOpenAIVideo()
	video.ID = task.TaskID
	video.Model = modelName
	video.Status = task.Status.ToVideoStatus()
	video.SetProgressStr(task.Progress)
	video.CreatedAt = task.CreatedAt
	if task.FinishTime > 0 {
		video.CompletedAt = task.FinishTime
	} else if task.UpdatedAt > 0 {
		video.CompletedAt = task.UpdatedAt
	}

	return common.Marshal(video)
}

// ============================
// helpers
// ============================

func encodeLocalTaskID(name string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(name))
}

func decodeLocalTaskID(local string) (string, error) {
	b, err := base64.RawURLEncoding.DecodeString(local)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

var modelRe = regexp.MustCompile(`models/([^/]+)/operations/`)

func extractModelFromOperationName(name string) string {
	if name == "" {
		return ""
	}
	if m := modelRe.FindStringSubmatch(name); len(m) == 2 {
		return m[1]
	}
	if idx := strings.Index(name, "models/"); idx >= 0 {
		s := name[idx+len("models/"):]
		if p := strings.Index(s, "/operations/"); p > 0 {
			return s[:p]
		}
	}
	return ""
}
```

